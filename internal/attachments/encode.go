package attachments

import (
	"bytes"
	"image"
	"image/jpeg"
)

const jpegQuality = 85

func encodeJPEG(w *bytes.Buffer, img image.Image) error {
	return jpeg.Encode(w, img, &jpeg.Options{Quality: jpegQuality})
}
