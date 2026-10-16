## Why

Three everyday frictions in edit mode: the line-number gutter changes width
at line 10 and line 100, so the text column shifts and indented blocks (such
as a template's YAML `form`) look misaligned while being correct; Tab inserts
nothing, so there is no way to indent without typing spaces; and once a note
is open there is no visible way to quit, because Esc bounces between preview
and finder and the only exit, `ctrl+c`, is unhinted and discards unsaved
edits.

## What Changes

- Edit mode renders its own line-number gutter with a width derived from the
  note's line count, so every row's text starts in the same column, including
  soft-wrapped continuation rows. The text area's built-in numbers are no
  longer used.
- Tab inserts two spaces at the cursor; Shift+Tab removes up to two leading
  spaces from the current line.
- `ctrl+q` quits clipt from every screen and is shown in every screen's key
  hints. In edit mode it saves the note first exactly as Esc does, and on a
  write failure stays in edit mode with the error instead of quitting. From
  the template picker or form it abandons the insertion, saves the note and
  quits. `ctrl+c` keeps its current meaning: quit immediately without saving.

Non-goals:

- Migrating to the Charm v2 libraries (separate change).
- Auto-indenting new lines on Enter.
- A configurable indent width.
- Changing how tab characters read from disk are expanded (stays four spaces).
- Showing a line number in preview or any vim-style navigation.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `editor`: the gutter width is fixed per note and text aligns across rows;
  Tab and Shift+Tab indent and dedent; `ctrl+q` saves and quits; the status
  bar hints `ctrl+q`.
- `finder`: `ctrl+q` quits and is hinted in both scopes.
- `preview`: `ctrl+q` quits from the view, rename and delete states and is
  hinted.
- `template-insert`: `ctrl+q` from the picker or form abandons the insertion,
  saves the note and quits; the status bar hints it.
- `notes-dir-selection`: `ctrl+q` quits from any picker state and is hinted.

## Impact

- `internal/ui/editor`: gutter rendering (a copy of the text area's wrap
  function plus a per-row prompt callback), Tab/Shift+Tab handling, a
  save-then-quit path, hint text.
- `internal/ui/finder`, `internal/ui/preview`, `internal/ui/tmpl`,
  `internal/ui/picker`: `ctrl+q` handling and hint text.
- `internal/ui/app`: routes the quit request coming out of the template
  insert mode through the editor's save.
- `README.md`: key tables gain `ctrl+q` and Tab/Shift+Tab.
- No new dependencies. Bubbles stays at v1.0.0.
