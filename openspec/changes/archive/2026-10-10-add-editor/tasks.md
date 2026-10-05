## 1. Store

- [x] 1.1 Add `Normalize(content string) string` to `internal/notes` (trim trailing newlines, append exactly one) and verify a table test covers empty, no newline, one newline and several newlines
- [x] 1.2 Add `Store.Write(name, content string) error` writing `<name>.md` in place and verify temp-dir tests cover creating a new file, overwriting an existing one, recreating a removed one, and failing on a read-only directory

## 2. Theme

- [x] 2.1 Add `theme.EditorStyles()` returning focused and blurred textarea styles from the gruvbox palette (gray line numbers, yellow cursor line number, normal text, no cursor line background) and verify `go vet ./...` passes and no color literal appears outside `internal/ui/theme`

## 3. Editor model

- [x] 3.1 Create `internal/ui/editor` with `New(store, note) (Model, error)` that reads the note, converts `\r\n` to `\n`, loads a textarea configured per design (line numbers on, empty prompt, no char/height/width limits, default key map, theme styles, 80x24 default size) and records the normalized baseline; verify a temp-dir test that `New` fails on a missing file and that a CRLF file shows two lines with no blank line between
- [x] 3.2 Implement `Update`: `ctrl+c` returns `tea.Quit`, Esc writes `Normalize(Value())` only when it differs from the baseline and emits `DoneMsg`, write failures set `err` and keep the content, every other key goes to the textarea and clears `err`; verify temp-dir tests for typed-then-Esc writes normalized content, Esc without edits leaves file bytes and mtime (set via `os.Chtimes`) untouched, type-then-delete does not write, file without trailing newline is not rewritten, tab file is untouched without edits and has four spaces after an edit, trailing blank lines collapse to one newline, read-only dir keeps content and sets `err`, Esc after the dir is writable again succeeds, and a removed file is recreated
- [x] 3.3 Implement `View` with the textarea above a status bar (`esc save & back` left, note name right, left-truncated with an ellipsis when narrow) and the error line replacing the last content line, plus `tea.WindowSizeMsg` resizing the textarea to `height-1`; verify tests that the View contains the note name, the hint, the error when set, and the last line of a note taller than the terminal on entry
- [x] 3.4 Return the textarea blink command from `Init` and verify the cursor blinks when running the app manually

## 4. Root integration

- [x] 4.1 Route `finder.NoteCreatedMsg` and `preview.EditMsg` to `editor.New`, falling back to the finder with the existing "cannot read" error on failure, and handle `editor.DoneMsg` by re-entering preview of the current note; forward `tea.WindowSizeMsg` to the editor while in edit mode; verify app tests that create-from-finder lands in the editor, Esc after typing lands in preview showing the typed content, and `e` then Esc returns to preview
- [x] 4.2 Delete `internal/ui/app/placeholder.go` and its tests and verify `go build ./...` and `go test ./...` pass with no reference to the placeholder

## 5. Verification

- [x] 5.1 Run `gofmt -l .`, `go vet ./...` and `go test ./...` and verify all are clean
- [x] 5.2 Manual pass on Linux: create a note from the finder, type, Esc, confirm preview renders it; open an old note, press `e` then Esc and confirm its finder position does not change; resize the terminal in edit mode and confirm the text area fills it
