## Context

See proposal.md for motivation. What is already in place:

- The root model in `internal/ui/app` has an `openEdit` path used by both
  `finder.NoteCreatedMsg` and `preview.EditMsg`, and a `noteBackMsg` handler
  that re-enters preview of the current note (or the finder when there is
  none). Both currently target `placeholder.go`.
- `preview.New` re-reads the note from disk, so the editor only has to write
  the file before emitting its back message; no content is handed around.
- `internal/notes.Store` has List, Exists, Create, Read, Rename, Delete and
  ValidateName. The editor needs Write.
- `internal/ui/theme` holds every style; mode packages never use color
  literals.
- Bubbles v1.0.0 `textarea`: no character limit by default, but `MaxHeight`
  is 99 and `MaxWidth` 500 and both clamp `SetHeight`/`SetWidth`. Line
  numbers are on by default with a fixed 4-column gutter. Input passes
  through a sanitizer that turns every `\r` and `\n` into a line break, every
  tab into four spaces, and drops other control characters and invalid UTF-8.
  `SetValue` resets the buffer and inserts the text at the cursor, which
  leaves the cursor at the end of the document. Esc and `ctrl+c` are not
  bound by its default key map.

## Goals / Non-Goals

**Goals:**
- An editor model with no knowledge of the finder, preview or root; it only
  emits a message when it is done.
- A write operation in the store, tested on a temp dir with no TUI involved.
- A change-detection rule that never rewrites a note the user did not touch,
  including notes with tabs, CRLF endings or a missing trailing newline.

**Non-Goals:**
- Replacing or wrapping the textarea's renderer (what syntax colouring would
  need).
- Any editor state that survives leaving edit mode.

## Decisions

### Package and messages

```
internal/ui/editor      Model, New(store, note) (Model, error)
                        -> DoneMsg{}
internal/ui/app         routes NoteCreatedMsg / EditMsg into editor,
                        handles DoneMsg with the existing "back to current
                        note" path; placeholder.go is deleted
internal/notes          Write(name, content string) error
                        Normalize(content string) string
internal/ui/theme       Editor textarea styles
```

`New` reads the file up front and returns the error, mirroring preview. On
error the root falls back to the finder with the existing "cannot read"
message and clears the current note.

`DoneMsg` carries nothing: the file is already written and the root re-enters
preview of `m.current`, which re-reads it. The root's existing `noteBackMsg`
handling is renamed to the editor's message; its "no current note" branch
stays for safety but is unreachable in practice.

Root table additions:

| Event                 | current | mode                               |
|-----------------------|---------|------------------------------------|
| finder.NoteCreatedMsg | set     | editor (empty note)                |
| preview.EditMsg       | kept    | editor                             |
| editor.DoneMsg        | kept    | preview of current (re-read)       |
| New returns error     | nil     | finder, error "cannot read <name>" |

Alternative considered: `DoneMsg{Content}` so preview can render without a
second read. Rejected; preview's contract is to re-read on entry, and
handing content around would create two sources of truth.

### Normalization and change detection

`notes.Normalize(s)` returns `strings.TrimRight(s, "\n") + "\n"`. This is
the on-disk form of every note clipt writes: exactly one trailing newline.

On entry the editor:

1. reads the raw file content;
2. replaces `\r\n` with `\n` before handing it to the textarea, so the
   sanitizer does not turn each CRLF into two line breaks;
3. drops one trailing `\n` (the one the on-disk form guarantees), otherwise
   the textarea shows an extra empty last line and puts the cursor there
   instead of at the end of the last line of text;
4. calls `SetValue` and then records `baseline = Normalize(ta.Value())`.

`baseline` is what the textarea holds after the sanitizer ran, so tabs are
already four spaces in it. On Esc the editor computes
`now = Normalize(ta.Value())` and writes only when `now != baseline`. A
visit with no effective edit therefore never touches the file, even when the
file has tabs, CRLF endings, no trailing newline or several trailing newlines;
the file is rewritten in normalized form only once the user changes it.

Alternative considered: compare against the raw file bytes. Rejected; a note
with a tab or without a trailing newline would then count as changed on
every visit and get rewritten, bumping its mtime and finder position.

Alternative considered: a dirty flag set on every keypress. Rejected; typing
`x` and deleting it would count as a change, and the baseline comparison is
just as cheap.

### Textarea configuration

