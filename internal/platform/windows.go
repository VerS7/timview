//go:build windows
// +build windows

package platform

import (
	"golang.org/x/sys/windows"
)

func IsTerminal(fd int) bool {
	var mode uint32

	err := windows.GetConsoleMode(windows.Handle(fd), &mode)
	return err == nil
}

func GetSize(fd int) (int, int, error) {
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(fd), &info); err != nil {
		return 0, 0, err
	}
	return int(info.Window.Right - info.Window.Left + 1), int(info.Window.Bottom - info.Window.Top + 1), nil
}

func EnableColoredOutput(fd int) bool {
	var mode uint32
	handle := windows.Handle(fd)

	err := windows.GetConsoleMode(handle, &mode)
	if err != nil {
		return false
	}

	mode |= 0x0004

	err = windows.SetConsoleMode(handle, mode)
	return err == nil
}
