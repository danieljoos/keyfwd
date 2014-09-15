// +build windows
package main

import (
	"github.com/AllenDang/w32"
)

type KeyboardEmitter struct {
	input [1]w32.INPUT
}

func NewKeyboardEmitter() *KeyboardEmitter {
	ret := new(KeyboardEmitter)
	ret.input[0].Type = w32.INPUT_KEYBOARD
	ret.input[0].Ki.WScan = 0
	ret.input[0].Ki.Time = 0
	ret.input[0].Ki.DwExtraInfo = 0
	ret.input[0].Ki.WVk = 0
	ret.input[0].Ki.DwFlags = 0
	return ret
}

func (t *KeyboardEmitter) SendKey(key int) {
	t.input[0].Ki.WVk = uint16(key)
	t.input[0].Ki.DwFlags = 0
	w32.SendInput(t.input[:])
	t.input[0].Ki.DwFlags = 2
	w32.SendInput(t.input[:])
}
