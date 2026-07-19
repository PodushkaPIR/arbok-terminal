package terminal

import (
	"fmt"
	"log/slog"
	"unicode/utf8"
)

var ansiColors = map[int]Color{
	30: ColorRGB(0, 0, 0),
	31: ColorRGB(128, 0, 0),
	32: ColorRGB(0, 128, 0),
	33: ColorRGB(128, 128, 0),
	34: ColorRGB(0, 0, 128),
	35: ColorRGB(128, 0, 128),
	36: ColorRGB(0, 128, 128),
	37: ColorRGB(192, 192, 192),
	90: ColorRGB(128, 128, 128),
	91: ColorRGB(255, 0, 0),
	92: ColorRGB(0, 255, 0),
	93: ColorRGB(255, 255, 0),
	94: ColorRGB(0, 0, 255),
	95: ColorRGB(255, 0, 255),
	96: ColorRGB(0, 255, 255),
	97: ColorRGB(255, 255, 255),
}

var ansiBgColors = map[int]Color{
	40:  ColorRGB(0, 0, 0),
	41:  ColorRGB(128, 0, 0),
	42:  ColorRGB(0, 128, 0),
	43:  ColorRGB(128, 128, 0),
	44:  ColorRGB(0, 0, 128),
	45:  ColorRGB(128, 0, 128),
	46:  ColorRGB(0, 128, 128),
	47:  ColorRGB(192, 192, 192),
	100: ColorRGB(128, 128, 128),
	101: ColorRGB(255, 0, 0),
	102: ColorRGB(0, 255, 0),
	103: ColorRGB(255, 255, 0),
	104: ColorRGB(0, 0, 255),
	105: ColorRGB(255, 0, 255),
	106: ColorRGB(0, 255, 255),
	107: ColorRGB(255, 255, 255),
}

