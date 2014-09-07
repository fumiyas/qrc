// +build !darwin,!windows,!freebsd

package tty

import (
	"syscall"
)

const (
	ioctlGetTermios = syscall.TCGETS
	ioctlSetTermios = syscall.TCSETS
)
