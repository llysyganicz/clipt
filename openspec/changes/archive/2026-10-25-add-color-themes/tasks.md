## 1. Palette and theme file format

- [x] 1.1 Add `theme.Palette` (background, foreground, selection, the sixteen ANSI slots) and `theme.Parse(r io.Reader, name string) (Palette, error)` that decodes the flat TOML, requires every key, validates `#rrggbb`, defaults `selection` to `bright_black` and ignores unknown keys; verify with table tests covering a complete file, omitted selection, an unknown key, a missing key (error names the key), a bad color (error names the key) and invalid TOML
- [x] 1.2 Create `internal/ui/theme/builtin/gruvbox_dark.toml` and `gruvbox_light.toml` with the full gruvbox palettes and explicit `selection`; verify a test parses both and asserts `gruvbox_dark` yields exactly the previous constants (`#282828`, `#504945`, `#ebdbb2`, `#fabd2f`, `#8ec07c`, `#fb4934`, `#928374`)
- [x] 1.3 Embed `builtin/*.toml` in the theme package and add `theme.Builtin()` listing names; verify a test asserts the list is exactly `gruvbox_dark`, `gruvbox_light`

## 2. Applying the palette to styles

- [x] 2.1 Add `theme.Apply(Palette)` that rebuilds `Normal`, `Selected`, `Create`, `Hint`, `Error`, `Prompt`, `Tag`, `RowTag`, `LineNumber`, `CursorLineNumber`, the cursor color and `Markdown` from the palette roles in the theme spec, and make the package self-initialize with the parsed built-in `gruvbox_dark`; verify a test renders every style before and after `Apply(gruvboxDark)` and gets identical output, and after `Apply(gruvboxLight)` gets the light colors
- [x] 2.2 Rewrite `markdownStyle(p Palette)` so every `Color`/`BackgroundColor` the dark Glamour config defines, plus headings, table headers, links, code, block quotes, rules, image markers, code block borders, list markers and task markers, comes from the palette with no heading or inline code background; verify a reflection test walks the resulting `StyleConfig` and asserts each non-nil color is one of the palette's hex strings, and `TestMarkdownRenders` still passes
- [x] 2.3 Keep `TestNoColorLiteralsOutsideTheme` passing with the `builtin/` folder (non-Go files are already skipped) and extend it to also reject hex literals in Go files under `theme/` outside the test file, since the palette now comes from TOML; verify `go test ./internal/ui/theme/`

## 3. Theme lookup

- [x] 3.1 Add `config.Dir()` returning the `clipt` config directory (parent of `config.Path()`) and `theme.Load(name, dir string) (Palette, error)` that reads `<dir>/themes/<name>.toml` when present, falls back to the embedded file, and otherwise returns an error naming the theme and the looked-up path; verify tests in a `t.TempDir()` cover user file, built-in fallback, user file shadowing `gruvbox_dark`, and an unknown name whose error contains both the name and the path
- [x] 3.2 Make `theme.Load` wrap parse errors from both sources with the file path (embedded files report `builtin/<name>.toml`); verify a test in a temp dir with a `red = "tomato"` theme gets an error containing the path and `red`

## 4. Config carries the theme

- [x] 4.1 Change `config.Load` to return a `Config{NotesDir, Theme string}` (plus status and error) and add `Theme string \`toml:"theme,omitempty"\`` to the file struct; verify tests read a file with and without `theme` and the unknown-key test now uses `editor = "vim"`
- [x] 4.2 Change `config.Save(cfg Config)` to write `notes_dir` and, when non-empty, `theme`; verify temp-dir tests show a `theme` key survives a rewrite and a file saved without a theme has no `theme` line
- [x] 4.3 Thread the theme name from `main` through `app.New`/`app.NewPicker` into the picker so `config.Save` receives it on directory changes; verify an app test drives ctrl+o and a new directory and asserts the written config (in a temp `XDG_CONFIG_HOME`) still contains the theme

## 5. Startup wiring

- [x] 5.1 In `cmd/clipt/run()`, after a successful config load (any status), resolve the theme name (empty means `gruvbox_dark`), call `theme.Load` then `theme.Apply`, and return load errors so `main` prints them and exits 1; verify `go build ./...` and a manual run with `theme = "nope"` prints an error naming `nope` and the themes path, while `theme = "gruvbox_light"` starts in light colors
- [x] 5.2 Confirm the picker path (first run, bad dir) also applies the theme before the program starts; verify by running with an empty config dir and `theme = "gruvbox_light"` set in a config whose `notes_dir` is missing

## 6. Release and docs

- [x] 6.1 Add `src: internal/ui/theme/builtin/*.toml`, `dst: themes` entries to the `.goreleaser.yaml` archive `files`; verify `go run github.com/goreleaser/goreleaser/v2@v2.18.3 release --snapshot --clean` produces archives containing `themes/gruvbox_dark.toml` and `themes/gruvbox_light.toml` next to `README.md`
- [x] 6.2 Add a "Themes" section to `README.md`: the `theme` key, lookup order, the two built-in names, the full key list with the gruvbox dark file as the example, the `selection` default, and error behavior; update the config section's key table; verify by reading the rendered README
- [x] 6.3 Update `openspec/config.yaml` project context: replace the "config override is a later change" theme line with the themes summary (key, lookup, format, built-ins); verify `openspec validate add-color-themes` still passes

## 7. Verification

- [x] 7.1 Run `gofmt -l .`, `go vet ./...` and `go test ./...` and confirm all clean
- [x] 7.2 Manually compare the default appearance before and after the change in finder, preview with a markdown note, edit mode with a selection, and the template form, confirming no visible difference under `gruvbox_dark`
