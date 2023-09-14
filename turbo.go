package cimg

// The darwin cgo paths assume you've installed jpeg-turbo using homebrew

/*
#cgo pkg-config: libturbojpeg
#include <turbojpeg.h>
*/
import "C"

import (
	"bytes"
	"fmt"
	"unsafe"
)

type Sampling C.int

const (
	Sampling444  Sampling = C.TJSAMP_444
	Sampling422  Sampling = C.TJSAMP_422
	Sampling420  Sampling = C.TJSAMP_420
	SamplingGray Sampling = C.TJSAMP_GRAY
	Sampling440  Sampling = C.TJSAMP_440
	Sampling411  Sampling = C.TJSAMP_411
)

type PixelFormat C.int

const (
	PixelFormatRGB     PixelFormat = C.TJPF_RGB
	PixelFormatBGR     PixelFormat = C.TJPF_BGR
	PixelFormatRGBX    PixelFormat = C.TJPF_RGBX
	PixelFormatBGRX    PixelFormat = C.TJPF_BGRX
	PixelFormatXBGR    PixelFormat = C.TJPF_XBGR
	PixelFormatXRGB    PixelFormat = C.TJPF_XRGB
	PixelFormatGRAY    PixelFormat = C.TJPF_GRAY
	PixelFormatRGBA    PixelFormat = C.TJPF_RGBA
	PixelFormatBGRA    PixelFormat = C.TJPF_BGRA
	PixelFormatABGR    PixelFormat = C.TJPF_ABGR
	PixelFormatARGB    PixelFormat = C.TJPF_ARGB
	PixelFormatCMYK    PixelFormat = C.TJPF_CMYK
	PixelFormatUNKNOWN PixelFormat = C.TJPF_UNKNOWN
)

type Flags C.int

const (
	FlagAccurateDCT   Flags = C.TJFLAG_ACCURATEDCT
	FlagBottomUp      Flags = C.TJFLAG_BOTTOMUP
	FlagFastDCT       Flags = C.TJFLAG_FASTDCT
	FlagFastUpsample  Flags = C.TJFLAG_FASTUPSAMPLE
	FlagNoRealloc     Flags = C.TJFLAG_NOREALLOC
	FlagProgressive   Flags = C.TJFLAG_PROGRESSIVE
	FlagStopOnWarning Flags = C.TJFLAG_STOPONWARNING
)

func makeError(handler C.tjhandle, returnVal C.int) error {
	if returnVal == 0 {
		return nil
	}
	str := C.GoString(C.tjGetErrorStr2(handler))
	return fmt.Errorf("turbojpeg error: %v", str)
}

// CompressParams are the TurboJPEG compression parameters
type CompressParams struct {
	Sampling Sampling
	Quality  int // 1 .. 100
	Flags    Flags
}

// MakeCompressParams returns a fully populated CompressParams struct
func MakeCompressParams(sampling Sampling, quality int, flags Flags) CompressParams {
	return CompressParams{
		Sampling: sampling,
		Quality:  quality,
		Flags:    flags,
	}
}

// Compress compresses an image using TurboJPEG
func Compress(img *Image, params CompressParams) ([]byte, error) {
	encoder := C.tjInitCompress()
	defer C.tjDestroy(encoder)

	var outBuf *C.uchar
	var outBufSize C.ulong

	if img.Format == PixelFormatGRAY {
		// This is the only valid sampling, so just fix it up if the user screwed it up
		params.Sampling = SamplingGray
	}

	// int tjCompress2(tjhandle handle, const unsigned char *srcBuf, int width, int pitch, int height, int pixelFormat,
	// unsigned char **jpegBuf, unsigned long *jpegSize, int jpegSubsamp, int jpegQual, int flags);
	res := C.tjCompress2(
		encoder,
		(*C.uchar)(&img.Pixels[0]),
		C.int(img.Width),
		C.int(img.Stride),
		C.int(img.Height),
		C.int(img.Format),
		&outBuf,
		&outBufSize,
		C.int(params.Sampling),
		C.int(params.Quality),
		C.int(params.Flags),
	)

	var enc []byte
	err := makeError(encoder, res)
	if outBuf != nil {
		enc = C.GoBytes(unsafe.Pointer(outBuf), C.int(outBufSize))
		C.tjFree(outBuf)
	}

	if err != nil {
		return nil, err
	}
	return enc, nil
}

