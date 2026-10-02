package main

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/llysyganicz/clipt/internal/config"
	"github.com/llysyganicz/clipt/internal/ui/app"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "clipt:", err)
		os.Exit(1)
	}
}

func run() error {
	dir, status, err := config.Load()
	var perr *config.ParseError
	if errors.As(err, &perr) {
		return err // message names the config path
	}

	var m app.Model
	switch status {
	case config.OK:
		m = app.New(dir)
	case config.BadDir:
		m = app.NewPicker(dir, err.Error())
	default:
		if err != nil {
			return err
		}
		m = app.NewPicker("", "")
	}
	_, err = tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}
