// Package live manages the station's "delayed live" mode. A live source is
// encoded by FFmpeg into the same HLS format as ordinary tracks, so the
// existing /stream endpoint and elapsed-time engine keep working. Listeners
// stay a few seconds behind the speaker.
//
// Two kinds of source are supported:
//
//   - ingest: one continuous WebM/Opus body streamed from the studio browser
//     (microphone) and piped straight into FFmpeg. The station is live for
//     exactly as long as the request body is open.
//   - external: an HTTP/RTMP/file URL that FFmpeg pulls and segments itself.
package live

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cheatsnake/airstation/internal/station"
)

const (
	playlistName = "live.m3u8"
	segmentGlob  = "live*.ts"

	// Short segments keep the delay low; the listener needs a few of them in
	// the playlist before playback can start.
	segmentSeconds = 2
	playlistSize   = 6

	// Live audio is deliberately encoded at a low bitrate so the encoder can
	// produce segments quickly and listeners do not build a large buffer.
	liveBitRate = 36
)

// ErrAlreadyLive is returned when a second source tries to take over while a
// live session is already running.
var ErrAlreadyLive = errors.New("a live source is already connected")

// Manager owns the active live source and the FFmpeg process that produces it.
type Manager struct {
	mu sync.Mutex

	dir       string // directory where live segments and the playlist are written
	urlPrefix string // public URL prefix the segments are served from

	active    bool
	mode      string // "mic" or "external"
	startedAt time.Time
	// segmentPrefix is unique per session so a browser can never serve a cached
	// segment from an earlier show that happened to use the same filename.
	segmentPrefix string
	// session identifies the current live session. It lets a delayed cleanup
	// avoid deleting files belonging to a newer session.
	session uint64

	cmd   *exec.Cmd
	stdin io.WriteCloser

	station *station.Service
	log     *slog.Logger
}

// NewManager creates a live manager that writes segments to dir and advertises
// them under urlPrefix (for example "/static/tmp/live").
func NewManager(dir, urlPrefix string, stationSvc *station.Service, log *slog.Logger) *Manager {
	return &Manager{dir: dir, urlPrefix: urlPrefix, station: stationSvc, log: log}
}

// audioBitRate is fixed for live sessions. A lower bitrate keeps encoder
// latency and listener buffering small.
func (m *Manager) audioBitRate() int {
	return liveBitRate
}

// IsActive reports whether a live source is currently running.
func (m *Manager) IsActive() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active
}

// Mode reports "mic", "external", or an empty string when off air.
func (m *Manager) Mode() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mode
}

// StartedAt returns the time the current live session began.
func (m *Manager) StartedAt() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.startedAt
}

// SessionID returns an identifier for the current live session. It can be
// captured before Stop and passed to CleanupSession later.
func (m *Manager) SessionID() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.session
}

// SegmentCount returns the number of segments currently in the live window.
func (m *Manager) SegmentCount() int {
	m.mu.Lock()
	dir, active := m.dir, m.active
	m.mu.Unlock()
	if !active {
		return 0
	}
	files, err := filepath.Glob(filepath.Join(dir, segmentGlob))
	if err != nil {
		return 0
	}
	return len(files)
}

// StartIngest starts FFmpeg with a WebM/Opus stream on stdin. The caller must
// feed the request body through Write and call Stop when the body ends.
func (m *Manager) StartIngest() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.active {
		return ErrAlreadyLive
	}
	if err := m.prepareDir(); err != nil {
		return err
	}

	args := []string{
		"-hide_banner", "-loglevel", "warning",
		"-fflags", "nobuffer",
		"-f", "webm", "-i", "pipe:0",
	}
	args = append(args, m.hlsOutputArgs()...)

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to open ffmpeg stdin: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg ingest: %w", err)
	}

	m.cmd = cmd
	m.stdin = stdin
	m.active = true
	m.mode = "mic"
	m.startedAt = time.Now()
	m.log.Info("Live source started (browser ingest)")
	return nil
}

// StartExternal starts an FFmpeg relay that pulls url and segments it into the
// live playlist.
func (m *Manager) StartExternal(url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.active {
		return ErrAlreadyLive
	}
	if url == "" {
		return errors.New("stream url is required")
	}
	if err := m.prepareDir(); err != nil {
		return err
	}

	args := []string{
		"-hide_banner", "-loglevel", "warning",
		"-re", "-i", url,
	}
	args = append(args, m.hlsOutputArgs()...)

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg relay: %w", err)
	}

	m.cmd = cmd
	m.stdin = nil
	m.active = true
	m.mode = "external"
	m.startedAt = time.Now()
	m.log.Info("Live source started (external)", slog.String("url", url))
	return nil
}

