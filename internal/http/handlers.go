package http

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/cheatsnake/airstation/internal/live"
	"github.com/cheatsnake/airstation/internal/pkg/sse"
	"github.com/cheatsnake/airstation/internal/station"
	"github.com/cheatsnake/airstation/internal/track"
	"github.com/cheatsnake/airstation/internal/voiceover"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

const multipartChunkLimit = 64 * 1024 * 1024 // 64 MB
const copyBufferSize = 256 * 1024            // 256 KB

// liveUpgrader keeps the default same-origin check. The endpoint is also behind
// jwtAuth, so a cross-site page cannot start a show with the studio cookie.
var liveUpgrader = websocket.Upgrader{
	ReadBufferSize:  8 * 1024,
	WriteBufferSize: 4 * 1024,
}

func (s *Server) handleHLSPlaylist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "audio/mpegurl")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	playlist, playing := s.playbackState.PlaylistSnapshot()
	if playing {
		fmt.Fprint(w, playlist)
	}
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	eventChan := make(chan *sse.Event)
	s.eventsEmitter.Subscribe(eventChan)

	closeNotify := r.Context().Done()
	go func() {
		<-closeNotify
		s.eventsEmitter.Unsubscribe(eventChan)
		close(eventChan)
	}()

	// Send current number of listeners immediately
	countEvent := s.countListeners()
	fmt.Fprint(w, countEvent.Stringify())
	w.(http.Flusher).Flush()

	for {
		event, isOpen := <-eventChan
		if !isOpen {
			break
		}

		fmt.Fprint(w, event.Stringify())
		w.(http.Flusher).Flush()
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	body, err := parseJSONBody[struct {
		Secret string `json:"secret"`
	}](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed.")
		return
	}

	isValidSecret := subtle.ConstantTimeCompare([]byte(body.Secret), []byte(s.config.SecretKey)) == 1
	if !isValidSecret {
		jsonForbidden(w, "Wrong secret, access denied.")
		return
	}

	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := jwt.MapClaims{
		"iss": "airstation",
		"exp": expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWTSign))
	if err != nil {
		s.logger.Debug("Failed to generate token: " + err.Error())
		jsonInternalError(w, "Failed to generate token.")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Expires:  expirationTime,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.config.SecureCookie,
		SameSite: http.SameSiteStrictMode,
	})

	s.logger.Info(fmt.Sprintf("New login succeed from %s with secureCookie=%v", r.Host, s.config.SecureCookie))

	jsonOK(w, "Login succeed.")
}

func (s *Server) handleTracks(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	page := parseIntQuery(queries, "page", 1)
	limit := parseIntQuery(queries, "limit", 20)
	search := queries.Get("search")
	sortBy := queries.Get("sort_by")
	sortOrder := queries.Get("sort_order")
	kind := queries.Get("kind")

	result, err := s.trackService.Tracks(page, limit, search, sortBy, sortOrder, kind)
	if err != nil {
		jsonBadRequest(w, "Tracks retrieving failed: "+err.Error())
		return
	}

	jsonResponse(w, result)
}

func (s *Server) handleTracksUpload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(multipartChunkLimit)
	if err != nil {
		jsonBadRequest(w, "Failed to parse multipart form: "+err.Error())
		return
	}

	files := r.MultipartForm.File["tracks"]
	if len(files) == 0 {
		jsonBadRequest(w, "No files uploaded")
		return
	}

	kind := r.FormValue("kind")
	if kind == "" {
		kind = track.KindMusic
	}
	if !track.IsValidKind(kind) {
		jsonBadRequest(w, "Unsupported asset kind: "+kind)
		return
	}

	paths := make([]string, 0, len(files))
	for _, fileHeader := range files {
		path, err := s.saveFile(fileHeader)
		if err != nil {
			jsonBadRequest(w, err.Error())
			return
		}
		paths = append(paths, path)
	}

	tracks, errs := s.trackService.LoadUploadedTracks(paths, kind)
	if len(tracks) == 0 {
		if len(errs) > 0 {
			jsonBadRequest(w, "Upload rejected: "+errs[0].Error())
			return
		}
		jsonBadRequest(w, "No assets were processed")
		return
	}

	msg := fmt.Sprintf("%d %s asset(s) added to the library.", len(tracks), kind)
	if len(errs) > 0 {
		msg += fmt.Sprintf(" %d file(s) rejected: %s", len(errs), errs[0].Error())
	}
	jsonOK(w, msg)
}

