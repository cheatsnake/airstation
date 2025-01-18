package http

import (
	"fmt"
	"net/http"

	"github.com/cheatsnake/airstation/internal/playback"
)

func handleHLSPlaylist(state *playback.State) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/mpegurl")

		playlist := state.GenerateHLSPlaylist()
		fmt.Fprint(w, playlist)
	}
}

func handleStaticDir(prefix string, path string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.Dir(path)))
}
