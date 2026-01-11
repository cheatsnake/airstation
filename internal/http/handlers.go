package http

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/cheatsnake/airstation/internal/pkg/sse"
	"github.com/cheatsnake/airstation/internal/station"
	"github.com/cheatsnake/airstation/internal/track"
	"github.com/golang-jwt/jwt/v5"
)

const multipartChunkLimit = 64 * 1024 * 1024 // 64 MB
const copyBufferSize = 256 * 1024            // 256 KB

func (s *Server) handleHLSPlaylist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "audio/mpegurl")

	if s.playbackState.IsPlaying {
		fmt.Fprint(w, s.playbackState.PlaylistStr)
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

	result, err := s.trackService.Tracks(page, limit, search, sortBy, sortOrder)
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

	for _, fileHeader := range files {
		_, err := s.saveFile(fileHeader)
		if err != nil {
			jsonBadRequest(w, err.Error())
			return
		}
	}

	go s.trackService.LoadTracksFromDisk(s.config.TracksDir)

	msg := fmt.Sprintf("%d track(s) uploaded successfully. They will be available in your library once processed.", len(files))
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
	}

	jsonResponse(w, pls)
}

func (s *Server) handlePlaylist(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	pl, err := s.playlistService.Playlist(id)
	if err != nil {
		jsonBadRequest(w, "Playlist retrieving failed: "+err.Error())
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
