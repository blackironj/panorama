# Interpolation Algorithm Diversification & Performance Optimization

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add nearest/bilinear/bicubic interpolation selection and eliminate PNG/JPEG performance penalty by normalizing to `*image.RGBA` before processing.

**Architecture:** Introduce an `Interpolator` interface with three implementations. Change `ReadImage` to return `*image.RGBA` by adding a `toRGBA` normalization step. Thread the interpolator through the conversion pipeline and expose it via a `--interpolation` CLI flag.

**Tech Stack:** Go 1.25, standard library `image`, `image/draw`, `math`

**Spec:** `docs/specs/2026-03-24-interpolation-and-performance-design.md`

---

### Task 1: RGBA Normalization in ReadImage (Performance Fix)

**Files:**
- Modify: `conv/img.go:22-43`
- Modify: `conv/img_test.go`

- [ ] **Step 1: Write failing test for toRGBA**

Add to `conv/img_test.go`:

```go
func TestToRGBA_AlreadyRGBA(t *testing.T) {
	t.Parallel()
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.Set(0, 0, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	got := toRGBA(src)
	// Should return the same pointer (no copy)
	if got != src {
		t.Error("expected same pointer for *image.RGBA input")
	}
}

func TestToRGBA_FromNRGBA(t *testing.T) {
	t.Parallel()
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	src.Set(0, 0, color.NRGBA{R: 100, G: 150, B: 200, A: 255})
	got := toRGBA(src)
	r, g, b, a := got.At(0, 0).RGBA()
	if r>>8 != 100 || g>>8 != 150 || b>>8 != 200 || a>>8 != 255 {
		t.Errorf("pixel mismatch: got RGBA(%d,%d,%d,%d)", r>>8, g>>8, b>>8, a>>8)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./conv/ -run "TestToRGBA" -v`
Expected: FAIL — `toRGBA` undefined

- [ ] **Step 3: Implement toRGBA**

Add to `conv/img.go`:

```go
import "image/draw"

func toRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return rgba
}
```

- [ ] **Step 4: Update ReadImage to return *image.RGBA**

Change `ReadImage` signature and body in `conv/img.go`:

```go
func ReadImage(imagePath string) (*image.RGBA, string, error) {
	// ... existing decode logic unchanged ...

	if ext == "jpg" || ext == "jpeg" || ext == "png" {
		return toRGBA(imgIn), ext, nil
	}

	return nil, "", errors.New("unsupported image format: " + ext)
}
```

- [ ] **Step 5: Update existing tests for new return type**

In `conv/img_test.go`, update `TestReadImage_JPEG` and `TestReadImage_PNG`:
- Change `img.Bounds()` assertions — these should still work since `*image.RGBA` embeds bounds
- Add assertion that returned type is `*image.RGBA` (it always will be now)

In `conv/convert_test.go`, update `TestInterpolateXYZtoColor` and `TestConvertEquirectangularToCubeMap` and `TestConvert_OutputNonZero`:
- These already use `image.NewRGBA` so no changes needed, but the `ConvertEquirectangularToCubeMap` signature will change in Task 3. For now, keep `image.Image` param.

- [ ] **Step 6: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git add conv/img.go conv/img_test.go
git commit -m "perf: normalize decoded images to *image.RGBA for direct pixel access

Eliminates per-pixel color model conversion overhead that caused PNG
images to process ~5x slower than JPEG (#6). JPEG also benefits from
skipping YCbCr→RGBA conversion in the hot loop."
```

---

### Task 2: Interpolator Interface and Implementations

**Files:**
- Create: `conv/interpolation.go`
- Create: `conv/interpolation_test.go`

- [ ] **Step 1: Write failing tests for all three interpolators**

Create `conv/interpolation_test.go`:

```go
package conv

import (
	"image"
	"image/color"
	"testing"
)

// newUniformRGBA creates a solid-color RGBA image.
func newUniformRGBA(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := range w {
		for y := range h {
			img.Set(x, y, c)
		}
	}
	return img
}

// newGradientRGBA creates a 4x4 image with distinct pixel values for interpolation testing.
func newGradientRGBA() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for x := range 4 {
		for y := range 4 {
			v := uint8(x*60 + y*15)
			img.Set(x, y, color.RGBA{R: v, G: v, B: v, A: 255})
		}
	}
	return img
}

