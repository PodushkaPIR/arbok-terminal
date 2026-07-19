package ui

import (
	"image/color"

	"arbok-terminal/internal/terminal"
)

func ColorToRGBA(c terminal.Color, fallback color.Color) color.RGBA {
	if c.Default {
		r, g, b, _ := fallback.RGBA()
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
	}
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: 255}
}
