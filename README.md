## Equirectangular panorama to Cubemap 

Porting c++ to go from https://github.com/denivip/panorama

Convert an equirectangular panorama image into cubemap image. this simple app is written by Go

### Usage
It is possible to convert **JPEG** and **PNG** image format
```
Usage:
  panorama [flags]

Flags:
  -h, --help         help for panorama
  -i, --in string    in image file path (required)
  -l, --len int      edge length of a cube face (default 1024)
  -o, --out string   out file dir path (default ".")
```
```
# example
./panorama --in ./sample_image.jpg --out ./dist --len 512
```

### Installation
```
git clone https://github.com/blackironj/panorama.gitgit clone 

cd panorama

go build -o panorama
```
Or [Download here](https://github.com/blackironj/panorama/releases/tag/1.0)

### TODO
- Optimize code 
> It uses 1 go-routine per each face to convert. (use 6 go-routines)