func TestNearestInterpolator_UniformImage(t *testing.T) {
	t.Parallel()
	img := newUniformRGBA(4, 4, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	interp := NearestInterpolator{}
	r, g, b := interp.Interpolate(img, 1.7, 2.3, 4, 4)
	if r != 100 || g != 150 || b != 200 {
		t.Errorf("nearest on uniform: got (%d,%d,%d), want (100,150,200)", r, g, b)
	}
}

func TestNearestInterpolator_PicksClosest(t *testing.T) {
	t.Parallel()
	img := newGradientRGBA()
	interp := NearestInterpolator{}
	// uf=1.9, vf=0.1 should pick pixel (2,0)
	r1, _, _ := interp.Interpolate(img, 1.9, 0.1, 4, 4)
	// pixel (2,0) has value 2*60 + 0*15 = 120
	if r1 != 120 {
		t.Errorf("nearest at (1.9,0.1): got R=%d, want 120", r1)
	}
}

func TestBilinearInterpolator_UniformImage(t *testing.T) {
	t.Parallel()
	img := newUniformRGBA(4, 4, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	interp := BilinearInterpolator{}
	r, g, b := interp.Interpolate(img, 1.5, 1.5, 4, 4)
	if r != 100 || g != 150 || b != 200 {
		t.Errorf("bilinear on uniform: got (%d,%d,%d), want (100,150,200)", r, g, b)
	}
}

func TestBilinearInterpolator_Blends(t *testing.T) {
	t.Parallel()
	// 2x1 image: pixel(0,0)=R:0, pixel(1,0)=R:200
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	img.Set(1, 0, color.RGBA{R: 200, G: 0, B: 0, A: 255})
	interp := BilinearInterpolator{}
	r, _, _ := interp.Interpolate(img, 0.5, 0.0, 2, 1)
	// At midpoint, expect ~100
	if r < 90 || r > 110 {
		t.Errorf("bilinear blend at midpoint: got R=%d, want ~100", r)
	}
}

func TestBicubicInterpolator_UniformImage(t *testing.T) {
	t.Parallel()
	img := newUniformRGBA(8, 8, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	interp := BicubicInterpolator{}
	r, g, b := interp.Interpolate(img, 4.0, 4.0, 8, 8)
	if r != 100 || g != 150 || b != 200 {
		t.Errorf("bicubic on uniform: got (%d,%d,%d), want (100,150,200)", r, g, b)
	}
}

func TestBicubicInterpolator_Blends(t *testing.T) {
	t.Parallel()
	img := image.NewRGBA(image.Rect(0, 0, 4, 1))
	img.Set(0, 0, color.RGBA{R: 0, A: 255})
	img.Set(1, 0, color.RGBA{R: 0, A: 255})
	img.Set(2, 0, color.RGBA{R: 200, A: 255})
	img.Set(3, 0, color.RGBA{R: 200, A: 255})
	interp := BicubicInterpolator{}
	r, _, _ := interp.Interpolate(img, 1.5, 0.0, 4, 1)
	// Should produce a blended value between 0 and 200
	if r == 0 || r == 200 {
		t.Errorf("bicubic should blend, got R=%d", r)
	}
}

func TestParseInterpolation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"nearest", false},
		{"bilinear", false},
		{"bicubic", false},
		{"invalid", true},
		{"", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseInterpolation(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInterpolation(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./conv/ -run "TestNearest|TestBilinear|TestBicubic|TestParseInterpolation" -v`
Expected: FAIL — types undefined

- [ ] **Step 3: Implement interpolation.go**

Create `conv/interpolation.go`:

```go
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

// readPixel reads RGBA values directly from the Pix slice, avoiding interface dispatch.
func readPixel(img *image.RGBA, x, y int) (r, g, b float64) {
	off := (y-img.Rect.Min.Y)*img.Stride + (x-img.Rect.Min.X)*4
	return float64(img.Pix[off]), float64(img.Pix[off+1]), float64(img.Pix[off+2])
}

// NearestInterpolator samples the single closest pixel.
type NearestInterpolator struct{}

func (NearestInterpolator) Interpolate(img *image.RGBA, uf, vf float64, sw, sh int) (r, g, b uint8) {
	x := safeIndex(math.Round(uf), float64(sw))
	y := safeIndex(math.Round(vf), float64(sh))
	pr, pg, pb := readPixel(img, x, y)
	return uint8(pr), uint8(pg), uint8(pb)
}

// BilinearInterpolator uses weighted average of 4 surrounding pixels.
type BilinearInterpolator struct{}

func (BilinearInterpolator) Interpolate(img *image.RGBA, uf, vf float64, sw, sh int) (r, g, b uint8) {
	ui := safeIndex(math.Floor(uf), float64(sw))
	vi := safeIndex(math.Floor(vf), float64(sh))
	u2 := safeIndex(float64(ui)+1.0, float64(sw))
	v2 := safeIndex(float64(vi)+1.0, float64(sh))

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
			px := safeIndex(float64(ui+n), float64(sw))
			py := safeIndex(float64(vi+m), float64(sh))
			pr, pg, pb := readPixel(img, px, py)
			w := wu * wv
			sr += pr * w
			sg += pg * w
			sb += pb * w
		}
	}

	return uint8(clamp(sr)), uint8(clamp(sg)), uint8(clamp(sb))
}

// cubicWeight computes the Catmull-Rom cubic kernel weight.
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
	return math.Max(0, math.Min(255, v))
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./conv/ -run "TestNearest|TestBilinear|TestBicubic|TestParseInterpolation" -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add conv/interpolation.go conv/interpolation_test.go
git commit -m "feat: add Interpolator interface with nearest, bilinear, bicubic implementations

Introduces pluggable interpolation strategy with direct Pix slice access
for performance. Nearest is fastest, bicubic uses Catmull-Rom kernel for
highest quality."
```

---

### Task 3: Wire Interpolator into Conversion Pipeline

**Files:**
- Modify: `conv/convert.go:45-137`
- Modify: `conv/convert_test.go`

- [ ] **Step 1: Update ConvertEquirectangularToCubeMap signature**

In `conv/convert.go`, change:

```go
func ConvertEquirectangularToCubeMap(edgeLen int, imgIn image.Image, sides []string) ([]*image.RGBA, error) {
```

to:

```go
func ConvertEquirectangularToCubeMap(edgeLen int, imgIn *image.RGBA, sides []string, interp Interpolator) ([]*image.RGBA, error) {
```

- [ ] **Step 2: Update convert function**

Change `convert` to accept `Interpolator` and `*image.RGBA`:

```go
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
			imgOut.Pix[(j*edge+i)*4] = r
			imgOut.Pix[(j*edge+i)*4+1] = g
			imgOut.Pix[(j*edge+i)*4+2] = b
			imgOut.Pix[(j*edge+i)*4+3] = 255
		}
	}
	return nil
}
```

- [ ] **Step 3: Update goroutine call in ConvertEquirectangularToCubeMap**

Pass `interp` through:

```go
go func(idx, side int, canvas *image.RGBA) {
    defer wg.Done()
    errs[idx] = convert(edgeLen, side, sw, sh, imgIn, canvas, interp)
}(i, sidesInt[i], canvases[i])
```

- [ ] **Step 4: Remove old interpolateXYZtoColor function**

Delete `interpolateXYZtoColor`, `mix`, `Vec3`, and `Number` from `convert.go` — these are replaced by the `Interpolator` implementations. Keep `outImgToXYZ`, `safeIndex`, and `piHalf`.

- [ ] **Step 5: Update convert_test.go**

Update all test functions that call `ConvertEquirectangularToCubeMap` or `convert` to pass a `BilinearInterpolator{}` and use `*image.RGBA` input:

- `TestConvertEquirectangularToCubeMap`: add `BilinearInterpolator{}` arg
- `TestConvertEquirectangularToCubeMap_AllSides`: add `BilinearInterpolator{}` arg
- `TestConvert_OutputNonZero`: add `BilinearInterpolator{}` arg
- Remove `TestInterpolateXYZtoColor` (function no longer exists)
- Remove `TestMix` (function no longer exists)

- [ ] **Step 6: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git add conv/convert.go conv/convert_test.go
git commit -m "refactor: wire Interpolator into conversion pipeline

ConvertEquirectangularToCubeMap now accepts *image.RGBA and Interpolator.
Coordinate mapping (XYZ→spherical→equirectangular) moved into convert(),
interpolation delegated to Interpolator. Removes old interpolateXYZtoColor."
```

---

### Task 4: Add --interpolation CLI Flag

**Files:**
- Modify: `cmd/root.go:27-34,58-84,98-118`
- Modify: `cmd/root_test.go`

- [ ] **Step 1: Write test for resolveInterpolation**

Add to `cmd/root_test.go`:

```go
func TestResolveInterpolation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"nearest", false},
		{"bilinear", false},
		{"bicubic", false},
		{"invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := conv.ParseInterpolation(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInterpolation(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
```

- [ ] **Step 2: Add flag and wire through**

In `cmd/root.go`:

Add variable:
```go
var interpolation string
```

In `init()`, add:
```go
rootCmd.Flags().StringVarP(&interpolation, "interpolation", "p", "bilinear", "interpolation method: nearest, bilinear, bicubic")
```

In `run()`, after `resolveTargetSides`:
```go
interp, err := conv.ParseInterpolation(interpolation)
if err != nil {
    exitWithError(err)
}
```

Update `processSingleImage` to accept `conv.Interpolator` and pass it to `ConvertEquirectangularToCubeMap`:

```go
func processSingleImage(inPath, outDir string, targetSides []string, interp conv.Interpolator, needSubdir bool) error {
    inImage, ext, err := conv.ReadImage(inPath)
    // ...
    canvases, err := conv.ConvertEquirectangularToCubeMap(edgeLen, inImage, targetSides, interp)
    // ...
}
```

Update all callers of `processSingleImage` in `run()` and `processDirectory()` to pass `interp`.

- [ ] **Step 3: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 4: Build and verify CLI help**

Run: `go build -o panorama . && ./panorama --help`
Expected: Shows `--interpolation` / `-p` flag with description

- [ ] **Step 5: Commit**

```bash
git add cmd/root.go cmd/root_test.go
git commit -m "feat: add --interpolation flag for algorithm selection

Users can now choose nearest (fast), bilinear (default), or bicubic
(high quality) interpolation via -p flag."
```

---

### Task 5: Benchmarks

**Files:**
- Create: `conv/benchmark_test.go`

- [ ] **Step 1: Write benchmarks**

Create `conv/benchmark_test.go`:

```go
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
```

- [ ] **Step 2: Run benchmarks**

Run: `go test ./conv/ -bench=. -benchmem -count=3`
Expected: Results showing relative speed of each algorithm

- [ ] **Step 3: Commit**

```bash
git add conv/benchmark_test.go
git commit -m "test: add benchmarks for interpolation algorithms and RGBA normalization"
```

---

### Task 6: Final Verification

- [ ] **Step 1: Run full test suite**

Run: `go test ./... -v -count=1`
Expected: ALL PASS

- [ ] **Step 2: Run linter**

Run: `go vet ./...`
Expected: No issues

- [ ] **Step 3: Build and smoke test**

Run: `go build -o panorama .`
Verify: `./panorama --help` shows all flags including `--interpolation`
