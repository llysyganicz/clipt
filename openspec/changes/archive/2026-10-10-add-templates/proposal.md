## Why

Many notes start from the same skeleton (a meeting, a daily log, a book
entry), and today the only way to get one is to retype it or copy it from an
older note. Templates with a small form in the front matter let the user
insert a prefilled skeleton at the cursor in a few keystrokes, which is the
next core feature after the editor landed.

## What Changes

- Add templates: markdown files under `.templates/` inside the notes
  directory. A template may start with YAML front matter whose `form` key
  lists fields (`name`, `label`, `type`, and `options` for the `option`
  type). The markdown body refers to fields as `{{name}}`.
- Add template insertion to edit mode: `ctrl+t` opens a fuzzy picker over the
  templates, then a form with one input per field (`text` empty, `number`
  prefilled with `0`, `option` cycling through its list starting at the
  first). Submitting expands the body and inserts it at the cursor; Esc
  returns to the editor unchanged. A template without fields is inserted
  straight away.
- Add a template scope to the finder: `ctrl+t` toggles between the notes
  listing and the templates listing. In template scope Enter opens the
  template directly in the editor, the create row writes a template with a
  skeleton front matter, and Esc returns to the notes scope. Templates never
  become the current note; leaving the editor on a template returns to the
  finder in template scope.
- Templates are never listed as notes. The `.templates/` directory is
  created on first template creation and a missing directory lists as empty.
- Drop the text area's default `ctrl+t` binding (transpose character) in
  edit mode; `ctrl+t` means templates everywhere.
- New dependency: a YAML parser (`github.com/goccy/go-yaml`) for the front
  matter.

## Capabilities

### New Capabilities

- `templates`: the template file format and storage: location, front matter
  and fence rules, field definitions and types, defaults, placeholder
  expansion, body stripping, and parse errors.
- `template-insert`: the insert mode entered from the editor: the template
  picker, the field form, validation, and the Esc/submit transitions back to
  the editor.

### Modified Capabilities

- `finder`: a template scope toggled with `ctrl+t` (listing, prompt, hints,
  empty hint, create row behaviour, opening into the editor), and Esc in
  template scope returning to the notes scope instead of leaving the finder.
- `editor`: `ctrl+t` is reserved for template insertion, inserted text lands
  at the cursor and counts as a change, and Esc on a template returns to the
  finder in template scope rather than to preview.

## Non-goals

- Templates nested in templates: `ctrl+t` does nothing while editing a
  template, and an inserted body is not expanded again.
- Any front matter in notes themselves; only templates are parsed.
- TOML front matter (`+++`). The fence detector only knows `---`; a second
  fence is a small later change.
- Field types beyond `text`, `number` and `option`, required fields, regex
  validation, default values other than `0` and the first option.
- Date or time placeholders, conditionals, loops, or any template language
  beyond literal `{{name}}` replacement.
- Rename and delete of templates from inside clipt; templates open straight
  into the editor and have no preview. Manage the files on disk for now.
- Previewing a template's rendered markdown.
- Templates in the preview mode or a "new note from template" flow in the
  finder; insertion is only from edit mode.
- Remembering the last used template or form values.

## Impact

- New package `internal/templates` (parse, expand, directory helpers,
  skeleton) with temp-dir tests and no TUI dependency.
- New package `internal/ui/tmpl` for the insert mode (picker and form).
- `internal/ui/finder` gains a scope, a toggle key and template-aware create
  and open behaviour; its messages carry the scope.
- `internal/ui/editor` gains the `ctrl+t` hook, an insert-at-cursor method
  and a template kind that changes the exit route and status bar.
- `internal/ui/app` routes the new messages and owns the insert mode.
- `internal/notes` is reused unchanged on the `.templates` subdirectory.
- `go.mod` gains `github.com/goccy/go-yaml`.
- Writes stay inside the notes directory: `.templates/` and the files in it
  are the only new paths touched.
