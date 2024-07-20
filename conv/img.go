package conv

import (
	"errors"
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
		return nil, "", err
	}

	imgFile, _ := os.Open(imagePath)
	defer imgFile.Close()

	imgIn, ext, err := image.Decode(imgFile)
	if err != nil {
		return nil, "", err
	}

	if ext == "jpeg" || ext == "png" {
		return imgIn, ext, nil
	}

	return nil, "", errors.New("We do not support this format: " + ext)
}

func WriteImage(canvases []*image.RGBA, writeDirPath, imgExt string, sides []string) error {
	if len(canvases) != len(sides) {
		return errors.New("Mismatched face size and sides length")
	}

	if _, err := os.Stat(writeDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(writeDirPath, os.ModePerm); err != nil {
			return err
		}
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
			return errors.New("Unsupported image file format: " + imgExt)
		}
	}
	return nil
}
