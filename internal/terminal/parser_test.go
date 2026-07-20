package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestVT() *VirtualTerminal {
	return NewVirtualTerminal(80, 24)
}

func parseStr(p *Parser, s string) {
	p.Parse([]byte(s))
}

func TestParser_plainText(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "Hello")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	requireCellChar(t, s, 0, 0, 'H')
	requireCellChar(t, s, 1, 0, 'e')
	requireCellChar(t, s, 2, 0, 'l')
	requireCellChar(t, s, 3, 0, 'l')
	requireCellChar(t, s, 4, 0, 'o')
	assert.Equal(t, 5, s.CursorX)
}

func TestParser_carriageReturn(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "AB\x0DC")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	requireCellChar(t, s, 0, 0, 'C')
	assert.Equal(t, 1, s.CursorX)
}

func TestParser_lineFeed(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "A\x0AB")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	requireCellChar(t, s, 0, 0, 'A')
	requireCellChar(t, s, 1, 1, 'B')
}

func TestParser_backspace(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "AB\x08C")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	requireCellChar(t, s, 0, 0, 'A')
	requireCellChar(t, s, 1, 0, 'C')
	assert.Equal(t, 2, s.CursorX)
}

func TestParser_tab(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\tX")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	requireCellChar(t, s, 8, 0, 'X')
	assert.Equal(t, 9, s.CursorX)
}

