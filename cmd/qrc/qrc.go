package main

import (
	"fmt"
	"io"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mattn/go-colorable"
	"rsc.io/qr"

	"github.com/fumiyas/qrc/internal/qrc"
	"github.com/fumiyas/qrc/internal/tty"
)

type cmdOptions struct {
	Help         bool   `short:"h" long:"help" description:"show this help message"`
	Inverse      bool   `short:"i" long:"invert" description:"invert color"`
	OutputFormat string `short:"f" long:"output-format" choice:"auto" choice:"ansi" choice:"sixel" default:"auto" description:"output format"`
}

func showHelp() {
	const v = `Usage: qrc [OPTIONS] [TEXT]

Options:
  -h, --help
    Show this help message
  -i, --invert
    Invert color
  -f, --output-format=<auto|ansi|sixel>
    Output format (default: auto)
      auto   Sixel if the terminal supports it, otherwise ansi
      ansi   ANSI background color escape sequences
      sixel  Sixel graphics

Text examples:
  http://www.example.jp/
  MAILTO:foobar@example.jp
  WIFI:S:myssid;T:WPA;P:pass123;;
`

	os.Stderr.Write([]byte(v))
}

// resolveOutputFormat returns the concrete output format to use.
// When format is "auto", probe the terminal via DA1 to choose between
// "sixel" and "ansi".
func resolveOutputFormat(format string, w *os.File) string {
	if format != "auto" {
		return format
	}
	da1, err := tty.GetDeviceAttributes1(w)
	if err == nil && da1[tty.DA1_SIXEL] {
		return "sixel"
	}
	return "ansi"
}

func pErr(format string, a ...interface{}) {
	fmt.Fprint(os.Stderr, os.Args[0], ": ")
	fmt.Fprintf(os.Stderr, format, a...)
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
		textBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			pErr("read from stdin failed: %v\n", err)
			ret = 1
			return
		}
		text = string(textBytes)
	}

	code, err := qr.Encode(text, qr.L)
	if err != nil {
		pErr("encode failed: %v\n", err)
		ret = 1
		return
	}

	switch resolveOutputFormat(opts.OutputFormat, os.Stdout) {
	case "sixel":
		qrc.PrintSixel(os.Stdout, code, opts.Inverse)
	case "ansi":
		stdout := colorable.NewColorableStdout()
		qrc.PrintAA(stdout, code, opts.Inverse)
	}
}
