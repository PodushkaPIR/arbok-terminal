package ui

import (
	"arbok-terminal/internal/input"
	"arbok-terminal/internal/terminal"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type TerminalWidget struct {
	widget.BaseWidget

	vt         *terminal.VirtualTerminal
	input      *input.Handler
	FontSize   float32
	CellWidth  float32
	CellHeight float32
}

type Option func(*TerminalWidget)

func WithFontSize(size float32) Option {
	return func(w *TerminalWidget) { w.FontSize = size }
}

func WithCellSize(w, h float32) Option {
	return func(tw *TerminalWidget) { tw.CellWidth = w; tw.CellHeight = h }
}

func New(vt *terminal.VirtualTerminal, input *input.Handler, opts ...Option) *TerminalWidget {
	tw := &TerminalWidget{
		vt:         vt,
		input:      input,
		FontSize:   14,
		CellWidth:  9,
		CellHeight: 17,
	}
	for _, opt := range opts {
		opt(tw)
	}
	tw.ExtendBaseWidget(tw)
	return tw
}

func (tw *TerminalWidget) CreateRenderer() fyne.WidgetRenderer {
	return newRenderer(tw)
}

func (tw *TerminalWidget) MinSize() fyne.Size {
	return fyne.NewSize(tw.CellWidth, tw.CellHeight)
}

func (tw *TerminalWidget) Resize(size fyne.Size) {
	tw.BaseWidget.Resize(size)
}

func (tw *TerminalWidget) Size() fyne.Size {
	return tw.BaseWidget.Size()
}

func (tw *TerminalWidget) TypedRune(r rune) {
	if tw.input != nil {
		tw.input.HandleRune(r)
	}
}

func (tw *TerminalWidget) TypedKey(event *fyne.KeyEvent) {
	if event.Name == fyne.KeyPageUp && !tw.vt.IsAltScreen() {
		tw.vt.Lock()
		tw.vt.ScrollUpView(tw.vt.Height())
		tw.vt.Unlock()
		tw.Refresh()
		return
	}
	if event.Name == fyne.KeyPageDown && !tw.vt.IsAltScreen() {
		tw.vt.Lock()
		tw.vt.ScrollDownView(tw.vt.Height())
		tw.vt.Unlock()
		tw.Refresh()
		return
	}
	if event.Name == fyne.KeyEscape && tw.vt.IsScrolling() {
		tw.vt.Lock()
		tw.vt.ResetScroll()
		tw.vt.Unlock()
		tw.Refresh()
		return
	}
	if tw.input != nil {
		tw.input.HandleKey(event)
	}
}

func (tw *TerminalWidget) Scrolled(ev *fyne.ScrollEvent) {
	tw.vt.Lock()
	defer tw.vt.Unlock()

	lines := int(ev.Scrolled.DY)
	if lines == 0 {
		return
	}

	if lines > 0 {
		tw.vt.ScrollUpView(lines)
	} else {
		tw.vt.ScrollDownView(-lines)
	}
	tw.Refresh()
}

func (tw *TerminalWidget) FocusGained() {}
func (tw *TerminalWidget) FocusLost()   {}
