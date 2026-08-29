# mpc

Write cards in a small text format, get print-ready images for
[MakePlayingCards](https://www.makeplayingcards.com).

```
$ mpc fetch                          # one-off: get Card Conjurer
$ mpc build examples/wedding.mpc
[1/3] Sir Whiskers, Snack Thief
[2/3] Afternoon Sunbeam
[3/3] Open Bar

3 card(s) in 30.7s
  previews  examples/out/cards
  upload    examples/out/mpc  (2202x3006, 804 DPI, 0.12in bleed per edge)
  editable  examples/out/cardconjurer  (load in Card Conjurer to hand-tweak)
```

Rendering is done by **Card Conjurer**, created by Kyle Burton — see
[ATTRIBUTION.md](ATTRIBUTION.md), which you should read before you print
anything.

## How it works

Card Conjurer is a browser app with no CLI, so `mpc build` drives the real
thing: it serves a local checkout on loopback, opens it in headless Chromium,
fills in the fields, and reads the finished canvas back. You get its exact
output — legend crowns, hybrid two-colour frames, inline mana symbols,
shrink-to-fit rules text, flavour bars — rather than an approximation.

Each card comes out three ways:

| directory | what it is |
| --- | --- |
| `out/cards/` | 2010x2814, rounded transparent corners. For looking at. |
| `out/mpc/` | 2202x3006, square corners, 0.12in bleed on every edge. **This is what you upload.** |
| `out/cardconjurer/` | A Card Conjurer save file. Load it in the app to hand-tweak one card. |

`out/mpc/manifest.csv` lists each file and how many copies to order.

### Why the bleed step matters

MPC prints onto a sheet larger than the finished card and then cuts it down. If
you upload an image that is exactly card-sized, MPC scales it up to cover the
bleed area and you lose a strip of border on every edge — the black border ends
up uneven and the corners cut inside the frame.

So `out/mpc/` puts the card at its true 2.5in x 3.5in size on a 2.74in x 3.74in
canvas, fills the margin by extending the card's own edge pixels outward, and
squares off the corners (a transparent corner prints white; MPC rounds them with
a die anyway).

## The markup

A `.mpc` file is blocks separated by a line of `---`. `#` starts a comment.

Lines beginning with `@` are directives. Everything else is card text, read in
the order a Magic card reads:

```
Name {mana cost}
Type — Subtype
rules text, one line per paragraph
> flavor text
power/toughness
```

Every part is optional except the name.

```
# Deck-level defaults. Every card inherits these.
@art-dir   art
@out-dir   out
@set       WED
@artist    Robert Aldrich
@copyright Robert & Bibbo • 2026

---
# A reusable template; cards opt in with "@template legend".
@deftemplate legend
@frame  regular
@rarity mythic
@copies 2

---
@template legend
@art     IMG_2456.JPEG
@number  001

Sir Whiskers, Snack Thief {2}{G}{W}
Legendary Creature — Cat Noble

Vigilance
When ~ enters the battlefield, create a Food token.
{T}, Sacrifice a Food: Draw a card.
> That was never yours to begin with.
3/4
```

Precedence is **deck defaults < template < the card's own directives**.

### Directives

Written `@key value`; `@key: value` and `@key=value` also work.

**Deck-level** (in the first block)

| | |
| --- | --- |
| `@art-dir DIR` | where `@art` filenames are resolved, relative to the `.mpc` file |
| `@out-dir DIR` | where images are written (default `./out`) |
| `@include FILE` | pull in another `.mpc` file's defaults, templates and cards |

**Anywhere**

| | |
| --- | --- |
| `@frame NAME` | frame style — see `mpc frames`. Aliases: `regular`, `accurate`, `fullart`, `borderless`, `extended`, `etched`, `phyrexian`, `ub`, `8th`, `seventh`, `circuit` |
| `@art FILE` | art file inside `@art-dir`, or an `http(s)://` URL |
| `@art-fit cover\|contain` | fill the art box (default) or letterbox inside it |
| `@art-zoom N` | zoom percent; switches to manual placement |
| `@art-x N` / `@art-y N` | manual position, as a fraction of the card |
| `@art-scale N` | multiply the fitted zoom, e.g. `1.15` to punch in a little |
| `@art-nudge-x N` / `@art-nudge-y N` | shift the fitted art, as a fraction of the card |
| `@art-rotate DEG` | rotate the art |
| `@artist NAME` | artist credit. **Fill this in** — see ATTRIBUTION.md |
| `@set CODE` / `@number N` / `@rarity R` | collector line. Rarity takes `common`/`uncommon`/`rare`/`mythic` or a letter |
| `@language`, `@year`, `@note` | rest of the collector line |
| `@copyright TEXT` | replaces the Wizards line at bottom right; `none` removes it |
| `@no-collector` | drop the bottom collector line entirely |
| `@set-symbol CODE` | a real set's symbol, e.g. `dom` |
| `@set-symbol-image FILE` | your own set symbol |
| `@watermark FILE` / `@watermark-opacity N` | watermark behind the rules text |
| `@nickname TEXT` | nickname bar, on frames that have one |
| `@copies N` | how many to order; ends up in `manifest.csv` |
| `@out NAME` | output filename, instead of `<number> <name>` |
| `@name`, `@cost`, `@type`, `@pt` | override what was parsed from the card text |
| `@template NAME` | apply a template |
| `@deftemplate NAME` | makes this block a template rather than a card |

### Text markup

`*asterisks*` become italics. Otherwise the rules text is passed to Card
Conjurer as-is, so its own codes work:

| | |
| --- | --- |
| `{W} {U} {B} {R} {G} {C} {T} {2}` … | mana and tap symbols |
| `~` or `{cardname}` | the card's own name |
| `{i}` … `{/i}` | italics |
| `{flavor}` | flavour bar (a `>` line emits this for you) |
| `{-}` | em dash |
| `//` | line break within a paragraph |
| `{divider}` | horizontal rule |

Colours and the frame layout are chosen from the mana cost and type line, the
same way Card Conjurer's "Automatic Frame" does it. `Legendary` in the type line
gets a crown; two colours get a split frame; `Snow` and `Enchantment Creature`
get their own treatments. To change any of that, use `@frame`.

## Commands

```
mpc build <file.mpc>    render every card in the file
mpc check <file.mpc>    parse only; print what would be rendered
mpc serve               run Card Conjurer in your own browser
mpc fetch               clone/update the Card Conjurer checkout
mpc frames              list the @frame styles
```

Useful `build` flags: `-only NAME` to re-render one card, `-headed` to watch the
browser do it, `-o DIR`, `-bleed IN`, `-no-bleed`, `-chrome PATH`.

## Hand-tweaking a card

When a card needs manual attention, `mpc build` has already written a save file
for it:

```
$ mpc serve -art examples/art
Card Conjurer at http://127.0.0.1:4242
```

Open that, go to **Import/Save**, and load
`out/cardconjurer/<card>.json`. Everything is where the generated card left it.

## Layout

```
cmd/mpc/            the CLI
internal/deck/      the .mpc markup language and its parser
internal/render/    local server, headless browser driver, injected page script
internal/print/     bleed and corner geometry for MPC uploads
examples/           a worked example
```

`internal/render/inject.js` is the interesting file: it runs inside the Card
Conjurer page and is the whole of the integration.

## Requirements

Chromium or Chrome, and `git` for `mpc fetch`. Nothing else.
