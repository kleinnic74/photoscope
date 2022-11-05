package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

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
	"github.com/gorilla/mux"
	"github.com/kleinnic74/fflags"
	"go.etcd.io/bbolt"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

type App struct {
	dir string

	db       *bbolt.DB
	bus      *events.Stream
	executor tasks.TaskExecutor
	taskRepo *tasks.TaskRepository
	peers    *swarm.Controller
	router   *mux.Router

	addr string

	shutdownHandlers shutdownHandlers
}

type Options struct {
	LibDir string `json:"libdir"`
	Port   uint   `json:"port"`
}

type shutdownHandler func(context.Context, *App)

const (
	dbName = "photos.db"
)

type shutdownHandlers struct {
	h []shutdownHandler
}

func (hdls *shutdownHandlers) Add(h shutdownHandler) {
	hdls.h = append(hdls.h, h)
}

func (hdls shutdownHandlers) Execute(ctx context.Context, a *App) {
	for i := len(hdls.h) - 1; i >= 0; i-- {
		hdls.h[i](ctx, a)
	}
}

func NewApp(ctx context.Context, o Options) (a *App, err error) {
	logger, ctx := logging.SubFrom(ctx, "app")

	logger.Info("Library directory", zap.String("dir", o.LibDir))
	if err = os.MkdirAll(o.LibDir, os.ModePerm); err != nil {
		return nil, err
	}

	a = &App{
		dir:      o.LibDir,
		addr:     fmt.Sprintf(":%d", o.Port),
		taskRepo: tasks.NewTaskRepository(),
		router:   mux.NewRouter(),
		bus:      events.NewStream(),
	}
	defer func() {
		if err != nil {
			a.shutdownHandlers.Execute(ctx, a)
		}
	}()

	a.db, err = bolt.Open(filepath.Join(o.LibDir, dbName), 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialite data store: %w", err)
	}
	a.shutdownHandlers.Add(func(ctx context.Context, a *App) {
		a.db.Close()
		logging.From(ctx).Info("Closed data store")
	})

	tasks.RegisterTasks(a.taskRepo)
	importer.RegisterTasks(a.taskRepo)

	var migrator *index.MigrationCoordinator
	if migrator, err = index.NewMigrationCoordinator(a.db); err != nil {
		return nil, fmt.Errorf("Failed to initialize migration coordinator: %w", err)
	}

	var indexTracker index.Tracker
	if indexTracker, err = boltstore.NewIndexTracker(a.db); err != nil {
		return nil, fmt.Errorf("Failed to initialize library: %w", err)
	}

	var store library.ClosableStore
	if store, err = boltstore.NewBoltStore(a.db); err != nil {
		return nil, fmt.Errorf("Failed to initialize library: %w", err)
	}

	thumbers := &domain.Thumbers{}
	nbParallelThumbers := domain.CalculateOptimumParallelism()
	thumbers.Add(domain.NewParallelThumber(ctx, domain.LocalThumber{}, nbParallelThumbers), 1)
	logger.Info("Initialized Thumber", zap.Int("parallelism", nbParallelThumbers))

	var lib *library.BasicPhotoLibrary
	if lib, err = library.NewBasicPhotoLibrary(o.LibDir, store, thumbers); err != nil {
		return nil, fmt.Errorf("Failed to initialize library: %w", err)
	}
	logger.Info("Opened photo library", zap.String("path", o.LibDir))
	migrator.AddInstances(lib)
	a.executor = tasks.NewSerialTaskExecutor(lib)
	indexer := index.NewIndexer(indexTracker, a.executor)
	indexer.RegisterTasks(a.taskRepo)
	lib.AddCallback(indexer.Add)

	if err = fflags.IfEnabled(fflags.Define("index.geo"), func() error {
		var geoindex library.GeoIndex
		if geoindex, err = boltstore.NewBoltGeoIndex(a.db); err != nil {
			return err
		}
		migrator.AddStructure("geo", geoindex)

		geocoder := geocoding.NewGeocoder(geoindex, openstreetmap.NewResolver("de,en"))
		geocoder.RegisterTasks(a.taskRepo)
		indexer.RegisterDefered("geo", boltstore.GeoIndexVersion, geocoder.LookupPhotoOnAdd)

		geo := rest.NewGeoHandler(geoindex, lib)
		geo.InitRoutes(a.router)

		geocache := rest.NewGeoCacheHandler(geocoder.Cache)
		geocache.InitRoutes(a.router)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("Failed to initialize geoindex: %w", err)
	}

	var dateindex *boltstore.DateIndex
	if dateindex, err = boltstore.NewDateIndex(a.db); err != nil {
		return nil, fmt.Errorf("Failed to initialize dateindex: %w", err)
	}
	migrator.AddStructure("date", dateindex)
	indexer.RegisterDirect("date", boltstore.DateIndexVersion, dateindex.Add)

	if err = fflags.IfEnabled(fflags.Define("index.events"), func() error {
		eventindex, err := boltstore.NewEventIndex(a.db)
		if err != nil {
			return err
		}
		classification.RegisterTasks(a.taskRepo, eventindex)
		events := rest.NewEventsHandler(eventindex, lib)
		events.InitRoutes(a.router)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("Failed to initialize event database: %w", err)
	}

	registerMigrationTask(a.taskRepo, migrator, indexer)

	var instance *swarm.Instance
	instance, err = swarm.NewInstance(ctx, swarm.InstanceID(lib.ID), DefaultInstanceProperties()...)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize swarm with unique local ID: %w", err)
	}
	logger, ctx = logging.FromWithFields(ctx, zap.Stringer("instance", instance.ID))

	a.peers, err = swarm.NewController(instance, o.Port)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize swarm controller: %w", err)
	}
	a.peers.OnPeerDetected(swarm.SkipSelf(addRemoteThumber(instance.ID, thumbers)))
	a.peers.OnPeerDetected(swarm.SkipSelf(addRemoteSync(a.executor)))

	// REST Handlers

	metrics := rest.NewMetricsHandler()
	metrics.InitRoutes(a.router)

	if consts.IsDevMode() {
		logs := rest.NewLogsHandler()
		logs.InitRoutes(a.router)
		debugService := DebugHandler{}
		debugService.InitRoutes(a.router)
	}

	sse := rest.NewSSEHandler(a.bus)
	sse.InitRoutes(a.router)

	photoApp := rest.NewApp(lib)
	photoApp.InitRoutes(a.router)

	timeline := rest.NewTimelineHandler(dateindex, lib)
	timeline.InitRoutes(a.router)

	tasksApp := rest.NewTaskHandler(a.taskRepo, a.executor)
	tasksApp.InitRoutes(a.router)

	indexesRest := rest.NewIndexes(indexer, migrator)
	indexesRest.Init(a.router)

	peersRest := rest.NewPeersAPI(a.peers)
	peersRest.InitRoutes(a.router)

	thumbService := rest.NewThumberAPI(domain.LocalThumber{})
	thumbService.InitRoutes(a.router)

	tmpdir := filepath.Join(o.LibDir, "tmp")
	var wdh http.Handler
	wdh, err = wdav.NewWebDavHandler(tmpdir, backgroundImport(a.executor))
	if err != nil {
		return nil, fmt.Errorf("Error initializing webdav interface: %w", err)
	}
	a.router.PathPrefix("/dav/").Handler(wdh)
	a.router.PathPrefix("/").Handler(rest.Embedder())

	return a, nil
}

