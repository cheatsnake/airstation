package main

import (
	"log/slog"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/cheatsnake/airstation/internal/config"
	"github.com/cheatsnake/airstation/internal/http"
	"github.com/cheatsnake/airstation/internal/logger"
	"github.com/cheatsnake/airstation/internal/pkg/fs"
	"github.com/cheatsnake/airstation/internal/storage"
	"github.com/cheatsnake/airstation/internal/storage/sqlite"
)

func main() {
	conf := config.Load()

	fs.DeleteDirIfExists(conf.TmpDir)
	fs.MustDir(conf.TmpDir)
	fs.MustDir(conf.TracksDir)
	fs.MustDir(conf.DBDir)

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	log := logger.New()
	store, err := sqlite.New(path.Join(conf.DBDir, conf.DBFile), log.WithGroup("storage"))
	if err != nil {
		log.Error("Failed connect to database: " + err.Error())
		os.Exit(1)
	}

	httpServer := http.NewServer(store, conf, log)
	go httpServer.Run()

	<-stopSignal
	shutdown(log, store)
}

func shutdown(log *slog.Logger, store storage.Storage) {
	println()
	log.Info("Shutting down the app...")

	err := store.Close()
	if err != nil {
		log.Error("Failed to close database connection: " + err.Error())
	}

	log.Info("App gracefully stopped")
}
