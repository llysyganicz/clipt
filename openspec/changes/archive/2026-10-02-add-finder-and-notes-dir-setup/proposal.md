## Why

clipt has no code yet. The first bootable slice is the loop a user hits every
session: load the config, pick a notes directory when none is configured, and
land in the fuzzy finder to open or create a note. Everything else (preview,
edit, rename, delete) hangs off the finder, so it has to exist first.

## What Changes

- Add the `cmd/clipt` entrypoint and a root Bubble Tea model that routes between
  modes and holds the optional "current note".
- Add config loading and saving: `clipt/config.toml` under the platform config
  directory, with a single `notes_dir` key.
- Add the notes directory picker mode: a single path input with `~` expansion,
  shown on first run, when the configured directory is missing, or on `ctrl+o`
  from the finder. Offers to create a missing directory after a y/n prompt.
- Add the finder mode: lists `*.md` notes by modification time (newest first),
  filters them by a case-insensitive subsequence match on the query while
  keeping that order, hides the current note, and appends a
  "Create `<query>`" row. Enter on a note opens it in preview; enter on the
  create row writes `<query>.md` and enters edit.
- Define the finder as a switcher: Esc returns to the current note's preview,
  or quits when no note is current.
- Target Linux, Windows and macOS through Go's platform-neutral APIs
  (`os.UserConfigDir`, `os.UserHomeDir`, `filepath`). Development and testing
  happen on Linux only; Windows and macOS get a manual pass after v1 and any
  fixes arrive as separate changes.

## Capabilities

### New Capabilities

- `config`: location, shape, reading and writing of the config file, and what
  counts as "not configured".
- `notes-dir-selection`: the picker mode: entry conditions, path validation,
  create-or-refuse, config write, and Esc semantics for both entry points.
- `finder`: listing, ranking, navigation, the create row, enter semantics,
  current-note handling and the `ctrl+o` / Esc / `ctrl+c` keys.

### Modified Capabilities

None.

## Non-goals

- Preview and edit modes. The finder transitions into them; this change only
  needs stubs for the transitions to compile. Rename and delete come with
  preview.
- Content search, folders, tags, sync, vim normal mode (out of scope for v1).
- Live feedback in the picker (note counts, path completion, directory
  browsing).
- A full Windows file-name validation set. Only path separators, empty names,
  `.` and `..` are rejected. Other OS rejections surface as an error in the
  finder and are fixed in a follow-up change if they matter.
- Platform-specific testing. Tests run on Linux; platform differences are
  handled by the standard library.

## Impact

- New Go module at `github.com/llysyganicz/clipt`.
- New packages under `internal/`: config, notes store, picker model, finder
  model, root model.
- New dependencies: Bubble Tea, Bubbles (textinput), Lip Gloss,
  BurntSushi/toml. Fuzzy matching is a small in-house subsequence match; no
  matching library.
- Writes files only inside the notes directory, except for the config file
  under the platform config directory.
