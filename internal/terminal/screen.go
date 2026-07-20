package terminal

var defaultCell = Cell{}

type Cell struct {
	Char       rune
	Foreground Color
	Background Color
	Attributes Attributes
}

type Attributes struct {
	Bold      bool
	Dim       bool
	Italic    bool
	Underline bool
	Blink     bool
	Reverse   bool
	Strike    bool
}

type Screen struct {
	Width  int
	Height int
	Grid   [][]Cell

	CursorX       int
	CursorY       int
	CursorVisible bool
	SavedX        int
	SavedY        int

	ScrollTop    int
	ScrollBottom int
}

func NewScreen(width, height int) *Screen {
	s := &Screen{
		Width:         width,
		Height:        height,
		CursorVisible: true,
		ScrollTop:     0,
		ScrollBottom:  height - 1,
	}
	s.Grid = make([][]Cell, height)
	for y := 0; y < height; y++ {
		s.Grid[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			s.Grid[y][x] = defaultCell
		}
	}
	return s
}

func (s *Screen) writeChar(ch rune, fg, bg Color, attrs Attributes) {
	if s.CursorX >= s.Width {
		s.newline()
	}
	if s.CursorY >= 0 && s.CursorY < s.Height && s.CursorX >= 0 && s.CursorX < s.Width {
		s.Grid[s.CursorY][s.CursorX] = Cell{
			Char:       ch,
			Foreground: fg,
			Background: bg,
			Attributes: attrs,
		}
	}
	s.CursorX++
}

func (s *Screen) newline() {
	s.CursorX = 0
	if s.CursorY < s.ScrollBottom {
		s.CursorY++
	} else if s.CursorY == s.ScrollBottom {
		s.scrollUp(1)
	}
}

func (s *Screen) moveCursor(x, y int) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	s.CursorX = x
	s.CursorY = y
}

func (s *Screen) moveUp(lines int) {
	s.CursorY -= lines
	if s.CursorY < s.ScrollTop {
		s.CursorY = s.ScrollTop
	}
}

func (s *Screen) moveDown(lines int) {
	s.CursorY += lines
	if s.CursorY > s.ScrollBottom {
		s.CursorY = s.ScrollBottom
	}
}

func (s *Screen) moveLeft(cols int) {
	s.CursorX -= cols
	if s.CursorX < 0 {
		s.CursorX = 0
	}
}

func (s *Screen) moveRight(cols int) {
	s.CursorX += cols
	if s.CursorX >= s.Width {
		s.CursorX = s.Width - 1
	}
}

func (s *Screen) clear() {
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			s.Grid[y][x] = defaultCell
		}
	}
	s.CursorX = 0
	s.CursorY = 0
}

func (s *Screen) clearToEndOfLine() {
	for x := s.CursorX; x < s.Width; x++ {
		s.Grid[s.CursorY][x] = defaultCell
	}
}

func (s *Screen) clearToBeginningOfLine() {
	for x := 0; x <= s.CursorX; x++ {
		s.Grid[s.CursorY][x] = defaultCell
	}
}

func (s *Screen) clearToEndOfScreen() {
	for x := s.CursorX; x < s.Width; x++ {
		s.Grid[s.CursorY][x] = defaultCell
	}
	for y := s.CursorY + 1; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			s.Grid[y][x] = defaultCell
		}
	}
}

func (s *Screen) clearToBeginningOfScreen() {
	for x := 0; x <= s.CursorX; x++ {
		s.Grid[s.CursorY][x] = defaultCell
	}
	for y := 0; y < s.CursorY; y++ {
		for x := 0; x < s.Width; x++ {
			s.Grid[y][x] = defaultCell
		}
	}
}

func (s *Screen) clearLine() {
	for x := 0; x < s.Width; x++ {
		s.Grid[s.CursorY][x] = defaultCell
	}
}

func (s *Screen) backspace() {
	if s.CursorX > 0 {
		s.CursorX--
	}
}

func (s *Screen) tab() {
	s.CursorX = (s.CursorX/8 + 1) * 8
	if s.CursorX >= s.Width {
		s.CursorX = s.Width - 1
	}
}

func (s *Screen) scrollUp(n int) {
	if n <= 0 || s.ScrollTop >= s.ScrollBottom {
		return
	}
	if n > s.ScrollBottom-s.ScrollTop+1 {
		n = s.ScrollBottom - s.ScrollTop + 1
	}
	for y := s.ScrollTop; y <= s.ScrollBottom-n; y++ {
		copy(s.Grid[y], s.Grid[y+n])
	}
	for y := s.ScrollBottom - n + 1; y <= s.ScrollBottom; y++ {
		for x := 0; x < s.Width; x++ {
			s.Grid[y][x] = defaultCell
		}
	}
}

