package ffmpeg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

func HLSPlaylist(filePath, outDir, segName string, segDur int) error {
	hlsTime := strconv.Itoa(segDur)
	hlsSegName := fmt.Sprintf("%s/%s", outDir, segName) + "%d.ts"
	hlsPlName := fmt.Sprintf("%s/%s", outDir, segName) + ".m3u8"

	cmd := exec.Command(
		ffmpegBin,
		"-i", filePath,
		"-codec:", "copy",
		"-start_number", "0",
		"-hls_time", hlsTime,
		"-hls_playlist_type", "event",
		"-hls_segment_filename", hlsSegName,
		hlsPlName,
	)

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running ffmpeg: %v\n%s", err, errBuf.String())
	}

	return nil
}

// TrackMetadata return track's duration in seconds and bitrate in kbps
func TrackMetadata(filePath string) (float64, int, error) {
	cmd := exec.Command(
		ffprobeBin,
		"-i", filePath,
		"-v", "error",
		"-show_entries", "format=duration,bit_rate",
		"-of", "json",
	)

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()
	if err != nil {
		return 0, 0, fmt.Errorf("error running ffprobe: %v\n%s", err, outBuf.String())
	}

	var metadata metadata

	err = json.Unmarshal(outBuf.Bytes(), &metadata)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing ffprobe output: %v", err)
	}

	duration, err := strconv.ParseFloat(metadata.Format.Duration, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing duration: %v", err)
	}

	bitrate, err := strconv.ParseInt(metadata.Format.Bitrate, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing bitrate: %v", err)
	}

	return duration, int(bitrate / 1000), nil
}
