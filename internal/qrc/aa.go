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

// PrintAA renders the QR code using ANSI background-color escape
// sequences and pairs of spaces. Each module is rendered as 2*scale
// horizontal spaces, and each module row is repeated scale times
// vertically. A quiet zone of border modules is added on all four
// sides. scale must be >= 1, border must be >= 0.
func PrintAA(wIn io.Writer, code *qr.Code, inverse bool, scale, border int) {
	if scale < 1 {
		scale = 1
	}
	if border < 0 {
		border = 0
	}
	// Buffering required for Windows (go-colorable) support
	w := bufio.NewWriterSize(wIn, 1024)

	reset := ansiReset
	black := ansiBlack
	white := ansiWhite
	if inverse {
		black, white = white, black
	}

	size := code.Size
	moduleSpaces := 2 * scale
	pad := fmt.Sprintf("%*s", border*moduleSpaces, "")
	margin := white + fmt.Sprintf("%*s", (size+2*border)*moduleSpaces, "") + reset + "\n"

	for i := 0; i < border*scale; i++ {
		fmt.Fprint(w, margin)
	}
	for y := 0; y < size; y++ {
		for r := 0; r < scale; r++ {
			fmt.Fprint(w, white, pad)
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
				fmt.Fprint(w, fmt.Sprintf("%*s", moduleSpaces, ""))
			}
			fmt.Fprint(w, white, pad, reset, "\n")
			w.Flush()
		}
	}
	for i := 0; i < border*scale; i++ {
		fmt.Fprint(w, margin)
	}
	w.Flush()
}
