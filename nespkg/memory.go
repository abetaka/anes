package nespkg

const ppuRegMask = 0xF007
const mmPageShift = 11
const mmPageSize = 1 << mmPageShift
const mmMemorySpaceSize = 0x10000

type MainMemory struct {
	mem [mmMemorySpaceSize / mmPageSize][]uint8
	nes *Nes
}

func isPpuRegAddress(address uint16) bool {
	if address == 0x4014 {
		return true
	}
	ppuAddress := address & 0xE007
	return (ppuAddress >= 0x2000 && ppuAddress <= 0x2007)
}

func isGamepadAddress(address uint16) bool {
	if address == 0x4016 {
		return true
	}
	return false
}

func page(address uint16) uint {
	return (uint(address) >> mmPageShift)
}

func offset(address uint16) uint {
	return (uint(address) & (mmPageSize - 1))
}

func (m *MainMemory) Read8(address uint16) uint8 {
	if isPpuRegAddress(address) {
		return m.nes.ppu.readMmapReg(address)
	} else if isGamepadAddress(address) {
		return m.nes.Pad.regRead(address)
	}
	v := m.mem[page(address)][offset(address)]
	if address < 0x1000 {
		Debug("  Rd8: %04X -> %02X\n", address, v)
	}
	return v
}

func (m *MainMemory) isRam(address uint16) bool {
	if address >= 0 && address < 0x2000 {
		return true
	}
	return false
}

func (m *MainMemory) Write8(address uint16, val uint8) {
	if address < 0x1000 {
		Debug("  Wt8: %04X <- %02X\n", address, val)
	}
	if m.isRam(address) {
		m.mem[page(address)][offset(address)] = val
	} else if isPpuRegAddress(address) {
		m.nes.ppu.writeMmapReg(address, val)
	} else if isGamepadAddress(address) {
		m.nes.Pad.regWrite(address, val)
	} else if address >= 0x8000 && address <= 0xffff {
		m.nes.mapper.regWrite8(address, val)
	}
}

func (m *MainMemory) Read16(address uint16) uint16 {
	return uint16(m.Read8(address)) | uint16(m.Read8(address+1))<<8
}

func (m *MainMemory) Write16(address uint16, v uint16) {
	m.Write8(address, uint8(v&0x0ff))
	m.Write8(address+1, uint8(v>>8))
}

func (m *MainMemory) setNrom128Mirror() {
	for a := 0xC000; a < 0x10000; a += mmPageSize {
		m.mem[page(uint16(a))] = m.mem[page(uint16(a-0x4000))]
	}
}

func NewMainMemory(nes *Nes) *MainMemory {
	m := new(MainMemory)
	m.mem[page(0x0000)] = make([]uint8, mmPageSize)
	m.mem[page(0x0800)] = m.mem[0]
	m.mem[page(0x1000)] = m.mem[0]
	m.mem[page(0x1800)] = m.mem[0]
	m.mem[page(0x2000)] = nil
	m.mem[page(0x2800)] = nil
	m.mem[page(0x3000)] = nil
	m.mem[page(0x3800)] = nil
	for addr := 0x4000; addr < mmMemorySpaceSize; addr += mmPageSize {
		m.mem[page(uint16(addr))] = make([]uint8, mmPageSize)
	}
	m.nes = nes
	return m
}

func (m *MainMemory) mapExtMem(address uint16, extmem []uint8, bytes int) {
	for i := uint16(0); i < uint16(bytes); i += mmPageSize {
		m.mem[page(address+i)] = extmem[i : i+mmPageSize]
	}
}
