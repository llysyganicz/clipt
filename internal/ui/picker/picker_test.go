package picker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/llysyganicz/clipt/internal/config"
	"github.com/llysyganicz/clipt/internal/ui/uitest"
)

func TestExpandPath(t *testing.T) {
	cwd, _ := os.Getwd()
	tests := []struct{ in, want string }{
		{"~", "/fake/home"},
		{"~/x", "/fake/home/x"},
		{"  ~/notes  ", "/fake/home/notes"},
		{"/abs/p", "/abs/p"},
		{"rel/p", filepath.Join(cwd, "rel/p")},
	}
	for _, tc := range tests {
		got, err := expandPath(tc.in, "/fake/home")
		if err != nil || got != tc.want {
			t.Errorf("expandPath(%q) = %q, %v; want %q", tc.in, got, err, tc.want)
		}
	}
	if _, err := expandPath("   ", "/fake/home"); err == nil {
		t.Error("empty input accepted")
	}
}

// send feeds a key and returns the new model and the message its command emits.
func send(m Model, s string) (Model, tea.Msg) {
	nm, cmd := m.Update(uitest.Key(s))
	return nm, run(cmd)
}

func setup(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
}

func TestEnterExistingDir(t *testing.T) {
	setup(t)
	dir := t.TempDir()
	m := New(Startup, dir, "")
	_, msg := send(m, "enter")
	if got, ok := msg.(DirAcceptedMsg); !ok || got.Dir != dir {
		t.Fatalf("msg = %#v, want DirAcceptedMsg{%q}", msg, dir)
	}
	if cfg, status, _ := config.Load(); status != config.OK || cfg != dir {
		t.Errorf("config = (%q, %v)", cfg, status)
	}
}

func TestEnterFile(t *testing.T) {
	setup(t)
	f := filepath.Join(t.TempDir(), "f")
	os.WriteFile(f, nil, 0o644)
	m, msg := send(New(Startup, f, ""), "enter")
	if msg != nil {
		t.Fatalf("unexpected msg %#v", msg)
	}
	if !strings.Contains(m.err, "not a directory") || m.input.Value() != f {
		t.Errorf("err = %q, input = %q", m.err, m.input.Value())
	}
}

func TestEmptyInput(t *testing.T) {
	setup(t)
	m, msg := send(New(Startup, "", ""), "enter")
	if msg != nil || !strings.Contains(m.err, "path is required") {
		t.Errorf("msg = %#v, err = %q", msg, m.err)
	}
}

func TestMissingPathConfirmFlow(t *testing.T) {
	setup(t)
	dir := filepath.Join(t.TempDir(), "a", "b")
	m, _ := send(New(Startup, dir, ""), "enter")
	if m.state != stateConfirm || !strings.Contains(m.View(), dir) {
		t.Fatalf("state = %v, view = %q", m.state, m.View())
	}
	for _, k := range []string{"n", "esc"} {
		back, msg := send(m, k)
		if back.state != stateInput || msg != nil || back.input.Value() != dir {
			t.Errorf("%q: state = %v, msg = %#v, input = %q", k, back.state, msg, back.input.Value())
		}
		if _, err := os.Stat(dir); err == nil {
			t.Errorf("%q created the directory", k)
		}
	}
	_, msg := send(m, "y")
	if got, ok := msg.(DirAcceptedMsg); !ok || got.Dir != dir {
		t.Fatalf("msg = %#v", msg)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Errorf("directory not created: %v", err)
	}
	if cfg, _, _ := config.Load(); cfg != dir {
		t.Errorf("config dir = %q, want %q", cfg, dir)
	}
}

func TestEscInInputCancels(t *testing.T) {
	_, msg := send(New(Finder, "x", ""), "esc")
	if _, ok := msg.(PickerCancelledMsg); !ok {
		t.Errorf("msg = %#v", msg)
	}
}

func TestMkdirFailureKeepsPickerOpen(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	setup(t)
	parent := t.TempDir()
	os.Chmod(parent, 0o555)
	t.Cleanup(func() { os.Chmod(parent, 0o755) })
	dir := filepath.Join(parent, "notes")
	m, _ := send(New(Startup, dir, ""), "enter")
	m, msg := send(m, "y")
	if msg != nil || m.state != stateInput || m.err == "" {
		t.Errorf("msg = %#v, state = %v, err = %q", msg, m.state, m.err)
	}
	if _, _, err := config.Load(); err != nil {
		t.Fatal(err)
	}
	if _, status, _ := config.Load(); status != config.MissingFile {
		t.Errorf("config was written, status = %v", status)
	}
}

func TestConfigSaveFailureKeepsPickerOpen(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	os.Chmod(xdg, 0o555)
	t.Cleanup(func() { os.Chmod(xdg, 0o755) })
	m, msg := send(New(Startup, t.TempDir(), ""), "enter")
	if msg != nil || m.err == "" {
		t.Errorf("msg = %#v, err = %q", msg, m.err)
	}
}

func TestErrorClearedOnKeypress(t *testing.T) {
	m := New(Startup, "/x", "boom")
	m, _ = send(m, "a")
	if m.err != "" {
		t.Errorf("err = %q", m.err)
	}
}

// run executes a command, giving up quickly on blocking ones (cursor blink).
func run(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	ch := make(chan tea.Msg, 1)
	go func() { ch <- cmd() }()
	select {
	case m := <-ch:
		if _, ok := m.(tea.BatchMsg); ok {
			return nil
		}
		return m
	case <-time.After(20 * time.Millisecond):
		return nil
	}
}

func TestCtrlQQuitsFromAnyState(t *testing.T) {
	setup(t)
	_, msg := send(New(Finder, "x", ""), "ctrl+q")
	if msg != tea.Msg(tea.QuitMsg{}) {
		t.Errorf("ctrl+q in the input -> %#v", msg)
	}
	if cfg, _, _ := config.Load(); cfg != "" {
		t.Errorf("config written by ctrl+q: %q", cfg)
	}
	dir := filepath.Join(t.TempDir(), "a", "b")
	m, _ := send(New(Startup, dir, ""), "enter")
	if m.state != stateConfirm {
		t.Fatalf("state = %v", m.state)
	}
	if _, msg := send(m, "ctrl+q"); msg != tea.Msg(tea.QuitMsg{}) {
		t.Errorf("ctrl+q in the confirmation -> %#v", msg)
	}
	if _, err := os.Stat(dir); err == nil {
		t.Error("ctrl+q created the directory")
	}
	if !strings.Contains(New(Finder, "x", "").View(), "ctrl+q quit") {
		t.Error("hint lacks ctrl+q")
	}
}