```
ta := textarea.New()
ta.ShowLineNumbers = true
ta.Prompt = ""          // line numbers are the only gutter
ta.CharLimit = 0        // already the default, set explicitly
ta.MaxHeight = 0        // lift the 99-row clamp for tall terminals
ta.MaxWidth = 0         // lift the 500-column clamp
ta.KeyMap = textarea.DefaultKeyMap
ta.SetWidth(width); ta.SetHeight(height - 1)
ta.Focus()
```

Width and height are set on `tea.WindowSizeMsg` and on construction with the
same defaults preview uses (80x24) so tests without a size message work.
`SetWidth` must run after `Prompt` and `ShowLineNumbers` are set because it
reserves gutter width from them.

Styles come from the theme package: `theme.EditorStyles()` returns a
`textarea.Style` pair (focused, blurred) built from the palette: line numbers
gray, the cursor line's number yellow, text in the normal foreground, no
cursor line background, end-of-buffer blank. The editor applies them with
`ta.FocusedStyle`/`ta.BlurredStyle`. Keeping the constructor in the theme
package preserves the "no color literal outside theme" rule; importing
`textarea` there is acceptable since it already imports Glamour for the
markdown style.

### Editor update loop

```
            esc, changed      store.Write ok -> DoneMsg
view -----> esc, unchanged -> DoneMsg
            esc, write fails -> err shown, stay (content kept)
            ctrl+c           -> tea.Quit
            anything else    -> textarea.Update
```

Esc and `ctrl+c` are intercepted in the editor's `Update` before the message
reaches the textarea; neither is in its key map, so nothing is lost. `err` is
a single string rendered in place of the last content line, exactly as
preview does, and cleared on the next key press. The textarea's blink
command is returned from `Init` so the cursor blinks like the rename prompt.

### Store.Write

`os.WriteFile(filepath.Join(dir, name+".md"), []byte(content), 0o644)`.
Plain in-place write by decision; it creates the file when it was removed
during editing. `Write` does not call `Normalize` itself so the store stays a
thin file layer; the editor passes normalized content.

### Status bar

Same layout code as preview: hints left in `theme.Hint`, name right-aligned
with `PlaceHorizontal`, name truncated from the left with an ellipsis when
the width is short. Hint text: `esc save & back`. The helper is duplicated
rather than shared for now; a shared status-bar helper is a small refactor
for a later change if a third mode needs it.

### Testing

- `notes`: `Write` creates and overwrites on `t.TempDir()`, fails on a
  read-only dir; `Normalize` table test (no newline, one, several, empty).
- `editor`: `Update` with synthetic `tea.KeyMsg`s on a temp dir. Asserts:
  typing then Esc writes normalized content and emits `DoneMsg`; Esc without
  edits leaves the file bytes and mtime untouched (set an old mtime with
  `os.Chtimes` first); type-then-delete does not write; CRLF file shows two
  lines; tab file untouched without edits and rewritten with spaces after
  one; write failure on a read-only dir keeps the content and sets `err`;
  file removed mid-edit is recreated; `ctrl+c` returns `tea.Quit`; the View
  of a long note contains its last line on entry.
- `app`: create-from-finder lands in the editor and Esc lands in preview
  showing the typed content; `e` then Esc returns to preview; the
  placeholder tests are removed.

## Risks / Trade-offs

- [Viewport may not follow the cursor to the end on entry] -> The editor
  test asserts the last line is in the View. The textarea only repositions
  in `Update`, and its viewport has no lines until the first `View`, so the
  editor renders once and then sends a no-op `Update` after every resize.
- [Tabs are rewritten as spaces on the first real edit] -> Accepted and
  specified; a note with no edits is never rewritten.
- [Control characters and invalid UTF-8 are dropped by the sanitizer] ->
  Same rule: only written after an edit. Notes are plain markdown, accepted.
- [Line number gutter is fixed at three digits] -> Notes over 999 lines get a
  misaligned gutter. Accepted for v1.
- [Bubbles textarea caps the buffer at 10,000 lines (`maxLines`, a private
  constant not lifted by `MaxHeight = 0`)] -> A note with more lines is
  truncated on load, and an edit would then save the truncated note. Open
  question: guard against it (refuse to edit such a note) or accept it for
  v1; not handled by this change.
- [`ctrl+v` paste uses a system clipboard library that may be unavailable in
  some terminals] -> Terminal bracketed paste still arrives as key runes and
  is inserted; `ctrl+v` failing is a no-op.
- [`ctrl+c` discards unsaved edits] -> By decision; it is the emergency exit.
  Esc is the only normal way out and always saves.
- [Last writer wins on external changes] -> Accepted non-goal.
