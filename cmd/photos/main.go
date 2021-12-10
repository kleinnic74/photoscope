// photos project main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"time"

	"github.com/gorilla/mux"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"

	"bitbucket.org/kleinnic74/photos/classification"
	"bitbucket.org/kleinnic74/photos/consts"
	"bitbucket.org/kleinnic74/photos/domain"
	"bitbucket.org/kleinnic74/photos/events"
	"bitbucket.org/kleinnic74/photos/geocoding"
	"bitbucket.org/kleinnic74/photos/geocoding/openstreetmap"
	"bitbucket.org/kleinnic74/photos/importer"
	"bitbucket.org/kleinnic74/photos/index"
	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/library/boltstore"
	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/rest"
	"bitbucket.org/kleinnic74/photos/rest/wdav"
	"bitbucket.org/kleinnic74/photos/swarm"
	"bitbucket.org/kleinnic74/photos/tasks"
	"github.com/kleinnic74/fflags"
)

var (
	dbName = "photos.db"

	libDir string
	uiDir  string
	port   uint

	logger *zap.Logger
	ctx    context.Context

	eventFeature = fflags.Define("index.events")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s  [options]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&libDir, "l", "gophotos", "Path to photo library")
	flag.StringVar(&uiDir, "ui", "", "Path to the frontend static assets")
	flag.UintVar(&port, "p", 8080, "HTTP server port")
	logger, ctx = logging.SubFrom(context.Background(), "main")

	flag.Parse()

	absdir, err := filepath.Abs(libDir)
	if err != nil {
		logger.Fatal("Could not determine path", zap.String("dir", libDir), zap.Error(err))
	}
	libDir = absdir
	logger.Info("Photoscope starting", zap.String("gitCommit", consts.GitCommit), zap.String("gitRepo", consts.GitRepo))
	logger.Info("Library directory", zap.String("dir", libDir))
}

