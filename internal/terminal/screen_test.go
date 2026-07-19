package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScreen(t *testing.T) {
	s := NewScreen(80, 24)
	assert.Equal(t, 80, s.Width)
	assert.Equal(t, 24, s.Height)
	assert.Equal(t, 0, s.CursorX)
	assert.Equal(t, 0, s.CursorY)
	assert.True(t, s.CursorVisible)
	assert.Equal(t, 0, s.ScrollTop)
	assert.Equal(t, 23, s.ScrollBottom)
	assert.Len(t, s.Grid, 24)
	assert.Len(t, s.Grid[0], 80)
}

func TestScreen_writeChar(t *testing.T) {
	s := NewScreen(10, 3)

	s.writeChar('A', ColorDefault, ColorDefault, Attributes{})
	assert.Equal(t, 'A', s.Grid[0][0].Char)
	assert.Equal(t, 1, s.CursorX)

	s.writeChar('B', ColorDefault, ColorDefault, Attributes{})
	assert.Equal(t, 'B', s.Grid[0][1].Char)
	assert.Equal(t, 2, s.CursorX)
}

func TestScreen_writeChar_withColor(t *testing.T) {
	s := NewScreen(10, 3)
	fg := ColorRGB(255, 0, 0)
	bg := ColorRGB(0, 0, 255)
	attrs := Attributes{Bold: true, Underline: true}

	s.writeChar('X', fg, bg, attrs)
	cell := s.Grid[0][0]
	assert.Equal(t, 'X', cell.Char)
	assert.Equal(t, fg, cell.Foreground)
	assert.Equal(t, bg, cell.Background)
	assert.True(t, cell.Attributes.Bold)
	assert.True(t, cell.Attributes.Underline)
}

func TestScreen_writeChar_wraps(t *testing.T) {
	s := NewScreen(5, 3)

	for i := 0; i < 5; i++ {
		s.writeChar(rune('A'+i), ColorDefault, ColorDefault, Attributes{})
	}
	assert.Equal(t, 5, s.CursorX)
	assert.Equal(t, 0, s.CursorY)

	s.writeChar('F', ColorDefault, ColorDefault, Attributes{})
	assert.Equal(t, 1, s.CursorX)
	assert.Equal(t, 1, s.CursorY)
	assert.Equal(t, 'F', s.Grid[1][0].Char)
}

func TestScreen_newline(t *testing.T) {
	s := NewScreen(10, 3)

	s.newline()
	assert.Equal(t, 0, s.CursorX)
	assert.Equal(t, 1, s.CursorY)

	s.CursorY = 2
	s.newline()
	assert.Equal(t, 0, s.CursorX)
	assert.Equal(t, 2, s.CursorY)
}

func TestScreen_moveCursor(t *testing.T) {
	s := NewScreen(10, 3)
	s.moveCursor(5, 2)
	assert.Equal(t, 5, s.CursorX)
	assert.Equal(t, 2, s.CursorY)
}

func TestScreen_moveCursor_clamps(t *testing.T) {
	s := NewScreen(10, 3)
	s.moveCursor(-5, -5)
	assert.Equal(t, 0, s.CursorX)
	assert.Equal(t, 0, s.CursorY)
}

func TestScreen_moveUp(t *testing.T) {
	s := NewScreen(10, 5)
	s.CursorY = 3
	s.moveUp(2)
	assert.Equal(t, 1, s.CursorY)
}

func TestScreen_moveUp_clamps(t *testing.T) {
	s := NewScreen(10, 5)
	s.CursorY = 0
	s.moveUp(5)
	assert.Equal(t, 0, s.CursorY)
}

func TestScreen_moveDown(t *testing.T) {
	s := NewScreen(10, 5)
	s.CursorY = 1
	s.moveDown(2)
	assert.Equal(t, 3, s.CursorY)
}

func TestScreen_moveDown_clamps(t *testing.T) {
	s := NewScreen(10, 5)
	s.CursorY = 4
	s.moveDown(5)
	assert.Equal(t, 4, s.CursorY)
}

func TestScreen_moveLeft(t *testing.T) {
	s := NewScreen(10, 3)
	s.CursorX = 5
	s.moveLeft(3)
	assert.Equal(t, 2, s.CursorX)
}

func TestScreen_moveLeft_clamps(t *testing.T) {
	s := NewScreen(10, 3)
	s.CursorX = 0
	s.moveLeft(5)
	assert.Equal(t, 0, s.CursorX)
}

func TestScreen_moveRight(t *testing.T) {
	s := NewScreen(10, 3)
	s.CursorX = 2
	s.moveRight(3)
	assert.Equal(t, 5, s.CursorX)
}

