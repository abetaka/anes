package nespkg

const ScreenSizePixX = 256
const ScreenSizePixY = 240
const tileSizePixX = 8
const tileSizePixY = 8
const screenSizeTileX = ScreenSizePixX / tileSizePixX
const screenSizeTileY = ScreenSizePixY / tileSizePixY
const vramSize = 0x4000
const vramPageShift = 8
const vramPageSize = 1 << vramPageShift
const vramPages = vramSize / vramPageSize

type Palette struct {
	bgPalette     [4][]uint8
	spritePalette [4][]uint8
}

type Ppu struct {
	ppuctrl         uint8
	ppumask         uint8
	ppustatus       uint8
	oamaddr         uint8
	oamdata         uint8
	oamdma          uint8
	ppuscrollx      uint8
	ppuscrolly      uint8
	ppuscrollynew   uint8
	ppuscrollw      bool
	ppuaddr         uint16
	ppudata         uint8
	ppuaddrw        bool
	vram            [vramSize]uint8
	lvram           [vramPages][]uint8
	patterntable    [2][]uint8
	nametable       [4][]uint8
	attributetable  [4][]uint8
	oam             [4 * 64]uint8
	palette         Palette
	bgPalette       [4][]uint8
	spritePalette   [4][]uint8
	screen          [ScreenSizePixY][ScreenSizePixX]uint8
	oddframe        bool
	currentScanline uint
	clock           uint
	nes             *Nes
}

func (ppu *Ppu) writeMmapReg(address uint16, v uint8) {
	if address == 0x4014 {
		ppu.writeOamdma(v)
	} else {
		switch address & 0x07 {
		case 0:
			ppu.writePpuctrl(v)
		case 1:
			ppu.writePpumask(v)
		case 3:
			ppu.writeOamaddr(v)
		case 4:
			ppu.writeOamdata(v)
		case 5:
			ppu.writePpuscroll(v)
		case 6:
			ppu.writePpuaddr(v)
		case 7:
			ppu.writePpudata(v)
		}
	}
}

func (ppu *Ppu) readMmapReg(address uint16) uint8 {
	switch address & 0x07 {
	case 2:
		return ppu.readPpustatus()
	case 4:
		return ppu.readOamdata()
	case 7:
		return ppu.readPpudata()
	}
	return 0
}

func (ppu *Ppu) writePpuctrl(v uint8) {
	ppu.ppuctrl = v
}

func (ppu *Ppu) baseNametableAddress() uint {
	return bits(uint(ppu.ppuctrl), 0, 3)
}

func (ppu *Ppu) vramIncMode() uint {
	return bits(uint(ppu.ppuctrl), 2, 1)
}

func (ppu *Ppu) spritePatternTableAddress8x8() uint {
	return bits(uint(ppu.ppuctrl), 3, 1)
}

func (ppu *Ppu) bgPatternTableAddress() uint {
	return bits(uint(ppu.ppuctrl), 4, 1)
}

func (ppu *Ppu) spriteSize() uint {
	return bits(uint(ppu.ppuctrl), 5, 1)
}

func (ppu *Ppu) ppuMasterSlave() uint {
	return bits(uint(ppu.ppuctrl), 6, 1)
}

func (ppu *Ppu) vblankNmi() bool {
	return bits(uint(ppu.ppuctrl), 7, 1) == 1
}

func (ppu *Ppu) writePpumask(v uint8) {
	ppu.ppumask = v
}

func (ppu *Ppu) greyscale() uint {
	return bits(uint(ppu.ppumask), 0, 1)
}

func (ppu *Ppu) showBgInLeftmost() uint {
	return bits(uint(ppu.ppumask), 1, 1)
}

func (ppu *Ppu) showSpriteInLeftmost() uint {
	return bits(uint(ppu.ppumask), 2, 1)
}

func (ppu *Ppu) showBg() uint {
	return bits(uint(ppu.ppumask), 3, 1)
}

