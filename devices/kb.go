package devices

// Keyboard implements simulation of Фахівець-85 keyboard.
// It's 12x6 matrix connected to the IO controller.
// 6 rows are mapped to the port B, pins 2-7 (pin 2 - row 6, pin 7 - row 1).
// First 8 columns of the matrix are mapped to the pins of port A.
// Last 4 columns are mapped to lower part of the port C.
// So the implementation of the keyboard sends values to ports A, B, and lower C, on the keystrokes.
type Keyboard struct {
	ctl IoController

	events chan keyEvent
}

func NewKeyboard(ctl IoController) *Keyboard {
	kb := &Keyboard{ctl: ctl, events: make(chan keyEvent)}
	go kb.run()
	return kb
}

func (kb *Keyboard) ShutDown() {
	close(kb.events)
}

func (kb *Keyboard) Event(code KeyCode, state KeyState) {
	kb.events <- keyEvent{code: code, state: state}
}

func (kb *Keyboard) RunSequence(seq []KeyCode) {
	for _, code := range seq {
		kb.Event(code, KeyStateDown)
		kb.Event(code, KeyStateUp)
	}
}

func (kb *Keyboard) run() {
	var matrix kbMatrix

	// Sync initial state.
	kb.syncPorts(matrix.portValues())

	// Process events.
	for event := range kb.events {
		if matrix.event(event) {
			kb.syncPorts(matrix.portValues())
		}
	}
}

func (kb *Keyboard) syncPorts(a, b, cl byte) {
	// Columns.
	kb.ctl.SendA(a)
	kb.ctl.SendCLow(cl)
	// Rows.
	kb.ctl.SendB(b)
}

type keyEvent struct {
	code  KeyCode
	state KeyState
}

type KeyState byte

const (
	KeyStateUp KeyState = iota
	KeyStateDown
)

type KeyCode byte

func MatrixKeyCode(row, col int) KeyCode {
	return KeyCode(byte(col&0x0F)<<4 | byte(row&0x0F))
}

func (kc KeyCode) matrix() (r, c int) {
	r = int(kc & 0x0F)
	c = int(kc >> 4)
	return
}

type kbMatrix struct {
	states [6][12]KeyState
}

func (kb *kbMatrix) event(event keyEvent) bool {
	row, col := event.code.matrix()
	if kb.states[row][col] != event.state {
		kb.states[row][col] = event.state
		return true
	}
	return false
}

func (kb *kbMatrix) portValues() (A, B, CLow byte) {
	rowBits, colBits := uint16(0), uint16(0)
	for row := range kb.states {
		for col := range kb.states[row] {
			if kb.states[row][col] == KeyStateDown {
				rowBits |= 1 << row
				colBits |= 1 << col
			}
		}
	}

	// Down is represented with 0.
	rowBits = ^rowBits
	colBits = ^colBits

	// First 8 columns are connected to port A.
	A = byte(colBits & 0xFF)
	// Last 4 columns are mapped to the lower C.
	CLow = byte(colBits>>8) & 0x0F
	// Rows are mapped to port B (row 1 - pin 7, row 6 - pin 2).
	B = reverseBits(byte(rowBits & 0xFF))
	return
}

func reverseBits(v byte) byte {
	v = (v & 0xAA >> 1) | (v & 0x55 << 1)
	v = (v & 0xCC >> 2) | (v & 0x33 << 2)
	v = (v & 0xF0 >> 4) | (v & 0x0F << 4)
	return v
}
