## Context

See proposal.md - Why. Technical facts that shape the approach:

- Edit mode wraps the Bubbles v1.0.0 `textarea`. Its built-in line numbers
  are padded to the digit count of `MaxHeight`, and the editor sets
  `MaxHeight = 0` to lift the 99-line cap, so the gutter is 3 columns wide
  for lines 1-9, 4 for 10-99 and 5 from 100. `SetWidth` always reserves 4
  columns for numbers, so from line 100 the text also overflows the
  viewport by one column and is clipped.
- The text area offers `SetPromptFunc(width, fn)`: a per-row callback that
  receives the absolute visual row index (counting soft-wrapped rows, before
  viewport scrolling) and whose width is reserved exactly. It is called with
  the row index only, so it cannot tell which rows start a logical line.
- The text area's `wrap(runes, width)` is unexported and pure. Bubbles v2
  carries the identical function, and v2 keeps the prompt callback.
- Tab arrives as a `KeyTab` message with no runes; the text area's fallback
  inserts the runes, so nothing happens. A literal tab character would be
  expanded to four spaces by the text area's input sanitizer.
- `ctrl+c` returns `tea.Quit` from every model. Each mode is its own Bubble
  Tea model and the root model in `internal/ui/app` routes between them;
  the editor keeps the note buffer while the template picker and form are
  shown.

## Goals / Non-Goals

**Goals:**

- A gutter whose width depends only on the note's line count, with text in
  the same column on every row, for any number of lines.
- Tab and Shift+Tab as two-space indent and dedent inside the editor model.
- One quit key, `ctrl+q`, that never loses an edit, handled where each mode
  already handles `ctrl+c`.

**Non-Goals:**

- Touching the text area library (no fork, no vendoring) or upgrading it.
- Changing the on-disk tab expansion (stays four spaces).
- Any change to the finder, preview or picker beyond the key and its hint.

## Decisions

### Own gutter through the prompt callback

The editor turns `ShowLineNumbers` off and installs a prompt callback of
width `digits(lineCount) + 2`. After every update that may change the value
or the width, the editor rebuilds a `[]string` of gutter labels, one per
visual row: the line number for the first row of each logical line, blanks
for continuation rows. The callback indexes that slice. The cursor line's
number uses the existing yellow style; the callback returns an already
styled string, which the text area passes through its (empty) prompt style.

To know which rows start a logical line, the editor package carries a copy
of the text area's `wrap` function (about sixty lines). Rows are computed by
wrapping every logical line at the text area's current width, exactly as the
text area does for rendering.

When the digit count changes, the editor calls `SetPromptFunc` with the new
width and then `SetWidth` again so the text area re-reserves the columns
and re-wraps.

Alternatives considered:

- Keep the built-in numbers and set `MaxHeight` to a large positive value:
  caps the note at that many lines, and any width past two digits overflows
  the fixed 4-column reservation and gets clipped.
- Keep the built-in numbers and set `MaxHeight` to a negative number such as
  `-9`: two digits and no cap, so lines 1-99 align, but the shift and the
  clipping return at line 100. Relies on an accident of the implementation.
- Render the gutter outside the text area by prefixing its `View` output:
  needs the viewport's scroll offset, which the text area does not expose.
- Drop line numbers: the user relies on them for navigation and for talking
  about notes with others.
- Upgrade to Bubbles v2: same coupling of digits to `MaxHeight`, and a
  migration of every UI package. Planned as a separate change; this design
  carries over because v2 keeps the prompt callback and the same `wrap`.

### Gutter labels are rebuilt, not memoized

Labels are rebuilt after each key press and each resize by wrapping the whole
value. Wrapping is linear in the note's size and the text area itself
re-wraps on render, so this is at most a constant-factor cost. Memoization
is not worth its complexity until a note proves too large; see Risks.

### Tab and Shift+Tab are editor-level keys

The editor intercepts `KeyTab` and `KeyShiftTab` before the text area, like
Esc and `ctrl+t`. Tab calls the text area's `InsertString("  ")`, so the
insertion follows the same path as typing and counts for auto-save.
Shift+Tab reads the cursor's line from `Value()`, removes up to two leading
spaces, puts the value back and restores the cursor row and column (minus
the removed count, floored at zero), reusing the reposition logic that the
resize path already has. The text area exposes no line-level setter, so the
whole value is replaced; the gutter is rebuilt afterwards like after any
edit.

Alternative considered: binding Tab inside the text area's key map. The key
map only maps to the text area's own actions, none of which inserts a fixed
string.

### ctrl+q is a per-model key, with save in the editor

Each model handles `ctrl+q` next to its `ctrl+c`:

- finder, preview, notes-dir picker: return `tea.Quit` directly. They hold
  nothing unsaved.
- editor: run the same save as Esc. On success return `tea.Quit` instead of
  `DoneMsg`; on failure stay open with the error, exactly as Esc does.
- template picker/form: emit a `QuitMsg`. The root model switches back to
  edit mode, discarding the insertion, and asks the editor to save-and-quit.
  The root can only do this because the editor model is still held while the
  template mode is shown.

Alternative considered: handling `ctrl+q` once in the root model. The root
would need to know which modes hold unsaved state and to reach into the
editor's save, which is the editor's business; and it would break the
convention that each mode owns its keys, including `ctrl+c`.

`ctrl+c` keeps its meaning (quit without writing) so an emergency exit that
never touches disk still exists, as in every Bubble Tea app.

### Hints

`ctrl+q quit` is appended to every hint string. The editor's hint line gets
`ctrl+q save & quit`. The hint strings are plain constants today and stay
that way.

## Risks / Trade-offs

- [Copied `wrap` drifts from the library's] → Bubbles v1 is frozen, and a
  test renders the text area with a wrapped line and asserts the editor's
  row count matches the number of rows the text area produced. v2 ships the
  identical function, so the migration keeps the test green.
- [Styled string returned from the prompt callback is padded by width] → The
  text area pads by the string's display width, which ignores ANSI escapes,
  so a styled label of the right visible width is not padded further. The
  gutter test checks the rendered column of the text, not the label.
- [Rebuilding labels on every key press is slow on huge notes] → Linear in
  note size; a 5,000-line note is a few hundred microseconds. If a real note
  proves slow, memoize per logical line keyed by content and width.
- [Shift+Tab replaces the whole value, which resets the scroll offset] → The
  cursor is restored and the view is scrolled to it, the same way the
  resize path already recovers; a test covers a dedent on a line below the
  first screen.
- [`ctrl+q` may be swallowed by terminals with software flow control] →
  Bubble Tea puts the terminal in raw mode, which disables flow control. If
  a user's terminal still eats it, `ctrl+c` remains as the exit.
- [Quit from the template form silently drops typed form values] → The form
  values were never going to reach disk without Submit; the note's own edits
  are what is saved. The hint says `quit`, not `submit`.

## Migration Plan

No data or config changes. Ship in the next release; the README key tables
are updated in the same change.
