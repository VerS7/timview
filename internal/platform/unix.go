//go:build !windows
// +build !windows

package platform

import (
	"golang.org/x/sys/unix"
)

const (
	ioctlReadTermios = unix.TCGETS
	ioctlGetWinSize  = unix.TIOCGWINSZ
)

func IsTerminal(fd int) bool {
	_, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	return err == nil
}

func GetSize(fd int) (int, int, error) {
	ws, err := unix.IoctlGetWinsize(fd, ioctlGetWinSize)
	if err != nil {
		return 0, 0, err
	}
	return int(ws.Col), int(ws.Row), nil
}

func EnableColoredOutput(fd int) bool {
	// TODO: check for coloring in UNIX
	return true
}
