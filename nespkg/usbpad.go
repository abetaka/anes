package nespkg

import (
	"github.com/simulatedsimian/joystick"
	"time"
)

type UsbGamePad struct {
	js            joystick.Joystick
	centerX       int
	centerY       int
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
		x := state.AxisData[0] - pad.centerX
		if x < 0 {
			pad.buttonPressed[ButtonLeft] = true
		} else if x > 0 {
			pad.buttonPressed[ButtonRight] = true
		}

		y := state.AxisData[1] - pad.centerY
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
	time.Sleep(time.Second * 1)
	state, err := pad.js.Read()
	if err == nil {
		pad.centerX = state.AxisData[0]
		pad.centerY = state.AxisData[1]
	} else {
		pad.centerX = -256
		pad.centerY = -256
	}

	Debug("centerX=%d centerY=%d\n", pad.centerX, pad.centerY)

	pad.strobe = false
	for i := range pad.buttonPressed {
		pad.buttonPressed[i] = false
	}
	pad.currentButton = ButtonA
	return pad
}
