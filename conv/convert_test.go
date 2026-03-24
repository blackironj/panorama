package conv

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestSafeIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		n    float64
		size float64
		want int
	}{
		{"within bounds", 5.0, 10.0, 5},
		{"negative clamped to zero", -3.0, 10.0, 0},
		{"exceeds size clamped to max", 15.0, 10.0, 9},
		{"at boundary", 9.0, 10.0, 9},
		{"zero", 0.0, 10.0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := safeIndex(tt.n, tt.size)
			if got != tt.want {
				t.Errorf("safeIndex(%v, %v) = %d, want %d", tt.n, tt.size, got, tt.want)
			}
		})
	}
}

func TestOutImgToXYZ_ValidFaces(t *testing.T) {
	t.Parallel()

	inLen := 2.0 / 4.0 // edge=4

	for face := range 6 {
		t.Run(faceNames[face], func(t *testing.T) {
			t.Parallel()
			v, err := outImgToXYZ(0, 0, face, inLen)
			if err != nil {
				t.Fatalf("unexpected error for face %d: %v", face, err)
			}
			if v.X == 0 && v.Y == 0 && v.Z == 0 {
				t.Errorf("face %d produced zero vector", face)
			}
		})
	}
}

func TestOutImgToXYZ_InvalidFace(t *testing.T) {
	t.Parallel()

	_, err := outImgToXYZ(0, 0, 99, 0.5)
	if err == nil {
		t.Fatal("expected error for invalid face index, got nil")
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
		{0, 1, 1, "X"}, // back: X=-1
		{2, 1, 1, "X"}, // front: X=1
		{1, 1, 1, "Y"}, // left: Y=-1
		{3, 1, 1, "Y"}, // right: Y=1
		{4, 1, 1, "Z"}, // top: Z=1
		{5, 1, 1, "Z"}, // bottom: Z=-1
	}

	for _, tt := range tests {
		t.Run(faceNames[tt.face], func(t *testing.T) {
			t.Parallel()
			v, err := outImgToXYZ(tt.i, tt.j, tt.face, inLen)
			if err != nil {
				t.Fatal(err)
			}
			absX, absY, absZ := math.Abs(v.X), math.Abs(v.Y), math.Abs(v.Z)
			switch tt.wantDom {
			case "X":
				if absX < absY || absX < absZ {
					t.Errorf("face %s: expected X dominant, got %v", faceNames[tt.face], v)
				}
			case "Y":
				if absY < absX || absY < absZ {
					t.Errorf("face %s: expected Y dominant, got %v", faceNames[tt.face], v)
				}
			case "Z":
				if absZ < absX || absZ < absY {
					t.Errorf("face %s: expected Z dominant, got %v", faceNames[tt.face], v)
				}
			}
		})
	}
}

var faceNames = [6]string{"back", "left", "front", "right", "top", "bottom"}

func TestConvertEquirectangularToCubeMap(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 8, 4))
	for x := range 8 {
		for y := range 4 {
			img.Set(x, y, color.RGBA{R: 128, G: 64, B: 32, A: 255})
		}
	}

	edgeLen := 4
	sides := []string{"front", "back"}

	canvases, err := ConvertEquirectangularToCubeMap(edgeLen, img, sides, BilinearInterpolator{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(canvases) != 2 {
		t.Fatalf("expected 2 canvases, got %d", len(canvases))
	}

	for i, canvas := range canvases {
		bounds := canvas.Bounds()
		if bounds.Dx() != edgeLen || bounds.Dy() != edgeLen {
			t.Errorf("canvas[%d] size = %dx%d, want %dx%d", i, bounds.Dx(), bounds.Dy(), edgeLen, edgeLen)
		}
	}
}

func TestConvertEquirectangularToCubeMap_AllSides(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 8, 4))
	for x := range 8 {
		for y := range 4 {
			img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	allSides := []string{"back", "left", "front", "right", "top", "bottom"}
	canvases, err := ConvertEquirectangularToCubeMap(4, img, allSides, BilinearInterpolator{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(canvases) != 6 {
		t.Fatalf("expected 6 canvases, got %d", len(canvases))
	}

	for i, canvas := range canvases {
		_, _, _, a := canvas.At(0, 0).RGBA()
		if a>>8 != 255 {
			t.Errorf("canvas[%d] (%s) pixel (0,0) has alpha %d, want 255", i, allSides[i], a>>8)
		}
	}
}

func TestConvert_OutputNonZero(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 8, 4))
	for x := range 8 {
		for y := range 4 {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}

	edge := 4
	out := image.NewRGBA(image.Rect(0, 0, edge, edge))

	if err := convert(edge, 2, 8, 4, img, out, BilinearInterpolator{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r, g, b, _ := out.At(edge/2, edge/2).RGBA()
	if r == 0 && g == 0 && b == 0 {
		t.Error("center pixel is black, expected non-zero color")
	}
}
