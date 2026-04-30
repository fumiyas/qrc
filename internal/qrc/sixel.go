package qrc

import (
	"github.com/qpliu/qrencode-go/qrencode"
	"fmt"
	"io"
)

func PrintSixel(w io.Writer, grid *qrencode.BitGrid, inverse bool) {
	black := "0"
	white := "1"

	fmt.Fprint(w,
		"\x1BPq\"1;1",
		"#", black, ";2;0;0;0",
		"#", white, ";2;100;100;100",
	)

	if inverse {
		black, white = white, black
	}

	height := grid.Height()
	width := grid.Width()
	line := "#" + white + "!" + fmt.Sprintf("%d", (width+2)*6) + "~"

	fmt.Fprint(w, line, "-")
	for y := 0; y < height; y++ {
		fmt.Fprint(w, "#", white)
		color := white
		repeat := 6
		var current string
		for x := 0; x < width; x++ {
			if grid.Get(x, y) {
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
