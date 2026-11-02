package voiceover

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cheatsnake/airstation/internal/pkg/mix"
	"github.com/cheatsnake/airstation/internal/pkg/ulid"
	"github.com/cheatsnake/airstation/internal/station"
	"github.com/cheatsnake/airstation/internal/track"
)

type Service struct {
	store    Store
	tracks   TrackFinder
	station  *station.Service
	renderer mix.Renderer
	log      *slog.Logger
}

func NewService(store Store, tracks TrackFinder, stationSvc *station.Service, renderer mix.Renderer, log *slog.Logger) *Service {
	return &Service{
		store:    store,
		tracks:   tracks,
		station:  stationSvc,
		renderer: renderer,
		log:      log,
	}
}

// Create validates and stores a new voice-over request. For the "current"
// target the caller must supply the currently playing track ID.
func (s *Service) Create(voiceTrackID, target string, fadeIn, fadeOut, duckDB, startOffset float64, currentTrackID string) (*VoiceOver, error) {
	if target == "" {
		target = TargetNext
	}
	if target != TargetNext && target != TargetCurrent {
		return nil, fmt.Errorf("target must be %q or %q", TargetNext, TargetCurrent)
	}
	if target == TargetCurrent && currentTrackID == "" {
		return nil, fmt.Errorf("there is no current track to talk over")
	}

	resolved, err := s.tracks.TracksByIDs([]string{voiceTrackID})
	if err != nil {
		return nil, fmt.Errorf("failed to load voice asset: %w", err)
	}
	if len(resolved) == 0 {
		return nil, fmt.Errorf("voice asset not found")
	}
	if resolved[0].Kind != track.KindVoice {
		return nil, fmt.Errorf("asset is not a voice clip")
	}

	vo := &VoiceOver{
		ID:            newID(),
		VoiceTrackID:  voiceTrackID,
		VoiceName:     resolved[0].Name,
		Target:        target,
		TargetTrackID: "",
		FadeIn:        clamp(fadeIn, 0, 15),
		FadeOut:       clamp(fadeOut, 0, 15),
		DuckDB:        clamp(duckDB, -30, 0),
		StartOffset:   clamp(startOffset, 0, 600),
		Status:        StatusPending,
		CreatedAt:     time.Now().Unix(),
	}
	if target == TargetCurrent {
		vo.TargetTrackID = currentTrackID
	}

	if err := s.store.Add(vo); err != nil {
		return nil, err
	}

	return vo, nil
}

func (s *Service) List() ([]*VoiceOver, error) {
	return s.store.List()
}

func (s *Service) Cancel(id string) error {
	vo, err := s.store.ByID(id)
	if err == nil && vo != nil && vo.MixPath != "" {
		_ = os.Remove(vo.MixPath)
	}
	return s.store.MarkStatus(id, StatusCancelled)
}

// MarkPlayed marks a rendered voice-over as aired so it is not mixed into the
// same track the next time it rotates around. The current track keeps showing
// the voice-over through playback.State.CurrentVoiceOver.
func (s *Service) MarkPlayed(id string) {
	if id == "" {
		return
	}
	if err := s.store.MarkStatus(id, StatusPlayed); err != nil {
		s.log.Warn("Failed to mark voice-over as played: " + err.Error())
	}
}

