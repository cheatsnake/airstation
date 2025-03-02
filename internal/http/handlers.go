package http

import (
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/cheatsnake/airstation/internal/track"
	trackservice "github.com/cheatsnake/airstation/internal/track/service"
	"github.com/golang-jwt/jwt/v5"
)

const chunkLimit = 64 * 1024 * 1024 // 64 MB

func (s *Server) handleHLSPlaylist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "audio/mpegurl")

	playlist := s.state.GenerateHLSPlaylist()
	fmt.Fprint(w, playlist)
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

func (s *Server) handleTrackUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(chunkLimit)

	file, handler, err := r.FormFile("track")
	if err != nil {
		s.logger.Debug(err.Error())
		jsonBadRequest(w, "Track parsing failed: "+err.Error())
		return
	}
	defer file.Close()

	trackPath := path.Join(s.config.TracksDir, handler.Filename)
	dst, err := os.Create(trackPath)
	if err != nil {
		s.logger.Debug(err.Error())
		jsonInternalError(w, "Track saving to disk failed")
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		s.logger.Debug(err.Error())
		jsonInternalError(w, "Track upload failed")
		return
	}

	track, err := s.trackService.AddTrack(handler.Filename, trackPath)
	if err != nil {
		s.logger.Debug(err.Error())
		jsonBadRequest(w, "Track saving to database failed")
		return
	}

	jsonResponse(w, track)
}

func (s *Server) handleTracksUpload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(chunkLimit)
	if err != nil {
		s.logger.Debug(err.Error())
		jsonBadRequest(w, "Failed to parse multipart form: "+err.Error())
		return
	}

	files := r.MultipartForm.File["tracks"]
	if len(files) == 0 {
		jsonBadRequest(w, "No files uploaded")
		return
	}

	var uploadedTracks []*track.Track

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			s.logger.Debug(err.Error())
			jsonBadRequest(w, "Failed to open file: "+err.Error())
			return
		}
		defer file.Close()

		trackPath := path.Join(s.config.TracksDir, fileHeader.Filename)
		dst, err := os.Create(trackPath)
		if err != nil {
			s.logger.Debug(err.Error())
			jsonInternalError(w, "Failed to create file on disk: "+err.Error())
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			s.logger.Debug(err.Error())
			jsonInternalError(w, "Failed to save file: "+err.Error())
			return
		}

		track, err := s.trackService.AddTrack(fileHeader.Filename, trackPath)
		if err != nil {
			s.logger.Debug(err.Error())
			jsonBadRequest(w, "Failed to save track to database: "+err.Error())
			return
		}

		uploadedTracks = append(uploadedTracks, track)
	}

	jsonResponse(w, uploadedTracks)
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

	current := s.state.CurrentTrack
	hasCurrent := slices.Contains(ids.IDs, current.ID)
	if hasCurrent {
		jsonBadRequest(w, "Can't delete a track that is being played")
		return
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

func (s *Server) handleStaticDir(prefix string, path string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.Dir(path)))
}
