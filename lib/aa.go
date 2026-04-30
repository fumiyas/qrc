package qrc

import (
	"bufio"
	"fmt"
	"io"

	"github.com/qpliu/qrencode-go/qrencode"
)

// ANSI SGR sequences. The exact byte sequences are kept identical to
// the previous github.com/mgutz/ansi based implementation so that the
// rendered output is unchanged.
const (
	ansiReset = "\x1b[0m"
	ansiBlack = "\x1b[0;30;40m" // black foreground on black background
	ansiWhite = "\x1b[0;30;47m" // black foreground on white background
)

func PrintAA(w_in io.Writer, grid *qrencode.BitGrid, inverse bool) {
	// Buffering required for Windows (go-colorable) support
	w := bufio.NewWriterSize(w_in, 1024)

	reset := ansiReset
	black := ansiBlack
	white := ansiWhite
	if inverse {
		black, white = white, black
	}

	height := grid.Height()
	width := grid.Width()
	line := white + fmt.Sprintf("%*s", width*2+2, "") + reset + "\n"

	fmt.Fprint(w, line)
	for y := 0; y < height; y++ {
		fmt.Fprint(w, white, " ")
		color_prev := white
		for x := 0; x < width; x++ {
			if grid.Get(x, y) {
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
