// Package voiceover manages RJ voice-over requests and mixes them into the
// music bed before HLS segmentation.
package voiceover

import "github.com/cheatsnake/airstation/internal/track"

const (
	TargetNext    = "next"
	TargetCurrent = "current"

	StatusPending   = "pending"
	StatusRendered  = "rendered"
	StatusPlayed    = "played"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
)

// VoiceOver is a request to place a voice clip over a music track.
type VoiceOver struct {
	ID             string  `json:"id"`
	VoiceTrackID   string  `json:"voiceTrackId"`
	VoiceName      string  `json:"voiceName,omitempty"`
	Target         string  `json:"target"`
	TargetTrackID  string  `json:"targetTrackId,omitempty"`
	FadeIn         float64 `json:"fadeIn"`
	FadeOut        float64 `json:"fadeOut"`
	DuckDB         float64 `json:"duckDb"`
	StartOffset    float64 `json:"startOffset"`
	Status         string  `json:"status"`
	AppliedTrackID string  `json:"appliedTrackId,omitempty"`
	MixPath        string  `json:"-"`
	MixID          string  `json:"-"`
	CreatedAt      int64   `json:"createdAt"`
}

type Store interface {
	Add(vo *VoiceOver) error
	List() ([]*VoiceOver, error)
	ByID(id string) (*VoiceOver, error)
	PendingFor(trackID, target string) (*VoiceOver, error)
	RenderedFor(trackID string) (*VoiceOver, error)
	MarkRendered(id, appliedTrackID, mixPath, mixID string) error
	MarkStatus(id, status string) error
}

// TrackFinder resolves the voice asset metadata.
type TrackFinder interface {
	TracksByIDs(IDs []string) ([]*track.Track, error)
}
