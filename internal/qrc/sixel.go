package qrc

import (
	"fmt"
	"io"

	"rsc.io/qr"
)

// PrintSixel renders the QR code as a Sixel image. Each module becomes
// a 6*scale by 6*scale pixel block. A quiet zone of border modules is
// added on all four sides. scale must be >= 1, border must be >= 0.
func PrintSixel(w io.Writer, code *qr.Code, inverse bool, scale, border int) {
	if scale < 1 {
		scale = 1
	}
	if border < 0 {
		border = 0
	}
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

	size := code.Size
	pxPerModule := 6 * scale
	marginPx := border * pxPerModule
	line := "#" + white + "!" + fmt.Sprintf("%d", (size+2*border)*pxPerModule) + "~"

	// Top quiet zone: border*scale all-white bands.
	for i := 0; i < border*scale; i++ {
		fmt.Fprint(w, line, "-")
	}
	for y := 0; y < size; y++ {
		// Build the band content for this module row once, then emit
		// the same band scale times to scale vertically.
		for b := 0; b < scale; b++ {
			fmt.Fprint(w, "#", white)
			color := white
			repeat := marginPx
			var current string
			for x := 0; x < size; x++ {
				if code.Black(x, y) {
					current = black
				} else {
					current = white
				}
				if current != color {
					if repeat > 0 {
						fmt.Fprint(w, "#", color, "!", repeat, "~")
					}
					color = current
					repeat = 0
				}
				repeat += pxPerModule
			}
			if marginPx == 0 {
				if repeat > 0 {
					fmt.Fprintf(w, "#%s!%d~", color, repeat)
				}
			} else if color == white {
				fmt.Fprintf(w, "#%s!%d~", white, repeat+marginPx)
			} else {
				fmt.Fprintf(w, "#%s!%d~#%s!%d~", color, repeat, white, marginPx)
			}
			// Trailing band separator unless this is the last band and
			// there is no bottom quiet zone.
			lastBand := y == size-1 && b == scale-1 && border == 0
			if !lastBand {
				fmt.Fprint(w, "-")
			}
		}
	}
	// Bottom quiet zone: border*scale bands. The very last band has no
	// trailing '-' because there is no following band.
	if border > 0 {
		for i := 0; i < border*scale-1; i++ {
			fmt.Fprint(w, line, "-")
		}
		fmt.Fprint(w, line)
	}
	fmt.Fprint(w, "\x1B\\")
}
