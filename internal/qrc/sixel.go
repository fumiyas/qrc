package qrc

import (
	"fmt"
	"io"

	"rsc.io/qr"
)

// PrintSixel renders the QR code as a Sixel image. Each module becomes
// a 6*scale by 6*scale pixel block. scale must be >= 1.
func PrintSixel(w io.Writer, code *qr.Code, inverse bool, scale int) {
	if scale < 1 {
		scale = 1
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
	line := "#" + white + "!" + fmt.Sprintf("%d", (size+2)*pxPerModule) + "~"

	// Top quiet zone: scale bands of all-white.
	for i := 0; i < scale; i++ {
		fmt.Fprint(w, line, "-")
	}
	for y := 0; y < size; y++ {
		// Build the band content for this module row once, then emit
		// the same band scale times to scale vertically.
		for b := 0; b < scale; b++ {
			fmt.Fprint(w, "#", white)
			color := white
			repeat := pxPerModule
			var current string
			for x := 0; x < size; x++ {
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
				repeat += pxPerModule
			}
			if color == white {
				fmt.Fprintf(w, "#%s!%d~", white, repeat+pxPerModule)
			} else {
				fmt.Fprintf(w, "#%s!%d~#%s!%d~", color, repeat, white, pxPerModule)
			}
			fmt.Fprint(w, "-")
		}
	}
	// Bottom quiet zone: scale bands. The very last band has no trailing
	// '-' because there is no following band.
	for i := 0; i < scale-1; i++ {
		fmt.Fprint(w, line, "-")
	}
	fmt.Fprint(w, line)
	fmt.Fprint(w, "\x1B\\")
}
