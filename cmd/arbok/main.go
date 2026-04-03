package main

import (
	"image/color"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"github.com/gdamore/tcell/v2"

	"arbok-terminal/internal/pty"
	"arbok-terminal/internal/terminal"
)

type InputHandler struct {
	onInput               func([]byte)
	applicationCursorKeys bool
}

func NewInputHandler() *InputHandler {
	return &InputHandler{}
}

func (h *InputHandler) SetOnInput(fn func([]byte)) {
	h.onInput = fn
}

func (h *InputHandler) HandleKey(event *fyne.KeyEvent) {
	data := h.handleKey(event)
	if data != nil && h.onInput != nil {
		h.onInput(data)
	}
}

func (h *InputHandler) HandleRune(r rune) {
	if h.onInput != nil {
		h.onInput([]byte{byte(r)})
	}
}

func (h *InputHandler) handleKey(event *fyne.KeyEvent) []byte {
	key := event.Name

	switch key {
	case fyne.KeyEscape:
		if h.applicationCursorKeys {
			return []byte{0x1B, 'O', 'A'}
		}
		return []byte{0x1B}

	case fyne.KeyReturn:
		return []byte{0x0D}

	case fyne.KeyBackspace:
		return []byte{0x7F}

	case fyne.KeyTab:
		return []byte{0x09}

	case fyne.KeyUp:
		if h.applicationCursorKeys {
			return []byte{0x1B, 'O', 'A'}
		}
		return []byte{0x1B, '[', 'A'}

	case fyne.KeyDown:
		if h.applicationCursorKeys {
			return []byte{0x1B, 'O', 'B'}
		}
		return []byte{0x1B, '[', 'B'}

	case fyne.KeyRight:
		if h.applicationCursorKeys {
			return []byte{0x1B, 'O', 'C'}
		}
		return []byte{0x1B, '[', 'C'}

	case fyne.KeyLeft:
		if h.applicationCursorKeys {
			return []byte{0x1B, 'O', 'D'}
		}
		return []byte{0x1B, '[', 'D'}

	case fyne.KeyHome:
		return []byte{0x1B, '[', 'H'}

	case fyne.KeyEnd:
		return []byte{0x1B, '[', 'F'}

	case fyne.KeyDelete:
		return []byte{0x1B, '[', '3', '~'}

	case fyne.KeyPageUp:
		return []byte{0x1B, '[', '5', '~'}

	case fyne.KeyPageDown:
		return []byte{0x1B, '[', '6', '~'}

	case fyne.KeyF1:
		return []byte{0x1B, 'O', 'P'}
	case fyne.KeyF2:
		return []byte{0x1B, 'O', 'Q'}
	case fyne.KeyF3:
		return []byte{0x1B, 'O', 'R'}
	case fyne.KeyF4:
		return []byte{0x1B, 'O', 'S'}
	case fyne.KeyF5:
		return []byte{0x1B, '[', '1', '5', '~'}
	case fyne.KeyF6:
		return []byte{0x1B, '[', '1', '7', '~'}
	case fyne.KeyF7:
		return []byte{0x1B, '[', '1', '8', '~'}
	case fyne.KeyF8:
		return []byte{0x1B, '[', '1', '9', '~'}
	case fyne.KeyF9:
		return []byte{0x1B, '[', '2', '0', '~'}
	case fyne.KeyF10:
		return []byte{0x1B, '[', '2', '1', '~'}
	case fyne.KeyF11:
		return []byte{0x1B, '[', '2', '3', '~'}
	case fyne.KeyF12:
		return []byte{0x1B, '[', '2', '4', '~'}
	}

	return nil
}

type TerminalWidget struct {
	buffer     *terminal.Buffer
	fontSize   float32
	cellWidth  float32
	cellHeight float32
	input      *InputHandler

	onResize func(cols, rows int)
}

func NewTerminalWidget(buffer *terminal.Buffer, input *InputHandler) *TerminalWidget {
	return &TerminalWidget{
		buffer:     buffer,
		fontSize:   14,
		cellWidth:  9,
		cellHeight: 17,
		input:      input,
	}
}

func (tw *TerminalWidget) SetBuffer(buffer *terminal.Buffer) {
	tw.buffer = buffer
}

func (tw *TerminalWidget) SetInput(input *InputHandler) {
	tw.input = input
}

func (tw *TerminalWidget) MinSize() fyne.Size {
	if tw.buffer == nil {
		return fyne.NewSize(640, 480)
	}
	return fyne.NewSize(
		float32(tw.buffer.Width)*tw.cellWidth,
		float32(tw.buffer.Height)*tw.cellHeight,
	)
}

func (tw *TerminalWidget) CreateRenderer() fyne.WidgetRenderer {
	return &TerminalRenderer{
		widget: tw,
		bg:     canvas.NewRectangle(theme.Color(theme.ColorNameBackground)),
	}
}

func (tw *TerminalWidget) Hide()                   {}
func (tw *TerminalWidget) Show()                   {}
func (tw *TerminalWidget) Move(pos fyne.Position)  {}
func (tw *TerminalWidget) Resize(size fyne.Size)   {}
func (tw *TerminalWidget) Position() fyne.Position { return fyne.NewPos(0, 0) }
func (tw *TerminalWidget) Size() fyne.Size         { return tw.MinSize() }
func (tw *TerminalWidget) Refresh()                {}
func (tw *TerminalWidget) Visible() bool           { return true }

