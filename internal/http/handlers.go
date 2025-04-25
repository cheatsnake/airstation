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
	"strings"
	"time"

	"github.com/cheatsnake/airstation/internal/events"
	trackservice "github.com/cheatsnake/airstation/internal/track/service"
	"github.com/golang-jwt/jwt/v5"
)

const multipartChunkLimit = 64 * 1024 * 1024 // 64 MB
const copyBufferSize = 256 * 1024            // 256 KB

func (s *Server) handleHLSPlaylist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "audio/mpegurl")

	if s.state.IsPlaying {
		fmt.Fprint(w, s.state.PlaylistStr)
	}
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	eventChan := make(chan *events.Event)
	s.eventsEmitter.Subscribe(eventChan)

	closeNotify := r.Context().Done()
	go func() {
		<-closeNotify
		s.eventsEmitter.Unsubscribe(eventChan)
		close(eventChan)
	}()

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

	secureCookie := true
	if strings.HasPrefix(r.Host, "localhost:") {
		secureCookie = false // Allow insecure cookies for local development
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Expires:  expirationTime,
		Path:     "/",
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteStrictMode,
	})

	jsonOK(w, "Login succeed.")
}

func (s *Server) handleTracks(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	page := parseIntQuery(queries, "page", 1)
	limit := parseIntQuery(queries, "limit", 20)
	search := queries.Get("search")

	result, err := s.trackService.Tracks(page, limit, search)
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
	ids, err := parseJSONBody[trackservice.TrackIDs](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed: "+err.Error())
		return
	}

	err = s.trackService.DeleteTracks(ids)
	if err != nil {
		s.logger.Debug(err.Error())
		jsonBadRequest(w, "Deleting tracks failed")
		return
	}

	jsonOK(w, "Tracks deleted")
}

func (s *Server) handleQueue(w http.ResponseWriter, _ *http.Request) {
	queue, err := s.trackService.Queue()
	if err != nil {
		s.logger.Debug(err.Error())
		jsonBadRequest(w, "Queue retrieving failed")
		return
	}

	jsonResponse(w, queue)
}

func (s *Server) handleAddToQueue(w http.ResponseWriter, r *http.Request) {
	ids, err := parseJSONBody[trackservice.TrackIDs](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed: "+err.Error())
		return
	}

	tracks, err := s.trackService.FindTracks(ids)
	if err != nil {
		jsonBadRequest(w, "Adding tracks to queue failed: "+err.Error())
		return
	}

	err = s.trackService.AddToQueue(tracks)
	if err != nil {
		jsonBadRequest(w, "Adding tracks to queue failed: "+err.Error())
		return
	}

	err = s.state.Load()
	if err != nil {
		s.logger.Debug(err.Error())
	}

	jsonOK(w, "Tracks added")
}

func (s *Server) handleReorderQueue(w http.ResponseWriter, r *http.Request) {
	ids, err := parseJSONBody[trackservice.TrackIDs](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed: "+err.Error())
		return
	}

	err = s.trackService.ReorderQueue(ids)
	if err != nil {
		jsonBadRequest(w, "Queue reordering failed: "+err.Error())
		return
	}

	err = s.state.Load()
	if err != nil {
		s.logger.Debug(err.Error())
	}

	jsonOK(w, "Queue reordered")
}

func (s *Server) handleRemoveFromQueue(w http.ResponseWriter, r *http.Request) {
	ids, err := parseJSONBody[trackservice.TrackIDs](r)
	if err != nil {
		jsonBadRequest(w, "Parsing request body failed: "+err.Error())
		return
	}

	if s.state.CurrentTrack != nil {
		hasCurrent := slices.Contains(ids.IDs, s.state.CurrentTrack.ID)
		if hasCurrent {
			jsonBadRequest(w, "Can't delete a track that is being played")
			return
		}
	}

	err = s.trackService.RemoveFromQueue(ids)
	if err != nil {
		jsonBadRequest(w, "Removing from queue failed: "+err.Error())
		return
	}

	err = s.state.Load()
	if err != nil {
		s.logger.Debug(err.Error())
	}

	jsonOK(w, "Tracks removed")
}

func (s *Server) handlePlaybackState(w http.ResponseWriter, _ *http.Request) {
	jsonResponse(w, s.state)
}

func (s *Server) handlePausePlayback(w http.ResponseWriter, _ *http.Request) {
	s.state.Pause()
	jsonResponse(w, s.state)
}

func (s *Server) handlePlayPlayback(w http.ResponseWriter, _ *http.Request) {
	err := s.state.Play()
	if err != nil {
		jsonBadRequest(w, "Playback failed to start: "+err.Error())
		return
	}

	jsonResponse(w, s.state)
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
