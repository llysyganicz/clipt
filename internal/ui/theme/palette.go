package theme

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

// Default is the theme used when the config names none.
const Default = "gruvbox_dark"

// Palette is a terminal color scheme: the background and foreground, the
// editor selection background and the sixteen standard ANSI colors. Every
// value is a lower-case "#rrggbb" string.
type Palette struct {
	Background, Foreground, Selection string

	Black, Red, Green, Yellow, Blue, Magenta, Cyan, White                                                 string
	BrightBlack, BrightRed, BrightGreen, BrightYellow, BrightBlue, BrightMagenta, BrightCyan, BrightWhite string
}

// Colors returns every palette value, so callers can check that an output
// uses nothing else.
func (p Palette) Colors() []string {
	return []string{
		p.Background, p.Foreground, p.Selection,
		p.Black, p.Red, p.Green, p.Yellow, p.Blue, p.Magenta, p.Cyan, p.White,
		p.BrightBlack, p.BrightRed, p.BrightGreen, p.BrightYellow,
		p.BrightBlue, p.BrightMagenta, p.BrightCyan, p.BrightWhite,
	}
}

// file is the on-disk shape of a theme: one string key per color.
type file struct {
	Background    string `toml:"background"`
	Foreground    string `toml:"foreground"`
	Selection     string `toml:"selection"`
	Black         string `toml:"black"`
	Red           string `toml:"red"`
	Green         string `toml:"green"`
	Yellow        string `toml:"yellow"`
	Blue          string `toml:"blue"`
	Magenta       string `toml:"magenta"`
	Cyan          string `toml:"cyan"`
	White         string `toml:"white"`
	BrightBlack   string `toml:"bright_black"`
	BrightRed     string `toml:"bright_red"`
	BrightGreen   string `toml:"bright_green"`
	BrightYellow  string `toml:"bright_yellow"`
	BrightBlue    string `toml:"bright_blue"`
	BrightMagenta string `toml:"bright_magenta"`
	BrightCyan    string `toml:"bright_cyan"`
	BrightWhite   string `toml:"bright_white"`
}

// required lists the keys a theme file must set, in file order, with the
// palette field each one fills.
var required = []struct {
	key string
	get func(*file) *string
	set func(*Palette) *string
}{
	{"background", func(f *file) *string { return &f.Background }, func(p *Palette) *string { return &p.Background }},
	{"foreground", func(f *file) *string { return &f.Foreground }, func(p *Palette) *string { return &p.Foreground }},
	{"black", func(f *file) *string { return &f.Black }, func(p *Palette) *string { return &p.Black }},
	{"red", func(f *file) *string { return &f.Red }, func(p *Palette) *string { return &p.Red }},
	{"green", func(f *file) *string { return &f.Green }, func(p *Palette) *string { return &p.Green }},
	{"yellow", func(f *file) *string { return &f.Yellow }, func(p *Palette) *string { return &p.Yellow }},
	{"blue", func(f *file) *string { return &f.Blue }, func(p *Palette) *string { return &p.Blue }},
	{"magenta", func(f *file) *string { return &f.Magenta }, func(p *Palette) *string { return &p.Magenta }},
	{"cyan", func(f *file) *string { return &f.Cyan }, func(p *Palette) *string { return &p.Cyan }},
	{"white", func(f *file) *string { return &f.White }, func(p *Palette) *string { return &p.White }},
	{"bright_black", func(f *file) *string { return &f.BrightBlack }, func(p *Palette) *string { return &p.BrightBlack }},
	{"bright_red", func(f *file) *string { return &f.BrightRed }, func(p *Palette) *string { return &p.BrightRed }},
	{"bright_green", func(f *file) *string { return &f.BrightGreen }, func(p *Palette) *string { return &p.BrightGreen }},
	{"bright_yellow", func(f *file) *string { return &f.BrightYellow }, func(p *Palette) *string { return &p.BrightYellow }},
	{"bright_blue", func(f *file) *string { return &f.BrightBlue }, func(p *Palette) *string { return &p.BrightBlue }},
	{"bright_magenta", func(f *file) *string { return &f.BrightMagenta }, func(p *Palette) *string { return &p.BrightMagenta }},
	{"bright_cyan", func(f *file) *string { return &f.BrightCyan }, func(p *Palette) *string { return &p.BrightCyan }},
	{"bright_white", func(f *file) *string { return &f.BrightWhite }, func(p *Palette) *string { return &p.BrightWhite }},
}

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// Parse decodes a theme file. Every key but selection is required, each
// value must be a "#rrggbb" color and unknown keys are ignored. selection
// defaults to bright_black. Errors mention name (the file's path) and, for a
// missing or bad key, the key.
func Parse(r io.Reader, name string) (Palette, error) {
	var f file
	md, err := toml.NewDecoder(r).Decode(&f)
	if err != nil {
		return Palette{}, fmt.Errorf("invalid theme file %s: %w", name, err)
	}
	var p Palette
	for _, k := range required {
		if !md.IsDefined(k.key) {
			return Palette{}, fmt.Errorf("invalid theme file %s: missing key %q", name, k.key)
		}
		v := *k.get(&f)
		if !hexColor.MatchString(v) {
			return Palette{}, fmt.Errorf("invalid theme file %s: key %q: %q is not a #rrggbb color", name, k.key, v)
		}
		*k.set(&p) = strings.ToLower(v)
	}
	switch {
	case !md.IsDefined("selection"):
		p.Selection = p.BrightBlack
	case !hexColor.MatchString(f.Selection):
		return Palette{}, fmt.Errorf("invalid theme file %s: key %q: %q is not a #rrggbb color", name, "selection", f.Selection)
	default:
		p.Selection = strings.ToLower(f.Selection)
	}
	return p, nil
}

//go:embed builtin/*.toml
var builtin embed.FS

// Builtin returns the names of the themes compiled into the binary, sorted.
func Builtin() []string {
	entries, err := fs.ReadDir(builtin, "builtin")
	if err != nil {
		panic(err) // the embedded directory always exists
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, strings.TrimSuffix(e.Name(), ".toml"))
	}
	sort.Strings(names)
	return names
}

// Load returns the theme called name: <dir>/themes/<name>.toml when that
// file exists, otherwise the built-in theme of that name. A failure to read
// or parse either is reported with the file's path (built-in themes as
// builtin/<name>.toml); an unknown name is reported with the path that was
// looked for.
func Load(name, dir string) (Palette, error) {
	user := filepath.Join(dir, "themes", name+".toml")
	f, err := os.Open(user)
	if err == nil {
		defer f.Close()
		return Parse(f, user)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return Palette{}, fmt.Errorf("cannot read theme file %s: %w", user, err)
	}
	embedded := path.Join("builtin", name+".toml")
	e, err := builtin.Open(embedded)
	if err != nil {
		return Palette{}, fmt.Errorf("theme %q not found (looked in %s)", name, user)
	}
	defer e.Close()
	return Parse(e, embedded)
}

// mustBuiltin parses an embedded theme; the files are part of the binary
// and tested, so a failure is a programming error.
func mustBuiltin(name string) Palette {
	f, err := builtin.Open(path.Join("builtin", name+".toml"))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	p, err := Parse(f, "builtin/"+name+".toml")
	if err != nil {
		panic(err)
	}
	return p
}
