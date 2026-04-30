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
	OutputFormat string `short:"f" long:"output-format" choice:"auto" choice:"ansi" choice:"sixel" choice:"unicode" default:"auto" description:"output format"`
	ECLevel      string `short:"l" long:"ec-level" choice:"L" choice:"M" choice:"Q" choice:"H" default:"L" description:"QR error correction level"`
	Scale        int    `short:"s" long:"scale" default:"1" description:"integer scale factor (>= 1)"`
	Border       int    `short:"b" long:"border" default:"1" description:"quiet zone width in modules (>= 0)"`
}

func showHelp() {
	const v = `Usage: qrc [OPTIONS] [TEXT]

Options:
  -h, --help
    Show this help message
  -i, --invert
    Invert color
  -f, --output-format=<auto|ansi|sixel|unicode>
    Output format (default: auto)
      auto     Sixel if the terminal supports it, otherwise ansi
      ansi     ANSI background color escape sequences
      sixel    Sixel graphics
      unicode  Unicode half-block characters (▀ ▄ █)
  -l, --ec-level=<L|M|Q|H>
    QR error correction level (default: L)
      L  Low      (~7%)
      M  Medium   (~15%)
      Q  Quartile (~25%)
      H  High     (~30%)
  -s, --scale=<N>
    Integer scale factor (default: 1, must be >= 1)
  -b, --border=<N>
    Quiet zone width in modules around the QR code
    (default: 1, must be >= 0)

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

// qrLevel maps the --ec-level string to the rsc.io/qr level constant.
func qrLevel(s string) qr.Level {
	switch s {
	case "M":
		return qr.M
	case "Q":
		return qr.Q
	case "H":
		return qr.H
	default:
		return qr.L
	}
}

func pErr(format string, a ...any) {
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

	code, err := qr.Encode(text, qrLevel(opts.ECLevel))
	if err != nil {
		pErr("encode failed: %v\n", err)
		ret = 1
		return
	}

	if opts.Scale < 1 {
		pErr("invalid --scale: %d (must be >= 1)\n", opts.Scale)
		ret = 1
		return
	}
	if opts.Border < 0 {
		pErr("invalid --border: %d (must be >= 0)\n", opts.Border)
		ret = 1
		return
	}

	switch resolveOutputFormat(opts.OutputFormat, os.Stdout) {
	case "sixel":
		qrc.PrintSixel(os.Stdout, code, opts.Inverse, opts.Scale, opts.Border)
	case "unicode":
		qrc.PrintUnicode(os.Stdout, code, opts.Inverse, opts.Scale, opts.Border)
	case "ansi":
		stdout := colorable.NewColorableStdout()
		qrc.PrintAA(stdout, code, opts.Inverse, opts.Scale, opts.Border)
	}
}
