// +build linux

package status

import (
	"syscall"
	"unsafe"
)

// Copied from golang.org/x/crypto/ssh/terminal.IsTerminal.
func isatty(fd uintptr) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCGETS,
		uintptr(unsafe.Pointer(&termios)), 0, 0, 0) // nolint: gosec, OK use of unsafe.
	return err == 0
}
