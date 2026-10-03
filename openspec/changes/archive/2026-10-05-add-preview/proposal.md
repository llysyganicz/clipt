## Why

The finder can open and create notes, but every transition out of it lands on
a placeholder that only shows the note name. Preview is the screen users
actually read notes on, and it is where rename and delete live, so it is the
next slice that makes clipt usable day to day.

## What Changes

- Add the preview mode: the current note's markdown is rendered with Glamour
  and shown in a scrolling viewport that fills the terminal above a one-line
  status bar. The status bar shows the key hints on the left and the note name
  on the right. There is no title bar.
- Keys: `e` enters edit, `r` starts a rename, `d` asks for delete
  confirmation, Esc returns to the finder, `ctrl+c` quits. Up/Down scroll one
  line, PgUp/PgDn one page.
- Rename: a prompt prefilled with the current name replaces the status bar.
  Enter renames the file and updates the current note; the same name or Esc
  cancels. An invalid name or an existing target is refused with an error and
  no file change.
- Delete: a `y/n` confirmation replaces the status bar. `y` removes the file,
  clears the current note and returns to the finder; `n` or Esc cancels.
- Entering preview re-reads the note from disk. If the read fails, the user is
  sent back to the finder with an error and the current note is cleared.
- Returning to the finder from preview (Esc or after delete) enters it with an
  empty query.
- Resizing the terminal re-renders the markdown at the new width. The scroll
  position starts at the top on every entry.
- Add a gruvbox Glamour style to the theme package so rendered markdown
  matches the rest of the UI.
- Replace the preview half of the placeholder in `internal/ui/app`; the
  placeholder shrinks to stand in for edit only.

## Capabilities

### New Capabilities

- `preview`: rendering and scrolling of the current note, the status bar, the
  key map, the rename and delete flows, the Esc/edit transitions and the
  missing-file behaviour.

### Modified Capabilities

None. The finder already re-reads its listing on entry, so renames and
deletes show up without a finder requirement change.

## Non-goals

- Edit mode. `e` transitions to a stub until the edit change lands.
- Remembering the scroll position between visits, or the finder query across
  a round trip through preview.
- Any editing of note content from preview, including changing the title or
  first heading.
- Case-only renames on case-insensitive file systems (`todo` to `Todo`) are
  refused as "already exists". Accepted for v1.
- Watching the file for external changes while preview is open.
- A user-configurable Glamour style; the gruvbox style is built in. If it
  proves too complex, it is dropped from this change and Glamour's built-in
  dark style ships instead.

## Impact

- New package `internal/ui/preview`; `internal/ui/app` routes to it and
  handles rename and delete messages on the current note.
- `internal/notes` gains read, rename and delete operations with temp-dir
  tests.
- `internal/ui/theme` gains a Glamour style built from the existing palette.
- New dependencies: `charmbracelet/glamour` and Bubbles `viewport` (already
  in the Bubbles module).
- Writes stay inside the notes directory: rename and delete touch only
  `<name>.md` files there.