func (s *Server) handleDeleteTracks(w http.ResponseWriter, r *http.Request) {
	body, err := parseJSONBody[track.BodyWithIDs](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed: "+err.Error())
		return
	}

	err = s.trackService.DeleteTracks(body.IDs)
	if err != nil {
		s.logger.Debug(err.Error())
		jsonBadRequest(w, "Deleting tracks failed")
		return
	}

	jsonOK(w, "Tracks deleted")
}

func (s *Server) handleQueue(w http.ResponseWriter, _ *http.Request) {
	queue, err := s.queueService.Queue()
	if err != nil {
		s.logger.Debug(err.Error())
		jsonBadRequest(w, "Queue retrieving failed: "+err.Error())
		return
	}

	jsonResponse(w, queue)
}

func (s *Server) handleAddToQueue(w http.ResponseWriter, r *http.Request) {
	body, err := parseJSONBody[track.BodyWithIDs](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed: "+err.Error())
		return
	}

	tracks, err := s.trackService.FindTracks(body.IDs)
	if err != nil {
		jsonBadRequest(w, "Adding tracks to queue failed: "+err.Error())
		return
	}

	err = s.queueService.AddToQueue(tracks)
	if err != nil {
		jsonBadRequest(w, "Adding tracks to queue failed: "+err.Error())
		return
	}

	err = s.playbackState.Reload()
	if err != nil {
		s.logger.Debug("Playback reload failed: " + err.Error())
	}

	jsonOK(w, "Tracks added")
}

func (s *Server) handleReorderQueue(w http.ResponseWriter, r *http.Request) {
	body, err := parseJSONBody[track.BodyWithIDs](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed: "+err.Error())
		return
	}

	err = s.queueService.ReorderQueue(body.IDs)
	if err != nil {
		jsonBadRequest(w, "Queue reordering failed: "+err.Error())
		return
	}

	err = s.playbackState.Reload()
	if err != nil {
		s.logger.Debug("Playback reload failed: " + err.Error())
	}

	jsonOK(w, "Queue reordered")
}

func (s *Server) handleRemoveFromQueue(w http.ResponseWriter, r *http.Request) {
	body, err := parseJSONBody[track.BodyWithIDs](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed: "+err.Error())
		return
	}

	if s.playbackState.CurrentTrack != nil {
		hasCurrent := slices.Contains(body.IDs, s.playbackState.CurrentTrack.ID)
		if hasCurrent {
			s.playbackState.Pause()
		}
	}

	err = s.queueService.RemoveFromQueue(body.IDs)
	if err != nil {
		jsonBadRequest(w, "Removing from queue failed: "+err.Error())
		return
	}

	err = s.playbackState.Reload()
	if err != nil {
		s.logger.Debug("Playback reload failed: " + err.Error())
	}

	jsonOK(w, "Tracks removed")
}

func (s *Server) handlePlaybackState(w http.ResponseWriter, _ *http.Request) {
	jsonResponse(w, s.playbackState)
}

func (s *Server) handlePausePlayback(w http.ResponseWriter, _ *http.Request) {
	s.playbackState.Pause()
	jsonResponse(w, s.playbackState)
}

func (s *Server) handlePlayPlayback(w http.ResponseWriter, _ *http.Request) {
	err := s.playbackState.Play()
	if err != nil {
		jsonBadRequest(w, "Playback failed to start: "+err.Error())
		return
	}

	jsonResponse(w, s.playbackState)
}

func (s *Server) handlePlaybackHistory(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	limit := parseIntQuery(queries, "limit", 50)
	history, err := s.playbackService.RecentPlaybackHistory(limit)
	if err != nil {
		s.logger.Debug(err.Error())
		jsonBadRequest(w, "Playback history retrieving failed")
		return
	}

	jsonResponse(w, history)
}

func (s *Server) handleCarts(w http.ResponseWriter, _ *http.Request) {
	carts, err := s.trackService.Carts()
	if err != nil {
		jsonBadRequest(w, "Carts retrieving failed: "+err.Error())
		return
	}

	jsonResponse(w, carts)
}

func (s *Server) handlePlayCart(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	body, err := parseJSONBody[struct {
		Mode string `json:"mode"`
	}](r)
	if err != nil {
		// A body is optional: the default is "play next".
		body = &struct {
			Mode string `json:"mode"`
		}{}
	}

	mode := body.Mode
	if mode == "" {
		mode = "next"
	}
	if mode != "now" && mode != "next" {
		jsonBadRequest(w, `mode must be "now" or "next"`)
		return
	}

	tracks, err := s.trackService.FindTracks([]string{id})
	if err != nil || len(tracks) == 0 {
		jsonBadRequest(w, "Cart not found")
		return
	}

	if err := s.enqueueTrack(id, mode == "now"); err != nil {
		jsonBadRequest(w, "Failed to queue cart: "+err.Error())
		return
	}

	jsonOK(w, "Cart queued")
}

