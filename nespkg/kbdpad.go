package nespkg

type KbdPad struct {
	reader        *KbdReader
	strobe        bool
	currentButton GamepadButton
}

func (pad *KbdPad) regRead() uint8 {
	var val uint8 = 0
	if pad.strobe {
		if pad.reader.buttonPressed[ButtonA] {
			val |= 0x01
		}
	} else {
		if pad.reader.buttonPressed[pad.currentButton] {
			val |= 0x01
		}
		if pad.currentButton < ButtonMax {
			pad.currentButton++
		}
	}
	return val
}

func (pad *KbdPad) regWrite(val uint8) {
	pad.strobe = (val&0x01 != 0)
	pad.currentButton = ButtonA
}

func NewKbdGamepad(reader *KbdReader) *KbdPad {
	pad := new(KbdPad)
	pad.reader = reader
	Debug("Kbd pad installed")
	return pad
}

type DummyGamepad struct {
}

func (pad *DummyGamepad) regRead() uint8 {
	return 0
}

func (pad *DummyGamepad) regWrite(val uint8) {
	return
}

func NewDummyGamepad() *DummyGamepad {
	pad := new(DummyGamepad)
	return pad
}
