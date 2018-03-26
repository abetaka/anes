package nespkg

import (
	"github.com/simulatedsimian/joystick"
)

type UsbGamePad struct {
	js            joystick.Joystick
	calibX        int
	calibY        int
	strobe        bool
	buttonPressed [ButtonMax]bool
	currentButton GamepadButton
}

func (pad *UsbGamePad) regRead() uint8 {
	var val uint8 = 0
	if pad.strobe {
		if pad.buttonPressed[ButtonA] {
			val |= 0x01
		}
	} else {
		if pad.buttonPressed[pad.currentButton] {
			val |= 0x01
		}
		if pad.currentButton < ButtonMax-1 {
			pad.currentButton++
		}
	}
	return val
}

func (pad *UsbGamePad) scan() {
	for i := range pad.buttonPressed {
		pad.buttonPressed[i] = false
	}

	state, err := pad.js.Read()
	if err == nil {
		//Debug("X=%d Y=%d Buttons=%X\n", state.AxisData[0], state.AxisData[1], state.Buttons)
		x := state.AxisData[0] - pad.calibX
		if x < 0 {
			pad.buttonPressed[ButtonLeft] = true
		} else if x > 0 {
			pad.buttonPressed[ButtonRight] = true
		}

		y := state.AxisData[1] - pad.calibY
		if y < 0 {
			pad.buttonPressed[ButtonUp] = true
		} else if y > 0 {
			pad.buttonPressed[ButtonDown] = true
		}

		if state.Buttons&0x08 != 0 {
			pad.buttonPressed[ButtonA] = true
		}
		if state.Buttons&0x04 != 0 {
			pad.buttonPressed[ButtonB] = true
		}
		if state.Buttons&0x80 != 0 {
			pad.buttonPressed[ButtonStart] = true
		}
		if state.Buttons&0x40 != 0 {
			pad.buttonPressed[ButtonSelect] = true
		}
	}
}

func (pad *UsbGamePad) regWrite(val uint8) {
	pad.strobe = (val&0x01 != 0)
	if !pad.strobe {
		pad.scan()
	}
	pad.currentButton = ButtonA
}

func NewUsbGamepad(jsid int) *UsbGamePad {
	js, err := joystick.Open(jsid)
	if err != nil {
		return nil
	}

	pad := new(UsbGamePad)
	pad.js = js

	pad.calibX = -256
	pad.calibY = -256
	Debug("calibX=%d calibY=%d\n", pad.calibX, pad.calibY)

	pad.strobe = false
	for i := range pad.buttonPressed {
		pad.buttonPressed[i] = false
	}
	pad.currentButton = ButtonA
	return pad
}
