package qrc

import (
	"code.google.com/p/rsc/qr"
	"fmt"
	"io"
)

func PrintSixel(w io.Writer, code *qr.Code, inverse bool) {
	black := "0"
	white := "1"

	fmt.Fprint(w,
		"\x1BPq",
		"#", black, ";2;0;0;0",
		"#", white, ";2;100;100;100",
	)

	if inverse {
		black, white = white, black
	}

	line := "#" + white + "!" + fmt.Sprintf("%d", (code.Size+2)*6) + "~"
	fmt.Fprint(w, line, "-")

	for y := 0; y < code.Size; y++ {
		fmt.Fprint(w, "#", white)
		color := white
		repeat := 6
		var current string
		for x := 0; x < code.Size; x++ {
			if code.Black(x, y) {
				current = black
			} else {
				current = white
			}
			if current != color {
				fmt.Fprint(w, "#", color, "!", repeat, "~")
				color = current
				repeat = 0
			}
			repeat += 6
		}
		if color == white {
			fmt.Fprintf(w, "#%s!%d~", white, repeat+6)
		} else {
			fmt.Fprintf(w, "#%s!%d~#%s!6~", color, repeat, white)
		}
		fmt.Fprint(w, "-")
	}
	fmt.Fprint(w, line)
	fmt.Fprint(w, "\x1B\\")
}
