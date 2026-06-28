package attachments

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/image/webp"
)

const (
	MaxUploadBytes    = 5 * 1024 * 1024
	MaxImageDimension = 1920
)

var (
	ErrTooLarge        = errors.New("file too large")
	ErrUnsupportedType = errors.New("unsupported file type")
	ErrEmptyFile       = errors.New("empty file")
)

type ProcessedFile struct {
	Filename    string
	MimeType    string
	SizeBytes   int64
	StorageName string
	Data        []byte
}

func ProcessUpload(originalName string, data []byte) (*ProcessedFile, error) {
	if len(data) == 0 {
		return nil, ErrEmptyFile
	}
	if len(data) > MaxUploadBytes {
		return nil, ErrTooLarge
	}
	kind, err := detectKind(data)
	if err != nil {
		return nil, err
	}
	safeName := sanitizeFilename(originalName)
	if safeName == "" {
		safeName = "attachment"
	}
	switch kind {
	case kindJPEG, kindPNG, kindWebP:
		out, mime, ext, compressErr := compressImage(data, kind)
		if compressErr != nil {
			return nil, fmt.Errorf("compress image: %w", compressErr)
		}
		if len(out) > MaxUploadBytes {
			return nil, ErrTooLarge
		}
		return &ProcessedFile{
			Filename: safeName, MimeType: mime, SizeBytes: int64(len(out)),
			StorageName: uuid.New().String() + ext, Data: out,
		}, nil
	case kindPDF:
		return &ProcessedFile{
			Filename: safeName, MimeType: "application/pdf", SizeBytes: int64(len(data)),
			StorageName: uuid.New().String() + ".pdf", Data: data,
		}, nil
	default:
		return nil, ErrUnsupportedType
	}
}

func WriteFile(dir string, pf *ProcessedFile) (string, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("mkdir attachments: %w", err)
	}
	path := filepath.Join(dir, pf.StorageName)
	if err := os.WriteFile(path, pf.Data, 0o640); err != nil {
		return "", fmt.Errorf("write attachment: %w", err)
	}
	return pf.StorageName, nil
}

func RemoveFile(dir, storagePath string) error {
	if storagePath == "" || strings.Contains(storagePath, "..") {
		return nil
	}
	path := filepath.Join(dir, storagePath)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove attachment file: %w", err)
	}
	return nil
}

func ReadAllLimited(r io.Reader, max int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, fmt.Errorf("read upload: %w", err)
	}
	if int64(len(data)) > max {
		return nil, ErrTooLarge
	}
	return data, nil
}

type fileKind int

const (
	kindUnknown fileKind = iota
	kindJPEG
	kindPNG
	kindWebP
	kindPDF
)

func detectKind(data []byte) (fileKind, error) {
	switch {
	case len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return kindJPEG, nil
	case len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		return kindPNG, nil
	case len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")):
		return kindWebP, nil
	case len(data) >= 5 && bytes.Equal(data[:5], []byte("%PDF-")):
		return kindPDF, nil
	default:
		return kindUnknown, ErrUnsupportedType
	}
}

func compressImage(data []byte, kind fileKind) ([]byte, string, string, error) {
	img, err := decodeImage(data, kind)
	if err != nil {
		return nil, "", "", err
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dstW, dstH := w, h
	if w > MaxImageDimension || h > MaxImageDimension {
		if w >= h {
			dstW = MaxImageDimension
			dstH = h * MaxImageDimension / w
		} else {
			dstH = MaxImageDimension
			dstW = w * MaxImageDimension / h
		}
		if dstW < 1 {
			dstW = 1
		}
		if dstH < 1 {
			dstH = 1
		}
	}
	resized := img
	if dstW != w || dstH != h {
		resized = resizeNearest(img, dstW, dstH)
	}
	var buf bytes.Buffer
	if err := encodeJPEG(&buf, resized); err != nil {
		return nil, "", "", err
	}
	return buf.Bytes(), "image/jpeg", ".jpg", nil
}

func decodeImage(data []byte, kind fileKind) (image.Image, error) {
	r := bytes.NewReader(data)
	if kind == kindWebP {
		img, err := webp.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("decode webp: %w", err)
		}
		return img, nil
	}
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	return img, nil
}

func sanitizeFilename(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == "/" {
		return ""
	}
	var b strings.Builder
	for _, r := range base {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '.', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('_')
		}
	}
	out := b.String()
	if len(out) > 200 {
		out = out[:200]
	}
	return out
}
