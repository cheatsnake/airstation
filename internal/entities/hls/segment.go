package hls

type Segment struct {
	Duration float64
	Path     string
}

func NewSegment(duration float64, path string) *Segment {
	return &Segment{
		Duration: duration,
		Path:     path,
	}
}
