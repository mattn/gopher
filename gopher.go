package gopher

import (
	"encoding/json"
	"errors"
	"github.com/cwchiu/go-winapi"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	WINDOW_CLASS = "Gopher"
)

type tagCOPYDATASTRUCT struct {
	DwData uintptr
	CbData uint32
	LpData unsafe.Pointer
}

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procGetClassName  = user32.NewProc("GetClassNameW")
	ErrInvalidRequest = errors.New("Invalid request")
	ErrTooManyRequest = errors.New("Too many request")
)

type Msg struct {
	Method  string `json:"method"`
	Content string `json:"content"`
	Link    string `json:"link"`
}

type Gopher struct {
	wnd winapi.HWND
}

var cb = syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
	var w [256]uint16
	procGetClassName.Call(uintptr(h), uintptr(unsafe.Pointer(&w[0])), 255)
	if syscall.UTF16ToString(w[:]) == WINDOW_CLASS {
		gophers := (*[]*Gopher)(unsafe.Pointer(p))
		*gophers = append(*gophers, &Gopher{winapi.HWND(h)})
	}
	return 1
})

// Lookup find gophers available.
func Lookup() []*Gopher {
	gophers := []*Gopher{}
	winapi.EnumChildWindows(0, cb, uintptr(unsafe.Pointer(&gophers)))
	return gophers
}

// LookupByName find gophers by name.
func LookupByName(name string) []*Gopher {
	gophers := []*Gopher{}
	for _, t := range Lookup() {
		var buf [1024]uint16
		if winapi.GetWindowText(t.wnd, &buf[0], 1024) > 0 {
			if syscall.UTF16ToString(buf[:]) == name {
				gophers = append(gophers, &Gopher{winapi.HWND(t.wnd)})
			}
		}
	}
	return gophers
}

// Create gopher.
func Create(name string) error {
	return exec.Command("gopher", "-n", name).Start()
}

func sendMessage(hWnd winapi.HWND, m *Msg) error {
	b, err := json.Marshal(&m)
	if err != nil {
		return err
	}
	var cds tagCOPYDATASTRUCT
	cds.CbData = uint32(len(b))
	cds.LpData = unsafe.Pointer(&b[0])
	ret := winapi.SendMessage(hWnd, winapi.WM_COPYDATA, 0, uintptr(unsafe.Pointer(&cds)))
	switch ret {
	case 1:
		return ErrInvalidRequest
	case 2:
		return ErrTooManyRequest
	}
	return nil
}

// Terminate gopher. If the message can't be arrived to gopher, or message
// queue is full of capacity, return error.
func (g *Gopher) Terminate() error {
	return sendMessage(g.wnd, &Msg{Method: "terminate"})
}

// Message send text and link. Both can be empty string. If the message can't
// be arrived to gopher, or message queue is full of capacity, return error.
func (g *Gopher) Message(m string, l string) error {
	return sendMessage(g.wnd, &Msg{
		Method:  "message",
		Content: m,
		Link:    l,
	})
}

// Clear request to clear message queue. If the message can't be arrived to
// gopher, return error.
func (g *Gopher) Clear() error {
	return sendMessage(g.wnd, &Msg{
		Method: "clear",
	})
}

// Jump send request jumping to gopher. If the message can't be arrived to
// gopher, or message queue is full of capacity, return error.
func (g *Gopher) Jump() error {
	return sendMessage(g.wnd, &Msg{
		Method: "jump",
	})
}

func (g *Gopher) Name() string {
	var buf [1024]uint16
	if winapi.GetWindowText(g.wnd, &buf[0], 1024) > 0 {
		return syscall.UTF16ToString(buf[:])
	}
	return ""
}

func (g *Gopher) Hwnd() winapi.HWND {
	return g.wnd
}
