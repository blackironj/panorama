package conv

import (
	"fmt"
	"image"
	"math"
	"sync"
)

const piHalf = math.Pi / 2.0

type vec3 struct {
	X, Y, Z float64
}

func outImgToXYZ(i, j, face int, inLen float64) (vec3, error) {
	a := inLen*float64(i) - 1.0
	b := inLen*float64(j) - 1.0

	var res vec3
	switch face {
	case 0: // back
		res = vec3{-1.0, -a, -b}
	case 1: // left
		res = vec3{a, -1.0, -b}
	case 2: // front
		res = vec3{1.0, a, -b}
	case 3: // right
		res = vec3{-a, 1.0, -b}
	case 4: // top
		res = vec3{b, a, 1.0}
	case 5: // bottom
		res = vec3{-b, a, -1.0}
	default:
		return vec3{}, fmt.Errorf("invalid face index: %d", face)
	}
	return res, nil
}

func ConvertEquirectangularToCubeMap(edgeLen int, imgIn *image.RGBA, sides []string, interp Interpolator) ([]*image.RGBA, error) {
	sw := imgIn.Bounds().Max.X
	sh := imgIn.Bounds().Max.Y
	sidesCount := len(sides)

	sidesInt := make([]int, 0, sidesCount)
	for i := range sidesCount {
		sidesInt = append(sidesInt, reversedFaceMap[sides[i]])
	}

	var wg sync.WaitGroup

	canvases := make([]*image.RGBA, sidesCount)
	for i := range sidesCount {
		canvases[i] = image.NewRGBA(image.Rect(0, 0, edgeLen, edgeLen))
	}

	errs := make([]error, sidesCount)
	for i := range sidesCount {
		wg.Add(1)
		go func(idx, side int, canvas *image.RGBA) {
			defer wg.Done()
			errs[idx] = convert(edgeLen, side, sw, sh, imgIn, canvas, interp)
		}(i, sidesInt[i], canvases[i])
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return canvases, nil
}

func convert(edge, face, sw, sh int, imgIn *image.RGBA, imgOut *image.RGBA, interp Interpolator) error {
	inLen := 2.0 / float64(edge)
	dividedH := float64(sh) / math.Pi

	for i := range edge {
		for j := range edge {
			xyz, err := outImgToXYZ(i, j, face, inLen)
			if err != nil {
				return err
			}

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
	return nil
}

func safeIndex(n, size float64) int {
	return int(math.Min(math.Max(n, 0), size-1))
}
