package input

import (
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

type Handler struct {
	onInput               func([]byte)
	ApplicationCursorKeys bool
	ctrlDown              bool
	ctrlConsumed          bool
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) SetOnInput(fn func([]byte)) {
	h.onInput = fn
}

func (h *Handler) SetApplicationCursorKeys(app bool) {
	slog.Info("input: application cursor keys", "enabled", app)
	h.ApplicationCursorKeys = app
}

func (h *Handler) HandleKey(event *fyne.KeyEvent) {
	key := event.Name
	slog.Debug("input: key", "key", string(key), "ctrl", h.ctrlDown)

	if key == desktop.KeyControlLeft || key == desktop.KeyControlRight {
		h.ctrlDown = true
		h.ctrlConsumed = false
		return
	}

	if h.ctrlDown {
		h.ctrlDown = false
		ctrlKey := translateCtrlKey(key)
		if ctrlKey != 0 {
			h.onInput([]byte{ctrlKey})
			h.ctrlConsumed = true
			return
		}
		h.ctrlConsumed = true
	}

	data := h.translateKey(event)
	if data != nil && h.onInput != nil {
		h.onInput(data)
	}
}

func (h *Handler) HandleRune(r rune) {
	if h.ctrlConsumed {
		h.ctrlConsumed = false
		return
	}
	h.ctrlDown = false
	slog.Debug("input: rune", "rune", string(r))
	if h.onInput != nil {
		data := []byte(string(r))
		h.onInput(data)
	}
}

func translateCtrlKey(key fyne.KeyName) byte {
	switch key {
	case fyne.KeyName("A"), fyne.KeyName("a"):
		return 0x01
	case fyne.KeyName("B"), fyne.KeyName("b"):
		return 0x02
	case fyne.KeyName("C"), fyne.KeyName("c"):
		return 0x03
	case fyne.KeyName("D"), fyne.KeyName("d"):
		return 0x04
	case fyne.KeyName("E"), fyne.KeyName("e"):
		return 0x05
	case fyne.KeyName("F"), fyne.KeyName("f"):
		return 0x06
	case fyne.KeyName("G"), fyne.KeyName("g"):
		return 0x07
	case fyne.KeyName("H"), fyne.KeyName("h"):
		return 0x08
	case fyne.KeyName("I"), fyne.KeyName("i"):
		return 0x09
	case fyne.KeyName("J"), fyne.KeyName("j"):
		return 0x0A
	case fyne.KeyName("K"), fyne.KeyName("k"):
		return 0x0B
	case fyne.KeyName("L"), fyne.KeyName("l"):
		return 0x0C
	case fyne.KeyName("M"), fyne.KeyName("m"):
		return 0x0D
	case fyne.KeyName("N"), fyne.KeyName("n"):
		return 0x0E
	case fyne.KeyName("O"), fyne.KeyName("o"):
		return 0x0F
	case fyne.KeyName("P"), fyne.KeyName("p"):
		return 0x10
	case fyne.KeyName("Q"), fyne.KeyName("q"):
		return 0x11
	case fyne.KeyName("R"), fyne.KeyName("r"):
		return 0x12
	case fyne.KeyName("S"), fyne.KeyName("s"):
		return 0x13
	case fyne.KeyName("T"), fyne.KeyName("t"):
		return 0x14
	case fyne.KeyName("U"), fyne.KeyName("u"):
		return 0x15
	case fyne.KeyName("V"), fyne.KeyName("v"):
		return 0x16
	case fyne.KeyName("W"), fyne.KeyName("w"):
		return 0x17
	case fyne.KeyName("X"), fyne.KeyName("x"):
		return 0x18
	case fyne.KeyName("Y"), fyne.KeyName("y"):
		return 0x19
	case fyne.KeyName("Z"), fyne.KeyName("z"):
		return 0x1A
	case fyne.KeyBackslash:
		return 0x1C
	case fyne.KeyLeftBracket:
		return 0x1B
	case fyne.KeyRightBracket:
		return 0x1D
	case fyne.KeyMinus:
		return 0x1F
	}
	return 0
}

func (h *Handler) translateKey(event *fyne.KeyEvent) []byte {
	key := event.Name
	app := h.ApplicationCursorKeys

	switch key {
	case fyne.KeyEscape:
		if app {
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
		if app {
			return []byte{0x1B, 'O', 'A'}
		}
		return []byte{0x1B, '[', 'A'}
	case fyne.KeyDown:
		if app {
			return []byte{0x1B, 'O', 'B'}
		}
		return []byte{0x1B, '[', 'B'}
	case fyne.KeyRight:
		if app {
			return []byte{0x1B, 'O', 'C'}
		}
		return []byte{0x1B, '[', 'C'}
	case fyne.KeyLeft:
		if app {
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
