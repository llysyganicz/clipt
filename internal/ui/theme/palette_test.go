package theme

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// complete is a valid theme file whose colors are all distinct from the
// built-in ones.
const complete = `background = "#000001"
foreground = "#000002"
selection = "#000003"
black = "#000010"
red = "#000011"
green = "#000012"
yellow = "#000013"
blue = "#000014"
magenta = "#000015"
cyan = "#000016"
white = "#000017"
bright_black = "#000020"
bright_red = "#000021"
bright_green = "#000022"
bright_yellow = "#000023"
bright_blue = "#000024"
bright_magenta = "#000025"
bright_cyan = "#000026"
bright_white = "#000027"
`

func without(t *testing.T, key string) string {
	t.Helper()
	var out []string
	found := false
	for _, l := range strings.Split(complete, "\n") {
		if strings.HasPrefix(l, key+" ") {
			found = true
			continue
		}
		out = append(out, l)
	}
	if !found {
		t.Fatalf("key %q not in fixture", key)
	}
	return strings.Join(out, "\n")
}

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		in        string
		selection string
		errSubs   []string
	}{
		{name: "complete", in: complete, selection: "#000003"},
		{name: "selection omitted", in: without(t, "selection"), selection: "#000020"},
		{name: "unknown key ignored", in: complete + "cursor = \"#ffffff\"\n", selection: "#000003"},
		{name: "upper case is lowered", in: strings.Replace(complete, `selection = "#000003"`, `selection = "#ABCDEF"`, 1), selection: "#abcdef"},
		{name: "missing key", in: without(t, "bright_yellow"), errSubs: []string{"t.toml", "bright_yellow"}},
		{name: "bad color", in: strings.Replace(complete, `red = "#000011"`, `red = "tomato"`, 1), errSubs: []string{"t.toml", `"red"`, "tomato"}},
		{name: "short color", in: strings.Replace(complete, `blue = "#000014"`, `blue = "#fff"`, 1), errSubs: []string{"t.toml", `"blue"`}},
		{name: "bad selection", in: strings.Replace(complete, `selection = "#000003"`, `selection = "x"`, 1), errSubs: []string{"t.toml", `"selection"`}},
		{name: "wrong type", in: strings.Replace(complete, `red = "#000011"`, `red = 5`, 1), errSubs: []string{"t.toml"}},
		{name: "invalid toml", in: "red = \n", errSubs: []string{"t.toml"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(strings.NewReader(tc.in), "t.toml")
			if len(tc.errSubs) > 0 {
				if err == nil {
					t.Fatalf("Parse() = %+v, want error", p)
				}
				for _, s := range tc.errSubs {
					if !strings.Contains(err.Error(), s) {
						t.Errorf("error %q does not contain %q", err, s)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if p.Selection != tc.selection {
				t.Errorf("Selection = %q, want %q", p.Selection, tc.selection)
			}
			if p.Background != "#000001" || p.BrightWhite != "#000027" || p.Cyan != "#000016" {
				t.Errorf("palette = %+v", p)
			}
		})
	}
}

func TestBuiltin(t *testing.T) {
	got := Builtin()
	want := []string{"gruvbox_dark", "gruvbox_light"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Builtin() = %v, want %v", got, want)
	}
	for _, n := range got {
		if _, err := Load(n, t.TempDir()); err != nil {
			t.Errorf("Load(%q) = %v", n, err)
		}
	}
}

func TestGruvboxDarkMatchesPreviousConstants(t *testing.T) {
	p := mustBuiltin("gruvbox_dark")
	want := map[string]string{
		"Background":   "#282828",
		"Selection":    "#504945",
		"Foreground":   "#ebdbb2",
		"BrightYellow": "#fabd2f",
		"BrightCyan":   "#8ec07c",
		"BrightRed":    "#fb4934",
		"BrightBlack":  "#928374",
	}
	v := reflect.ValueOf(p)
	for field, w := range want {
		if got := v.FieldByName(field).String(); got != w {
			t.Errorf("%s = %q, want %q", field, got, w)
		}
	}
	if l := mustBuiltin("gruvbox_light"); l.Background != "#fbf1c7" || l.Selection == l.BrightBlack {
		t.Errorf("gruvbox_light = %+v", l)
	}
}

func writeTheme(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, "themes", name+".toml")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoad(t *testing.T) {
	t.Run("user file", func(t *testing.T) {
		dir := t.TempDir()
		writeTheme(t, dir, "solarized", complete)
		p, err := Load("solarized", dir)
		if err != nil || p.Background != "#000001" {
			t.Errorf("Load() = %+v, %v", p, err)
		}
	})
	t.Run("built-in fallback", func(t *testing.T) {
		p, err := Load("gruvbox_light", t.TempDir())
		if err != nil || p.Background != "#fbf1c7" {
			t.Errorf("Load() = %+v, %v", p, err)
		}
	})
	t.Run("user file shadows built-in", func(t *testing.T) {
		dir := t.TempDir()
		writeTheme(t, dir, "gruvbox_dark", complete)
		p, err := Load("gruvbox_dark", dir)
		if err != nil || p.Background != "#000001" {
			t.Errorf("Load() = %+v, %v", p, err)
		}
	})
	t.Run("unknown", func(t *testing.T) {
		dir := t.TempDir()
		_, err := Load("nope", dir)
		if err == nil {
			t.Fatal("Load() succeeded")
		}
		want := filepath.Join(dir, "themes", "nope.toml")
		if !strings.Contains(err.Error(), `"nope"`) || !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not name nope and %s", err, want)
		}
	})
	t.Run("user parse error names path and key", func(t *testing.T) {
		dir := t.TempDir()
		content := strings.Replace(complete, `red = "#000011"`, `red = "tomato"`, 1)
		p := writeTheme(t, dir, "bad", content)
		_, err := Load("bad", dir)
		if err == nil || !strings.Contains(err.Error(), p) || !strings.Contains(err.Error(), `"red"`) {
			t.Errorf("error = %v, want naming %s and red", err, p)
		}
		got, _ := os.ReadFile(p)
		if string(got) != content {
			t.Error("theme file modified")
		}
	})
	t.Run("user file missing key", func(t *testing.T) {
		dir := t.TempDir()
		p := writeTheme(t, dir, "short", without(t, "bright_yellow"))
		_, err := Load("short", dir)
		if err == nil || !strings.Contains(err.Error(), p) || !strings.Contains(err.Error(), "bright_yellow") {
			t.Errorf("error = %v", err)
		}
	})
	t.Run("user file invalid toml", func(t *testing.T) {
		dir := t.TempDir()
		p := writeTheme(t, dir, "broken", "red = \n")
		_, err := Load("broken", dir)
		if err == nil || !strings.Contains(err.Error(), p) {
			t.Errorf("error = %v", err)
		}
	})
	t.Run("unreadable user file", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root ignores file permissions")
		}
		dir := t.TempDir()
		p := writeTheme(t, dir, "locked", complete)
		os.Chmod(p, 0o000)
		t.Cleanup(func() { os.Chmod(p, 0o644) })
		_, err := Load("locked", dir)
		if err == nil || !strings.Contains(err.Error(), p) {
			t.Errorf("error = %v", err)
		}
	})
}

