package nespkg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Gamepad interface {
	regWrite(val uint8)
	regRead() uint8
}

const (
	ButtonA = iota
	ButtonB
	ButtonSelect
	ButtonStart
	ButtonUp
	ButtonDown
	ButtonLeft
	ButtonRight
	ButtonMax
)

type GamepadButton int

type KbdReader struct {
	buttonPressed [ButtonMax]bool
}

func (r *KbdReader) KeyUpCallback(button GamepadButton) {
	r.buttonPressed[button] = false
}

func (r *KbdReader) KeyDownCallback(button GamepadButton) {
	r.buttonPressed[button] = true
}

func NewKbdReader() *KbdReader {
	r := new(KbdReader)
	for i := range r.buttonPressed {
		r.buttonPressed[i] = false
	}
	return r
}

func bits(v uint, pos uint, width uint) uint {
	return (v >> pos) & ((1 << width) - 1)
}

type Nes struct {
	cpu     *Cpu
	ppu     *Ppu
	Pad     [2]Gamepad
	Kbd     *KbdReader
	mem     *MainMemory
	rom     *NesRom
	mapper  Mapper
	display Display
	dbg     *Debugger
}

type Display interface {
	Render(screen *[ScreenSizePixY][ScreenSizePixX]uint8)
}

type Conf struct {
	DebugEnable    bool
	TraceEnable    bool
	MemTraceEnable bool
}

var DebugEnable bool = false
var MemTraceEnable bool = false

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

func NewNes(conf *Conf, d Display) *Nes {
	DebugEnable = conf.DebugEnable
	MemTraceEnable = conf.MemTraceEnable
	nes := new(Nes)
	nes.cpu = NewCpu(nes)
	nes.ppu = NewPpu(nes)
	nes.mem = NewMainMemory(nes)
	nes.Kbd = NewKbdReader()
	nes.cpu.mem = nes.mem
	nes.display = d
	nes.Pad[0] = NewUsbGamepad(0)
	if nes.Pad[0] == nil {
		nes.Pad[0] = NewKbdGamepad(nes.Kbd)
	}
	nes.dbg = NewDebugger(conf, nes)
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
	var err3 error
	nes.mapper, err3 = MakeMapper(nes, rom.mapperNum)
	if err3 != nil {
		return err3
	}
	nes.mapper.Init()
	Debug("calling PostRomLoadSetup\n")
	nes.ppu.PostRomLoadSetup()
	Debug("returning from LoadRom\n")
	return nil
}

func (nes *Nes) Stop() {
	nes.dbg.step = true
}

const framePeriodMicroSeconds = time.Microsecond * 16666

func (nes *Nes) Run() {
	nes.Reset()
	lastRefreshTime := time.Now()
	for {
		cycle := nes.cpu.executeInst()
		if nes.ppu.giveCpuClockDelta(cycle) {
			t := time.Since(lastRefreshTime)
			time.Sleep(framePeriodMicroSeconds - t)
			lastRefreshTime = time.Now()
		}
		nes.dbg.hook()
	}
}

type Debugger struct {
	nes     *Nes
	ibp     [8]uint16
	step    bool
	trace   bool
	scanner *bufio.Scanner
	prevCmd DbgCmd
}

func NewDebugger(conf *Conf, nes *Nes) *Debugger {
	dbg := new(Debugger)
	dbg.nes = nes
	dbg.step = false
	dbg.trace = conf.TraceEnable
	dbg.prevCmd = nil
	for i := range dbg.ibp {
		dbg.ibp[i] = 0
	}
	dbg.scanner = bufio.NewScanner(os.Stdin)
	return dbg
}

func (dbg *Debugger) Break() bool {
	if dbg.step {
		dbg.step = false
		return true
	}

	for _, ibp := range dbg.ibp {
		if ibp == dbg.nes.cpu.pc {
			return true
		}
	}

	return false
}

type DbgCmdTableEntry struct {
	cmdMaker func([]string) (DbgCmd, error)
}

type DbgCmd interface {
	execCmd(dbg *Debugger) bool
}

