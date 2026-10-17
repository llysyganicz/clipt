## 1. Dependencies and test helper

- [x] 1.1 Pin `charm.land/bubbletea/v2@v2.1.0`, `charm.land/bubbles/v2@v2.2.1`, `charm.land/lipgloss/v2@v2.0.6`, `charm.land/glamour/v2@v2.0.1` and `github.com/charmbracelet/x/ansi` in `go.mod`, drop the four v1 requirements, and verify `go mod tidy` leaves no `github.com/charmbracelet/bubbletea`, `bubbles`, `lipgloss` or `glamour` line while keeping `mattn/go-runewidth` and `rivo/uniseg` direct
- [x] 1.2 Add `internal/ui/uitest` with `Key(name) tea.KeyPressMsg` (named keys, `ctrl+`/`shift+` modifiers, single printable characters with `Text` set) and `Strip(s)` over `ansi.Strip`; verify with unit tests that `Key("ctrl+q").String() == "ctrl+q"`, `Key("shift+tab").String() == "shift+tab"`, `Key("esc").String() == "esc"`, `Key("e").Text == "e"`, and that `Strip` removes a Lip Gloss-rendered style and an OSC 8 hyperlink

## 2. Theme

- [x] 2.1 Switch `internal/ui/theme` to `charm.land/lipgloss/v2`, change `EditorStyles` to return one `textarea.Styles` (same eight styles per state, empty `Selection`, cursor as a blinking block in the theme foreground) and add `InputStyles() textinput.Styles` (prompt `Prompt`, unstyled text, `Hint` placeholder and suggestion, same cursor); verify `theme_test.go` updated to the new shapes passes, including the no-background check on `Focused.CursorLine` and a new check that `InputStyles().Blurred.Text` has no foreground
- [x] 2.2 Switch `theme/markdown.go` and the Glamour test to `charm.land/glamour/v2/...` and verify `TestMarkdownRenders` and `TestNoColorLiteralsOutsideTheme` pass

## 3. Mode models

- [x] 3.1 Port `internal/ui/finder` to v2: `Update` returns `(Model, tea.Cmd)`, `tea.KeyPressMsg` matched by name, `SetStyles(theme.InputStyles())`; rewrite the test key parser over `uitest.Key`, replace the regex stripper with `uitest.Strip`, drop the `nm.(Model)` assertions; verify `go test ./internal/ui/finder/` passes (temp-dir stores as today)
- [x] 3.2 Port `internal/ui/picker` the same way and verify `go test ./internal/ui/picker/` passes, including the `ctrl+q` and Esc cases
- [x] 3.3 Port `internal/ui/preview`: viewport via `viewport.New(WithWidth, WithHeight)` and `SetWidth`/`SetHeight` in `render`, key names for `esc`/`e`/`r`/`d`/`y`/`n`/`ctrl+c`/`ctrl+q`, rename input styles from the theme; rewrite the test helpers over `uitest.Key` and verify `go test ./internal/ui/preview/` passes, including scrolling with `up`/`down`/`pgup`/`pgdown` and the narrow status bar width check
- [x] 3.4 Port `internal/ui/tmpl`: key names (`tab`, `shift+tab`, `left`, `right`, arrows, `ctrl+p`/`ctrl+n`), form inputs styled from the theme, `Update` returning `Model`; rewrite the key parser and stripper over `uitest` and verify `go test ./internal/ui/tmpl/` passes (temp-dir template stores)
- [x] 3.5 Port `internal/ui/editor`: `textarea.DefaultKeyMap()` with transpose, the six selection bindings, `SelectAll` and `CopySelection` disabled; `SetStyles(theme.EditorStyles())`; `SetCursorColumn`; prompt callbacks taking `textarea.PromptInfo`; key names for `esc`/`tab`/`shift+tab`/`ctrl+c`/`ctrl+q`/`ctrl+t`; update the `wrap.go` comment to cite Bubbles v2.2.1; verify `go test ./internal/ui/editor/` passes with gutter assertions (`HasPrefix`, `textCol`) reading `uitest.Strip` output, and add a test that `shift+right` leaves value and cursor unchanged
- [x] 3.6 Port `internal/ui/app`: assign sub-model results directly (no assertions), `View` returns `tea.NewView(...)` with `AltScreen = true`; rewrite `press`/`feed`/`rune1` over `uitest.Key`; verify `go test ./internal/ui/app/` passes (temp-dir notes) and a new test asserts `View().AltScreen` is true in finder, preview and edit modes

## 4. Entrypoint and gate

- [x] 4.1 Switch `cmd/clipt/main.go` to `charm.land/bubbletea/v2`, drop `tea.WithAltScreen()`, and verify `go build ./...` succeeds
- [x] 4.2 Run `gofmt -l .` (empty output), `go vet ./...` and `go test ./...` and verify all pass; verify `grep -r "github.com/charmbracelet/\(bubbletea\|bubbles\|lipgloss\|glamour\)" --include=*.go .` finds nothing
- [x] 4.3 Run `go run ./cmd/clipt` in a Linux terminal and walk the manual pass from design.md (finder, preview incl. page keys and rename/delete, editor gutter across line 10 plus Tab/Shift+Tab and `ctrl+t` insert, notes-dir picker, resize, Esc and `ctrl+q` in every mode, `shift+arrows` inert in the editor); verify no visual or key regression against the current release
