package input

import (
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

func (h *Handler) HandleKey(event *fyne.KeyEvent) {
	key := event.Name

	if key == desktop.KeyControlLeft || key == desktop.KeyControlRight {
		h.ctrlDown = true
		return
	}

	if h.ctrlDown {
		h.ctrlDown = false
		if key == fyne.KeyName("L") || key == fyne.KeyName("l") {
			data := []byte{0x0C}
			if h.onInput != nil {
				h.onInput(data)
			}
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
	if h.onInput != nil {
		data := []byte(string(r))
		h.onInput(data)
	}
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
