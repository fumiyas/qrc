// +build darwin freebsd

package tty

import (
	"syscall"
)

const (
	ioctlGetTermios = syscall.TIOCGETA
	ioctlSetTermios = syscall.TIOCSETA
)
