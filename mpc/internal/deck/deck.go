// Package deck models a set of Magic-style cards written in the .mpc markup
// language, and lowers them into the shape Card Conjurer needs to render.
package deck

import (
	"fmt"
	"strings"
)

// Deck is a parsed .mpc file: file-level defaults, reusable templates, and the
// cards themselves.
type Deck struct {
	Source    string
	Defaults  Attrs
	Templates map[string]Attrs
	Cards     []*Card
}

// Card is a single card, already flattened: Attrs holds the deck defaults, the
// template it named, and its own directives merged in that order.
type Card struct {
	Attrs
	Line int // line in the source file the card block starts on

	Name    string
	Cost    string
	Type    string
	Rules   string // Card Conjurer markup, newline separated
	Power   string
	Tough   string
	Loyalty string
}

// PT renders the power/toughness box text, empty when the card has neither.
func (c *Card) PT() string {
	if c.Power == "" && c.Tough == "" {
		return c.Loyalty
	}
	return c.Power + "/" + c.Tough
}

// OutName is the base filename (no extension) used for this card's images.
func (c *Card) OutName() string {
	if n := c.Get("out"); n != "" {
		return n
	}
	name := c.Name
	if num := c.Get("number"); num != "" {
		name = num + " " + name
	}
	return sanitize(name)
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-', r == '_', r == ' ', r == '.':
			b.WriteRune(r)
		case r == ',' || r == '\'' || r == '"':
			// dropped
		default:
			b.WriteRune('-')
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		out = "unnamed"
	}
	return out
}

// Attrs is an ordered set of @directives. Later Set calls win, which is how
// deck < template < card precedence is implemented.
type Attrs struct {
	keys []string
	vals map[string]string
}

func (a *Attrs) Set(k, v string) {
	if a.vals == nil {
		a.vals = map[string]string{}
	}
	if _, ok := a.vals[k]; !ok {
		a.keys = append(a.keys, k)
	}
	a.vals[k] = v
}

func (a Attrs) Get(k string) string { return a.vals[k] }

func (a Attrs) Has(k string) bool { _, ok := a.vals[k]; return ok }

func (a Attrs) Keys() []string { return a.keys }

// Merge returns a copy of a with every attribute of b layered on top.
func (a Attrs) Merge(b Attrs) Attrs {
	out := Attrs{}
	for _, k := range a.keys {
		out.Set(k, a.vals[k])
	}
	for _, k := range b.keys {
		out.Set(k, b.vals[k])
	}
	return out
}

// Bool reads a flag-ish attribute. A bare directive (`@legendary`) counts as true.
func (a Attrs) Bool(k string) bool {
	v, ok := a.vals[k]
	if !ok {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "1", "true", "yes", "on":
		return true
	}
	return false
}

// Float reads a numeric attribute, falling back to def when absent or unparseable.
func (a Attrs) Float(k string, def float64) float64 {
	v, ok := a.vals[k]
	if !ok {
		return def
	}
	var f float64
	if _, err := fmt.Sscanf(strings.TrimSpace(v), "%g", &f); err != nil {
		return def
	}
	return f
}

// Int reads an integer attribute, falling back to def.
func (a Attrs) Int(k string, def int) int {
	v, ok := a.vals[k]
	if !ok {
		return def
	}
	var n int
	if _, err := fmt.Sscanf(strings.TrimSpace(v), "%d", &n); err != nil {
		return def
	}
	return n
}
