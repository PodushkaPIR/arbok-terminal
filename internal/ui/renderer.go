package ui

import (
	"image/color"

	"arbok-terminal/internal/terminal"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

type TerminalRenderer struct {
	widget    *TerminalWidget
	bg        *canvas.Rectangle
	cursor    *canvas.Rectangle
	cellRects [][]*canvas.Rectangle
	cellTexts [][]*canvas.Text
	objects   []fyne.CanvasObject

	prevGrid   [][]terminal.Cell
	prevW, prevH int

	dirtyLines    []bool
	prevCursorX   int
	prevCursorY   int
	cursorActive  bool
}

func newRenderer(tw *TerminalWidget) *TerminalRenderer {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	cursor := canvas.NewRectangle(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	cursor.Hide()

	r := &TerminalRenderer{
		widget: tw,
		bg:     bg,
		cursor: cursor,
	}

	if tw.vt != nil {
		r.buildCache()
	}

	return r
}

func (r *TerminalRenderer) buildCache() {
	screen := r.widget.vt.CurrentScreen()
	h := screen.Height
	w := screen.Width

	r.cellRects = make([][]*canvas.Rectangle, h)
	r.cellTexts = make([][]*canvas.Text, h)
	r.objects = make([]fyne.CanvasObject, 0, 2+h*w*2)
	r.objects = append(r.objects, r.bg)

	r.prevGrid = make([][]terminal.Cell, h)
	r.dirtyLines = make([]bool, h)
	r.prevW = w
	r.prevH = h
	r.prevCursorX = -1
	r.prevCursorY = -1
	r.cursorActive = false

	bgDef := theme.Color(theme.ColorNameBackground)

	for y := 0; y < h; y++ {
		r.cellRects[y] = make([]*canvas.Rectangle, w)
		r.cellTexts[y] = make([]*canvas.Text, w)
		r.prevGrid[y] = make([]terminal.Cell, w)
		r.dirtyLines[y] = true
		for x := 0; x < w; x++ {
			rect := canvas.NewRectangle(bgDef)
			text := canvas.NewText(" ", color.White)
			text.TextSize = r.widget.FontSize
			text.TextStyle.Monospace = true
			text.Hide()
			r.cellRects[y][x] = rect
			r.cellTexts[y][x] = text
			r.objects = append(r.objects, rect, text)
		}
	}
	r.objects = append(r.objects, r.cursor)
}

func (r *TerminalRenderer) Layout(size fyne.Size) {
	if r.widget.vt == nil {
		return
	}

	screen := r.widget.vt.CurrentScreen()
	h := screen.Height
	w := screen.Width

	if r.prevH != h || r.prevW != w {
		r.buildCache()
		h = r.prevH
		w = r.prevW
	}

	r.bg.Resize(size)

	cellW := size.Width / float32(w)
	cellH := size.Height / float32(h)
	r.widget.CellWidth = cellW
	r.widget.CellHeight = cellH

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			pos := fyne.NewPos(float32(x)*cellW, float32(y)*cellH)
			sz := fyne.NewSize(cellW, cellH)
			r.cellRects[y][x].Move(pos)
			r.cellRects[y][x].Resize(sz)
			r.cellTexts[y][x].Move(pos)
			r.cellTexts[y][x].Resize(sz)
		}
	}

	cursorX := screen.CursorX
	cursorY := screen.CursorY
	if cursorX >= 0 && cursorX < w && cursorY >= 0 && cursorY < h {
		r.cursor.Move(fyne.NewPos(float32(cursorX)*cellW, float32(cursorY)*cellH))
		r.cursor.Resize(fyne.NewSize(cellW, cellH))
	}
}

func (r *TerminalRenderer) MinSize() fyne.Size {
	return r.widget.MinSize()
}

