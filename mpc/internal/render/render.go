package render

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

//go:embed inject.js
var injectJS string

// Result is one rendered card.
type Result struct {
	// Rounded is the card as Card Conjurer draws it, with transparent
	// rounded corners. Square is the same card with the corner cutout
	// skipped, which is what the print bleed step wants.
	Rounded  []byte
	Square   []byte
	Width    int
	Height   int
	Save     json.RawMessage // a save file the Card Conjurer UI can load
	Warnings []string
}

// Renderer owns one headless browser tab pointed at a running Card Conjurer.
type Renderer struct {
	alloc  context.CancelFunc
	cancel context.CancelFunc
	ctx    context.Context
	srv    *Server
	log    func(string, ...any)
}

// Options configure the browser.
type Options struct {
	// ExecPath overrides the Chromium binary chromedp would pick.
	ExecPath string
	// Headed shows the browser window, which is useful when a card renders
	// wrong and you want to poke at it.
	Headed bool
	// Timeout bounds a single card render.
	Timeout time.Duration
	// Verbose forwards browser console output to Log.
	Log func(string, ...any)
}

// New starts a browser and loads the creator page served by srv.
func New(parent context.Context, srv *Server, opt Options) (*Renderer, error) {
	baseURL := srv.Addr
	if opt.Timeout == 0 {
		opt.Timeout = 90 * time.Second
	}

	flags := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	flags = append(flags,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		// The creator sizes some inputs off the viewport; give it a real one.
		chromedp.WindowSize(1400, 1000),
	)
	if opt.ExecPath != "" {
		flags = append(flags, chromedp.ExecPath(opt.ExecPath))
	}
	if opt.Headed {
		flags = append(flags, chromedp.Flag("headless", false))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(parent, flags...)

	var ctxOpts []chromedp.ContextOption
	if opt.Log != nil {
		ctxOpts = append(ctxOpts, chromedp.WithLogf(opt.Log))
	}
	ctx, cancel := chromedp.NewContext(allocCtx, ctxOpts...)

	r := &Renderer{alloc: allocCancel, cancel: cancel, ctx: ctx, srv: srv, log: opt.Log}
	r.watchPage(ctx)

	// The first Run allocates the browser and binds its lifetime to the
	// context it is given, so it has to be the long-lived one. Doing this as
	// a bare Run means the boot steps below can safely carry a deadline.
	if err := chromedp.Run(ctx); err != nil {
		r.Close()
		return nil, fmt.Errorf("starting browser: %w", err)
	}

	boot, bootCancel := context.WithTimeout(ctx, 90*time.Second)
	defer bootCancel()

	err := chromedp.Run(boot,
		// Installed before the creator's own scripts so it can observe every
		// image load the app starts.
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(injectJS).Do(ctx)
			return err
		}),
		// creator/index.html is an htmx fragment, not a page: it has no
		// <head> and relies on the shell at / having already loaded
		// main-1.js. Loading it directly leaves the creator's own script
		// calling functions that do not exist yet.
		chromedp.Navigate(baseURL+"/"),
		chromedp.WaitReady(`#content`, chromedp.ByQuery),
		// localStorage is only reachable from the server's own origin, and
		// the creator reads these as it boots, so set them before the swap.
		chromedp.Evaluate(`
			localStorage.setItem('autoLoadFrameVersion', 'true');
			localStorage.setItem('enableCollectorInfo', 'true');
			localStorage.setItem('autoFit', 'true');
			localStorage.setItem('lockSetSymbolCode', '');
			localStorage.setItem('lockSetSymbolURL', '');
		`, nil),
		chromedp.Evaluate(`htmx.ajax('GET', '/creator/index.html', '#content')`, nil, awaitPromise),
		chromedp.WaitReady(`#loadFrameVersion`, chromedp.ByQuery),
		chromedp.Evaluate(`window.__mpc.waitForCreator()
			.then(() => window.__mpc.settle(400))
			.then(() => window.__mpc.loadFonts())`, nil, awaitPromise),
	)
	if err != nil {
		r.Close()
		return nil, fmt.Errorf("starting Card Conjurer: %w", err)
	}
	return r, nil
}

// watchPage surfaces page-side failures, which otherwise show up in Go only
// as an opaque "context canceled" when the tab dies.
func (r *Renderer) watchPage(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev any) {
		switch e := ev.(type) {
		case *runtime.EventExceptionThrown:
			r.note("page exception: %s", e.ExceptionDetails.Error())
		case *runtime.EventConsoleAPICalled:
			if r.log == nil && e.Type != "error" && e.Type != "warning" {
				return
			}
			var parts []string
			for _, a := range e.Args {
				parts = append(parts, strings.Trim(string(a.Value), `"`))
			}
			r.note("console.%s: %s", e.Type, strings.Join(parts, " "))
		}
	})
}

func (r *Renderer) note(format string, a ...any) {
	if r.log != nil {
		r.log(format, a...)
	}
}

func (r *Renderer) Close() {
	if r.cancel != nil {
		r.cancel()
	}
	if r.alloc != nil {
		r.alloc()
	}
}

func awaitPromise(p *runtime.EvaluateParams) *runtime.EvaluateParams {
	return p.WithAwaitPromise(true)
}

type rawResult struct {
	Width    int             `json:"width"`
	Height   int             `json:"height"`
	Save     json.RawMessage `json:"save"`
	Warnings []string        `json:"warnings"`
}

// Render draws one card.
func (r *Renderer) Render(spec *Spec) (*Result, error) {
	payload, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(r.ctx, 120*time.Second)
	defer cancel()

	r.srv.Reset()

	var raw rawResult
	expr := fmt.Sprintf(`window.__mpc.render(%s)`, string(payload))
	if err := chromedp.Run(ctx, chromedp.Evaluate(expr, &raw, awaitPromise)); err != nil {
		return nil, fmt.Errorf("rendering %q: %w", spec.Title, err)
	}

	rounded, ok := r.srv.Take("rounded")
	if !ok {
		return nil, fmt.Errorf("rendering %q: browser did not return an image", spec.Title)
	}
	square, ok := r.srv.Take("square")
	if !ok {
		return nil, fmt.Errorf("rendering %q: browser did not return a print image", spec.Title)
	}
	return &Result{
		Rounded:  rounded,
		Square:   square,
		Width:    raw.Width,
		Height:   raw.Height,
		Save:     raw.Save,
		Warnings: raw.Warnings,
	}, nil
}