func (tw *TerminalWidget) FocusGained() {}
func (tw *TerminalWidget) FocusLost()   {}

func (tw *TerminalWidget) TypedRune(r rune) {
	if tw.input != nil {
		tw.input.HandleRune(r)
	}
}

func (tw *TerminalWidget) TypedKey(event *fyne.KeyEvent) {
	if tw.input != nil {
		tw.input.HandleKey(event)
	}
}

func (tw *TerminalWidget) cellPosition(x, y int) fyne.Position {
	return fyne.NewPos(
		float32(x)*tw.cellWidth,
		float32(y)*tw.cellHeight,
	)
}

func tcellColorToRGB(c tcell.Color) (r, g, b uint8) {
	if c == tcell.ColorDefault {
		return 255, 255, 255
	}
	ri, gi, bi := c.RGB()
	return uint8(ri), uint8(gi), uint8(bi)
}

func tcellColorToGoColor(c tcell.Color) color.Color {
	r, g, b := tcellColorToRGB(c)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

type TerminalRenderer struct {
	widget *TerminalWidget
	bg     *canvas.Rectangle
}

func (r *TerminalRenderer) Layout(size fyne.Size) {
	if r.widget.buffer == nil {
		return
	}

	r.bg.Resize(size)

	cellWidth := size.Width / float32(r.widget.buffer.Width)
	cellHeight := size.Height / float32(r.widget.buffer.Height)

	r.widget.cellWidth = cellWidth
	r.widget.cellHeight = cellHeight
}

func (r *TerminalRenderer) MinSize() fyne.Size {
	return r.widget.MinSize()
}

func (r *TerminalRenderer) Objects() []fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, 0)

	if r.widget.buffer == nil {
		return append(objects, r.bg)
	}

	objects = append(objects, r.bg)

	for y := 0; y < r.widget.buffer.Height; y++ {
		for x := 0; x < r.widget.buffer.Width; x++ {
			cell := r.widget.buffer.Grid[y][x]

			text := canvas.NewText(string(cell.Char), tcellColorToGoColor(cell.Foreground))
			text.TextSize = r.widget.fontSize
			text.TextStyle.Monospace = true
			text.Move(r.widget.cellPosition(x, y))
			text.Resize(fyne.NewSize(r.widget.cellWidth, r.widget.cellHeight))

			objects = append(objects, text)
		}
	}

	return objects
}

func (r *TerminalRenderer) Refresh() {
	r.bg.Refresh()
}

func (r *TerminalRenderer) Destroy() {}

var _ fyne.Focusable = (*TerminalWidget)(nil)

func main() {
	fyneApp := app.NewWithID("arbok.terminal")
	window := fyneApp.NewWindow("Arbok Terminal")

	fontSize := float32(14)
	cellWidth := float32(9)
	cellHeight := float32(17)

	calcTerminalSize := func(windowSize fyne.Size) (cols, rows int) {
		cols = int(float32(windowSize.Width) / cellWidth)
		rows = int(float32(windowSize.Height) / cellHeight)
		if cols < 1 {
			cols = 1
		}
		if rows < 1 {
			rows = 1
		}
		return
	}

	initialWidth, initialHeight := calcTerminalSize(fyne.NewSize(800, 600))

	buffer := terminal.NewBuffer(initialWidth, initialHeight)
	parser := terminal.NewParser(buffer, nil)
	parser.TitleHandler = func(title string) {
		window.SetTitle(title)
	}

	ptym, err := pty.New(os.Getenv("SHELL"), initialWidth, initialHeight)
	if err != nil {
		panic(err)
	}
	defer ptym.Close()

	input := NewInputHandler()
	termWidget := NewTerminalWidget(buffer, input)
	termWidget.fontSize = fontSize
	termWidget.cellWidth = cellWidth
	termWidget.cellHeight = cellHeight

	input.SetOnInput(func(data []byte) {
		ptym.Write(data)
	})

	go func() {
		for data := range ptym.OutputCh {
			parser.Parse(data)
			window.Canvas().Refresh(termWidget)
		}
	}()

	go func() {
		for size := range ptym.SizeCh {
			cols := int(size.Cols)
			rows := int(size.Rows)
			buffer.Resize(cols, rows)
			ptym.Resize(cols, rows)
			window.Canvas().Refresh(termWidget)
		}
	}()

	window.SetContent(termWidget)
	window.Resize(fyne.NewSize(800, 600))
	window.SetMaster()

	window.Canvas().Focus(termWidget)

	window.Show()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		var lastCols, lastRows int

		for {
			<-ticker.C
			size := window.Canvas().Size()
			cols, rows := calcTerminalSize(size)
			if cols != lastCols || rows != lastRows {
				if cols != buffer.Width || rows != buffer.Height {
					buffer.Resize(cols, rows)
					ptym.Resize(cols, rows)
					window.Canvas().Refresh(termWidget)
				}
				lastCols, lastRows = cols, rows
			}
		}
	}()

	fyneApp.Run()
}
