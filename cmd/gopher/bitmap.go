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
	hRgn    winapi.HRGN
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
	return hBitmap, nil
}

func loadImage(filename string) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewBuffer(MustAsset(filename)))
	return img, err
}

func reverseImage(img image.Image) image.Image {
	// make reversed image
	newimg := image.NewRGBA(img.Bounds())
	Dx := img.Bounds().Dx()
	for y := img.Bounds().Min.Y; y != img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x != img.Bounds().Max.X; x++ {
			newimg.Set(Dx-x, y, img.At(x, y))
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

	// make scene 1, 2, 3, and schene waiting
	for i, fname := range files {
		img[i], err = loadImage(fname)
		if err != nil {
			return nil, err
		}
		hBitmap[i], err = hBitmapFromImage(img[i])
		if err != nil {
			return nil, err
		}

		// create region for window
		hRgn[i] = winapi.CreateRectRgn(0, 0, 0, 0)
		for y := img[i].Bounds().Min.Y; y != img[i].Bounds().Max.Y; y++ {
			for x := img[i].Bounds().Min.X; x != img[i].Bounds().Max.X; x++ {
				_, _, _, a := img[i].At(x, y).RGBA()
				// combine transparent colors
				if a > 0 {
					mask := winapi.CreateRectRgn(int32(x), int32(y), int32(x+1), int32(y+1))
					winapi.CombineRgn(hRgn[i], mask, hRgn[i], winapi.RGN_OR)
					winapi.DeleteObject(winapi.HGDIOBJ(mask))
				}
			}
		}
	}

	// make reverse step 4, 5, 6, and waiting for right
	for i := 4; i <= 7; i++ {
		img[i] = reverseImage(img[i-4])
		hBitmap[i], err = hBitmapFromImage(img[i])
		if err != nil {
			return nil, err
		}

		hRgn[i] = winapi.CreateRectRgn(0, 0, 0, 0)
		for y := img[i].Bounds().Min.Y; y != img[i].Bounds().Max.Y; y++ {
			for x := img[i].Bounds().Min.X; x != img[i].Bounds().Max.X; x++ {
				_, _, _, a := img[i].At(x, y).RGBA()
				if a > 0 {
					mask := winapi.CreateRectRgn(int32(x), int32(y), int32(x+1), int32(y+1))
					winapi.CombineRgn(hRgn[i], mask, hRgn[i], winapi.RGN_OR)
					winapi.DeleteObject(winapi.HGDIOBJ(mask))
				}
			}
		}
	}

	bounds := img[0].Bounds()
	var rc winapi.RECT
	winapi.SystemParametersInfo(winapi.SPI_GETWORKAREA, 0, unsafe.Pointer(&rc), 0)

	screenWidth = int(rc.Right) - bounds.Dx()
	scenes = [2][5]sceneInfo{
		{
			{hBitmap[0], hRgn[0], 0},
			{hBitmap[1], hRgn[1], 2},
			{hBitmap[2], hRgn[2], 4},
			{hBitmap[1], hRgn[1], 2},
			{hBitmap[3], hRgn[3], 0},
		},
		{
			{hBitmap[4], hRgn[4], 0},
			{hBitmap[5], hRgn[5], 2},
			{hBitmap[6], hRgn[6], 4},
			{hBitmap[5], hRgn[5], 2},
			{hBitmap[7], hRgn[7], 0},
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
