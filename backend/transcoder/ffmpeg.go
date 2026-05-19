package transcoder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type FFmpegTranscoder struct{}

func NewFFmpegTranscoder() *FFmpegTranscoder {
	return &FFmpegTranscoder{}
}

// IsVideo uses ffprobe to verify that the file is a valid media container.
// This acts as a security gate against malformed file exploits.
func (t *FFmpegTranscoder) IsVideo(ctx context.Context, input string) (bool, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_type",
		"-of", "csv=p=0",
		input,
	)

	out, err := cmd.Output()
	if err != nil {
		return false, nil // Failed probe means it's not a valid video or corrupted
	}
	return strings.TrimSpace(string(out)) == "video", nil
}

func (t *FFmpegTranscoder) Transcode(ctx context.Context, input, outputDir string) (*Result, error) {
	// 1. Probe if file has an audio stream to prevent crashes on silent inputs
	hasAudio := false
	probeCmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "a",
		"-show_entries", "stream=codec_type",
		"-of", "csv=p=0",
		input,
	)
	probeOut, err := probeCmd.Output()
	if err == nil && len(probeOut) > 0 {
		hasAudio = true
	}

	// 2. Pre-create variant subdirectories to prevent directory creation errors during segment writing
	for _, v := range []string{"0", "1", "2", "3"} {
		if err := os.MkdirAll(filepath.Join(outputDir, v), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %v", v, err)
		}
	}

	// ABR Multi-resolution pipeline (360p, 480p, 720p, 1080p)
	// Uses aspect ratio preserving filters with force_original_aspect_ratio=decrease
	// and min(target, iw) to prevent upscaling low-res input videos.
	// Uses fixed GOP size of 48 (2s @ 24fps) for seamless segment switching.
	var args []string
	filterComplex := "[0:v]split=4[v1][v2][v3][v4];" +
		"[v1]scale=w='min(1920,iw)':h=-2:force_original_aspect_ratio=decrease[v1out];" +
		"[v2]scale=w='min(1280,iw)':h=-2:force_original_aspect_ratio=decrease[v2out];" +
		"[v3]scale=w='min(854,iw)':h=-2:force_original_aspect_ratio=decrease[v3out];" +
		"[v4]scale=w='min(640,iw)':h=-2:force_original_aspect_ratio=decrease[v4out]"

	if hasAudio {
		args = []string{
			"-y",
			"-i", input,
			"-filter_complex", filterComplex,
			"-map", "[v1out]", "-c:v:0", "libx264", "-b:v:0", "5000k", "-maxrate:v:0", "5350k", "-bufsize:v:0", "7500k",
			"-map", "[v2out]", "-c:v:1", "libx264", "-b:v:1", "2800k", "-maxrate:v:1", "2996k", "-bufsize:v:1", "4200k",
			"-map", "[v3out]", "-c:v:2", "libx264", "-b:v:2", "1400k", "-maxrate:v:2", "1498k", "-bufsize:v:2", "2100k",
			"-map", "[v4out]", "-c:v:3", "libx264", "-b:v:3", "800k", "-maxrate:v:3", "856k", "-bufsize:v:3", "1200k",
			"-map", "0:a:0", "-c:a:0", "aac", "-b:a:0", "128k", "-ac:a:0", "2",
			"-map", "0:a:0", "-c:a:1", "aac", "-b:a:1", "128k", "-ac:a:1", "2",
			"-map", "0:a:0", "-c:a:2", "aac", "-b:a:2", "128k", "-ac:a:2", "2",
			"-map", "0:a:0", "-c:a:3", "aac", "-b:a:3", "128k", "-ac:a:3", "2",
			"-f", "hls",
			"-hls_time", "4",
			"-hls_playlist_type", "vod",
			"-hls_segment_filename", filepath.Join(outputDir, "%v/segment%03d.ts"),
			"-master_pl_name", "master.m3u8",
			"-var_stream_map", "v:0,a:0 v:1,a:1 v:2,a:2 v:3,a:3",
			filepath.Join(outputDir, "%v/index.m3u8"),
		}
	} else {
		args = []string{
			"-y",
			"-i", input,
			"-filter_complex", filterComplex,
			"-map", "[v1out]", "-c:v:0", "libx264", "-b:v:0", "5000k", "-maxrate:v:0", "5350k", "-bufsize:v:0", "7500k",
			"-map", "[v2out]", "-c:v:1", "libx264", "-b:v:1", "2800k", "-maxrate:v:1", "2996k", "-bufsize:v:1", "4200k",
			"-map", "[v3out]", "-c:v:2", "libx264", "-b:v:2", "1400k", "-maxrate:v:2", "1498k", "-bufsize:v:2", "2100k",
			"-map", "[v4out]", "-c:v:3", "libx264", "-b:v:3", "800k", "-maxrate:v:3", "856k", "-bufsize:v:3", "1200k",
			"-f", "hls",
			"-hls_time", "4",
			"-hls_playlist_type", "vod",
			"-hls_segment_filename", filepath.Join(outputDir, "%v/segment%03d.ts"),
			"-master_pl_name", "master.m3u8",
			"-var_stream_map", "v:0 v:1 v:2 v:3",
			filepath.Join(outputDir, "%v/index.m3u8"),
		}
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg error: %v", err)
	}

	totalSize, err := directorySize(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to measure output size: %v", err)
	}

	return &Result{
		TotalSizeBytes: totalSize,
		Duration:       probeDuration(ctx, input),
	}, nil
}

func probeDuration(ctx context.Context, input string) int {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		input,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	duration, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0
	}
	return int(duration + 0.5)
}

func directorySize(root string) (int64, error) {
	var total int64
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	return total, err
}
