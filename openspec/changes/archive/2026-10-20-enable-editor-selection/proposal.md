## Why

Edit mode is insert-only: the text area's selection keys are switched off,
so replacing or removing a span of text means deleting it one character at a
time. Selecting with Shift plus the cursor keys is the one editing gesture
users of every other editor expect and miss here, and it makes the external
editor option (explored and dropped) unnecessary.

## What Changes

- Enable keyboard selection in edit mode: Shift+Left/Right select by
  character, Shift+Up/Down by line, Ctrl+Shift+Left/Right by word, and
  `ctrl+g` selects all. Plain cursor movement clears the selection.
- Typing, Enter, paste and template insertion replace the selection;
  Backspace and Delete remove it.
- `ctrl+shift+c`, the text area's default, copies the selection to the
  system clipboard. All selection keys are the text area's defaults; the
  only default clipt overrides is `ctrl+t`, reserved for templates.
- While a selection exists the status bar hints show the copy key.
- The selection is drawn with a visible background from the theme palette.
- The same keys work when editing a template.

Non-goals: cut, mouse selection, undo, syntax highlighting, an external
editor, and any change to the existing Esc, `ctrl+q`, `ctrl+c`, `ctrl+t`,
Tab and Shift+Tab keys.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `editor`: the "Editing keys" requirement gains selection; the "Status
  bar" requirement gains the contextual copy hint.

## Impact

- `internal/ui/editor`: stop disabling the selection and copy bindings,
  show the copy hint and the text area's copy errors.
- `internal/ui/theme`: a selection style for the text area.
- Dependencies: none added. The text area copies through its own clipboard
  library; on Linux that needs `xclip`, `xsel` or `wl-copy` installed,
  Windows and macOS need nothing.
- No change to saving, file layout, config or the other modes.
