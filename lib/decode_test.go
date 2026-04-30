package qrc

import (
	"bytes"
	"image"
	"image/color"
	"strconv"
	"testing"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// parseAA reverses PrintAA's output back into a 2D module grid where
// true means a "dark" (black) module. The layout produced by PrintAA is:
//
//   - one full-width white margin row at top and bottom
//   - each data row starts and ends with a 1-character white margin and
//     uses 2 spaces per module; the active background color is set via
//     ESC[0;30;47m (white) or ESC[0;30;40m (black).
//
// Color state changes are emitted with SGR sequences. Consecutive spaces
// that share a color are coalesced; we therefore walk the byte stream,
// track the current background color, and convert every 2 consecutive
// spaces into one module of that color.
func parseAA(t *testing.T, data []byte) [][]bool {
	t.Helper()

	type pixel struct {
		dark bool
	}
	var rows [][]pixel
	var current []pixel
	colorBlack := false
	hasColor := false

	i := 0
	for i < len(data) {
		c := data[i]
		switch {
		case c == 0x1b && i+1 < len(data) && data[i+1] == '[':
			// Parse SGR escape: ESC [ params m
			j := i + 2
			for j < len(data) && data[j] != 'm' {
				j++
			}
			if j >= len(data) {
				t.Fatalf("parseAA: unterminated escape at offset %d", i)
			}
			params := string(data[i+2 : j])
			// Look for 47 (white bg) or 40 (black bg).
			if containsParam(params, "47") {
				colorBlack = false
				hasColor = true
			} else if containsParam(params, "40") {
				colorBlack = true
				hasColor = true
			} else if params == "0" {
				hasColor = false
			}
			i = j + 1
		case c == '\n':
			rows = append(rows, current)
			current = nil
			i++
		case c == ' ':
			if !hasColor {
				t.Fatalf("parseAA: space without active color at offset %d", i)
			}
			current = append(current, pixel{dark: colorBlack})
			i++
		default:
			t.Fatalf("parseAA: unexpected byte %q at offset %d", c, i)
		}
	}
	if len(current) > 0 {
		rows = append(rows, current)
	}

	if len(rows) < 3 {
		t.Fatalf("parseAA: too few rows: %d", len(rows))
	}
	// Drop top/bottom margin rows.
	dataRows := rows[1 : len(rows)-1]

	grid := make([][]bool, 0, len(dataRows))
	for ri, r := range dataRows {
		// Strip 1-char left and right white margins.
		if len(r) < 4 {
			t.Fatalf("parseAA: row %d too short (%d)", ri, len(r))
		}
		inner := r[1 : len(r)-1]
		if len(inner)%2 != 0 {
			t.Fatalf("parseAA: row %d inner width %d not divisible by 2", ri, len(inner))
		}
		row := make([]bool, len(inner)/2)
		for x := 0; x < len(row); x++ {
			a, b := inner[x*2], inner[x*2+1]
			if a.dark != b.dark {
				t.Fatalf("parseAA: row %d module %d color mismatch", ri, x)
			}
			row[x] = a.dark
		}
		grid = append(grid, row)
	}
	return grid
}

// containsParam checks whether `;`-separated SGR params include p.
func containsParam(params, p string) bool {
	start := 0
	for i := 0; i <= len(params); i++ {
		if i == len(params) || params[i] == ';' {
			if params[start:i] == p {
				return true
			}
			start = i + 1
		}
	}
	return false
}

// parseSixel reverses PrintSixel's output back into a 2D module grid.
//
// PrintSixel renders each module as a 6x6 pixel block using two color
// registers (#0 = black, #1 = white). A row of modules is emitted as a
// single 6-row band terminated by '-'. Within a band, runs are written
// as "#N!K~" or "#N~" (single sixel), with '~' meaning all six vertical
// pixels are lit. We therefore only need to walk per-band runs and
// downsample by 6 horizontally.
func parseSixel(t *testing.T, data []byte) [][]bool {
	t.Helper()

	// Locate end of introducer (\x1BPq"...) up to first non-parameter character.
	if !bytes.HasPrefix(data, []byte("\x1bPq")) {
		t.Fatalf("parseSixel: missing introducer")
	}
	// Strip terminator \x1B\\ if present.
	if end := bytes.LastIndex(data, []byte("\x1b\\")); end >= 0 {
		data = data[:end]
	}
	// Skip past the "\x1bPq" and any aspect-ratio params up to the first '#'.
	p := 3
	for p < len(data) && data[p] != '#' {
		p++
	}

	var bands [][]bool // each entry is one band (one row of modules)
	var current []bool
	color := -1 // current selected color register

	for p < len(data) {
		c := data[p]
		switch c {
		case '#':
			// Either a color definition "#N;2;R;G;B" or a color selection "#N".
			p++
			n, np := readNum(data, p)
			p = np
			if p < len(data) && data[p] == ';' {
				// Color definition: skip ;2;R;G;B (5 number-or-semicolon segments).
				// Skip until we reach a non-digit, non-semicolon char.
				for p < len(data) && (data[p] == ';' || (data[p] >= '0' && data[p] <= '9')) {
					p++
				}
			} else {
				color = n
			}
		case '!':
			p++
			n, np := readNum(data, p)
			p = np
			if p >= len(data) {
				t.Fatalf("parseSixel: '!' run without sixel char")
			}
			six := data[p]
			p++
			appendSixelRun(t, &current, color, six, n)
		case '-':
			bands = append(bands, downsample6(t, current))
			current = nil
			p++
		case '$':
			// Carriage return within a band: would re-overlay the same band.
			// PrintSixel does not emit '$', so treat as unsupported.
			t.Fatalf("parseSixel: unexpected '$' at %d", p)
		default:
			if c >= 0x3f && c <= 0x7e {
				appendSixelRun(t, &current, color, c, 1)
				p++
			} else {
				t.Fatalf("parseSixel: unexpected byte %q at %d", c, p)
			}
		}
	}
	if len(current) > 0 {
		bands = append(bands, downsample6(t, current))
	}
	if len(bands) < 3 {
		t.Fatalf("parseSixel: too few bands: %d", len(bands))
	}
	// Drop the top and bottom all-white margin bands.
	dataBands := bands[1 : len(bands)-1]
	// Strip the 1-module left and right margins.
	grid := make([][]bool, 0, len(dataBands))
	for bi, b := range dataBands {
		if len(b) < 3 {
			t.Fatalf("parseSixel: band %d too narrow (%d)", bi, len(b))
		}
		row := make([]bool, len(b)-2)
		for i := range row {
			row[i] = b[i+1]
		}
		grid = append(grid, row)
	}
	return grid
}

func readNum(data []byte, p int) (int, int) {
	start := p
	for p < len(data) && data[p] >= '0' && data[p] <= '9' {
		p++
	}
	if start == p {
		return 0, p
	}
	n, _ := strconv.Atoi(string(data[start:p]))
	return n, p
}

func appendSixelRun(t *testing.T, current *[]bool, color int, six byte, n int) {
	t.Helper()
	if six != '~' {
		// PrintSixel only emits '~' (all six vertical bits set).
		t.Fatalf("parseSixel: only '~' supported, got %q", six)
	}
	if color < 0 {
		t.Fatalf("parseSixel: pixel without selected color")
	}
	dark := color == 0
	for i := 0; i < n; i++ {
		*current = append(*current, dark)
	}
}

func downsample6(t *testing.T, band []bool) []bool {
	t.Helper()
	if len(band)%6 != 0 {
		t.Fatalf("parseSixel: band width %d not divisible by 6", len(band))
	}
	out := make([]bool, len(band)/6)
	for i := range out {
		v := band[i*6]
		for j := 1; j < 6; j++ {
			if band[i*6+j] != v {
				t.Fatalf("parseSixel: non-uniform 6-pixel column at %d", i)
			}
		}
		out[i] = v
	}
	return out
}

// gridToImage renders a 2D module grid into a grayscale image suitable
// for QR decoding. Each module becomes scale x scale pixels and a quiet
// zone of `quiet` modules is added around the data.
func gridToImage(grid [][]bool) *image.Gray {
	const scale = 8
	const quiet = 4
	if len(grid) == 0 {
		return image.NewGray(image.Rect(0, 0, 1, 1))
	}
	h := len(grid)
	w := len(grid[0])
	pxW := (w + 2*quiet) * scale
	pxH := (h + 2*quiet) * scale
	img := image.NewGray(image.Rect(0, 0, pxW, pxH))
	// Fill with white.
	for i := range img.Pix {
		img.Pix[i] = 0xff
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !grid[y][x] {
				continue
			}
			x0 := (x + quiet) * scale
			y0 := (y + quiet) * scale
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					img.SetGray(x0+dx, y0+dy, color.Gray{Y: 0})
				}
			}
		}
	}
	return img
}

