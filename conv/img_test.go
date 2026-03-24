package conv

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestReadImage_FileNotFound(t *testing.T) {
	t.Parallel()

	_, _, err := ReadImage("/nonexistent/path/image.jpg")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestReadImage_JPEG(t *testing.T) {
	t.Parallel()

	path := createTestJPEG(t)
	img, ext, err := ReadImage(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ext != "jpeg" {
		t.Errorf("expected ext 'jpeg', got %q", ext)
	}
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 2 {
		t.Errorf("unexpected image size: %v", img.Bounds())
	}
}

func TestReadImage_PNG(t *testing.T) {
	t.Parallel()

	path := createTestPNG(t)
	img, ext, err := ReadImage(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ext != "png" {
		t.Errorf("expected ext 'png', got %q", ext)
	}
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 2 {
		t.Errorf("unexpected image size: %v", img.Bounds())
	}
}

func TestReadImage_InvalidFormat(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(path, []byte("not an image"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := ReadImage(path)
	if err == nil {
		t.Fatal("expected error for invalid image format, got nil")
	}
}

func TestWriteImage_JPEG(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	canvas := createTestCanvas()
	sides := []string{"front"}

	err := WriteImage([]*image.RGBA{canvas}, outDir, "jpeg", sides, 75)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outPath := filepath.Join(outDir, "front.jpeg")
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Fatalf("expected output file %s to exist", outPath)
	}

	// Verify it's a valid JPEG
	img, ext, err := ReadImage(outPath)
	if err != nil {
		t.Fatalf("failed to read back written image: %v", err)
	}
	if ext != "jpeg" {
		t.Errorf("expected ext 'jpeg', got %q", ext)
	}
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 4 {
		t.Errorf("unexpected size: %v", img.Bounds())
	}
}

func TestWriteImage_PNG(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	canvas := createTestCanvas()
	sides := []string{"back"}

	err := WriteImage([]*image.RGBA{canvas}, outDir, "png", sides, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outPath := filepath.Join(outDir, "back.png")
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Fatalf("expected output file %s to exist", outPath)
	}
}

func TestWriteImage_JPGNormalization(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	canvas := createTestCanvas()

	err := WriteImage([]*image.RGBA{canvas}, outDir, "jpg", []string{"top"}, 75)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// "jpg" should be normalized to "jpeg" in the output filename
	outPath := filepath.Join(outDir, "top.jpeg")
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Fatalf("expected output file %s to exist (jpg normalized to jpeg)", outPath)
	}
}

func TestWriteImage_MismatchedLength(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	canvas := createTestCanvas()

	err := WriteImage([]*image.RGBA{canvas}, outDir, "png", []string{"front", "back"}, 0)
	if err == nil {
		t.Fatal("expected error for mismatched canvases/sides, got nil")
	}
}

func TestWriteImage_MultipleSides(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	sides := []string{"front", "back", "left"}
	canvases := make([]*image.RGBA, len(sides))
	for i := range canvases {
		canvases[i] = createTestCanvas()
	}

	err := WriteImage(canvases, outDir, "png", sides, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, side := range sides {
		outPath := filepath.Join(outDir, side+".png")
		if _, err := os.Stat(outPath); os.IsNotExist(err) {
			t.Errorf("expected output file %s to exist", outPath)
		}
	}
}

func TestWriteImage_CreatesDirectory(t *testing.T) {
	t.Parallel()

	outDir := filepath.Join(t.TempDir(), "nested", "dir")
	canvas := createTestCanvas()

	err := WriteImage([]*image.RGBA{canvas}, outDir, "png", []string{"front"}, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outPath := filepath.Join(outDir, "front.png")
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Fatalf("expected output file %s to exist", outPath)
	}
}

func TestWriteImage_UnsupportedFormat(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	canvas := createTestCanvas()

	err := WriteImage([]*image.RGBA{canvas}, outDir, "bmp", []string{"front"}, 0)
	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
}

func TestReversedFaceMap(t *testing.T) {
	t.Parallel()

	expected := map[string]int{
		"back": 0, "left": 1, "front": 2,
		"right": 3, "top": 4, "bottom": 5,
	}

	for name, idx := range expected {
		got, ok := reversedFaceMap[name]
		if !ok {
			t.Errorf("reversedFaceMap missing key %q", name)
			continue
		}
		if got != idx {
			t.Errorf("reversedFaceMap[%q] = %d, want %d", name, got, idx)
		}
	}

	if len(reversedFaceMap) != len(expected) {
		t.Errorf("reversedFaceMap has %d entries, want %d", len(reversedFaceMap), len(expected))
	}
}

func TestToRGBA_AlreadyRGBA(t *testing.T) {
	t.Parallel()
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.Set(0, 0, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	got := toRGBA(src)
	if got != src {
		t.Error("expected same pointer for *image.RGBA input")
	}
}

func TestToRGBA_FromNRGBA(t *testing.T) {
	t.Parallel()
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	src.Set(0, 0, color.NRGBA{R: 100, G: 150, B: 200, A: 255})
	got := toRGBA(src)
	r, g, b, a := got.At(0, 0).RGBA()
	if r>>8 != 100 || g>>8 != 150 || b>>8 != 200 || a>>8 != 255 {
		t.Errorf("pixel mismatch: got RGBA(%d,%d,%d,%d)", r>>8, g>>8, b>>8, a>>8)
	}
}

// helpers

func createTestCanvas() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for x := range 4 {
		for y := range 4 {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	return img
}

func createTestJPEG(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.jpeg")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	for x := range 4 {
		for y := range 2 {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return path
}

func createTestPNG(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	for x := range 4 {
		for y := range 2 {
			img.Set(x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		}
	}

	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return path
}
