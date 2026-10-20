## Context

See proposal.md for motivation. The edit mode wraps the Bubbles v2 text
area, which already implements keyboard selection: anchor and head
positions, Shift+arrow and Ctrl+Shift+arrow bindings, select all, a
selection style, and "replace the selection" semantics for typing, Enter,
paste and the delete keys. clipt switches all of that off in
`internal/ui/editor/editor.go` when the text area is built, and the theme
leaves the selection style unset for the same reason.

The text area's default key map is the one clipt uses, with a single
override: `ctrl+t` (transpose) is disabled because it opens the template
picker. Copy errors do not reach the editor as a message (the text area's
copy message type is unexported), but the text area records them in its
exported `Err` field after handling the message.

## Goals / Non-Goals

**Goals:**

- Turn the existing selection on with the text area's default keys.
- Keep the editor model in charge of everything the status bar shows.

**Non-Goals:**

- Cut, mouse selection, undo, highlighting: see the proposal's non-goals.
- Any change to the text area's own cursor-movement or delete bindings.

## Decisions

**Re-enable the text area's bindings rather than re-implement selection.**
The selection, its rendering across soft-wrapped lines and the
replace-on-type behaviour all exist upstream. Removing the eight lines that
disable them is the whole change for the selection itself. Alternative: a
third-party editor component; rejected in exploration because none offers
Shift selection without a vim-style modal engine or a foreign framework.

**All selection keys are the text area's defaults, copy included.**
Shift+arrows, Ctrl+Shift+arrows (and the Alt+Shift variants the text area
also accepts), `ctrl+g` for select all and `ctrl+shift+c` for copy. The
text area runs the copy itself through its clipboard library; the editor
only has to stop disabling the bindings. Alternative considered: an
editor-owned copy on a key such as `ctrl+y`, because `ctrl+shift+c` is a
terminal chord and degrades to `ctrl+c` on terminals without enhanced key
reporting; rejected in favour of one key map with a single documented
override (`ctrl+t`), which is easier to learn and to explain.

**Copy errors are read from the text area, not routed as messages.** After
forwarding a key to the text area, the editor copies a non-nil `Err` from
it into its own status-line error and clears the text area's field, so copy
failures appear exactly where write failures do and vanish on the next key.

**The copy hint is contextual.** The hints line already lists three keys;
`ctrl+shift+c copy` is appended only while a selection exists, where it
also serves as the only indication that a selection is active besides the
highlight.

**Selection background comes from the gruvbox palette.** The theme gains a
`bg2` (`#504945`) constant used as the text area's selection background;
text keeps the normal foreground.

## Risks / Trade-offs

- [Copy fails on Linux without a clipboard utility] → the error names the
  missing tools (the clipboard library's message does); the selection and
  the note are untouched.
- [Ctrl+Shift+arrows may not reach the app in some terminals] → character
  and line selection with plain Shift still work everywhere; word selection
  is a convenience, not the only path.
- [`ctrl+shift+c` is intercepted by the terminal's own copy, or arrives as
  `ctrl+c` on terminals without enhanced key reporting, which quits without
  saving] → accepted for a single default key map; the hint shows the key
  while a selection is active, and the terminal's own copy still works on
  mouse selections. Revisit with its own change if it bites in practice.
