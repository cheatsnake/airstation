// Command live-hls is a reference implementation of the "mode switch" pattern for
// a web radio station: automation runs continuously, and a live DJ (or talk-show
// host) can take over the stream at any moment, with the listeners never seeing a
// gap and never changing player.
//
// It is deliberately dependency-free (standard library only) so it drops into a
// Go project such as Airstation without touching go.mod.
//
// Endpoints
//
//	GET  /                 studio page (push-to-talk mic)
//	GET  /listen.html      listener page
//	POST /ingest           chunked WebM/Opus upload, streamed straight into ffmpeg
//	GET  /stream           the m3u8 the player uses; switches between live and auto
//	GET  /hls/*            the segment files
//	GET  /api/v1/mode      {"mode":"live"|"automation"}
package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	hlsDir     = "hls"
	autoM3U8   = "auto.m3u8"
	liveM3U8   = "live.m3u8"
	segSeconds = "2"
	listSize   = "6"
)

func ffmpegBin() string {
	if v := os.Getenv("FFMPEG_BIN"); v != "" {
		return v
	}
	return "ffmpeg"
}

// ---------------------------------------------------------------- mode state

type modeState struct {
	mu      sync.RWMutex
	live    bool
	since   time.Time
	ingests int
}

func (m *modeState) set(live bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if live && !m.live {
		m.since = time.Now()
	}
	m.live = live
}

func (m *modeState) isLive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.live
}

// ------------------------------------------------------------ live ingest

type liveIngest struct {
	mu    sync.Mutex
	cmd   *exec.Cmd
	stdin io.WriteCloser
}

var errAlreadyLive = errors.New("a live source is already connected")

// start spawns ffmpeg reading a WebM/Opus stream from stdin and writing rolling
// HLS segments. It is idempotent: a second caller joins the existing session.
func (li *liveIngest) start() error {
	li.mu.Lock()
	defer li.mu.Unlock()
	if li.cmd != nil {
		return errAlreadyLive
	}

	args := []string{
		"-hide_banner", "-loglevel", "warning",
		"-fflags", "nobuffer",
		"-f", "webm", "-i", "pipe:0",
		"-c:a", "aac", "-b:a", "128k", "-ar", "48000", "-ac", "2",
		"-f", "hls",
		"-hls_time", segSeconds,
		"-hls_list_size", listSize,
		"-hls_flags", "delete_segments+omit_endlist+independent_segments",
		"-hls_segment_filename", filepath.Join(hlsDir, "live%05d.ts"),
		filepath.Join(hlsDir, liveM3U8),
	}

	cmd := exec.Command(ffmpegBin(), args...)
	cmd.Stderr = os.Stderr
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	li.cmd, li.stdin = cmd, in
	return nil
}

// write pushes captured bytes into ffmpeg. Called from the ingest handler.
func (li *liveIngest) write(r io.Reader) (int64, error) {
	li.mu.Lock()
	in := li.stdin
	li.mu.Unlock()
	if in == nil {
		return 0, errors.New("no live session")
	}
	return io.Copy(in, r)
}

// stop closes the pipe so ffmpeg finalises and exits, ending the live session.
func (li *liveIngest) stop() {
	li.mu.Lock()
	cmd, in := li.cmd, li.stdin
	li.cmd, li.stdin = nil, nil
	li.mu.Unlock()

	if in != nil {
		in.Close()
	}
	if cmd != nil {
		done := make(chan struct{})
		go func() { cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			cmd.Process.Kill()
		}
	}
}

func (li *liveIngest) running() bool {
	li.mu.Lock()
	defer li.mu.Unlock()
	return li.cmd != nil
}

// ------------------------------------------------------------- automation

// runAutomation keeps an "automation" stream on air at all times by looping a
// source file into its own rolling HLS playlist. In a real station this is the
// existing segment-timer playout; here it stands in for it so the mode switch is
// observable end to end.
func runAutomation(src string) {
	for {
		cmd := exec.Command(ffmpegBin(),
			"-hide_banner", "-loglevel", "error",
			"-stream_loop", "-1", "-i", src,
			"-c:a", "aac", "-b:a", "128k", "-ar", "48000", "-ac", "2",
			"-f", "hls",
			"-hls_time", segSeconds,
			"-hls_list_size", listSize,
			"-hls_flags", "delete_segments+omit_endlist+independent_segments",
			"-hls_segment_filename", filepath.Join(hlsDir, "auto%05d.ts"),
			filepath.Join(hlsDir, autoM3U8),
		)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Printf("automation exited: %v", err)
		}
		time.Sleep(time.Second)
	}
}

// ------------------------------------------------------------------ server

type server struct {
	mode *modeState
	live *liveIngest
}