func TestEmbeddedParseErrorNamesBuiltinPath(t *testing.T) {
	// The embedded files are valid, so exercise the wrapping through Parse
	// with the name Load passes for them.
	_, err := Parse(strings.NewReader("red = \n"), "builtin/x.toml")
	if err == nil || !strings.Contains(err.Error(), "builtin/x.toml") {
		t.Errorf("error = %v", err)
	}
}

// renderAll renders every exported style and the editor and input styles.
func renderAll() map[string]string {
	out := map[string]string{}
	for name, s := range map[string]lipgloss.Style{
		"Normal": Normal, "Selected": Selected, "Create": Create,
		"Hint": Hint, "Error": Error, "Prompt": Prompt,
		"Tag": Tag, "RowTag": RowTag,
		"LineNumber": LineNumber, "CursorLineNumber": CursorLineNumber,
	} {
		out[name] = s.Render("x")
	}
	e := EditorStyles()
	out["editor.Text"] = e.Focused.Text.Render("x")
	out["editor.Selection"] = e.Focused.Selection.Render("x")
	out["editor.CursorLineNumber"] = e.Focused.CursorLineNumber.Render("x")
	r, g, b, _ := e.Cursor.Color.RGBA()
	out["editor.Cursor"] = colorHex(r, g, b)
	i := InputStyles()
	out["input.Prompt"] = i.Focused.Prompt.Render("x")
	out["markdown.Heading"] = *Markdown.Heading.Color
	out["markdown.Document"] = *Markdown.Document.Color
	return out
}

func colorHex(r, g, b uint32) string {
	const digits = "0123456789abcdef"
	var s [7]byte
	s[0] = '#'
	for i, v := range []uint32{r >> 8, g >> 8, b >> 8} {
		s[1+2*i] = digits[v>>4]
		s[2+2*i] = digits[v&0xf]
	}
	return string(s[:])
}

