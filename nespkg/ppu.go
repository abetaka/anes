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
const nullSpIndex = 255

type Ppustatus struct {
	data uint8
}

const PPUSTATUS_V = uint8(0x80)
const PPUSTATUS_S = uint8(0x40)

type Ppu struct {
	ppuctrl         uint8
	ppumask         uint8
	ppustatus       uint8
	oamaddr         uint8
	oamdma          uint8
	ppuscrollx      uint8
	ppuscrollxnew   uint8
	ppuscrolly      uint8
	ppuscrollynew   uint8
	ppuscrollw      bool
	ppuaddr         uint16
	ppudata         uint8
	ppuaddrw        bool
	vram            [vramSize]uint8
	lvram           [vramPages][]uint8
	nametable       [4][]uint8
	attributetable  [4][]uint8
	oam             [4 * 64]uint8
	bgPalette       [4][]uint8
	spPalette       [4][]uint8
	screen          [ScreenSizePixY][ScreenSizePixX]uint8
	oamScreen       [ScreenSizePixY][ScreenSizePixX]uint8
	oamBehindBg     [ScreenSizePixY][ScreenSizePixX]bool
	oamMap          [ScreenSizePixY][ScreenSizePixX]uint8
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
	//Debug("writePpuctrl: %02Xh\n", v)
	ppu.ppuctrl = v
}

func (ppu *Ppu) baseNametableAddress() uint {
	return bits(uint(ppu.ppuctrl), 0, 3)
}

func (ppu *Ppu) vramIncMode() uint {
	return bits(uint(ppu.ppuctrl), 2, 1)
}

func (ppu *Ppu) sprite8x8PatternBase() uint16 {
	if bits(uint(ppu.ppuctrl), 3, 1) == 0 {
		return 0
	} else {
		return 0x1000
	}
}

func (ppu *Ppu) bgPatternBase() uint16 {
	if bits(uint(ppu.ppuctrl), 4, 1) == 0 {
		return 0
	} else {
		return 0x1000
	}
}

