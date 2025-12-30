package face

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"image"

	pigo "github.com/esimov/pigo/core"
)

//go:embed cascade/facefinder
var embeddedCascade []byte

// LocalDetector реализует детекцию лиц с использованием локальной модели pigo.
// Каскадный классификатор встроен в бинарник через embed.
type LocalDetector struct {
	classifier *pigo.Pigo
	cascade    []byte
}

// NewLocalDetector создает новый экземпляр локального детектора.
// Использует встроенный каскадный классификатор, не требует внешних файлов.
func NewLocalDetector() (*LocalDetector, error) {
	if len(embeddedCascade) == 0 {
		return nil, fmt.Errorf("embedded cascade file is empty")
	}

	p := pigo.NewPigo()
	classifier, err := p.Unpack(embeddedCascade)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack cascade: %w", err)
	}

	return &LocalDetector{
		classifier: classifier,
		cascade:    embeddedCascade,
	}, nil
}

// DetectFaces обнаруживает лица на изображении.
func (d *LocalDetector) DetectFaces(ctx context.Context, imageData []byte) ([]Face, error) {
	// Декодируем изображение
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Конвертируем изображение для pigo
	src := pigo.ImgToNRGBA(img)
	pixels := pigo.RgbToGrayscale(src)
	cols, rows := src.Bounds().Dx(), src.Bounds().Dy()

	// Параметры детекции
	params := pigo.CascadeParams{
		MinSize:     20,
		MaxSize:     1000,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,
		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   rows,
			Cols:   cols,
			Dim:    cols,
		},
	}

	// Детектируем лица
	dets := d.classifier.RunCascade(params, 0.0)

	// Кластеризуем детекции для удаления дубликатов
	dets = d.classifier.ClusterDetections(dets, 0.2)

	faces := make([]Face, 0, len(dets))
	for _, det := range dets {
		// Проверяем контекст на отмену
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Pigo возвращает координаты в формате (x, y, scale, q)
		// scale - размер области, q - качество детекции
		if det.Q > 50.0 { // Фильтруем по качеству
			x := det.Col - det.Scale/2
			y := det.Row - det.Scale/2
			size := det.Scale

			faces = append(faces, Face{
				Bounds: image.Rect(
					x, y,
					x+size, y+size,
				),
				Confidence: float64(det.Q),
			})
		}
	}

	return faces, nil
}

