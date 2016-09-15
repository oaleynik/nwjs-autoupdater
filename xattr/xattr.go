package xattr

import "syscall"

var noError = syscall.Errno(0)

// Error records an error and the operation, file path and attribute that caused it.
type Error struct {
	Op   string
	Path string
	Attr string
	Err  error
}

func (e *Error) Error() string {
	return e.Op + " " + e.Path + " " + e.Attr + ": " + e.Err.Error()
}