func (s *Screen) scrollDown(n int) {
	if n <= 0 || s.ScrollTop >= s.ScrollBottom {
		return
	}
	if n > s.ScrollBottom-s.ScrollTop+1 {
		n = s.ScrollBottom - s.ScrollTop + 1
	}
	for y := s.ScrollBottom; y >= s.ScrollTop+n; y-- {
		copy(s.Grid[y], s.Grid[y-n])
	}
	for y := s.ScrollTop; y < s.ScrollTop+n; y++ {
		for x := 0; x < s.Width; x++ {
			s.Grid[y][x] = defaultCell
		}
	}
}

func (s *Screen) setScrollRegion(top, bottom int) {
	if top < 0 {
		top = 0
	}
	if bottom >= s.Height {
		bottom = s.Height - 1
	}
	if top >= bottom {
		return
	}
	s.ScrollTop = top
	s.ScrollBottom = bottom
	s.CursorX = 0
	s.CursorY = 0
}

func (s *Screen) insertLines(n int) {
	if s.CursorY < s.ScrollTop || s.CursorY > s.ScrollBottom {
		return
	}
	if n <= 0 {
		n = 1
	}
	lines := s.ScrollBottom - s.CursorY
	if n > lines {
		n = lines
	}
	for y := s.ScrollBottom; y >= s.CursorY+n; y-- {
		copy(s.Grid[y], s.Grid[y-n])
	}
	for y := s.CursorY; y < s.CursorY+n; y++ {
		for x := 0; x < s.Width; x++ {
			s.Grid[y][x] = defaultCell
		}
	}
}

func (s *Screen) deleteLines(n int) {
	if s.CursorY < s.ScrollTop || s.CursorY > s.ScrollBottom {
		return
	}
	if n <= 0 {
		n = 1
	}
	lines := s.ScrollBottom - s.CursorY
	if n > lines {
		n = lines
	}
	for y := s.CursorY; y <= s.ScrollBottom-n; y++ {
		copy(s.Grid[y], s.Grid[y+n])
	}
	for y := s.ScrollBottom - n + 1; y <= s.ScrollBottom; y++ {
		for x := 0; x < s.Width; x++ {
			s.Grid[y][x] = defaultCell
		}
	}
}

func (s *Screen) deleteChars(n int) {
	if n <= 0 {
		n = 1
	}
	if s.CursorX+n > s.Width {
		n = s.Width - s.CursorX
	}
	for x := s.CursorX; x < s.Width-n; x++ {
		s.Grid[s.CursorY][x] = s.Grid[s.CursorY][x+n]
	}
	for x := s.Width - n; x < s.Width; x++ {
		s.Grid[s.CursorY][x] = defaultCell
	}
}

func (s *Screen) insertChars(n int) {
	if n <= 0 {
		n = 1
	}
	if s.CursorX+n > s.Width {
		n = s.Width - s.CursorX
	}
	for x := s.Width - 1; x >= s.CursorX+n; x-- {
		s.Grid[s.CursorY][x] = s.Grid[s.CursorY][x-n]
	}
	for x := s.CursorX; x < s.CursorX+n; x++ {
		s.Grid[s.CursorY][x] = defaultCell
	}
}

func (s *Screen) eraseChars(n int) {
	if n <= 0 {
		n = 1
	}
	end := s.CursorX + n
	if end > s.Width {
		end = s.Width
	}
	for x := s.CursorX; x < end; x++ {
		s.Grid[s.CursorY][x] = defaultCell
	}
}

func (s *Screen) resize(width, height int) {
	if s.Width == width && s.Height == height {
		return
	}
	newGrid := make([][]Cell, height)
	for y := 0; y < height; y++ {
		newGrid[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			if y < s.Height && x < s.Width {
				newGrid[y][x] = s.Grid[y][x]
			} else {
				newGrid[y][x] = defaultCell
			}
		}
	}
	s.Width = width
	s.Height = height
	s.Grid = newGrid
	if s.CursorX >= width {
		s.CursorX = width - 1
	}
	if s.CursorY >= height {
		s.CursorY = height - 1
	}
	s.ScrollBottom = height - 1
}
