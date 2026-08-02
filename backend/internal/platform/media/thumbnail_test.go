package media_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/platform/media"
)

func TestGenerateJPEGThumbnail(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 80, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	thumb, err := media.GenerateJPEGThumbnail(bytes.NewReader(buf.Bytes()), "image/jpeg")
	if err != nil {
		t.Fatal(err)
	}
	if len(thumb) < 100 {
		t.Fatalf("thumbnail too small: %d", len(thumb))
	}
	out, err := jpeg.Decode(bytes.NewReader(thumb))
	if err != nil {
		t.Fatal(err)
	}
	b := out.Bounds()
	if b.Dx() > media.MaxThumbSide || b.Dy() > media.MaxThumbSide {
		t.Fatalf("thumb size %dx%d exceeds max", b.Dx(), b.Dy())
	}
}