var DbgCmdTable = map[string]DbgCmdTableEntry{
	"a":   {NewDbgCmdAsm},
	"s":   {func(args []string) (DbgCmd, error) { return new(DbgCmdStep), nil }},
	"c":   {func(args []string) (DbgCmd, error) { return new(DbgCmdCont), nil }},
	"t":   {func(args []string) (DbgCmd, error) { return new(DbgCmdTrace), nil }},
	"mt":  {func(args []string) (DbgCmd, error) { return new(DbgCmdMemoryTrace), nil }},
	"p":   {func(args []string) (DbgCmd, error) { return new(DbgCmdPpureg), nil }},
	"m":   {NewDbgCmdMem},
	"v":   {NewDbgCmdVramRead},
	"r":   {func(args []string) (DbgCmd, error) { return new(DbgCmdRep), nil }},
	"":    {func(args []string) (DbgCmd, error) { return new(DbgCmdRep), nil }},
	"nop": {func(args []string) (DbgCmd, error) { return new(DbgCmdNop), nil }},
}

type DbgCmdBase struct {
	args []string
}

func (cmd *DbgCmdBase) execCmd(dbg *Debugger) bool {
	return true
}

type DbgCmdStep struct {
	DbgCmdBase
}

func (cmd *DbgCmdStep) execCmd(dbg *Debugger) bool {
	dbg.step = true
	return false
}

type DbgCmdPpureg struct {
	DbgCmdBase
}

func (cmd *DbgCmdPpureg) execCmd(dbg *Debugger) bool {
	ppu := dbg.nes.ppu
	fmt.Printf("ppuctrl        = %02Xh\n", ppu.ppuctrl)
	fmt.Printf("ppumask        = %02Xh\n", ppu.ppumask)
	fmt.Printf("ppustatus      = %02Xh\n", ppu.ppustatus)
	fmt.Printf("ppuscrollx     = %02Xh\n", ppu.ppuscrollx)
	fmt.Printf("ppuscrolly     = %02Xh\n", ppu.ppuscrolly)
	fmt.Printf("ppuscrollynew  = %02Xh\n", ppu.ppuscrollynew)
	fmt.Printf("ppuscrollw     = %t\n", ppu.ppuscrollw)
	fmt.Printf("ppuaddr        = %02Xh\n", ppu.ppuaddr)
	fmt.Printf("ppuaddrw       = %t\n", ppu.ppuaddrw)
	fmt.Printf("ppudata        = %02Xh\n", ppu.ppudata)
	return true
}

type DbgCmdRep struct {
	DbgCmdBase
}

func (cmd *DbgCmdRep) execCmd(dbg *Debugger) bool {
	if dbg.prevCmd != nil {
		return dbg.prevCmd.execCmd(dbg)
	} else {
		return false
	}
}

type DbgCmdCont struct {
	DbgCmdBase
}

func (cmd *DbgCmdCont) execCmd(dbg *Debugger) bool {
	dbg.step = false
	return false
}

type DbgCmdNop struct {
	DbgCmdBase
}

func (cmd *DbgCmdNop) execCmd(dbg *Debugger) bool {
	return true
}

type DbgCmdTrace struct {
	DbgCmdBase
}

func (cmd *DbgCmdTrace) execCmd(dbg *Debugger) bool {
	if dbg.trace {
		dbg.trace = false
	} else {
		dbg.trace = true
	}
	fmt.Printf("trace = %t\n", dbg.trace)
	return true
}

type DbgCmdMemoryTrace struct {
	DbgCmdBase
}

func (cmd *DbgCmdMemoryTrace) execCmd(dbg *Debugger) bool {
	if MemTraceEnable {
		MemTraceEnable = false
	} else {
		MemTraceEnable = true
	}
	fmt.Printf("memory trace = %t\n", MemTraceEnable)
	return true
}

type DbgCmdMem struct {
	DbgCmdBase
	address uint16
	length  int
}