func TestApply(t *testing.T) {
	dark := mustBuiltin("gruvbox_dark")
	light := mustBuiltin("gruvbox_light")
	t.Cleanup(func() { Apply(dark) })

	before := renderAll()
	Apply(dark)
	if after := renderAll(); !reflect.DeepEqual(before, after) {
		t.Errorf("Apply(gruvbox_dark) changed the default styles:\n%v\n%v", before, after)
	}

	Apply(light)
	got := renderAll()
	if reflect.DeepEqual(before, got) {
		t.Fatal("Apply(gruvbox_light) changed nothing")
	}
	checks := map[string]struct {
		style lipgloss.Style
		want  string
	}{
		"Selected":         {Selected, light.BrightYellow},
		"Prompt":           {Prompt, light.BrightYellow},
		"CursorLineNumber": {CursorLineNumber, light.BrightYellow},
		"Create":           {Create, light.BrightCyan},
		"Tag":              {Tag, light.BrightCyan},
		"Error":            {Error, light.BrightRed},
		"Hint":             {Hint, light.BrightBlack},
		"RowTag":           {RowTag, light.BrightBlack},
		"LineNumber":       {LineNumber, light.BrightBlack},
		"Normal":           {Normal, light.Foreground},
	}
	for name, c := range checks {
		if fg := c.style.GetForeground(); fg != lipgloss.Color(c.want) {
			t.Errorf("%s foreground = %v, want %s", name, fg, c.want)
		}
	}
	if bg := Normal.GetBackground(); bg != (lipgloss.NoColor{}) {
		t.Errorf("Normal background = %v, want none", bg)
	}
	if Background() != lipgloss.Color(light.Background) || Foreground() != lipgloss.Color(light.Foreground) {
		t.Errorf("Background() = %v, Foreground() = %v, want %s, %s", Background(), Foreground(), light.Background, light.Foreground)
	}
	if sel := EditorStyles().Focused.Selection.GetBackground(); sel != lipgloss.Color(light.Selection) {
		t.Errorf("selection background = %v, want %s", sel, light.Selection)
	}
	if got["editor.Cursor"] != light.Foreground {
		t.Errorf("cursor = %s, want %s", got["editor.Cursor"], light.Foreground)
	}
	if *Markdown.Heading.Color != light.BrightYellow || *Markdown.Document.Color != light.Foreground {
		t.Errorf("markdown heading %s document %s", *Markdown.Heading.Color, *Markdown.Document.Color)
	}
}

// colorFields collects every non-nil Color and BackgroundColor in v, keyed
// by field path.
func colorFields(prefix string, v reflect.Value, out map[string]string) {
	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			colorFields(prefix, v.Elem(), out)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			name := prefix + "." + f.Name
			if (f.Name == "Color" || f.Name == "BackgroundColor") && f.Type == reflect.TypeOf((*string)(nil)) {
				if !v.Field(i).IsNil() {
					out[name] = v.Field(i).Elem().String()
				}
				continue
			}
			colorFields(name, v.Field(i), out)
		}
	}
}

func TestMarkdownColorsComeFromPalette(t *testing.T) {
	for _, name := range Builtin() {
		p := mustBuiltin(name)
		allowed := map[string]bool{}
		for _, c := range p.Colors() {
			allowed[c] = true
		}
		got := map[string]string{}
		colorFields("", reflect.ValueOf(markdownStyle(p)), got)
		if len(got) == 0 {
			t.Fatal("no colors found")
		}
		for field, c := range got {
			if !allowed[c] {
				t.Errorf("%s: %s = %q is not a palette color", name, field, c)
			}
		}
		c := markdownStyle(p)
		want := map[string]struct{ got, want *string }{
			"Document":       {c.Document.Color, &p.Foreground},
			"Heading":        {c.Heading.Color, &p.BrightYellow},
			"H1":             {c.H1.Color, &p.BrightYellow},
			"H6":             {c.H6.Color, &p.BrightYellow},
			"Table":          {c.Table.Color, &p.BrightYellow},
			"Link":           {c.Link.Color, &p.BrightCyan},
			"LinkText":       {c.LinkText.Color, &p.BrightCyan},
			"Code":           {c.Code.Color, &p.BrightCyan},
			"CodeBlock":      {c.CodeBlock.Color, &p.BrightCyan},
			"BlockQuote":     {c.BlockQuote.Color, &p.BrightBlack},
			"HorizontalRule": {c.HorizontalRule.Color, &p.BrightBlack},
			"Image":          {c.Image.Color, &p.BrightBlack},
			"ImageText":      {c.ImageText.Color, &p.BrightBlack},
			"Item":           {c.Item.Color, &p.BrightBlue},
			"Enumeration":    {c.Enumeration.Color, &p.BrightBlue},
			"Task":           {c.Task.Color, &p.BrightGreen},
		}
		for el, w := range want {
			if w.got == nil || *w.got != *w.want {
				t.Errorf("%s: %s color = %v, want %s", name, el, w.got, *w.want)
			}
		}
		if c.H1.BackgroundColor != nil || c.Code.BackgroundColor != nil || c.Heading.BackgroundColor != nil {
			t.Errorf("%s: heading or code has a background", name)
		}
		if c.Emph.Color != nil || c.Strong.Color != nil || c.Emph.Italic == nil || c.Strong.Bold == nil {
			t.Errorf("%s: emphasis or strong not plain foreground with italic/bold", name)
		}
		if c.CodeBlock.Chroma != nil {
			t.Errorf("%s: code block keeps chroma colors", name)
		}
	}
}