func (ppu *Ppu) showSprite() uint {
	return bits(uint(ppu.ppumask), 4, 1)
}

func (ppu *Ppu) emphasizeRed() uint {
	return bits(uint(ppu.ppumask), 5, 1)
}

func (ppu *Ppu) emphasizeGreen() uint {
	return bits(uint(ppu.ppumask), 6, 1)
}

func (ppu *Ppu) emphasizeBlue() uint {
	return bits(uint(ppu.ppumask), 7, 1)
}

func (ppu *Ppu) readPpustatus() uint8 {
	return ppu.ppustatus
}

func (ppu *Ppu) writePpuscroll(v uint8) {
	if ppu.ppuscrollw {
		ppu.ppuscrollynew = v
		ppu.ppuscrollw = false
	} else {
		ppu.ppuscrollx = v
		ppu.ppuscrollw = true
	}
}

func (ppu *Ppu) updatePpuscrolly() {
	ppu.ppuscrolly = ppu.ppuscrollynew
}

func (ppu *Ppu) writePpuaddr(v uint8) {
	if ppu.ppuaddrw {
		ppu.ppuaddr = ppu.ppuaddr&0xff00 | uint16(v)
		ppu.ppuaddrw = false
	} else {
		ppu.ppuaddr = ppu.ppuaddr&0x00ff | uint16(v)<<8
		ppu.ppuaddrw = true
	}
}

func (ppu *Ppu) incPpuaddr() {
	if ppu.vramIncMode() == 0 {
		ppu.ppuaddr++
	} else {
		ppu.ppuaddr += 32
	}
}

func (ppu *Ppu) writePpudata(v uint8) {
	Debug("ppu.ppuaddr=%04X\n", ppu.ppuaddr)
	ppu.lvram[vramPage(ppu.ppuaddr)][vramOffest(ppu.ppuaddr)] = v
	ppu.incPpuaddr()
}

func (ppu *Ppu) readPpudata() uint8 {
	var v uint8
	if ppu.ppuaddr < 0x3f00 {
		v = ppu.ppudata
		ppu.ppudata = ppu.lvram[vramPage(ppu.ppuaddr)][vramOffest(ppu.ppuaddr)]
	} else {
		ppu.ppudata = ppu.lvram[vramPage(ppu.ppuaddr)][vramOffest(ppu.ppuaddr)]
		v = ppu.ppudata
	}
	ppu.incPpuaddr()
	return v
}

func (ppu *Ppu) writeOamaddr(v uint8) {
	ppu.oamaddr = v
}

func (ppu *Ppu) writeOamdata(v uint8) {
	ppu.oam[ppu.oamaddr] = v
	ppu.oamaddr++
}

func (ppu *Ppu) readOamdata() uint8 {
	return ppu.oam[ppu.oamaddr]
}

func (ppu *Ppu) writeOamdma(hi uint8) {
	const oamdmalen = 0x100
	cpuaddr := uint16(hi << 8)
	for i := 0; i < oamdmalen; i++ {
		ppu.oam[int(hi)+i] = ppu.nes.mem.Read8(cpuaddr)
		cpuaddr++
	}
}

func (ppu *Ppu) RenderScreen() {
	for row := range ppu.screen {
		ppu.renderScanline(uint(row))
	}
}

func (ppu *Ppu) renderScanline(row uint) {
	for col := 0; col < ScreenSizePixX; col++ {
		ppu.renderPixel(uint(col), row)
	}
}

func getIndexInNametable(x uint, y uint) uint {
	return screenSizeTileX*((y%ScreenSizePixY)/8) + ((x % ScreenSizePixX) / 8)
}

func getPaletteIndex(attributetable []uint8, x uint, y uint) uint8 {
	const attrMetatileSize = 32
	const attrTileSize = 16
	const attrMetatilesInRow = 8

	u := x % ScreenSizePixX
	v := y % ScreenSizePixY
	index := u/attrMetatileSize + (u/attrMetatileSize)*attrMetatilesInRow
	attr := attributetable[index]
	s := (u / attrTileSize) % 2
	t := (v / attrTileSize) % 2
	return (attr >> ((s + t*2) * 2) & 0x03)
}

