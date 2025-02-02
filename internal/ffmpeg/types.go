package ffmpeg

// AudioMetadata holds metadata information about an audio file.
type AudioMetadata struct {
	Duration     float64 // The total duration of the audio file in seconds.
	BitRate      int     // The bit rate of the audio file in kbps (kilobits per second).
	CodecName    string  // The name of the codec used for encoding the audio.
	SampleRate   int     // The sample rate of the audio file in Hz (hertz).
	ChannelCount int     // The number of audio channels (e.g., 1 for mono, 2 for stereo).
}

type rawAudioMetadata struct {
	Format struct {
		Duration string `json:"duration"`
		BitRate  string `json:"bit_rate"`
	} `json:"format"`
	Streams []struct {
		CodecName  string `json:"codec_name"`
		SampleRate string `json:"sample_rate"`
		Channels   int    `json:"channels"`
	} `json:"streams"`
}