func main() {
	defer func() {
		if consts.IsDevMode() {
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 1)
		}
	}()

	if err := os.MkdirAll(libDir, os.ModePerm); err != nil {
		logger.Fatal("Failed to create directory", zap.String("dir", libDir), zap.Error(err))
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		oscall := <-signals
		logger.Info("Received signal", zap.Any("signal", oscall))
		cancel()
	}()

	router := mux.NewRouter()

	db, err := bolt.Open(filepath.Join(libDir, dbName), 0600, nil)
	if err != nil {
		logger.Fatal("Failed to initialize data store", zap.Error(err))
	}
	defer func() {
		db.Close()
		logger.Info("Closed data store")
	}()

	taskRepo := tasks.NewTaskRepository()
	tasks.RegisterTasks(taskRepo)
	importer.RegisterTasks(taskRepo)

	migrator, err := index.NewMigrationCoordinator(db)
	if err != nil {
		logger.Fatal("Failed to initialize migration coordinator", zap.Error(err))
	}

	indexTracker, err := boltstore.NewIndexTracker(db)
	if err != nil {
		logger.Fatal("Failed to initialize library", zap.Error(err))
	}

	store, err := boltstore.NewBoltStore(db)
	if err != nil {
		logger.Fatal("Failed to initialize library", zap.Error(err))
	}

	thumbers := &domain.Thumbers{}
	thumbers.Add(domain.LocalThumber{}, 1)

	lib, err := library.NewBasicPhotoLibrary(libDir, store, thumbers)
	if err != nil {
		logger.Fatal("Failed to initialize library", zap.Error(err))
	}
	logger.Info("Opened photo library", zap.String("path", libDir))
	migrator.AddInstances(lib)

	geoindex, err := boltstore.NewBoltGeoIndex(db)
	if err != nil {
		logger.Fatal("Failed to initialize geoindex", zap.Error(err))
	}
	migrator.AddStructure("geo", geoindex)

	geocoder := geocoding.NewGeocoder(geoindex, openstreetmap.NewResolver("de,en"))
	geocoder.RegisterTasks(taskRepo)

	dateindex, err := boltstore.NewDateIndex(db)
	if err != nil {
		logger.Fatal("Failed to initialize dataindex", zap.Error(err))
	}
	migrator.AddStructure("date", dateindex)

	fflags.IfEnabled(eventFeature, func() error {
		eventindex, err := boltstore.NewEventIndex(db)
		if err != nil {
			return fmt.Errorf("Failed to initialize event database: %w", err)
		}
		classification.RegisterTasks(taskRepo, eventindex)
		events := rest.NewEventsHandler(eventindex, lib)
		events.InitRoutes(router)
		return nil
	})

	bus := events.NewStream()
	go bus.Dispatch(ctx)

	executor := tasks.NewSerialTaskExecutor(lib)
	go executor.DrainTasks(ctx, func(e tasks.Execution) {
		bus.Publish(events.Event{Name: "tasks", Action: "completed"})
	})

	indexer := index.NewIndexer(indexTracker, executor)
	indexer.RegisterDirect("date", boltstore.DateIndexVersion, dateindex.Add)
	indexer.RegisterDefered("geo", boltstore.GeoIndexVersion, geocoder.LookupPhotoOnAdd)

	indexer.RegisterTasks(taskRepo)

	RegisterMigrationTask(taskRepo, migrator, indexer)

	lib.AddCallback(indexer.Add)

	go launchStartupTasks(ctx, taskRepo, executor)

	instance, err := NewInstance(ctx, lib.ID, DefaultInstanceProperties()...)
	if err != nil {
		logger.Fatal("Failed to initialize library unique ID", zap.Error(err))
	}
	logger, ctx = logging.FromWithFields(ctx, zap.String("instance", instance.ID))

	peers, err := swarm.NewController(instance, port)
	peers.OnPeerDetected(swarm.SkipSelf(addRemoteThumber(instance.ID, thumbers)))
	peers.OnPeerDetected(swarm.SkipSelf(addRemoteSync(executor)))

	go peers.ListenAndServe(ctx)

	// REST Handlers

	metrics := rest.NewMetricsHandler()
	metrics.InitRoutes(router)

	if consts.IsDevMode() {
		logs := rest.NewLogsHandler()
		logs.InitRoutes(router)
	}

	sse := rest.NewSSEHandler(bus)
	sse.InitRoutes(router)

	photoApp := rest.NewApp(lib)
	photoApp.InitRoutes(router)

	timeline := rest.NewTimelineHandler(dateindex, lib)
	timeline.InitRoutes(router)

	geo := rest.NewGeoHandler(geoindex, lib)
	geo.InitRoutes(router)

	geocache := rest.NewGeoCacheHandler(geocoder.Cache)
	geocache.InitRoutes(router)

	tasksApp := rest.NewTaskHandler(taskRepo, executor)
	tasksApp.InitRoutes(router)

	indexesRest := rest.NewIndexes(indexer, migrator)
	indexesRest.Init(router)

	peersRest := rest.NewPeersAPI(peers)
	peersRest.InitRoutes(router)

	thumbService := rest.NewThumberAPI(domain.LocalThumber{})
	thumbService.InitRoutes(router)

	tmpdir := filepath.Join(libDir, "tmp")
	wdav, err := wdav.NewWebDavHandler(tmpdir, backgroundImport(executor))
	if err != nil {
		logger.Fatal("Error initializing webdav interface", zap.Error(err))
	}
	router.PathPrefix("/dav/").Handler(wdav)
	if consts.IsDevMode() && uiDir != "" {
		router.PathPrefix("/").Handler(http.FileServer(http.Dir(uiDir)))
	} else {
		router.PathPrefix("/").Handler(rest.Embedder())
	}

	if ifs, err := net.Interfaces(); err == nil {
		for _, intf := range ifs {
			if addr, err := intf.Addrs(); err == nil {
				for _, a := range addr {
					ip, _, _ := net.ParseCIDR(a.String())
					if ip.IsLoopback() || !ip.IsGlobalUnicast() {
						continue
					}
					logger.Info("Address", zap.String("if", intf.Name),
						zap.String("net", a.Network()),
						zap.String("addr", a.String()),
						zap.Bool("loopback", ip.IsLoopback()),
						zap.Bool("global", ip.IsGlobalUnicast()))
				}
			}
		}
	}
	server := http.Server{
		Addr:        fmt.Sprintf(":%d", port),
		Handler:     rest.WithMiddleWares(router, "rest"),
		BaseContext: func(l net.Listener) context.Context { return ctx },
	}
	go func() {
		logger.Info("Starting HTTP server...", zap.Uint("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
		logger.Info("HTTP server stopped")
	}()

	<-ctx.Done()

	logger.Info("Stopping server...")

	peers.Shutdown()

	ctxShutdown, cancelServerShutdown := context.WithTimeout(ctx, 5*time.Second)
	defer func() {
		cancelServerShutdown()
	}()
	if err := server.Shutdown(ctxShutdown); err != nil {
		logger.Error("Failed to shutdown HTTP server", zap.Error(err))
	}

	logger.Info("Terminated gracefully")
}

func launchStartupTasks(ctx context.Context, tasksRepo *tasks.TaskRepository, executor tasks.TaskExecutor) {
	for _, t := range tasksRepo.DefinedTasks() {
		if t.RunOnStart {
			logging.From(ctx).Debug("Launching startup task", zap.String("task", t.Name))
			task, err := tasksRepo.CreateTask(t.Name)
			if err != nil {
				logging.From(ctx).Warn("StartupTasks", zap.Error(err))
				continue
			}
			executor.Submit(ctx, task)
		}
	}
}

func backgroundImport(executor tasks.TaskExecutor) wdav.UploadedFunc {
	return func(ctx context.Context, path string) {
		task := importer.NewImportFileTaskWithParams(false, path, true)
		if _, err := executor.Submit(ctx, task); err != nil {
			logging.From(ctx).Warn("Could not import file", zap.String("path", path), zap.Error(err))
		}
	}
}

func addRemoteThumber(self string, thumbers *domain.Thumbers) swarm.PeerHandler {
	return func(ctx context.Context, peer swarm.Peer) {
		if self == peer.ID {
			// Do not add ourselves as remote thumber
			return
		}
		thumber, err := swarm.NewRemoteThumber(fmt.Sprintf("%s/thumb", peer.URL), domain.JPEG)
		if err != nil {
			logging.From(ctx).Warn("Failed to create remote thumber", zap.Error(err))
			return
		}
		logging.From(ctx).Info("Remote thumber added", zap.String("peer.url", peer.URL))
		thumbers.Add(thumber, 1)
	}
}
