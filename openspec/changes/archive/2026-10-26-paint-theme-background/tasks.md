## 1. Theme package

- [x] 1.1 Add `theme.Background()` and `theme.Foreground()` returning the current palette's colors as `color.Color` (built from `lipgloss.Color`), refreshed by `Apply`; verify a theme test applies `gruvbox_light` and asserts both accessors equal the light palette's `background` and `foreground`
- [x] 1.2 Make `Normal` a foreground-only style; verify a theme test renders `Normal` under both built-in themes and asserts the output contains the foreground color but no background sequence, and `TestNoColorLiteralsOutsideTheme` still passes
- [x] 1.3 Set the typed-text style in `InputStyles` (both focus states) to the palette foreground; verify a theme test asserts `InputStyles().Focused.Text` and `.Blurred.Text` render text in the current foreground after `Apply(gruvboxLight)`

## 2. Root view paints the screen

- [x] 2.1 In `app.Model.View`, set `BackgroundColor` and `ForegroundColor` on the returned `tea.View` from the theme accessors, next to `AltScreen`; verify an app test applies `gruvbox_light`, drives the model through picker, finder, preview, edit and template insert, and asserts both colors on `View()` equal the light palette in every mode (same shape as the existing `AltScreen` checks)
- [x] 2.2 Confirm the finder and template picker rows no longer carry a background without code changes in those packages; verify finder and tmpl tests that render rows assert no background sequence appears in an unselected row under `gruvbox_light`

## 3. Docs and context

- [x] 3.1 Add to the README's "Themes" section that clipt paints the terminal background and foreground from the theme while it runs and restores them on exit, and that a terminal or multiplexer ignoring the request shows the palette colors on its own background; verify by reading the rendered README
- [x] 3.2 Update the themes paragraph in `openspec/config.yaml` project context to say the root view paints the screen from the palette and no element except the editor selection sets a background; verify `openspec validate --changes paint-theme-background` passes

## 4. Verification

- [x] 4.1 Run `gofmt -l .`, `go vet ./...` and `go test ./...` and confirm all clean
- [x] 4.2 Run the built binary in a dark terminal with `theme = "gruvbox_light"` and check the finder, the ctrl+o directory picker, preview, edit with a selection and the template form are drawn on the cream background with no row boxes, then quit and confirm the terminal's own colors are back; repeat with `gruvbox_dark` and confirm no visible difference from before the change