func decodeQR(t *testing.T, grid [][]bool) string {
	t.Helper()
	img := gridToImage(grid)
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		t.Fatalf("NewBinaryBitmapFromImage: %v", err)
	}
	res, err := qrcode.NewQRCodeReader().Decode(bmp, nil)
	if err != nil {
		t.Fatalf("QR decode: %v", err)
	}
	return res.GetText()
}

func TestRoundTripAA(t *testing.T) {
	for _, in := range testInputs {
		in := in
		t.Run(in.name, func(t *testing.T) {
			grid := encode(t, in.text)
			var buf bytes.Buffer
			PrintAA(&buf, grid, false)
			parsed := parseAA(t, buf.Bytes())
			got := decodeQR(t, parsed)
			if got != in.text {
				t.Errorf("AA round-trip mismatch: got %q, want %q", got, in.text)
			}
		})
	}
}

func TestRoundTripSixel(t *testing.T) {
	for _, in := range testInputs {
		in := in
		t.Run(in.name, func(t *testing.T) {
			grid := encode(t, in.text)
			var buf bytes.Buffer
			PrintSixel(&buf, grid, false)
			parsed := parseSixel(t, buf.Bytes())
			got := decodeQR(t, parsed)
			if got != in.text {
				t.Errorf("Sixel round-trip mismatch: got %q, want %q", got, in.text)
			}
		})
	}
}
