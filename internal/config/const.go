package config

import "path"

var (
	defaultTracksDir = path.Join("static", "tracks")
	defaultTmpDir    = path.Join("static", "tmp")
)

const defaultHLSSegmentDuration = 5 // seconds
