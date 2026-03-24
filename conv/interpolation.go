package conv

import (
	"fmt"
	"image"
	"math"
)

// Interpolator samples a color from an RGBA image at fractional coordinates.
type Interpolator interface {
	Interpolate(img *image.RGBA, uf, vf float64, sw, sh int) (r, g, b uint8)
}

// ParseInterpolation returns an Interpolator for the given name.
func ParseInterpolation(name string) (Interpolator, error) {
	switch name {
	case "nearest":
		return NearestInterpolator{}, nil
	case "bilinear":
		return BilinearInterpolator{}, nil
	case "bicubic":
		return BicubicInterpolator{}, nil
	default:
		return nil, fmt.Errorf("unknown interpolation method: %q (valid: nearest, bilinear, bicubic)", name)
	}
}

func readPixel(img *image.RGBA, x, y int) (r, g, b float64) {
	off := (y-img.Rect.Min.Y)*img.Stride + (x-img.Rect.Min.X)*4
	return float64(img.Pix[off]), float64(img.Pix[off+1]), float64(img.Pix[off+2])
}

// NearestInterpolator samples the single closest pixel.
type NearestInterpolator struct{}

func (NearestInterpolator) Interpolate(img *image.RGBA, uf, vf float64, sw, sh int) (r, g, b uint8) {
	x := safeIndex(int(math.Round(uf)), sw)
	y := safeIndex(int(math.Round(vf)), sh)
	pr, pg, pb := readPixel(img, x, y)
	return uint8(pr), uint8(pg), uint8(pb)
}

// BilinearInterpolator uses weighted average of 4 surrounding pixels.
type BilinearInterpolator struct{}

func (BilinearInterpolator) Interpolate(img *image.RGBA, uf, vf float64, sw, sh int) (r, g, b uint8) {
	ui := safeIndex(int(math.Floor(uf)), sw)
	vi := safeIndex(int(math.Floor(vf)), sh)
	u2 := safeIndex(ui+1, sw)
	v2 := safeIndex(vi+1, sh)

	mu := uf - float64(ui)
	nu := vf - float64(vi)

	r00, g00, b00 := readPixel(img, ui, vi)
	r10, g10, b10 := readPixel(img, u2, vi)
	r01, g01, b01 := readPixel(img, ui, v2)
	r11, g11, b11 := readPixel(img, u2, v2)

	fr := lerp(lerp(r00, r10, mu), lerp(r01, r11, mu), nu)
	fg := lerp(lerp(g00, g10, mu), lerp(g01, g11, mu), nu)
	fb := lerp(lerp(b00, b10, mu), lerp(b01, b11, mu), nu)

	return uint8(clamp(fr)), uint8(clamp(fg)), uint8(clamp(fb))
}

// BicubicInterpolator uses weighted average of 16 surrounding pixels with cubic kernel.
type BicubicInterpolator struct{}

func (BicubicInterpolator) Interpolate(img *image.RGBA, uf, vf float64, sw, sh int) (r, g, b uint8) {
	ui := int(math.Floor(uf))
	vi := int(math.Floor(vf))
	du := uf - float64(ui)
	dv := vf - float64(vi)

	var sr, sg, sb float64
	for m := -1; m <= 2; m++ {
		wv := cubicWeight(dv - float64(m))
		for n := -1; n <= 2; n++ {
			wu := cubicWeight(du - float64(n))
			px := safeIndex(ui+n, sw)
			py := safeIndex(vi+m, sh)
			pr, pg, pb := readPixel(img, px, py)
			w := wu * wv
			sr += pr * w
			sg += pg * w
			sb += pb * w
		}
	}

	return uint8(clamp(sr)), uint8(clamp(sg)), uint8(clamp(sb))
}

func cubicWeight(d float64) float64 {
	d = math.Abs(d)
	if d < 1.0 {
		return 1.5*d*d*d - 2.5*d*d + 1.0
	}
	if d < 2.0 {
		return -0.5*d*d*d + 2.5*d*d - 4.0*d + 2.0
	}
	return 0.0
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}
