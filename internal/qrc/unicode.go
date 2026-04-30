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
// times. scale must be >= 1.
//
// A 1-module quiet zone is added around the code; when the resulting
// height is odd, an extra blank module row is appended at the bottom so
// that the output remains an integer number of half-block rows.
func PrintUnicode(wIn io.Writer, code *qr.Code, inverse bool, scale int) {
	if scale < 1 {
		scale = 1
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

	totalRows := (size + 3) / 2
	for r := 0; r < totalRows; r++ {
		yTop := r*2 - 1
		yBot := yTop + 1
		for vr := 0; vr < scale; vr++ {
			for x := -1; x <= size; x++ {
				ch := cell(isDark(x, yTop), isDark(x, yBot))
				for hr := 0; hr < scale; hr++ {
					fmt.Fprintf(w, "%c", ch)
				}
			}
			fmt.Fprintln(w)
		}
	}
	w.Flush()
}
