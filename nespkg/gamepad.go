package nespkg

const (
	ButtonA = iota
	ButtonB
	ButtonSelect
	ButtonStart
	ButtonUp
	ButtonDown
	ButtonLeft
	ButtonRight
	ButtonEnd
)

type GamepadButton int

type Gamepad struct {
	strobe        bool
	buttonPressed [9]bool
	currentButton GamepadButton
}

func (pad *Gamepad) regRead(address uint16) uint8 {
	if address != 0x4016 {
		return 0
	}
	var val uint8 = 0
	if pad.strobe {
		if pad.buttonPressed[ButtonA] {
			val |= 0x01
		}
	} else {
		if pad.buttonPressed[pad.currentButton] {
			val |= 0x01
		}
		if pad.currentButton < ButtonEnd {
			pad.currentButton++
		}
	}
	return val
}

func (pad *Gamepad) regWrite(address uint16, val uint8) {
	if address != 0x4016 {
		return
	}
	pad.strobe = (val&0x01 != 0)
	pad.currentButton = ButtonA
}

func (pad *Gamepad) SetButtonState(button GamepadButton, pressed bool) {
	pad.buttonPressed[button] = pressed
}

func NewGamepad() *Gamepad {
	pad := new(Gamepad)
	pad.strobe = false
	for i := range pad.buttonPressed {
		pad.buttonPressed[i] = false
	}
	pad.currentButton = ButtonA
	return pad
}