// Write copies an ingest body into the live FFmpeg process. It blocks until the
// reader ends or FFmpeg exits.
func (m *Manager) Write(r io.Reader) (int64, error) {
	m.mu.Lock()
	stdin := m.stdin
	m.mu.Unlock()

	if stdin == nil {
		return 0, errors.New("no live ingest is running")
	}
	return io.Copy(stdin, r)
}

// WriteBytes sends one chunk of an ingest stream (for example one WebSocket
// binary message) to the live FFmpeg process.
func (m *Manager) WriteBytes(p []byte) (int, error) {
	m.mu.Lock()
	stdin := m.stdin
	m.mu.Unlock()

	if stdin == nil {
		return 0, errors.New("no live ingest is running")
	}
	return stdin.Write(p)
}

// Stop closes the ingest pipe (letting FFmpeg finalise) and waits briefly for
// it to exit. The last playlist and segments are deliberately left on disk:
// the playback engine keeps advertising them until the automation playlist is
// ready, so listeners never see an empty /stream during the switch. Call
// Cleanup once those files are no longer needed.
func (m *Manager) Stop() {
	m.mu.Lock()
	cmd, stdin := m.cmd, m.stdin
	wasActive := m.active
	m.cmd, m.stdin = nil, nil
	m.active = false
	m.mode = ""
	m.mu.Unlock()

	if stdin != nil {
		_ = stdin.Close()
	}
	if cmd != nil {
		done := make(chan struct{})
		go func() {
			_ = cmd.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			<-done
		}
	}

	if wasActive {
		m.log.Info("Live source stopped")
	}
}

// CleanupSession removes the playlist and segments belonging to the live
// session identified by session. It is a no-op when that session is still on
// air or when a newer session has already started, so a delayed cleanup can
// never remove live audio from a show that began in the meantime.
func (m *Manager) CleanupSession(session uint64) {
	m.mu.Lock()
	if m.active || m.session != session {
		m.mu.Unlock()
		return
	}
	// Removing the playlist under the lock keeps a new session from starting
	// between the check above and the removal.
	_ = os.Remove(filepath.Join(m.dir, playlistName))
	dir := m.dir
	prefix := m.segmentPrefix
	m.mu.Unlock()

	files, err := filepath.Glob(filepath.Join(dir, prefix+"*.ts"))
	if err != nil {
		return
	}
	for _, file := range files {
		_ = os.Remove(file)
	}
}

// BuildPlaylist returns the current live playlist with segment paths rewritten
// to their public URL. It returns an empty string until FFmpeg has produced its
// first segment, which lets callers keep serving the previous playlist during
// the takeover gap.
func (m *Manager) BuildPlaylist() string {
	m.mu.Lock()
	dir, prefix, active := m.dir, m.urlPrefix, m.active
	m.mu.Unlock()

	if !active {
		return ""
	}
	raw, err := os.ReadFile(filepath.Join(dir, playlistName))
	if err != nil {
		return ""
	}

	var out strings.Builder
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "#EXT-X-ENDLIST" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			out.WriteString(line)
			out.WriteString("\n")
			continue
		}
		out.WriteString(prefix)
		out.WriteString("/")
		out.WriteString(path.Base(line))
		out.WriteString("\n")
	}

	playlist := out.String()
	if !strings.Contains(playlist, "#EXTINF") {
		return ""
	}
	return playlist
}

// hlsOutputArgs builds the codec and HLS muxer arguments shared by both source
// kinds. Callers prepend the input options.
func (m *Manager) hlsOutputArgs() []string {
	return []string{
		"-c:a", "aac",
		"-b:a", strconv.Itoa(m.audioBitRate()) + "k",
		"-ar", "44100", "-ac", "2",
		"-f", "hls",
		"-hls_time", strconv.Itoa(segmentSeconds),
		"-hls_list_size", strconv.Itoa(playlistSize),
		"-hls_flags", "delete_segments+omit_endlist+independent_segments+discont_start",
		"-hls_segment_filename", filepath.Join(m.dir, m.segmentPrefix+"%05d.ts"),
		filepath.Join(m.dir, playlistName),
	}
}

// prepareDir creates the segment directory and removes leftovers from an
// earlier session. The caller must hold m.mu.
func (m *Manager) prepareDir() error {
	if err := os.MkdirAll(m.dir, 0o755); err != nil {
		return err
	}
	// The caller already holds m.mu.
	m.cleanupDir(m.dir)
	m.session++
	m.segmentPrefix = "live" + strconv.FormatInt(time.Now().UnixNano(), 36) + "-"
	return nil
}

// cleanupDir removes the playlist and any segments from dir.
func (m *Manager) cleanupDir(dir string) {
	_ = os.Remove(filepath.Join(dir, playlistName))
	files, err := filepath.Glob(filepath.Join(dir, segmentGlob))
	if err != nil {
		return
	}
	for _, file := range files {
		_ = os.Remove(file)
	}
}