func (r *TerminalRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *TerminalRenderer) Refresh() {
	if r.widget.vt == nil {
		return
	}

	r.widget.vt.Lock()
	defer r.widget.vt.Unlock()

	screen := r.widget.vt.CurrentScreen()
	h := screen.Height
	w := screen.Width

	if r.prevH != h || r.prevW != w {
		r.buildCache()
		h = r.prevH
		w = r.prevW
	}

	bgDef := theme.Color(theme.ColorNameBackground)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := screen.Grid[y][x]
			prev := r.prevGrid[y][x]

			if cell == prev {
				continue
			}

			isDefaultBg := cell.Background == terminal.ColorDefault || (!cell.Background.Default && cell.Background.R == 0 && cell.Background.G == 0 && cell.Background.B == 0)
			if isDefaultBg {
				r.cellRects[y][x].FillColor = bgDef
			} else {
				r.cellRects[y][x].FillColor = ColorToRGBA(cell.Background, bgDef)
			}

			isSpace := cell.Char == 0 || cell.Char == ' '
			if isSpace {
				r.cellTexts[y][x].Hide()
			} else {
				r.cellTexts[y][x].Text = string(cell.Char)
				isDefaultFg := cell.Foreground == terminal.ColorDefault || (!cell.Foreground.Default && cell.Foreground.R == 255 && cell.Foreground.G == 255 && cell.Foreground.B == 255)
				if isDefaultFg {
					r.cellTexts[y][x].Color = color.White
				} else {
					r.cellTexts[y][x].Color = ColorToRGBA(cell.Foreground, bgDef)
				}
				r.cellTexts[y][x].Show()
			}

			r.cellRects[y][x].Refresh()
			if !isSpace {
				r.cellTexts[y][x].Refresh()
			}

			r.prevGrid[y][x] = cell
			r.dirtyLines[y] = true
		}
	}

	r.bg.Refresh()

	cursorX := screen.CursorX
	cursorY := screen.CursorY
	showCursor := screen.CursorVisible && cursorX >= 0 && cursorX < w && cursorY >= 0 && cursorY < h

	if showCursor {
		cellW := r.widget.CellWidth
		cellH := r.widget.CellHeight

		if r.cursorActive && (r.prevCursorX != cursorX || r.prevCursorY != cursorY) {
			r.restoreCellAt(r.prevCursorX, r.prevCursorY, screen, bgDef)
		}

		r.invertCellAt(cursorX, cursorY, screen, bgDef)

		r.cursor.Move(fyne.NewPos(float32(cursorX)*cellW, float32(cursorY)*cellH))
		r.cursor.Resize(fyne.NewSize(cellW, cellH))
		r.cursor.Hide()
		r.prevCursorX = cursorX
		r.prevCursorY = cursorY
		r.cursorActive = true
	} else if r.cursorActive {
		r.restoreCellAt(r.prevCursorX, r.prevCursorY, screen, bgDef)
		r.cursor.Hide()
		r.cursorActive = false
	}
}

func (r *TerminalRenderer) invertCellAt(x, y int, screen *terminal.Screen, bgDef color.Color) {
	if x < 0 || x >= r.prevW || y < 0 || y >= r.prevH {
		return
	}
	cell := screen.Grid[y][x]

	isDefaultBg := cell.Background == terminal.ColorDefault
	isDefaultFg := cell.Foreground == terminal.ColorDefault

	var fg, bg color.RGBA
	if isDefaultFg {
		fg = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	} else {
		fg = ColorToRGBA(cell.Foreground, bgDef)
	}
	if isDefaultBg {
		bg = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	} else {
		bg = ColorToRGBA(cell.Background, bgDef)
	}

	r.cellRects[y][x].FillColor = fg
	r.cellRects[y][x].Refresh()

	ch := string(cell.Char)
	if cell.Char == 0 || ch == "" {
		ch = " "
	}
	r.cellTexts[y][x].Text = ch
	r.cellTexts[y][x].Color = bg
	r.cellTexts[y][x].Show()
	r.cellTexts[y][x].Refresh()
}

func (r *TerminalRenderer) restoreCellAt(x, y int, screen *terminal.Screen, bgDef color.Color) {
	if x < 0 || x >= r.prevW || y < 0 || y >= r.prevH {
		return
	}
	cell := screen.Grid[y][x]

	isDefaultBg := cell.Background == terminal.ColorDefault || (!cell.Background.Default && cell.Background.R == 0 && cell.Background.G == 0 && cell.Background.B == 0)
	if isDefaultBg {
		r.cellRects[y][x].FillColor = bgDef
	} else {
		r.cellRects[y][x].FillColor = ColorToRGBA(cell.Background, bgDef)
	}
	r.cellRects[y][x].Refresh()

	isSpace := cell.Char == 0 || cell.Char == ' '
	if isSpace {
		r.cellTexts[y][x].Hide()
	} else {
		isDefaultFg := cell.Foreground == terminal.ColorDefault || (!cell.Foreground.Default && cell.Foreground.R == 255 && cell.Foreground.G == 255 && cell.Foreground.B == 255)
		if isDefaultFg {
			r.cellTexts[y][x].Color = color.White
		} else {
			r.cellTexts[y][x].Color = ColorToRGBA(cell.Foreground, bgDef)
		}
		r.cellTexts[y][x].Text = string(cell.Char)
		r.cellTexts[y][x].Show()
		r.cellTexts[y][x].Refresh()
	}
}

func (r *TerminalRenderer) Destroy() {}
