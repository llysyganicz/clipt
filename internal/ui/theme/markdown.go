package theme

import (
	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
)

func ptr[T any](v T) *T { return &v }

// Markdown styles rendered markdown with the gruvbox palette. It starts from
// a copy of Glamour's dark style and overrides only the colored elements;
// override fields are replaced, never mutated, so the original is untouched.
var Markdown = markdownStyle()

func markdownStyle() ansi.StyleConfig {
	c := styles.DarkStyleConfig

	c.Document.Color = ptr(fg)
	c.Heading.Color = ptr(yellow)
	c.H1.Color = ptr(yellow)
	c.H1.BackgroundColor = nil
	c.H6.Color = ptr(yellow)
	c.BlockQuote.Color = ptr(gray)
	c.HorizontalRule.Color = ptr(gray)
	c.Link.Color = ptr(aqua)
	c.LinkText.Color = ptr(aqua)
	c.Code.Color = ptr(aqua)
	c.Code.BackgroundColor = nil
	c.CodeBlock.Color = ptr(aqua)
	return c
}
