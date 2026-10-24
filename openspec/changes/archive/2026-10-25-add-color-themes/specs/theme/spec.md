## Purpose

Lets the user replace the built-in gruvbox dark palette with another terminal
color scheme, chosen by name in the config file and defined in a small TOML
theme file, so clipt matches the terminal it runs in.

## ADDED Requirements

### Requirement: Theme selection
The optional `theme` key in the config file SHALL name the theme to use. When
the key is absent or empty, the theme SHALL be `gruvbox_dark`, which renders
exactly as clipt did before themes existed.

#### Scenario: No theme configured
- **WHEN** the config file has no `theme` key
- **THEN** clipt starts with the `gruvbox_dark` theme

#### Scenario: Theme configured
- **WHEN** the config file contains `theme = "gruvbox_light"`
- **THEN** clipt starts with the `gruvbox_light` theme

### Requirement: Theme lookup
A theme named `<name>` SHALL be read from `<config dir>/clipt/themes/<name>.toml`
when that file exists, where `<config dir>` is the same platform config
directory that holds `config.toml`. Otherwise it SHALL be read from the
themes built into the binary. A user file SHALL take precedence over a
built-in theme of the same name. When neither exists, clipt SHALL print an
error naming the theme and the path it looked for, and exit with a non-zero
status.

#### Scenario: User theme file
- **WHEN** `theme = "solarized"` and `<config dir>/clipt/themes/solarized.toml`
  is a valid theme file
- **THEN** clipt starts with the palette from that file

#### Scenario: Built-in theme
- **WHEN** `theme = "gruvbox_light"` and no `gruvbox_light.toml` exists in the
  themes directory
- **THEN** clipt starts with the built-in `gruvbox_light` palette

#### Scenario: User file shadows a built-in theme
- **WHEN** `theme = "gruvbox_dark"` and `<config dir>/clipt/themes/gruvbox_dark.toml`
  exists with a different `background`
- **THEN** clipt uses the background from the user's file

#### Scenario: Unknown theme
- **WHEN** `theme = "nope"` and neither a user file nor a built-in theme has
  that name
- **THEN** clipt prints an error that names `nope` and the path
  `<config dir>/clipt/themes/nope.toml`, and exits with a non-zero status

### Requirement: Built-in themes
The binary SHALL contain exactly two themes: `gruvbox_dark` and
`gruvbox_light`. Both SHALL be complete theme files in the same format a
user would write, and the repository SHALL hold them as the single source
for the binary and for release archives.

#### Scenario: Built-in names
- **WHEN** the built-in themes are listed
- **THEN** they are `gruvbox_dark` and `gruvbox_light` and no others

### Requirement: Theme file format
A theme file SHALL be TOML with the string keys `background`, `foreground`,
`black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`,
`bright_black`, `bright_red`, `bright_green`, `bright_yellow`,
`bright_blue`, `bright_magenta`, `bright_cyan`, `bright_white`, each holding
a color written as `#` followed by six hexadecimal digits. The key
`selection` (a color in the same form) SHALL be optional and SHALL default
to the value of `bright_black`. Unknown keys SHALL be ignored.

#### Scenario: Complete theme file
- **WHEN** a theme file sets every required key to a valid color and
  `selection = "#504945"`
- **THEN** the theme loads and its selection background is `#504945`

#### Scenario: Selection omitted
- **WHEN** a theme file sets every required key, with `bright_black = "#928374"`,
  and has no `selection` key
- **THEN** the theme loads and its selection background is `#928374`

#### Scenario: Unknown key ignored
- **WHEN** a theme file is complete and also contains `cursor = "#ffffff"`
- **THEN** the theme loads without error

### Requirement: Invalid theme file
A theme file that cannot be read, is not valid TOML, is missing a required
key, or holds a value that is not a `#rrggbb` color SHALL be reported as a
startup error naming the theme file path and, for a bad or missing key, the
key name. clipt SHALL then exit with a non-zero status without modifying
any file.

#### Scenario: Missing key
- **WHEN** the selected theme file has no `bright_yellow` key
- **THEN** clipt prints an error naming the file and `bright_yellow`, and
  exits with a non-zero status

#### Scenario: Bad color value
- **WHEN** the selected theme file contains `red = "tomato"`
- **THEN** clipt prints an error naming the file and `red`, and exits with a
  non-zero status

#### Scenario: Invalid TOML
- **WHEN** the selected theme file contains `red = ` (invalid TOML)
- **THEN** clipt prints an error naming the file and exits with a non-zero
  status, leaving the file unchanged

### Requirement: Palette roles in the interface
The interface SHALL draw every element from the theme palette and nothing
else, with these roles: ordinary text in `foreground`; the finder's normal
rows on `background`; the selected row, prompts, confirmation questions and
the cursor's line number in `bright_yellow`; the create row and tags in
`bright_cyan`; errors in `bright_red`; hints, dimmed row tags and the other
line numbers in `bright_black`; the editor selection on the `selection`
background; the text cursor in `foreground`.

#### Scenario: Light theme colors the finder
- **WHEN** clipt runs with `gruvbox_light` and the finder shows a selected
  row, a create row and a hint line
- **THEN** the selected row uses the light theme's `bright_yellow`, the
  create row its `bright_cyan` and the hint its `bright_black`

#### Scenario: No color outside the palette
- **WHEN** any screen is rendered under any theme
- **THEN** every color in the output is one of that theme's palette values

### Requirement: Palette roles in rendered markdown
Rendered markdown in preview SHALL take every color from the theme palette:
document text in `foreground`; headings and table headers in
`bright_yellow`; links, link text and inline or block code text in
`bright_cyan`; block quotes, horizontal rules, image markers and code block
borders in `bright_black`; list markers in `bright_blue`; completed task
markers in `bright_green`; emphasis and strong text keep `foreground` with
italic and bold. No heading or inline code SHALL have its own background.

#### Scenario: Markdown under a light theme
- **WHEN** a note with a heading, a list, a link and a code block is
  previewed under `gruvbox_light`
- **THEN** the heading is in that theme's `bright_yellow`, list markers in
  its `bright_blue`, the link and code in its `bright_cyan`, and no element
  uses a color outside the palette
