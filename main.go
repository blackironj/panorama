package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"strconv"
	"time"

	"github.com/blackironj/panorama/img"
)

const (
	rValue   = 2048
	testLoop = 20
)

func main() {
	imgFile, _ := os.Open("sample.jpg")
	defer imgFile.Close()
	imgIn, _, _ := image.Decode(imgFile)

	var totalTime time.Duration
	fmt.Println("[GO] Panoramal to cube map")

	canvases := make([]*image.RGBA, 6)
	for i := 0; i < testLoop; i++ {
		start := time.Now()
		canvases = img.ConverBack(rValue, imgIn)

		fmt.Printf("[GO] elapsed time : %v\n", time.Since(start))
		totalTime += time.Since(start)
	}
	fmt.Printf("[GO] average time : %v\n", totalTime/testLoop)

	opt := jpeg.Options{
		Quality: 90,
	}

	for i := 0; i < 6; i++ {
		newFile, _ := os.Create("test" + strconv.Itoa(i) + ".jpg")
		jpeg.Encode(newFile, canvases[i], &opt)
	}
}