// enqueueTrack inserts an asset at the front of the queue ("now") or directly
// after the current track ("next"), then asks the playback engine to reload so
// the upcoming HLS segments reflect the change.
func (s *Server) enqueueTrack(trackID string, playNow bool) error {
	queue, err := s.queueService.Queue()
	if err != nil {
		return err
	}

	ids := make([]string, 0, len(queue)+1)
	if playNow || !s.playbackState.IsPlaying || len(queue) == 0 {
		ids = append(ids, trackID)
	} else {
		ids = append(ids, queue[0].ID)
		if queue[0].ID != trackID {
			ids = append(ids, trackID)
		}
		queue = queue[1:]
	}

	for _, t := range queue {
		if t.ID == trackID {
			continue
		}
		ids = append(ids, t.ID)
	}

	if err := s.queueService.ReorderQueue(ids); err != nil {
		return err
	}

	return s.playbackState.Reload()
}

func (s *Server) handleVoiceOvers(w http.ResponseWriter, _ *http.Request) {
	list, err := s.voiceoverService.List()
	if err != nil {
		jsonBadRequest(w, "Voice-overs retrieving failed: "+err.Error())
		return
	}

	jsonResponse(w, list)
}

func (s *Server) handleCreateVoiceOver(w http.ResponseWriter, r *http.Request) {
	body, err := parseJSONBody[struct {
		VoiceTrackID string  `json:"voiceTrackId"`
		Target       string  `json:"target"`
		FadeIn       float64 `json:"fadeIn"`
		FadeOut      float64 `json:"fadeOut"`
		DuckDB       float64 `json:"duckDb"`
		StartOffset  float64 `json:"startOffset"`
	}](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed.")
		return
	}

	currentTrackID := ""
	if body.Target == voiceover.TargetCurrent {
		currentTrackID = s.playbackState.CurrentTrackID()
	}

	vo, err := s.voiceoverService.Create(
		body.VoiceTrackID,
		body.Target,
		body.FadeIn,
		body.FadeOut,
		body.DuckDB,
		body.StartOffset,
		currentTrackID,
	)
	if err != nil {
		jsonBadRequest(w, "Voice-over creation failed: "+err.Error())
		return
	}

	if body.Target == voiceover.TargetCurrent {
		if err := s.playbackState.Restart(); err != nil {
			_ = s.voiceoverService.Cancel(vo.ID)
			jsonBadRequest(w, "Voice-over playback failed: "+err.Error())
			return
		}
	} else if err := s.playbackState.Reload(); err != nil {
		s.logger.Warn("Playback reload after voice-over failed: " + err.Error())
	}

	s.eventsEmitter.RegisterEvent(eventVoiceOver, vo.VoiceName)
	jsonResponse(w, vo)
}

func (s *Server) handleCancelVoiceOver(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := s.voiceoverService.Cancel(id); err != nil {
		jsonBadRequest(w, "Voice-over cancellation failed: "+err.Error())
		return
	}

	if err := s.playbackState.Reload(); err != nil {
		s.logger.Warn("Playback reload after voice-over cancel failed: " + err.Error())
	}

	jsonOK(w, "Voice-over cancelled")
}

func (s *Server) handleScheduleRules(w http.ResponseWriter, _ *http.Request) {
	rules, err := s.scheduleService.Rules()
	if err != nil {
		jsonBadRequest(w, "Schedule retrieving failed: "+err.Error())
		return
	}

	jsonResponse(w, rules)
}

func (s *Server) handleCreateScheduleRule(w http.ResponseWriter, r *http.Request) {
	body, err := parseJSONBody[struct {
		Name         string `json:"name"`
		Kind         string `json:"kind"`
		EveryNTracks int    `json:"everyNTracks"`
		AtMinute     int    `json:"atMinute"`
		Enabled      *bool  `json:"enabled"`
	}](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed.")
		return
	}

	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	rule, err := s.scheduleService.AddRule(body.Name, body.Kind, body.EveryNTracks, body.AtMinute, enabled)
	if err != nil {
		jsonBadRequest(w, "Schedule rule creation failed: "+err.Error())
		return
	}

	jsonResponse(w, rule)
}

