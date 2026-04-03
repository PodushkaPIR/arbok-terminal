package terminal

import (
	"github.com/gdamore/tcell/v2"
)

type Attributes struct {
	Bold      bool
	Dim       bool
	Italic    bool
	Underline bool
	Blink     bool
	Reverse   bool
	Strike    bool
}

type Cell struct {
	Char       rune
	Foreground tcell.Color
	Background tcell.Color
	Attributes Attributes
}

type Buffer struct {
	Width  int
	Height int
	Grid   [][]Cell

	CursorX int
	CursorY int
}

func NewBuffer(width, height int) *Buffer {
	grid := make([][]Cell, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			grid[y][x] = Cell{
				Char:       ' ',
				Foreground: tcell.ColorDefault,
				Background: tcell.ColorDefault,
				Attributes: Attributes{},
			}
		}
	}

	return &Buffer{
		Width:   width,
		Height:  height,
		Grid:    grid,
		CursorX: 0,
		CursorY: 0,
	}
}

func (b *Buffer) Resize(width, height int) {
	if b.Width == width && b.Height == height {
		return
	}

	newGrid := make([][]Cell, height)
	for y := 0; y < height; y++ {
		newGrid[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			if y < b.Height && x < b.Width {
				newGrid[y][x] = b.Grid[y][x]
			} else {
				newGrid[y][x] = Cell{
					Char:       ' ',
					Foreground: tcell.ColorDefault,
					Background: tcell.ColorDefault,
					Attributes: Attributes{},
				}
			}
		}
	}

	b.Width = width
	b.Height = height
	b.Grid = newGrid

	if b.CursorX >= width {
		b.CursorX = width - 1
	}
	if b.CursorY >= height {
		b.CursorY = height - 1
	}
}

func (b *Buffer) WriteChar(ch rune, fg, bg tcell.Color, attrs Attributes) {
	if b.CursorX >= b.Width {
		b.Newline()
	}

	if b.CursorY >= 0 && b.CursorY < b.Height && b.CursorX >= 0 && b.CursorX < b.Width {
		b.Grid[b.CursorY][b.CursorX] = Cell{
			Char:       ch,
			Foreground: fg,
			Background: bg,
			Attributes: attrs,
		}
	}

	b.CursorX++
}

func (b *Buffer) Newline() {
	b.CursorX = 0
	b.CursorY++
}

func (b *Buffer) MoveCursor(x, y int) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	b.CursorX = x
	b.CursorY = y
}

func (b *Buffer) MoveUp(lines int) {
	b.CursorY -= lines
	if b.CursorY < 0 {
		b.CursorY = 0
	}
}

func (b *Buffer) MoveDown(lines int) {
	b.CursorY += lines
	if b.CursorY >= b.Height {
		b.CursorY = b.Height - 1
	}
}

func (b *Buffer) MoveLeft(cols int) {
	b.CursorX -= cols
	if b.CursorX < 0 {
		b.CursorX = 0
	}
}

func (b *Buffer) MoveRight(cols int) {
	b.CursorX += cols
	if b.CursorX >= b.Width {
		b.CursorX = b.Width - 1
	}
}

func (b *Buffer) Clear() {
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			b.Grid[y][x] = Cell{
				Char:       ' ',
				Foreground: tcell.ColorDefault,
				Background: tcell.ColorDefault,
				Attributes: Attributes{},
			}
		}
	}
	b.CursorX = 0
	b.CursorY = 0
}

func (b *Buffer) ClearToEnd() {
	for x := b.CursorX; x < b.Width; x++ {
		b.Grid[b.CursorY][x] = Cell{
			Char:       ' ',
			Foreground: tcell.ColorDefault,
			Background: tcell.ColorDefault,
			Attributes: Attributes{},
		}
	}
}

func (b *Buffer) ClearToBeginning() {
	for x := 0; x <= b.CursorX; x++ {
		b.Grid[b.CursorY][x] = Cell{
			Char:       ' ',
			Foreground: tcell.ColorDefault,
			Background: tcell.ColorDefault,
			Attributes: Attributes{},
		}
	}
}

func (b *Buffer) ClearLine() {
	for x := 0; x < b.Width; x++ {
		b.Grid[b.CursorY][x] = Cell{
			Char:       ' ',
			Foreground: tcell.ColorDefault,
			Background: tcell.ColorDefault,
			Attributes: Attributes{},
		}
	}
	b.CursorX = 0
}

func (b *Buffer) Backspace() {
	if b.CursorX > 0 {
		b.CursorX--
	}
}

func (b *Buffer) Tab() {
	b.CursorX = (b.CursorX/8 + 1) * 8
	if b.CursorX >= b.Width {
		b.CursorX = b.Width - 1
	}
}
