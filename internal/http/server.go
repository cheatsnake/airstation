package http

import (
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/cheatsnake/airstation/internal/config"
	"github.com/cheatsnake/airstation/internal/events"
	"github.com/cheatsnake/airstation/internal/hls"
	"github.com/cheatsnake/airstation/internal/playback"
	trackservice "github.com/cheatsnake/airstation/internal/track/service"
	"github.com/rs/cors"
)

type Server struct {
	state         *playback.State
	eventsEmitter *events.Emitter
	trackService  *trackservice.Service
	config        *config.Config
	logger        *slog.Logger
	mux           *http.ServeMux
}

func NewServer(state *playback.State, trackService *trackservice.Service, conf *config.Config, logger *slog.Logger) *Server {
	return &Server{
		state:         state,
		eventsEmitter: events.NewEmitter(),
		trackService:  trackService,
		config:        conf,
		logger:        logger,
		mux:           http.NewServeMux(),
	}
}

func (s *Server) Run() {
	s.registerMP2TMimeType()

	// Public handlers
	s.mux.HandleFunc("GET /stream", s.handleHLSPlaylist)
	s.mux.HandleFunc("GET /api/v1/events", s.handleEvents)
	s.mux.HandleFunc("POST /api/v1/login", s.handleLogin)
	s.mux.Handle("GET /static/tmp/", s.handleStaticDirWithoutCache("/static/tmp", s.config.TmpDir))
	s.mux.Handle("GET /api/v1/playback", http.HandlerFunc(s.handlePlaybackState))

	// Protected handlers
	s.mux.Handle("POST /api/v1/track", s.jwtAuth(http.HandlerFunc(s.handleTrackUpload)))
	s.mux.Handle("POST /api/v1/tracks", s.jwtAuth(http.HandlerFunc(s.handleTracksUpload)))
	s.mux.Handle("GET /api/v1/tracks", s.jwtAuth(http.HandlerFunc(s.handleTracks)))
	s.mux.Handle("DELETE /api/v1/tracks", s.jwtAuth(http.HandlerFunc(s.handleDeleteTracks)))
	s.mux.Handle("GET /api/v1/queue", s.jwtAuth(http.HandlerFunc(s.handleQueue)))
	s.mux.Handle("POST /api/v1/queue", s.jwtAuth(http.HandlerFunc(s.handleAddToQueue)))
	s.mux.Handle("PUT /api/v1/queue", s.jwtAuth(http.HandlerFunc(s.handleReorderQueue)))
	s.mux.Handle("DELETE /api/v1/queue", s.jwtAuth(http.HandlerFunc(s.handleRemoveFromQueue)))
	s.mux.Handle("POST /api/v1/playback/pause", s.jwtAuth(http.HandlerFunc(s.handlePausePlayback)))
	s.mux.Handle("POST /api/v1/playback/play", s.jwtAuth(http.HandlerFunc(s.handlePlayPlayback)))
	s.mux.Handle("GET /static/tracks/", s.jwtAuth(s.handleStaticDir("/static/tracks", s.config.TracksDir)))

	s.mux.Handle("GET /", s.handleStaticDir("", s.config.WebDir))

	s.listenEvents()
	err := s.state.Play()
	if err != nil {
		s.logger.Warn("Auto start playing failed: " + err.Error())
	}

	go s.state.Run()

	server := cors.Default().Handler(s.mux) // CORS middleware

	s.logger.Info("Server starts on http://localhost:" + s.config.HTTPPort)
	err = http.ListenAndServe(":"+s.config.HTTPPort, server)
	if err != nil {
		s.logger.Error("Listen and serve failed", slog.String("info", err.Error()))
	}
}

func (s *Server) registerMP2TMimeType() {
	err := mime.AddExtensionType(hls.SegmentExtension, "video/mp2t")
	if err != nil {
		s.logger.Error("MP2T mime type registration failed", slog.String("info", err.Error()))
	}
}

func (s *Server) listenEvents() {
	countConnectionTicker := time.Tick(5 * time.Second)

	// TODO: add context for gracefull shutdown

	go func() {
		for range countConnectionTicker {
			count := s.eventsEmitter.CountSubscribers()
			s.eventsEmitter.RegisterEvent(eventCountListeners, strconv.Itoa(count))
		}
	}()

	go func() {
		for {
			select {
			case trackName := <-s.state.NewTrackNotify:
				s.eventsEmitter.RegisterEvent(eventNewTrack, trackName)
			case <-s.state.PlayNotify:
				s.eventsEmitter.RegisterEvent(eventPlay, s.state.CurrentTrack.Name)
			case <-s.state.PauseNotify:
				s.eventsEmitter.RegisterEvent(eventPause, " ")

			}
		}
	}()
}
