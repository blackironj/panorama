# Equirectangular panorama to Cubemap

Porting c++ to go from <https://github.com/denivip/panorama>

Convert an equirectangular panorama image into cubemap image. this simple app is written by Go

## Screenshot

![example](https://user-images.githubusercontent.com/43738420/112742708-bf90c100-8fcb-11eb-8159-cecaf834ef2c.png)
> Image source: <a href="https://unsplash.com/@oldfieldart?utm_content=creditCopyText&utm_medium=referral&utm_source=unsplash">Timothy Oldfield</a> on <a href="https://unsplash.com/photos/blue-and-gray-docks-luufnHoChRU?utm_content=creditCopyText&utm_medium=referral&utm_source=unsplash">Unsplash</a>

### Usage

It is possible to convert **JPEG** and **PNG** image format

``` sh
Usage:
  panorama [flags]

Flags:
  -h, --help         help for panorama
  -i, --in string    input image file path (required if --indir is not specified)
  -d, --indir string input directory path (required if --in is not specified)
  -l, --len int      edge length of a cube face (default 1024)
  -o, --out string   out file dir path (default ".")
  -s, --sides array  list of sides splited by "," (optional)
  -q, --quality int  jpeg file output quality ranges from 1 to 100 inclusive, higher is better (optional, default 75)
```

``` sh
# example
./panorama --in ./sample_image.jpg --out ./dist --len 512 --sides left,right,top,bottom,front,back
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
