package main

import (
	"os"
	"fmt"
	"bufio"
	"github.com/jessevdk/go-flags"
	"code.google.com/p/rsc/qr"

	"github.com/fumiyas/qrc/lib"
	"github.com/fumiyas/qrc/tty"
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

	da1, err := tty.GetDeviceAttributes1(os.Stdout)
	if err == nil && da1[tty.DA1_SIXEL] {
		fmt.Printf("FIXME: Print QR Code in Sixel format\n")
	} else {
		qrc.PrintAA(code, opts.Inverse)
	}
}
