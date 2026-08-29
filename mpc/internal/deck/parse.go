package deck

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Parse reads a .mpc file. Includes are resolved relative to path.
func ParseFile(path string) (*Deck, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parse(f, path, map[string]bool{})
}

// Parse reads .mpc markup from r. name is used in error messages and to
// resolve @include paths.
func Parse(r io.Reader, name string) (*Deck, error) {
	return parse(r, name, map[string]bool{})
}

type block struct {
	line  int
	attrs Attrs
	body  []string
}

func parse(r io.Reader, path string, seen map[string]bool) (*Deck, error) {
	blocks, err := readBlocks(r, path)
	if err != nil {
		return nil, err
	}

	d := &Deck{Source: path, Templates: map[string]Attrs{}}

	for i, b := range blocks {
		switch {
		case b.attrs.Has("deftemplate"):
			name := strings.TrimSpace(b.attrs.Get("deftemplate"))
			if name == "" {
				return nil, fmt.Errorf("%s:%d: @deftemplate needs a name", path, b.line)
			}
			t := b.attrs
			t.vals = copyMap(t.vals)
			delete(t.vals, "deftemplate")
			t.keys = without(t.keys, "deftemplate")
			d.Templates[name] = t

		case i == 0 && !hasText(b.body):
			// Leading block with no card text is the deck header.
			d.Defaults = b.attrs
			if inc := b.attrs.Get("include"); inc != "" {
				if err := d.include(inc, path, seen); err != nil {
					return nil, err
				}
			}

		case !hasText(b.body) && len(b.attrs.Keys()) == 0:
			// blank block between separators; ignore

		default:
			c, err := buildCard(d, b, path)
			if err != nil {
				return nil, err
			}
			d.Cards = append(d.Cards, c)
		}
	}
	return d, nil
}

func (d *Deck) include(list, path string, seen map[string]bool) error {
	for _, rel := range strings.Split(list, ",") {
		rel = strings.TrimSpace(rel)
		if rel == "" {
			continue
		}
		p := rel
		if !filepath.IsAbs(p) {
			p = filepath.Join(filepath.Dir(path), p)
		}
		abs, _ := filepath.Abs(p)
		if seen[abs] {
			return fmt.Errorf("%s: @include cycle at %s", path, rel)
		}
		seen[abs] = true

		f, err := os.Open(p)
		if err != nil {
			return fmt.Errorf("%s: @include %s: %w", path, rel, err)
		}
		sub, err := parse(f, p, seen)
		f.Close()
		if err != nil {
			return err
		}
		// Included defaults sit *under* this file's own, and included
		// templates are only used when this file doesn't redefine them.
		d.Defaults = sub.Defaults.Merge(d.Defaults)
		for k, v := range sub.Templates {
			if _, ok := d.Templates[k]; !ok {
				d.Templates[k] = v
			}
		}
		d.Cards = append(d.Cards, sub.Cards...)
	}
	return nil
}

var sepRe = regexp.MustCompile(`^\s*(-{3,}|={3,})\s*$`)

func readBlocks(r io.Reader, path string) ([]block, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	var blocks []block
	cur := block{line: 1}
	flush := func(next int) {
		blocks = append(blocks, cur)
		cur = block{line: next}
	}

	lineNo := 0
	for sc.Scan() {
		lineNo++
		raw := strings.TrimRight(sc.Text(), " \t\r")
		trimmed := strings.TrimSpace(raw)

		switch {
		case strings.HasPrefix(trimmed, "#"):
			// comment
		case sepRe.MatchString(raw):
			flush(lineNo + 1)
		case strings.HasPrefix(trimmed, "@"):
			k, v := splitDirective(trimmed)
			if k == "" {
				return nil, fmt.Errorf("%s:%d: empty directive", path, lineNo)
			}
			cur.attrs.Set(k, v)
		default:
			cur.body = append(cur.body, raw)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	flush(lineNo + 1)
	return blocks, nil
}

func splitDirective(s string) (key, val string) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "@")
	// `@key value`, `@key: value` and `@key=value` all work.
	i := strings.IndexAny(s, " \t:=")
	if i < 0 {
		return strings.ToLower(s), ""
	}
	key = strings.ToLower(s[:i])
	val = strings.TrimSpace(strings.TrimLeft(s[i:], " \t:="))
	return key, val
}

