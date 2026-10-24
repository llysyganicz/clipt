package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/llysyganicz/clipt/internal/config"
	"github.com/llysyganicz/clipt/internal/ui/app"
	"github.com/llysyganicz/clipt/internal/ui/theme"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "clipt:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, status, cfgErr := config.Load()
	if cfgErr != nil && status != config.BadDir {
		return cfgErr // a parse error names the config path
	}

	// The theme is applied before any model exists, so every mode is built
	// with the final styles.
	name := cfg.Theme
	if name == "" {
		name = theme.Default
	}
	cfgDir, err := config.Dir()
	if err != nil {
		return err
	}
	palette, err := theme.Load(name, cfgDir)
	if err != nil {
		return err
	}
	theme.Apply(palette)

	var m app.Model
	switch status {
	case config.OK:
		m = app.New(cfg.NotesDir, cfg.Theme)
	case config.BadDir:
		m = app.NewPicker(cfg.NotesDir, cfgErr.Error(), cfg.Theme)
	default:
		m = app.NewPicker("", "", cfg.Theme)
	}
	_, err = tea.NewProgram(m).Run()
	return err
}
