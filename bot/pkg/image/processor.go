package image

import (
	"bytes"
	"image"
	"image/jpeg"
)

// EncodeJPEG кодирует изображение в JPEG формат.
func EncodeJPEG(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeImage декодирует изображение из байтов.
func DecodeImage(data []byte) (image.Image, string, error) {
	reader := bytes.NewReader(data)
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", err
	}
	return img, format, nil
}

// ConvertToJPEG конвертирует изображение в JPEG формат.
func ConvertToJPEG(img image.Image) ([]byte, error) {
	return EncodeJPEG(img, 90)
}