var (
	costRe    = regexp.MustCompile(`\s*((?:\{[^}]*\}\s*)+)$`)
	ptRe      = regexp.MustCompile(`^\s*([0-9+\-*X]+)\s*/\s*([0-9+\-*X]+)\s*$`)
	loyaltyRe = regexp.MustCompile(`^\s*(?:\[\s*([0-9+\-*X]+)\s*\]|[Ll]oyalty:?\s*([0-9+\-*X]+))\s*$`)
	italicRe  = regexp.MustCompile(`\*([^*\n]+)\*`)
)

func buildCard(d *Deck, b block, path string) (*Card, error) {
	attrs := d.Defaults
	if t := b.attrs.Get("template"); t != "" {
		tpl, ok := d.Templates[t]
		if !ok {
			return nil, fmt.Errorf("%s:%d: unknown @template %q", path, b.line, t)
		}
		attrs = attrs.Merge(tpl)
	} else if t := d.Defaults.Get("template"); t != "" {
		if tpl, ok := d.Templates[t]; ok {
			attrs = attrs.Merge(tpl)
		} else {
			return nil, fmt.Errorf("%s: unknown @template %q in header", path, t)
		}
	}
	attrs = attrs.Merge(b.attrs)

	c := &Card{Attrs: attrs, Line: b.line}

	lines := trimBlankEdges(b.body)
	if len(lines) == 0 {
		if c.Name = attrs.Get("name"); c.Name == "" {
			return nil, fmt.Errorf("%s:%d: card block has no text and no @name", path, b.line)
		}
		return c, nil
	}

	// 1. title line: name plus an optional trailing mana cost
	title := strings.TrimSpace(lines[0])
	if m := costRe.FindStringSubmatch(title); m != nil {
		c.Cost = strings.Join(strings.Fields(m[1]), "")
		title = strings.TrimSpace(title[:len(title)-len(m[0])])
	}
	c.Name = title
	lines = lines[1:]

	// 2. type line
	if len(lines) > 0 {
		c.Type = strings.TrimSpace(lines[0])
		lines = lines[1:]
	}

	// 3. trailing power/toughness or loyalty
	lines = trimBlankEdges(lines)
	if n := len(lines); n > 0 {
		last := lines[n-1]
		if m := ptRe.FindStringSubmatch(last); m != nil {
			c.Power, c.Tough = m[1], m[2]
			lines = trimBlankEdges(lines[:n-1])
		} else if m := loyaltyRe.FindStringSubmatch(last); m != nil {
			c.Loyalty = m[1] + m[2]
			lines = trimBlankEdges(lines[:n-1])
		}
	}

	c.Rules = rulesMarkup(lines)

	// Explicit directives still win over anything scraped from the body.
	if v := attrs.Get("cost"); v != "" {
		c.Cost = v
	}
	if v := attrs.Get("type"); v != "" {
		c.Type = v
	}
	if v := attrs.Get("pt"); v != "" {
		if m := ptRe.FindStringSubmatch(v); m != nil {
			c.Power, c.Tough = m[1], m[2]
		}
	}
	if v := attrs.Get("name"); v != "" {
		c.Name = v
	}
	return c, nil
}

// rulesMarkup turns the rules/flavor body into Card Conjurer's text markup.
// Lines starting with "> " are flavor text; the first one opens the flavor bar.
func rulesMarkup(lines []string) string {
	var rules, flavor []string
	inFlavor := false
	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, ">") {
			inFlavor = true
			flavor = append(flavor, strings.TrimSpace(strings.TrimPrefix(t, ">")))
			continue
		}
		if inFlavor {
			// A non-"> " line after flavor started is still flavor; that
			// keeps multi-line quotes and attributions readable.
			flavor = append(flavor, t)
			continue
		}
		rules = append(rules, ln)
	}

	out := strings.Join(trimBlankEdges(rules), "\n")
	if len(flavor) > 0 {
		f := strings.Join(trimBlankEdges(flavor), "\n")
		if out != "" {
			out += "\n"
		}
		out += "{flavor}" + f
	}
	return emphasis(out)
}

// emphasis rewrites *asterisk* spans as Card Conjurer italics, leaving any
// literal {i}/{/i} the author typed alone.
func emphasis(s string) string {
	return italicRe.ReplaceAllString(s, "{i}$1{/i}")
}

func hasText(lines []string) bool {
	for _, ln := range lines {
		if strings.TrimSpace(ln) != "" {
			return true
		}
	}
	return false
}

func trimBlankEdges(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func copyMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func without(ss []string, drop string) []string {
	out := ss[:0:0]
	for _, s := range ss {
		if s != drop {
			out = append(out, s)
		}
	}
	return out
}
