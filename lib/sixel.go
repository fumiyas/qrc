package qrc

import (
	"code.google.com/p/rsc/qr"
	"fmt"
)

func PrintSixel(code *qr.Code, inverse bool) {
	black := "0"
	white := "1"
	if inverse {
		black, white = white, black
	}

	line := "#" + white + "!" + fmt.Sprintf("%d", (code.Size+2)*6) + "~"

	fmt.Print(
		"\x1BPq",
		"#", black, ";2;0;0;0",
		"#", white, ";2;100;100;100",
		line,
		"-",
	)

	for x := 0; x < code.Size; x++ {
		fmt.Print("#", white)
		color := white
		run := 6
		var current string
		for y := 0; y < code.Size; y++ {
			if code.Black(x, y) {
				current = black
			} else {
				current = white
			}
			if current != color {
				fmt.Print("#", color, "!", run, "~")
				color = current
				run = 0
			}
			run += 6
		}
		if color == white {
			fmt.Printf("#%s!%d~", white, run+6)
		} else {
			fmt.Printf("#%s!%d~#%s!6~", color, run, white)
		}
		fmt.Print("-")
	}
	fmt.Print(line)
	fmt.Print("\x1B\\")
}
