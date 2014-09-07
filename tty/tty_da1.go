package tty

import (
	"bytes"
	"fmt"
	"os"
)

const (
	DA1_132_COLUMNS                         = 1
	DA1_PRINTER_PORT                        = 2
	DA1_SIXEL                               = 4
	DA1_SELECTIVE_ERASE                     = 6
	DA1_DRCS                                = 7
	DA1_SOFT_CHARACTER_SET                  = 7
	DA1_UDKS                                = 8
	DA1_USER_DEFINED_KEYS                   = 8
	DA1_NRCS                                = 9
	DA1_NATIONAL_REPLACEMENT_CHARACTER_SETS = 9
	DA1_SCS                                 = 12
	DA1_YUGOSLAVIAN                         = 12
	DA1_TECHNICAL_CHARACTER_SET             = 15
	DA1_WINDOWING_CAPABILITY                = 18
	DA1_HORIZONTAL_SCROLLING                = 21
	DA1_GREEK                               = 23
	DA1_TURKISH                             = 24
	DA1_ISO_LATIN2_CHARACTER_SET            = 42
	DA1_PCTERM                              = 44
	DA1_SOFT_KEY_MAP                        = 45
	DA1_ASCII_EMULATION                     = 46
	DA1_MAX                                 = 64
)

type DeviceAttributes1 [DA1_MAX]bool

func GetDeviceAttributes1(file *os.File) (DeviceAttributes1, error) {
	var err error
	var termios_save Termios
	var da1 DeviceAttributes1

	termios_save, err = MakeRaw(file)
	if err != nil {
		return da1, err
	}
	defer SetTermios(file, termios_save)

	file.WriteString("\x1B[c")

	buf := make([]byte, 3)
	_, err = file.Read(buf)
	if err != nil {
		return da1, fmt.Errorf("cannot read DA1: %v", err)
	}
	if bytes.Compare(buf, []byte("\x1b[?")) != 0 {
		return da1, fmt.Errorf("invalid DA1 response")
	}

	var attr byte
LOOP:
	for {
		_, err = file.Read(buf[0:1])
		if err != nil {
			return da1, fmt.Errorf("cannot read DA1: %v", err)
		}
		switch {
		case buf[0] >= '0' && buf[0] <= '9':
			attr *= 10
			attr += buf[0] - '0'
		case buf[0] == ';' || buf[0] == 'c':
			if attr <= DA1_MAX {
				da1[attr] = true
			}
			if buf[0] == 'c' {
				break LOOP
			}
			attr = 0
		}
	}

	return da1, nil
}