func (a *App) Run(ctx context.Context) {
	logger, ctx := logging.SubFrom(ctx, "app")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		logger, ctx := logging.SubFrom(ctx, "eventbus")
		a.bus.Dispatch(ctx)
		logger.Info("DONE")
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		logger, ctx := logging.SubFrom(ctx, "tasks")
		a.executor.DrainTasks(ctx, func(e tasks.Execution) {
			a.bus.Publish(events.Event{Name: "tasks", Action: "completed"})
		})
		logger.Info("DONE")
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		logger, ctx := logging.SubFrom(ctx, "startuptasks")
		launchStartupTasks(ctx, a.taskRepo, a.executor)
		logger.Info("DONE")
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		logger, ctx := logging.SubFrom(ctx, "swarm")
		a.peers.ListenAndServe(ctx)
		logger.Info("DONE")
		wg.Done()
	}()

	server := http.Server{
		Addr:        a.addr,
		Handler:     rest.WithMiddleWares(a.router, "rest"),
		BaseContext: func(l net.Listener) context.Context { return ctx },
	}
	wg.Add(1)
	go func() {
		logger, _ := logging.SubFrom(ctx, "http")
		logger.Info("Starting HTTP server...", zap.String("bindAddr", a.addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
		logger.Info("DONE")
		wg.Done()
	}()

	<-ctx.Done()

	logger.Info("Stopping...")

	a.peers.Shutdown()

	ctxShutdown, cancelServerShutdown := context.WithTimeout(ctx, 5*time.Second)
	defer func() {
		cancelServerShutdown()
	}()
	if err := server.Shutdown(ctxShutdown); err != nil {
		logger.Error("Failed to shutdown HTTP server", zap.Error(err))
	}

	wg.Wait()

	logger.Info("Terminated gracefully")

}

func backgroundImport(executor tasks.TaskExecutor) wdav.UploadedFunc {
	return func(ctx context.Context, path string) {
		task := importer.NewImportFileTaskWithParams(false, path, true)
		if _, err := executor.Submit(ctx, task); err != nil {
			logging.From(ctx).Warn("Could not import file", zap.String("path", path), zap.Error(err))
		}
	}
}