// Load an image into memory.
// JPEG: Uses TurboJPEG
// PNG: Uses Go's native PNG library
// TIFF: Uses golang.org/x/image/tiff
// The resulting image is RGB for JPEGs, or RGBA/Gray for PNG
func Decompress(encoded []byte, outFormat PixelFormat) (*Image, error) {
	if len(encoded) > 4 && bytes.Compare(encoded[:4], []byte("II*\x00")) == 0 {
		return decompressTIFF(encoded)
	}
	if len(encoded) > 8 && bytes.Compare(encoded[:8], []byte("\x89\x50\x4e\x47\x0d\x0a\x1a\x0a")) == 0 {
		return decompressPNG(encoded)
	}

	decoder := C.tjInitDecompress()
	defer C.tjDestroy(decoder)

	width := C.int(0)
	height := C.int(0)
	sampling := C.int(0)
	colorspace := C.int(0)

	err := makeError(
		decoder,
		C.tjDecompressHeader3(
			decoder,
			(*C.uchar)(&encoded[0]),
			C.ulong(len(encoded)),
			&width,
			&height,
			&sampling,
			&colorspace,
		),
	)
	if err != nil {
		return nil, err
	}

	pixelSize := NChan(outFormat)
	outBuf := make([]byte, width*height*C.int(pixelSize))
	stride := C.int(width * C.int(pixelSize))

	// int tjDecompress2(tjhandle handle, const unsigned char *jpegBuf, unsigned long jpegSize, unsigned char *dstBuf,
	// int width, int pitch, int height, int pixelFormat, int flags);
	err = makeError(
		decoder,
		C.tjDecompress2(
			decoder,
			(*C.uchar)(&encoded[0]),
			C.ulong(len(encoded)),
			(*C.uchar)(&outBuf[0]),
			width,
			stride,
			height,
			C.int(outFormat),
			0,
		),
	)
	if err != nil {
		return nil, err
	}

	img := &Image{
		Width:  int(width),
		Height: int(height),
		Stride: int(stride),
		Format: outFormat,
		Pixels: outBuf,
	}
	return img, nil
}

func alignFloor(value, base int) int {
	return ((value) & ^((base) - 1))
}
func alignCeil(value, base int) int {
	return alignFloor((value)+((base)-1), base)
}

func alignRound(value, base int) int {
	return alignFloor((value)+((base)/2), base)
}

// JPEG image CROP using TurboJPEG
func Transform(jpegBytes []byte, x, y, w, h int, jpegSampling Sampling, flags Flags) ([]byte, error) {
	// 初始化句柄
	decoder := C.tjInitTransform()
	defer C.tjDestroy(decoder)

	// 坐标与16对齐
	alignX := alignRound(x, int(C.tjMCUWidth[jpegSampling]))
	alignY := alignRound(y, int(C.tjMCUHeight[jpegSampling]))
	fixedW := w + (x - alignX)
	fixedH := h + (y - alignY)

	// 裁剪参数
	var xform C.tjtransform
	xform.r.x = C.int(alignX)
	xform.r.y = C.int(alignY)
	xform.r.w = C.int(fixedW)
	xform.r.h = C.int(fixedH)
	xform.options |= C.TJXOPT_CROP

	// 声明输出参数
	var dstBuf *C.uchar
	var outBufSize C.ulong
	defer func() {
		// 是否需要释放
		if dstBuf != nil {
			C.tjFree(dstBuf)
		}
	}()

	// 执行裁剪
	err := makeError(decoder, C.tjTransform(
		decoder,
		(*C.uchar)(&jpegBytes[0]),
		C.ulong(len(jpegBytes)),
		1,
		&dstBuf,
		&outBufSize,
		&xform,
		C.int(flags),
	))
	if err != nil {
		return nil, err
	}

	// 提取裁剪图像
	datImg := C.GoBytes(unsafe.Pointer(dstBuf), C.int(outBufSize))

	// OK
	return datImg, nil
}
