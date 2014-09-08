// +build darwin freebsd netbsd openbsd

package tty

import (
	"syscall"
)

const (
	ioctlGetTermios = syscall.TIOCGETA
	ioctlSetTermios = syscall.TIOCSETA
)
