// +build windows
package main

import (
	"github.com/AllenDang/w32"
)

var isHidden bool = false

// Hide the console window
func HideConsoleWindow() {
	w32.ShowWindow(w32.GetConsoleWindow(), w32.SW_HIDE)
	isHidden = true
}

// Show the console window
func ShowConsoleWindow() {
	w32.ShowWindow(w32.GetConsoleWindow(), w32.SW_SHOW)
	isHidden = false
}

// Toggle the visibility of the console window
func ToggleShowConsoleWindow() {
	if isHidden {
		ShowConsoleWindow()
	} else {
		HideConsoleWindow()
	}
}