func NewDbgCmdMem(args []string) (DbgCmd, error) {
	if len(args) != 2 {
		return nil, errors.New("mem: invalid arguments")
	}
	c := new(DbgCmdMem)
	if a, err := strconv.ParseUint(args[0], 16, 16); err == nil {
		c.address = uint16(a)
	} else {
		return nil, errors.New("mem: invalid arguments")
	}
	if l, err := strconv.ParseInt(args[1], 16, 0); err == nil {
		c.length = int(l)
	} else {
		return nil, errors.New("mem: invalid arguments")
	}
	return c, nil
}

type DbgCmdVramRead struct {
	DbgCmdBase
	address uint16
	length  int
}

func NewDbgCmdVramRead(args []string) (DbgCmd, error) {
	if len(args) != 2 {
		return nil, errors.New("vramread: invalid arguments")
	}
	c := new(DbgCmdVramRead)
	if a, err := strconv.ParseUint(args[0], 16, 16); err == nil {
		c.address = uint16(a)
	} else {
		return nil, errors.New("vramread: invalid arguments")
	}
	if l, err := strconv.ParseInt(args[1], 16, 0); err == nil {
		c.length = int(l)
	} else {
		return nil, errors.New("vramread: invalid arguments")
	}
	return c, nil
}

func (cmd *DbgCmdVramRead) execCmd(dbg *Debugger) bool {
	r := func(a uint16) uint8 { return dbg.nes.ppu.vramRead8(a) }
	dumpMem(r, cmd.address, cmd.length)
	return true
}

type DbgCmdAsm struct {
	DbgCmdBase
	address uint16
	length  int
}

func NewDbgCmdAsm(args []string) (DbgCmd, error) {
	if len(args) != 2 {
		return nil, errors.New("invalid arguments")
	}
	c := new(DbgCmdAsm)
	if a, err := strconv.ParseUint(args[0], 16, 16); err == nil {
		c.address = uint16(a)
	} else {
		return nil, errors.New("invalid arguments")
	}
	if l, err := strconv.ParseInt(args[1], 16, 0); err == nil {
		c.length = int(l)
	} else {
		return nil, errors.New("invalid arguments")
	}
	return c, nil
}

func (cmd *DbgCmdAsm) execCmd(dbg *Debugger) bool {
	for pc := cmd.address; pc < cmd.address+uint16(cmd.length); {
		err, b, s := GetAsmStr(dbg.nes.mem, pc)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println(s)
		pc += uint16(b)
	}
	return true
}

func dumpMem(r func(uint16) uint8, start uint16, length int) {
	fmt.Println("      00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f 10 11 12 13 14 15 16 17 18 19 1a 1b 1c 1d 1e 1f")
	fmt.Println("-----------------------------------------------------------------------------------------------------")
	for i := 0; i < length; i++ {
		address := start + uint16(i)
		if i%32 == 0 {
			fmt.Printf("%04X:", address)
		}
		fmt.Printf(" %02X", r(address))
		if i%32 == 31 {
			fmt.Println("")
		}
	}
	fmt.Println("")
}

func (cmd *DbgCmdMem) execCmd(dbg *Debugger) bool {
	r := func(a uint16) uint8 { return dbg.nes.cpu.mem.Read8(a) }
	dumpMem(r, cmd.address, cmd.length)
	return true
}

func (dbg *Debugger) NewCmd() (DbgCmd, error) {
	t := []string{"nop"}
	if dbg.scanner.Scan() {
		t = strings.Split(dbg.scanner.Text(), " ")
	}
	e, ok := DbgCmdTable[t[0]]
	if ok {
		return e.cmdMaker(t[1:])

	} else {
		return nil, errors.New("Invalid command")
	}
}

func (dbg *Debugger) hook() {
	if !dbg.Break() {
		return
	}

	inDebug := true
	for inDebug {
		fmt.Print("dbg> ")
		cmd, err := dbg.NewCmd()
		if err == nil {
			inDebug = cmd.execCmd(dbg)
			switch cmd.(type) {
			case *DbgCmdRep:
			default:
				dbg.prevCmd = cmd
			}
		} else {
			fmt.Println(err)
		}
	}
}
