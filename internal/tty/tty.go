// +build !windows

package tty

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

type Termios syscall.Termios

func GetTermios(file *os.File) (Termios, error) {
	var termios Termios

	r, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		file.Fd(),
		uintptr(ioctlGetTermios),
		uintptr(unsafe.Pointer(&termios)))

	if int(r) == -1 {
		return termios, fmt.Errorf("ioctl failed: %s", syscall.Errno(errno).Error())
	}

	return termios, nil
}

func SetTermios(file *os.File, termios Termios) error {
	r, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		file.Fd(),
		uintptr(ioctlSetTermios),
		uintptr(unsafe.Pointer(&termios)))

	if int(r) == -1 {
		return fmt.Errorf("ioctl failed: %s", syscall.Errno(errno).Error())
	}

	return nil
}

func MakeRaw(file *os.File) (Termios, error) {
	termios, err := GetTermios(file)
	if err != nil {
		return termios, err
	}

	termios_save := termios

	termios.Iflag &^= (syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON)
	termios.Oflag &^= syscall.OPOST
	termios.Lflag &^= (syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN)
	termios.Cflag &^= (syscall.CSIZE | syscall.PARENB)
	termios.Cflag |= syscall.CS8

	err = SetTermios(file, termios)

	return termios_save, err
}

// IsTty checks if the given fd is a tty
func IsTty(file *os.File) bool {
	_, err := GetTermios(file)

	return (err == nil)
}
