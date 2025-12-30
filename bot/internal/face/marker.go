package face

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// MarkFaces рисует прямоугольники вокруг обнаруженных лиц на изображении.
func MarkFaces(img image.Image, faces []Face) (image.Image, error) {
	// Создаем копию изображения для рисования
	bounds := img.Bounds()
	marked := image.NewRGBA(bounds)
	draw.Draw(marked, bounds, img, bounds.Min, draw.Src)

	// Цвет для прямоугольников (красный)
	rectColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	// Толщина линии
	lineWidth := 3

	// Рисуем прямоугольники вокруг каждого лица
	for i, face := range faces {
		drawRectangle(marked, face.Bounds, rectColor, lineWidth)

		// Добавляем номер лица в верхний левый угол прямоугольника
		label := fmt.Sprintf("№%d", i+1)
		drawLabel(marked, face.Bounds.Min.X, face.Bounds.Min.Y-5, label)
	}

	return marked, nil
}

// drawRectangle рисует прямоугольник на изображении.
func drawRectangle(img *image.RGBA, rect image.Rectangle, c color.Color, width int) {
	// Верхняя линия
	for x := rect.Min.X; x <= rect.Max.X; x++ {
		for w := 0; w < width; w++ {
			if y := rect.Min.Y + w; y >= 0 && y < img.Bounds().Dy() && x >= 0 && x < img.Bounds().Dx() {
				img.Set(x, y, c)
			}
		}
	}

	// Нижняя линия
	for x := rect.Min.X; x <= rect.Max.X; x++ {
		for w := 0; w < width; w++ {
			if y := rect.Max.Y - w; y >= 0 && y < img.Bounds().Dy() && x >= 0 && x < img.Bounds().Dx() {
				img.Set(x, y, c)
			}
		}
	}

	// Левая линия
	for y := rect.Min.Y; y <= rect.Max.Y; y++ {
		for w := 0; w < width; w++ {
			if x := rect.Min.X + w; x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, c)
			}
		}
	}

	// Правая линия
	for y := rect.Min.Y; y <= rect.Max.Y; y++ {
		for w := 0; w < width; w++ {
			if x := rect.Max.X - w; x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, c)
			}
		}
	}
}

// drawLabel рисует текст на изображении.
func drawLabel(img *image.RGBA, x, y int, label string) {
	point := fixed.Point26_6{
		X: fixed.Int26_6(x * 64),
		Y: fixed.Int26_6(y * 64),
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{R: 255, G: 255, B: 255, A: 255}),
		Face: basicfont.Face7x13,
		Dot:  point,
	}

	// Рисуем черный контур для лучшей читаемости
	black := image.NewUniform(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx != 0 || dy != 0 {
				d.Src = black
				d.Dot = fixed.Point26_6{
					X: fixed.Int26_6((x + dx) * 64),
					Y: fixed.Int26_6((y + dy) * 64),
				}
				d.DrawString(label)
			}
		}
	}

	// Рисуем белый текст поверх
	d.Src = image.NewUniform(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	d.Dot = point
	d.DrawString(label)
}

