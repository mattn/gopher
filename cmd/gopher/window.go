package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"syscall"
	"time"
	"unsafe"

	"github.com/cwchiu/go-winapi"
	gopherlib "github.com/mattn/gopher"
)

const (
	DT_TOP                  = 0x00000000
	DT_LEFT                 = 0x00000000
	DT_CENTER               = 0x00000001
	DT_RIGHT                = 0x00000002
	DT_VCENTER              = 0x00000004
	DT_BOTTOM               = 0x00000008
	DT_WORDBREAK            = 0x00000010
	DT_SINGLELINE           = 0x00000020
	DT_EXPANDTABS           = 0x00000040
	DT_TABSTOP              = 0x00000080
	DT_NOCLIP               = 0x00000100
	DT_EXTERNALLEADING      = 0x00000200
	DT_CALCRECT             = 0x00000400
	DT_NOPREFIX             = 0x00000800
	DT_INTERNAL             = 0x00001000
	DT_EDITCONTROL          = 0x00002000
	DT_PATH_ELLIPSIS        = 0x00004000
	DT_END_ELLIPSIS         = 0x00008000
	DT_MODIFYSTRING         = 0x00010000
	DT_RTLREADING           = 0x00020000
	DT_WORD_ELLIPSIS        = 0x00040000
	DT_NOFULLWIDTHCHARBREAK = 0x00080000
	DT_HIDEPREFIX           = 0x00100000
	DT_PREFIXONLY           = 0x00200000
	LWA_COLORKEY            = 0x00001
	LWA_ALPHA               = 0x00002
)

type tagCOPYDATASTRUCT struct {
	DwData uintptr
	CbData uint32
	LpData unsafe.Pointer
}

var (
	user32                         = syscall.NewLazyDLL("user32.dll")
	procSetWindowRgn               = user32.NewProc("SetWindowRgn")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	procReplyMessage               = user32.NewProc("ReplyMessage")
	procDrawText                   = user32.NewProc("DrawTextW")
	shell32                        = syscall.NewLazyDLL("shell32.dll")
	procShellExecute               = shell32.NewProc("ShellExecuteW")
)

var (
	hBitmap [8]winapi.HBITMAP
	hRgn    [8]winapi.HRGN
	hFont   winapi.HFONT
	white   = winapi.RGB(0xFF, 0xFF, 0xFF)
	black   = winapi.RGB(0x00, 0x00, 0x00)

	gopher *Gopher

	screenWidth int // screen width

	// sceneInfo have two directions. first is going right, second is left.
	// they have four scenesInfo, last one is gopher that have balloon.
	scenes [2][5]sceneInfo

	name   = flag.String("n", "", "name of gopher")
	slmode = flag.Bool("sl", false, "sl mode")
)

func updateWindowRegion(hWnd winapi.HWND) {
	s := gopher.CurrentSceneInfo()
	tmp := winapi.CreateRectRgn(0, 0, 0, 0)
	winapi.CombineRgn(tmp, s.hRgn, 0, winapi.RGN_COPY)
	winapi.SetWindowPos(hWnd, 0, int32(gopher.X()), int32(gopher.Y()-s.off), 0, 0,
		winapi.SWP_NOSIZE|winapi.SWP_NOZORDER|winapi.SWP_NOOWNERZORDER)
	procSetWindowRgn.Call(uintptr(hWnd), uintptr(tmp), uintptr(1))
	winapi.InvalidateRect(hWnd, nil, false)
}

func openBrowser(hWnd winapi.HWND, uri string) {
	if _, err := url.Parse(uri); err != nil {
		return
	}
	procShellExecute.Call(
		uintptr(hWnd),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("open"))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(uri))),
		0,
		0,
		uintptr(winapi.SW_SHOW))
}

func paintGopher(hWnd winapi.HWND) {
	var ps winapi.PAINTSTRUCT
	s := gopher.CurrentSceneInfo()
	hdc := winapi.BeginPaint(hWnd, &ps)
	hCompatDC := winapi.CreateCompatibleDC(hdc)
	winapi.SelectObject(hCompatDC, winapi.HGDIOBJ(s.hBitmap))
	winapi.BitBlt(hdc, 0, 0, int32(gopher.W()), int32(gopher.H()), hCompatDC, 0, 0, winapi.SRCCOPY)
	if gopher.Mode() == Waiting {
		pt, _ := syscall.UTF16PtrFromString(gopher.Content())
		winapi.SetTextColor(hdc, black)
		winapi.SetBkMode(hdc, winapi.TRANSPARENT)
		rc := winapi.RECT{13, 135, 190, 186}
		old := winapi.SelectObject(hdc, winapi.HGDIOBJ(hFont))
		procDrawText.Call(
			uintptr(hdc),
			uintptr(unsafe.Pointer(pt)), uintptr(len([]rune(gopher.Content()))),
			uintptr(unsafe.Pointer(&rc)),
			DT_LEFT|DT_VCENTER|DT_NOPREFIX|DT_EDITCONTROL|DT_WORDBREAK)
		winapi.SelectObject(hdc, old)
	}
	winapi.DeleteDC(hCompatDC)
	winapi.EndPaint(hWnd, &ps)
}

