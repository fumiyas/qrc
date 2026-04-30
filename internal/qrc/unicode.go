package qrc

import (
	"bufio"
	"fmt"
	"io"

	"rsc.io/qr"
)

// Unicode half-block characters used to pack two QR modules into one
// character cell. The upper half corresponds to one module and the
// lower half to the module directly below it.
const (
	uBlockNone  = ' '      // both halves light
	uBlockUpper = '\u2580' // ▀ upper dark, lower light
	uBlockLower = '\u2584' // ▄ upper light, lower dark
	uBlockFull  = '\u2588' // █ both halves dark
)

// PrintUnicode renders the QR code using Unicode half-block characters,
// packing two vertically-adjacent modules into a single character cell.
//
// Each cell is repeated horizontally scale times and vertically scale
// times. scale must be >= 1, border must be >= 0.
//
// A quiet zone of border modules is added on all four sides. Because
// each character row spans two modules, the bottom quiet zone is
// padded by one extra module when (size + 2*border) is odd so that the
// output remains an integer number of half-block rows.
func PrintUnicode(wIn io.Writer, code *qr.Code, inverse bool, scale, border int) {
	if scale < 1 {
		scale = 1
	}
	if border < 0 {
		border = 0
	}
	w := bufio.NewWriterSize(wIn, 1024)
	size := code.Size

	// Unicode half-block characters paint with the terminal's foreground
	// color, which on the typical dark-themed terminal appears bright.
	// To keep the visual appearance consistent with the ansi/sixel
	// formats (dark QR modules look dark), we treat block characters as
	// LIGHT modules and spaces as DARK modules — i.e. internally invert
	// the requested mode. The user-facing --invert flag still flips on
	// top of that, so -i restores the alternate look.
	inverse = !inverse

	isDark := func(x, y int) bool {
		if x < 0 || x >= size || y < 0 || y >= size {
			return inverse
		}
		return code.Black(x, y) != inverse
	}

	cell := func(top, bot bool) rune {
		switch {
		case top && bot:
			return uBlockFull
		case top:
			return uBlockUpper
		case bot:
			return uBlockLower
		default:
			return uBlockNone
		}
	}

	totalRows := (size + 2*border + 1) / 2
	for r := range totalRows {
		yTop := r*2 - border
		yBot := yTop + 1
		for range scale {
			for x := -border; x < size+border; x++ {
				ch := cell(isDark(x, yTop), isDark(x, yBot))
				for range scale {
					fmt.Fprintf(w, "%c", ch)
				}
			}
			fmt.Fprintln(w)
		}
	}
	w.Flush()
}
