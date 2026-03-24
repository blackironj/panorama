# Interpolation Algorithm Diversification & PNG Performance Optimization

**Date:** 2026-03-24
**Status:** Approved
**Related Issue:** https://github.com/blackironj/panorama/issues/6

## Problem

1. Only bilinear interpolation is available. Users need faster (nearest neighbor) or higher quality (bicubic) options.
2. PNG images process ~5x slower than JPEG at the same resolution (6720x3360: JPG 4.5s vs PNG 24s). Root cause: `image.At()` on non-RGBA types triggers color model conversion on every call, and the conversion loop calls it millions of times.

## Goals

- Support 3 interpolation algorithms: nearest neighbor, bilinear, bicubic
- User selects via CLI flag `--interpolation` / `-p` (default: `bilinear`)
- Eliminate PNG performance penalty by normalizing all images to `*image.RGBA` before processing
- Maintain backward compatibility (default behavior unchanged)

## Design

### 1. Interpolator Interface

New file: `conv/interpolation.go`

```go
type Interpolator interface {
    Interpolate(img *image.RGBA, uf, vf float64, sw, sh int) (r, g, b uint8)
}
```

Three implementations:

- **NearestInterpolator** — Samples the single closest pixel. Fastest, lowest quality. Useful for previews or when speed matters more than visual fidelity.
- **BilinearInterpolator** — Weighted average of 4 surrounding pixels (current behavior). Good balance of speed and quality.
- **BicubicInterpolator** — Weighted average of 16 surrounding pixels using cubic kernel. Highest quality, sharpest edges, ~2-3x slower than bilinear.

All implementations access `image.RGBA.Pix` slice directly for maximum throughput (no interface dispatch on pixel access).

### 2. RGBA Normalization

New function in `conv/img.go`:

```go
func toRGBA(img image.Image) *image.RGBA
```

- If the image is already `*image.RGBA`, return as-is (no copy)
- Otherwise, draw into a new `*image.RGBA` canvas using `draw.Draw`
- Called once after `ReadImage()`, before the conversion loop
- This eliminates per-pixel color model conversion that causes the PNG slowdown

`ReadImage` return type changes from `image.Image` to `*image.RGBA`.

### 3. Convert Pipeline Changes

`conv/convert.go`:

- `ConvertEquirectangularToCubeMap` accepts an `Interpolator` parameter
- `convert()` accepts an `Interpolator` parameter
- The interpolation call in the inner loop delegates to the `Interpolator` instead of calling `interpolateXYZtoColor` directly
- `outImgToXYZ` remains unchanged

### 4. CLI Changes

`cmd/root.go`:

- New flag: `--interpolation` / `-p` with values `nearest`, `bilinear`, `bicubic` (default: `bilinear`)
- Resolve the flag value to an `Interpolator` instance before calling conversion

### 5. File Changes

| File | Change |
|------|--------|
| `conv/interpolation.go` | New — Interpolator interface + 3 implementations |
| `conv/interpolation_test.go` | New — Unit tests for each algorithm |
| `conv/convert.go` | Modified — Accept Interpolator parameter |
| `conv/convert_test.go` | Modified — Pass Interpolator in tests |
| `conv/img.go` | Modified — Add `toRGBA()`, change `ReadImage` return type |
| `conv/img_test.go` | Modified — Update for new return type |
| `cmd/root.go` | Modified — Add `--interpolation` flag |
| `cmd/root_test.go` | Modified — Test interpolation flag resolution |

### 6. Testing Strategy

- Unit tests for each interpolator with known input/output pixel values
- Benchmark tests comparing all 3 algorithms at a fixed image size
- Benchmark test comparing RGBA vs non-RGBA pixel access to validate the performance fix
- Existing tests updated to pass an interpolator parameter
