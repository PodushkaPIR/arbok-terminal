package terminal

import "sync"

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
	Foreground Color
	Background Color
	Attributes Attributes
}

var defaultCell = Cell{
	Char:       ' ',
	Foreground: ColorDefault,
	Background: ColorDefault,
}

type Buffer struct {
	mu sync.RWMutex

	Width  int
	Height int
	Grid   [][]Cell

	CursorX int
	CursorY int

	dirty bool
}

func NewBuffer(width, height int) *Buffer {
	grid := make([][]Cell, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			grid[y][x] = defaultCell
		}
	}

	return &Buffer{
		Width:  width,
		Height: height,
		Grid:   grid,
	}
}

func (b *Buffer) Lock()   { b.mu.Lock() }
func (b *Buffer) Unlock() { b.mu.Unlock() }
func (b *Buffer) RLock()  { b.mu.RLock() }
func (b *Buffer) RUnlock() { b.mu.RUnlock() }

func (b *Buffer) IsDirty() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.dirty
}

func (b *Buffer) ClearDirty() {
	b.mu.Lock()
	b.dirty = false
	b.mu.Unlock()
}

func (b *Buffer) Resize(width, height int) {
	b.mu.Lock()
	defer b.mu.Unlock()

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
				newGrid[y][x] = defaultCell
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
	b.dirty = true
}

func (b *Buffer) WriteChar(ch rune, fg, bg Color, attrs Attributes) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.writeCharLocked(ch, fg, bg, attrs)
}

func (b *Buffer) writeCharLocked(ch rune, fg, bg Color, attrs Attributes) {
	if b.CursorX >= b.Width {
		b.newlineLocked()
	}

	if b.CursorY >= 0 && b.CursorY < b.Height && b.CursorX >= 0 && b.CursorX < b.Width {
		b.Grid[b.CursorY][b.CursorX] = Cell{
			Char:       ch,
			Foreground: fg,
			Background: bg,
			Attributes: attrs,
		}
		b.dirty = true
	}

	b.CursorX++
}

func (b *Buffer) Newline() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.newlineLocked()
}

func (b *Buffer) newlineLocked() {
	b.CursorX = 0
	if b.CursorY < b.Height-1 {
		b.CursorY++
	}
	b.dirty = true
}

func (b *Buffer) MoveCursor(x, y int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.moveCursorLocked(x, y)
}

func (b *Buffer) moveCursorLocked(x, y int) {
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
	b.mu.Lock()
	defer b.mu.Unlock()
	b.moveUpLocked(lines)
}

func (b *Buffer) moveUpLocked(lines int) {
	b.CursorY -= lines
	if b.CursorY < 0 {
		b.CursorY = 0
	}
}

func (b *Buffer) MoveDown(lines int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.moveDownLocked(lines)
}

func (b *Buffer) moveDownLocked(lines int) {
	b.CursorY += lines
	if b.CursorY >= b.Height {
		b.CursorY = b.Height - 1
	}
}

func (b *Buffer) MoveLeft(cols int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.moveLeftLocked(cols)
}

func (b *Buffer) moveLeftLocked(cols int) {
	b.CursorX -= cols
	if b.CursorX < 0 {
		b.CursorX = 0
	}
}

func (b *Buffer) MoveRight(cols int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.moveRightLocked(cols)
}

func (b *Buffer) moveRightLocked(cols int) {
	b.CursorX += cols
	if b.CursorX >= b.Width {
		b.CursorX = b.Width - 1
	}
}

func (b *Buffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clearLocked()
}

func (b *Buffer) clearLocked() {
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			b.Grid[y][x] = defaultCell
		}
	}
	b.CursorX = 0
	b.CursorY = 0
	b.dirty = true
}

func (b *Buffer) ClearToEnd() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clearToEndLocked()
}

func (b *Buffer) clearToEndLocked() {
	for x := b.CursorX; x < b.Width; x++ {
		b.Grid[b.CursorY][x] = defaultCell
	}
	for y := b.CursorY + 1; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			b.Grid[y][x] = defaultCell
		}
	}
	b.dirty = true
}

func (b *Buffer) ClearToBeginning() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clearToBeginningLocked()
}

func (b *Buffer) clearToBeginningLocked() {
	for x := 0; x <= b.CursorX; x++ {
		b.Grid[b.CursorY][x] = defaultCell
	}
	for y := 0; y < b.CursorY; y++ {
		for x := 0; x < b.Width; x++ {
			b.Grid[y][x] = defaultCell
		}
	}
	b.dirty = true
}

func (b *Buffer) ClearLine() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clearLineLocked()
}

func (b *Buffer) clearLineLocked() {
	for x := 0; x < b.Width; x++ {
		b.Grid[b.CursorY][x] = defaultCell
	}
	b.CursorX = 0
	b.dirty = true
}

func (b *Buffer) Backspace() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.backspaceLocked()
}

func (b *Buffer) backspaceLocked() {
	if b.CursorX > 0 {
		b.CursorX--
	}
}

func (b *Buffer) Tab() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tabLocked()
}

func (b *Buffer) tabLocked() {
	b.CursorX = (b.CursorX/8 + 1) * 8
	if b.CursorX >= b.Width {
		b.CursorX = b.Width - 1
	}
}
