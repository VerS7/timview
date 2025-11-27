//go:build windows
// +build windows

package platform

import (
	"golang.org/x/sys/windows"
)

type State struct {
	mode uint32
}

func MakeRaw(fd int) (*State, error) {
	var st uint32
	if err := windows.GetConsoleMode(windows.Handle(fd), &st); err != nil {
		return nil, err
	}
	raw := st &^ (windows.ENABLE_ECHO_INPUT | windows.ENABLE_PROCESSED_INPUT | windows.ENABLE_LINE_INPUT)
	raw |= windows.ENABLE_VIRTUAL_TERMINAL_INPUT
	if err := windows.SetConsoleMode(windows.Handle(fd), raw); err != nil {
		return nil, err
	}
	return &State{mode: st}, nil
}

func GetState(fd int) (*State, error) {
	var st uint32
	if err := windows.GetConsoleMode(windows.Handle(fd), &st); err != nil {
		return nil, err
	}
	return &State{mode: st}, nil
}

func Restore(fd int, state *State) error {
	return windows.SetConsoleMode(windows.Handle(fd), state.mode)
}

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
