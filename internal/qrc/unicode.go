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
// A 1-module quiet zone is added around the code; when the resulting
// height is odd, an extra blank module row is appended at the bottom so
// that the output remains an integer number of half-block rows.
func PrintUnicode(wIn io.Writer, code *qr.Code, inverse bool) {
	w := bufio.NewWriterSize(wIn, 1024)
	size := code.Size

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

	// One quiet module on each side. The total module height (size + 2)
	// is rounded up to the next even number so we always emit complete
	// half-block rows.
	totalRows := (size + 3) / 2
	for r := 0; r < totalRows; r++ {
		yTop := r*2 - 1
		yBot := yTop + 1
		for x := -1; x <= size; x++ {
			fmt.Fprintf(w, "%c", cell(isDark(x, yTop), isDark(x, yBot)))
		}
		fmt.Fprintln(w)
	}
	w.Flush()
}
