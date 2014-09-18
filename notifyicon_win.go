// +build windows
package main

import (
	"github.com/AllenDang/w32"
	"log"
	"syscall"
	"unsafe"
)

const (
	niUID             uint32 = 100
	niCallbackMessage uint32 = 101
)

var (
	niIconEntries []_ICONDIRENTRY = getIconEntries(NotifyIconData)
	IconClient    w32.HICON       = getIconHandle(NotifyIconData, niIconEntries[0])
	IconServer    w32.HICON       = getIconHandle(NotifyIconData, niIconEntries[1])
)

// Type used for tracking interaction with the notify icon
type NotifyIconButton int

const (
	LeftMouseButton NotifyIconButton = iota
	RightMouseButton
)

type _NotifyIcon struct {
	nid     _NOTIFYICONDATA
	hwnd    w32.HWND
	tooltip string
	onClick chan NotifyIconButton
}

// Create a new notify icon with the given tooltip and icon handle
// The tooltip will be shown on mouse-over of the notify icon.
// Returns the notify icon object or a possible error.
func NewNotifyIcon(tooltip string, iconHandle w32.HICON) (*_NotifyIcon, error) {
	ret := new(_NotifyIcon)
	ret.tooltip = tooltip
	ret.nid.HIcon = iconHandle
	ret.onClick = make(chan NotifyIconButton)
	err := ret.createCallbackWindow()
	return ret, err
}

// Adds the notification icon and starts handling the mouse
// interaction with the icon.
// Calls to this function will block until the Stop() function
// will be called from another goroutine/thread.
// Returns an error, if adding the notify icons fails.
func (t *_NotifyIcon) Start() (err error) {
	log.Println("Creating notification icon")

	tooltipUtf16, _ := syscall.UTF16FromString(t.tooltip)
	t.nid.CbSize = w32.DWORD(unsafe.Sizeof(&t.nid))
	t.nid.HWnd = t.hwnd
	t.nid.UFlags = _NIF_MESSAGE | _NIF_ICON | _NIF_TIP
	t.nid.UID = niUID
	t.nid.UCallbackMessage = niCallbackMessage
	copy(t.nid.SzTip[:], tooltipUtf16)
	err = shellNotifyIcon(_NIM_ADD, &t.nid)
	if err != nil {
		return
	}

	var msg w32.MSG
	for w32.GetMessage(&msg, t.hwnd, uint32(0), uint32(0)) != 0 {
		w32.TranslateMessage(&msg)
		w32.DispatchMessage(&msg)
	}
	return nil
}

// Stops handling the interaction with the notify icon and
// removes the icon.
// The WM_QUIT message will be sent to all waiting GetMessage loops inside this process.
func (t *_NotifyIcon) Stop() {
	log.Println("Removig notification icon")
	shellNotifyIcon(_NIM_DELETE, &t.nid)
	w32.DestroyIcon(t.nid.HIcon)
	w32.PostQuitMessage(0)
}

// Returns a channel object, which will be filled on left- or right click on the
// notify icon.
func (t *_NotifyIcon) OnClick() chan NotifyIconButton {
	return t.onClick
}

// WndProc of the notify icon
func (t *_NotifyIcon) iconCallback(hwnd w32.HWND, msg uint32, wparam w32.WPARAM, lparam w32.LPARAM) w32.LRESULT {
	if msg == niCallbackMessage {
		switch lparam {
		case w32.WM_LBUTTONUP:
			select {
			case t.onClick <- LeftMouseButton:
			default:
			}
		case w32.WM_RBUTTONUP:
			select {
			case t.onClick <- RightMouseButton:
			default:
			}
		}
	}
	return w32.LRESULT(w32.DefWindowProc(hwnd, msg, uintptr(wparam), uintptr(lparam)))
}

// Creates a hidden window, required for capturing the mouse interaction with the
// notify icon. Windows will call the WndProc of this window, whenever something happens
// on the notify icon (e.g. mouse click).
func (t *_NotifyIcon) createCallbackWindow() (err error) {
	className := "KeyFwdWindowClass"

	err = registerWindowClass(className, t.iconCallback)
	if err != nil {
		return
	}

	classNamePtr, _ := syscall.UTF16PtrFromString(className)
	t.hwnd = w32.CreateWindowEx(
		uint(w32.WS_EX_LEFT|w32.WS_EX_LTRREADING|w32.WS_EX_WINDOWEDGE),
		classNamePtr,
		nil,
		uint(w32.WS_OVERLAPPED|w32.WS_MINIMIZEBOX|w32.WS_SYSMENU|w32.WS_CLIPSIBLINGS|w32.WS_CAPTION),
		w32.CW_USEDEFAULT,
		w32.CW_USEDEFAULT,
		10,
		10,
		w32.HWND_DESKTOP,
		w32.HMENU(0),
		w32.GetModuleHandle(""),
		unsafe.Pointer(uintptr(0)),
	)
	if t.hwnd == 0 {
		return syscall.GetLastError()
	}

	return nil
}

