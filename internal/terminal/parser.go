package terminal

import (
	"github.com/gdamore/tcell/v2"
)

var ansiColors = map[int]tcell.Color{
	30: tcell.ColorBlack,
	31: tcell.ColorMaroon,
	32: tcell.ColorGreen,
	33: tcell.ColorOlive,
	34: tcell.ColorNavy,
	35: tcell.ColorPurple,
	36: tcell.ColorTeal,
	37: tcell.ColorSilver,
	90: tcell.ColorGray,
	91: tcell.ColorRed,
	92: tcell.ColorLime,
	93: tcell.ColorYellow,
	94: tcell.ColorBlue,
	95: tcell.ColorFuchsia,
	96: tcell.ColorAqua,
	97: tcell.ColorWhite,
}

var ansiBgColors = map[int]tcell.Color{
	40:  tcell.ColorBlack,
	41:  tcell.ColorMaroon,
	42:  tcell.ColorGreen,
	43:  tcell.ColorOlive,
	44:  tcell.ColorNavy,
	45:  tcell.ColorPurple,
	46:  tcell.ColorTeal,
	47:  tcell.ColorSilver,
	100: tcell.ColorGray,
	101: tcell.ColorRed,
	102: tcell.ColorLime,
	103: tcell.ColorYellow,
	104: tcell.ColorBlue,
	105: tcell.ColorFuchsia,
	106: tcell.ColorAqua,
	107: tcell.ColorWhite,
}

type Parser struct {
	state  int
	params []int
	buf    []byte
	buffer *Buffer

	currentFg    tcell.Color
	currentBg    tcell.Color
	currentAttrs Attributes

	savedX int
	savedY int

	handler      func(rune, tcell.Color, tcell.Color, Attributes)
	TitleHandler func(string)
}

const (
	stateGround = iota
	stateEscape
	stateEscapeIntermediate
	stateCSI
	stateCSIParam
	stateCSIIntermediate
	stateOSC
)

func NewParser(buf *Buffer, handler func(rune, tcell.Color, tcell.Color, Attributes)) *Parser {
	return &Parser{
		state:     stateGround,
		buffer:    buf,
		params:    make([]int, 0, 16),
		currentFg: tcell.ColorDefault,
		currentBg: tcell.ColorDefault,
		handler:   handler,
	}
}

func (p *Parser) Parse(data []byte) {
	for _, b := range data {
		p.parseByte(byte(b))
	}
}

func (p *Parser) parseByte(b byte) {
	switch p.state {
	case stateGround:
		p.handleGround(b)
	case stateEscape:
		p.handleEscape(b)
	case stateEscapeIntermediate:
		p.handleEscapeIntermediate(b)
	case stateCSI:
		p.handleCSI(b)
	case stateCSIParam:
		p.handleCSIParam(b)
	case stateCSIIntermediate:
		p.handleCSIIntermediate(b)
	case stateOSC:
		p.handleOSC(b)
	}
}

func (p *Parser) handleGround(b byte) {
	switch b {
	case 0x1B:
		p.state = stateEscape
		p.params = p.params[:0]
	case 0x07:
		// Bell - ignore for now
	case 0x08:
		// Backspace
		p.buffer.Backspace()
	case 0x09:
		// Tab
		p.buffer.Tab()
	case 0x0A, 0x0B, 0x0C:
		// Line feed, vertical tab, form feed
		p.buffer.Newline()
	case 0x0D:
		// Carriage return
		p.buffer.CursorX = 0
	case 0x7F:
		// Delete - ignore
	default:
		if b >= 0x20 {
			p.emitChar(rune(b))
		}
	}
}

func (p *Parser) handleEscape(b byte) {
	switch b {
	case '[':
		p.state = stateCSI
		p.params = p.params[:0]
	case ']':
		p.state = stateOSC
		p.buf = p.buf[:0]
	case '7':
		p.savedX = p.buffer.CursorX
		p.savedY = p.buffer.CursorY
		p.state = stateGround
	case '8':
		p.buffer.MoveCursor(p.savedX, p.savedY)
		p.state = stateGround
	case 'D':
		p.buffer.MoveDown(1)
		p.state = stateGround
	case 'M':
		p.buffer.MoveUp(1)
		p.state = stateGround
	case 'c':
		p.buffer.Clear()
		p.currentFg = tcell.ColorDefault
		p.currentBg = tcell.ColorDefault
		p.currentAttrs = Attributes{}
		p.state = stateGround
	case 'P', 'X', '^', '=': // DCS, SOS, PM, APC - ignore until ST
		p.buf = p.buf[:0]
		p.state = stateGround
	default:
		p.state = stateGround
	}
}

func (p *Parser) handleEscapeIntermediate(b byte) {
	p.state = stateGround
}

func (p *Parser) handleCSIIntermediate(b byte) {
	p.state = stateGround
}

func (p *Parser) handleOSC(b byte) {
	switch b {
	case 0x07:
		p.executeOSC()
		p.state = stateGround
	case 0x1B:
		p.state = stateEscape
	default:
		p.buf = append(p.buf, b)
	}
}

func (p *Parser) executeOSC() {
	if len(p.buf) < 2 {
		return
	}

	oscType := 0
	i := 0

	for i < len(p.buf) && p.buf[i] >= '0' && p.buf[i] <= '9' {
		oscType = oscType*10 + int(p.buf[i]-'0')
		i++
	}

	if i < len(p.buf) && p.buf[i] == ';' {
		i++
	}

	title := string(p.buf[i:])

	switch oscType {
	case 0, 1, 2:
		if p.TitleHandler != nil {
			p.TitleHandler(title)
		}
	}
}

