package main

import (
	"testing"

	"rsc.io/qr"
)

func TestResolveOutputFormatExplicit(t *testing.T) {
	// Explicit values must pass through unchanged regardless of the
	// terminal capabilities of the writer.
	for _, f := range []string{"ansi", "sixel", "unicode"} {
		if got := resolveOutputFormat(f, nil); got != f {
			t.Errorf("resolveOutputFormat(%q) = %q, want %q", f, got, f)
		}
	}
}

func TestResolveOutputFormatAutoFallsBackToANSI(t *testing.T) {
	// A nil writer cannot be probed for DA1, so auto must fall back to
	// the safe default of "ansi".
	if got := resolveOutputFormat("auto", nil); got != "ansi" {
		t.Errorf("resolveOutputFormat(auto, nil) = %q, want ansi", got)
	}
}

func TestQRLevel(t *testing.T) {
	cases := []struct {
		in   string
		want qr.Level
	}{
		{"L", qr.L},
		{"M", qr.M},
		{"Q", qr.Q},
		{"H", qr.H},
		{"", qr.L},        // default
		{"unknown", qr.L}, // safe fallback
	}
	for _, c := range cases {
		if got := qrLevel(c.in); got != c.want {
			t.Errorf("qrLevel(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
