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
	router        *http.ServeMux
}

func NewServer(state *playback.State, trackService *trackservice.Service, conf *config.Config, logger *slog.Logger) *Server {
	return &Server{
		state:         state,
		eventsEmitter: events.NewEmitter(),
		trackService:  trackService,
		config:        conf,
		logger:        logger,
		router:        http.NewServeMux(),
	}
}

func (s *Server) Run() {
	s.registerMP2TMimeType()

	// Public handlers
	s.router.HandleFunc("GET /stream", s.handleHLSPlaylist)
	s.router.HandleFunc("GET /api/v1/events", s.handleEvents)
	s.router.HandleFunc("POST /api/v1/login", s.handleLogin)
	s.router.Handle("GET /static/tmp/", s.handleStaticDirWithoutCache("/static/tmp", s.config.TmpDir))
	s.router.Handle("GET /api/v1/playback", http.HandlerFunc(s.handlePlaybackState))

	// Protected handlers
	s.router.Handle("POST /api/v1/tracks", s.jwtAuth(http.HandlerFunc(s.handleTracksUpload)))
	s.router.Handle("GET /api/v1/tracks", s.jwtAuth(http.HandlerFunc(s.handleTracks)))
	s.router.Handle("DELETE /api/v1/tracks", s.jwtAuth(http.HandlerFunc(s.handleDeleteTracks)))
	s.router.Handle("GET /api/v1/queue", s.jwtAuth(http.HandlerFunc(s.handleQueue)))
	s.router.Handle("POST /api/v1/queue", s.jwtAuth(http.HandlerFunc(s.handleAddToQueue)))
	s.router.Handle("PUT /api/v1/queue", s.jwtAuth(http.HandlerFunc(s.handleReorderQueue)))
	s.router.Handle("DELETE /api/v1/queue", s.jwtAuth(http.HandlerFunc(s.handleRemoveFromQueue)))
	s.router.Handle("POST /api/v1/playback/pause", s.jwtAuth(http.HandlerFunc(s.handlePausePlayback)))
	s.router.Handle("POST /api/v1/playback/play", s.jwtAuth(http.HandlerFunc(s.handlePlayPlayback)))
	s.router.Handle("GET /static/tracks/", s.jwtAuth(s.handleStaticDir("/static/tracks", s.config.TracksDir)))

	s.router.Handle("GET /studio/", s.handleStaticDir("/studio/", s.config.StudioDir))
	s.router.Handle("GET /", s.handleStaticDir("/", s.config.PlayerDir))

	s.listenEvents()

	err := s.state.Play()
	if err != nil {
		s.logger.Warn("Auto start playing failed: " + err.Error())
	}

	go s.state.Run()
	go s.trackService.LoadTracksFromDisk(s.config.TracksDir)

	s.logger.Info("Server starts on http://localhost:" + s.config.HTTPPort)
	err = http.ListenAndServe(":"+s.config.HTTPPort, cors.Default().Handler(s.router))
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
			case <-s.state.PlayNotify:
				s.eventsEmitter.RegisterEvent(eventPlay, s.state.CurrentTrack.Name)
			case <-s.state.PauseNotify:
				s.eventsEmitter.RegisterEvent(eventPause, " ")
			case trackName := <-s.state.NewTrackNotify:
				s.eventsEmitter.RegisterEvent(eventNewTrack, trackName)
			case loadedTracks := <-s.trackService.LoadedTracksNotify:
				s.eventsEmitter.RegisterEvent(eventLoadedTracks, strconv.Itoa(loadedTracks))
			}
		}
	}()
}
