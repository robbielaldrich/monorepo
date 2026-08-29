// Command mpc turns a .mpc markup file into print-ready card images.
//
// Rendering is done by Card Conjurer (https://github.com/Investigamer/cardconjurer),
// created by Kyle Burton, driven headlessly. See mpc/ATTRIBUTION.md.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/robbielaldrich/monorepo/mpc/internal/deck"
	"github.com/robbielaldrich/monorepo/mpc/internal/print"
	"github.com/robbielaldrich/monorepo/mpc/internal/render"
)

const usage = `mpc — build Make Playing Cards uploads from a .mpc file

usage:
  mpc build <file.mpc> [flags]   render every card in the file
  mpc check <file.mpc>           parse only; report what would be rendered
  mpc serve [-art DIR]           run Card Conjurer in your own browser
  mpc fetch                      clone/update the Card Conjurer checkout
  mpc frames                     list the @frame styles available

build flags:
  -o DIR         output directory (default: alongside the .mpc file, in ./out)
  -art DIR       art directory (default: @art-dir, else the .mpc file's dir)
  -cc DIR        Card Conjurer checkout (default: $MPC_CARDCONJURER or the cache)
  -only NAME     render only cards whose name contains NAME (repeatable)
  -no-bleed      skip the MPC bleed images
  -no-save       skip the .json Card Conjurer save files
  -bleed IN      bleed per edge in inches (default 0.12, MPC's poker template)
  -headed        show the browser, for debugging a card that renders wrong
  -chrome PATH   Chromium/Chrome binary to use
  -v             verbose

Cards go to DIR/cards (preview, rounded corners) and DIR/mpc (upload, with bleed).
`

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "mpc: "+err.Error())
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		return errors.New("no command given")
	}
	switch os.Args[1] {
	case "build":
		return build(os.Args[2:])
	case "check":
		return check(os.Args[2:])
	case "fetch":
		return fetch(os.Args[2:])
	case "serve":
		return serve(os.Args[2:])
	case "frames":
		for _, f := range render.Frames {
			fmt.Println(f)
		}
		return nil
	case "-h", "--help", "help":
		fmt.Print(usage)
		return nil
	default:
		fmt.Print(usage)
		return fmt.Errorf("unknown command %q", os.Args[1])
	}
}

// splitArgs pulls positional arguments out from between flags, so that both
// `mpc build deck.mpc -v` and `mpc build -v deck.mpc` work.
func splitArgs(args []string, valueFlags map[string]bool) (flags, positional []string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--":
			return flags, append(positional, args[i+1:]...)
		case strings.HasPrefix(a, "-") && a != "-":
			flags = append(flags, a)
			name := strings.TrimLeft(a, "-")
			if !strings.Contains(a, "=") && valueFlags[name] && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		default:
			positional = append(positional, a)
		}
	}
	return flags, positional
}

type stringList []string