func handleMethod(hWnd winapi.HWND, method string) {
	switch method {
	case "terminate":
		gopher.SetMode(Stopping)
	case "jump":
		gopher.SetMode(HighJumping)
	case "message":
		gopher.SetMode(Waiting)
		winapi.SetWindowPos(hWnd, 0, 0, 0, 0, 0,
			winapi.SWP_NOMOVE|winapi.SWP_NOSIZE|winapi.SWP_NOOWNERZORDER)
		updateWindowRegion(hWnd)
	default:
		gopher.SetMode(Waiting)
		gopher.SetContent("What do you mean?")
		updateWindowRegion(hWnd)
	}
}

func randomAction() {
	if rand.Int()%10 == 0 {
		gopher.SetMode(Jumping)
	} else if rand.Int()%30 == 0 {
		gopher.Turn()
	}
}

func animateGopher(hWnd winapi.HWND) {
	switch gopher.Mode() {
	case Stopping:
		if gopher.Idle() <= 0 {
			winapi.DestroyWindow(hWnd)
		}
		gopher.NextScene()
		procSetLayeredWindowAttributes.Call(uintptr(hWnd), uintptr(white), uintptr(gopher.Time()*25), LWA_ALPHA)
		gopher.Motion()
	case Waiting:
		if gopher.Idle() <= 0 {
			gopher.SetMode(Walking)
		}
	case SL:
		gopher.NextScene()
		gopher.Motion()
	case Walking:
		if gopher.NextScene() == 0 {
			method, ok := gopher.Take()
			if ok {
				handleMethod(hWnd, method)
			} else {
				randomAction()
			}
		}
		gopher.Motion()
	default:
		gopher.Motion()
	}

	if gopher.Mode() == SL {
		if gopher.X() > screenWidth+gopher.W() {
			winapi.DestroyWindow(hWnd)
			return
		}
	} else {
		// turn over
		if (gopher.Dx() > 0 && gopher.X() > screenWidth) || (gopher.Dx() < 0 && gopher.X() < 0) {
			gopher.Turn()
		}
	}

	// redraw one
	if gopher.Mode() != Waiting {
		updateWindowRegion(hWnd)
	}
}

func clickGopher(hWnd winapi.HWND) {
	if winapi.GetKeyState(winapi.VK_SHIFT) < 0 {
		gopher.ClearMsg()
		gopher.PushMsg(gopherlib.Msg{Method: "terminate"})
		return
	}

	switch gopher.Mode() {
	case Waiting:
		if gopher.Link() != "" {
			openBrowser(hWnd, gopher.Link())
			gopher.WakeUp()
		}
	case Walking:
		gopher.SetMode(Jumping)
	}
}

func ipcGopher(hWnd winapi.HWND, cds *tagCOPYDATASTRUCT) {
	if gopher.Busy() {
		procReplyMessage.Call(2)
		return
	}
	b := make([]byte, cds.CbData)
	copy(b, (*(*[1 << 20]byte)(cds.LpData))[:])
	var msg gopherlib.Msg
	if err := json.Unmarshal(b, &msg); err != nil {
		procReplyMessage.Call(1)
		return
	}
	procReplyMessage.Call(0)

	switch msg.Method {
	case "terminate":
		gopher.ClearMsg()
		gopher.PushMsg(gopherlib.Msg{Method: "terminate"})
	case "clear":
		gopher.ClearMsg()
	default:
		gopher.PushMsg(msg)
	}
}

func showErrorMessage(hWnd winapi.HWND, msg string) {
	s, _ := syscall.UTF16PtrFromString(msg)
	t, _ := syscall.UTF16PtrFromString(gopherlib.WINDOW_CLASS)
	winapi.MessageBox(hWnd, s, t, winapi.MB_ICONWARNING|winapi.MB_OK)
}

