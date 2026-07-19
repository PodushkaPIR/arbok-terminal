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
	r.prevW = w
	r.prevH = h

	bgDef := theme.Color(theme.ColorNameBackground)

	for y := 0; y < h; y++ {
		r.cellRects[y] = make([]*canvas.Rectangle, w)
		r.cellTexts[y] = make([]*canvas.Text, w)
		r.prevGrid[y] = make([]terminal.Cell, w)
		for x := 0; x < w; x++ {
			rect := canvas.NewRectangle(bgDef)
			text := canvas.NewText(" ", color.White)
			text.TextSize = r.widget.FontSize
			text.TextStyle.Monospace = true
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

			ch := string(cell.Char)
			if cell.Char == 0 || ch == "" {
				ch = " "
			}
			r.cellTexts[y][x].Text = ch
			isDefaultFg := cell.Foreground == terminal.ColorDefault || (!cell.Foreground.Default && cell.Foreground.R == 255 && cell.Foreground.G == 255 && cell.Foreground.B == 255)
			if isDefaultFg {
				r.cellTexts[y][x].Color = color.White
			} else {
				r.cellTexts[y][x].Color = ColorToRGBA(cell.Foreground, bgDef)
			}

			r.cellRects[y][x].Refresh()
			r.cellTexts[y][x].Refresh()

			r.prevGrid[y][x] = cell
		}
	}

	r.bg.Refresh()

	cursorX := screen.CursorX
	cursorY := screen.CursorY
	if screen.CursorVisible && cursorX >= 0 && cursorX < w && cursorY >= 0 && cursorY < h {
		cellW := r.widget.CellWidth
		cellH := r.widget.CellHeight
		r.cursor.Move(fyne.NewPos(float32(cursorX)*cellW, float32(cursorY)*cellH))
		r.cursor.Resize(fyne.NewSize(cellW, cellH))
		r.cursor.Show()
	} else {
		r.cursor.Hide()
	}
}

func (r *TerminalRenderer) Destroy() {}
