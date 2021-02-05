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

	return nil, "", errors.New("We do not support this format : " + ext)
}

func WriteImage(canvases []*image.RGBA, writeDirPath, imgExt string) error {
	if len(canvases) != faceLen {
		return errors.New("Wrong face size")
	}

	if _, err := os.Stat(writeDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(writeDirPath, os.ModePerm); err != nil {
			return err
		}
	}

	for i := 0; i < faceLen; i++ {
		path := filepath.Join(writeDirPath, faceMap[i]+"."+imgExt)
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
			return errors.New("Wrong image file format : " + imgExt)
		}
	}
	return nil
}
