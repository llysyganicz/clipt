package theme

import (
	"image/color"
	"regexp"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// cells parses a painted screen back into its grid of cells.
func cells(s string) []uv.Line { return uv.NewStyledString(s).Lines(ansi.WcWidth) }

func sameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

func TestPaintScreenPaintsEveryCell(t *testing.T) {
	light := withLight(t)
	bg := lipgloss.Color(light.Background)
	fg := lipgloss.Color(light.Foreground)
	sel := lipgloss.NewStyle().Background(lipgloss.Color(light.Selection))
	content := "ab\n" + Selected.Render("cd") + "\n" + sel.Render("ef")
	const width, height = 6, 5
	out := PaintScreen(content, width, height)

	lines := cells(out)
	if len(lines) != height {
		t.Fatalf("painted %d lines, want %d:\n%q", len(lines), height, out)
	}
	for y, line := range lines {
		if len(line) != width {
			t.Fatalf("line %d has %d cells, want %d:\n%q", y, len(line), width, out)
		}
		for x := range line {
			c := line.At(x)
			wantBg := bg
			if y == 2 && x < 2 {
				wantBg = lipgloss.Color(light.Selection)
			}
			if !sameColor(c.Style.Bg, wantBg) {
				t.Errorf("cell (%d,%d) background = %v, want %v", x, y, c.Style.Bg, wantBg)
			}
			wantFg := fg
			if y == 1 && x < 2 {
				wantFg = lipgloss.Color(light.BrightYellow)
			}
			if !sameColor(c.Style.Fg, wantFg) {
				t.Errorf("cell (%d,%d) foreground = %v, want %v", x, y, c.Style.Fg, wantFg)
			}
		}
	}
	if got := ansi.Strip(out); !strings.HasPrefix(got, "ab    \n") {
		t.Errorf("first line is not padded to the width: %q", got)
	}
}

func TestPaintScreenWithoutSizeReturnsContent(t *testing.T) {
	withLight(t)
	s := "ab\n" + Selected.Render("cd")
	for _, size := range [][2]int{{0, 0}, {5, 0}, {0, 5}} {
		if got := PaintScreen(s, size[0], size[1]); got != s {
			t.Errorf("PaintScreen(s, %d, %d) = %q, want s unchanged", size[0], size[1], got)
		}
	}
}

// sgrColor matches a truecolor foreground or background SGR parameter.
var sgrColor = regexp.MustCompile(`[34]8;2;(\d+);(\d+);(\d+)`)

func TestPaintScreenUsesOnlyPaletteColors(t *testing.T) {
	t.Cleanup(func() { Apply(mustBuiltin("gruvbox_dark")) })
	for _, name := range Builtin() {
		p := mustBuiltin(name)
		Apply(p)
		allowed := map[string]bool{}
		for _, hex := range p.Colors() {
			allowed[strings.TrimPrefix(uitestFg(hex), "38;2;")] = true
		}
		sel := lipgloss.NewStyle().Background(lipgloss.Color(p.Selection))
		content := strings.Join([]string{
			Normal.Render("normal") + " plain",
			Selected.Render("selected") + " " + Create.Render("create"),
			Hint.Render("hint") + " " + Error.Render("error") + " " + Prompt.Render("prompt"),
			Tag.Render("tag") + " " + RowTag.Render("rowtag") + " " + LineNumber.Render("1") + CursorLineNumber.Render("2"),
			sel.Render("selection"),
		}, "\n")
		out := PaintScreen(content, 40, 8)
		for _, m := range sgrColor.FindAllStringSubmatch(out, -1) {
			rgb := m[1] + ";" + m[2] + ";" + m[3]
			if !allowed[rgb] {
				t.Errorf("%s: painted output uses color %s outside the palette", name, rgb)
			}
		}
		if !strings.Contains(out, "48;2;") {
			t.Errorf("%s: painted output sets no background", name)
		}
	}
}
