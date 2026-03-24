package conv

import (
	"image"
	"math"
	"sync"
)

const piHalf = math.Pi / 2.0

const (
	faceBack   = 0
	faceLeft   = 1
	faceFront  = 2
	faceRight  = 3
	faceTop    = 4
	faceBottom = 5
	faceCount  = 6
)

type vec3 struct {
	X, Y, Z float64
}

func outImgToXYZ(i, j, face int, inLen float64) vec3 {
	a := inLen*float64(i) - 1.0
	b := inLen*float64(j) - 1.0

	switch face {
	case faceBack:
		return vec3{-1.0, -a, -b}
	case faceLeft:
		return vec3{a, -1.0, -b}
	case faceFront:
		return vec3{1.0, a, -b}
	case faceRight:
		return vec3{-a, 1.0, -b}
	case faceTop:
		return vec3{b, a, 1.0}
	default: // faceBottom
		return vec3{-b, a, -1.0}
	}
}

func ConvertEquirectangularToCubeMap(edgeLen int, imgIn *image.RGBA, sides []string, interp Interpolator) ([]*image.RGBA, error) {
	sw := imgIn.Bounds().Max.X
	sh := imgIn.Bounds().Max.Y
	sidesCount := len(sides)

	sidesInt := make([]int, sidesCount)
	for i := range sidesCount {
		sidesInt[i] = reversedFaceMap[sides[i]]
	}

	var wg sync.WaitGroup

	canvases := make([]*image.RGBA, sidesCount)
	for i := range sidesCount {
		canvases[i] = image.NewRGBA(image.Rect(0, 0, edgeLen, edgeLen))
	}

	for i := range sidesCount {
		wg.Add(1)
		go func(idx, side int, canvas *image.RGBA) {
			defer wg.Done()
			convert(edgeLen, side, sw, sh, imgIn, canvas, interp)
		}(i, sidesInt[i], canvases[i])
	}
	wg.Wait()

	return canvases, nil
}

func convert(edge, face, sw, sh int, imgIn *image.RGBA, imgOut *image.RGBA, interp Interpolator) {
	inLen := 2.0 / float64(edge)
	dividedH := float64(sh) / math.Pi

	for i := range edge {
		for j := range edge {
			xyz := outImgToXYZ(i, j, face, inLen)

			theta := math.Atan2(xyz.Y, xyz.X)
			rad := math.Hypot(xyz.X, xyz.Y)
			phi := math.Atan2(xyz.Z, rad)

			uf := (theta + math.Pi) * dividedH
			vf := (piHalf - phi) * dividedH

			r, g, b := interp.Interpolate(imgIn, uf, vf, sw, sh)
			off := (j*edge + i) * 4
			imgOut.Pix[off] = r
			imgOut.Pix[off+1] = g
			imgOut.Pix[off+2] = b
			imgOut.Pix[off+3] = 255
		}
	}
}

func safeIndex(n, size int) int {
	if n < 0 {
		return 0
	}
	if n >= size {
		return size - 1
	}
	return n
}
