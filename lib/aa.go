package qrc

import (
	"bufio"
	"code.google.com/p/rsc/qr"
	"fmt"
	"github.com/mgutz/ansi"
	"io"
)

func PrintAA(w_in io.Writer, code *qr.Code, inverse bool) {
	// Buffering required for Windows (go-colorable) support
	w := bufio.NewWriterSize(w_in, 1024)

	reset := ansi.ColorCode("reset")
	black := ansi.ColorCode(":black")
	white := ansi.ColorCode(":white")
	if inverse {
		black, white = white, black
	}

	line := white + fmt.Sprintf("%*s", code.Size*2+2, "") + reset + "\n"

	fmt.Fprint(w, line)
	for y := 0; y < code.Size; y++ {
		fmt.Fprint(w, white, " ")
		color_prev := white
		for x := 0; x < code.Size; x++ {
			if code.Black(x, y) {
				if color_prev != black {
					fmt.Fprint(w, black)
					color_prev = black
				}
			} else {
				if color_prev != white {
					fmt.Fprint(w, white)
					color_prev = white
				}
			}
			fmt.Fprint(w, "  ")
		}
		fmt.Fprint(w, white, " ", reset, "\n")
		w.Flush()
	}
	fmt.Fprint(w, line)
	w.Flush()
}
