// Package mix builds and renders the audio graph used for RJ voice-overs.
//
// A voice-over is not a separate audio stream: it is mixed into the music bed
// before HLS segmentation, so listeners hear one continuous programme. The
// music is ducked underneath the voice and both sides use the fade durations
// chosen by the operator.
package mix

import (
	"fmt"
	"math"
)

// Options describes a voice-over mix.
type Options struct {
	BasePath      string
	VoicePath     string
	BaseDuration  float64
	VoiceDuration float64
	StartOffset   float64
	FadeIn        float64
	FadeOut       float64
	DuckDB        float64
	BitRate       int
}

// Renderer turns a music bed and a voice clip into a single audio file.
type Renderer interface {
	RenderVoiceOver(basePath, voicePath, outPath string, opts Options) error
}

// BuildFilter validates opts and returns the FFmpeg filter_complex graph.
//
// The graph keeps the music playing for the whole base track, delays the voice
// to StartOffset and fades it in/out. The music is ducked with a time-based
// envelope: it ramps down while the voice fades in, holds at the requested
// DuckDB level, then ramps back up while the voice fades out. Using an envelope
// (instead of a level-triggered sidechain compressor) makes the duck depth
// exactly match the value the operator chose and guarantees the mix keeps the
// full duration of the music bed.
func BuildFilter(opts Options) (string, error) {
	if opts.BaseDuration <= 0 {
		return "", fmt.Errorf("base duration must be positive")
	}
	if opts.VoiceDuration <= 0 {
		return "", fmt.Errorf("voice duration must be positive")
	}
	if opts.VoiceDuration > opts.BaseDuration+0.05 {
		return "", fmt.Errorf("voice clip (%.2fs) is longer than the music bed (%.2fs)", opts.VoiceDuration, opts.BaseDuration)
	}

	start := clamp(opts.StartOffset, 0, math.Max(0, opts.BaseDuration-opts.VoiceDuration))
	fadeIn := clamp(opts.FadeIn, 0, opts.VoiceDuration/2)
	fadeOut := clamp(opts.FadeOut, 0, opts.VoiceDuration/2)
	if fadeIn+fadeOut > opts.VoiceDuration {
		fadeOut = math.Max(0, opts.VoiceDuration-fadeIn)
	}
	duckDB := clamp(opts.DuckDB, -30, 0)

	voiceEnd := start + opts.VoiceDuration
	outFadeStart := math.Max(0, opts.VoiceDuration-fadeOut)
	delayMS := int(math.Round(start * 1000))

	var filter string
	filter += "[0:a]aformat=sample_fmts=fltp:sample_rates=44100:channel_layouts=stereo"

	if duckDB < 0 {
		gain := math.Pow(10, duckDB/20)
		filter += fmt.Sprintf(",volume=volume='%s':eval=frame", duckEnvelope(gain, start, voiceEnd, fadeIn, fadeOut))
	}
	filter += "[music];"

	filter += "[1:a]aformat=sample_fmts=fltp:sample_rates=44100:channel_layouts=stereo"
	if fadeIn > 0 {
		filter += fmt.Sprintf(",afade=t=in:st=0:d=%.3f", fadeIn)
	}
	if fadeOut > 0 {
		filter += fmt.Sprintf(",afade=t=out:st=%.3f:d=%.3f", outFadeStart, fadeOut)
	}
	if delayMS > 0 {
		filter += fmt.Sprintf(",adelay=delays=%d:all=1", delayMS)
	}
	// apad makes the (short) voice stream last as long as the music bed, so
	// amix does not truncate the programme when the voice clip ends.
	filter += fmt.Sprintf(",apad=whole_dur=%.3f[voice];", opts.BaseDuration)

	filter += "[music][voice]amix=inputs=2:duration=first:dropout_transition=0:normalize=0,alimiter=limit=0.95:level=false[out]"

	return filter, nil
}

// duckEnvelope builds the volume expression used to fade the music down and up
// around the voice clip. The result is 1 outside the voice window, ramps to
// gain over fadeIn, stays there until fadeOut starts, then ramps back to 1.
func duckEnvelope(gain, start, end, fadeIn, fadeOut float64) string {
	rampIn := "1"
	if fadeIn > 0 {
		rampIn = fmt.Sprintf("max(0,min(1,(t-%.3f)/%.3f))", start, fadeIn)
	} else if start > 0 {
		rampIn = fmt.Sprintf("if(lt(t,%.3f),0,1)", start)
	}

	rampOut := "1"
	if fadeOut > 0 {
		rampOut = fmt.Sprintf("max(0,min(1,(%.3f-t)/%.3f))", end, fadeOut)
	} else {
		rampOut = fmt.Sprintf("if(lt(t,%.3f),1,0)", end)
	}

	return fmt.Sprintf("1-(1-%.4f)*min(%s,%s)", gain, rampIn, rampOut)
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