func TestParser_cursorUp(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[3B\x1b[2A")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, 1, s.CursorY)
}

func TestParser_cursorDown(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[3B")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, 3, s.CursorY)
}

func TestParser_cursorLeft(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "ABCDE\x1b[3D")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, 2, s.CursorX)
}

func TestParser_cursorRight(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[5C")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, 5, s.CursorX)
}

func TestParser_cursorPosition(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[10;5H")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, 4, s.CursorX)
	assert.Equal(t, 9, s.CursorY)
}

func TestParser_cursorPosition_default(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[H")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, 0, s.CursorX)
	assert.Equal(t, 0, s.CursorY)
}

func TestParser_cursorVerticalAbsolute(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[10d")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, 0, s.CursorX)
	assert.Equal(t, 9, s.CursorY)
}

func TestParser_cursorHorizontalAbsolute(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[5G")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, 4, s.CursorX)
	assert.Equal(t, 0, s.CursorY)
}

func TestParser_eraseDisplay(t *testing.T) {
	tests := []struct {
		name  string
		seq   string
		check func(t *testing.T, s *Screen)
	}{
		{
			name: "erase below cursor",
			seq:  "AB\x1b[1;1HA\x1b[0J",
			check: func(t *testing.T, s *Screen) {
				requireCellChar(t, s, 0, 0, 'A')
				assert.Equal(t, defaultCell, s.Grid[0][1])
			},
		},
		{
			name: "erase above cursor",
			seq:  "AB\x1b[1;2H\x1b[1J",
			check: func(t *testing.T, s *Screen) {
				assert.Equal(t, defaultCell, s.Grid[0][0])
				assert.Equal(t, defaultCell, s.Grid[0][1])
			},
		},
		{
			name: "erase entire display",
			seq:  "AB\x1b[2J",
			check: func(t *testing.T, s *Screen) {
				for y := 0; y < s.Height; y++ {
					for x := 0; x < s.Width; x++ {
						assert.Equal(t, defaultCell, s.Grid[y][x])
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vt := newTestVT()
			p := NewParser(vt)
			parseStr(p, tt.seq)
			vt.mu.Lock()
			defer vt.mu.Unlock()
			tt.check(t, vt.CurrentScreen())
		})
	}
}

func TestParser_eraseLine(t *testing.T) {
	tests := []struct {
		name  string
		seq   string
		check func(t *testing.T, s *Screen)
	}{
		{
			name: "erase to end of line",
			seq:  "ABC\x1b[1;2H\x1b[0K",
			check: func(t *testing.T, s *Screen) {
				requireCellChar(t, s, 0, 0, 'A')
				assert.Equal(t, defaultCell, s.Grid[0][1])
				assert.Equal(t, defaultCell, s.Grid[0][2])
			},
		},
		{
			name: "erase to beginning of line",
			seq:  "ABC\x1b[1;2H\x1b[1K",
			check: func(t *testing.T, s *Screen) {
				assert.Equal(t, defaultCell, s.Grid[0][0])
				assert.Equal(t, defaultCell, s.Grid[0][1])
				requireCellChar(t, s, 2, 0, 'C')
			},
		},
		{
			name: "erase entire line",
			seq:  "ABC\x1b[2K",
			check: func(t *testing.T, s *Screen) {
				for x := 0; x < s.Width; x++ {
					assert.Equal(t, defaultCell, s.Grid[0][x])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vt := newTestVT()
			p := NewParser(vt)
			parseStr(p, tt.seq)
			vt.mu.Lock()
			defer vt.mu.Unlock()
			tt.check(t, vt.CurrentScreen())
		})
	}
}

func TestParser_sgr_bold(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[1mB")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.True(t, s.Grid[0][0].Attributes.Bold)
}

func TestParser_sgr_colors(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[31mR")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, ansiColors[31], s.Grid[0][0].Foreground)
}

func TestParser_sgr_bgColor(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[42mG")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, ansiBgColors[42], s.Grid[0][0].Background)
}

func TestParser_sgr_256color(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[38;5;196mR")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	expected := ColorIndex(196)
	assert.Equal(t, expected, s.Grid[0][0].Foreground)
}

func TestParser_sgr_truecolor(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[38;2;10;20;30mX")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, ColorRGB(10, 20, 30), s.Grid[0][0].Foreground)
}

func TestParser_sgr_reset(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[1;31mR\x1b[0mX")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.False(t, s.Grid[0][1].Attributes.Bold)
	assert.Equal(t, ColorDefault, s.Grid[0][1].Foreground)
}

func TestParser_saveRestoreCursor(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "AB\x1b7")   // save cursor
	parseStr(p, "\x1b[1;1H") // move to home
	parseStr(p, "\x1b8")     // restore cursor

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, 2, s.CursorX)
	assert.Equal(t, 0, s.CursorY)
}

func TestParser_altScreen(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "main")
	parseStr(p, "\x1b[?1049h") // switch to alt
	parseStr(p, "alt")
	parseStr(p, "\x1b[?1049l") // switch back to main

	vt.mu.Lock()
	defer vt.mu.Unlock()

	// alt screen should have "alt"
	alt := vt.alt
	requireCellChar(t, alt, 0, 0, 'a')
	requireCellChar(t, alt, 1, 0, 'l')
	requireCellChar(t, alt, 2, 0, 't')

	// main screen should have "main"
	main := vt.main
	requireCellChar(t, main, 0, 0, 'm')
	requireCellChar(t, main, 1, 0, 'a')
	requireCellChar(t, main, 2, 0, 'i')
	requireCellChar(t, main, 3, 0, 'n')
}

func TestParser_cursorVisibility(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[?25l") // hide cursor

	vt.mu.Lock()
	assert.False(t, vt.CurrentScreen().CursorVisible)
	vt.mu.Unlock()

	parseStr(p, "\x1b[?25h") // show cursor

	vt.mu.Lock()
	assert.True(t, vt.CurrentScreen().CursorVisible)
	vt.mu.Unlock()
}

func TestParser_insertDeleteLines(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorX = 0
	vt.CurrentScreen().CursorY = 1
	vt.WriteChar('B', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorX = 0
	vt.CurrentScreen().CursorY = 2
	vt.WriteChar('C', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorX = 0
	vt.CurrentScreen().CursorY = 3
	vt.WriteChar('D', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorY = 1
	vt.Unlock()

	parseStr(p, "\x1b[1L") // insert 1 line at row 1

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	requireCellChar(t, s, 0, 0, 'A')
	assert.Equal(t, defaultCell, s.Grid[1][0])
	requireCellChar(t, s, 0, 2, 'B')
	requireCellChar(t, s, 0, 3, 'C')
}

func TestParser_insertDeleteChars(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "ABCDE")

	vt.mu.Lock()
	vt.CurrentScreen().CursorX = 1
	vt.mu.Unlock()

	parseStr(p, "\x1b[1P") // delete 1 char at cursor

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	requireCellChar(t, s, 0, 0, 'A')
	requireCellChar(t, s, 1, 0, 'C')
	requireCellChar(t, s, 2, 0, 'D')
	requireCellChar(t, s, 3, 0, 'E')
}

func TestParser_scrollUp(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorX = 0
	vt.CurrentScreen().CursorY = 1
	vt.WriteChar('B', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorX = 0
	vt.CurrentScreen().CursorY = 2
	vt.WriteChar('C', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorX = 0
	vt.CurrentScreen().CursorY = 0
	vt.Unlock()

	parseStr(p, "\x1b[1S") // scroll up 1

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	requireCellChar(t, s, 0, 0, 'B')
	requireCellChar(t, s, 0, 1, 'C')
}

func TestParser_scrollRegion(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b[2;4r") // set scroll region rows 2-4

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, 1, s.ScrollTop)
	assert.Equal(t, 3, s.ScrollBottom)
}

func TestParser_osc_title(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "\x1b]0;My Title\x07")

	title := vt.PendingTitle()
	assert.Equal(t, "My Title", title)
}

func TestParser_utf8(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	// "Привет" in UTF-8
	parseStr(p, "Привет")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	requireCellChar(t, s, 0, 0, 'П')
	requireCellChar(t, s, 1, 0, 'р')
	requireCellChar(t, s, 2, 0, 'и')
	requireCellChar(t, s, 3, 0, 'в')
	requireCellChar(t, s, 4, 0, 'е')
	requireCellChar(t, s, 5, 0, 'т')
}

func TestParser_bell(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "A\x07B")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	requireCellChar(t, s, 0, 0, 'A')
	requireCellChar(t, s, 1, 0, 'B')
}

func TestParser_deleteChar(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "ABCDE")

	vt.mu.Lock()
	vt.CurrentScreen().CursorX = 2
	vt.mu.Unlock()

	parseStr(p, "\x1b[1X") // erase 1 char at cursor

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	requireCellChar(t, s, 0, 0, 'A')
	requireCellChar(t, s, 1, 0, 'B')
	assert.Equal(t, defaultCell, s.Grid[0][2])
	requireCellChar(t, s, 3, 0, 'D')
}

func TestParser_indexReverseIndex(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	// ESC D = index (move down, scroll if at bottom)
	parseStr(p, "\x1bD")

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, 1, s.CursorY)
}

func TestParser_reset(t *testing.T) {
	vt := newTestVT()
	p := NewParser(vt)

	parseStr(p, "test")
	parseStr(p, "\x1bc") // reset

	vt.mu.Lock()
	defer vt.mu.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, 0, s.CursorX)
	assert.Equal(t, 0, s.CursorY)
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			assert.Equal(t, defaultCell, s.Grid[y][x])
		}
	}
}
