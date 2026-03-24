package conv

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadImage_FileNotFound(t *testing.T) {
	t.Parallel()
	_, _, err := ReadImage("/nonexistent/path/image.jpg")
	assert.Error(t, err)
}

func TestReadImage_JPEG(t *testing.T) {
	t.Parallel()
	path := createTestJPEG(t)

	img, ext, err := ReadImage(path)
	require.NoError(t, err)
	assert.Equal(t, "jpeg", ext)
	assert.Equal(t, 4, img.Bounds().Dx())
	assert.Equal(t, 2, img.Bounds().Dy())
}

func TestReadImage_PNG(t *testing.T) {
	t.Parallel()
	path := createTestPNG(t)

	img, ext, err := ReadImage(path)
	require.NoError(t, err)
	assert.Equal(t, "png", ext)
	assert.Equal(t, 4, img.Bounds().Dx())
	assert.Equal(t, 2, img.Bounds().Dy())
}

func TestReadImage_InvalidFormat(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "test.txt")
	require.NoError(t, os.WriteFile(path, []byte("not an image"), 0o644))

	_, _, err := ReadImage(path)
	assert.Error(t, err)
}

func TestToRGBA_AlreadyRGBA(t *testing.T) {
	t.Parallel()
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.Set(0, 0, color.RGBA{R: 100, G: 150, B: 200, A: 255})

	got := toRGBA(src)
	assert.Same(t, src, got, "should return same pointer for *image.RGBA input")
}

func TestToRGBA_FromNRGBA(t *testing.T) {
	t.Parallel()
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	src.Set(0, 0, color.NRGBA{R: 100, G: 150, B: 200, A: 255})

	got := toRGBA(src)
	r, g, b, a := got.At(0, 0).RGBA()
	assert.Equal(t, uint32(100), r>>8)
	assert.Equal(t, uint32(150), g>>8)
	assert.Equal(t, uint32(200), b>>8)
	assert.Equal(t, uint32(255), a>>8)
}

func TestWriteImage_JPEG(t *testing.T) {
	t.Parallel()
	outDir := t.TempDir()
	canvas := createTestCanvas()

	err := WriteImage([]*image.RGBA{canvas}, outDir, "jpeg", []string{"front"}, 75)
	require.NoError(t, err)

	outPath := filepath.Join(outDir, "front.jpeg")
	assert.FileExists(t, outPath)

	img, ext, err := ReadImage(outPath)
	require.NoError(t, err)
	assert.Equal(t, "jpeg", ext)
	assert.Equal(t, 4, img.Bounds().Dx())
}

func TestWriteImage_PNG(t *testing.T) {
	t.Parallel()
	outDir := t.TempDir()

	err := WriteImage([]*image.RGBA{createTestCanvas()}, outDir, "png", []string{"back"}, 0)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(outDir, "back.png"))
}

func TestWriteImage_JPGNormalization(t *testing.T) {
	t.Parallel()
	outDir := t.TempDir()

	err := WriteImage([]*image.RGBA{createTestCanvas()}, outDir, "jpg", []string{"top"}, 75)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(outDir, "top.jpeg"), "jpg should be normalized to jpeg")
}

func TestWriteImage_MismatchedLength(t *testing.T) {
	t.Parallel()
	err := WriteImage([]*image.RGBA{createTestCanvas()}, t.TempDir(), "png", []string{"front", "back"}, 0)
	assert.Error(t, err)
}

func TestWriteImage_MultipleSides(t *testing.T) {
	t.Parallel()
	outDir := t.TempDir()
	sides := []string{"front", "back", "left"}
	canvases := make([]*image.RGBA, len(sides))
	for i := range canvases {
		canvases[i] = createTestCanvas()
	}

	require.NoError(t, WriteImage(canvases, outDir, "png", sides, 0))
	for _, side := range sides {
		assert.FileExists(t, filepath.Join(outDir, side+".png"))
	}
}

func TestWriteImage_CreatesDirectory(t *testing.T) {
	t.Parallel()
	outDir := filepath.Join(t.TempDir(), "nested", "dir")

	require.NoError(t, WriteImage([]*image.RGBA{createTestCanvas()}, outDir, "png", []string{"front"}, 0))
	assert.FileExists(t, filepath.Join(outDir, "front.png"))
}

func TestWriteImage_UnsupportedFormat(t *testing.T) {
	t.Parallel()
	err := WriteImage([]*image.RGBA{createTestCanvas()}, t.TempDir(), "bmp", []string{"front"}, 0)
	assert.Error(t, err)
}

func TestReversedFaceMap(t *testing.T) {
	t.Parallel()

	expected := map[string]int{
		"back": 0, "left": 1, "front": 2,
		"right": 3, "top": 4, "bottom": 5,
	}

	assert.Equal(t, expected, reversedFaceMap)
}

// helpers

func createTestCanvas() *image.RGBA {
	return newTestImage(4, 4, color.RGBA{R: 100, G: 150, B: 200, A: 255})
}

func createTestJPEG(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.jpeg")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	img := newTestImage(4, 2, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	require.NoError(t, jpeg.Encode(f, img, &jpeg.Options{Quality: 90}))
	return path
}

func createTestPNG(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.png")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	img := newTestImage(4, 2, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	require.NoError(t, png.Encode(f, img))
	return path
}
