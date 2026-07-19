package terminal

import (
	"sync"
	"sync/atomic"
)

type VirtualTerminal struct {
	mu       sync.Mutex
	main     *Screen
	alt      *Screen
	scrollback *ScrollbackBuffer
	modes    DECModeSet

	pendingTitle string

	dirty      atomic.Bool
	generation atomic.Uint64
}

type DECModeSet struct {
	AltScreen    bool
	CursorVisible bool
	CursorBlink  bool
	AutoWrap     bool
	OriginMode   bool
	InsertMode   bool
	TabWidth     int
}

func NewVirtualTerminal(width, height int) *VirtualTerminal {
	return &VirtualTerminal{
		main: NewScreen(width, height),
		alt:  NewScreen(width, height),
		scrollback: NewScrollbackBuffer(1000),
		modes: DECModeSet{
			CursorVisible: true,
			AutoWrap:      true,
			TabWidth:      8,
		},
	}
}

func (vt *VirtualTerminal) CurrentScreen() *Screen {
	if vt.modes.AltScreen {
		return vt.alt
	}
	return vt.main
}

func (vt *VirtualTerminal) Lock()   { vt.mu.Lock() }
func (vt *VirtualTerminal) Unlock() { vt.mu.Unlock() }

func (vt *VirtualTerminal) Width() int  { return vt.main.Width }
func (vt *VirtualTerminal) Height() int { return vt.main.Height }

func (vt *VirtualTerminal) IsDirty() bool     { return vt.dirty.Load() }
func (vt *VirtualTerminal) ClearDirty()        { vt.dirty.Store(false) }
func (vt *VirtualTerminal) Generation() uint64 { return vt.generation.Load() }

func (vt *VirtualTerminal) markDirty() {
	vt.dirty.Store(true)
	vt.generation.Add(1)
}

func (vt *VirtualTerminal) PendingTitle() string {
	t := vt.pendingTitle
	vt.pendingTitle = ""
	return t
}

func (vt *VirtualTerminal) SetPendingTitle(t string) {
	vt.pendingTitle = t
}

func (vt *VirtualTerminal) SetMode(mode int, on bool) {
	switch mode {
	case 1049: // Alt screen
		if on && !vt.modes.AltScreen {
			vt.main.CursorX, vt.main.CursorY = 0, 0
			vt.modes.AltScreen = true
		} else if !on && vt.modes.AltScreen {
			vt.modes.AltScreen = false
			vt.alt.CursorX, vt.alt.CursorY = 0, 0
		}
	case 1048: // Save cursor
		s := vt.CurrentScreen()
		if on {
			s.SavedX = s.CursorX
			s.SavedY = s.CursorY
		} else {
			s.CursorX = s.SavedX
			s.CursorY = s.SavedY
		}
	case 1047: // Alt screen (no cursor save)
		if on && !vt.modes.AltScreen {
			vt.main.CursorX, vt.main.CursorY = 0, 0
			vt.modes.AltScreen = true
		} else if !on && vt.modes.AltScreen {
			vt.modes.AltScreen = false
			vt.alt.CursorX, vt.alt.CursorY = 0, 0
		}
	case 25: // Cursor visible
		vt.modes.CursorVisible = on
		vt.CurrentScreen().CursorVisible = on
	case 7: // Auto-wrap
		vt.modes.AutoWrap = on
	case 6: // Origin mode
		vt.modes.OriginMode = on
	case 4: // Insert mode
		vt.modes.InsertMode = on
	}
	vt.markDirty()
}

func (vt *VirtualTerminal) SetDECPrivateMode(mode int, on bool) {
	vt.SetMode(mode, on)
}

func (vt *VirtualTerminal) Resize(width, height int) {
	if width == vt.main.Width && height == vt.main.Height {
		return
	}
	if len(vt.main.Grid) > 0 {
		for y := 0; y < vt.main.Height; y++ {
			vt.scrollback.Push(vt.main.Grid[y])
		}
	}
	vt.main.resize(width, height)
	vt.alt.resize(width, height)
	vt.main.ScrollBottom = height - 1
	vt.alt.ScrollBottom = height - 1
	vt.markDirty()
}

// Locked methods - must be called with vt.mu held

func (vt *VirtualTerminal) WriteChar(ch rune, fg, bg Color, attrs Attributes) {
	s := vt.CurrentScreen()
	if vt.modes.InsertMode {
		s.insertChars(1)
	}
	s.writeChar(ch, fg, bg, attrs)
	vt.markDirty()
}

func (vt *VirtualTerminal) CursorUp(n int) {
	vt.CurrentScreen().moveUp(n)
	vt.markDirty()
}

func (vt *VirtualTerminal) CursorDown(n int) {
	vt.CurrentScreen().moveDown(n)
	vt.markDirty()
}

func (vt *VirtualTerminal) CursorLeft(n int) {
	vt.CurrentScreen().moveLeft(n)
	vt.markDirty()
}

func (vt *VirtualTerminal) CursorRight(n int) {
	vt.CurrentScreen().moveRight(n)
	vt.markDirty()
}

func (vt *VirtualTerminal) CursorPosition(x, y int) {
	s := vt.CurrentScreen()
	if vt.modes.OriginMode {
		x += 0
		y += s.ScrollTop
	}
	s.moveCursor(x-1, y-1)
	vt.markDirty()
}

