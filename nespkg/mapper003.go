package nespkg

type Mapper003 struct {
	MapperBase
}

func (mapper *Mapper003) regWrite8(address uint16, val uint8) {
	if address >= 0x8000 && address <= 0xffff {
		bank := int(val & 0x03)
		Debug("Mapper003 bank=%d\n", bank)
		nes := mapper.nes
		nes.ppu.mapExtMem(0, nes.rom.chrRom[0x2000*bank:0x2000*(bank+1)], 0x2000)
	}
}

func NewMapper003(nes *Nes) Mapper {
	mapper := new(Mapper003)
	mapper.mapperNum = 3
	mapper.nes = nes
	return mapper
}
