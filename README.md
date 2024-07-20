# Equirectangular panorama to Cubemap

Porting c++ to go from <https://github.com/denivip/panorama>

Convert an equirectangular panorama image into cubemap image. this simple app is written by Go

## Screenshot

![example](https://user-images.githubusercontent.com/43738420/112742708-bf90c100-8fcb-11eb-8159-cecaf834ef2c.png)

### Usage

It is possible to convert **JPEG** and **PNG** image format

``` sh
Usage:
  panorama [flags]

Flags:
  -h, --help         help for panorama
  -i, --in string    in image file path (required)
  -l, --len int      edge length of a cube face (default 1024)
  -o, --out string   out file dir path (default ".")
  -s, --sides array  list of sides splited by "," (optional)
```

``` sh
# example
./panorama --in ./sample_image.jpg --out ./dist --len 512 --sides left,right,top,buttom,front,back
```

### Installation

``` sh
git clone https://github.com/blackironj/panorama.gitgit clone

cd panorama

go build -o panorama
```

Or [Download here](https://github.com/blackironj/panorama/releases/tag/1.0)

### TODO

- Optimize code
  - It uses 1 go-routine per each face to convert. (use 6 go-routines)
- Add more interpolation algorithms