func getNametableIndex(x uint, y uint) uint {
	return (x/ScreenSizePixX)%2 + (y/ScreenSizePixY)%2
}

func (ppu *Ppu) renderPixel(col uint, row uint) {
	const patternEntryBytes = 16
	const tileSize = 8
	const hiOffset = 8

	x := col + uint(ppu.ppuscrolly)
	y := row + uint(ppu.ppuscrolly)
	nametableIndex := getNametableIndex(x, y)
	nametable := ppu.nametable[nametableIndex]

	indexInNametable := getIndexInNametable(x, y)
	patternIndex := uint(nametable[indexInNametable])*patternEntryBytes + y%tileSize

	patterntable := ppu.patterntable[0]
	lo := patterntable[patternIndex]
	hi := patterntable[patternIndex+hiOffset]

	attributetable := ppu.attributetable[nametableIndex]
	paletteIndex := getPaletteIndex(attributetable, x, y)

	pix := bits(uint(lo), x%tileSize, 1) | (bits(uint(hi), x%tileSize, 1) << 1)
	if pix == 0 {
		ppu.screen[row][col] = ppu.bgPalette[0][0]
	} else {
		ppu.screen[row][col] = ppu.bgPalette[paletteIndex][pix]
	}
}

func vramPage(a uint16) uint {
	return uint(a) >> vramPageShift
}

func vramOffest(a uint16) uint {
	// 0x3F20-0x3FFF: mirrors of 0x3F00-0x3F1F
	if a >= 0x3F20 && a < 0x4000 {
		a &= 0xFF1F
	}

	return uint(a) & (vramPageSize - 1)
}

func (ppu *Ppu) vramWrite8(address uint16, v uint8) {
	ppu.lvram[vramPage(address)][vramOffest(address)] = v
}

func (ppu *Ppu) vramRead8(address uint16) uint8 {
	return ppu.lvram[vramPage(address)][vramOffest(address)]
}

func (ppu *Ppu) reset() {
	ppu.currentScanline = 0
	ppu.clock = 0
}

func toPpuClockDelta(cpuclockDelta uint) uint {
	const ppuCpuClockRatio = 3
	return cpuclockDelta * ppuCpuClockRatio
}

const firstVisibleScanline = 0
const lastVisibleScanline = 239
const firstVBlankScanline = 241
const lastScanline = 261

func scanlineToClock(row uint) uint {
	return (row + 1) * 341
}

func (ppu *Ppu) giveCpuClockDelta(cpuclockDelta uint) {
	ppu.clock += toPpuClockDelta(cpuclockDelta)
	for row := ppu.currentScanline; ppu.clock >= scanlineToClock(row); row++ {
		if row >= firstVisibleScanline && row <= lastVisibleScanline {
			ppu.renderScanline(row)
		} else if row == firstVBlankScanline && ppu.vblankNmi() {
			ppu.nes.cpu.setNmi()
		}
		ppu.currentScanline = row

		if row == lastVisibleScanline {
			Debug("lastVisibleScanline\n")
			ppu.nes.display.Render(&ppu.screen)
		} else if row == lastScanline {
			ppu.currentScanline = 0
			ppu.clock = 0
		}
	}
}

func NewPpu(nes *Nes) *Ppu {
	ppu := new(Ppu)
	ppu.nes = nes

	//
	// Initialize VRAM
	//
	for i := range ppu.lvram {
		ppu.lvram[i] = ppu.vram[vramPageSize*i : vramPageSize*(i+1)]
	}
	Debug("VRAM initialized\n")

	return ppu
}

