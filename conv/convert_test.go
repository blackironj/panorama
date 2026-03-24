package conv

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		n    int
		size int
		want int
	}{
		{"within bounds", 5, 10, 5},
		{"negative clamped to zero", -3, 10, 0},
		{"exceeds size clamped to max", 15, 10, 9},
		{"at boundary", 9, 10, 9},
		{"zero", 0, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, safeIndex(tt.n, tt.size))
		})
	}
}

func TestOutImgToXYZ_ValidFaces(t *testing.T) {
	t.Parallel()

	inLen := 2.0 / 4.0

	for face := range faceCount {
		t.Run(faceNames[face], func(t *testing.T) {
			t.Parallel()
			v := outImgToXYZ(0, 0, face, inLen)
			assert.False(t, v.X == 0 && v.Y == 0 && v.Z == 0, "face %d produced zero vector", face)
		})
	}
}

func TestOutImgToXYZ_FaceDirections(t *testing.T) {
	t.Parallel()

	inLen := 1.0 // edge=2

	tests := []struct {
		face    int
		i, j    int
		wantDom string
	}{
		{faceBack, 1, 1, "X"},
		{faceFront, 1, 1, "X"},
		{faceLeft, 1, 1, "Y"},
		{faceRight, 1, 1, "Y"},
		{faceTop, 1, 1, "Z"},
		{faceBottom, 1, 1, "Z"},
	}

	for _, tt := range tests {
		t.Run(faceNames[tt.face], func(t *testing.T) {
			t.Parallel()
			v := outImgToXYZ(tt.i, tt.j, tt.face, inLen)
			absX, absY, absZ := math.Abs(v.X), math.Abs(v.Y), math.Abs(v.Z)
			switch tt.wantDom {
			case "X":
				assert.True(t, absX >= absY && absX >= absZ, "face %s: expected X dominant, got %v", faceNames[tt.face], v)
			case "Y":
				assert.True(t, absY >= absX && absY >= absZ, "face %s: expected Y dominant, got %v", faceNames[tt.face], v)
			case "Z":
				assert.True(t, absZ >= absX && absZ >= absY, "face %s: expected Z dominant, got %v", faceNames[tt.face], v)
			}
		})
	}
}

var faceNames = [faceCount]string{"back", "left", "front", "right", "top", "bottom"}

func TestConvertEquirectangularToCubeMap(t *testing.T) {
	t.Parallel()

	img := newTestImage(8, 4, color.RGBA{R: 128, G: 64, B: 32, A: 255})
	edgeLen := 4
	sides := []string{"front", "back"}

	canvases, err := ConvertEquirectangularToCubeMap(edgeLen, img, sides, BilinearInterpolator{})
	require.NoError(t, err)
	require.Len(t, canvases, 2)

	for i, canvas := range canvases {
		bounds := canvas.Bounds()
		assert.Equal(t, edgeLen, bounds.Dx(), "canvas[%d] width", i)
		assert.Equal(t, edgeLen, bounds.Dy(), "canvas[%d] height", i)
	}
}

func TestConvertEquirectangularToCubeMap_AllSides(t *testing.T) {
	t.Parallel()

	img := newTestImage(8, 4, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	allSides := []string{"back", "left", "front", "right", "top", "bottom"}

	canvases, err := ConvertEquirectangularToCubeMap(4, img, allSides, BilinearInterpolator{})
	require.NoError(t, err)
	require.Len(t, canvases, faceCount)

	for i, canvas := range canvases {
		_, _, _, a := canvas.At(0, 0).RGBA()
		assert.Equal(t, uint32(255), a>>8, "canvas[%d] (%s) alpha", i, allSides[i])
	}
}

func TestConvert_OutputNonZero(t *testing.T) {
	t.Parallel()

	img := newTestImage(8, 4, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	edge := 4
	out := image.NewRGBA(image.Rect(0, 0, edge, edge))

	convert(edge, faceFront, 8, 4, img, out, BilinearInterpolator{})

	r, g, b, _ := out.At(edge/2, edge/2).RGBA()
	assert.False(t, r == 0 && g == 0 && b == 0, "center pixel should not be black")
}

func newTestImage(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := range w {
		for y := range h {
			img.Set(x, y, c)
		}
	}
	return img
}
