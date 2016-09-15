// +build darwin

package xattr

import (
	"syscall"
	"unsafe"
)

// int removexattr(const char *path, const char *name, int options);
func removexattr(path string, name string) (e error) {
	_, _, e1 := syscall.Syscall(
		syscall.SYS_REMOVEXATTR,
		uintptr(unsafe.Pointer(syscall.StringBytePtr(path))),
		uintptr(unsafe.Pointer(syscall.StringBytePtr(name))),
		0)

	if e1 != noError {
		e = e1
	}

	return
}

// Remove the extended attribute.
func Remove(path, attr string) error {
	if err := removexattr(path, attr); err != nil {
		return &Error{"removexattr", path, attr, err}
	}

	return nil
}
