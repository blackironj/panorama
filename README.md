# Equirectangular Panorama to Cubemap

[![CI](https://github.com/blackironj/panorama/actions/workflows/ci.yml/badge.svg)](https://github.com/blackironj/panorama/actions/workflows/ci.yml)

Convert equirectangular panorama images into cubemap images. Written in Go.

Inspired by [denivip/panorama](https://github.com/denivip/panorama).

## Screenshot

![example](https://user-images.githubusercontent.com/43738420/112742708-bf90c100-8fcb-11eb-8159-cecaf834ef2c.png)
> Image source: [Timothy Oldfield](https://unsplash.com/@oldfieldart) on [Unsplash](https://unsplash.com/photos/blue-and-gray-docks-luufnHoChRU)

## Features

- Supports **JPEG** and **PNG** input/output
- Three interpolation algorithms: **nearest**, **bilinear** (default), **bicubic**
- Selective face extraction (front, back, left, right, top, bottom)
- Batch directory processing with concurrent execution
- Configurable output quality and cube face size

## Usage

```sh
panorama [flags]

Flags:
  -i, --in string              input image file path (required if --indir is not specified)
  -d, --indir string           input directory path (required if --in is not specified)
  -o, --out string             output file directory path (default ".")
  -l, --len int                edge length of a cube face (default 1024)
  -s, --sides strings          list of sides: front,back,left,right,top,bottom (default: all)
  -q, --quality int            jpeg output quality, 1-100 (default 75)
  -p, --interpolation string   interpolation method: nearest, bilinear, bicubic (default "bilinear")
  -h, --help                   help for panorama
```

### Examples

```sh
# Basic conversion
./panorama -i ./sample.jpg -o ./dist

# High quality bicubic with 2048px faces
./panorama -i ./sample.jpg -o ./dist -l 2048 -p bicubic -q 95

# Only front and back faces
./panorama -i ./sample.jpg -o ./dist -s front,back

# Batch process a directory
./panorama -d ./input_dir -o ./dist
```

## Installation

### Build from source

```sh
git clone https://github.com/blackironj/panorama.git
cd panorama
go build -o panorama
```

### Download

Pre-built binaries are available on the [Releases](https://github.com/blackironj/panorama/releases) page.

Supported platforms: Linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64).

## Interpolation Methods

| Method | Speed | Quality | Description |
|--------|-------|---------|-------------|
| `nearest` | Fastest | Low | Nearest pixel sampling, good for previews |
| `bilinear` | Balanced | Medium | 4-pixel weighted average (default) |
| `bicubic` | Slowest | High | 16-pixel Catmull-Rom kernel, sharpest output |

## License

[MIT](LICENSE)
