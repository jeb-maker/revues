package attachments_test

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/jeb-maker/revues/internal/attachments"
)

func TestProcessUpload_RejectsInvalidType(t *testing.T) {
	t.Parallel()
	_, err := attachments.ProcessUpload("evil.exe", []byte{0x4D, 0x5A, 0x90, 0x00})
	if !errors.Is(err, attachments.ErrUnsupportedType) {
		t.Fatalf("err = %v", err)
	}
}

func TestProcessUpload_RejectsTooLarge(t *testing.T) {
	t.Parallel()
	data := make([]byte, attachments.MaxUploadBytes+1)
	data[0], data[1], data[2] = 0xFF, 0xD8, 0xFF
	_, err := attachments.ProcessUpload("big.jpg", data)
	if !errors.Is(err, attachments.ErrTooLarge) {
		t.Fatalf("err = %v", err)
	}
}

func TestProcessUpload_CompressesJPEG(t *testing.T) {
	t.Parallel()
	src := image.NewRGBA(image.Rect(0, 0, 2400, 1200))
	var raw bytes.Buffer
	if err := jpeg.Encode(&raw, src, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatalf("encode: %v", err)
	}
	pf, err := attachments.ProcessUpload("photo.jpg", raw.Bytes())
	if err != nil {
		t.Fatalf("ProcessUpload(): %v", err)
	}
	img, _, err := image.Decode(bytes.NewReader(pf.Data))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	b := img.Bounds()
	if b.Dx() > attachments.MaxImageDimension || b.Dy() > attachments.MaxImageDimension {
		t.Fatalf("dimensions %dx%d too large", b.Dx(), b.Dy())
	}
}

func TestProcessUpload_AcceptsPDF(t *testing.T) {
	t.Parallel()
	data := []byte("%PDF-1.4 test")
	pf, err := attachments.ProcessUpload("doc.pdf", data)
	if err != nil {
		t.Fatalf("ProcessUpload(): %v", err)
	}
	if !bytes.Equal(pf.Data, data) {
		t.Fatal("pdf modified")
	}
}

func TestProcessUpload_AcceptsPNG(t *testing.T) {
	t.Parallel()
	src := image.NewRGBA(image.Rect(0, 0, 64, 64))
	var raw bytes.Buffer
	if err := png.Encode(&raw, src); err != nil {
		t.Fatalf("encode: %v", err)
	}
	pf, err := attachments.ProcessUpload("shot.png", raw.Bytes())
	if err != nil {
		t.Fatalf("ProcessUpload(): %v", err)
	}
	if pf.MimeType != "image/jpeg" {
		t.Fatalf("mime = %q", pf.MimeType)
	}
}
