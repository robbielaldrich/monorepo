// Package print prepares rendered cards for MakePlayingCards.
//
// MPC prints onto a sheet slightly larger than the finished card and then cuts
// it down, so an upload that is exactly card-sized gets scaled up to cover the
// bleed area and loses a strip of border on every edge. The fix is to hand MPC
// an image that already includes the bleed: the artwork at its true size,
// centred on a larger canvas whose margin is filled from the card's own edge
// pixels.
package print

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// Standard poker-card geometry, in inches.
const (
	CardWidthIn  = 2.5
	CardHeightIn = 3.5
	// BleedIn is the margin MPC adds on each edge: their 300 DPI poker
	// template is 822x1122 px, i.e. 2.74in x 3.74in.
	BleedIn = 0.12
	// CornerRadiusIn is the finished card's corner radius.
	CornerRadiusIn = 0.125
)

// Geometry describes the bleed canvas for a card image of the given pixel size.
type Geometry struct {
	CardW, CardH int
	OutW, OutH   int
	MarginX      int
	MarginY      int
	DPI          float64
}

// GeometryFor computes the bleed canvas for a w x h card image, assuming the
// image covers exactly the 2.5in x 3.5in finished card.
func GeometryFor(w, h int, bleedIn float64) Geometry {
	dpi := float64(w) / CardWidthIn
	mx := int(math.Round(bleedIn * dpi))
	my := int(math.Round(bleedIn * (float64(h) / CardHeightIn)))
	return Geometry{
		CardW: w, CardH: h,
		OutW: w + 2*mx, OutH: h + 2*my,
		MarginX: mx, MarginY: my,
		DPI: dpi,
	}
}

// AddBleed centres src on a larger canvas and fills the margin by extending
// src's outermost pixels outwards. Any transparency in src is flattened onto
// bg first, since MPC prints on opaque stock and a transparent corner would
// come back white.
func AddBleed(src image.Image, bleedIn float64, bg color.Color) *image.RGBA {
	b := src.Bounds()
	g := GeometryFor(b.Dx(), b.Dy(), bleedIn)

	flat := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(flat, flat.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
	draw.Draw(flat, flat.Bounds(), src, b.Min, draw.Over)

	out := image.NewRGBA(image.Rect(0, 0, g.OutW, g.OutH))
	draw.Draw(out, image.Rect(g.MarginX, g.MarginY, g.MarginX+g.CardW, g.MarginY+g.CardH),
		flat, image.Point{}, draw.Src)

	// Edge-replicate into the four margins, then the four corners.
	for y := 0; y < g.OutH; y++ {
		sy := clamp(y-g.MarginY, 0, g.CardH-1)
		for x := 0; x < g.MarginX; x++ {
			out.Set(x, y, flat.At(0, sy))
			out.Set(g.OutW-1-x, y, flat.At(g.CardW-1, sy))
		}
	}
	for x := 0; x < g.OutW; x++ {
		sx := clamp(x-g.MarginX, 0, g.CardW-1)
		for y := 0; y < g.MarginY; y++ {
			out.Set(x, y, flat.At(sx, 0))
			out.Set(x, g.OutH-1-y, flat.At(sx, g.CardH-1))
		}
	}
	return out
}

// RoundCorners punches transparent rounded corners into a card-sized image,
// for previews. Print files keep square corners: MPC does the rounding with a
// die, and a transparent corner in an upload prints white.
func RoundCorners(src image.Image, radiusIn float64) *image.RGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(out, out.Bounds(), src, b.Min, draw.Src)

	r := radiusIn * (float64(w) / CardWidthIn)
	fw, fh := float64(w), float64(h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			px, py := float64(x), float64(y)
			// Only the four corner boxes can fall outside the rounded rect.
			if (px > r && px < fw-r) || (py > r && py < fh-r) {
				continue
			}
			cx, cy := r, r
			if px > fw/2 {
				cx = fw - r
			}
			if py > fh/2 {
				cy = fh - r
			}
			if a := coverage(px, py, cx, cy, r); a < 1 {
				c := out.RGBAAt(x, y)
				out.SetRGBA(x, y, color.RGBA{
					R: uint8(float64(c.R) * a),
					G: uint8(float64(c.G) * a),
					B: uint8(float64(c.B) * a),
					A: uint8(float64(c.A) * a),
				})
			}
		}
	}
	return out
}

// coverage supersamples one pixel against a circle to keep the corner arc from
// looking stair-stepped.
func coverage(px, py, cx, cy, r float64) float64 {
	const n = 4
	inside := 0
	for sy := 0; sy < n; sy++ {
		for sx := 0; sx < n; sx++ {
			ox := px + (float64(sx)+0.5)/n
			oy := py + (float64(sy)+0.5)/n
			if math.Hypot(ox-cx, oy-cy) <= r {
				inside++
			}
		}
	}
	return float64(inside) / (n * n)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
