## Context

See proposal.md for motivation. What is already in place:

- The root model in `internal/ui/app` routes between modes and owns
  `current`. It reacts to messages from mode models and already has the
  "current note vanished" path for Esc from the finder.
- A `placeholder` model stands in for both preview and edit. The archived
  finder design promised the preview change would remove it.
- `internal/notes.Store` has List, Exists, Create and ValidateName. Preview
  needs Read, Rename and Delete.
- `internal/ui/theme` holds every style; mode packages never use color
  literals. Glamour styling has to follow the same rule.
- The picker already implements the prompt-plus-y/n-confirm pattern that
  rename and delete reuse.

## Goals / Non-Goals

**Goals:**
- A preview model with no knowledge of the finder or the root; it only emits
  messages.
- Rename and delete file operations in the store, tested on a temp dir with
  no TUI involved.
- Markdown rendering that matches the gruvbox palette without any color
  literal outside the theme package.

**Non-Goals:**
- Caching rendered output across visits.
- File watching.
- The edit model. A trimmed placeholder keeps the `e` transition exercisable.

## Decisions

### Package and messages

```
internal/ui/preview     Model, New(store, note) (Model, error)
                        -> BackMsg{} | EditMsg{Note} | RenamedMsg{Note} | DeletedMsg{}
internal/ui/app         routes OpenNoteMsg / finder BackMsg into preview,
                        handles the four preview messages
internal/notes          Read(name) (string, error), Rename(old, new) error, Delete(name) error
internal/ui/theme       Glamour style for rendered markdown
```

`New` reads the file up front and returns the error instead of a model, so
the root can fall back to the finder before preview is ever shown. This keeps
the "cannot be read" requirement a root-level decision, like the existing
Esc path, rather than a preview state.

Root table additions:

| Event              | current | mode                               |
|--------------------|---------|------------------------------------|
| preview.RenamedMsg | renamed | stays preview                      |
| preview.DeletedMsg | nil     | finder, empty query                |
| preview.BackMsg    | kept    | finder, empty query                |
| preview.EditMsg    | kept    | edit (placeholder for now)         |
| New returns error  | nil     | finder, error "cannot read <name>" |

The finder is rebuilt with an empty query on every return, which is what
`noteBackMsg` already does today.

Alternative considered: preview reads lazily in `Init` and shows an error
state. Rejected; the user asked for "back to the finder with an error", and
a model that may or may not have content complicates every View branch.

### Rendering and scrolling

Content is rendered with Glamour at the current width and the result is
handed to a Bubbles `viewport` sized to `height - 1` (status bar) and the
full width. On `tea.WindowSizeMsg` the markdown is re-rendered and the
viewport reset to the top, since Glamour's word wrap depends on width.
Scroll keys are the viewport's own Up/Down/PgUp/PgDn handling; its default
key map also binds `j`/`k`, `d`/`u`, `f`/`b` and space, which collide with
the preview keys, so the key map is replaced with the four arrow/page keys
only. The viewport width is set so wrapping is done by Glamour rather than by
the viewport.

Rendering happens on entry and on resize, not on every keypress. A note of a
few thousand lines renders in well under a frame on Glamour, so there is no
need for an async render command.

Alternative considered: render once per note and keep the viewport across
visits. Rejected; the spec requires re-reading on entry and starting at the
top, so there is nothing to keep.

### Preview state machine

```
view --r--> rename  --enter--> validate -> ok: store.Rename -> RenamedMsg, view
                               (same name) -> view
                               (invalid / exists / os error) -> err, stay
            rename  --esc----> view
view --d--> confirm --y------> store.Delete -> ok: DeletedMsg
                                              fail: err, view
            confirm --n/esc--> view
view --e--> EditMsg
view --esc> BackMsg
```

`err` is a single string shown above the status bar and cleared on the next
keypress, as in the finder and picker. The rename prompt reuses Bubbles
`textinput`, prefilled and with the cursor at the end, exactly like the
picker.

The rename succeeds without re-reading or re-rendering the content; only the
name shown in the status bar changes.

### Refusing to overwrite on rename

`os.Rename` replaces the destination on POSIX, so `Store.Rename` checks
`Exists(new)` first and returns an "already exists" error. The check-then-
rename race is accepted for a single-user tool. On case-insensitive file
systems `Exists("Todo")` is true when `todo.md` exists, so a case-only rename
is refused; the proposal lists this as a v1 non-goal.

Alternative considered: `os.Link` then `os.Remove` for an atomic
no-overwrite rename. Rejected; hard links are not available on every
Windows file system and the behaviour would differ per platform.

### Status bar layout

Hints are rendered with the existing `theme.Hint` style on the left; the
note name is right-aligned with Lip Gloss `PlaceHorizontal` on the remaining
width. When the width is too small for both, the name is truncated from the
left with an ellipsis so the hints stay readable. Rename and delete prompts
replace the whole line.

### Glamour style in the theme package

Glamour accepts a `StyleConfig` struct; the theme package exposes one built
from the existing palette constants (headings yellow, links and code aqua,
emphasis fg, block quotes and horizontal rules gray) and preview renders with
`glamour.WithStyles(theme.Markdown)` plus `WithWordWrap(width)`. Only the
handful of elements that need a color are set; everything else falls back to
Glamour's dark style values by copying that config and overriding fields.

Fallback agreed with the user: if the style config fights the palette or
takes disproportionate effort, the task is dropped from this change and
`glamour.WithStandardStyle("dark")` ships. The preview code is unchanged
either way; only the theme value differs.

Alternative considered: a JSON style file embedded with `embed`. Rejected;
colors would then live in a JSON literal outside the Go palette constants,
breaking the one-theme-package rule.

### Placeholder trimming

The placeholder loses its preview role and is kept only for the `e`
transition: it shows "<name> (edit)" and returns to preview on Esc, which
matches "edit always returns to preview". The edit change deletes it.

### Testing

- `notes`: Read, Rename (success, target exists, invalid name, source
  missing), Delete (success, missing) on `t.TempDir()`.
- `preview`: `Update` with synthetic `tea.KeyMsg` for rename/delete/esc/e,
  asserting on emitted messages and on files in a temp dir. Rendering is
  asserted only loosely (non-empty output, contains the note name in the
  status bar) because Glamour output is ANSI-heavy.
- `app`: the switcher tests gain rename-updates-current,
  delete-clears-current-and-returns-to-finder, and
  unreadable-note-returns-to-finder.

## Risks / Trade-offs

- [Glamour adds a sizeable dependency tree (goldmark, chroma, termenv)] ->
  Accepted; it is the chosen renderer and there is no lighter option that
  handles tables and code blocks.
- [Glamour rendering of very large notes on every resize] -> Resize events
  are rare; if it ever matters, debounce in a later change.
- [Viewport default key map collisions] -> Replaced explicitly; a test
  presses `d` and asserts the confirm prompt appears rather than a page
  scroll.
- [Check-then-rename race] -> Single-user tool, accepted.
- [Status bar truncation at narrow widths] -> Hints win over the name;
  below the hint width the line simply clips.
- [Custom Glamour style may not cover every element] -> Unset fields inherit
  from the copied dark style, so nothing renders unstyled.

## Open Questions

- Exact hint wording and separator glyphs in the status bar are left to
  implementation; they do not affect any scenario.
