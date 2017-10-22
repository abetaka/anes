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

func page(address uint16) uint {
	return (uint(address) >> mmPageShift)
}

func offset(address uint16) uint {
	return (uint(address) & (mmPageSize - 1))
}

func (m *MainMemory) Read8(address uint16) uint8 {
	if isPpuRegAddress(address) {
		return m.nes.ppu.readMmapReg(address)
	}
	return m.mem[page(address)][offset(address)]
}

func (m *MainMemory) Write8(address uint16, value uint8) {
	if isPpuRegAddress(address) {
		m.nes.ppu.writeMmapReg(address, value)
	} else {
		m.mem[page(address)][offset(address)] = value
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
