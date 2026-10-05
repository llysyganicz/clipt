## Why

Preview and the finder are done, but both transitions into edit (`e` in
preview, Enter on the finder's create row) still land on a placeholder that
only shows the note name. Without an editor clipt cannot write a note, so this
is the last slice needed for a usable v1.

## What Changes

- Add the edit mode: the current note's content in a full-height text area
  with line numbers above a one-line status bar, using the same layout as
  preview (key hints left, note name right). Editing is insert-only; there is
  no vim normal mode.
- Keys: Esc leaves edit and returns to preview, `ctrl+c` quits. Every other
  key goes to the text area with its default key bindings. There is no limit
  on the length of a note.
- Auto-save: on Esc the note is written back to `<name>.md` only when its
  content differs from what was loaded, so a visit without edits leaves the
  file and its modification time untouched. There is no explicit save command.
- The saved file always ends with exactly one newline. Content counts as
  unchanged when it matches the loaded content apart from that final newline.
- When the write fails the editor stays open with the error and the content
  preserved; Esc retries.
- The cursor starts at the end of the document on every entry.
- Resizing the terminal resizes the text area to fill the space above the
  status bar.
- Remove the edit placeholder from `internal/ui/app`; the root routes both
  entry points to the new editor model.

## Capabilities

### New Capabilities

- `editor`: the edit mode's content area and status bar, the key map, the
  cursor start position, the Esc/auto-save transition back to preview, the
  unchanged-content rule, the trailing-newline rule, write-failure behaviour
  and resize.

### Modified Capabilities

None. The finder already creates the empty file before entering edit, and
preview already re-reads the note from disk on entry, so both existing specs
hold as written.

## Non-goals

- Syntax colouring of markdown while editing. Bubbles textarea renders its
  whole buffer in one style, so colouring needs a custom renderer; it is a
  candidate for a later change once the plain editor is in daily use.
- Undo and redo.
- Markdown-aware editing such as list continuation or heading helpers.
- Detecting that the file changed on disk while it is being edited; the last
  writer wins.
- Remembering the cursor position between visits.
- Saving on `ctrl+c`. It is the emergency exit and quits immediately,
  discarding unsaved edits, as in every other mode.
- Deleting a freshly created note that is left empty; it stays as an empty
  note and can be deleted from preview.
- Atomic writes through a temporary file; the note is written in place.
- Vim normal mode and any modal editing.

## Impact

- New package `internal/ui/editor`; `internal/ui/app` routes to it and drops
  `placeholder.go`.
- `internal/notes` gains a write operation with temp-dir tests.
- `internal/ui/theme` gains the styles the text area needs (line numbers,
  cursor line) built from the existing palette.
- New dependency: Bubbles `textarea` (already in the Bubbles module).
- Writes stay inside the notes directory: the editor only ever writes the
  current note's `<name>.md`.
