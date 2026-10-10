## Context

See proposal.md for motivation. What is already in place and shapes the
approach:

- `internal/notes.Store` is a plain directory path with List, Exists,
  Create, Read, Write, Rename, Delete and ValidateName. `List` skips entries
  that start with `.` or do not end in `.md`, so a `.templates` directory is
  already invisible to the notes listing. `List` returns the `os.ReadDir`
  error for a missing directory.
- The finder (`internal/ui/finder`) is built from a store, an optional
  current note and a query; `BuildRows`/`Match` are pure functions over the
  listing. It emits `OpenNoteMsg`, `NoteCreatedMsg`, `ChangeDirMsg` and
  `BackMsg`; the root decides what they mean.
- The editor (`internal/ui/editor`) wraps Bubbles `textarea`. `InsertString`
  inserts at the cursor, splits on `\n`, runs the sanitizer, and leaves the
  cursor after the inserted text. The textarea only scrolls to the cursor in
  `Update`, which the editor already works around with a no-op message.
  `DefaultKeyMap` binds `ctrl+t` to `TransposeCharacterBackward`.
- The root (`internal/ui/app`) keeps one model per mode as a field and
  switches `mode`; the editor model survives in `m.edit` while another mode
  is shown, which is what template insertion needs.
- Preview re-reads the note from disk; nothing else holds note content.
- No YAML parser is in `go.mod`; `BurntSushi/toml` is (config only).

## Goals / Non-Goals

**Goals:**
- A `templates` package with no TUI dependency: parse, expand, locate, and
  the skeleton, all unit-tested with strings and temp dirs.
- Reuse `notes.Store` on the `.templates` directory for every file
  operation; no second file layer.
- The finder and editor learn about templates through one small flag each
  (scope, kind), not through parallel models.
- The insert mode knows nothing about the editor beyond the message it
  returns.

**Non-Goals:**
- A generic form component. The form is three input kinds in a list and is
  built directly from Bubbles `textinput`.
- Preserving the finder query across scope toggles.
- Sharing the status-bar helper between modes (still duplicated, as in
  preview and editor).

## Decisions

### Package layout and messages

```
internal/templates        Dir(notesDir) string           -> notesDir/.templates
                          Store(notesDir) notes.Store    -> Store{Dir: Dir(notesDir)}
                          EnsureDir(notesDir) error      -> os.MkdirAll
                          Parse(content) (Template, error)
                          (Template) Expand(values map[string]string) string
                          Skeleton string
                          Template{Fields []Field, Body string}
                          Field{Name, Label, Type string; Options []string}
                          (Field) Default() string

internal/ui/finder        Scope (ScopeNotes, ScopeTemplates)
                          New(store, current, query)        unchanged, notes
                          NewTemplates(store, query)        template scope
                          OpenNoteMsg{Note, Scope}, NoteCreatedMsg{Note, Scope}
                          ToggleScopeMsg{}                  ctrl+t
                          BackMsg{}                         Esc (both scopes)

internal/ui/editor        New(store, note)                  unchanged, note
                          NewTemplate(store, note)          template kind
                          (Model) Insert(text) Model
                          DoneMsg{Template bool}
                          InsertTemplateMsg{}               ctrl+t on a note

internal/ui/tmpl          New(store notes.Store) Model      picker over templates
                          InsertMsg{Text string}
                          CancelMsg{}

internal/ui/app           modeTemplate; routes everything above
```

`internal/templates` is the only new package with logic; `internal/ui/tmpl`
is the only new mode. The names are distinct on purpose: `templates` is the
format, `tmpl` is the screen.

Alternative considered: a `Templates` interface or a second store type.
Rejected; `notes.Store{Dir: ...}` already does everything needed and
`ValidateName` applies unchanged to template names.

### Root routing

| Event                                   | current | mode / state                                  |
|-----------------------------------------|---------|-----------------------------------------------|
| finder.ToggleScopeMsg (notes)           | kept    | finder, template scope, empty query           |
| finder.ToggleScopeMsg (templates)       | kept    | finder, notes scope, empty query              |
| finder.BackMsg (template scope)         | kept    | finder, notes scope, empty query              |
| finder.BackMsg (notes scope)            |         | unchanged: preview of current or quit         |
| finder.OpenNoteMsg{Scope: Templates}    | kept    | editor on template (NewTemplate)              |
| finder.NoteCreatedMsg{Scope: Templates} | kept    | editor on template (NewTemplate)              |
| editor.DoneMsg{Template: true}          | kept    | finder, template scope, empty query           |
| editor.DoneMsg{Template: false}         |         | unchanged: preview of current                 |
| editor.InsertTemplateMsg                | kept    | tmpl (New over templates store); edit kept    |
| tmpl.InsertMsg{Text}                    | kept    | edit; `m.edit = m.edit.Insert(Text)`          |
| tmpl.CancelMsg                          | kept    | edit, untouched                               |
| finder.ChangeDirMsg (either scope)      |         | unchanged: picker; returns to notes scope     |
| New/NewTemplate returns error           | nil*    | finder with "cannot read" error               |

