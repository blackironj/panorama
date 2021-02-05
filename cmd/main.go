package main

import (
	"image"
	"image/jpeg"
	"os"
	"strconv"

	"github.com/blackironj/panorama/conv"
)

const (
	rValue   = 2048
	testLoop = 20
)

func main() {
	imgFile, _ := os.Open("sample.jpg")
	defer imgFile.Close()
	imgIn, _, _ := image.Decode(imgFile)

	canvases := make([]*image.RGBA, 6)
	canvases = conv.ConverPanoramaToCubemap(rValue, imgIn)

	opt := jpeg.Options{
		Quality: 90,
	}

	for i := 0; i < 6; i++ {
		newFile, _ := os.Create("test" + strconv.Itoa(i) + ".jpg")
		jpeg.Encode(newFile, canvases[i], &opt)
	}
}
