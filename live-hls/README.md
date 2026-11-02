# Live mode for Airstation — reference implementation

A working, tested implementation of the **mode switch** pattern for a web radio
station: automation runs continuously, a DJ or talk-show host takes over the
stream from the browser at any moment, and listeners never change player and
never hear a gap.

Standard library only — no modules, no second daemon, no WebSocket library. It
drops into a Go project such as Airstation without touching `go.mod`.

## What it proves

`go test -v` drives the whole lifecycle against a real ffmpeg:

```
automation on air  ->  DJ connects  ->  live on air  ->  DJ disconnects  ->  automation
```

Specifically it asserts that:

1. `/stream` serves the automation playlist, with segment paths rewritten to absolute URLs.
2. A chunked `POST /ingest` flips `/api/v1/mode` to `live`.
3. ffmpeg produces a real rolling HLS playlist from the piped WebM/Opus.
4. `/stream` switches to the live playlist, and that playlist references real segments.
5. The segments are structurally valid MPEG-TS (0x47 sync, whole 188-byte packets).
6. When the upload ends, the station falls back to automation automatically.

```
--- PASS: TestModeSwitch (7.49s)
    segment live00000.ts: 35720 bytes, 190 valid TS packets
mode switch verified: automation -> live -> automation
```

## Running it

```sh
# ffmpeg must be on PATH, or set FFMPEG_BIN
go run .            # studio at http://localhost:8080/
                    # listen at http://localhost:8080/listen.html

go test -v          # end-to-end mode switch test
```

`AUTOMATION_SRC` points at the file used as the automation bed (default
`tone.m4a` — generate one with
`ffmpeg -f lavfi -i "sine=frequency=220:duration=6" -c:a aac tone.m4a`).

## Why it is built this way

**Push-to-talk gates a gain node, not the capture track.** The microphone stream
runs continuously from page load; the PTT button moves a `GainNode` between 0
and 1 with an 8 ms ramp. This matters because ffmpeg needs a *continuous* input:
if you stop and start the recorder, the pipe stalls, the HLS muxer stops emitting
segments, and listeners rebuffer. Gating gain means the silence is real audio and
the stream never breaks. Unmuting is instant.

**One continuous WebM stream, not one file per chunk.** `MediaRecorder` emits a
blob every 250 ms, but only the first contains the EBML header. Posting each blob
as its own file gives ffmpeg one decodable file and a pile of headerless clusters.
Feeding them into a single `ReadableStream` gives ffmpeg one valid WebM — which is
why `-f webm -i pipe:0` works here.

**Mode is derived from the connection, not from a flag.** The station is live for
exactly as long as an ingest body is open. No timer to get out of sync, no way to
be left live by accident after a browser crash.

## Wiring it into Airstation

The prototype mirrors Airstation's shape so the port is mechanical:

| Prototype | Airstation |
|---|---|
| `modeState` | a `Live bool` + `LiveSince time.Time` on `playback.State` |
| `handleIngest` | new `POST /api/v1/live/ingest`, wrapped in the existing `s.jwtAuth` |
| `handleMode` | new `GET /api/v1/live/mode` |
| `handleStream` | replace the body of `handleHLSPlaylist` |
| `liveIngest` | new `internal/live/service.go` |

The three changes that matter:

**1. `State.Run()` must stop advancing while live.** Right now it ticks
`CurrentTrackElapsed += 1` every second and regenerates the playlist. While a DJ
is on air, skip the increment (and skip `loadNextTrack`) but keep the ticker
running so the mode is re-checked:

```go
if !s.IsPlaying || s.Live {
    continue
}
```

**2. `handleHLSPlaylist` becomes a switch.** When `Live` is true, serve
`hls/live.m3u8` (rewritten to absolute segment paths, as `absolutise` does);
otherwise serve `playlist.Generate(elapsed)` exactly as today.

**3. Ending a live session needs a decision.** On disconnect you can either
`Play()` (restarts the current track from zero — audible jump) or wait for the
current track to finish and resume at the next boundary (clean, but up to a few
minutes of automation bed). For a talk show, resuming at the next track boundary
is what listeners expect. The `#EXT-X-DISCONTINUITY` tag on the first segment of
the resumed track is what tells the player to reset its timeline.

Broadcast the transition over the SSE emitter that already exists
(`internal/pkg/sse`, `GET /api/v1/events`) as a `mode` event, so every open studio
client updates without polling.

## Multi-mic talk shows

The prototype accepts **one** ingest at a time and returns `409` to a second.
For a panel you need a mixer, and the shape is:

- one ingest per participant, each decoded by its own ffmpeg to raw PCM;
- a single mixing loop that sums the frames with per-participant gain;
- one encoder fed by the mix, writing the live HLS.

Summing float32 frames from N sources is easy; keeping them aligned is not. Each
participant needs a jitter buffer, and you need a common timeline — otherwise
one person's audio drifts against another's by a few tens of milliseconds and it
sounds like a badly tuned radio. Budget for that; it is the actual work in a
multi-party feature.

A simpler first step that covers most talk shows: one host on the live path, and
guests pre-recorded or dialled in over a separate call that the host mixes
locally before it reaches the browser.

## Cloudflare tunnels

`trycloudflare.com` quick tunnels are for testing, and a few things follow from
that:

- **HTTPS is why your mic works.** `getUserMedia` requires a secure context.
  `https://*.trycloudflare.com` qualifies; `http://192.168.x.x` does not. That is
  probably the reason you reached for the tunnel in the first place, and it is
  the right instinct.
- **The URL is ephemeral.** A new random subdomain on every restart, no SLA, rate
  limited. Fine for a rehearsal, wrong for a scheduled show. Move to a named
  tunnel on a domain you own before you tell anyone to tune in.
- **WebRTC will not work over a quick tunnel.** Tunnels proxy HTTP; WebRTC needs
  UDP. This is the practical reason HLS is the right distribution choice for you —
  it is the one live format that survives an HTTP-only path.
- **HLS is a good fit for a tunnel** because it is many small requests rather than
  one long-lived connection. The chunked upload is the only long-lived request in
  the design.
- **Check Cloudflare's terms before a public launch.** Their free tier is not
  intended as a media CDN, and disproportionate audio delivery is explicitly out
  of scope. For a handful of listeners nobody will notice; for a real audience,
  put the segments behind a proper storage origin or a streaming-friendly CDN.

## Known limits of the prototype

- One live source at a time (see above).
- `fetch()` with a streaming request body (`duplex: 'half'`) is Chrome/Edge only.
  Safari cannot do it — for a Safari studio client, switch the ingest transport to
  a WebSocket, which every browser supports and which tunnels pass through fine.
  That is the one change worth making before you rely on it.
- 6–10 seconds of latency with 2-second segments. Lower it with `-hls_time 1` at
  the cost of more requests; go below ~2 s and you want LL-HLS instead.
- No auth. Add the JWT middleware before this goes anywhere public, or anyone who
  finds `/ingest` can take over your station.
