package main

import (
	"os"
	"fmt"
	"bufio"
	"github.com/jessevdk/go-flags"
	"code.google.com/p/rsc/qr"
	"github.com/mgutz/ansi"
)

type cmdOptions struct {
	Help		bool	`short:"h" long:"help" description:"show this help message and exit"`
	Inverse		bool	`short:"i" long:"invert" description:"invert black and white"`
}

func showHelp() {
	const v = `Usage: qrc [OPTIONS] [TEXT]
`

	os.Stderr.Write([]byte(v))
}

func printAA(code *qr.Code, inverse bool) {
	reset := ansi.ColorCode("reset")
	black := ansi.ColorCode(":black")
	white := ansi.ColorCode(":white")

	if inverse {
		black, white = white, black
	}

	line := white + fmt.Sprintf("%*s", code.Size * 2 + 2, "") + reset + "\n"

	fmt.Print(line);
	for y := 0; y < code.Size; y++ {
		fmt.Print(white, " ")
		color_prev := white
		for x := 0; x < code.Size; x++ {
			if code.Black(x, y) {
				if color_prev != black {
					fmt.Print(black)
					color_prev = black
				}
			} else {
				if color_prev != white {
					fmt.Print(white)
					color_prev = white
				}
			}
			fmt.Print("  ")
		}
		fmt.Print(white, " ", reset, "\n")
	}
	fmt.Print(line);
}

func main() {
	ret := 0
	defer func() { os.Exit(ret) }()

	opts := &cmdOptions{}
	optsParser := flags.NewParser(opts, flags.PrintErrors)
	args, err := optsParser.Parse()
	if err != nil || len(args) > 1 {
		showHelp()
		ret = 1
		return
	}
	if opts.Help {
		showHelp()
		return
	}

	var text string
	if len(args) == 1 {
		text = args[0]
	} else {
		// FIXME: Read all input
		rd := bufio.NewReaderSize(os.Stdin, 1024)
		text_bytes, _, _ := rd.ReadLine()
		text = string(text_bytes)
	}

	code, _ := qr.Encode(text, qr.L)

	printAA(code, opts.Inverse)
}

