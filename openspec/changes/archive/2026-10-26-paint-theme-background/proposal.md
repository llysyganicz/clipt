## Why

A theme only colors the text: clipt never sets the terminal background, so
the palette's `background` shows up solely as a box behind each unselected
finder row. Under `gruvbox_dark` on a dark terminal that box was invisible;
under `gruvbox_light` on a dark terminal it is a cream rectangle around every
row and the rest of the screen stays dark. A theme should decide how the whole
screen looks, not only the ink.

## What Changes

- The root view sets the terminal background and foreground from the theme
  palette for as long as clipt runs, so every mode (picker, finder, preview,
  edit, template insert) is drawn on the palette's `background` in its
  `foreground`. The terminal's own colors return when clipt exits.
- No interface element paints a background of its own any more; the
  finder's and template picker's unselected rows lose the per-row box. The
  editor selection keeps the `selection` background, which is its purpose.
- Text typed into the finder query and the directory path input uses the
  palette `foreground` instead of the terminal's default text color.
- A terminal that ignores the background request shows the palette colors
  on its own background, which is what happens today minus the row boxes.

Non-goals: detecting whether the terminal is light or dark, picking a theme
automatically, per-element background colors, any change to the theme file
format. The format keeps the full sixteen-color palette, required keys
included, even though only some of them are drawn today; the rest are
reserved for later features.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `theme`: the "Palette roles in the interface" requirement changes from
  "the finder's normal rows on `background`" to the whole screen on
  `background`, with ordinary text in `foreground`, and no element except
  the editor selection setting a background.

## Impact

- `internal/ui/app`: the root `View` sets the view's background and
  foreground colors from the theme.
- `internal/ui/theme`: `Normal` becomes a foreground-only style; the package
  exposes the palette background and foreground as colors; `InputStyles`
  styles typed text with the foreground.
- `internal/ui/finder` and `internal/ui/tmpl`: no code change intended; their
  rows simply stop carrying a background through `Normal`.
- Tests in `internal/ui/app` and `internal/ui/theme` cover the view colors
  and the styles. No new dependencies: Bubble Tea v2 already supports
  setting the terminal colors from a view and resets them on exit.
