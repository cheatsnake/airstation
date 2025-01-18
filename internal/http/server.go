package http

import (
	"log"
	"mime"
	"net/http"

	"github.com/cheatsnake/airstation/internal/config"
	"github.com/cheatsnake/airstation/internal/hls"
	"github.com/cheatsnake/airstation/internal/playback"
)

func StartServer(state *playback.State) {
	conf := config.GetConfig()
	addMP2TMimeType()

	http.HandleFunc("/stream", handleHLSPlaylist(state))
	http.Handle("/static/tmp/", handleStaticDir("/static/tmp", conf.TmpDir))
	http.Handle("/", handleStaticDir("", conf.WebDir))

	log.Println("Server starts on http://localhost:" + conf.HTTPPort)
	log.Fatal(http.ListenAndServe(":"+conf.HTTPPort, nil))
}

func addMP2TMimeType() {
	err := mime.AddExtensionType(hls.SegmentExtension, "video/mp2t")
	if err != nil {
		log.Fatalf("Failed to add MIME type: %s", err)
	}
}
