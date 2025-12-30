package service

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/disintegration/imaging"
)

type ImageService struct{}

func NewImageService() *ImageService {
	return &ImageService{}
}

func (s *ImageService) CreateGrid(imagePaths []string, outputPath string) error {
	const gridCols = 5
	const gridRows = 4
	const thumbWidth = 200
	const thumbHeight = 120

	// We need up to 20 images
	numImages := len(imagePaths)
	if numImages > 20 {
		numImages = 20
	}

	dst := image.NewRGBA(image.Rect(0, 0, gridCols*thumbWidth, gridRows*thumbHeight))

	for i := 0; i < numImages; i++ {
		img, err := imaging.Open(imagePaths[i])
		if err != nil {
			return fmt.Errorf("failed to open image %s: %v", imagePaths[i], err)
		}

		resized := imaging.Fill(img, thumbWidth, thumbHeight, imaging.Center, imaging.Lanczos)

		col := i % gridCols
		row := i / gridCols

		rect := image.Rect(col*thumbWidth, row*thumbHeight, (col+1)*thumbWidth, (row+1)*thumbHeight)
		draw.Draw(dst, rect, resized, image.Point{0, 0}, draw.Src)
	}

	if err := imaging.Save(dst, outputPath); err != nil {
		return fmt.Errorf("failed to save grid image: %v", err)
	}

	return nil
}
