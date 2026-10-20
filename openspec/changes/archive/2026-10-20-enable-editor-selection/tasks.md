## 1. Theme

- [x] 1.1 Add a `bg2` gruvbox constant and set the text area's focused and blurred `Selection` style to that background in `EditorStyles`; verify `TestEditorStylesRender` (extended with a selection render check) passes.

## 2. Selection

- [x] 2.1 Stop disabling the text area's selection and copy bindings in the editor's `load` (only `TransposeCharacterBackward` stays disabled); verify editor tests for Shift+Left, Shift+Down, Ctrl+Shift+Right and `ctrl+g` select the expected text via `SelectedText()`.
- [x] 2.2 Add editor tests for the replace and clear semantics: typing, Backspace, Delete across lines, unshifted movement, and Tab and template insertion replacing a selection (the editor's `Insert` deletes the selection before inserting, so Tab and `InsertTemplate` both go through it); verify they pass.
- [x] 2.3 Add an editor test that selecting all and pressing Esc on an unchanged note does not write the file (temp dir); verify it passes.
- [x] 2.4 Add an editor test that Shift+Left selects in a template opened with `NewTemplate`; verify it passes.

## 3. Copy

- [x] 3.1 After forwarding a message to the text area, move a non-nil `Err` from the text area into the editor's status-line error and clear the text area's field; verify editor tests that `ctrl+shift+c` leaves the content, cursor and selection unchanged (with and without a selection) and that an error set on the text area is rendered above the status bar and gone after the next key press.

## 4. Status bar

- [x] 4.1 Append `ctrl+shift+c copy` to the hints (note and template variants) while `HasSelection()` is true; verify editor tests check the hint appears after Shift+Left and disappears after Right.

## 5. Wrap-up

- [x] 5.1 Run `gofmt -l .`, `go vet ./...` and `go test ./...`; verify all pass.
- [x] 5.2 Update the product context in `openspec/config.yaml` (edit mode is no longer "insert-only, no selection") and the README's key list if it documents edit-mode keys; verify the text matches the spec.
