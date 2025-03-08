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

	fs.DeleteFolderIfExists(conf.TmpDir)
	fs.MustDir(conf.TmpDir)
	fs.MustDir(conf.TracksDir)

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	log := logger.New()
	storageLog := log.WithGroup("storage")
	store, err := sqlite.Open("storage.db", storageLog)
	if err != nil {
		log.Error("Failed connect to database: " + err.Error())
		os.Exit(1)
	}

	ffmpegCLI := ffmpeg.NewCLI()
	trackServiceLog := log.WithGroup("trackservice")
	trackService := trackservice.New(store, ffmpegCLI, trackServiceLog)

	playbackLog := log.WithGroup("playback")
	state := playback.NewState(trackService, conf.TmpDir, playbackLog)
	err = state.Play()
	if err != nil {
		log.Error("Auto start playing failed: " + err.Error())
	}

	go state.Run()

	httpLog := log.WithGroup("http")
	server := http.NewServer(state, trackService, conf, httpLog)
	go server.Run()

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
