package conv

import (
	"image"
	"image/color"
	"log"
	"math"
	"sync"
)

const Pi_2 = math.Pi / 2.0

type Number interface {
	uint32 | float64
}
type Vec3[T Number] struct {
	X, Y, Z T
}

func outImgToXYZ(i, j, face, edge int, inLen float64) Vec3[float64] {
	a := inLen*float64(i) - 1.0
	b := inLen*float64(j) - 1.0

	var res Vec3[float64]
	switch face {
	case 0: //back
		res = Vec3[float64]{-1.0, -a, -b}
	case 1: //left
		res = Vec3[float64]{a, -1.0, -b}
	case 2: //front
		res = Vec3[float64]{1.0, a, -b}
	case 3: //right
		res = Vec3[float64]{-a, 1.0, -b}
	case 4: //top
		res = Vec3[float64]{b, a, 1.0}
	case 5: //bottom
		res = Vec3[float64]{-b, a, -1.0}
	default:
		log.Fatal("Wrong face")
	}
	return res
}

func interpolateXYZtoColor(xyz Vec3[float64], imgIn image.Image, sw, sh int) Vec3[uint32] {
	theta := math.Atan2(xyz.Y, xyz.X)
	rad := math.Hypot(xyz.X, xyz.Y) // range -pi to pi
	phi := math.Atan2(xyz.Z, rad)   // range -pi/2 to pi/2

	//source img coords
	dividedH := float64(sh) / math.Pi
	uf := (theta + math.Pi) * dividedH
	vf := (Pi_2 - phi) * dividedH

	// Use bilinear interpolation between the four surrounding pixels
	ui := safeIndex(math.Floor(uf), float64(sw))
	vi := safeIndex(math.Floor(vf), float64(sh))
	u2 := safeIndex(float64(ui)+1.0, float64(sw))
	v2 := safeIndex(float64(vi)+1.0, float64(sh))

	mu := uf - float64(ui)
	nu := vf - float64(vi)

	read := func(x, y int) Vec3[float64] {
		red, green, blue, _ := imgIn.At(x, y).RGBA()
		return Vec3[float64]{
			X: float64(red >> 8),
			Y: float64(green >> 8),
			Z: float64(blue >> 8),
		}
	}

	A := read(ui, vi)
	B := read(u2, vi)
	C := read(ui, v2)
	D := read(u2, v2)

	val := mix(mix(A, B, mu), mix(C, D, mu), nu)
	return Vec3[uint32]{
		X: uint32(val.X),
		Y: uint32(val.Y),
		Z: uint32(val.Z),
	}
}

func ConvertEquirectangularToCubeMap(rValue int, imgIn image.Image) []*image.RGBA {
	sw := imgIn.Bounds().Max.X
	sh := imgIn.Bounds().Max.Y

	var wg sync.WaitGroup

	canvases := make([]*image.RGBA, 6)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		canvases[i] = image.NewRGBA(image.Rect(0, 0, rValue, rValue))
		start := i * rValue
		end := start + rValue

		go func() {
			defer wg.Done()
			convert(start, end, rValue, sw, sh, imgIn, canvases)
		}()
	}
	wg.Wait()

	return canvases
}

func convert(start, end, edge, sw, sh int, imgIn image.Image, imgOut []*image.RGBA) {
	inLen := 2.0 / float64(edge)

	for k := start; k < end; k++ {
		face := k / edge
		i := k % edge

		for j := 0; j < edge; j++ {
			xyz := outImgToXYZ(i, j, face, edge, inLen)
			clr := interpolateXYZtoColor(xyz, imgIn, sw, sh)

			imgOut[face].Set(i, j, color.RGBA{uint8(clr.X), uint8(clr.Y), uint8(clr.Z), 255})
		}
	}
}

func safeIndex(n, size float64) int {
	return int(math.Min(math.Max(n, 0), size-1))
}

func mix(one, other Vec3[float64], c float64) Vec3[float64] {
	x := (other.X-one.X)*c + one.X
	y := (other.Y-one.Y)*c + one.Y
	z := (other.Z-one.Z)*c + one.Z

	return Vec3[float64]{
		X: x,
		Y: y,
		Z: z,
	}
}
