package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 64
	screenHeight = 32
	scale        = 10
)

type Game struct {
	// display is the CHIP-8 framebuffer; true means the pixel is on.
	display [screenHeight][screenWidth]bool
	pixels  []byte
}

// Update runs at 60 ticks per second, which matches the CHIP-8 timer rate.
func (g *Game) Update() error {
	// TODO: run CPU cycles, decrement timers, read keys.
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for y := range screenHeight {
		for x := range screenWidth {
			var c byte
			if g.display[y][x] {
				c = 0xff
			}
			i := (y*screenWidth + x) * 4
			g.pixels[i], g.pixels[i+1], g.pixels[i+2], g.pixels[i+3] = c, c, c, 0xff
		}
	}

	screen.WritePixels(g.pixels)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	g := &Game{pixels: make([]byte, screenWidth*screenHeight*4)}

	ebiten.SetWindowSize(screenWidth*scale, screenHeight*scale)
	ebiten.SetWindowTitle("CHIP-8")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
