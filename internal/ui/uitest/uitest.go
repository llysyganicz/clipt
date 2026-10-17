// Package uitest holds helpers for the UI packages' tests: building key
// press messages from the names the models match on, and stripping ANSI
// escape sequences from rendered views. It has no production callers.
package uitest

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

var named = map[string]rune{
	"enter":     tea.KeyEnter,
	"esc":       tea.KeyEscape,
	"tab":       tea.KeyTab,
	"backspace": tea.KeyBackspace,
	"delete":    tea.KeyDelete,
	"space":     tea.KeySpace,
	"up":        tea.KeyUp,
	"down":      tea.KeyDown,
	"left":      tea.KeyLeft,
	"right":     tea.KeyRight,
	"home":      tea.KeyHome,
	"end":       tea.KeyEnd,
	"pgup":      tea.KeyPgUp,
	"pgdown":    tea.KeyPgDown,
}

// Key builds the key press a terminal would report for name: a named key
// ("esc", "pgdown"), one with modifiers ("ctrl+q", "shift+tab",
// "ctrl+shift+right"), or a single printable character ("e", "É"), whose
// Text is set. Key(name).String() == name for every form. It panics on
// anything else.
func Key(name string) tea.KeyPressMsg {
	k, ok := parse(name)
	if !ok {
		panic("uitest.Key: unknown key " + name)
	}
	return k
}

// Type returns the key presses that produce s: a single press when s is a
// key name Key accepts, otherwise one press per character, so that
// Type("todo") types the word and Type("enter") presses Enter.
func Type(s string) []tea.KeyPressMsg {
	if k, ok := parse(s); ok {
		return []tea.KeyPressMsg{k}
	}
	var out []tea.KeyPressMsg
	for _, r := range s {
		out = append(out, Key(string(r)))
	}
	return out
}

func parse(name string) (tea.KeyPressMsg, bool) {
	var k tea.Key
	parts := strings.Split(name, "+")
	for _, mod := range parts[:len(parts)-1] {
		switch mod {
		case "ctrl":
			k.Mod |= tea.ModCtrl
		case "shift":
			k.Mod |= tea.ModShift
		case "alt":
			k.Mod |= tea.ModAlt
		default:
			return tea.KeyPressMsg{}, false
		}
	}
	last := parts[len(parts)-1]
	if code, ok := named[last]; ok {
		k.Code = code
		return tea.KeyPressMsg(k), true
	}
	runes := []rune(last)
	if len(runes) != 1 {
		return tea.KeyPressMsg{}, false
	}
	k.Code = runes[0]
	if k.Mod == 0 {
		k.Text = last
	}
	return tea.KeyPressMsg(k), true
}

// Strip removes every ANSI escape sequence (styles, hyperlinks, cursor
// movement) from s, leaving the visible text.
func Strip(s string) string { return ansi.Strip(s) }
