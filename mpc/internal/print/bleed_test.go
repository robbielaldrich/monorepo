package print

import (
	"image"
	"image/color"
	"testing"
)

func TestGeometryMatchesMPCTemplate(t *testing.T) {
	// MPC's 300 DPI poker template is 822x1122 with the card at 750x1050.
	g := GeometryFor(750, 1050, BleedIn)
	if g.OutW != 822 || g.OutH != 1122 {
		t.Fatalf("got %dx%d, want 822x1122", g.OutW, g.OutH)
	}
	if g.DPI != 300 {
		t.Errorf("dpi = %v, want 300", g.DPI)
	}
}

func TestGeometryAtCardConjurerResolution(t *testing.T) {
	g := GeometryFor(2010, 2814, BleedIn)
	if g.MarginX != 96 || g.MarginY != 96 {
		t.Fatalf("margins = %d,%d, want 96,96", g.MarginX, g.MarginY)
	}
	if g.OutW != 2202 || g.OutH != 3006 {
		t.Fatalf("got %dx%d, want 2202x3006", g.OutW, g.OutH)
	}
	// The bleed canvas must keep MPC's aspect ratio or the print is stretched.
	want := (CardWidthIn + 2*BleedIn) / (CardHeightIn + 2*BleedIn)
	got := float64(g.OutW) / float64(g.OutH)
	if diff := got - want; diff > 0.001 || diff < -0.001 {
		t.Errorf("aspect = %v, want %v", got, want)
	}
}

func TestAddBleedReplicatesEdges(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 750, 1050))
	red := color.RGBA{255, 0, 0, 255}
	for y := 0; y < 1050; y++ {
		for x := 0; x < 750; x++ {
			src.SetRGBA(x, y, red)
		}
	}
	// A distinctive top-left pixel proves the corner is filled from the
	// nearest card pixel rather than left blank.
	blue := color.RGBA{0, 0, 255, 255}
	src.SetRGBA(0, 0, blue)

	out := AddBleed(src, BleedIn, color.Black)
	g := GeometryFor(750, 1050, BleedIn)
	if out.Bounds().Dx() != g.OutW || out.Bounds().Dy() != g.OutH {
		t.Fatalf("size = %v", out.Bounds())
	}
	if got := out.RGBAAt(0, 0); got != blue {
		t.Errorf("corner bleed = %v, want %v", got, blue)
	}
	if got := out.RGBAAt(g.OutW/2, 0); got != red {
		t.Errorf("top bleed = %v, want %v", got, red)
	}
	if got := out.RGBAAt(g.OutW-1, g.OutH-1); got != red {
		t.Errorf("bottom-right bleed = %v, want %v", got, red)
	}
	if got := out.RGBAAt(g.MarginX, g.MarginY); got != blue {
		t.Errorf("card origin = %v, want %v", got, blue)
	}
}

func TestAddBleedFlattensTransparency(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 750, 1050)) // fully transparent
	out := AddBleed(src, BleedIn, color.Black)
	if got := out.RGBAAt(400, 500); got != (color.RGBA{0, 0, 0, 255}) {
		t.Errorf("got %v, want opaque black", got)
	}
}

func TestRoundCorners(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 750, 1050))
	for y := 0; y < 1050; y++ {
		for x := 0; x < 750; x++ {
			src.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	out := RoundCorners(src, CornerRadiusIn)
	if a := out.RGBAAt(0, 0).A; a != 0 {
		t.Errorf("corner alpha = %d, want 0", a)
	}
	if a := out.RGBAAt(375, 525).A; a != 255 {
		t.Errorf("centre alpha = %d, want 255", a)
	}
	if a := out.RGBAAt(0, 525).A; a != 255 {
		t.Errorf("mid-edge alpha = %d, want 255", a)
	}
	if a := out.RGBAAt(749, 1049).A; a != 0 {
		t.Errorf("far corner alpha = %d, want 0", a)
	}
}