type Parser struct {
	state   int
	params  []int
	buf     []byte
	utf8Buf []byte
	vt      *VirtualTerminal

	currentFg    Color
	currentBg    Color
	currentAttrs Attributes

	paramStarted bool
	privateMode   bool
	TitleHandler  func(string)
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

func NewParser(vt *VirtualTerminal) *Parser {
	return &Parser{
		state:     stateGround,
		vt:        vt,
		params:    make([]int, 0, 16),
		currentFg: ColorDefault,
		currentBg: ColorDefault,
	}
}

func (p *Parser) Parse(data []byte) {
	slog.Debug("parser: parse", "bytes", len(data), "hex", fmt.Sprintf("%x", data))
	p.vt.mu.Lock()
	for _, b := range data {
		p.parseByte(byte(b))
	}
	p.vt.mu.Unlock()
}

func (p *Parser) SetTitle(title string) {
	if p.TitleHandler != nil {
		p.TitleHandler(title)
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
	if len(p.utf8Buf) > 0 {
		p.utf8Buf = append(p.utf8Buf, b)
		if r, _ := utf8.DecodeRune(p.utf8Buf); r != utf8.RuneError || len(p.utf8Buf) >= 4 {
			if r != utf8.RuneError {
				p.emitChar(r)
			}
			p.utf8Buf = p.utf8Buf[:0]
		}
		return
	}

	switch b {
	case 0x1B:
		p.state = stateEscape
		p.params = p.params[:0]
	case 0x07:
		// Bell - ignore for now
	case 0x08:
		p.vt.Backspace()
	case 0x09:
		p.vt.Tab()
	case 0x0A, 0x0B, 0x0C:
		p.vt.LineFeed()
	case 0x0D:
		p.vt.CarriageReturn()
	case 0x7F:
		// Delete - ignore
	default:
		if b >= 0x20 {
			if b < 0x80 {
				p.emitChar(rune(b))
			} else {
				p.utf8Buf = append(p.utf8Buf[:0], b)
			}
		}
	}
}

func (p *Parser) handleEscape(b byte) {
	switch b {
	case '[':
		p.state = stateCSI
		p.params = p.params[:0]
		p.privateMode = false
	case ']':
		p.state = stateOSC
		p.buf = p.buf[:0]
	case '7':
		p.vt.SaveCursor()
		p.state = stateGround
	case '8':
		p.vt.RestoreCursor()
		p.state = stateGround
	case 'D':
		p.vt.CursorDownLine(1)
		p.state = stateGround
	case 'M':
		p.vt.CursorUpLine(1)
		p.state = stateGround
	case 'c':
		p.vt.Clear()
		p.currentFg = ColorDefault
		p.currentBg = ColorDefault
		p.currentAttrs = Attributes{}
		p.state = stateGround
	case 'P', 'X', '^', '=':
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
		p.vt.SetPendingTitle(title)
	}
}

func (p *Parser) handleCSI(b byte) {
	if b == '?' || b == '>' || b == '=' {
		p.privateMode = true
		return
	}

	if b >= '0' && b <= '9' {
		p.state = stateCSIParam
		p.params = append(p.params, int(b-'0'))
		p.paramStarted = b != '0'
		return
	}

	if b == ';' {
		p.state = stateCSIParam
		return
	}

	if b >= 0x40 && b < 0x80 {
		if p.privateMode {
			p.executePrivateMode(b, p.params)
		} else {
			p.executeCSI(b, p.params)
		}
		p.privateMode = false
		p.state = stateGround
	}
}

func (p *Parser) handleCSIParam(b byte) {
	if b >= '0' && b <= '9' {
		if len(p.params) == 0 {
			p.params = append(p.params, 0)
		}
		last := len(p.params) - 1
		if p.params[last] == 0 && !p.paramStarted {
			p.params[last] = int(b - '0')
		} else {
			p.params[last] = p.params[last]*10 + int(b-'0')
		}
		if b != '0' {
			p.paramStarted = true
		}
		return
	}

	if b == ';' {
		p.params = append(p.params, 0)
		p.paramStarted = false
		return
	}

	if b >= 0x40 && b < 0x80 {
		if p.privateMode {
			p.executePrivateMode(b, p.params)
		} else {
			p.executeCSI(b, p.params)
		}
		p.privateMode = false
		p.state = stateGround
	}
}

func (p *Parser) executeCSI(cmd byte, params []int) {
	getParam := func(idx, defaultVal int) int {
		if idx < len(params) {
			return params[idx]
		}
		return defaultVal
	}

	slog.Debug("parser: CSI", "cmd", string(cmd), "params", params)

	switch cmd {
	case 'm':
		p.handleSGR(params)
	case 'H', 'f':
		y := getParam(0, 1)
		x := getParam(1, 1)
		p.vt.CursorPosition(x, y)
	case 'A':
		p.vt.CursorUp(getParam(0, 1))
	case 'B':
		p.vt.CursorDown(getParam(0, 1))
	case 'C':
		p.vt.CursorRight(getParam(0, 1))
	case 'D':
		p.vt.CursorLeft(getParam(0, 1))
	case 'E':
		p.vt.CursorDownLine(getParam(0, 1))
	case 'F':
		p.vt.CursorUpLine(getParam(0, 1))
	case 'G':
		p.vt.CurrentScreen().moveCursor(max(getParam(0, 1)-1, 0), p.vt.CurrentScreen().CursorY)
	case 'J':
		p.vt.EraseDisplay(getParam(0, 0))
	case 'K':
		p.vt.EraseLine(getParam(0, 0))
	case 'L':
		p.vt.InsertLines(max(getParam(0, 1), 1))
	case 'M':
		p.vt.DeleteLines(max(getParam(0, 1), 1))
	case 'P':
		p.vt.DeleteChars(max(getParam(0, 1), 1))
	case '@':
		p.vt.InsertChars(max(getParam(0, 1), 1))
	case 'S':
		p.vt.ScrollUp(max(getParam(0, 1), 1))
	case 'T':
		p.vt.ScrollDown(max(getParam(0, 1), 1))
	case 'X':
		p.vt.EraseChars(max(getParam(0, 1), 1))
	case 'd':
		p.vt.CursorPosition(1, getParam(0, 1))
	case 'r':
		p.vt.SetScrollRegion(getParam(0, 1)-1, getParam(1, p.vt.Height())-1)
	case 'n':
		// DeviceStatusReport
		switch getParam(0, 0) {
		case 5:
			// Status report: OK
			p.vt.SendResponse([]byte("\x1b[0n"))
		case 6:
			// Cursor position report
			x := p.vt.CurrentScreen().CursorX + 1
			y := p.vt.CurrentScreen().CursorY + 1
			p.vt.SendResponse([]byte(fmt.Sprintf("\x1b[%d;%dR", y, x)))
		}
	case 'c':
		// DeviceAttributes — basic VT100
		p.vt.SendResponse([]byte("\x1b[?1;2c"))
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
			p.currentFg = ColorDefault
			p.currentBg = ColorDefault
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
			p.currentFg = ColorDefault
		case code >= 40 && code <= 47:
			p.currentBg = ansiBgColors[code]
		case code == 49:
			p.currentBg = ColorDefault

		case code >= 90 && code <= 97:
			p.currentFg = ansiColors[code]
		case code >= 100 && code <= 107:
			p.currentBg = ansiBgColors[code]

		case code == 38 && i < len(params):
			if params[i] == 5 && i+1 < len(params) {
				p.currentFg = ColorIndex(params[i+1])
				i += 2
			} else if params[i] == 2 && i+3 < len(params) {
				r, g, b := params[i+1], params[i+2], params[i+3]
				p.currentFg = ColorRGB(uint8(r), uint8(g), uint8(b))
				i += 4
			}

		case code == 48 && i < len(params):
			if params[i] == 5 && i+1 < len(params) {
				p.currentBg = ColorIndex(params[i+1])
				i += 2
			} else if params[i] == 2 && i+3 < len(params) {
				r, g, b := params[i+1], params[i+2], params[i+3]
				p.currentBg = ColorRGB(uint8(r), uint8(g), uint8(b))
				i += 4
			}
		}
	}
}

func (p *Parser) emitChar(ch rune) {
	p.vt.WriteChar(ch, p.currentFg, p.currentBg, p.currentAttrs)
}

func (p *Parser) executePrivateMode(cmd byte, params []int) {
	getParam := func(idx, defaultVal int) int {
		if idx < len(params) {
			return params[idx]
		}
		return defaultVal
	}

	mode := getParam(0, 0)
	on := cmd == 'h'

	slog.Info("parser: DEC mode", "mode", mode, "on", on)

	switch mode {
	case 1:
		p.vt.SetDECPrivateMode(1, on)
	case 7:
		p.vt.SetDECPrivateMode(7, on)
	case 12, 25:
		p.vt.SetDECPrivateMode(25, on)
	case 47, 1047, 1049:
		p.vt.SetDECPrivateMode(1049, on)
	case 1048:
		p.vt.SetDECPrivateMode(1048, on)
	case 2004:
		p.vt.SetDECPrivateMode(2004, on)
	}
}
