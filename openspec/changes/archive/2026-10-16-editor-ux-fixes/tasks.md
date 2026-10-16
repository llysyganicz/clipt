## 1. Line number gutter

- [x] 1.1 Copy the text area's `wrap` function into `internal/ui/editor` with a test that, for a note with one line longer than the width, compares the editor's visual row count with the rows the text area renders before its end-of-buffer padding
- [x] 1.2 Add gutter label building (number on the first row of each logical line, blanks on continuation rows, width `digits(lineCount)+2`, cursor line styled yellow) and verify with unit tests on a 12-line and a 120-line value and on a wrapped line
- [x] 1.3 Turn the text area's own line numbers off, install the prompt callback, rebuild labels after every update and resize, and re-call `SetPromptFunc` plus `SetWidth` when the digit count changes; verify with rendered-view tests (temp dir note) that the text column is identical on lines 9 and 10, on lines 99 and 100, and on both rows of a wrapped line
- [x] 1.4 Verify the existing editor tests for cursor visibility on entry, resize and `Insert` still pass with the new gutter (`go test ./internal/ui/editor/`)

## 2. Indentation keys

- [x] 2.1 Handle Tab in the editor by inserting two spaces at the cursor; verify with tests for Tab at line start, Tab inside a line, and that the note is written with the spaces on Esc (temp dir)
- [x] 2.2 Handle Shift+Tab by removing up to two leading spaces from the cursor's line and restoring the cursor; verify with tests for four, one and zero leading spaces, the cursor column after removal, and a dedent on a line below the first screen keeping the view on the cursor

## 3. ctrl+q

- [x] 3.1 Editor: `ctrl+q` runs the Esc save path and returns `tea.Quit` on success, stays open with the error on failure; verify with temp-dir tests for changed content (file written, quit command), unchanged content (mtime kept, quit command) and a read-only directory (no quit, error shown)
- [x] 3.2 Finder, preview and notes-dir picker: `ctrl+q` returns `tea.Quit` from every state (both finder scopes, preview rename and delete prompts, picker path input and create confirmation); verify with a test per model and state
- [x] 3.3 Template picker and form: `ctrl+q` emits `tmpl.QuitMsg`; root model returns to edit mode discarding the insertion and invokes the editor's save-and-quit; verify with app tests (temp dir) for a successful save and for a failing write that lands back in edit mode with the error
- [x] 3.4 Add `ctrl+q` to every hint string (`ctrl+q save & quit` in the editor, `ctrl+q quit` elsewhere) and verify with the existing status-bar and footer tests updated to expect it

## 4. Docs and checks

- [x] 4.1 Update the README key tables with `ctrl+q`, Tab and Shift+Tab and verify the four mode tables each list `ctrl+q`
- [x] 4.2 Run `gofmt -l .`, `go vet ./...` and `go test ./...` and verify all pass
