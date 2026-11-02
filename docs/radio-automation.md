# Radio automation

Phase 1 turns Airstation from a queue-based music player into a small radio
station: the library knows about more than songs, the station can insert
jingles and ads on its own, recorded RJ links are mixed into the music, and the
studio has a broadcast console for firing carts and links.

Everything described here runs on the existing HLS engine, so the public
`/stream` endpoint and the listener player keep working exactly as before.

## Asset kinds

Every library item has a `kind`:

| Kind         | Used for                                   | Minimum length |
| ------------ | ------------------------------------------ | -------------- |
| `music`      | Songs for rotation                         | 15 seconds     |
| `jingle`     | Station stings, sweeps, short branding      | 1 second       |
| `ad`         | Commercials                                | 1 second       |
| `station_id` | Top-of-hour IDs, frequency/name announcements | 1 second    |
| `voice`      | Pre-recorded RJ links                      | 1 second       |

Uploads are processed synchronously, so the studio reports a rejected file
(for example an ad that is too short) instead of accepting it and failing
silently. The upload form in **Library → Add** lets you pick the kind before
choosing files; the library can also be filtered by kind.

The old rule still applies to songs: music must be at least 15 seconds because
the HLS playlist keeps three 5-second segments buffered. Short station assets
only need to be long enough to form a segment.

## Cart wall

All non-music assets (`jingle`, `ad`, `station_id`, `voice`) are available as
carts on the **Console** tab. Each cart has two actions:

- **Next** – insert the cart directly after the current track. The current
  track keeps playing to its end, then the cart plays and the rest of the queue
  follows.
- **Now** – put the cart at the front of the queue and restart playout. The
  stream jumps to the cart immediately (listeners may hear a short
  discontinuity while their player re-buffers).

Carts are just library assets, so uploading a new jingle makes it appear on the
wall automatically.

## Schedule rules

The **Schedule** tab creates rules that insert jingles, ads or station IDs into
the program log automatically. A rule can use either trigger, or both:

- **Every N music tracks** – fire once N music tracks have finished since the
  last automatic interlude. `0` disables this trigger.
- **Clock trigger** – fire at a specific minute of the hour (`0`–`59`), once per
  wall-clock hour. `-1` disables this trigger.

Clock rules are evaluated against the server process's local timezone. Clock
triggers take priority over rotation triggers, and the station will not stack
two interludes back to back: if an interlude is already playing or queued next,
the rule waits.

When a rule fires, the station picks a random asset of the configured kind that
is not already in the queue. If the library has no asset of that kind, the rule
is skipped until one is uploaded.

## Voice-over (RJ links)

The voice-over composer on the **Console** takes a `voice` asset and mixes it
into a music track *before* the track is split into HLS segments. Listeners
therefore hear one continuous programme: the music is ducked under the voice,
and the link fades in and out at the values the RJ chose.

Settings:

| Setting     | Meaning                                                             |
| ----------- | ------------------------------------------------------------------- |
| `Start at`  | Seconds into the music track where the link begins.                 |
| `Fade in`   | Length of the voice fade-in.                                        |
| `Fade out`  | Length of the voice fade-out.                                       |
| `Duck music`| How far the music drops while the link is talking (`-6` to `-24` dB).|
| `Target`    | `next` mixes into the upcoming track, `current` restarts the current track and talks over it. |

Behavior details:

- The ducking depth is exact. The mix uses a time-based FFmpeg volume envelope
  rather than a level-triggered compressor, so the music beds down by exactly
  the requested number of decibels and comes back up during the voice fade-out.
- The rendered mix keeps the full length of the music bed. It is written as
  `<trackID>-vo<shortID>.m4a` in the temporary directory and segmented like any
  other track, so elapsed-time tracking and HLS buffering are unchanged.
- The voice fade-in and fade-out are clamped to half the voice length, and the
  start offset is clamped so the link never runs past the end of the song.
- A voice clip longer than the chosen music track is rejected.
- Voice-overs are **one-shot**. Once the mixed track has aired, the request is
  marked `played` and the track is clean the next time it rotates around.
- Cancelling a pending or rendered link removes the mixed file. If it was
  already mixed into the upcoming track, the station re-segments that track
  without the link.
- If mixing fails for any reason, the station logs a warning and plays the
  original track instead of going off air.

Targeting `current` restarts the current track from the top, so the link always
starts at the requested offset. This is why the console asks for playback to be
running before it enables that button.

### Mix semantics

The rendered graph is built in `internal/pkg/mix` and validated by unit tests:

1. Both inputs are converted to 44.1 kHz stereo float.
2. The voice is faded with `afade`, delayed to the start offset with `adelay`,
   and padded to the music length with `apad`.
3. The music is attenuated by a volume expression that ramps from 1 to the
   duck gain over the voice fade-in, holds for the body of the link, and ramps
   back to 1 over the voice fade-out.
4. `amix` keeps the length of the first input (the full music bed), and
   `alimiter` prevents clipping.

## API

All automation endpoints require the usual `jwt` login cookie.

### Assets and carts