func runGopher() int {
	var buf bytes.Buffer
	flag.CommandLine.SetOutput(&buf)
	flag.Usage = func() {
		fmt.Fprintf(&buf, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		showErrorMessage(0, buf.String())
		os.Exit(1)
	}
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	hInstance := winapi.GetModuleHandle(nil)

	if registerWindowClass(hInstance) == 0 {
		showErrorMessage(0, "registerWindowClass failed")
		return 1
	}

	if err := initializeInstance(hInstance, winapi.SW_SHOW); err != nil {
		showErrorMessage(0, err.Error())
		return 1
	}

	var msg winapi.MSG
	for winapi.GetMessage(&msg, 0, 0, 0) != 0 {
		winapi.TranslateMessage(&msg)
		winapi.DispatchMessage(&msg)
	}

	finalizeInstance(hInstance)

	return int(msg.WParam)
}

func registerWindowClass(hInstance winapi.HINSTANCE) winapi.ATOM {
	var wc winapi.WNDCLASSEX

	wc.CbSize = uint32(unsafe.Sizeof(winapi.WNDCLASSEX{}))
	wc.Style = 0
	wc.LpfnWndProc = syscall.NewCallback(wndProc)
	wc.CbClsExtra = 0
	wc.CbWndExtra = 0
	wc.HInstance = hInstance
	wc.HIcon = winapi.LoadIcon(hInstance, winapi.MAKEINTRESOURCE(132))
	wc.HCursor = winapi.LoadCursor(0, winapi.MAKEINTRESOURCE(winapi.IDC_HAND))
	wc.HbrBackground = winapi.HBRUSH(winapi.GetStockObject(winapi.WHITE_BRUSH))
	wc.LpszMenuName = nil
	wc.LpszClassName, _ = syscall.UTF16PtrFromString(gopherlib.WINDOW_CLASS)

	return winapi.RegisterClassEx(&wc)
}

func initializeInstance(hInstance winapi.HINSTANCE, nCmdShow int) error {
	var err error
	gopher, err = makeGopher()
	if err != nil {
		return err
	}

	hFont = winapi.CreateFont(
		15, 0, 0, 0, winapi.FW_NORMAL, 0, 0, 0,
		winapi.ANSI_CHARSET, winapi.OUT_DEVICE_PRECIS,
		winapi.CLIP_DEFAULT_PRECIS, winapi.DEFAULT_QUALITY,
		winapi.VARIABLE_PITCH|winapi.FF_ROMAN, nil)

	title := *name
	if title == "" {
		title = gopherlib.WINDOW_CLASS
	}
	pc, _ := syscall.UTF16PtrFromString(gopherlib.WINDOW_CLASS)
	pt, _ := syscall.UTF16PtrFromString(title)
	hWnd := winapi.CreateWindowEx(
		winapi.WS_EX_TOOLWINDOW|winapi.WS_EX_TOPMOST|winapi.WS_EX_NOACTIVATE|winapi.WS_EX_LAYERED,
		pc, pt, winapi.WS_POPUP,
		int32(gopher.X()),
		int32(gopher.X()),
		int32(gopher.W()),
		int32(gopher.H()),
		0, 0, hInstance, nil)
	if hWnd == 0 {
		return errors.New("CreateWindowEx failed")
	}

	updateWindowRegion(hWnd)

	procSetLayeredWindowAttributes.Call(uintptr(hWnd), uintptr(white), 255, LWA_ALPHA)
	winapi.ShowWindow(hWnd, int32(nCmdShow))
	winapi.SetTimer(hWnd, 1, 50, 0)
	return nil
}

func finalizeInstance(hInstance winapi.HINSTANCE) error {
	winapi.DeleteObject(winapi.HGDIOBJ(hFont))
	for _, h := range hBitmap {
		winapi.DeleteObject(winapi.HGDIOBJ(h))
	}
	for _, h := range hRgn {
		winapi.DeleteObject(winapi.HGDIOBJ(h))
	}
	return nil
}

func wndProc(hWnd winapi.HWND, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	switch msg {
	case winapi.WM_COPYDATA:
		ipcGopher(hWnd, (*tagCOPYDATASTRUCT)(unsafe.Pointer(lParam)))
	case winapi.WM_LBUTTONDOWN:
		clickGopher(hWnd)
	case winapi.WM_PAINT:
		paintGopher(hWnd)
	case winapi.WM_TIMER:
		animateGopher(hWnd)
	case winapi.WM_DESTROY:
		winapi.PostQuitMessage(0)
	default:
		return winapi.DefWindowProc(hWnd, msg, wParam, lParam)
	}
	return 0
}