func (p *Parser) handleCSI(b byte) {
	if b >= '0' && b <= '9' {
		p.state = stateCSIParam
		p.params = append(p.params, int(b-'0'))
		return
	}

	if b == ';' {
		p.state = stateCSIParam
		return
	}

	if b >= 0x40 && b < 0x80 {
		p.executeCSI(b, p.params)
		p.state = stateGround
	}
}

func (p *Parser) handleCSIParam(b byte) {
	if b >= '0' && b <= '9' {
		if len(p.params) == 0 {
			p.params = append(p.params, 0)
		}
		last := len(p.params) - 1
		if p.params[last] == 0 && !p.isInMiddleOfParam() {
			p.params[last] = int(b - '0')
		} else {
			p.params[last] = p.params[last]*10 + int(b-'0')
		}
		return
	}

	if b == ';' {
		p.params = append(p.params, 0)
		return
	}

	if b >= 0x40 && b < 0x80 {
		p.executeCSI(b, p.params)
		p.state = stateGround
	}
}

func (p *Parser) isInMiddleOfParam() bool {
	for _, v := range p.params {
		if v != 0 {
			return true
		}
	}
	return false
}

func (p *Parser) executeCSI(cmd byte, params []int) {
	getParam := func(idx, defaultVal int) int {
		if idx < len(params) {
			return params[idx]
		}
		return defaultVal
	}

	switch cmd {
	case 'm':
		p.handleSGR(params)
	case 'H', 'f':
		y := getParam(0, 1)
		x := getParam(1, 1)
		p.buffer.MoveCursor(x-1, y-1)
	case 'A':
		p.buffer.MoveUp(getParam(0, 1))
	case 'B':
		p.buffer.MoveDown(getParam(0, 1))
	case 'C':
		p.buffer.MoveRight(getParam(0, 1))
	case 'D':
		p.buffer.MoveLeft(getParam(0, 1))
	case 'J':
		mode := getParam(0, 0)
		switch mode {
		case 0:
			p.buffer.ClearToEnd()
		case 1:
			p.buffer.ClearToBeginning()
		case 2, 3:
			p.buffer.Clear()
		}
	case 'K':
		mode := getParam(0, 0)
		switch mode {
		case 0:
			p.buffer.ClearToEnd()
		case 1:
			p.buffer.ClearToBeginning()
		case 2:
			p.buffer.ClearLine()
		}
	}
}

func (p *Parser) handleSGR(params []int) {
	if len(params) == 0 {
		params = []int{0}
	}

	i := 0
	for i < len(params) {
		code := params[i]
		i++

		switch {
		case code == 0:
			p.currentFg = tcell.ColorDefault
			p.currentBg = tcell.ColorDefault
			p.currentAttrs = Attributes{}

		case code == 1:
			p.currentAttrs.Bold = true
		case code == 2:
			p.currentAttrs.Dim = true
		case code == 3:
			p.currentAttrs.Italic = true
		case code == 4:
			p.currentAttrs.Underline = true
		case code == 5:
			p.currentAttrs.Blink = true
		case code == 7:
			p.currentAttrs.Reverse = true
		case code == 9:
			p.currentAttrs.Strike = true

		case code == 21 || code == 22:
			p.currentAttrs.Bold = false
			p.currentAttrs.Dim = false
		case code == 23:
			p.currentAttrs.Italic = false
		case code == 24:
			p.currentAttrs.Underline = false
		case code == 25:
			p.currentAttrs.Blink = false
		case code == 27:
			p.currentAttrs.Reverse = false
		case code == 29:
			p.currentAttrs.Strike = false

		case code >= 30 && code <= 37:
			p.currentFg = ansiColors[code]
		case code == 39:
			p.currentFg = tcell.ColorDefault
		case code >= 40 && code <= 47:
			p.currentBg = ansiBgColors[code]
		case code == 49:
			p.currentBg = tcell.ColorDefault

		case code >= 90 && code <= 97:
			p.currentFg = ansiColors[code]
		case code >= 100 && code <= 107:
			p.currentBg = ansiBgColors[code]

		case code == 38 && i < len(params):
			if params[i] == 5 && i+1 < len(params) {
				p.currentFg = tcell.Color(uint32(params[i+1]) + 1)
				i += 2
			} else if params[i] == 2 && i+3 < len(params) {
				r, g, b := params[i+1], params[i+2], params[i+3]
				p.currentFg = tcell.NewRGBColor(int32(r), int32(g), int32(b))
				i += 4
			}

		case code == 48 && i < len(params):
			if params[i] == 5 && i+1 < len(params) {
				p.currentBg = tcell.Color(uint32(params[i+1]) + 1)
				i += 2
			} else if params[i] == 2 && i+3 < len(params) {
				r, g, b := params[i+1], params[i+2], params[i+3]
				p.currentBg = tcell.NewRGBColor(int32(r), int32(g), int32(b))
				i += 4
			}
		}
	}
}

func (p *Parser) emitChar(ch rune) {
	p.buffer.WriteChar(ch, p.currentFg, p.currentBg, p.currentAttrs)
	if p.handler != nil {
		p.handler(ch, p.currentFg, p.currentBg, p.currentAttrs)
	}
}