func (s *Server) handleEditScheduleRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	body, err := parseJSONBody[struct {
		Name         string `json:"name"`
		Kind         string `json:"kind"`
		EveryNTracks int    `json:"everyNTracks"`
		AtMinute     int    `json:"atMinute"`
		Enabled      *bool  `json:"enabled"`
	}](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed.")
		return
	}

	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	rule, err := s.scheduleService.EditRule(id, body.Name, body.Kind, body.EveryNTracks, body.AtMinute, enabled)
	if err != nil {
		jsonBadRequest(w, "Schedule rule update failed: "+err.Error())
		return
	}

	jsonResponse(w, rule)
}

func (s *Server) handleDeleteScheduleRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := s.scheduleService.DeleteRule(id); err != nil {
		jsonBadRequest(w, "Schedule rule deletion failed: "+err.Error())
		return
	}

	jsonOK(w, "Schedule rule deleted")
}

func (s *Server) handleAddPlaylist(w http.ResponseWriter, r *http.Request) {
	body, err := parseJSONBody[struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		TrackIDs    []string `json:"trackIDs"`
	}](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed: "+err.Error())
		return
	}

	pl, err := s.playlistService.AddPlaylist(body.Name, body.Description, body.TrackIDs)
	if err != nil {
		jsonBadRequest(w, "Playlist creation failed: "+err.Error())
		return
	}

	jsonResponse(w, pl)
}

func (s *Server) handlePlaylists(w http.ResponseWriter, r *http.Request) {
	pls, err := s.playlistService.Playlists()
	if err != nil {
		jsonBadRequest(w, "Playlists retrieving failed: "+err.Error())
		return
	}

	jsonResponse(w, pls)
}

func (s *Server) handlePlaylist(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	pl, err := s.playlistService.Playlist(id)
	if err != nil {
		jsonBadRequest(w, "Playlist retrieving failed: "+err.Error())
		return
	}

	jsonResponse(w, pl)
}

func (s *Server) handleEditPlaylist(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	body, err := parseJSONBody[struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		TrackIDs    []string `json:"trackIDs"`
	}](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed: "+err.Error())
		return
	}

	err = s.playlistService.EditPlaylist(id, body.Name, body.Description, body.TrackIDs)
	if err != nil {
		jsonBadRequest(w, "Playlist creation failed: "+err.Error())
		return
	}

	jsonOK(w, "Playlist updated")
}

func (s *Server) handleDeletePlaylist(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := s.playlistService.DeletePlaylist(id)
	if err != nil {
		jsonBadRequest(w, "Playlist deletion failed: "+err.Error())
		return
	}

	jsonOK(w, "Playlist deleted")
}

func (s *Server) handleStaticDir(prefix string, path string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.Dir(path)))
}

func (s *Server) handleStaticDirWithoutCache(prefix string, path string) http.Handler {
	fileHandler := http.StripPrefix(prefix, http.FileServer(http.Dir(path)))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		fileHandler.ServeHTTP(w, r)
	})
}

func (s *Server) handleStationInfo(w http.ResponseWriter, _ *http.Request) {
	info, err := s.stationService.Info()
	if err != nil {
		jsonBadRequest(w, "Failed to get station info: "+err.Error())
		return
	}

	jsonResponse(w, info)
}

func (s *Server) handleEditStationInfo(w http.ResponseWriter, r *http.Request) {
	body, err := parseJSONBody[station.Info](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed: "+err.Error())
		return
	}

	info, err := s.stationService.EditInfo(body)
	if err != nil {
		jsonBadRequest(w, "Station info editing failed: "+err.Error())
		return
	}

	s.eventsEmitter.RegisterEvent(eventChangeTheme, " ")

	jsonResponse(w, info)
}

func (s *Server) saveFile(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		msg := "Failed to open file: " + err.Error()
		s.logger.Debug(msg)
		return "", errors.New(msg)
	}

	fileName := filepath.Base(fileHeader.Filename)
	filePath := filepath.Join(s.config.TracksDir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		msg := "Failed to create file on disk: " + err.Error()
		s.logger.Debug(msg)
		return "", errors.New(msg)
	}

	_, err = io.CopyBuffer(dst, file, make([]byte, copyBufferSize))
	if err != nil {
		msg := "Failed to save file: " + err.Error()
		s.logger.Debug(msg)
		return "", errors.New(msg)
	}

	file.Close()
	dst.Close()

	return filePath, nil
}