func TestScreen_moveRight_clamps(t *testing.T) {
	s := NewScreen(10, 3)
	s.CursorX = 9
	s.moveRight(5)
	assert.Equal(t, 9, s.CursorX)
}

func TestScreen_clear(t *testing.T) {
	s := NewScreen(5, 3)
	s.writeChar('A', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('B', ColorDefault, ColorDefault, Attributes{})
	s.CursorX = 2
	s.CursorY = 1
	s.writeChar('C', ColorDefault, ColorDefault, Attributes{})

	s.clear()
	assert.Equal(t, 0, s.CursorX)
	assert.Equal(t, 0, s.CursorY)
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			assert.Equal(t, defaultCell, s.Grid[y][x])
		}
	}
}

func TestScreen_clearToEndOfLine(t *testing.T) {
	s := NewScreen(5, 3)
	s.writeChar('A', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('B', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('C', ColorDefault, ColorDefault, Attributes{})

	s.CursorX = 1
	s.clearToEndOfLine()
	assert.Equal(t, 'A', s.Grid[0][0].Char)
	assert.Equal(t, defaultCell, s.Grid[0][1])
	assert.Equal(t, defaultCell, s.Grid[0][2])
}

func TestScreen_clearToBeginningOfLine(t *testing.T) {
	s := NewScreen(5, 3)
	s.writeChar('A', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('B', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('C', ColorDefault, ColorDefault, Attributes{})

	s.CursorX = 1
	s.clearToBeginningOfLine()
	assert.Equal(t, defaultCell, s.Grid[0][0])
	assert.Equal(t, defaultCell, s.Grid[0][1])
	requireCellChar(t, s, 2, 0, 'C')
}

func TestScreen_clearToEndOfScreen(t *testing.T) {
	s := NewScreen(5, 3)
	s.writeChar('A', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('B', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('C', ColorDefault, ColorDefault, Attributes{})
	s.newline()
	s.writeChar('D', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('E', ColorDefault, ColorDefault, Attributes{})
	s.newline()
	s.writeChar('F', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('G', ColorDefault, ColorDefault, Attributes{})

	s.CursorX = 1
	s.CursorY = 1
	s.clearToEndOfScreen()

	assert.Equal(t, 'A', s.Grid[0][0].Char)
	assert.Equal(t, 'B', s.Grid[0][1].Char)
	assert.Equal(t, 'C', s.Grid[0][2].Char)
	assert.Equal(t, 'D', s.Grid[1][0].Char)
	assert.Equal(t, defaultCell, s.Grid[1][1])
	assert.Equal(t, defaultCell, s.Grid[1][2])
	for x := 0; x < s.Width; x++ {
		assert.Equal(t, defaultCell, s.Grid[2][x])
	}
}

func TestScreen_clearToBeginningOfScreen(t *testing.T) {
	s := NewScreen(5, 3)
	s.writeChar('A', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('B', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('C', ColorDefault, ColorDefault, Attributes{})
	s.newline()
	s.writeChar('D', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('E', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('X', ColorDefault, ColorDefault, Attributes{})
	s.newline()
	s.writeChar('F', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('G', ColorDefault, ColorDefault, Attributes{})

	s.CursorX = 1
	s.CursorY = 1
	s.clearToBeginningOfScreen()

	for x := 0; x < s.Width; x++ {
		assert.Equal(t, defaultCell, s.Grid[0][x])
	}
	assert.Equal(t, defaultCell, s.Grid[1][0])
	assert.Equal(t, defaultCell, s.Grid[1][1])
	requireCellChar(t, s, 2, 1, 'X')
	requireCellChar(t, s, 0, 2, 'F')
	requireCellChar(t, s, 1, 2, 'G')
}

func TestScreen_clearLine(t *testing.T) {
	s := NewScreen(5, 3)
	s.writeChar('A', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('B', ColorDefault, ColorDefault, Attributes{})

	s.CursorX = 3
	s.clearLine()
	assert.Equal(t, 3, s.CursorX)
	for x := 0; x < s.Width; x++ {
		assert.Equal(t, defaultCell, s.Grid[0][x])
	}
}

func TestScreen_backspace(t *testing.T) {
	s := NewScreen(10, 3)
	s.CursorX = 5
	s.backspace()
	assert.Equal(t, 4, s.CursorX)
}

func TestScreen_backspace_atStart(t *testing.T) {
	s := NewScreen(10, 3)
	s.CursorX = 0
	s.backspace()
	assert.Equal(t, 0, s.CursorX)
}

func TestScreen_tab(t *testing.T) {
	s := NewScreen(40, 3)
	s.CursorX = 0
	s.tab()
	assert.Equal(t, 8, s.CursorX)

	s.CursorX = 5
	s.tab()
	assert.Equal(t, 8, s.CursorX)

	s.CursorX = 8
	s.tab()
	assert.Equal(t, 16, s.CursorX)
}

func TestScreen_tab_clamps(t *testing.T) {
	s := NewScreen(10, 3)
	s.CursorX = 9
	s.tab()
	assert.Equal(t, 9, s.CursorX)
}

func TestScreen_scrollUp(t *testing.T) {
	s := NewScreen(5, 4)
	s.writeChar('0', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('1', ColorDefault, ColorDefault, Attributes{})
	s.newline()
	s.writeChar('2', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('3', ColorDefault, ColorDefault, Attributes{})
	s.newline()
	s.writeChar('4', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('5', ColorDefault, ColorDefault, Attributes{})
	s.newline()
	s.writeChar('6', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('7', ColorDefault, ColorDefault, Attributes{})

	s.scrollUp(1)

	assert.Equal(t, "23...", cellRowString(s.Grid[0]))
	assert.Equal(t, "45...", cellRowString(s.Grid[1]))
	assert.Equal(t, "67...", cellRowString(s.Grid[2]))
	for x := 0; x < s.Width; x++ {
		assert.Equal(t, defaultCell, s.Grid[3][x])
	}
}

func TestScreen_scrollUp_fullScreen(t *testing.T) {
	s := NewScreen(5, 3)
	s.writeChar('A', ColorDefault, ColorDefault, Attributes{})
	s.newline()
	s.writeChar('B', ColorDefault, ColorDefault, Attributes{})
	s.newline()
	s.writeChar('C', ColorDefault, ColorDefault, Attributes{})

	s.scrollUp(3)

	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			assert.Equal(t, defaultCell, s.Grid[y][x])
		}
	}
}

func TestScreen_scrollDown(t *testing.T) {
	s := NewScreen(5, 4)
	s.writeChar('0', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('1', ColorDefault, ColorDefault, Attributes{})
	s.newline()
	s.writeChar('2', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('3', ColorDefault, ColorDefault, Attributes{})
	s.newline()
	s.writeChar('4', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('5', ColorDefault, ColorDefault, Attributes{})
	s.newline()
	s.writeChar('6', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('7', ColorDefault, ColorDefault, Attributes{})

	s.scrollDown(1)

	for x := 0; x < s.Width; x++ {
		assert.Equal(t, defaultCell, s.Grid[0][x])
	}
	assert.Equal(t, "01...", cellRowString(s.Grid[1]))
	assert.Equal(t, "23...", cellRowString(s.Grid[2]))
	assert.Equal(t, "45...", cellRowString(s.Grid[3]))
}

func TestScreen_setScrollRegion(t *testing.T) {
	s := NewScreen(10, 10)
	s.setScrollRegion(2, 7)
	assert.Equal(t, 2, s.ScrollTop)
	assert.Equal(t, 7, s.ScrollBottom)
	assert.Equal(t, 0, s.CursorX)
	assert.Equal(t, 0, s.CursorY)
}

func TestScreen_setScrollRegion_invalid(t *testing.T) {
	s := NewScreen(10, 10)
	s.setScrollRegion(5, 3)
	assert.Equal(t, 0, s.ScrollTop)
	assert.Equal(t, 9, s.ScrollBottom)
}

func TestScreen_insertLines(t *testing.T) {
	s := NewScreen(5, 4)
	s.Grid[0][0] = Cell{Char: '0'}
	s.Grid[1][0] = Cell{Char: '1'}
	s.Grid[2][0] = Cell{Char: '2'}
	s.Grid[3][0] = Cell{Char: '3'}

	s.CursorY = 1
	s.insertLines(1)

	assert.Equal(t, '0', s.Grid[0][0].Char)
	for x := 0; x < s.Width; x++ {
		assert.Equal(t, defaultCell, s.Grid[1][x])
	}
	assert.Equal(t, '1', s.Grid[2][0].Char)
	assert.Equal(t, '2', s.Grid[3][0].Char)
}

func TestScreen_deleteLines(t *testing.T) {
	s := NewScreen(5, 4)
	s.Grid[0][0] = Cell{Char: '0'}
	s.Grid[1][0] = Cell{Char: '1'}
	s.Grid[2][0] = Cell{Char: '2'}
	s.Grid[3][0] = Cell{Char: '3'}

	s.CursorY = 1
	s.deleteLines(1)

	assert.Equal(t, '0', s.Grid[0][0].Char)
	assert.Equal(t, '2', s.Grid[1][0].Char)
	assert.Equal(t, '3', s.Grid[2][0].Char)
	for x := 0; x < s.Width; x++ {
		assert.Equal(t, defaultCell, s.Grid[3][x])
	}
}

func TestScreen_insertChars(t *testing.T) {
	s := NewScreen(5, 1)
	s.writeChar('A', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('B', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('C', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('D', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('E', ColorDefault, ColorDefault, Attributes{})

	s.CursorX = 1
	s.insertChars(1)

	assert.Equal(t, 'A', s.Grid[0][0].Char)
	assert.Equal(t, defaultCell, s.Grid[0][1])
	assert.Equal(t, 'B', s.Grid[0][2].Char)
	assert.Equal(t, 'C', s.Grid[0][3].Char)
}

func TestScreen_deleteChars(t *testing.T) {
	s := NewScreen(5, 1)
	s.writeChar('A', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('B', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('C', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('D', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('E', ColorDefault, ColorDefault, Attributes{})

	s.CursorX = 1
	s.deleteChars(1)

	assert.Equal(t, 'A', s.Grid[0][0].Char)
	assert.Equal(t, 'C', s.Grid[0][1].Char)
	assert.Equal(t, 'D', s.Grid[0][2].Char)
	assert.Equal(t, 'E', s.Grid[0][3].Char)
	assert.Equal(t, defaultCell, s.Grid[0][4])
}

func TestScreen_eraseChars(t *testing.T) {
	s := NewScreen(5, 1)
	s.writeChar('A', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('B', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('C', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('D', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('E', ColorDefault, ColorDefault, Attributes{})

	s.CursorX = 1
	s.eraseChars(2)

	assert.Equal(t, 'A', s.Grid[0][0].Char)
	assert.Equal(t, defaultCell, s.Grid[0][1])
	assert.Equal(t, defaultCell, s.Grid[0][2])
	assert.Equal(t, 'D', s.Grid[0][3].Char)
}

func TestScreen_resize(t *testing.T) {
	s := NewScreen(5, 3)
	s.writeChar('A', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('B', ColorDefault, ColorDefault, Attributes{})
	s.newline()
	s.writeChar('C', ColorDefault, ColorDefault, Attributes{})

	s.resize(8, 5)

	assert.Equal(t, 8, s.Width)
	assert.Equal(t, 5, s.Height)
	assert.Equal(t, 'A', s.Grid[0][0].Char)
	assert.Equal(t, 'B', s.Grid[0][1].Char)
	assert.Equal(t, 'C', s.Grid[1][0].Char)
	assert.Equal(t, 4, s.ScrollBottom)
}

func TestScreen_resize_smaller(t *testing.T) {
	s := NewScreen(10, 5)
	s.writeChar('A', ColorDefault, ColorDefault, Attributes{})
	s.writeChar('B', ColorDefault, ColorDefault, Attributes{})

	s.resize(5, 3)

	assert.Equal(t, 5, s.Width)
	assert.Equal(t, 3, s.Height)
	assert.Equal(t, 'A', s.Grid[0][0].Char)
	assert.Equal(t, 'B', s.Grid[0][1].Char)
	assert.Equal(t, 2, s.CursorX)
	assert.Equal(t, 0, s.CursorY)
}

func TestScreen_resize_clampsCursor(t *testing.T) {
	s := NewScreen(10, 5)
	s.CursorX = 8
	s.CursorY = 4

	s.resize(5, 3)

	assert.Equal(t, 4, s.CursorX)
	assert.Equal(t, 2, s.CursorY)
}

func TestScreen_scrollRegion_scrollsOnlyWithin(t *testing.T) {
	s := NewScreen(5, 5)
	s.setScrollRegion(1, 3)

	for i := 0; i < 3; i++ {
		s.CursorX = 0
		s.CursorY = 1 + i
		s.writeChar(rune('A'+i), ColorDefault, ColorDefault, Attributes{})
	}

	s.CursorY = 3
	s.scrollUp(1)

	assert.Equal(t, 'B', s.Grid[1][0].Char)
	assert.Equal(t, 'C', s.Grid[2][0].Char)
	for x := 0; x < s.Width; x++ {
		assert.Equal(t, defaultCell, s.Grid[3][x])
	}
}

func cellRowString(row []Cell) string {
	b := make([]byte, len(row))
	for i, c := range row {
		if c.Char == 0 {
			b[i] = '.'
		} else {
			b[i] = byte(c.Char)
		}
	}
	return string(b)
}

func requireCellChar(t *testing.T, s *Screen, x, y int, expected rune) {
	t.Helper()
	require.Less(t, y, s.Height, "y out of bounds")
	require.Less(t, x, s.Width, "x out of bounds")
	assert.Equal(t, expected, s.Grid[y][x].Char, "cell(%d,%d)", x, y)
}
