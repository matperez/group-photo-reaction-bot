package face

import (
	"context"
	"image"
)

// Face представляет обнаруженное лицо на изображении.
type Face struct {
	Bounds     image.Rectangle
	Confidence float64
}

// Detector определяет интерфейс для детекции лиц на изображениях.
type Detector interface {
	// DetectFaces обнаруживает лица на изображении и возвращает список найденных лиц.
	DetectFaces(ctx context.Context, imageData []byte) ([]Face, error)
}