func (s *stringList) String() string     { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error { *s = append(*s, v); return nil }

func fetch(args []string) error {
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	dir := fs.String("cc", render.CacheDir(), "checkout directory")
	fs.Parse(args)

	fmt.Fprintf(os.Stderr, "fetching Card Conjurer into %s\n", *dir)
	if err := render.Fetch(*dir, os.Stderr); err != nil {
		return err
	}
	if err := render.Check(*dir); err != nil {
		return err
	}
	fmt.Fprint(os.Stderr, "\nCard Conjurer was created by Kyle Burton and is maintained at\n"+
		"github.com/Investigamer/cardconjurer. Its frames, fonts and mana symbols are\n"+
		"not redistributed by this tool. See mpc/ATTRIBUTION.md before you print.\n")
	return nil
}

func serve(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	ccDir := fs.String("cc", render.CacheDir(), "Card Conjurer checkout")
	artDir := fs.String("art", ".", "directory to expose at /art/")
	port := fs.Int("port", 4242, "port to listen on")
	fs.Parse(args)

	srv, err := render.ServeOn(*ccDir, *artDir, *port)
	if err != nil {
		return err
	}
	defer srv.Close()
	fmt.Printf("Card Conjurer at %s\n", srv.Addr)
	fmt.Printf("art served from %s at %s/art/\n", *artDir, srv.Addr)
	fmt.Println("load a file from out/cardconjurer/ under Import/Save to hand-tweak a card")
	fmt.Println("ctrl-c to stop")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	<-ctx.Done()
	return nil
}

func check(args []string) error {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	flags, pos := splitArgs(args, nil)
	fs.Parse(flags)
	if len(pos) != 1 {
		return errors.New("check needs exactly one .mpc file")
	}
	d, err := deck.ParseFile(pos[0])
	if err != nil {
		return err
	}
	artBase := "http://localhost/art"
	for _, c := range d.Cards {
		spec, err := render.SpecFor(c, artBase)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n  %-10s %s\n  frame    %s\n  art      %s\n",
			c.Name, c.Cost, c.Type, spec.Frame, or(c.Get("art"), "(none)"))
		if pt := c.PT(); pt != "" {
			fmt.Printf("  pt       %s\n", pt)
		}
		if c.Rules != "" {
			fmt.Printf("  rules    %s\n", strings.ReplaceAll(c.Rules, "\n", "\n           "))
		}
		fmt.Printf("  out      %s.png\n\n", c.OutName())
	}
	fmt.Printf("%d card(s), %d template(s)\n", len(d.Cards), len(d.Templates))
	return nil
}

func build(args []string) error {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	out := fs.String("o", "", "output directory")
	artDir := fs.String("art", "", "art directory")
	ccDir := fs.String("cc", render.CacheDir(), "Card Conjurer checkout")
	bleedIn := fs.Float64("bleed", print.BleedIn, "bleed per edge, inches")
	noBleed := fs.Bool("no-bleed", false, "skip MPC bleed images")
	noSave := fs.Bool("no-save", false, "skip Card Conjurer save files")
	headed := fs.Bool("headed", false, "show the browser")
	chrome := fs.String("chrome", "", "Chromium binary")
	verbose := fs.Bool("v", false, "verbose")
	var only stringList
	fs.Var(&only, "only", "render only cards whose name contains this")

	flags, pos := splitArgs(args, map[string]bool{
		"o": true, "art": true, "cc": true, "bleed": true,
		"only": true, "chrome": true,
	})
	fs.Parse(flags)
	if len(pos) != 1 {
		return errors.New("build needs exactly one .mpc file")
	}
	src := pos[0]

	d, err := deck.ParseFile(src)
	if err != nil {
		return err
	}
	cards := filterCards(d.Cards, only)
	if len(cards) == 0 {
		return errors.New("no cards to render")
	}

	srcDir := filepath.Dir(src)
	art := pick(*artDir, resolveDir(srcDir, d.Defaults.Get("art-dir")), srcDir)
	outDir := pick(*out, resolveDir(srcDir, d.Defaults.Get("out-dir")), filepath.Join(srcDir, "out"))

	if err := render.Check(*ccDir); err != nil {
		return err
	}

	srv, err := render.Serve(*ccDir, art)
	if err != nil {
		return err
	}
	defer srv.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	opts := render.Options{ExecPath: *chrome, Headed: *headed}
	if *verbose {
		opts.Log = func(f string, a ...any) { fmt.Fprintf(os.Stderr, f+"\n", a...) }
	}
	fmt.Fprintf(os.Stderr, "starting Card Conjurer (%s)\n", *ccDir)
	r, err := render.New(ctx, srv, opts)
	if err != nil {
		return err
	}
	defer r.Close()

	cardsOut := filepath.Join(outDir, "cards")
	mpcOut := filepath.Join(outDir, "mpc")
	saveOut := filepath.Join(outDir, "cardconjurer")
	for _, dir := range []string{cardsOut, mpcOut, saveOut} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	start := time.Now()
	manifest := make([][2]string, 0, len(cards))
	for i, c := range cards {
		spec, err := render.SpecFor(c, srv.Addr+"/art")
		if err != nil {
			return err
		}
		if spec.Art != "" {
			if err := checkArt(art, c.Get("art")); err != nil {
				return fmt.Errorf("%s:%d: %w", src, c.Line, err)
			}
		}

		fmt.Fprintf(os.Stderr, "[%d/%d] %s\n", i+1, len(cards), c.Name)
		// Card Conjurer's terms make crediting the artist the card maker's
		// responsibility, and its own download button refuses without it.
		if spec.Art != "" && spec.Artist == "" {
			fmt.Fprintf(os.Stderr, "        warning: no @artist credit for %s\n", c.Get("art"))
		}
		res, err := r.Render(spec)
		if err != nil {
			return err
		}
		for _, w := range res.Warnings {
			fmt.Fprintf(os.Stderr, "        warning: %s\n", w)
		}

		name := c.OutName()
		if err := os.WriteFile(filepath.Join(cardsOut, name+".png"), res.Rounded, 0o644); err != nil {
			return err
		}
		if !*noSave && len(res.Save) > 0 {
			if err := os.WriteFile(filepath.Join(saveOut, name+".json"), res.Save, 0o644); err != nil {
				return err
			}
		}
		if !*noBleed {
			if err := writeBleed(filepath.Join(mpcOut, name+".png"), res.Square, *bleedIn); err != nil {
				return err
			}
			manifest = append(manifest, [2]string{name + ".png", c.Get("copies")})
		}
	}

	if !*noBleed {
		if err := writeManifest(filepath.Join(mpcOut, "manifest.csv"), manifest); err != nil {
			return err
		}
	}

	g := print.GeometryFor(2010, 2814, *bleedIn)
	fmt.Fprintf(os.Stderr, "\n%d card(s) in %s\n  previews  %s\n", len(cards),
		time.Since(start).Round(time.Millisecond), cardsOut)
	if !*noBleed {
		fmt.Fprintf(os.Stderr, "  upload    %s  (%dx%d, %.0f DPI, %.2fin bleed per edge)\n",
			mpcOut, g.OutW, g.OutH, g.DPI, *bleedIn)
	}
	if !*noSave {
		fmt.Fprintf(os.Stderr, "  editable  %s  (load in Card Conjurer to hand-tweak)\n", saveOut)
	}
	return nil
}

