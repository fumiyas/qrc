package qrc

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"rsc.io/qr"
)

var update = flag.Bool("update", false, "update golden files")

// testInputs lists the canonical inputs used to lock down output bytes.
// Keep this set small but representative of plain text and URL payloads
// so that any change in encoder or printer behavior is caught.
var testInputs = []struct {
	name string
	text string
}{
	{"hello", "Hello, World!"},
	{"url", "https://fumiyas.github.io/"},
	{"wifi", "WIFI:S:Our-ssid;T:WPA;P:secret;;"},
}

func encode(t *testing.T, text string) *qr.Code {
	t.Helper()
	code, err := qr.Encode(text, qr.L)
	if err != nil {
		t.Fatalf("qr.Encode(%q): %v", text, err)
	}
	return code
}

func checkGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with -update to create)", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("output for %s does not match golden file %s\n--- got (%d bytes) ---\n%q\n--- want (%d bytes) ---\n%q",
			name, path, len(got), got, len(want), want)
	}
}

func TestPrintAAGolden(t *testing.T) {
	for _, in := range testInputs {
		in := in
		for _, inv := range []bool{false, true} {
			inv := inv
			suffix := "normal"
			if inv {
				suffix = "invert"
			}
			t.Run(in.name+"_"+suffix, func(t *testing.T) {
				grid := encode(t, in.text)
				var buf bytes.Buffer
				PrintAA(&buf, grid, inv, 1, 1)
				checkGolden(t, "aa_"+in.name+"_"+suffix, buf.Bytes())
			})
		}
	}
}

func TestPrintSixelGolden(t *testing.T) {
	for _, in := range testInputs {
		in := in
		for _, inv := range []bool{false, true} {
			inv := inv
			suffix := "normal"
			if inv {
				suffix = "invert"
			}
			t.Run(in.name+"_"+suffix, func(t *testing.T) {
				grid := encode(t, in.text)
				var buf bytes.Buffer
				PrintSixel(&buf, grid, inv, 1, 1)
				checkGolden(t, "sixel_"+in.name+"_"+suffix, buf.Bytes())
			})
		}
	}
}

func TestPrintUnicodeGolden(t *testing.T) {
	for _, in := range testInputs {
		in := in
		for _, inv := range []bool{false, true} {
			inv := inv
			suffix := "normal"
			if inv {
				suffix = "invert"
			}
			t.Run(in.name+"_"+suffix, func(t *testing.T) {
				grid := encode(t, in.text)
				var buf bytes.Buffer
				PrintUnicode(&buf, grid, inv, 1, 1)
				checkGolden(t, "unicode_"+in.name+"_"+suffix, buf.Bytes())
			})
		}
	}
}