`*` For a template read failure the finder stays in template scope and the
current note is left alone; the existing nil-ing of `current` only applies
to notes. The root reads the finder's scope with `m.finder.Scope()` when it
needs it (BackMsg); the open/create messages carry it so the root does not
have to look it up.

The editor is not rebuilt for insertion: the root switches `mode` to
`modeTemplate` and keeps `m.edit`. On `InsertMsg` it calls `Insert`, which
runs `ta.InsertString` and then the editor's scroll-to-cursor step. On
`CancelMsg` it just switches back. `WindowSizeMsg` is forwarded to `m.tmpl`
while in that mode and to `m.edit` as today, so the editor is the right
size when it reappears.

Alternative considered: picker and form as sub-states inside the editor,
like rename inside preview. Rejected; the picker is a full-screen list with
its own input and the form has several inputs, which is a mode by the
project's rule, and keeping it out of the editor keeps the editor's key
handling flat.

Alternative considered: `tmpl` holding a pointer to the editor and inserting
directly. Rejected; models are values here and the root is the only place
that composes them.

### Front matter parsing

`Parse` works on the string after `\r\n` -> `\n` normalization:

1. If the content does not start with `---\n` (or equal `---`), the whole
   content is the body and there are no fields.
2. Otherwise scan line by line from the second line for a line equal to
   `---`. None: error "front matter is not closed".
3. Decode the lines in between with `yaml.Unmarshal` into
   `struct{ Form []rawField \`yaml:"form"\` }` where `rawField` has `Name`,
   `Label`, `Type string` and `Options []string`. A YAML error is returned
   as-is (goccy/go-yaml includes line and column).
4. Validate fields in order: name non-empty and free of whitespace, `{` and
   `}`; label defaulting to name; type one of `text`, `number`, `option`;
   `option` requires at least one entry in `options`; no duplicate names.
   Errors name the field by index (`form[2]`) and by name when it has one.
5. The body is the remaining lines joined, with leading whitespace-only
   lines dropped and trailing `\n` trimmed.

`Expand` runs one regexp pass over the body, `\{\{([^{}\s]+)\}\}`, replacing
a match with the value when the captured name is a declared field and
leaving it untouched otherwise. One pass means values are never re-expanded.

Alternative considered: `strings.ReplaceAll` per field. Rejected; the
order of fields would make `a={{b}}` expand twice, and the spec says
values are literal.

YAML library: `github.com/goccy/go-yaml` v1.19.x. Pure Go, maintained,
error messages carry positions. `yaml.v3` would work equally but is in
maintenance mode. Options are decoded into `[]string`; a bare number in the
list (`options: [1, 2]`) is decoded as its string form by goccy, so no
special case is needed. This is checked by a test rather than assumed.

### Skeleton

```
---
form:
  - name: title
    label: Title
    type: text
---
# {{title}}

```

The finder in template scope does `EnsureDir`, then `store.Create(name)`
(O_EXCL, so an existing file is never clobbered), then
`store.Write(name, Skeleton)`. The editor then loads it like any file.
`Skeleton` has a test asserting `Parse` accepts it.

### Finder scope

`Model` gains `scope Scope`. `NewTemplates` builds the same model with the
templates store, `current == nil`, the prompt `templates> ` and
`scope = ScopeTemplates`. Differences keyed on the scope:

- `reload`: in template scope an `fs.ErrNotExist` from `List` is an empty
  listing, not an error.
- `enter` on the create row: `EnsureDir` + `Create` + `Write(Skeleton)` in
  template scope; `Create` only in notes scope. Both emit `NoteCreatedMsg`
  with the scope.
- `ctrl+t`: emits `ToggleScopeMsg`.
- View: prompt text, empty hint ("No templates yet. ..."), footer hints
  (`enter edit · esc notes · ctrl+t notes · ctrl+o change directory`).

`BuildRows` is untouched: with `current == nil` nothing is hidden, and the
create row logic is the same. The root builds the finder for a scope with a
`newFinder(query)` / `newTemplatesFinder()` pair so the height propagation
stays in one place.

### Editor kind and insertion

`Model` gains `template bool`. `NewTemplate` is `New` with the flag set; the
textarea setup is shared. The flag drives:

- `ctrl+t`: emits `InsertTemplateMsg` when `!template`, no-op otherwise.
- `DoneMsg{Template: m.template}`.
- Status bar: hints `esc save & back · ctrl+t template` for a note, `esc
  save & back` for a template; the right side renders `template: <name>`
  for a template.

