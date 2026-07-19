package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVirtualTerminal_new(t *testing.T) {
	vt := NewVirtualTerminal(80, 24)
	assert.Equal(t, 80, vt.Width())
	assert.Equal(t, 24, vt.Height())
	assert.False(t, vt.IsDirty())
}

func TestVirtualTerminal_dirty(t *testing.T) {
	vt := NewVirtualTerminal(80, 24)
	vt.Lock()
	defer vt.Unlock()

	assert.False(t, vt.IsDirty())

	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	assert.True(t, vt.IsDirty())
	assert.Greater(t, vt.Generation(), uint64(0))

	vt.ClearDirty()
	assert.False(t, vt.IsDirty())
}

func TestVirtualTerminal_altScreen(t *testing.T) {
	vt := NewVirtualTerminal(80, 24)

	vt.Lock()
	vt.SetMode(1049, true)
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.Unlock()

	assert.Equal(t, vt.alt, vt.CurrentScreen())

	vt.Lock()
	vt.SetMode(1049, false)
	vt.Unlock()

	assert.Equal(t, vt.main, vt.CurrentScreen())
}

func TestVirtualTerminal_altScreen_cursorIndependence(t *testing.T) {
	vt := NewVirtualTerminal(80, 24)

	vt.Lock()
	vt.WriteChar('M', ColorDefault, ColorDefault, Attributes{})
	vt.SetMode(1049, true)
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.SetMode(1049, false)
	vt.Unlock()

	// main screen should have M at position 0,0
	assert.Equal(t, 'M', vt.main.Grid[0][0].Char)
	// alt screen should have A at position 0,0
	assert.Equal(t, 'A', vt.alt.Grid[0][0].Char)
}

func TestVirtualTerminal_scrollback(t *testing.T) {
	vt := NewVirtualTerminal(10, 3)

	vt.Lock()
	for i := 0; i < 5; i++ {
		vt.WriteChar(rune('A'+i), ColorDefault, ColorDefault, Attributes{})
		vt.LineFeed()
	}
	vt.Unlock()

	assert.Greater(t, vt.ScrollbackLines(), 0)
}

func TestVirtualTerminal_resize(t *testing.T) {
	vt := NewVirtualTerminal(80, 24)

	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.Unlock()

	vt.Resize(120, 40)

	assert.Equal(t, 120, vt.Width())
	assert.Equal(t, 40, vt.Height())
	assert.True(t, vt.IsDirty())
}

