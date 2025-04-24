package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cheatsnake/airstation/internal/config"
	"github.com/cheatsnake/airstation/internal/ffmpeg"
	"github.com/cheatsnake/airstation/internal/http"
	"github.com/cheatsnake/airstation/internal/logger"
	"github.com/cheatsnake/airstation/internal/playback"
	"github.com/cheatsnake/airstation/internal/storage"
	"github.com/cheatsnake/airstation/internal/storage/sqlite"
	"github.com/cheatsnake/airstation/internal/tools/fs"
	trackservice "github.com/cheatsnake/airstation/internal/track/service"
)

func main() {
	conf := config.Load()

	fs.DeleteDirIfExists(conf.TmpDir)
	fs.MustDir(conf.TmpDir)
	fs.MustDir(conf.TracksDir)

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	log := logger.New()
	store, err := sqlite.Open("storage.db", log.WithGroup("storage"))
	if err != nil {
		log.Error("Failed connect to database: " + err.Error())
		os.Exit(1)
	}

	ffmpegCLI := ffmpeg.NewCLI()
	trackService := trackservice.New(store, ffmpegCLI, log.WithGroup("trackservice"))
	go trackService.LoadTracksFromDisk(conf.TracksDir)

	playbackState := playback.NewState(trackService, conf.TmpDir, log.WithGroup("playback"))
	httpServer := http.NewServer(playbackState, trackService, conf, log.WithGroup("http"))
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