func (ppu *Ppu) spriteSize8x8() bool {
	return bits(uint(ppu.ppuctrl), 5, 1) == 0
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

func (ppu *Ppu) greyscale() bool {
	return ppu.ppumask&0x01 != 0
}

func (ppu *Ppu) showBgInLeftmost() bool {
	return ppu.ppumask&0x02 != 0
}

func (ppu *Ppu) showSpriteInLeftmost() bool {
	return ppu.ppumask&0x04 != 0
}

func (ppu *Ppu) showBg() bool {
	return ppu.ppumask&0x08 != 0
}

func (ppu *Ppu) showSprite() bool {
	return ppu.ppumask&0x10 != 0
}

func (ppu *Ppu) emphasizeRed() bool {
	return ppu.ppumask&0x20 != 0
}

func (ppu *Ppu) emphasizeGreen() bool {
	return ppu.ppumask&0x40 != 0
}

func (ppu *Ppu) emphasizeBlue() bool {
	return ppu.ppumask&0x80 != 0
}

func (ppu *Ppu) readPpustatus() uint8 {
	v := ppu.ppustatus
	ppu.ppustatus &= ^PPUSTATUS_V
	//Debug("ppustatus=%02Xh\n", v)
	return v
}

func (ppu *Ppu) writePpuscroll(v uint8) {
	//Debug("writePpuscroll: %02Xh\n", v)
	if ppu.ppuscrollw {
		ppu.ppuscrollynew = v
		ppu.ppuscrollw = false
	} else {
		ppu.ppuscrollxnew = v
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
	//Debug("ppu.ppuaddr=%04X data=%02X\n", ppu.ppuaddr, v)
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
	cpuaddr := uint16(hi) << 8
	for i := 0; i < 0x100; i++ {
		ppu.oam[i] = ppu.nes.mem.Read8(cpuaddr)
		cpuaddr++
	}
}

func (ppu *Ppu) RenderScreen() {
	for row := range ppu.screen {
		ppu.renderScanline(uint(row))
	}
}

func (ppu *Ppu) renderScanline(row uint) {
	ppu.ppuscrollx = ppu.ppuscrollxnew
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
	index := u/attrMetatileSize + (v/attrMetatileSize)*attrMetatilesInRow
	attr := attributetable[index]
	s := (u / attrTileSize) % 2
	t := (v / attrTileSize) % 2
	return (attr >> ((s + t*2) * 2) & 0x03)
}

func getNametableIndex(x uint, y uint) uint {
	return (x/ScreenSizePixX)%2 + ((y/ScreenSizePixY)%2)*2
}

func (ppu *Ppu) renderPixel(col uint, row uint) {
	const patternEntryBytes = 16
	const tileSize = 8
	const hiOffset = 8

	if ppu.showBg() {
		x := col + uint(ppu.ppuscrollx)
		if ppu.ppuctrl&0x01 != 0 {
			x += 256
		}
		y := row + uint(ppu.ppuscrolly)
		if ppu.ppuctrl&0x02 != 0 {
			y += 240
		}
		nametableIndex := getNametableIndex(x, y)
		nametable := ppu.nametable[nametableIndex]

		index := getIndexInNametable(x, y)
		patternIndex := uint16(uint(nametable[index])*patternEntryBytes + y%tileSize)

		lo := ppu.vramRead8(ppu.bgPatternBase() + patternIndex)
		hi := ppu.vramRead8(ppu.bgPatternBase() + patternIndex + hiOffset)

		attributetable := ppu.attributetable[nametableIndex]
		paletteIndex := getPaletteIndex(attributetable, x, y)

		pix := bits(uint(lo), tileSize-1-x%tileSize, 1) |
			(bits(uint(hi), tileSize-1-x%tileSize, 1) << 1)
		if pix == 0 {
			ppu.screen[row][col] = ppu.bgPalette[0][0]
			//Debug("ppu.bgPalette[0][0]=%02X\n", ppu.bgPalette[0][0])
		} else {
			ppu.screen[row][col] = ppu.bgPalette[paletteIndex][pix]
		}
	}

	if ppu.showSprite() && ppu.oamScreen[row][col] != 0 && !(ppu.oamBehindBg[row][col] && ppu.screen[row][col] != ppu.bgPalette[0][0]) {
		ppu.screen[row][col] = ppu.oamScreen[row][col]
	}

	if ppu.ppustatus&PPUSTATUS_S == 0 && ppu.showSprite() && ppu.showBg() {
		if ppu.oamMap[row][col] == 0 && ppu.screen[row][col] != 0 {
			//Debug("Sprite zero hit\n")
			ppu.ppustatus |= PPUSTATUS_S
		}
	}
}

func vramPage(a uint16) uint {
	return uint(a) >> vramPageShift
}

func vramAddressFix(a uint16) uint16 {
	if a >= 0x3F20 && a < 0x4000 {
		// 0x3F20-0x3FFF: mirrors of 0x3F00-0x3F1F
		a &= 0xFF1F
	} else if a == 0x3f10 || a == 0x3f14 || a == 0x3f18 || a == 0x3fc0 {
		a &= 0x3f0f
	}
	return a
}

func vramOffest(a uint16) uint {
	return uint(vramAddressFix(a)) & (vramPageSize - 1)
}

func (ppu *Ppu) mapExtMem(address uint16, extmem []uint8, bytes int) {
	for i := uint16(0); i < uint16(bytes); i += vramPageSize {
		ppu.lvram[vramPage(address+i)] = extmem[i : i+vramPageSize]
	}
}

func (ppu *Ppu) vramWrite8(address uint16, v uint8) {
	ppu.lvram[vramPage(address)][vramOffest(address)] = v
}

func (ppu *Ppu) vramRead8(address uint16) uint8 {
	return ppu.lvram[vramPage(address)][vramOffest(address)]
}

func (ppu *Ppu) reset() {
	ppu.ppuscrollx = 0
	ppu.ppuscrolly = 0
	ppu.ppuscrollynew = 0
	ppu.currentScanline = preRenderScanline
	ppu.clock = 0
}

func toPpuClockDelta(cpuclockDelta uint) uint {
	const ppuCpuClockRatio = 3
	return cpuclockDelta * ppuCpuClockRatio
}

const firstVisibleScanline = 0
const lastVisibleScanline = 239
const postRenderScanline = 240
const firstVBlankScanline = 241
const preRenderScanline = 261
const numScanlines = 262

func scanlineToClock(row uint) uint {
	return (row + 1) * 341
}

type Sprite struct {
	index int
	oam   []uint8
}

func (ppu *Ppu) getSprite(index int) *Sprite {
	sp := new(Sprite)
	sp.oam = ppu.oam[4*index : 4*(index+1)]
	sp.index = index
	return sp
}

func (sp *Sprite) posX() int {
	return int(sp.oam[3])
}

func (sp *Sprite) posY() int {
	return int(sp.oam[0] + 1)
}

func (sp *Sprite) visible() bool {
	if sp.oam[0] >= 0xef && sp.oam[0] <= 0xff {
		return false
	}
	return true
}

func (sp *Sprite) patternAddress(ppu *Ppu, bottomHalf bool) uint16 {
	if ppu.spriteSize8x8() {
		return uint16(sp.oam[1])*16 + ppu.sprite8x8PatternBase()
	} else {
		base := uint16(sp.oam[1]&0xfe)*16 + uint16(sp.oam[1]&0x01)<<12
		if bottomHalf {
			base += 16
		}
		return base
	}
}

func (sp *Sprite) paletteIndex() int {
	return int(bits(uint(sp.oam[2]), 0, 2))
}

func (sp *Sprite) behindBg() bool {
	return int(bits(uint(sp.oam[2]), 5, 1)) != 0
}

func (sp *Sprite) hFlip() bool {
	return bits(uint(sp.oam[2]), 6, 1) == 1
}

func (sp *Sprite) vFlip() bool {
	return bits(uint(sp.oam[2]), 7, 1) == 1
}

func (ppu *Ppu) putSpriteTile(sp *Sprite, x int, y int, bottomHalf bool, behindBg bool) {
	base := sp.patternAddress(ppu, bottomHalf)
	for j := 0; j < 8; j++ {
		var lo, hi uint8
		if sp.vFlip() {
			lo = ppu.vramRead8(base + 7 - uint16(j))
			hi = ppu.vramRead8(base + 7 - uint16(j) + 8)
		} else {
			lo = ppu.vramRead8(base + uint16(j))
			hi = ppu.vramRead8(base + uint16(j) + 8)
		}
		for i := 0; i < 8; i++ {
			var pix uint
			if sp.hFlip() {
				pix = bits(uint(lo), uint(i%8), 1) | (bits(uint(hi), uint(i%8), 1) << 1)
			} else {
				pix = bits(uint(lo), uint(7-i%8), 1) | (bits(uint(hi), uint(7-i%8), 1) << 1)
			}
			c := uint8(0)
			if pix != 0 {
				c = ppu.spPalette[sp.paletteIndex()][pix]
			}

			u := x + i
			v := y + j
			if c != 0 && u < ScreenSizePixX && v < ScreenSizePixY {
				ppu.oamScreen[v][u] = c
				ppu.oamBehindBg[v][u] = behindBg
				ppu.oamMap[v][u] = uint8(sp.index)
			}
		}
	}
}

func (ppu *Ppu) preRenderSprite(spriteIndex int) {
	sp := ppu.getSprite(spriteIndex)
	if sp.visible() {
		x := sp.posX()
		y := sp.posY()
		behindBg := sp.behindBg()
		ppu.putSpriteTile(sp, x, y, false, behindBg)
		if !ppu.spriteSize8x8() {
			ppu.putSpriteTile(sp, x, y+8, true, behindBg)
		}
	}
}

func (ppu *Ppu) prepSprite() {
	for i := 0; i < ScreenSizePixY; i++ {
		for j := 0; j < ScreenSizePixX; j++ {
			ppu.oamScreen[i][j] = 0
			ppu.oamBehindBg[i][j] = false
		}
	}

	for i := 0; i < ScreenSizePixY; i++ {
		for j := 0; j < ScreenSizePixX; j++ {
			ppu.oamMap[i][j] = nullSpIndex
		}
	}

	for i := 63; i >= 0; i-- {
		ppu.preRenderSprite(i)
	}
}

func (ppu *Ppu) giveCpuClockDelta(cpuclockDelta uint) bool {
	lvs := false
	ppu.clock += toPpuClockDelta(cpuclockDelta)
	for row := ppu.currentScanline; ppu.clock >= scanlineToClock(row); row++ {
		if row >= firstVisibleScanline && row <= lastVisibleScanline {
			ppu.renderScanline(row)
		}

		if row == postRenderScanline {
			//Debug("firstVBlankScanline\n")
			ppu.ppustatus |= PPUSTATUS_V
			if ppu.vblankNmi() {
				ppu.nes.cpu.setNmi()
			}
		}

		if row == lastVisibleScanline {
			//Debug("lastVisibleScanline\n")
			ppu.nes.display.Render(&ppu.screen)
			lvs = true
		}

		if row == preRenderScanline-1 {
			ppu.ppustatus &= ^(PPUSTATUS_V | PPUSTATUS_S)
			ppu.prepSprite()
		}

		if row == preRenderScanline {
			ppu.updatePpuscrolly()
			ppu.currentScanline = 0
			ppu.clock = 0
		} else {
			ppu.currentScanline++
		}
	}
	return lvs
}

func NewPpu(nes *Nes) *Ppu {
	ppu := new(Ppu)
	ppu.nes = nes

	//
	// Initialize VRAM
	//
	for i := range ppu.lvram {
		Debug("NewPpu: %04X-%04X\n", vramPageSize*i, vramPageSize*(i+1))
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
		Debug("Name table mirror setitng\n")
		if rom.verticalMirror {
			Debug("Vertical mirror setup\n")
			for i := uint16(0); i < 0x400; i += vramPageSize {
				ppu.lvram[vramPage(0x2800+i)] = ppu.lvram[vramPage(0x2000+i)]
				ppu.lvram[vramPage(0x2c00+i)] = ppu.lvram[vramPage(0x2400+i)]
			}
		} else {
			Debug("Horizontal mirror setup\n")
			for i := uint16(0); i < 0x400; i += vramPageSize {
				ppu.lvram[vramPage(0x2400+i)] = ppu.lvram[vramPage(0x2000+i)]
				ppu.lvram[vramPage(0x2c00+i)] = ppu.lvram[vramPage(0x2800+i)]
			}
		}
	}

	Debug("Setting VRAM mirror\n")
	for a := uint16(0x3000); a < 0x3f00; a += vramPageSize {
		ppu.lvram[vramPage(a)] = ppu.lvram[vramPage(a-0x1000)]
	}

	//
	// Map Name Tables and Attribute Tables
	//
	Debug("Name table slices setup\n")
	if rom.fourScreenVram {
		Debug("Four screens\n")
		ppu.nametable[0] = ppu.vram[0x2000:0x23c0]
		ppu.attributetable[0] = ppu.vram[0x23c0:0x2400]
		ppu.nametable[1] = ppu.vram[0x2400:0x27c0]
		ppu.attributetable[1] = ppu.vram[0x27c0:0x2800]
		ppu.nametable[2] = ppu.vram[0x2800:0x2bc0]
		ppu.attributetable[2] = ppu.vram[0x2bc0:0x2c00]
		ppu.nametable[3] = ppu.vram[0x2c00:0x2fc0]
		ppu.attributetable[3] = ppu.vram[0x2fc0:0x3000]
	} else if rom.verticalMirror {
		Debug("Vertical mirror\n")
		ppu.nametable[0] = ppu.vram[0x2000:0x23c0]
		ppu.attributetable[0] = ppu.vram[0x23c0:0x2400]
		ppu.nametable[1] = ppu.vram[0x2400:0x27c0]
		ppu.attributetable[1] = ppu.vram[0x27c0:0x2800]
		ppu.nametable[2] = ppu.nametable[0]
		ppu.attributetable[2] = ppu.attributetable[0]
		ppu.nametable[3] = ppu.nametable[1]
		ppu.attributetable[3] = ppu.attributetable[1]
	} else {
		Debug("Horizontal mirror\n")
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
	ppu.spPalette[0] = ppu.vram[0x3f10:0x3f14]
	ppu.spPalette[1] = ppu.vram[0x3f14:0x3f18]
	ppu.spPalette[2] = ppu.vram[0x3f18:0x3f1c]
	ppu.spPalette[3] = ppu.vram[0x3f1c:0x3f20]

}
