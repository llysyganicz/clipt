package theme

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// PaintScreen draws content into a screen of width by height cells and gives
// every cell without a background the palette background and every cell
// without a foreground the palette foreground. Cells that already carry one
// (the editor selection, colored text) keep it. Lines are padded to the width
// and the screen to the height, so the empty cells after each line and below
// the last line are painted too. The root view uses it on Windows, where the
// terminal's own default colors cannot be set and reset safely.
//
// With a zero width or height, before the first window size message, content
// is returned unchanged.
func PaintScreen(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return content
	}
	scr := uv.NewScreenBuffer(width, height)
	uv.NewStyledString(content).Draw(scr, scr.Bounds())
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := scr.CellAt(x, y)
			if c == nil || c.Width == 0 {
				continue // the continuation of a wide character
			}
			if c.Style.Bg == nil {
				c.Style.Bg = background
			}
			if c.Style.Fg == nil {
				c.Style.Fg = foreground
			}
		}
	}
	// The buffer's own renderer keeps the painted cells at the end of every
	// line; the Lip Gloss canvas would trim them as trailing spaces.
	return scr.Render()
}