`ta.KeyMap.TransposeCharacterBackward` is set to a disabled binding so the
textarea never sees `ctrl+t` as transpose even if the interception order
changes. `Insert(text)` is `ta.InsertString(text)` followed by the existing
`scrollToCursor`; it returns the modified model so the root can assign it.

### Insert mode (`internal/ui/tmpl`)

```
          enter, parse ok, no fields -> InsertMsg{Body}
pick ---> enter, parse ok, fields    -> form
          enter, read/parse error    -> err shown, stay
          esc                        -> CancelMsg
          other keys                 -> query / selection

          enter (not last) / down / tab  -> next row
          up / shift+tab                 -> previous row
form ---> left / right (option row)      -> cycle option
          enter (last) valid             -> InsertMsg{Expand(values)}
          enter (last) invalid number    -> err, focus that row, stay
          esc                            -> CancelMsg
          other keys (text/number row)   -> textinput
```

Picker: a `textinput` plus the listing from `store.List()` taken at `New`,
filtered with `finder.Match` (exported already) and ordered as listed. The
finder's `window` arithmetic is duplicated rather than shared; it is five
lines.

Form rows: a slice of `row{field Field; input textinput.Model; idx int}`
where `idx` is the selected option for `option` rows and `input` is used
for the other two. Only the focused row's input is focused so only one
cursor blinks. `number` validation is `strconv.ParseFloat(v, 64)` on submit;
the typed text is what gets expanded, so `1.50` stays `1.50`.

Layout: picker is query line, rows, error line, status bar, like the finder.
Form is `Label: value` per row (option values rendered as `< low >` with the
selected one highlighted via `theme.Selected`), error line, status bar.
Both take the terminal size from `WindowSizeMsg` with the usual 80x24
default for tests.

### Testing

- `templates`: table tests for `Parse` (no front matter, with front matter,
  CRLF, unterminated, invalid YAML, missing name, bad name, duplicate,
  unknown type, option without options, numeric options, other keys ignored,
  blank-line stripping, trailing newline) and `Expand` (repeat, unknown,
  value with placeholder, spaced placeholder not matched); `Dir`, `Store`
  and `EnsureDir` on `t.TempDir()`; `Skeleton` parses.
- `finder`: `NewTemplates` on a temp dir lists only `.templates/*.md`,
  missing dir is empty with no error, create row creates the dir and the
  skeleton file and emits `NoteCreatedMsg{Scope: ScopeTemplates}`,
  `ctrl+t` emits `ToggleScopeMsg`, current note is not hidden in template
  scope, View shows the prompt and the hints.
- `editor`: `Insert` places text at a mid-line cursor and the cursor after
  it; `ctrl+t` emits `InsertTemplateMsg` on a note and nothing on a
  template; `ctrl+t` does not transpose; `DoneMsg.Template` is set;
  template save writes under `.templates`; status bar names the template.
- `tmpl`: picker lists and filters, Enter on a field-less template emits
  `InsertMsg` with the body, Enter on a template with fields opens the form
  with defaults, invalid template shows the error, form navigation and
  option cycling, invalid number keeps the form open with the error and
  focus, valid submit emits the expanded text, Esc emits `CancelMsg` from
  both states, empty picker shows the hint.
- `app`: `ctrl+t` in the finder switches scope and back; Enter on a
  template opens the editor and Esc returns to template scope with the
  current note unchanged; `ctrl+t` in the editor, pick, fill, submit lands
  back in the editor with the text inserted; cancel lands back unchanged;
  `ctrl+o` from template scope returns to notes scope.

## Risks / Trade-offs

- [`ctrl+t` is swallowed by some terminals or multiplexers] -> Same class
  of risk as `ctrl+o` today; accepted. A configurable key map is a later
  change.
- [goccy/go-yaml decodes YAML booleans and nulls into strings in surprising
  ways (`options: [yes, no]` becomes `true`/`false` in some YAML modes)] ->
  Test the numeric and boolean cases; document in the templates spec's
  examples that option values are best quoted when they look like other
  scalars. Behaviour is what the library does, not special-cased.
- [A template with thousands of lines hits the textarea's 10,000-line cap on
  insertion] -> Same limitation as pasting; accepted, noted in the editor
  change already.
- [Insertion at a mid-line cursor glues the template's first line to the
  text before the cursor] -> By decision; the user asked for insertion at
  the cursor. Templates that want their own line start with a newline.
- [Templates are edited without a preview, so rename and delete are not
  available in clipt] -> Explicit non-goal; the files live in a visible
  place on disk.
- [The empty `.templates` directory is created even if the user then
  deletes the only template] -> Harmless; it is hidden from the listing.
- [Front matter in a note is not a template and will be rendered by
  preview as a rule and text] -> Unchanged from today; notes are never
  parsed.

## Open Questions

- Whether the picker should fall back to opening the form when a template
  has fields but all are `option` with a single entry. Not behaviour the
  specs fix; the form opens in every fielded case for now.
