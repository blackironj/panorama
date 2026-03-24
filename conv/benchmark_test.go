package conv

import (
	"image"
	"image/color"
	"testing"
)

func newBenchImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := range w {
		for y := range h {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: uint8((x + y) % 256),
				A: 255,
			})
		}
	}
	return img
}

func benchmarkInterpolator(b *testing.B, interp Interpolator, edgeLen int) {
	img := newBenchImage(edgeLen*2, edgeLen)
	sides := []string{"front"}
	b.ResetTimer()
	for range b.N {
		_, _ = ConvertEquirectangularToCubeMap(edgeLen, img, sides, interp)
	}
}

func BenchmarkNearest_256(b *testing.B)  { benchmarkInterpolator(b, NearestInterpolator{}, 256) }
func BenchmarkBilinear_256(b *testing.B) { benchmarkInterpolator(b, BilinearInterpolator{}, 256) }
func BenchmarkBicubic_256(b *testing.B)  { benchmarkInterpolator(b, BicubicInterpolator{}, 256) }

func BenchmarkNearest_1024(b *testing.B)  { benchmarkInterpolator(b, NearestInterpolator{}, 1024) }
func BenchmarkBilinear_1024(b *testing.B) { benchmarkInterpolator(b, BilinearInterpolator{}, 1024) }
func BenchmarkBicubic_1024(b *testing.B)  { benchmarkInterpolator(b, BicubicInterpolator{}, 1024) }

func BenchmarkRGBAvsNRGBA(b *testing.B) {
	rgba := newBenchImage(512, 256)

	nrgba := image.NewNRGBA(image.Rect(0, 0, 512, 256))
	for x := range 512 {
		for y := range 256 {
			nrgba.Set(x, y, color.NRGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: uint8((x + y) % 256),
				A: 255,
			})
		}
	}
	converted := toRGBA(nrgba)

	interp := BilinearInterpolator{}
	sides := []string{"front"}

	b.Run("DirectRGBA", func(b *testing.B) {
		for range b.N {
			_, _ = ConvertEquirectangularToCubeMap(256, rgba, sides, interp)
		}
	})
	b.Run("ConvertedNRGBA", func(b *testing.B) {
		for range b.N {
			_, _ = ConvertEquirectangularToCubeMap(256, converted, sides, interp)
		}
	})
}
