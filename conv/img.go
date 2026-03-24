package conv

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

var reversedFaceMap = map[string]int{
	"back":   0,
	"left":   1,
	"front":  2,
	"right":  3,
	"top":    4,
	"bottom": 5,
}

func toRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return rgba
}

func ReadImage(imagePath string) (*image.RGBA, string, error) {
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("file does not exist: %s", imagePath)
	}

	imgFile, err := os.Open(imagePath)
	if err != nil {
		return nil, "", fmt.Errorf("opening file: %w", err)
	}
	defer imgFile.Close()

	imgIn, ext, err := image.Decode(imgFile)
	if err != nil {
		return nil, "", fmt.Errorf("decoding image: %w", err)
	}

	if ext == "jpg" || ext == "jpeg" || ext == "png" {
		return toRGBA(imgIn), ext, nil
	}

	return nil, "", errors.New("unsupported image format: " + ext)
}

func WriteImage(canvases []*image.RGBA, writeDirPath, imgExt string, sides []string, quality int) error {
	if len(canvases) != len(sides) {
		return errors.New("mismatched face size and sides length")
	}

	if _, err := os.Stat(writeDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(writeDirPath, os.ModePerm); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	// Treat "jpg" as "jpeg"
	if imgExt == "jpg" {
		imgExt = "jpeg"
	}

	for i, canvas := range canvases {
		side := sides[i]
		path := filepath.Join(writeDirPath, side+"."+imgExt)

		if err := writeImageFile(path, imgExt, canvas, quality); err != nil {
			return fmt.Errorf("writing %s: %w", side, err)
		}
	}
	return nil
}

func writeImageFile(path, ext string, canvas *image.RGBA, quality int) (retErr error) {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && retErr == nil {
			retErr = fmt.Errorf("closing file: %w", cerr)
		}
	}()

	switch ext {
	case "jpeg":
		return jpeg.Encode(f, canvas, &jpeg.Options{Quality: quality})
	case "png":
		return png.Encode(f, canvas)
	default:
		return errors.New("unsupported image file format: " + ext)
	}
}
