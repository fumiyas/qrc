// +build !darwin,!windows,!freebsd,!netbsd,!openbsd

package tty

import (
	"syscall"
)

const (
	ioctlGetTermios = syscall.TCGETS
	ioctlSetTermios = syscall.TCSETS
)
