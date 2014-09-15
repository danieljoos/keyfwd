// +build windows
package main

import (
	"github.com/AllenDang/w32"
	"syscall"
	"unsafe"
)

type KeyboardCapture struct {
	keyboardHook  w32.HHOOK
	forwardedKeys []int

	KeyPressed chan int
}

// Create a new KeyboardCapture object.
// The object will only 'capture' the keys, specified in the given
// integer array.
// Specify keys by using the VK_* constants of Windows:
// http://msdn.microsoft.com/en-us/library/windows/desktop/dd375731(v=vs.85).aspx
func NewKeyboardCapture(forwardedKeys []int) *KeyboardCapture {
	ret := new(KeyboardCapture)
	ret.forwardedKeys = forwardedKeys
	ret.KeyPressed = make(chan int)
	return ret
}

// Creates a low-level keyboard hook using the SetWindowsHookEx function:
// http://msdn.microsoft.com/en-us/library/windows/desktop/ms644990(v=vs.85).aspx
//
// Each intercepted key, which was included in the 'forwardedKeys' configuration
// variable (see NewKeyboardCapture), will be pushed to the 'KeyPressed' channel field.
// Returns an error in case the initialization of the hook failed.
// Calls to this function will block until KeyboardCapture.Stop() was called or the
// WM_QUIT message was sent to the current process.
func (t *KeyboardCapture) SyncReceive() error {
	isValidKey := func(key w32.DWORD) bool {
		for _, e := range t.forwardedKeys {
			if e == int(key) {
				return true
			}
		}
		return false
	}
	t.KeyPressed = make(chan int)
	t.keyboardHook = w32.SetWindowsHookEx(w32.WH_KEYBOARD_LL,
		(w32.HOOKPROC)(func(code int, wparam w32.WPARAM, lparam w32.LPARAM) w32.LRESULT {
			if code >= 0 && wparam == w32.WM_KEYDOWN {
				kbdstruct := (*w32.KBDLLHOOKSTRUCT)(unsafe.Pointer(lparam))
				if isValidKey(kbdstruct.VkCode) {
					select {
					case t.KeyPressed <- int(kbdstruct.VkCode):
					default:
					}
				}
			}
			return w32.CallNextHookEx(t.keyboardHook, code, wparam, lparam)
		}), 0, 0)
	if t.keyboardHook == 0 {
		return syscall.GetLastError()
	}
	var msg w32.MSG
	for w32.GetMessage(&msg, 0, 0, 0) != 0 {
	}
	w32.UnhookWindowsHookEx(t.keyboardHook)
	t.keyboardHook = 0

	return nil
}

// Stops the key interception by sending the quit message (WM_QUIT) to the current
// process.
func (t *KeyboardCapture) Stop() {
	w32.PostQuitMessage(0)
}
