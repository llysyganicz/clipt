## 1. Notes store

- [x] 1.1 Add `Store.Read(name)` returning the file content; test on a temp dir for an existing note, a missing note and an empty file
- [x] 1.2 Add `Store.Rename(old, new)` that validates the new name, refuses when the target exists and otherwise renames; temp-dir tests for success, same name, target exists, invalid name and missing source
- [x] 1.3 Add `Store.Delete(name)`; temp-dir tests for success and a missing file, asserting no other file is touched

## 2. Theme

- [x] 2.1 Add Glamour to go.mod and verify `go build ./...` and `go mod tidy` leave a clean tree
- [x] 2.2 Add a Glamour style value to `internal/ui/theme` built from the palette constants on top of a copy of the dark style; test that rendering `# h` and a code block through it produces non-empty output. If this proves disproportionate, replace the value with the standard dark style, note it in design.md, and keep the test

## 3. Preview model

- [x] 3.1 Create `internal/ui/preview` with `New(store, note) (Model, error)` that reads the file, renders it with the theme style at the current width and fills a viewport; test that a missing file returns an error and a present file yields a model whose View contains the note name
- [x] 3.2 Replace the viewport key map with Up/Down/PgUp/PgDn only; test that `d`, `u`, `f`, `b`, `j`, `k` and space do not scroll
- [x] 3.3 Handle `tea.WindowSizeMsg` by re-rendering at the new width and resetting to the top; test that View changes width after a resize
- [x] 3.4 Render the status bar with hints on the left and the note name right-aligned, truncating the name when narrow; test the name is present at a wide width and the hints survive a narrow one
- [x] 3.5 Emit `BackMsg` on Esc, `EditMsg` on `e`, quit on `ctrl+c`, and ignore other keys; test each via `Update`
- [x] 3.6 Implement the rename prompt (prefilled textinput replacing the status bar): Enter renames through the store and emits `RenamedMsg`, same name or Esc cancels, invalid or existing target shows an error and keeps the prompt; temp-dir tests for each path, asserting on files and on the emitted message
- [x] 3.7 Implement the delete confirmation: `y` deletes through the store and emits `DeletedMsg`, `n`/Esc cancels, a failed delete shows the error; temp-dir tests for each path

## 4. Root routing

- [x] 4.1 Route `finder.OpenNoteMsg` and the Esc-from-finder path into `preview.New`; on error stay in the finder with a "cannot read" error and clear current; test with a note removed between listing and Enter
- [x] 4.2 Handle `preview.BackMsg` (finder, empty query), `preview.RenamedMsg` (update current, stay in preview) and `preview.DeletedMsg` (clear current, finder, empty query); extend `app_test.go` with rename-updates-current and delete-clears-current cases
- [x] 4.3 Trim the placeholder to the edit role only: created on `preview.EditMsg` and on `finder.NoteCreatedMsg`, returning to preview on Esc; test that Esc from the placeholder lands in preview of the same note
- [x] 4.4 Forward `tea.WindowSizeMsg` to the preview model and verify by running clipt and resizing the terminal

## 5. Verification

- [x] 5.1 `gofmt -l .`, `go vet ./...` and `go test ./...` pass
- [x] 5.2 Manual pass: open a long note and scroll, rename it, rename to an existing name, delete it, and confirm the finder listing and hidden current note behave as the spec scenarios describe
