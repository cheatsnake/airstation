package http

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
)

const chunkLimit = 64 * 1024 * 1024 // 64 MB

func (s *Server) handleHLSPlaylist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "audio/mpegurl")

	playlist := s.state.GenerateHLSPlaylist()
	fmt.Fprint(w, playlist)
}

func (s *Server) handleTracks(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	page := parseIntQuery(queries, "page", 1)
	limit := parseIntQuery(queries, "limit", 20)

	result, err := s.trackService.Tracks(page, limit)
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

func (s *Server) handleDeleteTracks(w http.ResponseWriter, r *http.Request) {
	idsQuery := r.URL.Query().Get("ids")
	ids := strings.Split(idsQuery, ",")
	err := s.trackService.DeleteTracks(ids)
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
	idsQuery := r.URL.Query().Get("ids")
	ids := strings.Split(idsQuery, ",")
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

func (s *Server) handleRemoveFromQueue(w http.ResponseWriter, r *http.Request) {
	idsQuery := r.URL.Query().Get("ids")
	ids := strings.Split(idsQuery, ",")

	current := s.state.CurrentTrack
	hasCurrent := slices.Contains(ids, current.ID)
	if hasCurrent {
		jsonBadRequest(w, "Can't delete a track that is being played")
		return
	}

	err := s.trackService.RemoveFromQueue(ids)
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
