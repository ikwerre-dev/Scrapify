package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type MediaService struct{}

func NewMediaService() *MediaService {
	return &MediaService{}
}

func (s *MediaService) GenerateSnapshots(videoPath string, outputDir string) ([]string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	// Capture frames every 5 seconds
	// -preset ultrafast for speed
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vf", "fps=1/5", "-preset", "ultrafast", filepath.Join(outputDir, "thumb%03d.jpg"))
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg snapshots failed: %v", err)
	}

	files, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, err
	}

	var snapshots []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".jpg" {
			snapshots = append(snapshots, filepath.Join(outputDir, file.Name()))
		}
	}
	return snapshots, nil
}

func (s *MediaService) ExtractAudio(videoPath string, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vn", "-acodec", "libmp3lame", "-y", "-preset", "ultrafast", outputPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg audio extraction failed: %v", err)
	}
	return nil
}
