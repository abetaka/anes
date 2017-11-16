package nespkg

import (
	"bytes"
	"fmt"
	"os"
)

func bits(v uint, pos uint, width uint) uint {
	return (v >> pos) & ((1 << width) - 1)
}

type Nes struct {
	cpu     *Cpu
	ppu     *Ppu
	Pad     *Gamepad
	mem     *MainMemory
	rom     *NesRom
	mapper  Mapper
	display Display
}

type Display interface {
	Render(screen *[ScreenSizePixY][ScreenSizePixX]uint8)
}

var DebugEnable bool = false

func Debug(f string, a ...interface{}) {
	if DebugEnable {
		fmt.Printf(f, a...)
	}
}

func (nes *Nes) Reset() {
	nes.cpu.Reset()
}

func (nes *Nes) Regdump() {
	nes.cpu.Regdump()
}

func NewNes(d Display) *Nes {
	nes := new(Nes)
	nes.cpu = NewCpu(nes)
	nes.ppu = NewPpu(nes)
	nes.mem = NewMainMemory(nes)
	nes.cpu.mem = nes.mem
	nes.display = d
	nes.Pad = NewGamepad()
	Debug("NewNes: nes=%p\n", nes)
	return nes
}

const maxRomImageSize = 0x100000

type NesRom struct {
	filename            string
	romImage            []uint8
	prgRom              []uint8
	chrRom              []uint8
	signature           [4]uint8
	prgRomSizeIn16KB    uint
	chrRomSizeIn8KB     uint
	prgRamSizeIn8KB     uint
	verticalMirror      bool
	batteryBackedPrgRam bool
	trainerPresent      bool
	fourScreenVram      bool
	mapperNum           int
	vsUnisystem         bool
	playChoice10        bool
	nes2format          bool
	tvSystem            uint
}

func NewNesRom(filename string, romImage []uint8) (*NesRom, error) {
	rom := new(NesRom)
	rom.filename = filename
	if bytes.Equal(romImage[0:4], []uint8{'N', 'E', 'S', 0x1a}) == false {
		return nil, fmt.Errorf("Invalid NES ROM File signature")
	}
	rom.romImage = romImage
	rom.prgRomSizeIn16KB = uint(romImage[4])
	rom.chrRomSizeIn8KB = uint(romImage[5])
	rom.prgRamSizeIn8KB = uint(romImage[8])
	rom.verticalMirror = romImage[6]&0x01 != 0
	rom.batteryBackedPrgRam = romImage[6]&0x02 != 0
	rom.trainerPresent = romImage[6]&0x04 != 0
	rom.fourScreenVram = romImage[6]&0x08 != 0
	rom.mapperNum = int(((romImage[6] & 0xf0) >> 4) | (romImage[7] & 0xf0))
	rom.vsUnisystem = romImage[7]&0x01 != 0
	rom.playChoice10 = romImage[7]&0x02 != 0
	rom.nes2format = (romImage[7]&0x0c)>>2 == 2
	rom.tvSystem = uint(romImage[9] & 0x03)

	prgstart := uint(16)
	if rom.trainerPresent {
		prgstart += 512
	}
	prgend := prgstart + 16*1024*rom.prgRomSizeIn16KB
	rom.prgRom = romImage[prgstart:prgend]

	chrstart := prgend
	chrend := chrstart + 8*1024*rom.chrRomSizeIn8KB
	rom.chrRom = romImage[chrstart:chrend]

	return rom, nil
}

func (rom *NesRom) PrintRomData() {
	Debug("filename=%s\n", rom.filename)
	Debug("prgRom=%x\n", rom.prgRom[0:16])
	Debug("prgRom=%x\n", rom.prgRom[len(rom.prgRom)-16:len(rom.prgRom)])
	Debug("chrRom=%x\n", rom.chrRom[0:16])
	Debug("prgRomSizeIn16KB=%d\n", rom.prgRomSizeIn16KB)
	Debug("chrRomSizeIn8KB=%d\n", rom.chrRomSizeIn8KB)
	Debug("prgRamSizeIn8KB=%d\n", rom.prgRamSizeIn8KB)
	Debug("verticalMirror=%t\n", rom.verticalMirror)
	Debug("batteryBackedPrgRam=%t\n", rom.batteryBackedPrgRam)
	Debug("trainerPresent=%t\n", rom.trainerPresent)
	Debug("fourScreenVram=%t\n", rom.fourScreenVram)
	Debug("mapperNum=%d\n", rom.mapperNum)
	Debug("vsUnisystem=%t\n", rom.vsUnisystem)
	Debug("playChoice10=%t\n", rom.playChoice10)
	Debug("nes2format=%t\n", rom.nes2format)
	Debug("tvSystem=%d\n", rom.tvSystem)
}

func (nes *Nes) LoadRom(filename string) error {
	f, err := os.Open(filename)
	romImage := make([]byte, maxRomImageSize)
	_, err2 := f.Read(romImage)
	if err2 != nil {
		return fmt.Errorf("ROM file open error")
	}
	defer f.Close()

	Debug("ROM file opened\n")
	rom, err := NewNesRom(filename, romImage)
	if err != nil {
		return fmt.Errorf("Not valid iNES file")
	}
	Debug("ROM header analyzed\n")
	rom.PrintRomData()
	nes.rom = rom
	nes.mapper = MakeMapper(nes, rom.mapperNum)
	nes.mapper.Init()
	Debug("calling PostRomLoadSetup\n")
	nes.ppu.PostRomLoadSetup()
	Debug("returning from LoadRom\n")
	return nil
}

func (nes *Nes) Run() {
	nes.Reset()
	//running := true
	//for i := 0; i < 32; i++ {
	for true {
		cycle := nes.cpu.executeInst()
		nes.ppu.giveCpuClockDelta(cycle)
	}
}
