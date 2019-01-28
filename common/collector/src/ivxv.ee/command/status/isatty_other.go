// +build !linux

package status

func isatty(fd uintptr) bool {
	return false
}