func (ppu *Ppu) PostRomLoadSetup() {
	rom := ppu.nes.rom
	//
	// Nametable mirror
	//
	if rom.fourScreenVram == false {
		if rom.verticalMirror {
			for a := uint16(0x2000); a < 0x2800; a += vramPageSize {
				ppu.lvram[vramPage(a+0x800)] = ppu.lvram[vramPage(a)]
				ppu.lvram[vramPage(a+0xc00)] = ppu.lvram[vramPage(a+0x400)]
			}
		} else {
			for a := uint16(0x2000); a < 0x2800; a += vramPageSize {
				ppu.lvram[vramPage(a+0x400)] = ppu.lvram[vramPage(a)]
				ppu.lvram[vramPage(a+0xc00)] = ppu.lvram[vramPage(a+0x800)]
			}
		}
	}

	Debug("Setting VRAM mirror\n")
	for a := uint16(0x3000); a < 0x3f00; a += vramPageSize {
		ppu.lvram[vramPage(a)] = ppu.lvram[vramPage(a-0x1000)]
	}

	//
	// Map Pattern Tables
	//
	Debug("Map pattern tables\n")
	ppu.patterntable[0] = ppu.vram[0:0x1000]
	ppu.patterntable[1] = ppu.vram[0x1000:0x2000]

	//
	// Map Name Tables and Attribute Tables
	//
	Debug("Map name tables\n")
	if rom.fourScreenVram {
		ppu.nametable[0] = ppu.vram[0x2000:0x23c0]
		ppu.attributetable[0] = ppu.vram[0x23c0:0x2400]
		ppu.nametable[1] = ppu.vram[0x2400:0x27c0]
		ppu.attributetable[1] = ppu.vram[0x27c0:0x2800]
		ppu.nametable[2] = ppu.vram[0x2800:0x2bc0]
		ppu.attributetable[2] = ppu.vram[0x2bc0:0x2c00]
		ppu.nametable[3] = ppu.vram[0x2c00:0x2fc0]
		ppu.attributetable[3] = ppu.vram[0x2fc0:0x3000]
	} else if rom.verticalMirror {
		ppu.nametable[0] = ppu.vram[0x2000:0x23c0]
		ppu.attributetable[0] = ppu.vram[0x23c0:0x2400]
		ppu.nametable[1] = ppu.vram[0x2400:0x27c0]
		ppu.attributetable[1] = ppu.vram[0x27c0:0x2800]
		ppu.nametable[2] = ppu.nametable[0]
		ppu.attributetable[2] = ppu.attributetable[0]
		ppu.nametable[3] = ppu.nametable[1]
		ppu.attributetable[3] = ppu.attributetable[1]
	} else {
		ppu.nametable[0] = ppu.vram[0x2000:0x23c0]
		ppu.attributetable[0] = ppu.vram[0x23c0:0x2400]
		ppu.nametable[1] = ppu.nametable[0]
		ppu.attributetable[1] = ppu.nametable[0]
		ppu.nametable[2] = ppu.vram[0x2800:0x2bc0]
		ppu.attributetable[2] = ppu.vram[0x2bc0:0x2c00]
		ppu.nametable[3] = ppu.nametable[2]
		ppu.attributetable[3] = ppu.attributetable[2]
	}

	//
	// Initialize Palettes
	//
	Debug("Initialize palettes\n")
	ppu.bgPalette[0] = ppu.vram[0x3f00:0x3f04]
	ppu.bgPalette[1] = ppu.vram[0x3f04:0x3f08]
	ppu.bgPalette[2] = ppu.vram[0x3f08:0x3f0c]
	ppu.bgPalette[3] = ppu.vram[0x3f0c:0x3f10]
	ppu.spritePalette[0] = ppu.vram[0x3f11:0x3f14]
	ppu.spritePalette[1] = ppu.vram[0x3f15:0x3f18]
	ppu.spritePalette[2] = ppu.vram[0x3f19:0x3f1c]
	ppu.spritePalette[3] = ppu.vram[0x3f1d:0x3f20]

}
