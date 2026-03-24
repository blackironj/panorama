package conv

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	r, g, b := NearestInterpolator{}.Interpolate(img, 1.7, 2.3, 4, 4)
	assert.Equal(t, uint8(100), r)
	assert.Equal(t, uint8(150), g)
	assert.Equal(t, uint8(200), b)
}

func TestNearestInterpolator_PicksClosest(t *testing.T) {
	t.Parallel()
	img := newGradientRGBA()
	r, _, _ := NearestInterpolator{}.Interpolate(img, 1.9, 0.1, 4, 4)
	assert.Equal(t, uint8(120), r)
}

func TestBilinearInterpolator_UniformImage(t *testing.T) {
	t.Parallel()
	img := newUniformRGBA(4, 4, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	r, g, b := BilinearInterpolator{}.Interpolate(img, 1.5, 1.5, 4, 4)
	assert.Equal(t, uint8(100), r)
	assert.Equal(t, uint8(150), g)
	assert.Equal(t, uint8(200), b)
}

func TestBilinearInterpolator_Blends(t *testing.T) {
	t.Parallel()
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	img.Set(1, 0, color.RGBA{R: 200, G: 0, B: 0, A: 255})

	r, _, _ := BilinearInterpolator{}.Interpolate(img, 0.5, 0.0, 2, 1)
	assert.InDelta(t, 100, int(r), 10, "bilinear blend at midpoint")
}

func TestBicubicInterpolator_UniformImage(t *testing.T) {
	t.Parallel()
	img := newUniformRGBA(8, 8, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	r, g, b := BicubicInterpolator{}.Interpolate(img, 4.0, 4.0, 8, 8)
	assert.Equal(t, uint8(100), r)
	assert.Equal(t, uint8(150), g)
	assert.Equal(t, uint8(200), b)
}

func TestBicubicInterpolator_Blends(t *testing.T) {
	t.Parallel()
	img := image.NewRGBA(image.Rect(0, 0, 4, 1))
	img.Set(0, 0, color.RGBA{R: 0, A: 255})
	img.Set(1, 0, color.RGBA{R: 0, A: 255})
	img.Set(2, 0, color.RGBA{R: 200, A: 255})
	img.Set(3, 0, color.RGBA{R: 200, A: 255})

	r, _, _ := BicubicInterpolator{}.Interpolate(img, 1.5, 0.0, 4, 1)
	assert.NotEqual(t, uint8(0), r, "bicubic should blend")
	assert.NotEqual(t, uint8(200), r, "bicubic should blend")
}

func TestParseInterpolation(t *testing.T) {
	t.Parallel()

	t.Run("valid methods", func(t *testing.T) {
		t.Parallel()
		for _, name := range []string{"nearest", "bilinear", "bicubic"} {
			interp, err := ParseInterpolation(name)
			require.NoError(t, err, name)
			assert.NotNil(t, interp, name)
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		t.Parallel()
		_, err := ParseInterpolation("invalid")
		assert.Error(t, err)
	})

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()
		_, err := ParseInterpolation("")
		assert.Error(t, err)
	})
}
