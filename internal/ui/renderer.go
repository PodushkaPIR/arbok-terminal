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
	cellRects [][]*canvas.Rectangle
	cellTexts [][]*canvas.Text
	objects   []fyne.CanvasObject

	prevGrid   [][]terminal.Cell
	prevW, prevH int
}

func newRenderer(tw *TerminalWidget) *TerminalRenderer {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))

	r := &TerminalRenderer{
		widget: tw,
		bg:     bg,
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
	r.objects = make([]fyne.CanvasObject, 0, 1+h*w*2)
	r.objects = append(r.objects, r.bg)

	r.prevGrid = make([][]terminal.Cell, h)
	r.prevW = w
	r.prevH = h

	for y := 0; y < h; y++ {
		r.cellRects[y] = make([]*canvas.Rectangle, w)
		r.cellTexts[y] = make([]*canvas.Text, w)
		r.prevGrid[y] = make([]terminal.Cell, w)
		for x := 0; x < w; x++ {
			rect := canvas.NewRectangle(color.RGBA{0, 0, 0, 255})
			text := canvas.NewText(" ", color.White)
			text.TextSize = r.widget.FontSize
			text.TextStyle.Monospace = true
			r.cellRects[y][x] = rect
			r.cellTexts[y][x] = text
			r.objects = append(r.objects, rect, text)
		}
	}
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

			if cell.Background == terminal.ColorDefault {
				r.cellRects[y][x].FillColor = bgDef
			} else {
				r.cellRects[y][x].FillColor = ColorToRGBA(cell.Background, bgDef)
			}

			ch := string(cell.Char)
			if ch == "" {
				ch = " "
			}
			r.cellTexts[y][x].Text = ch
			if cell.Foreground == terminal.ColorDefault {
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
}

func (r *TerminalRenderer) Destroy() {}