func newServer() *server {
	return &server{mode: &modeState{}, live: &liveIngest{}}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/hls/", http.StripPrefix("/hls/", http.FileServer(http.Dir(hlsDir))))
	mux.HandleFunc("/stream", s.handleStream)
	mux.HandleFunc("/api/v1/mode", s.handleMode)
	mux.HandleFunc("/ingest", s.handleIngest)
	mux.HandleFunc("/listen.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, listenHTML)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, studioHTML)
	})
	return mux
}

func (s *server) handleMode(w http.ResponseWriter, r *http.Request) {
	mode := "automation"
	if s.mode.isLive() {
		mode = "live"
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	io.WriteString(w, `{"mode":"`+mode+`"}`)
}

// handleStream serves whichever playlist is currently on air, rewriting the
// relative segment names to absolute URLs so the client resolves them correctly
// no matter what path it fetched the playlist from.
func (s *server) handleStream(w http.ResponseWriter, r *http.Request) {
	name := autoM3U8
	if s.mode.isLive() {
		name = liveM3U8
	}

	b, err := os.ReadFile(filepath.Join(hlsDir, name))
	if err != nil {
		http.Error(w, "stream not ready yet", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	io.WriteString(w, absolutise(string(b)))
}

func absolutise(m3u8 string) string {
	lines := strings.Split(m3u8, "\n")
	for i, ln := range lines {
		t := strings.TrimSpace(ln)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		if !strings.HasPrefix(t, "http") && !strings.HasPrefix(t, "/") {
			lines[i] = "/hls/" + t
		}
	}
	return strings.Join(lines, "\n")
}

// handleIngest is the heart of the mode switch: the moment bytes arrive the
// station is live, and when the body ends it falls back to automation.
func (s *server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if err := s.live.start(); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	s.mode.set(true)
	log.Println("mode: live (source connected)")

	n, err := s.live.write(r.Body)
	if err != nil {
		log.Printf("ingest ended: %v", err)
	}
	log.Printf("ingest: received %d bytes", n)

	s.live.stop()
	s.mode.set(false)
	log.Println("mode: automation (source disconnected)")
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	if err := os.MkdirAll(hlsDir, 0o755); err != nil {
		log.Fatal(err)
	}

	src := os.Getenv("AUTOMATION_SRC")
	if src == "" {
		src = "tone.m4a"
	}
	go runAutomation(src)

	s := newServer()
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("studio  http://localhost%s/", addr)
	log.Printf("listen  http://localhost%s/listen.html", addr)
	log.Fatal(http.ListenAndServe(addr, s.routes()))
}

// --------------------------------------------------------------------- pages

const studioHTML = `<!doctype html>
<html><head><meta charset="utf-8"><title>Studio — live source</title>
<style>
 body{font:15px/1.5 system-ui,sans-serif;background:#111;color:#eee;display:grid;place-items:center;height:100vh;margin:0}
 .card{width:420px;background:#1b1b1b;border:1px solid #2c2c2c;border-radius:14px;padding:24px}
 h1{font-size:16px;margin:0 0 16px;letter-spacing:.02em;text-transform:uppercase;color:#8a8a8a}
 #ptt{width:100%;height:140px;border-radius:12px;border:2px solid #333;background:#222;color:#eee;font-size:18px;cursor:pointer;user-select:none;transition:.12s}
 #ptt.on{background:#c62828;border-color:#e53935;box-shadow:0 0 40px #c6282855}
 #ptt:disabled{opacity:.35;cursor:not-allowed}
 .row{display:flex;justify-content:space-between;margin-top:14px;font-size:13px;color:#9a9a9a}
 .dot{display:inline-block;width:9px;height:9px;border-radius:50%;background:#444;margin-right:6px}
 .dot.live{background:#e53935;box-shadow:0 0 8px #e53935}
 .meter{height:6px;background:#2a2a2a;border-radius:3px;overflow:hidden;margin-top:14px}
 .meter>i{display:block;height:100%;width:0;background:#4caf50;transition:width .06s}
</style></head><body>
<div class="card">
  <h1>Live source</h1>
  <button id="ptt" disabled>Connecting…</button>
  <div class="meter"><i id="bar"></i></div>
  <div class="row"><span><span id="dot" class="dot"></span><span id="state">offline</span></span><span id="mode">—</span></div>
  <p style="font-size:12px;color:#777;margin-top:18px">
    Hold the button (or hold <b>Space</b>) to talk. Release to return to the automation
    bed. The stream never stops — only the microphone gain changes.
  </p>
</div>
<script>
let rec, gainNode, ctx, stream, wsReady=false, streamWriter=null;
const ptt=document.getElementById('ptt'), state=document.getElementById('state'),
      dot=document.getElementById('dot'), bar=document.getElementById('bar'),
      modeEl=document.getElementById('mode');

// Poll the mode endpoint so the operator can see what is on air.
setInterval(async()=>{try{
  const r=await fetch('/api/v1/mode',{cache:'no-store'});modeEl.textContent=(await r.json()).mode;
}catch(e){modeEl.textContent='—'}},1000);

async function connect(){
  stream=await navigator.mediaDevices.getUserMedia({audio:{
    channelCount:1, sampleRate:48000,
    echoCancellation:false, noiseSuppression:false, autoGainControl:false}});

  ctx=new AudioContext({sampleRate:48000, latencyHint:'interactive'});
  const src=ctx.createMediaStreamSource(stream);
  gainNode=ctx.createGain();
  gainNode.gain.value=0;                 // push-to-talk starts CLOSED
  const dest=ctx.createMediaStreamDestination();
  src.connect(gainNode).connect(dest);

  // level meter (pre-gain, so the operator sees the mic even while muted)
  const an=ctx.createAnalyser(); an.fftSize=512; src.connect(an);
  const buf=new Float32Array(an.fftSize);
  (function tick(){
    an.getFloatTimeDomainData(buf);
    let p=0; for(const v of buf) p=Math.max(p,Math.abs(v));
    bar.style.width=Math.min(100,p*220)+'%';
    requestAnimationFrame(tick);
  })();

  const mime=MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
    ?'audio/webm;codecs=opus':'audio/webm';
  rec=new MediaRecorder(dest.stream,{mimeType:mime,audioBitsPerSecond:64000});

  // One continuous ReadableStream = one continuous WebM. This is what makes the
  // server-side ffmpeg decode cleanly; posting each blob as its own file does not.
  let resolveChunk, chain=Promise.resolve();
  const rs=new ReadableStream({
    start(c){
      rec.ondataavailable=e=>{
        if(!e.data.size) return;
        chain=chain.then(async()=>c.enqueue(new Uint8Array(await e.data.arrayBuffer())))
                   .catch(()=>{});
      };
      rec.onstop=()=>chain.then(()=>c.close());
    }
  });
  rec.start(250);

  fetch('/ingest',{method:'POST',body:rs,duplex:'half',
    headers:{'Content-Type':'application/octet-stream'}}).catch(()=>{});

  ptt.disabled=false; ptt.textContent='HOLD TO TALK';
  state.textContent='ready'; dot.classList.add('live');
}

function talk(on){
  if(!gainNode) return;
  gainNode.gain.setTargetAtTime(on?1:0, ctx.currentTime, 0.008);
  ptt.classList.toggle('on',on);
}

ptt.addEventListener('pointerdown',e=>{e.preventDefault();ptt.setPointerCapture(e.pointerId);talk(true)});
ptt.addEventListener('pointerup',()=>talk(false));
ptt.addEventListener('pointercancel',()=>talk(false));
window.addEventListener('blur',()=>talk(false));
window.addEventListener('keydown',e=>{if(e.code==='Space'&&!e.repeat){e.preventDefault();talk(true)}});
window.addEventListener('keyup',e=>{if(e.code==='Space'){e.preventDefault();talk(false)}});

connect().catch(err=>{ptt.textContent='MIC ERROR'; state.textContent=err.message});
</script></body></html>`

const listenHTML = `<!doctype html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Listen</title>
<style>
 body{font:15px/1.5 system-ui,sans-serif;background:#0e0e10;color:#eee;display:grid;place-items:center;height:100vh;margin:0}
 .card{width:360px;text-align:center;background:#17171a;border:1px solid #26262b;border-radius:16px;padding:28px}
 button{width:96px;height:96px;border-radius:50%;border:0;background:#2f7d32;color:#fff;font-size:15px;cursor:pointer}
 button.on{background:#c62828}
 #mode{font-size:12px;color:#888;margin-top:16px;letter-spacing:.08em;text-transform:uppercase}
</style></head><body>
<div class="card">
  <button id="b">PLAY</button>
  <div id="mode">automation</div>
</div>
<audio id="a" preload="none"></audio>
<script src="https://cdn.jsdelivr.net/npm/hls.js@1/dist/hls.min.js"></script>
<script>
const a=document.getElementById('a'), b=document.getElementById('b'), modeEl=document.getElementById('mode');
let hls=null, playing=false;

function start(){
  const url='/stream';
  if(window.Hls && Hls.isSupported()){
    hls=new Hls({liveSyncDurationCount:3, lowLatencyMode:false});
    hls.loadSource(url); hls.attachMedia(a);
  } else { a.src=url; }        // Safari plays HLS natively
  a.play().catch(()=>{});
}

b.onclick=()=>{
  playing=!playing;
  b.classList.toggle('on',playing);
  b.textContent=playing?'STOP':'PLAY';
  if(playing) start(); else { if(hls) hls.destroy(); hls=null; a.pause(); }
};

setInterval(async()=>{try{
  const r=await fetch('/api/v1/mode',{cache:'no-store'});
  modeEl.textContent=(await r.json()).mode;
}catch(e){}},1000);
</script></body></html>`
