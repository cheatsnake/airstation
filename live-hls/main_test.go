package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// throttledReader paces a file out at roughly `bps` bytes per second, so ffmpeg
// sees something that behaves like a live microphone rather than a file copy.
type throttledReader struct {
	r      io.Reader
	perSec int
	start  time.Time
	sent   int
}

func (t *throttledReader) Read(p []byte) (int, error) {
	if t.start.IsZero() {
		t.start = time.Now()
	}
	// Block until the clock catches up rather than returning (0, nil), which the
	// HTTP transport treats as a stalled body.
	for {
		want := int(time.Since(t.start).Seconds() * float64(t.perSec))
		if allowed := want - t.sent; allowed > 0 {
			if len(p) > allowed {
				p = p[:allowed]
			}
			n, err := t.r.Read(p)
			t.sent += n
			return n, err
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func mustRun(t *testing.T, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %v: %v", name, args, err)
	}
}

func waitFor(t *testing.T, what string, d time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestModeSwitch(t *testing.T) {
	bin := ffmpegBin()
	if _, err := exec.LookPath(bin); err != nil {
		t.Skipf("ffmpeg not found: %v", err)
	}
	os.RemoveAll(hlsDir)
	if err := os.MkdirAll(hlsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// fixtures: an automation bed, and a "microphone" recording to push in
	mustRun(t, bin, "-y", "-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "sine=frequency=220:duration=6",
		"-c:a", "aac", "-b:a", "128k", "-ar", "48000", "-ac", "2", "tone.m4a")

	mustRun(t, bin, "-y", "-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "sine=frequency=880:duration=8",
		"-c:a", "libopus", "-b:a", "64k", "-ar", "48000", "-ac", "1", "mic.webm")

	// automation bed, started directly so the test can kill it
	auto := exec.Command(bin,
		"-hide_banner", "-loglevel", "error",
		"-stream_loop", "-1", "-i", "tone.m4a",
		"-c:a", "aac", "-b:a", "128k", "-ar", "48000", "-ac", "2",
		"-f", "hls", "-hls_time", segSeconds, "-hls_list_size", listSize,
		"-hls_flags", "delete_segments+omit_endlist+independent_segments",
		"-hls_segment_filename", filepath.Join(hlsDir, "auto%05d.ts"),
		filepath.Join(hlsDir, autoM3U8))
	auto.Stderr = os.Stderr
	if err := auto.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { auto.Process.Kill(); auto.Wait() })

	srv := httptest.NewServer(newServer().routes())
	t.Cleanup(srv.Close)

	// 1. automation is on air
	waitFor(t, "automation playlist", 20*time.Second, func() bool {
		_, err := os.Stat(filepath.Join(hlsDir, autoM3U8))
		return err == nil
	})

	resp, err := http.Get(srv.URL + "/api/v1/mode")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), "automation") {
		t.Fatalf("expected automation mode, got %s", body)
	}

	// /stream must serve a usable playlist with absolute segment paths
	pl := getString(t, srv.URL+"/stream")
	if !strings.Contains(pl, "#EXTM3U") || !strings.Contains(pl, "/hls/auto") {
		t.Fatalf("automation playlist looks wrong:\n%s", pl)
	}

	// 2. connect a live source
	f, err := os.Open("mic.webm")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	info, _ := f.Stat()
	rate := int(float64(info.Size()) / 6.0) // spread over ~6s
	if rate < 1000 {
		rate = 1000
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/ingest", &throttledReader{r: f, perSec: rate})
		req.Header.Set("Content-Type", "application/octet-stream")
		r, err := http.DefaultClient.Do(req)
		if err == nil {
			io.Copy(io.Discard, r.Body)
			r.Body.Close()
		}
	}()

	// 3. mode flips to live, and the live playlist is produced
	waitFor(t, "live mode", 25*time.Second, func() bool {
		return strings.Contains(getString(t, srv.URL+"/api/v1/mode"), "live")
	})

	waitFor(t, "live playlist", 30*time.Second, func() bool {
		_, err := os.Stat(filepath.Join(hlsDir, liveM3U8))
		return err == nil
	})

	livePL := getString(t, srv.URL+"/stream")
	if !strings.Contains(livePL, "/hls/live") {
		t.Fatalf("expected live playlist from /stream, got:\n%s", livePL)
	}

	// The segments must be real, finished media. We verify them structurally
	// (MPEG-TS framing) rather than by decoding: the static ffmpeg build used in
	// CI segfaults on any MPEG-TS read, including a plain stream copy.
	var segs []string
	waitFor(t, "two finished live segments", 25*time.Second, func() bool {
		segs, _ = filepath.Glob(filepath.Join(hlsDir, "live*.ts"))
		return len(segs) >= 2
	})

	assertValidTS(t, segs[0])

	if !strings.Contains(livePL, filepath.Base(segs[0])) {
		t.Fatalf("playlist does not reference %s:\n%s", filepath.Base(segs[0]), livePL)
	}

	// 4. source disconnects -> back to automation
	wg.Wait()
	waitFor(t, "fallback to automation", 20*time.Second, func() bool {
		return strings.Contains(getString(t, srv.URL+"/api/v1/mode"), "automation")
	})

	fmt.Println("mode switch verified: automation -> live -> automation")
}

// assertValidTS checks MPEG-TS framing: every 188-byte packet must begin with
// the 0x47 sync byte, and the file must be a whole number of packets.
func assertValidTS(t *testing.T, path string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 188*4 {
		t.Fatalf("%s is too small to be a segment (%d bytes)", path, len(b))
	}
	if len(b)%188 != 0 {
		t.Fatalf("%s is not a whole number of TS packets (%d bytes)", path, len(b))
	}
	for i := 0; i < len(b); i += 188 {
		if b[i] != 0x47 {
			t.Fatalf("%s lost TS sync at offset %d", path, i)
		}
	}
	t.Logf("segment %s: %d bytes, %d valid TS packets", filepath.Base(path), len(b), len(b)/188)
}

func getString(t *testing.T, url string) string {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return string(b)
}
