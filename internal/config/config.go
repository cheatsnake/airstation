package config

type Config struct {
	TracksDir          string
	TmpDir             string
	HLSSegmentDuration int
	PossibleBitrates   []int
}

func New() *Config {
	return &Config{
		TracksDir:          defaultTracksDir,
		TmpDir:             defaultTmpDir,
		HLSSegmentDuration: defaultHLSSegmentDuration,
		PossibleBitrates:   []int{320},
	}
}
