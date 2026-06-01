# CLAUDE.md

Equirectangular panorama → cubemap converter. Go CLI built with cobra.

## Commands

```sh
go build -o panorama .   # build binary
go test ./...            # run all tests
go test -race ./...      # race detector (CI runs this)
go vet ./...             # vet (CI runs this)
go test -bench=. ./conv  # benchmarks (conv/benchmark_test.go)
```

CI (`.github/workflows/ci.yml`) runs build + `test -race` + vet on Linux/macOS/Windows with Go 1.25. Release workflow cross-compiles binaries on GitHub release publish.

## Architecture

- `main.go` → `cmd.Execute()`.
- `cmd/root.go` — CLI: flag parsing, validation, single-file vs directory dispatch. Directory mode processes files concurrently with a `runtime.NumCPU()`-bounded semaphore and live progress reporting to stderr.
- `conv/` — core conversion library:
  - `convert.go` — `ConvertEquirectangularToCubeMap`: spawns one goroutine per requested face, maps each output pixel to spherical coords, samples via the `Interpolator`.
  - `interpolation.go` — `Interpolator` interface + `Nearest`/`Bilinear`/`Bicubic` implementations; `ParseInterpolation(name)`.
  - `img.go` — `ReadImage`/`WriteImage`, format handling, `ValidSides()`.

## Conventions & Gotchas

- Follow the Uber Go Style Guide.
- Face order is fixed by integer constants in `convert.go` (`faceBack=0` … `faceBottom=5`); `reversedFaceMap` in `img.go` maps side names → these indices. Keep both in sync.
- `image.Decode` reports JPEG as `"jpeg"`; on write, `"jpg"` is normalized to `"jpeg"`. Supported formats: jpg/jpeg/png only.
- `safeIndex` clamps source coordinates to image bounds (edge pixels repeat) — interpolators rely on this instead of wrapping.
- Default interpolation is `bilinear`; default JPEG quality is 75 (valid 1–100).
- `testdir/` holds sample panorama images for manual runs.
