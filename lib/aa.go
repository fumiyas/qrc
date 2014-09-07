package qrc

import (
	"code.google.com/p/rsc/qr"
	"fmt"
	"github.com/mgutz/ansi"
)

func PrintAA(code *qr.Code, inverse bool) {
	reset := ansi.ColorCode("reset")
	black := ansi.ColorCode(":black")
	white := ansi.ColorCode(":white")
	if inverse {
		black, white = white, black
	}

	line := white + fmt.Sprintf("%*s", code.Size*2+2, "") + reset + "\n"

	fmt.Print(line)
	for y := 0; y < code.Size; y++ {
		fmt.Print(white, " ")
		color_prev := white
		for x := 0; x < code.Size; x++ {
			if code.Black(x, y) {
				if color_prev != black {
					fmt.Print(black)
					color_prev = black
				}
			} else {
				if color_prev != white {
					fmt.Print(white)
					color_prev = white
				}
			}
			fmt.Print("  ")
		}
		fmt.Print(white, " ", reset, "\n")
	}
	fmt.Print(line)
}
