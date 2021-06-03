# C++ image library bindings for Go

`cimg` is a Go wrapper for various C/C++ image libraries, including:

* TurboJPEG
* stb_image_resize
* libexif (EXIF is gratuitous here - you could just as well use one of the native Go EXIF libraries)
* Unrotate image so that natural encoding orientation is same as display orientation

Why?

There are a host of high-performance C/C++ libraries out there for
image manipulation, and it's near impossible to write the same kind
of code in Go.

### Example: Compress/Decompress with TurboJPEG

```go
import "github.com/bmharper/cimg"

func compressImage(width, height int, rgba []byte) {
	raw := cimg.Image{
		Width: width,
		Height: height,
		Stride: width * 4,
		Pixels: rgba,
	}
	params := cimg.MakeCompressParams(cimg.PixelFormatRGBA, cimg.Sampling420, 35, 0)
	jpg, err := cimg.Compress(&raw, params)
}

func decompressImage(jpg []byte) (*Image, error) {
	return cimg.Decompress(jpg)
}
```

### Example: Resize with stb_image_resize

```go
import "github.com/bmharper/cimg"

// Resize from bytes
func resizeImage(srcWidth, srcHeight int, rgba []byte, dstWidth, dstHeight int) *cimg.Image {
	src := cimg.WrapImage(srcWidth, srcHeight, 4, rgba)
	return cimg.ResizeNew(src, dstWidth, dstHeight)
}
```

### C/C++ compiler optimizations
I was initially worried that I needed to add the directive `#cgo CXXFLAGS: -O2`, but it looks like
`cgo` compiles with optimizations on by default. You can verify this by adding `#cgo CXXFLAGS: -O0`
to `resize.go`, and run `go test -bench=.`. Compare with `-O0` and `-O2` and there should be
a big difference. Removing the comment entirely should give similar performance to `-O2`.

### System requirements
I have only tested this on Ubuntu 20.04 `amd64`.

To install the necessary packages:
```
apt install libexif-dev libturbojpeg0-dev
```

### Testing
Warning! Many of the Go unit tests don't actually validate their results. Instead, they
simply write out a JPEG file into the `test` directory. It's your job to visually see that they look correct.
