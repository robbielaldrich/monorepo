package render

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/robbielaldrich/monorepo/mpc/internal/deck"
)

// Spec is the flat description of one card handed to the browser.
type Spec struct {
	Frame string `json:"frame"`
	Pack  string `json:"pack"`

	Title    string `json:"title"`
	Mana     string `json:"mana"`
	Type     string `json:"type"`
	Rules    string `json:"rules"`
	PT       string `json:"pt"`
	Nickname string `json:"nickname,omitempty"`

	Art          string  `json:"art,omitempty"`
	ArtFit       string  `json:"artFit,omitempty"`
	ArtX         float64 `json:"artX,omitempty"`
	ArtY         float64 `json:"artY,omitempty"`
	ArtZoom      float64 `json:"artZoom,omitempty"`
	ArtZoomScale float64 `json:"artZoomScale,omitempty"`
	ArtNudgeX    float64 `json:"artNudgeX,omitempty"`
	ArtNudgeY    float64 `json:"artNudgeY,omitempty"`
	ArtRotate    float64 `json:"artRotate,omitempty"`

	SetSymbol        string  `json:"setSymbol,omitempty"`
	SetSymbolURL     string  `json:"setSymbolUrl,omitempty"`
	Watermark        string  `json:"watermark,omitempty"`
	WatermarkOpacity float64 `json:"watermarkOpacity,omitempty"`

	Artist    string `json:"artist,omitempty"`
	Set       string `json:"set,omitempty"`
	Rarity    string `json:"rarity,omitempty"`
	Number    string `json:"number,omitempty"`
	Language  string `json:"language,omitempty"`
	Note      string `json:"note,omitempty"`
	Year      string `json:"year,omitempty"`
	Copyright string `json:"copyright,omitempty"`
	Collector bool   `json:"collector"`
}

// Frames lists the @frame values the tool understands, mirroring Card
// Conjurer's "Automatic Frame" dropdown.
var Frames = []string{
	"M15Regular-1", "M15RegularNew", "FullArtNew", "M15BoxTopper",
	"M15ExtendedArtShort", "Borderless", "Etched", "Praetors", "UB", "UBNew",
	"8th", "Seventh", "Circuit", "M15Eighth", "M15EighthUB",
}

// frameAliases lets .mpc files say "regular" instead of "M15Regular-1".
var frameAliases = map[string]string{
	"regular":         "M15Regular-1",
	"m15":             "M15Regular-1",
	"accurate":        "M15RegularNew",
	"fullart":         "FullArtNew",
	"full-art":        "FullArtNew",
	"extended":        "M15BoxTopper",
	"extendedart":     "M15BoxTopper",
	"extendedshort":   "M15ExtendedArtShort",
	"borderless":      "Borderless",
	"etched":          "Etched",
	"phyrexian":       "Praetors",
	"ub":              "UB",
	"universesbeyond": "UB",
	"8th":             "8th",
	"eighth":          "8th",
	"seventh":         "Seventh",
	"circuit":         "Circuit",
}

func resolveFrame(name string) (string, error) {
	if name == "" {
		return "M15Regular-1", nil
	}
	for _, f := range Frames {
		if strings.EqualFold(f, name) {
			return f, nil
		}
	}
	key := strings.ToLower(strings.NewReplacer(" ", "", "_", "", "-", "").Replace(name))
	if f, ok := frameAliases[key]; ok {
		return f, nil
	}
	return "", fmt.Errorf("unknown @frame %q (try one of: %s)", name, strings.Join(Frames, ", "))
}

var rarityCodes = map[string]string{
	"common": "C", "c": "C",
	"uncommon": "U", "u": "U",
	"rare": "R", "r": "R",
	"mythic": "M", "m": "M", "mythicrare": "M",
	"special": "S", "s": "S",
	"bonus": "B", "b": "B",
	"promo": "P", "p": "P",
	"land": "L", "l": "L",
	"timeshifted": "T", "t": "T",
}

// SpecFor lowers a parsed card into the browser-facing Spec. artBase is the URL
// prefix the local server exposes the deck's art directory under.
func SpecFor(c *deck.Card, artBase string) (*Spec, error) {
	frame, err := resolveFrame(c.Get("frame"))
	if err != nil {
		return nil, fmt.Errorf("card %q: %w", c.Name, err)
	}

	rarity := c.Get("rarity")
	if code, ok := rarityCodes[strings.ToLower(strings.ReplaceAll(rarity, " ", ""))]; ok {
		rarity = code
	}

	s := &Spec{
		Frame:     frame,
		Pack:      "pack" + frame,
		Title:     c.Name,
		Mana:      c.Cost,
		Type:      c.Type,
		Rules:     c.Rules,
		PT:        c.PT(),
		Nickname:  c.Get("nickname"),
		Artist:    c.Get("artist"),
		Set:       c.Get("set"),
		Rarity:    rarity,
		Number:    c.Get("number"),
		Language:  or(c.Get("language"), "EN"),
		Note:      c.Get("note"),
		Year:      c.Get("year"),
		Copyright: c.Get("copyright"),
		Collector: !c.Has("no-collector"),

		ArtFit:       or(c.Get("art-fit"), "cover"),
		ArtZoom:      c.Float("art-zoom", 0) / 100,
		ArtZoomScale: c.Float("art-scale", 0),
		ArtNudgeX:    c.Float("art-nudge-x", 0),
		ArtNudgeY:    c.Float("art-nudge-y", 0),
		ArtRotate:    c.Float("art-rotate", 0),

		SetSymbol:        c.Get("set-symbol"),
		WatermarkOpacity: c.Float("watermark-opacity", 0) / 100,
	}
	if c.Has("art-zoom") {
		s.ArtFit = "manual"
		s.ArtX = c.Float("art-x", 0)
		s.ArtY = c.Float("art-y", 0)
	}
	if img := c.Get("set-symbol-image"); img != "" {
		s.SetSymbolURL = assetURL(img, artBase)
	}
	if wm := c.Get("watermark"); wm != "" {
		s.Watermark = assetURL(wm, artBase)
	}
	if a := c.Get("art"); a != "" {
		s.Art = assetURL(a, artBase)
	}
	return s, nil
}

// assetURL maps an @art value onto something the browser can fetch: absolute
// URLs and Card Conjurer's own /img paths pass through, everything else is
// treated as a file inside the deck's art directory.
func assetURL(ref, artBase string) string {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") ||
		strings.HasPrefix(ref, "data:") || strings.HasPrefix(ref, "/img/") {
		return ref
	}
	return strings.TrimSuffix(artBase, "/") + "/" + pathEscape(ref)
}

func pathEscape(p string) string {
	parts := strings.Split(path.Clean(p), "/")
	for i, s := range parts {
		parts[i] = url.PathEscape(s)
	}
	return strings.Join(parts, "/")
}

func or(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
