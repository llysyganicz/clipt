## Context

Greenfield Go module. See proposal.md for motivation. Constraints that shape
the design:

- Each mode is its own Bubble Tea model; a root model routes between them.
- Preview and edit do not exist yet, but the finder transitions into both.
- Linux is the only development and test platform; Windows and macOS must
  work through the standard library without platform-specific code.
- File writes are limited to the notes directory and the config file.

## Goals / Non-Goals

**Goals:**
- A root model whose routing and "current note" state survive the later
  addition of preview and edit without rework.
- A notes store package that preview/edit/rename/delete can reuse.
- Every file I/O path testable against a temp dir with no TUI involved.

**Non-Goals:**
- Any rendering of markdown, any text editing.
- Filesystem watching; the list is re-read on finder entry only.
- Persisting anything beyond `notes_dir`.

## Decisions

### Package layout

```
cmd/clipt/main.go          load config -> build root model -> tea.NewProgram
internal/config            Path(), Load(), Save(), Status (ok / missing / empty / bad dir)
internal/notes             Store{Dir}: List() []Note, Create(name) error, Exists(name), ValidateName
internal/ui/picker         Bubble Tea model for notes-dir-selection
internal/ui/finder         Bubble Tea model for the finder
internal/ui/app            root model: routing, current note, mode switching
```

`internal/notes` and `internal/config` are pure Go with no Bubble Tea
imports, so their tests are plain table tests on `t.TempDir()`. The UI models
are tested through `Update` with synthetic `tea.KeyMsg` values where that is
cheap, and otherwise rely on the store tests.

Alternative considered: a single `internal/tui` package. Rejected because
the mode-per-model rule reads better as one package per mode, and it keeps
the preview/edit changes from touching finder code.

### Root model and the current note

The root model holds:

```
mode        picker | finder | preview | edit
dir         string          // configured notes directory
current     *notes.Note     // nil when nothing is open
pickerFrom  startup | finder
```

Mode models do not know about each other. They return messages the root
reacts to:

```
picker  -> DirAcceptedMsg{Dir}   | PickerCancelledMsg
finder  -> OpenNoteMsg{Note}     | NoteCreatedMsg{Note} | ChangeDirMsg | BackMsg
```

The root model is the only place that sets or clears `current`:

| Event                                 | current        |
|---------------------------------------|----------------|
| OpenNoteMsg                           | that note      |
| NoteCreatedMsg                        | that note      |
| DirAcceptedMsg with a different dir   | nil            |
| BackMsg while current file is missing | nil            |
| (later) delete from preview           | nil            |
| (later) rename from preview           | renamed note   |

Alternative considered: letting the finder own the current note. Rejected
because preview and edit also change it, so it belongs above all modes.

### Preview and edit stubs

Until their own changes land, `OpenNoteMsg` and `NoteCreatedMsg` route to a
placeholder model that shows the note name and returns `BackMsg` on Esc.
This keeps the finder switcher loop (Enter, Esc, Enter) exercisable end to
end before preview exists. The placeholder lives in `internal/ui/app` and is
deleted by the preview change.

### Finder list construction

On every query change the finder rebuilds its rows from the cached listing:

```
notes      = store.List()               // sorted mtime desc, name asc
visible    = notes minus current
rows       = query == ""  ? visible
                          : filter(visible, subsequence(lower(query), lower(name)))
if trimmed(query) != "" && no note name == trimmed(query):
    rows = append(rows, createRow(trimmed(query)))
selection  = 0
```

`subsequence` is a single pass over the name's runes advancing a cursor
through the query; a name matches when the cursor reaches the end of the
query. There is no scoring: matches stay in mtime order, so adding a
character only removes rows and never reorders them. Case folding uses
`unicode.ToLower` per rune; the create-row exact-name check stays
case-sensitive because the file system decides what "exists" means.

Alternative considered: sahilm/fuzzy. Rejected; its relevance ranking would
reorder the list on every keystroke, which fights the "most recent on top"
switcher behaviour, and the match itself is a dozen lines.

The listing is read once on finder entry, not on every keystroke. Rows carry
either a `*notes.Note` or a create marker, so Enter dispatches on the row
kind rather than on list length.

Alternative considered: Bubbles `list` component. Rejected; it brings
filtering, pagination and styling opinions that fight the create-row and
hidden-current rules. A hand-rolled viewport over a slice is a few dozen
lines.

### Picker state machine

```
input --enter--> validate(path)
                   dir        -> save config -> DirAcceptedMsg
                   file       -> err, input
                   missing    -> confirm
confirm --y-->  MkdirAll -> ok: save config -> DirAcceptedMsg
                          -> fail: err, input
confirm --n/esc--> input
input --esc--> PickerCancelledMsg   (root quits on startup, else back to finder)
```

Config save happens inside the picker flow so that a failed write keeps the
user in the picker with the error, rather than entering a finder whose
directory will be forgotten next run.

### Platform neutrality

- `os.UserConfigDir()` for the config directory. On Linux it honors
  `XDG_CONFIG_HOME`, so tests set that env var to a temp dir.
- `os.UserHomeDir()` for `~` expansion; only a leading `~` or `~/` is
  expanded.
- `filepath` everywhere; no string concatenation of paths.
- `notes.ValidateName` rejects both `/` and `\` regardless of platform, so
  Linux tests cover the Windows separator rule.

### Error display

Each UI model keeps a single `err` string rendered under its input. It is
cleared on the next keypress. There is no error modal; errors never block
input.

### Theme

All styling goes through one `internal/ui/theme` package exposing named Lip
Gloss styles (selected row, create row, hint, error, prompt) built from the
gruvbox dark palette (bg `#282828`, fg `#ebdbb2`, yellow `#fabd2f`, aqua
`#8ec07c`, red `#fb4934`, gray `#928374`). Mode models never reference a
color literal, only a theme style. Lip Gloss downsamples hex colors on
terminals without true color, so no fallback palette is needed.

A user-configurable palette is a later change; it will only need to build
the same style set from config values. Keeping colors out of the mode
models now is what makes that change small.

Alternative considered: styling inline in each model with Lip Gloss
defaults. Rejected because the override change would then touch every
mode.

## Risks / Trade-offs

- [Windows file-name rules are broader than what is validated] -> OS errors
  from `os.OpenFile` are shown in the finder; a follow-up change tightens
  validation if real usage hits it.
- [mtime ordering makes the switcher follow edits, not views] -> Accepted
  trade-off (stateless, no history file). Revisit only if it proves annoying.
- [Directory listing on every finder entry] -> Flat directory of markdown
  files; `os.ReadDir` plus one `Stat` per entry is negligible for thousands
  of notes.
- [Placeholder preview/edit model ships in the code] -> Clearly named, no
  behavior beyond Esc, removed by the preview change.
- [Config save inside picker couples picker to config package] -> Acceptable;
  the picker is the only place that changes the directory.

## Open Questions

- Exact glyphs (selection marker, create-row prefix) are left to
  implementation; they do not affect any spec.
