package nespkg

type MapperBase struct {
	nes       *Nes
	mapperNum int
}

func (mapper *MapperBase) regWrite8(address uint16, val uint8) {
	/*
		if address >= 0x8000 && address <= 0xffff {
			bank := int(val & 0x03)
			nes := mapper.nes
			nes.ppu.mapExtMem(0, nes.rom.chrRom[0x2000*bank:0x2000*(bank+1)], 0x2000)
		}
	*/
	return
}

func (mapper *MapperBase) Init() {
	Debug("MapperBase Init()\n")
	nes := mapper.nes
	nes.mem.mapExtMem(0x8000, nes.rom.prgRom, len(nes.rom.prgRom))
	if nes.rom.prgRomSizeIn16KB == 1 {
		nes.mem.setNrom128Mirror()
	}

	nes.ppu.mapExtMem(0, nes.rom.chrRom, len(nes.rom.chrRom))
}

func NewMapperBase(nes *Nes) Mapper {
	mapper := new(MapperBase)
	mapper.mapperNum = 0
	mapper.nes = nes
	return mapper
}
