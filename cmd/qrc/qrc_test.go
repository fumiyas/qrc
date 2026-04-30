package main

import "testing"

func TestResolveOutputFormatExplicit(t *testing.T) {
	// Explicit values must pass through unchanged regardless of the
	// terminal capabilities of the writer.
	for _, f := range []string{"ansi", "sixel"} {
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
