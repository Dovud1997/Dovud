package media

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
	"strings"

	"golang.org/x/image/draw"
)

const MaxThumbSide = 256

func IsImageMIME(mime string) bool {
	m := strings.ToLower(strings.TrimSpace(mime))
	return strings.HasPrefix(m, "image/jpeg") || strings.HasPrefix(m, "image/jpg") ||
		strings.HasPrefix(m, "image/png") || m == "image/pjpeg"
}

func ThumbnailObjectKey(objectKey string) string {
	return objectKey + ".thumb.jpg"
}

// GenerateJPEGThumbnail decodes JPEG/PNG and returns a JPEG thumbnail.
func GenerateJPEGThumbnail(r io.Reader, mime string) ([]byte, error) {
	var img image.Image
	var err error
	m := strings.ToLower(mime)
	switch {
	case strings.Contains(m, "png"):
		img, err = png.Decode(r)
	default:
		img, err = jpeg.Decode(r)
	}
	if err != nil {
		// try generic decode
		if seeker, ok := r.(io.Seeker); ok {
			_, _ = seeker.Seek(0, io.SeekStart)
		}
		img, _, err = image.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("decode image: %w", err)
		}
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid image size")
	}
	scale := float64(MaxThumbSide) / float64(w)
	if h > w {
		scale = float64(MaxThumbSide) / float64(h)
	}
	if scale > 1 {
		scale = 1
	}
	tw := int(float64(w) * scale)
	th := int(float64(h) * scale)
	if tw < 1 {
		tw = 1
	}
	if th < 1 {
		th = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, tw, th))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 85}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