func (s *Server) handleLiveStatus(w http.ResponseWriter, _ *http.Request) {
	jsonResponse(w, map[string]any{
		"live":         s.playbackState.IsLive(),
		"mode":         s.playbackState.LiveManagerMode(),
		"startedAt":    s.playbackState.LiveStartedUnix(),
		"segmentCount": s.playbackState.LiveSegmentCount(),
	})
}

// handleLiveStart starts an external relay source. Microphone sessions use the
// streaming ingest endpoint instead, because the station is live for exactly as
// long as that request body stays open.
func (s *Server) handleLiveStart(w http.ResponseWriter, r *http.Request) {
	body, err := parseJSONBody[struct {
		Mode string `json:"mode"`
		URL  string `json:"url"`
	}](r)
	if err != nil {
		jsonBadRequest(w, "Failed to parse request body: "+err.Error())
		return
	}
	if body == nil || body.URL == "" {
		jsonBadRequest(w, "A stream URL is required")
		return
	}
	if body.Mode != "" && body.Mode != "external" {
		jsonBadRequest(w, "Microphone sessions use POST /api/v1/live/ingest")
		return
	}
	if err := s.playbackState.StartLiveExternal(body.URL); err != nil {
		if errors.Is(err, live.ErrAlreadyLive) {
			jsonMessage(w, http.StatusConflict, err.Error())
			return
		}
		jsonBadRequest(w, "Failed to start live: "+err.Error())
		return
	}
	jsonOK(w, "Live source started")
}

func (s *Server) handleLiveStop(w http.ResponseWriter, _ *http.Request) {
	if err := s.playbackState.StopLive(); err != nil {
		jsonBadRequest(w, "Failed to stop live: "+err.Error())
		return
	}
	jsonOK(w, "Live stopped")
}

// handleLiveIngest accepts microphone audio either as a WebSocket (used by the
// studio browser, because fetch request streams require HTTP/2) or as a plain
// streaming POST body (useful for scripts and server-side encoders).
func (s *Server) handleLiveIngest(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		s.handleLiveIngestWS(w, r)
		return
	}
	s.handleLiveIngestHTTP(w, r)
}

func (s *Server) handleLiveIngestWS(w http.ResponseWriter, r *http.Request) {
	conn, err := liveUpgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Warn(
			"Live ingest upgrade failed",
			slog.String("info", err.Error()),
			slog.String("origin", r.Header.Get("Origin")),
			slog.String("host", r.Host),
		)
		return
	}
	defer conn.Close()

	if err := s.playbackState.StartLiveIngest(); err != nil {
		code := websocket.CloseInternalServerErr
		if errors.Is(err, live.ErrAlreadyLive) {
			code = websocket.ClosePolicyViolation
		}
		_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, err.Error()), time.Now().Add(time.Second))
		return
	}
	defer func() {
		_ = s.playbackState.StopLive()
	}()

	// Opus chunks from MediaRecorder are tiny; a megabyte is already generous.
	conn.SetReadLimit(1 << 20)

	var total int64
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				s.logger.Warn("Live ingest socket ended", slog.String("info", err.Error()))
			}
			break
		}
		if messageType != websocket.BinaryMessage {
			continue
		}
		n, err := s.playbackState.WriteLiveIngestBytes(data)
		total += int64(n)
		if err != nil {
			s.logger.Warn("Live ingest write failed", slog.String("info", err.Error()), slog.Int64("bytes", total))
			break
		}
	}
	s.logger.Info("Live ingest finished", slog.Int64("bytes", total))
}

// handleLiveIngestHTTP holds the request open for the duration of a microphone
// session. FFmpeg reads the WebM/Opus body directly from the pipe; when the
// client disconnects the session ends and automation resumes.
func (s *Server) handleLiveIngestHTTP(w http.ResponseWriter, r *http.Request) {
	if err := s.playbackState.StartLiveIngest(); err != nil {
		if errors.Is(err, live.ErrAlreadyLive) {
			jsonMessage(w, http.StatusConflict, err.Error())
			return
		}
		jsonBadRequest(w, "Failed to start live: "+err.Error())
		return
	}
	defer func() {
		_ = s.playbackState.StopLive()
	}()

	n, err := s.playbackState.WriteLiveIngest(r.Body)
	if err != nil {
		s.logger.Warn("Live ingest ended with an error", slog.String("info", err.Error()), slog.Int64("bytes", n))
	} else {
		s.logger.Info("Live ingest finished", slog.Int64("bytes", n))
	}
	jsonOK(w, "Live session ended")
}
