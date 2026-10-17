package uitest

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func TestKeyNames(t *testing.T) {
	for _, name := range []string{
		"enter", "esc", "tab", "backspace", "delete", "up", "down", "left", "right",
		"home", "end", "pgup", "pgdown", "space",
		"ctrl+c", "ctrl+q", "ctrl+t", "ctrl+o", "ctrl+p", "ctrl+n",
		"shift+tab", "shift+right", "ctrl+shift+right", "alt+backspace",
		"e", "r", "d", "y", "N", "É",
	} {
		if got := Key(name).String(); got != name {
			t.Errorf("Key(%q).String() = %q", name, got)
		}
	}
}

func TestKeyFields(t *testing.T) {
	if k := Key("ctrl+q"); k.Code != 'q' || k.Mod != tea.ModCtrl || k.Text != "" {
		t.Errorf("ctrl+q = %+v", k)
	}
	if k := Key("shift+tab"); k.Code != tea.KeyTab || k.Mod != tea.ModShift {
		t.Errorf("shift+tab = %+v", k)
	}
	if k := Key("esc"); k.Code != tea.KeyEscape || k.Mod != 0 {
		t.Errorf("esc = %+v", k)
	}
	if k := Key("e"); k.Code != 'e' || k.Text != "e" || k.Mod != 0 {
		t.Errorf("e = %+v", k)
	}
}

func TestKeyPanicsOnUnknown(t *testing.T) {
	for _, name := range []string{"ctrl+", "hyper+a", "nosuchkey", ""} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Key(%q) did not panic", name)
				}
			}()
			Key(name)
		}()
	}
}

func TestType(t *testing.T) {
	if got := Type("enter"); len(got) != 1 || got[0].String() != "enter" {
		t.Errorf("Type(enter) = %v", got)
	}
	got := Type("to do")
	want := []string{"t", "o", "space", "d", "o"}
	if len(got) != len(want) {
		t.Fatalf("Type(to do) = %v", got)
	}
	for i, k := range got {
		if k.String() != want[i] {
			t.Errorf("Type(to do)[%d] = %q, want %q", i, k.String(), want[i])
		}
	}
	if got[2].Text != " " {
		t.Errorf("space text = %q", got[2].Text)
	}
	if got := Type(""); len(got) != 0 {
		t.Errorf("Type(\"\") = %v", got)
	}
}

func TestStrip(t *testing.T) {
	styled := lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true).Render("> a")
	if styled == "> a" {
		t.Fatal("lipgloss rendered no escape codes; the test is vacuous")
	}
	if got := Strip(styled); got != "> a" {
		t.Errorf("Strip(styled) = %q", got)
	}
	link := "see \x1b]8;;https://example.com\x1b\\here\x1b]8;;\x1b\\ now"
	if got := Strip(link); got != "see here now" {
		t.Errorf("Strip(link) = %q", got)
	}
	if got := Strip("plain"); got != "plain" {
		t.Errorf("Strip(plain) = %q", got)
	}
}
