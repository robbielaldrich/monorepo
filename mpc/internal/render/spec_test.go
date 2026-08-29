package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robbielaldrich/monorepo/mpc/internal/deck"
)

func card(t *testing.T, src string) *deck.Card {
	t.Helper()
	d, err := deck.Parse(strings.NewReader(src), "t.mpc")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Cards) != 1 {
		t.Fatalf("want 1 card, got %d", len(d.Cards))
	}
	return d.Cards[0]
}

func TestResolveFrame(t *testing.T) {
	cases := map[string]string{
		"":             "M15Regular-1",
		"regular":      "M15Regular-1",
		"M15Regular-1": "M15Regular-1",
		"full art":     "FullArtNew",
		"Borderless":   "Borderless",
		"borderless":   "Borderless",
		"phyrexian":    "Praetors",
	}
	for in, want := range cases {
		got, err := resolveFrame(in)
		if err != nil || got != want {
			t.Errorf("resolveFrame(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	if _, err := resolveFrame("nonsense"); err == nil {
		t.Error("want an error for an unknown frame")
	}
}

func TestSpecForDefaults(t *testing.T) {
	c := card(t, "---\n@art cat.jpg\n@rarity mythic\nX {G}\nCreature — Cat\n2/2\n")
	s, err := SpecFor(c, "http://127.0.0.1:1234/art")
	if err != nil {
		t.Fatal(err)
	}
	if s.Frame != "M15Regular-1" || s.Pack != "packM15Regular-1" {
		t.Errorf("frame/pack = %q/%q", s.Frame, s.Pack)
	}
	if s.Rarity != "M" {
		t.Errorf("rarity = %q, want M", s.Rarity)
	}
	if s.Art != "http://127.0.0.1:1234/art/cat.jpg" {
		t.Errorf("art = %q", s.Art)
	}
	if s.ArtFit != "cover" {
		t.Errorf("artFit = %q, want cover", s.ArtFit)
	}
	if s.Language != "EN" {
		t.Errorf("language = %q, want EN", s.Language)
	}
	if !s.Collector {
		t.Error("collector info should default on")
	}
}

func TestAssetURL(t *testing.T) {
	const base = "http://h/art"
	cases := map[string]string{
		"cat.jpg":           "http://h/art/cat.jpg",
		"sub dir/cat 2.jpg": "http://h/art/sub%20dir/cat%202.jpg",
		"https://x/y.png":   "https://x/y.png",
		"/img/blank.png":    "/img/blank.png",
	}
	for in, want := range cases {
		if got := assetURL(in, base); got != want {
			t.Errorf("assetURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestManualArtPlacement(t *testing.T) {
	c := card(t, "---\n@art cat.jpg\n@art-zoom 145\n@art-x 0.1\n@art-y -0.05\nX\nCreature\n")
	s, err := SpecFor(c, "http://h/art")
	if err != nil {
		t.Fatal(err)
	}
	if s.ArtFit != "manual" {
		t.Fatalf("artFit = %q, want manual", s.ArtFit)
	}
	if s.ArtZoom != 1.45 || s.ArtX != 0.1 || s.ArtY != -0.05 {
		t.Errorf("zoom/x/y = %v/%v/%v", s.ArtZoom, s.ArtX, s.ArtY)
	}
}

func TestNoCollector(t *testing.T) {
	c := card(t, "---\n@no-collector\nX\nCreature\n")
	s, _ := SpecFor(c, "http://h/art")
	if s.Collector {
		t.Error("@no-collector should turn the collector line off")
	}
}

// Every frame the CLI advertises must map to a pack script Card Conjurer
// actually ships, or the render fails deep inside the browser.
func TestFramesAreAllResolvable(t *testing.T) {
	for _, f := range Frames {
		got, err := resolveFrame(f)
		if err != nil || got != f {
			t.Errorf("resolveFrame(%q) = %q, %v", f, got, err)
		}
	}
}

// If a Card Conjurer checkout is around, confirm each advertised frame has a
// pack script. This is the failure that otherwise only shows up mid-render.
func TestFramePacksExist(t *testing.T) {
	dir := CacheDir()
	if err := Check(dir); err != nil {
		t.Skipf("no Card Conjurer checkout: %v", err)
	}
	for _, f := range Frames {
		path := filepath.Join(dir, "js", "frames", "pack"+f+".js")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("frame %q: %v", f, err)
		}
	}
}
