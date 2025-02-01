package ffmpeg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/cheatsnake/airstation/internal/tools/fs"
)

type CLI struct{}

func NewCLI() *CLI {
	return &CLI{}
}

func (cli *CLI) MakeHLSPlaylist(trackPath, outDir, segName string, segDur int) error {
	if err := fs.FileExists(trackPath); err != nil {
		return err
	}

	hlsTime := strconv.Itoa(segDur)
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

// AudioMetadata return track's duration in seconds and bitrate in kbps
func (cli *CLI) AudioMetadata(filePath string) (float64, int, error) {
	if err := fs.FileExists(filePath); err != nil {
		return 0, 0, err
	}

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
		return 0, 0, fmt.Errorf("metadata retrieve failed: %v\n%s", err, outBuf.String())
	}

	var metadata metadata

	err = json.Unmarshal(outBuf.Bytes(), &metadata)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing metadata retrieve failed: %v", err)
	}

	duration, err := strconv.ParseFloat(metadata.Format.Duration, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing metadata duration failed: %v", err)
	}

	bitrate, err := strconv.ParseInt(metadata.Format.Bitrate, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing metadata bitrate failed: %v", err)
	}

	return duration, int(bitrate / 1000), nil
}

func (cli *CLI) PadAudio(filePath string, padLength float64, bitrate int) error {
	if err := fs.FileExists(filePath); err != nil {
		return err
	}

	dir, name := filepath.Split(filePath)
	tmpFilePath := filepath.Join(dir, "xtmp-"+name)

	cmd := exec.Command(
		ffmpegBin,
		"-i", filePath,
		"-f", "lavfi",
		"-t", strconv.FormatFloat(padLength, 'f', 1, 64),
		"-i", "anullsrc=r=44100:cl=stereo",
		"-filter_complex", "[0:a][1:a]concat=n=2:v=0:a=1",
		"-b:a", strconv.Itoa(bitrate)+"k",
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

	fs.DeleteFile(tmpFilePath)

	return nil
}

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