func TestVirtualTerminal_resize_preservesContent(t *testing.T) {
	vt := NewVirtualTerminal(10, 5)

	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.WriteChar('B', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorX = 0
	vt.CurrentScreen().CursorY = 1
	vt.WriteChar('C', ColorDefault, ColorDefault, Attributes{})
	vt.Unlock()

	vt.Resize(15, 8)

	vt.Lock()
	defer vt.Unlock()
	s := vt.CurrentScreen()
	assert.Equal(t, 'A', s.Grid[0][0].Char)
	assert.Equal(t, 'B', s.Grid[0][1].Char)
	assert.Equal(t, 'C', s.Grid[1][0].Char)
}

func TestVirtualTerminal_scrollRegion(t *testing.T) {
	vt := NewVirtualTerminal(10, 5)

	vt.Lock()
	vt.SetScrollRegion(1, 3)
	vt.Unlock()

	assert.Equal(t, 1, vt.main.ScrollTop)
	assert.Equal(t, 3, vt.main.ScrollBottom)
}

func TestVirtualTerminal_insertDeleteLines(t *testing.T) {
	vt := NewVirtualTerminal(10, 5)

	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorX = 0
	vt.CurrentScreen().CursorY = 1
	vt.WriteChar('B', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorX = 0
	vt.CurrentScreen().CursorY = 2
	vt.WriteChar('C', ColorDefault, ColorDefault, Attributes{})

	vt.CurrentScreen().CursorY = 1
	vt.InsertLines(1)

	assert.Equal(t, 'A', vt.main.Grid[0][0].Char)
	assert.Equal(t, defaultCell, vt.main.Grid[1][0])
	assert.Equal(t, 'B', vt.main.Grid[2][0].Char)
	vt.Unlock()
}

func TestVirtualTerminal_insertDeleteChars(t *testing.T) {
	vt := NewVirtualTerminal(10, 3)

	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.WriteChar('B', ColorDefault, ColorDefault, Attributes{})
	vt.WriteChar('C', ColorDefault, ColorDefault, Attributes{})

	vt.CurrentScreen().CursorX = 1
	vt.DeleteChars(1)

	assert.Equal(t, 'A', vt.main.Grid[0][0].Char)
	assert.Equal(t, 'C', vt.main.Grid[0][1].Char)
	vt.Unlock()
}

func TestVirtualTerminal_scrollUp(t *testing.T) {
	vt := NewVirtualTerminal(10, 5)

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
	vt.ScrollUp(1)

	assert.Equal(t, 'B', vt.main.Grid[0][0].Char)
	assert.Equal(t, 'C', vt.main.Grid[1][0].Char)
	vt.Unlock()
}

func TestVirtualTerminal_scrollDown(t *testing.T) {
	vt := NewVirtualTerminal(10, 5)

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
	vt.ScrollDown(1)

	assert.Equal(t, defaultCell, vt.main.Grid[0][0])
	assert.Equal(t, 'A', vt.main.Grid[1][0].Char)
	assert.Equal(t, 'B', vt.main.Grid[2][0].Char)
	vt.Unlock()
}

func TestVirtualTerminal_lineFeed(t *testing.T) {
	vt := NewVirtualTerminal(10, 5)

	vt.Lock()
	vt.LineFeed()
	vt.LineFeed()
	vt.LineFeed()
	vt.Unlock()

	assert.Equal(t, 3, vt.main.CursorY)
}

func TestVirtualTerminal_lineFeed_scrolls(t *testing.T) {
	vt := NewVirtualTerminal(10, 3)

	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorY = 2
	vt.LineFeed() // should scroll, cursor stays at 2
	vt.Unlock()

	assert.Equal(t, 2, vt.main.CursorY)
	assert.Equal(t, defaultCell, vt.main.Grid[0][0])
	assert.Equal(t, 1, vt.ScrollbackLines())
}

func TestVirtualTerminal_reverseLineFeed(t *testing.T) {
	vt := NewVirtualTerminal(10, 5)

	vt.Lock()
	vt.CurrentScreen().CursorY = 2
	vt.ReverseLineFeed()
	vt.ReverseLineFeed()
	vt.ReverseLineFeed()
	vt.Unlock()

	assert.Equal(t, 0, vt.main.CursorY)
}

func TestVirtualTerminal_reverseLineFeed_scrolls(t *testing.T) {
	vt := NewVirtualTerminal(10, 3)

	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorY = 0
	vt.ReverseLineFeed() // should scroll down
	vt.Unlock()

	assert.Equal(t, 0, vt.main.CursorY)
}

func TestVirtualTerminal_saveRestoreCursor(t *testing.T) {
	vt := NewVirtualTerminal(10, 5)

	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.WriteChar('B', ColorDefault, ColorDefault, Attributes{})
	vt.SaveCursor()
	vt.CurrentScreen().CursorX = 0
	vt.CurrentScreen().CursorY = 3
	vt.RestoreCursor()
	vt.Unlock()

	assert.Equal(t, 2, vt.main.CursorX)
	assert.Equal(t, 0, vt.main.CursorY)
}

func TestVirtualTerminal_setCursorVisible(t *testing.T) {
	vt := NewVirtualTerminal(10, 5)

	vt.Lock()
	vt.SetCursorVisible(false)
	assert.False(t, vt.main.CursorVisible)
	vt.SetCursorVisible(true)
	assert.True(t, vt.main.CursorVisible)
	vt.Unlock()
}

func TestVirtualTerminal_clear(t *testing.T) {
	vt := NewVirtualTerminal(10, 5)

	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.Unlock()

	vt.Clear()

	vt.Lock()
	defer vt.Unlock()
	for y := 0; y < vt.Height(); y++ {
		for x := 0; x < vt.Width(); x++ {
			assert.Equal(t, defaultCell, vt.main.Grid[y][x])
			assert.Equal(t, defaultCell, vt.alt.Grid[y][x])
		}
	}
}

func TestVirtualTerminal_generation(t *testing.T) {
	vt := NewVirtualTerminal(10, 5)
	gen := vt.Generation()

	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.Unlock()

	assert.Greater(t, vt.Generation(), gen)
}

func TestVirtualTerminal_EraseDisplay_mode0(t *testing.T) {
	vt := NewVirtualTerminal(10, 5)
	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorX = 0
	vt.CurrentScreen().CursorY = 1
	vt.WriteChar('B', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorX = 0
	vt.CurrentScreen().CursorY = 2
	vt.WriteChar('C', ColorDefault, ColorDefault, Attributes{})

	vt.CurrentScreen().CursorX = 2
	vt.CurrentScreen().CursorY = 1
	vt.EraseDisplay(0)

	assert.Equal(t, 'A', vt.main.Grid[0][0].Char)
	assert.Equal(t, defaultCell, vt.main.Grid[1][2])
	assert.Equal(t, defaultCell, vt.main.Grid[1][3])
	for x := 0; x < vt.Width(); x++ {
		assert.Equal(t, defaultCell, vt.main.Grid[2][x])
	}
	vt.Unlock()
}

func TestVirtualTerminal_EraseDisplay_mode1(t *testing.T) {
	vt := NewVirtualTerminal(10, 5)
	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorX = 0
	vt.CurrentScreen().CursorY = 1
	vt.WriteChar('B', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorX = 0
	vt.CurrentScreen().CursorY = 2
	vt.WriteChar('C', ColorDefault, ColorDefault, Attributes{})

	vt.CurrentScreen().CursorX = 2
	vt.CurrentScreen().CursorY = 1
	vt.EraseDisplay(1)

	for x := 0; x < vt.Width(); x++ {
		assert.Equal(t, defaultCell, vt.main.Grid[0][x])
	}
	assert.Equal(t, defaultCell, vt.main.Grid[1][0])
	assert.Equal(t, defaultCell, vt.main.Grid[1][1])
	assert.Equal(t, defaultCell, vt.main.Grid[1][2])
	assert.Equal(t, 'C', vt.main.Grid[2][0].Char)
	vt.Unlock()
}

func TestVirtualTerminal_EraseDisplay_mode2(t *testing.T) {
	vt := NewVirtualTerminal(10, 5)
	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.CurrentScreen().CursorX = 5
	vt.CurrentScreen().CursorY = 3
	vt.EraseDisplay(2)

	assert.Equal(t, 0, vt.main.CursorX)
	assert.Equal(t, 0, vt.main.CursorY)
	for y := 0; y < vt.Height(); y++ {
		for x := 0; x < vt.Width(); x++ {
			assert.Equal(t, defaultCell, vt.main.Grid[y][x])
		}
	}
	vt.Unlock()
}

func TestVirtualTerminal_EraseLine_mode0(t *testing.T) {
	vt := NewVirtualTerminal(10, 3)
	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.WriteChar('B', ColorDefault, ColorDefault, Attributes{})
	vt.WriteChar('C', ColorDefault, ColorDefault, Attributes{})

	vt.CurrentScreen().CursorX = 1
	vt.EraseLine(0)

	assert.Equal(t, 'A', vt.main.Grid[0][0].Char)
	assert.Equal(t, defaultCell, vt.main.Grid[0][1])
	assert.Equal(t, defaultCell, vt.main.Grid[0][2])
	vt.Unlock()
}

func TestVirtualTerminal_EraseLine_mode1(t *testing.T) {
	vt := NewVirtualTerminal(10, 3)
	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.WriteChar('B', ColorDefault, ColorDefault, Attributes{})
	vt.WriteChar('C', ColorDefault, ColorDefault, Attributes{})

	vt.CurrentScreen().CursorX = 1
	vt.EraseLine(1)

	assert.Equal(t, defaultCell, vt.main.Grid[0][0])
	assert.Equal(t, defaultCell, vt.main.Grid[0][1])
	requireCellCharR(t, vt, 2, 0, 'C')
	vt.Unlock()
}

func TestVirtualTerminal_EraseLine_mode2(t *testing.T) {
	vt := NewVirtualTerminal(10, 3)
	vt.Lock()
	vt.WriteChar('A', ColorDefault, ColorDefault, Attributes{})
	vt.WriteChar('B', ColorDefault, ColorDefault, Attributes{})
	vt.WriteChar('C', ColorDefault, ColorDefault, Attributes{})

	vt.CurrentScreen().CursorX = 2
	vt.EraseLine(2)

	assert.Equal(t, 2, vt.main.CursorX)
	for x := 0; x < vt.Width(); x++ {
		assert.Equal(t, defaultCell, vt.main.Grid[0][x])
	}
	vt.Unlock()
}

func requireCellCharR(t *testing.T, vt *VirtualTerminal, x, y int, expected rune) {
	t.Helper()
	s := vt.CurrentScreen()
	require.Less(t, y, s.Height, "y out of bounds")
	require.Less(t, x, s.Width, "x out of bounds")
	assert.Equal(t, expected, s.Grid[y][x].Char, "cell(%d,%d)", x, y)
}
