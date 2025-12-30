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

func (s *ImageService) CreateGrids(imagePaths []string, outputPrefix string) ([]string, error) {
	const gridCols = 5
	const gridRows = 4
	const imagesPerGrid = gridCols * gridRows
	const thumbWidth = 200
	const thumbHeight = 120

	var gridPaths []string
	numImages := len(imagePaths)

	for i := 0; i < numImages; i += imagesPerGrid {
		end := i + imagesPerGrid
		if end > numImages {
			end = numImages
		}

		gridIdx := i/imagesPerGrid + 1
		gridPath := fmt.Sprintf("%s_grid_%d.jpg", outputPrefix, gridIdx)

		dst := image.NewRGBA(image.Rect(0, 0, gridCols*thumbWidth, gridRows*thumbHeight))

		batch := imagePaths[i:end]
		for j, imgPath := range batch {
			img, err := imaging.Open(imgPath)
			if err != nil {
				return nil, fmt.Errorf("failed to open image %s: %v", imgPath, err)
			}

			resized := imaging.Fill(img, thumbWidth, thumbHeight, imaging.Center, imaging.Lanczos)

			col := j % gridCols
			row := j / gridCols

			rect := image.Rect(col*thumbWidth, row*thumbHeight, (col+1)*thumbWidth, (row+1)*thumbHeight)
			draw.Draw(dst, rect, resized, image.Point{0, 0}, draw.Src)
		}

		if err := imaging.Save(dst, gridPath); err != nil {
			return nil, fmt.Errorf("failed to save grid image: %v", err)
		}
		gridPaths = append(gridPaths, gridPath)
	}

	return gridPaths, nil
}