| Method   | Path                          | Body / query                                        |
| -------- | ----------------------------- | --------------------------------------------------- |
| `POST`   | `/api/v1/tracks`              | multipart: `tracks` file(s), `kind`                 |
| `GET`    | `/api/v1/tracks`              | `kind=music\|jingle\|ad\|station_id\|voice` (optional) |
| `GET`    | `/api/v1/carts`               | –                                                   |
| `POST`   | `/api/v1/carts/{id}/play`     | `{"mode":"next"}` or `{"mode":"now"}`               |

`POST /api/v1/tracks` returns a summary of what was stored and what was
rejected. A rejected upload returns `400` with the reason.

### Voice-overs

| Method   | Path                        | Body                                                                 |
| -------- | --------------------------- | -------------------------------------------------------------------- |
| `GET`    | `/api/v1/voice-overs`       | –                                                                     |
| `POST`   | `/api/v1/voice-overs`       | `{"voiceTrackId","target","fadeIn","fadeOut","duckDb","startOffset"}` |
| `DELETE` | `/api/v1/voice-overs/{id}`  | –                                                                     |

Statuses: `pending`, `rendered`, `played`, `failed`, `cancelled`.

### Schedule

| Method   | Path                      | Body                                                        |
| -------- | ------------------------- | ----------------------------------------------------------- |
| `GET`    | `/api/v1/schedule`        | –                                                           |
| `POST`   | `/api/v1/schedule`        | `{"name","kind","everyNTracks","atMinute","enabled"}`       |
| `PUT`    | `/api/v1/schedule/{id}`   | same as create                                              |
| `DELETE` | `/api/v1/schedule/{id}`   | –                                                           |

### Live

| Method | Path                      | Body / notes                                              |
| ------ | ------------------------- | --------------------------------------------------------- |
| `GET`  | `/api/v1/live`            | status: `{"live","mode","startedAt","segmentCount"}`      |
| `GET`  | `/api/v1/live/ingest`     | WebSocket upgrade for the studio microphone               |
| `POST` | `/api/v1/live/ingest`     | raw `audio/webm` body held open for scripted clients      |
| `POST` | `/api/v1/live/start`      | `{"mode":"external","url":"..."}` for an external relay   |
| `POST` | `/api/v1/live/stop`       | –                                                         |

`GET /api/v1/live/ingest` returns a WebSocket that stays open until the operator
stops; `POST /api/v1/live/ingest` returns once the request body ends. A second
ingest or external source while one is active returns `409 Conflict`.

## Studio and player

- **Studio → Console** shows on-air state, current and next track, listener
  count, the cart wall, the voice-over composer, link history, the upcoming
  program log, and active automation rules.
- **Studio → Library** filters by kind and chooses the kind when uploading.
- **Studio → Queue** shows kind badges so interludes are visible in the log.
- **Studio → Schedule** creates, edits, enables and deletes rules.
- **Player** shows a microphone badge while a link is airing and an "Up next"
  line under the current track.

## Live mode (Phase 3, delayed)

The console can switch the station from automation to a **delayed live** source.
Nothing happens until the operator explicitly clicks **Go live**, so merely
opening the Console never takes the station off automation.

- **Microphone** – the browser opens one WebM/Opus capture and streams each
  recorder chunk over a **WebSocket** (`GET /api/v1/live/ingest`). The server
  pipes those messages straight into one FFmpeg process, which writes a rolling
  HLS window of **2-second segments** (six in the playlist). The station is live
  for exactly as long as the socket stays open; closing the tab, losing the
  network, or pressing **Stop** ends the session and resumes automation.
  WebSockets work over plain HTTP/1.1, so no TLS or HTTP/2 reverse proxy is
  required for microphone mode.
- **Push-to-talk** – the microphone runs continuously and PTT only ramps a
  WebAudio gain node (8 ms), so the upload pipe never stalls. Hold the **Talk**
  button or the space bar to speak, or switch on **Open mic** for a continuous
  show.
- **External stream** – give a URL (http, rtmp, or a local file). The server
  runs FFmpeg to pull and segment it into the same live playlist.
- **Stop** – ends the live source and resumes automation with the queue intact.

The live playlist is served from the same `/static/tmp/live/` route as other
generated segments, so `/stream` and the existing elapsed-time engine keep
working unchanged. While the first live segment is still being produced, the
previous automation playlist stays on air, which avoids a cold-start gap.
End-to-end latency is roughly **6–10 seconds** (encoder, segment duration and
player buffer).

The ingest endpoint requires the studio JWT cookie. The studio uses a
browser-native WebSocket, so microphone live mode works over HTTP/1.1 in all
modern browsers; the same path also accepts a plain chunked `POST` body for
scripts and server-side encoders. External streams work from any browser.

The console also has a **Record 8s voice clip** button that captures a short
clip and uploads it as a `voice` asset, so it can be fired from the cart wall
or used for a voice-over immediately afterwards.

## Still not included

- live ducking of automation music (the station is fully live while on air),
- a lower-latency (WebRTC) pipeline and a browser-agnostic ingest fallback,
- automatic level normalisation of uploaded voice clips,
- crossfades between music tracks.

Pre-recorded voice links are still pre-rendered into the music bed so the
elapsed-time engine stays the single source of truth for playout.
