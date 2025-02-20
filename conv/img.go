package conv

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

const faceLen = 6

var faceMap = map[int]string{
	0: "back",
	1: "left",
	2: "front",
	3: "right",
	4: "top",
	5: "bottom",
}

var revesedFaceMap = map[string]int{
	"back":   0,
	"left":   1,
	"front":  2,
	"right":  3,
	"top":    4,
	"bottom": 5,
}

func ReadImage(imagePath string) (image.Image, string, error) {
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("file does not exist: %s", imagePath)
	}

	imgFile, err := os.Open(imagePath)
	if err != nil {
		return nil, "", fmt.Errorf("error opening file: %s", err)
	}
	defer imgFile.Close()

	imgIn, ext, err := image.Decode(imgFile)
	if err != nil {
		return nil, "", fmt.Errorf("error decoding image: %s", err)
	}

	if ext == "jpg" || ext == "jpeg" || ext == "png" {
		return imgIn, ext, nil
	}

	return nil, "", errors.New("unsupported image format: " + ext)
}

func WriteImage(canvases []*image.RGBA, writeDirPath, imgExt string, sides []string) error {
	if len(canvases) != len(sides) {
		return errors.New("mismatched face size and sides length")
	}

	if _, err := os.Stat(writeDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(writeDirPath, os.ModePerm); err != nil {
			return err
		}
	}

	// Treat "jpg" as "jpeg"
	if imgExt == "jpg" {
		imgExt = "jpeg"
	}

	for i := 0; i < len(canvases); i++ {
		side := sides[i]
		path := filepath.Join(writeDirPath, side+"."+imgExt)
		newFile, _ := os.Create(path)

		switch imgExt {
		case "jpeg":
			if err := jpeg.Encode(newFile, canvases[i], nil); err != nil {
				return err
			}
		case "png":
			if err := png.Encode(newFile, canvases[i]); err != nil {
				return err
			}
		default:
			return errors.New("unsupported image file format: " + imgExt)
		}
	}
	return nil
}
