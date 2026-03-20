package main

import (
	"bytes"
	"errors"
	"image"
	_ "image/png"
	"unsafe"

	"github.com/cwchiu/go-winapi"
	glib "github.com/mattn/gopher"
)

type sceneInfo struct {
	hBitmap winapi.HBITMAP
	off     int // offset to play walking scene
}

func hBitmapFromImage(img image.Image) (winapi.HBITMAP, error) {
	var bi winapi.BITMAPV5HEADER
	bi.BiSize = uint32(unsafe.Sizeof(bi))
	bi.BiWidth = int32(img.Bounds().Dx())
	bi.BiHeight = -int32(img.Bounds().Dy())
	bi.BiPlanes = 1
	bi.BiBitCount = 32
	bi.BiCompression = winapi.BI_BITFIELDS
	bi.BV4RedMask = 0x00FF0000
	bi.BV4GreenMask = 0x0000FF00
	bi.BV4BlueMask = 0x000000FF
	bi.BV4AlphaMask = 0xFF000000

	hdc := winapi.GetDC(0)
	defer winapi.ReleaseDC(0, hdc)

	var bits unsafe.Pointer
	hBitmap := winapi.CreateDIBSection(
		hdc, &bi.BITMAPINFOHEADER, winapi.DIB_RGB_COLORS, &bits, 0, 0)
	switch hBitmap {
	case 0, winapi.ERROR_INVALID_PARAMETER:
		return 0, errors.New("CreateDIBSection failed")
	}

	ba := (*[1 << 30]byte)(unsafe.Pointer(bits))

	// Fast path: direct pixel access for NRGBA (PNG decode result)
	if nrgba, ok := img.(*image.NRGBA); ok {
		pix := nrgba.Pix
		stride := nrgba.Stride
		bounds := img.Bounds()
		i := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			row := (y - bounds.Min.Y) * stride
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				j := row + (x-bounds.Min.X)*4
				a := pix[j+3]
				if a == 0xFF {
					ba[i+0] = pix[j+2] // B
					ba[i+1] = pix[j+1] // G
					ba[i+2] = pix[j]   // R
					ba[i+3] = 0xFF
				} else if a > 0 {
					// Premultiply alpha (BGRA order)
					a16 := uint16(a)
					ba[i+0] = byte(uint16(pix[j+2]) * a16 / 255)
					ba[i+1] = byte(uint16(pix[j+1]) * a16 / 255)
					ba[i+2] = byte(uint16(pix[j]) * a16 / 255)
					ba[i+3] = a
				} else {
					ba[i+0] = 0
					ba[i+1] = 0
					ba[i+2] = 0
					ba[i+3] = 0
				}
				i += 4
			}
		}
	} else {
		// Fallback: interface-based access
		i := 0
		for y := img.Bounds().Min.Y; y != img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x != img.Bounds().Max.X; x++ {
				r, g, b, a := img.At(x, y).RGBA()
				ba[i+3] = byte(a >> 8)
				ba[i+2] = byte(r >> 8)
				ba[i+1] = byte(g >> 8)
				ba[i+0] = byte(b >> 8)
				i += 4
			}
		}
	}
	return hBitmap, nil
}

func createDIB32(w, h int32) (winapi.HBITMAP, unsafe.Pointer) {
	var bi winapi.BITMAPINFOHEADER
	bi.BiSize = uint32(unsafe.Sizeof(bi))
	bi.BiWidth = w
	bi.BiHeight = -h
	bi.BiPlanes = 1
	bi.BiBitCount = 32
	bi.BiCompression = winapi.BI_RGB

	hdc := winapi.GetDC(0)
	defer winapi.ReleaseDC(0, hdc)

	var bits unsafe.Pointer
	hBitmap := winapi.CreateDIBSection(hdc, &bi, winapi.DIB_RGB_COLORS, &bits, 0, 0)
	return hBitmap, bits
}

func loadImage(filename string) (image.Image, error) {
	b, err := assets.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(b))
	return img, err
}

func reverseImage(img image.Image) image.Image {
	bounds := img.Bounds()
	dx := bounds.Dx()

	// Fast path: direct pixel manipulation for NRGBA
	if nrgba, ok := img.(*image.NRGBA); ok {
		newimg := image.NewNRGBA(bounds)
		for y := 0; y < bounds.Dy(); y++ {
			srcRow := y * nrgba.Stride
			dstRow := y * newimg.Stride
			for x := 1; x < dx; x++ {
				si := srcRow + x*4
				di := dstRow + (dx-x)*4
				newimg.Pix[di] = nrgba.Pix[si]
				newimg.Pix[di+1] = nrgba.Pix[si+1]
				newimg.Pix[di+2] = nrgba.Pix[si+2]
				newimg.Pix[di+3] = nrgba.Pix[si+3]
			}
		}
		return newimg
	}

	// Fallback
	newimg := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y != bounds.Max.Y; y++ {
		for x := bounds.Min.X; x != bounds.Max.X; x++ {
			newimg.Set(dx-x, y, img.At(x, y))
		}
	}
	return newimg
}

func makeGopher() (*Gopher, error) {
	var err error
	var img [8]image.Image

	files := []string{
		"data/out01.png",
		"data/out02.png",
		"data/out03.png",
		"data/waiting.png",
	}

	// make scene 1, 2, 3, and scene waiting
	for i, fname := range files {
		img[i], err = loadImage(fname)
		if err != nil {
			return nil, err
		}
		hBitmap[i], err = hBitmapFromImage(img[i])
		if err != nil {
			return nil, err
		}
	}

	// make reverse step 4, 5, 6, and waiting for right
	for i := 4; i <= 7; i++ {
		img[i] = reverseImage(img[i-4])
		hBitmap[i], err = hBitmapFromImage(img[i])
		if err != nil {
			return nil, err
		}
	}

	bounds := img[0].Bounds()
	var rc winapi.RECT
	winapi.SystemParametersInfo(winapi.SPI_GETWORKAREA, 0, unsafe.Pointer(&rc), 0)

	screenWidth = int(rc.Right) - bounds.Dx()
	scenes = [2][5]sceneInfo{
		{
			{hBitmap[0], 0},
			{hBitmap[1], 2},
			{hBitmap[2], 4},
			{hBitmap[1], 2},
			{hBitmap[3], 0},
		},
		{
			{hBitmap[4], 0},
			{hBitmap[5], 2},
			{hBitmap[6], 4},
			{hBitmap[5], 2},
			{hBitmap[7], 0},
		},
	}

	mode := Walking
	dx := 10
	if *slmode {
		mode = SL
		dx = 20
	}

	// initialize gopher
	return &Gopher{
		task: make(chan glib.Msg, 50),
		x:    -bounds.Dx(),
		y:    int(rc.Bottom) - bounds.Dy(),
		w:    bounds.Dx(),
		h:    bounds.Dy(),
		dx:   dx,
		dy:   0,
		wait: 0,
		mode: mode,
	}, nil
}