func (vt *VirtualTerminal) CursorUpLine(n int) {
	s := vt.CurrentScreen()
	s.CursorX = 0
	s.moveUp(n)
	vt.markDirty()
}

func (vt *VirtualTerminal) CursorDownLine(n int) {
	s := vt.CurrentScreen()
	s.CursorX = 0
	s.moveDown(n)
	vt.markDirty()
}

func (vt *VirtualTerminal) EraseDisplay(mode int) {
	s := vt.CurrentScreen()
	switch mode {
	case 0:
		s.clearToEnd()
	case 1:
		s.clearToBeginning()
	case 2:
		s.clear()
	case 3:
		s.clear()
		vt.scrollback.Reset()
	}
	vt.markDirty()
}

func (vt *VirtualTerminal) EraseLine(mode int) {
	s := vt.CurrentScreen()
	switch mode {
	case 0:
		s.clearToEnd()
	case 1:
		s.clearToBeginning()
	case 2:
		s.clearLine()
	}
	vt.markDirty()
}

func (vt *VirtualTerminal) EraseChars(n int) {
	vt.CurrentScreen().eraseChars(n)
	vt.markDirty()
}

func (vt *VirtualTerminal) InsertLines(n int) {
	vt.CurrentScreen().insertLines(n)
	vt.markDirty()
}

func (vt *VirtualTerminal) DeleteLines(n int) {
	vt.CurrentScreen().deleteLines(n)
	vt.markDirty()
}

func (vt *VirtualTerminal) InsertChars(n int) {
	vt.CurrentScreen().insertChars(n)
	vt.markDirty()
}

func (vt *VirtualTerminal) DeleteChars(n int) {
	vt.CurrentScreen().deleteChars(n)
	vt.markDirty()
}

func (vt *VirtualTerminal) ScrollUp(n int) {
	vt.CurrentScreen().scrollUp(n)
	vt.markDirty()
}

func (vt *VirtualTerminal) ScrollDown(n int) {
	vt.CurrentScreen().scrollDown(n)
	vt.markDirty()
}

func (vt *VirtualTerminal) SetScrollRegion(top, bottom int) {
	vt.CurrentScreen().setScrollRegion(top, bottom)
	vt.markDirty()
}

func (vt *VirtualTerminal) SaveCursor() {
	s := vt.CurrentScreen()
	s.SavedX = s.CursorX
	s.SavedY = s.CursorY
}

func (vt *VirtualTerminal) RestoreCursor() {
	s := vt.CurrentScreen()
	s.CursorX = s.SavedX
	s.CursorY = s.SavedY
	vt.markDirty()
}

func (vt *VirtualTerminal) SetCursorVisible(visible bool) {
	vt.CurrentScreen().CursorVisible = visible
	vt.modes.CursorVisible = visible
	vt.markDirty()
}

func (vt *VirtualTerminal) Backspace() {
	vt.CurrentScreen().backspace()
	vt.markDirty()
}

func (vt *VirtualTerminal) Tab() {
	vt.CurrentScreen().tab()
	vt.markDirty()
}

func (vt *VirtualTerminal) CarriageReturn() {
	vt.CurrentScreen().CursorX = 0
	vt.markDirty()
}

func (vt *VirtualTerminal) LineFeed() {
	s := vt.CurrentScreen()
	if s.CursorY < s.ScrollBottom {
		s.CursorY++
	} else {
		s.scrollUp(1)
	}
	vt.markDirty()
}

func (vt *VirtualTerminal) ReverseLineFeed() {
	s := vt.CurrentScreen()
	if s.CursorY > s.ScrollTop {
		s.CursorY--
	} else {
		s.scrollDown(1)
	}
	vt.markDirty()
}

func (vt *VirtualTerminal) Clear() {
	vt.main.clear()
	vt.alt.clear()
	vt.scrollback.Reset()
	vt.markDirty()
}

func (vt *VirtualTerminal) ScrollbackLines() int {
	return vt.scrollback.Len()
}

func (vt *VirtualTerminal) ScrollbackLine(i int) []Cell {
	return vt.scrollback.Get(i)
}

type ScrollbackBuffer struct {
	lines [][]Cell
	size  int
	start int
	count int
	mu    sync.Mutex
}

func NewScrollbackBuffer(size int) *ScrollbackBuffer {
	return &ScrollbackBuffer{
		lines: make([][]Cell, size),
		size:  size,
	}
}

func (sb *ScrollbackBuffer) Push(line []Cell) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	copied := make([]Cell, len(line))
	copy(copied, line)
	sb.lines[(sb.start+sb.count)%sb.size] = copied
	if sb.count < sb.size {
		sb.count++
	} else {
		sb.start = (sb.start + 1) % sb.size
	}
}

func (sb *ScrollbackBuffer) Get(index int) []Cell {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	if index < 0 || index >= sb.count {
		return nil
	}
	return sb.lines[(sb.start+index)%sb.size]
}

func (sb *ScrollbackBuffer) Len() int {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.count
}

func (sb *ScrollbackBuffer) Reset() {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.count = 0
	sb.start = 0
}
