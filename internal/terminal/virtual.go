package terminal

import (
	"log/slog"
	"sync"
	"sync/atomic"
)

type VirtualTerminal struct {
	mu       sync.Mutex
	main     *Screen
	alt      *Screen
	scrollback *ScrollbackBuffer
	modes    DECModeSet

	scrollOffset int

	pendingTitle string

	dirty      atomic.Bool
	generation atomic.Uint64

	onCursorKeyMode func(app bool)
	onResponse       func(data []byte)
}

type DECModeSet struct {
	AltScreen     bool
	CursorVisible bool
	CursorBlink   bool
	AutoWrap      bool
	OriginMode    bool
	InsertMode    bool
	BracketedPaste bool
	TabWidth      int
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

func (vt *VirtualTerminal) SetCursorKeyModeCallback(fn func(bool)) {
	vt.onCursorKeyMode = fn
}

func (vt *VirtualTerminal) SetResponseCallback(fn func([]byte)) {
	vt.onResponse = fn
}

func (vt *VirtualTerminal) SendResponse(data []byte) {
	if vt.onResponse != nil {
		vt.onResponse(data)
	}
}

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
	slog.Debug("vt: set mode", "mode", mode, "on", on)
	switch mode {
	case 1: // DECCKM — cursor keys mode
		slog.Debug("vt: DECCKM", "application", on)
		if vt.onCursorKeyMode != nil {
			vt.onCursorKeyMode(on)
		}
	case 1049: // Alt screen with cursor save/restore
		if on && !vt.modes.AltScreen {
			vt.main.SavedX = vt.main.CursorX
			vt.main.SavedY = vt.main.CursorY
			vt.alt.clear()
			vt.modes.AltScreen = true
		} else if !on && vt.modes.AltScreen {
			vt.modes.AltScreen = false
			vt.main.CursorX = vt.main.SavedX
			vt.main.CursorY = vt.main.SavedY
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
			vt.modes.AltScreen = true
		} else if !on && vt.modes.AltScreen {
			vt.modes.AltScreen = false
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
	case 2004: // Bracketed paste
		slog.Debug("vt: bracketed paste", "enabled", on)
		vt.modes.BracketedPaste = on
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
	slog.Info("vt: resize", "width", width, "height", height, "old_width", vt.main.Width, "old_height", vt.main.Height)
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
		s.clearToEndOfScreen()
	case 1:
		s.clearToBeginningOfScreen()
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
		s.clearToEndOfLine()
	case 1:
		s.clearToBeginningOfLine()
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
	s := vt.CurrentScreen()
	count := min(n, s.ScrollBottom-s.ScrollTop+1)
	vt.pushScrollback(s, count)
	s.scrollUp(n)
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
		vt.pushScrollback(s, 1)
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

func (vt *VirtualTerminal) pushScrollback(s *Screen, n int) {
	for i := 0; i < n; i++ {
		vt.scrollback.Push(s.Grid[s.ScrollTop+i])
	}
}

func (vt *VirtualTerminal) ScrollbackLines() int {
	return vt.scrollback.Len()
}

func (vt *VirtualTerminal) ScrollbackLine(i int) []Cell {
	return vt.scrollback.Get(i)
}

func (vt *VirtualTerminal) ScrollUpView(lines int) {
	sbLen := vt.scrollback.Len()
	vt.scrollOffset += lines
	if vt.scrollOffset > sbLen {
		vt.scrollOffset = sbLen
	}
}

func (vt *VirtualTerminal) ScrollDownView(lines int) {
	vt.scrollOffset -= lines
	if vt.scrollOffset < 0 {
		vt.scrollOffset = 0
	}
}

func (vt *VirtualTerminal) ResetScroll() {
	vt.scrollOffset = 0
}

func (vt *VirtualTerminal) ScrollOffset() int {
	return vt.scrollOffset
}

func (vt *VirtualTerminal) IsScrolling() bool {
	return vt.scrollOffset > 0
}

func (vt *VirtualTerminal) IsAltScreen() bool {
	return vt.modes.AltScreen
}

func (vt *VirtualTerminal) DisplayLine(y int) []Cell {
	sbLen := vt.scrollback.Len()
	h := vt.main.Height

	if vt.scrollOffset == 0 {
		return vt.CurrentScreen().Grid[y]
	}

	idx := sbLen - vt.scrollOffset + y
	if idx >= 0 && idx < sbLen {
		line := vt.scrollback.Get(idx)
		if line != nil {
			return line
		}
	}
	if idx >= sbLen {
		screenY := idx - sbLen
		if screenY >= 0 && screenY < h {
			return vt.CurrentScreen().Grid[screenY]
		}
	}
	return vt.CurrentScreen().Grid[y]
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
