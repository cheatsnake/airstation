// Package ffmpeg provides a CLI wrapper for executing FFmpeg and FFprobe commands,
// enabling audio processing functionalities.
package ffmpeg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/cheatsnake/airstation/internal/tools/fs"
	"github.com/cheatsnake/airstation/internal/tools/ulid"
)

// CLI represents a command-line interface for interacting with FFmpeg and FFprobe.
type CLI struct{}

// NewCLI creates and returns a new instance of CLI.
func NewCLI() *CLI {
	return &CLI{}
}

// MakeHLSPlaylist converts an audio track into an HLS (HTTP Live Streaming) playlist with segmented files.
// It generates a playlist (.m3u8) and segment files (.ts) in the specified output directory.
//
// Parameters:
//   - trackPath: The path to the source audio file to be converted into HLS format.
//   - outDir: The directory where the HLS playlist and segments will be stored.
//   - segName: The base name for the segment files, which will be suffixed with an index.
//   - segDuration: The duration (in seconds) of each segment.
//
// Returns:
//   - An error if the input file does not exist, or if the HLS generation process fails.
func (cli *CLI) MakeHLSPlaylist(trackPath, outDir, segName string, segDuration int) error {
	if err := fs.FileExists(trackPath); err != nil {
		return err
	}

	hlsTime := strconv.Itoa(segDuration)
	hlsSegName := fmt.Sprintf("%s/%s", outDir, segName) + "%d.ts"
	hlsPlName := fmt.Sprintf("%s/%s", outDir, segName) + ".m3u8"

	cmd := exec.Command(
		ffmpegBin,
		"-i", trackPath,
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
		return fmt.Errorf("hls playlist generation failed: %v\n%s", err, errBuf.String())
	}

	return nil
}

// AudioMetadata extracts and returns metadata information from the specified audio file.
// It uses ffprobe to retrieve details such as duration, bit rate, codec name, sample rate, and channel count.
//
// Parameters:
//   - filePath: The path to the audio file whose metadata is to be retrieved.
//
// Returns:
//   - AudioMetadata: A struct containing the extracted metadata (duration, bit rate, codec, sample rate, and channels).
//   - An error if the file does not exist, ffprobe execution fails, or metadata parsing encounters an issue.
func (cli *CLI) AudioMetadata(filePath string) (AudioMetadata, error) {
	metadata := AudioMetadata{}

	if err := fs.FileExists(filePath); err != nil {
		return metadata, err
	}

	cmd := exec.Command(
		ffprobeBin,
		"-i", filePath,
		"-v", "error",
		"-show_entries", "format=duration,bit_rate:stream=codec_name,sample_rate,channels",
		"-of", "json",
	)

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()
	if err != nil {
		return metadata, fmt.Errorf("metadata retrieve failed: %v\n%s", err, outBuf.String())
	}

	var rawMetadata rawAudioMetadata

	if err = json.Unmarshal(outBuf.Bytes(), &rawMetadata); err != nil {
		return metadata, fmt.Errorf("parsing metadata retrieve failed: %v", err)
	}

	duration, err := strconv.ParseFloat(rawMetadata.Format.Duration, 64)
	if err != nil {
		return metadata, fmt.Errorf("parsing metadata duration failed: %v", err)
	}

	bitRate, err := strconv.Atoi(rawMetadata.Format.BitRate)
	if err != nil {
		return metadata, fmt.Errorf("parsing metadata bitrate failed: %v", err)
	}

	if len(rawMetadata.Streams) == 0 {
		return metadata, fmt.Errorf("couldn't extract all metadata")
	}

	channels := rawMetadata.Streams[0].Channels
	codecName := rawMetadata.Streams[0].CodecName
	sampleRate, err := strconv.Atoi(rawMetadata.Streams[0].SampleRate)
	if err != nil {
		return metadata, fmt.Errorf("parsing metadata sample rate failed: %v", err)
	}

	metadata.Duration = duration
	metadata.BitRate = int(bitRate / 1000)
	metadata.ChannelCount = channels
	metadata.CodecName = codecName
	metadata.SampleRate = sampleRate

	return metadata, nil
}

// PadAudio appends a period of silence to the given audio file, extending its duration by padDuration seconds.
// It generates a silence file based on the provided audio metadata and concatenates it with the original file.
//
// Parameters:
//   - filePath: The path to the audio file to be padded.
//   - padDuration: The duration of silence (in seconds) to be added at the end of the audio.
//   - meta: The metadata of the original audio file, including codec, bit rate, sample rate, and channel count.
//
// Returns:
//   - An error if the file does not exist, silence generation fails, padding fails, or file operations encounter an issue.
func (cli *CLI) PadAudio(filePath string, padDuration float64, meta AudioMetadata) error {
	if err := fs.FileExists(filePath); err != nil {
		return err
	}

	dir, name := filepath.Split(filePath)
	tmpFilePath := filepath.Join(dir, "xtmp-"+name)
	silenceFile := filepath.Join(dir, ulid.New()+"."+meta.CodecName)

	if err := cli.generateSilence(padDuration, meta.BitRate, meta.SampleRate, meta.ChannelCount, silenceFile); err != nil {
		return err
	}

	cmd := exec.Command(
		ffmpegBin,
		"-i", fmt.Sprintf("concat:%s|%s", filePath, silenceFile),
		"-c", "copy",
		tmpFilePath,
		"-y",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("padding audio failed: %v\nOutput: %s", err, string(output))
	}

	err = fs.RenameFile(tmpFilePath, filePath)
	if err != nil {
		return err
	}

	go fs.DeleteFile(tmpFilePath)
	go fs.DeleteFile(silenceFile)

	return nil
}

// TrimAudio trims the audio file at the specified filePath to the given totalDuration.
// It creates a temporary file, processes the trimming using ffmpeg, and replaces the original file.
//
// Parameters:
//   - filePath: The path to the audio file to be trimmed.
//   - totalDuration: The desired duration of the trimmed audio in seconds.
//
// Returns:
//   - An error if the file does not exist, the trimming process fails, or file operations encounter an issue.
func (cli *CLI) TrimAudio(filePath string, totalDuration float64) error {
	if err := fs.FileExists(filePath); err != nil {
		return err
	}

	dir, name := filepath.Split(filePath)
	tmpFilePath := filepath.Join(dir, "xtmp-"+name)

	cmd := exec.Command(
		ffmpegBin,
		"-i", filePath,
		"-t", strconv.FormatFloat(totalDuration, 'f', 1, 64),
		"-c:a", "copy",
		tmpFilePath,
		"-y",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("triming audio failed: %v\nOutput: %s", err, string(output))
	}

	err = fs.RenameFile(tmpFilePath, filePath)
	if err != nil {
		return err
	}

	fs.DeleteFile(tmpFilePath)

	return nil
}

// generateSilence generates a silent audio file with the specified duration, bitrate, sample rate,
// and number of channels. The resulting audio file is saved to the provided file path.
func (cli *CLI) generateSilence(duration float64, bitRate, sampleRate, channelCount int, filePath string) error {
	layout := "stereo"
	if channelCount == 1 {
		layout = "mono"
	}

	cmd := exec.Command(
		ffmpegBin,
		"-f", "lavfi",
		"-i", "anullsrc=r="+strconv.Itoa(sampleRate)+":cl="+layout,
		"-t", strconv.FormatFloat(duration, 'f', 1, 64),
		"-b:a", strconv.Itoa(bitRate)+"k",
		"-ar", strconv.Itoa(sampleRate),
		"-ac", strconv.Itoa(channelCount),
		filePath,
		"-y",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("generating silence audio failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}