// Prepare returns a synthetic track whose Path points at the mixed audio file,
// or nil when the track has no voice-over. role is either "current" or "next".
func (s *Service) Prepare(t *track.Track, role, dir string) (*track.Track, *VoiceOver, error) {
	if t == nil {
		return nil, nil, nil
	}

	if rendered, err := s.RenderedFor(t.ID); err != nil {
		return nil, nil, err
	} else if rendered != nil {
		return mixedTrack(t, rendered), rendered, nil
	}

	var vo *VoiceOver
	var err error
	switch role {
	case "current":
		vo, err = s.store.PendingFor(t.ID, TargetCurrent)
	case "next":
		vo, err = s.store.PendingFor(t.ID, TargetNext)
	default:
		return nil, nil, fmt.Errorf("unknown voice-over role: %s", role)
	}
	if err != nil || vo == nil {
		return nil, vo, err
	}

	voices, err := s.tracks.TracksByIDs([]string{vo.VoiceTrackID})
	if err != nil {
		s.log.Warn("Failed to load voice asset: " + err.Error())
		return nil, vo, err
	}
	if len(voices) == 0 {
		_ = s.store.MarkStatus(vo.ID, StatusFailed)
		return nil, vo, fmt.Errorf("voice asset not found")
	}
	voice := voices[0]

	if voice.Duration > t.Duration+0.05 {
		_ = s.store.MarkStatus(vo.ID, StatusFailed)
		return nil, vo, fmt.Errorf("voice clip %q is longer than music track %q", voice.Name, t.Name)
	}

	// Keep the requested start where possible, but never let the voice clip
	// run past the end of the music bed.
	maxStart := math.Max(0, t.Duration-voice.Duration)
	startOffset := clamp(vo.StartOffset, 0, maxStart)
	mixID := fmt.Sprintf("%s-vo%s", t.ID, shortID(vo.ID))
	mixPath := filepath.Join(dir, mixID+".m4a")

	if _, statErr := os.Stat(mixPath); statErr != nil {
		opts := mix.Options{
			BasePath:      t.Path,
			VoicePath:     voice.Path,
			BaseDuration:  t.Duration,
			VoiceDuration: voice.Duration,
			StartOffset:   startOffset,
			FadeIn:        vo.FadeIn,
			FadeOut:       vo.FadeOut,
			DuckDB:        vo.DuckDB,
			BitRate:       s.audioBitRate(),
		}

		if err := s.renderer.RenderVoiceOver(t.Path, voice.Path, mixPath, opts); err != nil {
			_ = s.store.MarkStatus(vo.ID, StatusFailed)
			return nil, vo, err
		}
	}

	if err := s.store.MarkRendered(vo.ID, t.ID, mixPath, mixID); err != nil {
		return nil, vo, err
	}

	vo.Status = StatusRendered
	vo.AppliedTrackID = t.ID
	vo.MixPath = mixPath
	vo.MixID = mixID

	s.log.Info(fmt.Sprintf("Rendered voice-over %q over %q", voice.Name, t.Name))
	return mixedTrack(t, vo), vo, nil
}

// RenderedFor returns the rendered mix for a track, if the file still exists.
func (s *Service) RenderedFor(trackID string) (*VoiceOver, error) {
	vo, err := s.store.RenderedFor(trackID)
	if err != nil || vo == nil {
		return nil, err
	}
	if _, statErr := os.Stat(vo.MixPath); statErr != nil {
		_ = s.store.MarkStatus(vo.ID, StatusFailed)
		return nil, nil
	}
	return vo, nil
}

// HasPendingForNext reports whether a next-track voice-over is waiting to be
// rendered. The playback engine uses this to re-segment the upcoming track.
func (s *Service) HasPendingForNext() bool {
	vo, err := s.store.PendingFor("", TargetNext)
	return err == nil && vo != nil
}

// IsRenderedFor reports whether trackID already has a rendered voice-over mix.
func (s *Service) IsRenderedFor(trackID string) bool {
	vo, err := s.RenderedFor(trackID)
	return err == nil && vo != nil
}

func mixedTrack(t *track.Track, vo *VoiceOver) *track.Track {
	copy := *t
	copy.ID = vo.MixID
	copy.Path = vo.MixPath
	return &copy
}

func shortID(id string) string {
	id = strings.ToLower(id)
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func newID() string {
	return ulid.New()
}

func (s *Service) audioBitRate() int {
	if s.station != nil {
		if info, err := s.station.Info(); err == nil && info.AudioBitRate > 0 {
			return info.AudioBitRate
		}
	}
	return 48
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
