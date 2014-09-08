package main

import (
	"bufio"
	"code.google.com/p/rsc/qr"
	"github.com/mattn/go-colorable"
	"github.com/jessevdk/go-flags"
	"os"

	"github.com/fumiyas/qrc/lib"
	"github.com/fumiyas/qrc/tty"
)

type cmdOptions struct {
	Help    bool `short:"h" long:"help" description:"show this help message"`
	Inverse bool `short:"i" long:"invert" description:"invert color"`
}

func showHelp() {
	const v = `Usage: qrc [OPTIONS] [TEXT]

Options:
  -h, --help
    Show this help message
  -i, --invert
    Invert color

Text examples:
  URLTO:http://www.example.jp/
  MAILTO:foobar@example.jp
  WIFI:S:myssid;T:WPA;P:pass123;;
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
		qrc.PrintSixel(os.Stdout, code, opts.Inverse)
	} else {
		stdout := colorable.NewColorableStdout()
		qrc.PrintAA(stdout, code, opts.Inverse)
	}
}
