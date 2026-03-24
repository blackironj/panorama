package conv

import (
	"image"
	"image/color"
	"testing"
)

func newUniformRGBA(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := range w {
		for y := range h {
			img.Set(x, y, c)
		}
	}
	return img
}

func newGradientRGBA() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for x := range 4 {
		for y := range 4 {
			v := uint8(x*60 + y*15)
			img.Set(x, y, color.RGBA{R: v, G: v, B: v, A: 255})
		}
	}
	return img
}

func TestNearestInterpolator_UniformImage(t *testing.T) {
	t.Parallel()
	img := newUniformRGBA(4, 4, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	interp := NearestInterpolator{}
	r, g, b := interp.Interpolate(img, 1.7, 2.3, 4, 4)
	if r != 100 || g != 150 || b != 200 {
		t.Errorf("nearest on uniform: got (%d,%d,%d), want (100,150,200)", r, g, b)
	}
}

func TestNearestInterpolator_PicksClosest(t *testing.T) {
	t.Parallel()
	img := newGradientRGBA()
	interp := NearestInterpolator{}
	r1, _, _ := interp.Interpolate(img, 1.9, 0.1, 4, 4)
	if r1 != 120 {
		t.Errorf("nearest at (1.9,0.1): got R=%d, want 120", r1)
	}
}

func TestBilinearInterpolator_UniformImage(t *testing.T) {
	t.Parallel()
	img := newUniformRGBA(4, 4, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	interp := BilinearInterpolator{}
	r, g, b := interp.Interpolate(img, 1.5, 1.5, 4, 4)
	if r != 100 || g != 150 || b != 200 {
		t.Errorf("bilinear on uniform: got (%d,%d,%d), want (100,150,200)", r, g, b)
	}
}

func TestBilinearInterpolator_Blends(t *testing.T) {
	t.Parallel()
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	img.Set(1, 0, color.RGBA{R: 200, G: 0, B: 0, A: 255})
	interp := BilinearInterpolator{}
	r, _, _ := interp.Interpolate(img, 0.5, 0.0, 2, 1)
	if r < 90 || r > 110 {
		t.Errorf("bilinear blend at midpoint: got R=%d, want ~100", r)
	}
}

func TestBicubicInterpolator_UniformImage(t *testing.T) {
	t.Parallel()
	img := newUniformRGBA(8, 8, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	interp := BicubicInterpolator{}
	r, g, b := interp.Interpolate(img, 4.0, 4.0, 8, 8)
	if r != 100 || g != 150 || b != 200 {
		t.Errorf("bicubic on uniform: got (%d,%d,%d), want (100,150,200)", r, g, b)
	}
}

func TestBicubicInterpolator_Blends(t *testing.T) {
	t.Parallel()
	img := image.NewRGBA(image.Rect(0, 0, 4, 1))
	img.Set(0, 0, color.RGBA{R: 0, A: 255})
	img.Set(1, 0, color.RGBA{R: 0, A: 255})
	img.Set(2, 0, color.RGBA{R: 200, A: 255})
	img.Set(3, 0, color.RGBA{R: 200, A: 255})
	interp := BicubicInterpolator{}
	r, _, _ := interp.Interpolate(img, 1.5, 0.0, 4, 1)
	if r == 0 || r == 200 {
		t.Errorf("bicubic should blend, got R=%d", r)
	}
}

func TestParseInterpolation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"nearest", false},
		{"bilinear", false},
		{"bicubic", false},
		{"invalid", true},
		{"", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseInterpolation(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInterpolation(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
