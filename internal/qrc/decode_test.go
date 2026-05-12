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

// parseAA reverses PrintANSI's output back into a 2D module grid where
// true means a "dark" (black) module. The layout produced by PrintANSI is:
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
func parseAA(t *testing.T, data []byte, border int) [][]bool {
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

	if len(rows) < 2*border+1 {
		t.Fatalf("parseAA: too few rows: %d (border=%d)", len(rows), border)
	}
	// Drop top/bottom margin rows.
	dataRows := rows[border : len(rows)-border]

	grid := make([][]bool, 0, len(dataRows))
	for ri, r := range dataRows {
		// Strip the border-module (= 2*border spaces) left and right white margins.
		if len(r) < 4*border {
			t.Fatalf("parseAA: row %d too short (%d)", ri, len(r))
		}
		inner := r[2*border : len(r)-2*border]
		if len(inner)%2 != 0 {
			t.Fatalf("parseAA: row %d inner width %d not divisible by 2", ri, len(inner))
		}
		row := make([]bool, len(inner)/2)
		for x := range len(row) {
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
func parseSixel(t *testing.T, data []byte, border int) [][]bool {
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
	if len(bands) < 2*border+1 {
		t.Fatalf("parseSixel: too few bands: %d (border=%d)", len(bands), border)
	}
	// Drop the top and bottom all-white margin bands.
	dataBands := bands[border : len(bands)-border]
	// Strip the border-module left and right margins.
	grid := make([][]bool, 0, len(dataBands))
	for bi, b := range dataBands {
		if len(b) < 2*border+1 {
			t.Fatalf("parseSixel: band %d too narrow (%d)", bi, len(b))
		}
		row := make([]bool, len(b)-2*border)
		for i := range row {
			row[i] = b[i+border]
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
	for range n {
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
	for y := range h {
		for x := range w {
			if !grid[y][x] {
				continue
			}
			x0 := (x + quiet) * scale
			y0 := (y + quiet) * scale
			for dy := range scale {
				for dx := range scale {
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
		t.Run(in.name, func(t *testing.T) {
			grid := encode(t, in.text)
			var buf bytes.Buffer
			PrintANSI(&buf, grid, false, 1, 1)
			parsed := parseAA(t, buf.Bytes(), 1)
			got := decodeQR(t, parsed)
			if got != in.text {
				t.Errorf("AA round-trip mismatch: got %q, want %q", got, in.text)
			}
		})
	}
}

func TestRoundTripSixel(t *testing.T) {
	for _, in := range testInputs {
		t.Run(in.name, func(t *testing.T) {
			grid := encode(t, in.text)
			var buf bytes.Buffer
			PrintSixel(&buf, grid, false, 1, 1)
			parsed := parseSixel(t, buf.Bytes(), 1)
			got := decodeQR(t, parsed)
			if got != in.text {
				t.Errorf("Sixel round-trip mismatch: got %q, want %q", got, in.text)
			}
		})
	}
}

// parseUnicode reverses PrintUnicode's output back into a 2D module
// grid. PrintUnicode packs two vertically-adjacent modules into one
// half-block character; we therefore expand each line of runes into
// two module rows and then strip the surrounding quiet zone.
//
// PrintUnicode renders LIGHT modules as block characters and DARK
// modules as spaces (see the comment in PrintUnicode), so this parser
// inverts that mapping when filling in the boolean grid.
func parseUnicode(t *testing.T, data []byte, border int) [][]bool {
	t.Helper()
	var modules [][]bool
	li := 0
	for line := range bytes.Lines(data) {
		line = bytes.TrimRight(line, "\n")
		runes := []rune(string(line))
		top := make([]bool, len(runes))
		bot := make([]bool, len(runes))
		for i, r := range runes {
			// dark = space, light = block character; here `true` means
			// "QR dark module".
			switch r {
			case uBlockNone:
				top[i] = true
				bot[i] = true
			case uBlockUpper:
				bot[i] = true
			case uBlockLower:
				top[i] = true
			case uBlockFull:
			default:
				t.Fatalf("parseUnicode: unexpected rune %q at line %d col %d", r, li, i)
			}
		}
		modules = append(modules, top, bot)
		li++
	}
	if li < 2 {
		t.Fatalf("parseUnicode: too few lines: %d", li)
	}
	allFalse := func(row []bool) bool {
		for _, v := range row {
			if v {
				return false
			}
		}
		return true
	}
	top, bot := 0, len(modules)
	for top < bot && allFalse(modules[top]) {
		top++
	}
	for bot > top && allFalse(modules[bot-1]) {
		bot--
	}
	if bot <= top {
		t.Fatalf("parseUnicode: no data rows found")
	}
	dataRows := modules[top:bot]
	if len(dataRows) == 0 || len(dataRows[0]) < 2*border {
		t.Fatalf("parseUnicode: row too narrow: %d (border=%d)", len(dataRows[0]), border)
	}
	grid := make([][]bool, len(dataRows))
	for i, row := range dataRows {
		grid[i] = row[border : len(row)-border]
	}
	return grid
}

func TestRoundTripUnicode(t *testing.T) {
	for _, in := range testInputs {
		t.Run(in.name, func(t *testing.T) {
			grid := encode(t, in.text)
			var buf bytes.Buffer
			PrintUnicode(&buf, grid, false, 1, 1)
			parsed := parseUnicode(t, buf.Bytes(), 1)
			got := decodeQR(t, parsed)
			if got != in.text {
				t.Errorf("Unicode round-trip mismatch: got %q, want %q", got, in.text)
			}
		})
	}
}

// downsampleGrid collapses each scale x scale block of identical
// modules into one module. It fails the test if any block is not
// uniform.
func downsampleGrid(t *testing.T, grid [][]bool, scale int) [][]bool {
	t.Helper()
	if scale < 1 {
		t.Fatalf("downsampleGrid: invalid scale %d", scale)
	}
	if scale == 1 {
		return grid
	}
	if len(grid)%scale != 0 {
		t.Fatalf("downsampleGrid: height %d not divisible by %d", len(grid), scale)
	}
	out := make([][]bool, len(grid)/scale)
	for y := range len(out) {
		row := grid[y*scale]
		if len(row)%scale != 0 {
			t.Fatalf("downsampleGrid: width %d not divisible by %d", len(row), scale)
		}
		out[y] = make([]bool, len(row)/scale)
		for x := range len(out[y]) {
			v := row[x*scale]
			for dy := range scale {
				for dx := range scale {
					if grid[y*scale+dy][x*scale+dx] != v {
						t.Fatalf("downsampleGrid: non-uniform block at (%d,%d)", x, y)
					}
				}
			}
			out[y][x] = v
		}
	}
	return out
}

// countNewlines counts '\n' bytes in data.
func countNewlines(data []byte) int {
	return bytes.Count(data, []byte{'\n'})
}

// TestPrintANSIScale verifies that scale=N produces output whose visible
// dimensions grow exactly by N. Combined with the scale=1 round-trip
// test, this implies scale=N output is also scannable.
func TestPrintANSIScale(t *testing.T) {
	in := testInputs[0]
	grid := encode(t, in.text)
	var b1, b2 bytes.Buffer
	PrintANSI(&b1, grid, false, 1, 1)
	PrintANSI(&b2, grid, false, 2, 1)
	if got, want := countNewlines(b2.Bytes()), countNewlines(b1.Bytes())*2; got != want {
		t.Errorf("AA scale=2 line count = %d, want %d", got, want)
	}
}

// TestPrintSixelScale verifies that scale=N produces N times as many
// sixel bands and exactly twice the per-band pixel width.
func TestPrintSixelScale(t *testing.T) {
	in := testInputs[0]
	grid := encode(t, in.text)
	var b1, b2 bytes.Buffer
	PrintSixel(&b1, grid, false, 1, 1)
	PrintSixel(&b2, grid, false, 2, 1)
	// '-' separates bands, so band count = dashes + 1 (final band has no
	// trailing dash). The total number of bands must double for scale=2.
	bands1 := bytes.Count(b1.Bytes(), []byte{'-'}) + 1
	bands2 := bytes.Count(b2.Bytes(), []byte{'-'}) + 1
	if bands2 != bands1*2 {
		t.Errorf("Sixel scale=2 band count = %d, want %d", bands2, bands1*2)
	}
	// All pixel run counts (between '!' and '~') must double.
	rs1 := extractSixelRuns(b1.Bytes())
	rs2 := extractSixelRuns(b2.Bytes())
	sum := func(xs []int) int {
		s := 0
		for _, x := range xs {
			s += x
		}
		return s
	}
	// Total pixel count grows quadratically: scale=2 doubles both the
	// per-band horizontal pixel count and the number of bands.
	if got, want := sum(rs2), sum(rs1)*4; got != want {
		t.Errorf("Sixel scale=2 total run pixels = %d, want %d", got, want)
	}
}

func extractSixelRuns(data []byte) []int {
	var runs []int
	for i := range len(data) {
		if data[i] != '!' {
			continue
		}
		j := i + 1
		for j < len(data) && data[j] >= '0' && data[j] <= '9' {
			j++
		}
		if n, err := strconv.Atoi(string(data[i+1 : j])); err == nil {
			runs = append(runs, n)
		}
	}
	return runs
}

// TestPrintUnicodeScale verifies that scale=N produces N times as many
// lines and N times as many runes per line.
func TestPrintUnicodeScale(t *testing.T) {
	in := testInputs[0]
	grid := encode(t, in.text)
	var b1, b2 bytes.Buffer
	PrintUnicode(&b1, grid, false, 1, 1)
	PrintUnicode(&b2, grid, false, 2, 1)
	if got, want := countNewlines(b2.Bytes()), countNewlines(b1.Bytes())*2; got != want {
		t.Errorf("Unicode scale=2 line count = %d, want %d", got, want)
	}
	first1 := bytes.SplitN(b1.Bytes(), []byte{'\n'}, 2)[0]
	first2 := bytes.SplitN(b2.Bytes(), []byte{'\n'}, 2)[0]
	r1, r2 := utf8RuneCount(first1), utf8RuneCount(first2)
	if r2 != r1*2 {
		t.Errorf("Unicode scale=2 first-line rune count = %d, want %d", r2, r1*2)
	}
}

func utf8RuneCount(b []byte) int { return len([]rune(string(b))) }

// TestPrintANSIBorder verifies that --border N changes the output
// dimensions linearly, that border=0 produces no quiet zone, and that
// the QR code remains decodable for both border=0 and border=2.
func TestPrintANSIBorder(t *testing.T) {
	in := testInputs[0]
	grid := encode(t, in.text)
	for _, border := range []int{0, 2} {
		var buf bytes.Buffer
		PrintANSI(&buf, grid, false, 1, border)
		// Line count = size + 2*border.
		if got, want := countNewlines(buf.Bytes()), grid.Size+2*border; got != want {
			t.Errorf("AA border=%d line count = %d, want %d", border, got, want)
		}
		parsed := parseAA(t, buf.Bytes(), border)
		got := decodeQR(t, parsed)
		if got != in.text {
			t.Errorf("AA border=%d round-trip mismatch: got %q, want %q", border, got, in.text)
		}
	}
}

// TestPrintSixelBorder verifies the same for the sixel format. The
// number of bands must equal size + 2*border.
func TestPrintSixelBorder(t *testing.T) {
	in := testInputs[0]
	grid := encode(t, in.text)
	for _, border := range []int{0, 2} {
		var buf bytes.Buffer
		PrintSixel(&buf, grid, false, 1, border)
		bands := bytes.Count(buf.Bytes(), []byte{'-'}) + 1
		if got, want := bands, grid.Size+2*border; got != want {
			t.Errorf("Sixel border=%d band count = %d, want %d", border, got, want)
		}
		parsed := parseSixel(t, buf.Bytes(), border)
		got := decodeQR(t, parsed)
		if got != in.text {
			t.Errorf("Sixel border=%d round-trip mismatch: got %q, want %q", border, got, in.text)
		}
	}
}

// TestPrintUnicodeBorder verifies the same for the unicode format.
// Because each character row spans two modules, the bottom may be
// padded by one extra module when (size + 2*border) is odd.
func TestPrintUnicodeBorder(t *testing.T) {
	in := testInputs[0]
	grid := encode(t, in.text)
	for _, border := range []int{0, 2} {
		var buf bytes.Buffer
		PrintUnicode(&buf, grid, false, 1, border)
		wantRows := (grid.Size + 2*border + 1) / 2
		if got := countNewlines(buf.Bytes()); got != wantRows {
			t.Errorf("Unicode border=%d line count = %d, want %d", border, got, wantRows)
		}
		parsed := parseUnicode(t, buf.Bytes(), border)
		got := decodeQR(t, parsed)
		if got != in.text {
			t.Errorf("Unicode border=%d round-trip mismatch: got %q, want %q", border, got, in.text)
		}
	}
}
