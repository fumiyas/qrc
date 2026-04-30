package qrc

import (
	"bufio"
	"fmt"
	"io"

	"rsc.io/qr"
)

// ANSI SGR sequences. The exact byte sequences are kept identical to
// the previous github.com/mgutz/ansi based implementation so that the
// rendered output is unchanged.
const (
	ansiReset = "\x1b[0m"
	ansiBlack = "\x1b[0;30;40m" // black foreground on black background
	ansiWhite = "\x1b[0;30;47m" // black foreground on white background
)

func PrintAA(wIn io.Writer, code *qr.Code, inverse bool) {
	// Buffering required for Windows (go-colorable) support
	w := bufio.NewWriterSize(wIn, 1024)

	reset := ansiReset
	black := ansiBlack
	white := ansiWhite
	if inverse {
		black, white = white, black
	}

	size := code.Size
	line := white + fmt.Sprintf("%*s", size*2+2, "") + reset + "\n"

	fmt.Fprint(w, line)
	for y := 0; y < size; y++ {
		fmt.Fprint(w, white, " ")
		colorPrev := white
		for x := 0; x < size; x++ {
			if code.Black(x, y) {
				if colorPrev != black {
					fmt.Fprint(w, black)
					colorPrev = black
				}
			} else {
				if colorPrev != white {
					fmt.Fprint(w, white)
					colorPrev = white
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