// Register a new window class for the current process. A window with the specified
// class name will use the given WndProc callback function.
func registerWindowClass(className string, callback _WindowProc) error {
	classNamePtr, _ := syscall.UTF16PtrFromString(className)
	var winClass w32.WNDCLASSEX
	winClass.Size = uint32(unsafe.Sizeof(winClass))
	winClass.Instance = w32.GetModuleHandle("")
	winClass.ClassName = classNamePtr
	winClass.WndProc = syscall.NewCallback(callback)

	if w32.RegisterClassEx(&winClass) == 0 {
		return syscall.GetLastError()
	}
	return nil
}

// Cast ICO file data into an ICONDIRENTRY array
func getIconEntries(icoData []byte) []_ICONDIRENTRY {
	ret := make([]_ICONDIRENTRY, icoData[4])
	for i := 0; i < len(ret); i++ {
		ret[i] = *((*_ICONDIRENTRY)(unsafe.Pointer(&icoData[6+(i*0x10)])))
	}
	return ret
}

// Fetch an icon handle using a ICONDIRENTRY struct
func getIconHandle(icoData []byte, dirEntry _ICONDIRENTRY) w32.HICON {
	ret, _ := createIconFromResourceEx(
		&icoData[dirEntry.dwImageOffset],
		w32.DWORD(0),
		true,
		w32.DWORD(0x30000),
		int32(dirEntry.bWidth),
		int32(dirEntry.bHeight),
		_LR_DEFAULT_COLOR)
	return w32.HICON(ret)
}

// Win32 constants:

const (
	_NIM_ADD    w32.DWORD = 0x00
	_NIM_DELETE w32.DWORD = 0x02

	_NIF_MESSAGE uint32 = 0x01
	_NIF_ICON    uint32 = 0x02
	_NIF_TIP     uint32 = 0x04

	_LR_DEFAULT_COLOR uint32 = 0x00
)

// Win32 data structures:

type _NOTIFYICONDATA struct {
	CbSize           w32.DWORD
	HWnd             w32.HWND
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            w32.HICON
	SzTip            [64]uint16
	DwState          w32.DWORD
	DwStateMask      w32.DWORD
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      w32.DWORD
	GuidItem         w32.GUID
}

type _WindowProc func(w32.HWND, uint32, w32.WPARAM, w32.LPARAM) w32.LRESULT

type _ICONDIRENTRY struct {
	bWidth        byte
	bHeight       byte
	bColorCount   byte
	bReserved     byte
	wPlanes       uint16
	wBitCount     uint16
	dwBytesInRes  uint32
	dwImageOffset uint32
}

// Win32 functions:

var (
	modShell32                   = syscall.NewLazyDLL("shell32.dll")
	modUser32                    = syscall.NewLazyDLL("user32.dll")
	procShell_NotifyIcon         = modShell32.NewProc("Shell_NotifyIconW")
	procCreateIconFromResourceEx = modUser32.NewProc("CreateIconFromResourceEx")
)

func shellNotifyIcon(message w32.DWORD, nid *_NOTIFYICONDATA) error {
	ret, _, _ := procShell_NotifyIcon.Call(
		uintptr(message),
		uintptr(unsafe.Pointer(nid)),
	)
	if ret == w32.FALSE {
		return syscall.GetLastError()
	}
	return nil
}

func createIconFromResourceEx(pbIconBits *byte, cbIconBits w32.DWORD, fIcon bool, dwVersion w32.DWORD,
	cxDesired, cyDesired int32, flags uint32) (w32.HICON, error) {
	ret, _, _ := procCreateIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(pbIconBits)),
		uintptr(cbIconBits),
		uintptr(w32.BoolToBOOL(fIcon)),
		uintptr(dwVersion),
		uintptr(cxDesired),
		uintptr(cyDesired),
		uintptr(flags),
	)
	if ret == 0 {
		return w32.HICON(0), syscall.GetLastError()
	}
	return w32.HICON(ret), nil
}
