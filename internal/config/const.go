package config

import (
	"path/filepath"
)

var (
	defaultTracksDir = filepath.Join("static", "tracks")
	defaultTmpDir    = filepath.Join("static", "tmp")
)

const defaultHLSSegmentDuration = 5 // seconds
