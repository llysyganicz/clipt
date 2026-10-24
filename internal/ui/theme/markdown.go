package theme

import (
	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
)

func ptr[T any](v T) *T { return &v }

// markdownStyle builds the rendered-markdown style for p. It takes the
// structural settings (indents, margins, prefixes) from Glamour's dark style
// and replaces every color it defines with a palette slot, so the result
// uses no color outside the palette and a light palette renders correctly.
// Fenced code is not syntax highlighted: all code text is bright cyan.
func markdownStyle(p Palette) ansi.StyleConfig {
	c := styles.DarkStyleConfig

	c.Document.Color = ptr(p.Foreground)
	c.Text.Color = nil
	c.Emph.Color = nil
	c.Strong.Color = nil

	c.Heading.Color = ptr(p.BrightYellow)
	c.H1.Color = ptr(p.BrightYellow)
	c.H1.BackgroundColor = nil
	c.H6.Color = ptr(p.BrightYellow)
	c.Table.Color = ptr(p.BrightYellow)

	c.Link.Color = ptr(p.BrightCyan)
	c.LinkText.Color = ptr(p.BrightCyan)
	c.Code.Color = ptr(p.BrightCyan)
	c.Code.BackgroundColor = nil
	c.CodeBlock.Color = ptr(p.BrightCyan)
	c.CodeBlock.Chroma = nil
	c.CodeBlock.Theme = ""

	c.BlockQuote.Color = ptr(p.BrightBlack)
	c.HorizontalRule.Color = ptr(p.BrightBlack)
	c.Image.Color = ptr(p.BrightBlack)
	c.ImageText.Color = ptr(p.BrightBlack)

	c.Item.Color = ptr(p.BrightBlue)
	c.Enumeration.Color = ptr(p.BrightBlue)
	c.Task.Color = ptr(p.BrightGreen)
	return c
}
