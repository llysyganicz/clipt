## 1. Dependency and templates package

- [x] 1.1 Add `github.com/goccy/go-yaml` (v1.19.x) to `go.mod` with `go get` and verify `go build ./...` and `go mod tidy` leave a clean diff apart from the new requirement
- [x] 1.2 Create `internal/templates` with `Dir`, `Store` and `EnsureDir` over the notes directory and verify temp-dir tests that `Dir` is `<notes>/.templates`, `Store(dir).List()` returns `fs.ErrNotExist` when the directory is missing, and `EnsureDir` creates it and is idempotent
- [x] 1.3 Implement `Parse` (CRLF normalization, `---` fence detection, YAML decode of `form`, field validation per the templates spec, body with leading blank lines and trailing newlines stripped) and verify a table test covers: no front matter, front matter with fields, CRLF input, unterminated fence, invalid YAML, missing name, name with braces or whitespace, duplicate name, unknown type, option without options, numeric and boolean option values, extra front matter keys ignored, label defaulting to name, one and several blank lines after the fence, blank lines inside the body kept
- [x] 1.4 Implement `Field.Default` (empty, `0`, first option) and `Template.Expand` as a single regexp pass and verify tests for repeated placeholders, unknown placeholders left intact, a value containing a placeholder not re-expanded, and `{{ name }}` with spaces not matched
- [x] 1.5 Add the `Skeleton` constant from the design and verify a test that `Parse(Skeleton)` succeeds with exactly one `text` field named `title`

## 2. Finder template scope

- [x] 2.1 Add `Scope` with `ScopeNotes`/`ScopeTemplates`, the `scope` field, `NewTemplates(store, query)`, `Scope()` and the `Scope` field on `OpenNoteMsg` and `NoteCreatedMsg`, and verify existing finder tests still pass with the notes scope reported on both messages
- [x] 2.2 Make `reload` treat `fs.ErrNotExist` as an empty listing in template scope and verify a temp-dir test that `NewTemplates` on a notes dir without `.templates` lists nothing and shows no error, and that with `.templates/meeting.md` and `todo.md` present only `meeting` is listed
- [x] 2.3 Make the create row in template scope call `EnsureDir`, `Create` and `Write(Skeleton)` and emit `NoteCreatedMsg{Scope: ScopeTemplates}`, and verify a temp-dir test that Enter on `Create "meeting"` creates `.templates/` and `.templates/meeting.md` with the skeleton, plus that `a/b` is rejected with the validation error and no file
- [x] 2.4 Bind `ctrl+t` to emit `ToggleScopeMsg` in both scopes and verify a test that the message is produced from each scope
- [x] 2.5 Render the `templates> ` prompt, the "No templates yet" hint and the template-scope footer hints, and verify View tests for the prompt, the hint on an empty listing, and that the current note is not hidden in template scope when a template shares its name

## 3. Editor insertion and template kind

- [x] 3.1 Add the `template` flag, `NewTemplate(store, note)`, `DoneMsg.Template`, and disable `TransposeCharacterBackward` in the key map; verify temp-dir tests that `NewTemplate` reads and saves `.templates/<name>.md`, that `DoneMsg.Template` is true for a template and false for a note, and that `ctrl+t` after `ab` leaves the line as `ab`
- [x] 3.2 Emit `InsertTemplateMsg` on `ctrl+t` for a note and nothing for a template, and verify tests for both cases
- [x] 3.3 Implement `Insert(text) Model` using `InsertString` plus scroll-to-cursor and verify a test that inserting `# Standup\n- ` with the cursor between `a` and `b` yields `a# Standup\n- b` with the cursor after the inserted text, and that Esc afterwards writes the inserted content
- [x] 3.4 Update the status bar hints (`ctrl+t template` on a note, none on a template) and the right-hand `template: <name>` label, and verify View tests for both kinds

## 4. Insert mode

- [x] 4.1 Create `internal/ui/tmpl` with `New(store)`, the picker state (query input, listing from `store.List()`, `finder.Match` filtering, selection, window), `InsertMsg`, `CancelMsg`, and verify temp-dir tests that the picker lists templates newest first, filters with the query, resets the selection on query change, does not wrap, shows the "create templates from the finder" hint when empty, and emits `CancelMsg` on Esc
- [x] 4.2 Implement Enter in the picker: read and `Parse` the template, emit `InsertMsg{Body}` for a field-less template, open the form otherwise, show read/parse errors in place; verify temp-dir tests for all three outcomes and for Enter on an empty list doing nothing
- [x] 4.3 Implement the form state: rows built from the fields with defaults, focus movement (Down/Tab/Enter next, Up/Shift+Tab previous), option cycling with Left/Right without wrapping, text and number rows editing via `textinput`; verify tests for initial values, cycling at both ends, and focus moving through three rows
- [x] 4.4 Implement submit on Enter at the last row: `ParseFloat` validation of number rows with the error naming the field and focus moved to it, otherwise `InsertMsg{Expand(values)}`; verify tests for an invalid number, a decimal `1.50` expanding verbatim, a valid submit emitting the expanded body, and Esc from the form emitting `CancelMsg`
- [x] 4.5 Implement `View` for both states (list or rows, error line, status bar hints) and `WindowSizeMsg` handling with the 80x24 default, and verify View tests that the picker shows the query and rows, the form shows each label with its value and highlights the selected option, and the status bar lists the keys

## 5. Root integration

- [x] 5.1 Add `modeTemplate` and the `tmpl` field, route `finder.ToggleScopeMsg` and template-scope `finder.BackMsg` between `newFinder` and a `newTemplatesFinder` helper, route `OpenNoteMsg`/`NoteCreatedMsg` with `ScopeTemplates` to `editor.NewTemplate` without touching `current`, and `editor.DoneMsg{Template: true}` back to template scope; verify app tests on a temp dir that `ctrl+t` toggles scope and back, that Enter on a template then Esc returns to template scope with `current` unchanged, and that `ctrl+o` from template scope lands in notes scope after cancel
- [x] 5.2 Route `editor.InsertTemplateMsg` to `tmpl.New` with the editor model kept, `tmpl.InsertMsg` to `m.edit.Insert` and back to edit mode, `tmpl.CancelMsg` back to edit mode, and forward `WindowSizeMsg` in `modeTemplate`; verify app tests that `ctrl+t`, pick, fill and submit lands in the editor with the expanded text at the cursor, that cancel lands back with content and cursor unchanged, and that a read failure on a template keeps the finder in template scope with the error and `current` intact

## 6. Verification

- [x] 6.1 Run `gofmt -l .`, `go vet ./...` and `go test ./...` and verify all are clean
- [x] 6.2 Manual pass on Linux: create a template from the finder, edit its form, return to a note, insert it mid-line with `ctrl+t`, confirm the form defaults and option cycling, confirm the note saves the inserted text, confirm `.templates` never shows in the notes list, and confirm a malformed template shows its error in the picker
