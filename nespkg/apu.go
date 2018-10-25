package nespkg

import "github.com/hajimehoshi/oto"
import "fmt"

//import "log"

const samplingRate = 48000
const bufferSize = samplingRate / 10

const apuRegAddressPulse1A = 0x4000
const apuRegAddressPulse1B = 0x4001
const apuRegAddressPulse1C = 0x4002
const apuRegAddressPulse1D = 0x4003

type Apu struct {
	nes    *Nes
	clock  int
	player *oto.Player
	pulse1 *PulseGen
	//Pulse2   PulseGen
	//Triangle TriangleGen
	//Noise    NoiseGen
	//Dmc      DmcGen
}

type PulseGen struct {
	duty              int
	envelopeLoop      int
	lengthCounterHalt bool
	constantVolume    bool
	volumeEnvelope    int
	sweepEnable       bool
	sweepPeriod       int
	sweepNegate       bool
	sweepShift        int
	timer             int
	lengthCounterLoad int
	lengthCounter     int
	sequencerMode     int
	interruptInhibit  bool
	interruptFlag     bool
}

const SEQUENCER_MODE_0 = 0
const SEQUENCER_MODE_1 = 1

func NewPulseGen() *PulseGen {
	pulse := new(PulseGen)
	pulse.duty = 0
	pulse.envelopeLoop = 0
	pulse.lengthCounterHalt = false
	pulse.constantVolume = false
	pulse.volumeEnvelope = 0
	pulse.sweepEnable = false
	pulse.sweepPeriod = 0
	pulse.sweepNegate = false
	pulse.sweepShift = 0
	pulse.timer = 0
	pulse.lengthCounterLoad = 0
	pulse.lengthCounter = 0
	pulse.sequencerMode = 0
	pulse.interruptInhibit = false
	pulse.interruptFlag = false
	return pulse
}

func NewApu(nes *Nes) *Apu {
	apu := new(Apu)
	apu.nes = nes
	apu.pulse1 = NewPulseGen()
	apu.clock = 0

	player, err := oto.NewPlayer(samplingRate, 1, 1, bufferSize)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Fail to create new player")
		return nil
	}
	apu.player = player

	return apu
}

func (pulse *PulseGen) readApuPulseReg(offset uint16) uint8 {
	v := uint8(0)
	switch offset {
	case 0:
		v |= uint8(pulse.duty) << 6
		if pulse.lengthCounterHalt {
			v |= 0x20
		}
		if pulse.constantVolume {
			v |= 0x10
		}
		v |= uint8(pulse.volumeEnvelope)
	case 1:
		if pulse.sweepEnable {
			v |= 0x80
		}
		v |= uint8(pulse.sweepPeriod) << 4
		if pulse.sweepNegate {
			v |= 0x08
		}
		v |= uint8(pulse.sweepShift)
	case 2:
		v |= uint8(pulse.timer)
	case 3:
		v |= uint8(pulse.lengthCounterLoad) << 3
		v |= uint8(pulse.timer>>8) & 0x07
	}
	return v
}

func (pulse *PulseGen) writeApuPulseReg(offset uint16, v uint8) {
	switch offset {
	case 0:
		pulse.duty = int((v & 0x0c0) >> 6)
		pulse.lengthCounterHalt = v&0x020 != 0
		pulse.constantVolume = v&0x010 != 0
		pulse.volumeEnvelope = int(v & 0x00f)
	case 1:
		pulse.sweepEnable = v&0x80 != 0
		pulse.sweepPeriod = int((v & 0x070) >> 4)
		pulse.sweepNegate = v&0x08 != 0
		pulse.sweepShift = int(v & 0x007)
	case 2:
		pulse.timer |= int(v)
	case 3:
		pulse.lengthCounterLoad |= int((v & 0x0F8) >> 3)
		pulse.timer |= int((v & 0x007) << 8)
	}
}

func (apu *Apu) WriteReg(address uint16, v uint8) {
	if address >= apuRegAddressPulse1A && address <= apuRegAddressPulse1D {
		apu.pulse1.writeApuPulseReg(address-apuRegAddressPulse1A, v)
	}
}

func (apu *Apu) ReadReg(address uint16) uint8 {
	if address >= apuRegAddressPulse1A && address <= apuRegAddressPulse1D {
		return apu.pulse1.readApuPulseReg(address - apuRegAddressPulse1A)
	}
	return 0
}

const CpuHz = 1789773

func timerToHz(t int) int {
	return CpuHz / (16 * (t + 1))
}

func (apu *Apu) giveFrameTiming() {
	apu.clock += bufferSize
	data := make([]byte, bufferSize)
	hz := timerToHz(apu.pulse1.timer)
	rectangleWave(data, apu.pulse1.duty, apu.clock, hz, apu.pulse1.volumeEnvelope)
	//log.Println(data)
	apu.player.Write(data)
}

func rectangleWave(data []byte, duty int, clock int, hz int, level int) {
	if hz <= 0 {
		for i := range data {
			data[i] = 0
		}
		return
	}
	clocksPerPeriod := samplingRate / hz
	Debug("hz=%d clocksPerPeriod=%d\n", hz, clocksPerPeriod)
	if clocksPerPeriod == 0 {
		for i := range data {
			data[i] = 0
		}
		return
	}

	upclocks := 0
	reverse := false
	switch duty {
	case 0:
		upclocks = clocksPerPeriod / 8
	case 1:
		upclocks = clocksPerPeriod / 4
	case 2:
		upclocks = clocksPerPeriod / 2
	case 3:
		upclocks = clocksPerPeriod / 4
		reverse = true
	}

	up := 0
	down := level
	if reverse {
		up = level
		down = 0
	}

	for i := range data {
		if (clock+i)%clocksPerPeriod < upclocks {
			data[i] = byte(up)
		} else {
			data[i] = byte(down)
		}
	}
}
