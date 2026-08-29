package deck

import (
	"strings"
	"testing"
)

const sample = `
# wedding deck
@set WED
@artist Robert
@art-dir art

---
@deftemplate hero
@frame M15Regular-1
@rarity mythic

---
@template hero
@art cat.jpg
@number 001
Sir Whiskers, Snack Thief {2}{G}{W}
Legendary Creature — Cat Noble

Vigilance
When ~ enters, create a Food token.
> *Mine now.*
3/4
`

func TestParse(t *testing.T) {
	d, err := Parse(strings.NewReader(sample), "test.mpc")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(d.Cards); got != 1 {
		t.Fatalf("cards = %d, want 1", got)
	}
	c := d.Cards[0]
	checks := []struct{ name, got, want string }{
		{"name", c.Name, "Sir Whiskers, Snack Thief"},
		{"cost", c.Cost, "{2}{G}{W}"},
		{"type", c.Type, "Legendary Creature — Cat Noble"},
		{"pt", c.PT(), "3/4"},
		{"set", c.Get("set"), "WED"},
		{"artist", c.Get("artist"), "Robert"},
		{"rarity", c.Get("rarity"), "mythic"},
		{"frame", c.Get("frame"), "M15Regular-1"},
		{"art", c.Get("art"), "cat.jpg"},
		{"out", c.OutName(), "001 Sir Whiskers Snack Thief"},
		{"rules", c.Rules, "Vigilance\nWhen ~ enters, create a Food token.\n{flavor}{i}Mine now.{/i}"},
	}
	for _, k := range checks {
		if k.got != k.want {
			t.Errorf("%s = %q, want %q", k.name, k.got, k.want)
		}
	}
}

func TestParseNoCostNoPT(t *testing.T) {
	d, err := Parse(strings.NewReader("---\nPlains\nBasic Land — Plains\n"), "t.mpc")
	if err != nil {
		t.Fatal(err)
	}
	c := d.Cards[0]
	if c.Name != "Plains" || c.Cost != "" || c.Type != "Basic Land — Plains" || c.PT() != "" {
		t.Fatalf("got %+v", c)
	}
}

func TestLoyalty(t *testing.T) {
	d, err := Parse(strings.NewReader("---\nBride {3}{W}{W}\nLegendary Planeswalker — Bride\n+1: Draw a card.\n[4]\n"), "t.mpc")
	if err != nil {
		t.Fatal(err)
	}
	if got := d.Cards[0].PT(); got != "4" {
		t.Fatalf("loyalty = %q, want 4", got)
	}
}

func TestPrecedence(t *testing.T) {
	src := "@rarity common\n@set AAA\n---\n@deftemplate t\n@rarity rare\n---\n@template t\n@set BBB\nX\nType\n"
	d, err := Parse(strings.NewReader(src), "t.mpc")
	if err != nil {
		t.Fatal(err)
	}
	c := d.Cards[0]
	if c.Get("rarity") != "rare" {
		t.Errorf("template should beat deck default, got %q", c.Get("rarity"))
	}
	if c.Get("set") != "BBB" {
		t.Errorf("card should beat deck default, got %q", c.Get("set"))
	}
}

func TestUnknownTemplate(t *testing.T) {
	_, err := Parse(strings.NewReader("---\n@template nope\nX\nType\n"), "t.mpc")
	if err == nil || !strings.Contains(err.Error(), "unknown @template") {
		t.Fatalf("err = %v", err)
	}
}

func TestDirectiveForms(t *testing.T) {
	for _, in := range []string{"@set WED", "@set: WED", "@set=WED", "@set:WED"} {
		k, v := splitDirective(in)
		if k != "set" || v != "WED" {
			t.Errorf("%q -> %q,%q", in, k, v)
		}
	}
	if k, v := splitDirective("@legendary"); k != "legendary" || v != "" {
		t.Errorf("bare -> %q,%q", k, v)
	}
}