// writeManifest records how many of each card to order, which is the one bit
// of the MPC upload that the images themselves cannot carry.
func writeManifest(path string, rows [][2]string) error {
	var b strings.Builder
	b.WriteString("file,copies\n")
	for _, r := range rows {
		copies := r[1]
		if copies == "" {
			copies = "1"
		}
		fmt.Fprintf(&b, "%q,%s\n", r[0], copies)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func writeBleed(path string, cardPNG []byte, bleedIn float64) error {
	img, err := png.Decode(bytes.NewReader(cardPNG))
	if err != nil {
		return fmt.Errorf("decoding rendered card: %w", err)
	}
	out := print.AddBleed(img, bleedIn, color.Black)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, out)
}

func checkArt(artDir, ref string) error {
	if artDir == "" || ref == "" || strings.Contains(ref, "://") || strings.HasPrefix(ref, "/img/") {
		return nil
	}
	if _, err := os.Stat(filepath.Join(artDir, ref)); err != nil {
		return fmt.Errorf("@art %s not found in %s", ref, artDir)
	}
	return nil
}

func filterCards(cards []*deck.Card, only stringList) []*deck.Card {
	if len(only) == 0 {
		return cards
	}
	var out []*deck.Card
	for _, c := range cards {
		for _, want := range only {
			if strings.Contains(strings.ToLower(c.Name), strings.ToLower(want)) {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

func resolveDir(base, rel string) string {
	if rel == "" {
		return ""
	}
	if strings.HasPrefix(rel, "~/") {
		rel = filepath.Join(os.Getenv("HOME"), rel[2:])
	}
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(base, rel)
}

func pick(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func or(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
