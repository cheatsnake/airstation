package http

import (
	"log/slog"
	"mime"
	"net/http"

	"github.com/cheatsnake/airstation/internal/config"
	"github.com/cheatsnake/airstation/internal/hls"
	"github.com/cheatsnake/airstation/internal/playback"
	trackservice "github.com/cheatsnake/airstation/internal/track/service"
	"github.com/rs/cors"
)

type Server struct {
	state        *playback.State
	trackService *trackservice.Service
	config       *config.Config
	logger       *slog.Logger
	mux          *http.ServeMux
}

func NewServer(state *playback.State, trackService *trackservice.Service, conf *config.Config, logger *slog.Logger) *Server {
	return &Server{
		state:        state,
		trackService: trackService,
		config:       conf,
		logger:       logger,
		mux:          http.NewServeMux(),
	}
}

func (s *Server) Run() {
	s.registerMP2TMimeType()

	// Public handlers
	s.mux.HandleFunc("GET /stream", s.handleHLSPlaylist)
	s.mux.HandleFunc("POST /v1/api/login", s.handleLogin)

	// Admin handlers
	s.mux.Handle("GET /v1/api/playback", s.jwtAuth(http.HandlerFunc(s.handlePlaybackState)))
	s.mux.Handle("POST /v1/api/track", s.jwtAuth(http.HandlerFunc(s.handleTrackUpload)))
	s.mux.Handle("POST /v1/api/tracks", s.jwtAuth(http.HandlerFunc(s.handleTracksUpload)))
	s.mux.Handle("GET /v1/api/tracks", s.jwtAuth(http.HandlerFunc(s.handleTracks)))
	s.mux.Handle("DELETE /v1/api/tracks", s.jwtAuth(http.HandlerFunc(s.handleDeleteTracks)))

	s.mux.Handle("GET /v1/api/queue", s.jwtAuth(http.HandlerFunc(s.handleQueue)))
	s.mux.Handle("POST /v1/api/queue", s.jwtAuth(http.HandlerFunc(s.handleAddToQueue)))
	s.mux.Handle("PUT /v1/api/queue", s.jwtAuth(http.HandlerFunc(s.handleReorderQueue)))
	s.mux.Handle("DELETE /v1/api/queue", s.jwtAuth(http.HandlerFunc(s.handleRemoveFromQueue)))

	// Static
	s.mux.Handle("GET /static/tmp/", s.handleStaticDir("/static/tmp", s.config.TmpDir))
	s.mux.Handle("GET /static/tracks/", s.jwtAuth(s.handleStaticDir("/static/tracks", s.config.TracksDir)))
	s.mux.Handle("GET /", s.handleStaticDir("", s.config.WebDir))

	server := cors.Default().Handler(s.mux) // CORS middleware

	s.logger.Info("Server starts on http://localhost:" + s.config.HTTPPort)
	err := http.ListenAndServe(":"+s.config.HTTPPort, server)
	if err != nil {
		s.logger.Error("Listend and serve failed", slog.String("info", err.Error()))
	}
}

func (s *Server) registerMP2TMimeType() {
	err := mime.AddExtensionType(hls.SegmentExtension, "video/mp2t")
	if err != nil {
		s.logger.Error("MP2T mime type registration failed", slog.String("info", err.Error()))
	}
}
