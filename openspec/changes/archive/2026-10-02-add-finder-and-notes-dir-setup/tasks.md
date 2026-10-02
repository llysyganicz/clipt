## 1. Module setup

- [x] 1.1 Initialize the Go module `github.com/llysyganicz/clipt`, add Bubble Tea, Bubbles, Lip Gloss and BurntSushi/toml, and verify `go build ./...` succeeds with a placeholder `cmd/clipt/main.go`
- [x] 1.2 Add a `.gitignore` for the built binary and verify `gofmt -l .` and `go vet ./...` are clean
- [x] 1.3 Add `internal/ui/theme` with the gruvbox dark palette and the named styles from design.md, and verify a test that every exported style renders a non-empty string and no mode package imports a color literal (grep for `#` hex colors outside `theme`)

## 2. Config package (`internal/config`)

- [x] 2.1 Implement `Path()` using `os.UserConfigDir` and verify a test with `XDG_CONFIG_HOME` set to a temp dir returns `<tmp>/clipt/config.toml`
- [x] 2.2 Implement `Load()` returning the notes dir and a status (ok, missing file, empty dir, bad dir) and verify temp-dir tests for: no file, empty `notes_dir`, unknown keys ignored, path that is a file, path that does not exist
- [x] 2.3 Implement `Save(dir)` creating the config directory and writing `notes_dir` and verify a temp-dir test round-trips through `Load()` and a test on a read-only config dir returns an error
- [x] 2.4 Make `Load()` return a distinct parse error for invalid TOML and verify a temp-dir test with `notes_dir = ` yields that error and leaves the file unchanged

## 3. Notes store (`internal/notes`)

- [x] 3.1 Implement `ValidateName` rejecting empty, `.`, `..`, `/` and `\` and verify a table test covers each rejection and accepts `foo.md` and `meeting notes`
- [x] 3.2 Implement `Store.List()` returning `.md` regular files, skipping dotfiles, directories and other extensions, sorted by mtime desc then name, with display name stripped of `.md`, and verify a temp-dir test with mixed entries and controlled mtimes (`os.Chtimes`)
- [x] 3.3 Implement `Store.Create(name)` writing an empty `<name>.md` without overwriting, and `Store.Exists(name)`, and verify temp-dir tests for creation, literal `foo.md` -> `foo.md.md`, refusal when the file exists, and an error on a read-only directory

## 4. Picker model (`internal/ui/picker`)

- [x] 4.1 Implement path expansion (`~`, trim, `filepath.Abs`) as a pure function and verify a test with a fake home dir covers `~`, `~/x`, relative and absolute inputs
- [x] 4.2 Implement the picker model with input and confirm states, origin (startup or finder) and prefill, emitting `DirAcceptedMsg` / `PickerCancelledMsg`, and verify `Update` tests with synthetic keys cover: Enter on an existing temp dir, Enter on a file path shows an error, Enter on a missing path shows the y/n prompt, `n` and Esc return to input, Esc in input emits cancel
- [x] 4.3 Wire directory creation and config save into the accept path and verify a temp-dir test that `y` creates nested dirs and writes the config, and that a failed mkdir keeps the picker open with an error and no config written

## 5. Finder model (`internal/ui/finder`)

- [x] 5.1 Implement the case-insensitive subsequence matcher as a pure function and verify a table test covers: match, non-match, case folding, empty query matches everything, literal `foo.md` against `foo` does not match, and multi-byte runes
- [x] 5.2 Implement row building (hide current, filter by matcher keeping mtime order, append create row) as a pure function over `[]notes.Note` and verify table tests cover: empty query order, current note hidden, filtered rows keep mtime order, create row position, no create row on exact name, no create row when the exact name is the hidden current note, no create row on blank query
- [x] 5.3 Implement the finder model: text input, selection with Up/Down and `ctrl+p`/`ctrl+n` without wrapping, selection reset on query change, `ctrl+o` -> `ChangeDirMsg`, Esc -> `BackMsg` without clearing the query, `ctrl+c` -> quit, and verify `Update` tests with synthetic keys
- [x] 5.4 Implement Enter: note row -> `OpenNoteMsg`, create row -> `Store.Create` then `NoteCreatedMsg`, invalid name or create failure -> inline error, empty list -> no-op, and verify temp-dir tests for the create path including the read-only directory error
- [x] 5.5 Implement the view: query line, rows with a selection marker, create row rendered as `Create "<query>"`, inline error line, and the empty-directory hint, and verify `View()` snapshot-style tests for the empty hint and for a list with a create row

## 6. Root model and entrypoint (`internal/ui/app`, `cmd/clipt`)

- [x] 6.1 Implement the root model with mode routing, `dir`, `current` and `pickerFrom`, handling `DirAcceptedMsg` (clear current only on a different dir, reset query), `PickerCancelledMsg` (quit at startup, otherwise restore finder), `OpenNoteMsg` / `NoteCreatedMsg` (set current), and `BackMsg` (preview of current, clear current with error if its file is gone, quit when none), and verify `Update` tests cover each transition with a temp notes dir
- [x] 6.2 Add the placeholder preview/edit model that shows the note name and returns `BackMsg` on Esc, and verify a test that Enter, Esc, Enter in the finder toggles between the two most recent notes
- [x] 6.3 Implement `cmd/clipt/main.go`: load config, print the config path and exit non-zero on a parse error, start in the picker (with stale path and error when the configured dir is bad) or the finder, and verify by running the binary with `XDG_CONFIG_HOME` pointing at an empty temp dir and seeing the picker prefilled with `~/notes`

## 7. Verification

- [x] 7.1 Run `gofmt -l .`, `go vet ./...` and `go test ./...` and verify all pass
- [x] 7.2 Walk the first-run flow manually on Linux against a temp notes dir: accept `~/notes` equivalent, decline then accept creation, create a note, open it, Esc back, confirm it is hidden, `ctrl+o` to another dir and back, and verify each step matches the scenarios in the three specs
