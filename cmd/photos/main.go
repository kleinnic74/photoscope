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
	"time"

	"bitbucket.org/kleinnic74/photos/library"
	"bitbucket.org/kleinnic74/photos/library/boltstore"
	"bitbucket.org/kleinnic74/photos/logging"
	"bitbucket.org/kleinnic74/photos/rest"
	"bitbucket.org/kleinnic74/photos/rest/wdav"
	"bitbucket.org/kleinnic74/photos/tasks"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"golang.org/x/net/webdav"
)

var (
	libDir string
	port   uint

	logger *zap.Logger
	ctx    context.Context
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s  [options]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&libDir, "l", "gophotos", "Path to photo library")
	flag.UintVar(&port, "p", 8080, "HTTP server port")
	ctx = logging.Context(context.Background(), nil)
	logger = logging.From(ctx)

	flag.Parse()
}

func main() {
	//	classifier := NewEventClassifier()
	lib, err := library.NewBasicPhotoLibrary(libDir, boltstore.NewBoltStore)
	if err != nil {
		logger.Fatal("Failed to initialize library", zap.NamedError("err", err))
	}
	logger.Info("Opened photo library", zap.String("path", libDir))
	router := mux.NewRouter()
	photoApp := rest.NewApp(lib)
	photoApp.InitRoutes(router)

	executor := tasks.NewSerialTaskExecutor(lib)
	executorContext, cancelExecutor := context.WithCancel(ctx)
	go executor.DrainTasks(executorContext)
	tasksApp := rest.NewTaskHandler(executor)
	tasksApp.InitRoutes(router)

	tmpdir := filepath.Join(libDir, "tmp")
	wdav, err := wdav.NewWebDavAdapter(tmpdir, func(ctx context.Context, path string) {
		task := tasks.NewImportFileTaskWithParams(false, path, true)
		if _, err := executor.Submit(ctx, task); err != nil {
			logging.From(ctx).Warn("Could not import file", zap.String("path", path), zap.Error(err))
		}
	})
	if err != nil {
		logger.Fatal("Failed to launch photos", zap.Error(err))
	}
	dav := &webdav.Handler{
		Prefix:     "/dav/",
		LockSystem: webdav.NewMemLS(),
		FileSystem: wdav,
	}
	router.PathPrefix("/dav/").HandlerFunc(dav.ServeHTTP)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	signalContext, cancel := context.WithCancel(ctx)

	go func() {
		oscall := <-c
		logger.Info("Received signal", zap.Any("signal", oscall))
		cancel()
	}()

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
		Addr:    fmt.Sprintf(":%d", port),
		Handler: rest.WithMiddleWares(router, "rest"),
	}
	go func() {
		logger.Info("Starting HTTP server...", zap.Uint("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
		logger.Info("HTTP server stopped")
	}()

	<-signalContext.Done()

	logger.Info("Stopping server...")

	ctxShutdown, cancelServerShutdown := context.WithTimeout(ctx, 5*time.Second)
	defer func() {
		cancelServerShutdown()
	}()
	if err := server.Shutdown(ctxShutdown); err != nil {
		logger.Fatal(("Failed to shutdown HTTP server"), zap.Error(err))
	}

	cancelExecutor()

	logger.Info("Terminated gracefully")

	// img := classifier.DistanceMatrixToImage()
	// log.Printf("Creating time-distance matrix image %s", matrixFilename)
	// out, err := os.Create(matrixFilename)
	// if err != nil {
	// 	log.Fatalf("Could not create distance matrix: %s", err)
	// }
	// defer out.Close()
	// png.Encode(out, img)
}
