package nespkg

import (
	"errors"
	"fmt"
)

const P_N_POS = 7
const P_V_POS = 6
const P_R_POS = 5
const P_B_POS = 4
const P_D_POS = 3
const P_I_POS = 2
const P_Z_POS = 1
const P_C_POS = 0

const P_N = (uint8(1) << P_N_POS)
const P_V = (uint8(1) << P_V_POS)
const P_R = (uint8(1) << P_R_POS)
const P_B = (uint8(1) << P_B_POS)
const P_D = (uint8(1) << P_D_POS)
const P_I = (uint8(1) << P_I_POS)
const P_Z = (uint8(1) << P_Z_POS)
const P_C = (uint8(1) << P_C_POS)

const VEC_NMI = 0xfffa
const VEC_RESET = 0xfffc
const VEC_IRQ = 0xfffe

const NES_SIZE_H = 256
const NES_SIZE_V = 240

type InstParams struct {
	mnemonic string
	mode     InstMode
	bytes    uint
	cycle    uint
}

type InstMode int

const (
	abs = iota
	abx
	aby
	acc
	imm
	imp
	ind
	inx
	iny
	rel
	zrp
	zpx
	zpy
)

type ModeOps struct {
	getAddress   func(cpu *Cpu) uint16
	getValue     func(cpu *Cpu) uint8
	setValue     func(cpu *Cpu, v uint8)
	getOpdString func(mem Memory, pc uint16) string
}

var modeOpsTable = map[InstMode]ModeOps{
	abs: {getAddrAbs, getValueAbs, setValueAbs, getOpdstrAbs},
	abx: {getAddrAbx, getValueAbx, setValueAbx, getOpdstrAbx},
	aby: {getAddrAby, getValueAby, setValueAby, getOpdstrAby},
	acc: {nil, getValueAcc, setValueAcc, getOpdstrAcc},
	imm: {nil, getValueImm, nil, getOpdstrImm},
	imp: {nil, nil, nil, getOpdstrImp},
	ind: {getAddrInd, nil, nil, getOpdstrInd},
	inx: {getAddrInx, getValueInx, setValueInx, getOpdstrInx},
	iny: {getAddrIny, getValueIny, setValueIny, getOpdstrIny},
	rel: {nil, getValueImm, nil, getOpdstrRel},
	zrp: {getAddrZrp, getValueZrp, setValueZrp, getOpdstrZrp},
	zpx: {getAddrZpx, getValueZpx, setValueZpx, getOpdstrZpx},
	zpy: {getAddrZpy, getValueZpy, setValueZpy, getOpdstrZpy},
}

var instTable = map[uint8]InstParams{
	0x00: {"brk", imp, 1, 7},
	0x01: {"ora", inx, 2, 6},
	0x05: {"ora", zrp, 2, 3},
	0x06: {"asl", zrp, 2, 5},
	0x08: {"php", imp, 1, 3},
	0x09: {"ora", imm, 2, 2},
	0x0A: {"asl", acc, 1, 2},
	0x0D: {"ora", abs, 3, 4},
	0x0E: {"asl", abs, 3, 6},
	0x10: {"bpl", rel, 2, 2},
	0x11: {"ora", iny, 2, 5},
	0x15: {"ora", zpx, 2, 4},
	0x16: {"asl", zpx, 2, 6},
	0x18: {"clc", imp, 1, 2},
	0x19: {"ora", aby, 3, 4},
	0x1D: {"ora", abx, 3, 4},
	0x1E: {"asl", abx, 3, 7},
	0x20: {"jsr", abs, 3, 6},
	0x21: {"and", inx, 2, 6},
	0x24: {"bit", zrp, 2, 3},
	0x25: {"and", zrp, 2, 3},
	0x26: {"rol", zrp, 2, 5},
	0x28: {"plp", imp, 1, 4},
	0x29: {"and", imm, 2, 2},
	0x2A: {"rol", acc, 1, 2},
	0x2C: {"bit", abs, 3, 4},
	0x2D: {"and", abs, 3, 4},
	0x2E: {"rol", abs, 3, 6},
	0x30: {"bmi", rel, 2, 2},
	0x31: {"and", iny, 2, 5},
	0x35: {"and", zpx, 2, 4},
	0x36: {"rol", zpx, 2, 6},
	0x38: {"sec", imp, 1, 2},
	0x39: {"and", aby, 3, 4},
	0x3D: {"and", abx, 3, 4},
	0x3E: {"rol", abx, 3, 7},
	0x40: {"rti", imp, 1, 6},
	0x41: {"eor", inx, 2, 6},
	0x45: {"eor", zrp, 2, 3},
	0x46: {"lsr", zrp, 2, 5},
	0x48: {"pha", imp, 1, 3},
	0x49: {"eor", imm, 2, 2},
	0x4A: {"lsr", acc, 1, 2},
	0x4C: {"jmp", abs, 3, 3},
	0x4D: {"eor", abs, 3, 4},
	0x4E: {"lsr", abs, 3, 6},
	0x50: {"bvc", rel, 2, 2},
	0x51: {"eor", iny, 2, 5},
	0x55: {"eor", zpx, 2, 4},
	0x56: {"lsr", zpx, 2, 6},
	0x58: {"cli", imp, 1, 2},
	0x59: {"eor", aby, 3, 4},
	0x5D: {"eor", abx, 3, 4},
	0x5E: {"lsr", abx, 3, 7},
	0x60: {"rts", imp, 1, 6},
	0x61: {"adc", inx, 2, 6},
	0x65: {"adc", zrp, 2, 3},
	0x66: {"ror", zrp, 2, 5},
	0x68: {"pla", imp, 1, 4},
	0x69: {"adc", imm, 2, 2},
	0x6A: {"ror", acc, 1, 2},
	0x6C: {"jmp", ind, 3, 5},
	0x6D: {"adc", abs, 3, 4},
	0x6E: {"ror", abs, 3, 6},
	0x70: {"bvs", rel, 2, 2},
	0x71: {"adc", iny, 2, 5},
	0x75: {"adc", zpx, 2, 4},
	0x76: {"ror", zpx, 2, 6},
	0x78: {"sei", imp, 1, 2},
	0x79: {"adc", aby, 3, 4},
	0x7D: {"adc", abx, 3, 4},
	0x7E: {"ror", abx, 3, 7},
	0x81: {"sta", inx, 2, 6},
	0x84: {"sty", zrp, 2, 3},
	0x85: {"sta", zrp, 2, 3},
	0x86: {"stx", zrp, 2, 3},
	0x88: {"dey", imp, 1, 2},
	0x8A: {"txa", imp, 1, 2},
	0x8C: {"sty", abs, 3, 4},
	0x8D: {"sta", abs, 3, 4},
	0x8E: {"stx", abs, 3, 4},
	0x90: {"bcc", rel, 2, 2},
	0x91: {"sta", iny, 2, 6},
	0x94: {"sty", zpx, 2, 4},
	0x95: {"sta", zpx, 2, 4},
	0x96: {"stx", zpy, 2, 4},
	0x98: {"tya", imp, 1, 2},
	0x99: {"sta", aby, 3, 5},
	0x9A: {"txs", imp, 1, 2},
	0x9D: {"sta", abx, 3, 5},
	0xA0: {"ldy", imm, 2, 2},
	0xA1: {"lda", inx, 2, 6},
	0xA2: {"ldx", imm, 2, 2},
	0xA4: {"ldy", zrp, 2, 3},
	0xA5: {"lda", zrp, 2, 3},
	0xA6: {"ldx", zrp, 2, 3},
	0xA8: {"tay", imp, 1, 2},
	0xA9: {"lda", imm, 2, 2},
	0xAA: {"tax", imp, 1, 2},
	0xAC: {"ldy", abs, 3, 4},
	0xAD: {"lda", abs, 3, 4},
	0xAE: {"ldx", abs, 3, 4},
	0xB0: {"bcs", rel, 2, 2},
	0xB1: {"lda", iny, 2, 5},
	0xB4: {"ldy", zpx, 2, 4},
	0xB5: {"lda", zpx, 2, 4},
	0xB6: {"ldx", zpy, 2, 4},
	0xB8: {"clv", imp, 1, 2},
	0xB9: {"lda", aby, 3, 4},
	0xBA: {"tsx", imp, 1, 2},
	0xBC: {"ldy", abx, 3, 4},
	0xBD: {"lda", abx, 3, 4},
	0xBE: {"ldx", aby, 3, 4},
	0xC0: {"cpy", imm, 2, 2},
	0xC1: {"cmp", inx, 2, 6},
	0xC4: {"cpy", zrp, 2, 3},
	0xC5: {"cmp", zrp, 2, 3},
	0xC6: {"dec", zrp, 2, 5},
	0xC8: {"iny", imp, 1, 2},
	0xC9: {"cmp", imm, 2, 2},
	0xCA: {"dex", imp, 1, 2},
	0xCC: {"cpy", abs, 3, 4},
	0xCD: {"cmp", abs, 3, 4},
	0xCE: {"dec", abs, 3, 6},
	0xD0: {"bne", rel, 2, 2},
	0xD1: {"cmp", iny, 2, 5},
	0xD5: {"cmp", zpx, 2, 4},
	0xD6: {"dec", zpx, 2, 6},
	0xD8: {"cld", imp, 1, 2},
	0xD9: {"cmp", aby, 3, 4},
	0xDD: {"cmp", abx, 3, 4},
	0xDE: {"dec", abx, 3, 7},
	0xE0: {"cpx", imm, 2, 2},
	0xE1: {"sbc", inx, 2, 6},
	0xE4: {"cpx", zrp, 2, 3},
	0xE5: {"sbc", zrp, 2, 3},
	0xE6: {"inc", zrp, 2, 5},
	0xE8: {"inx", imp, 1, 2},
	0xE9: {"sbc", imm, 2, 2},
	0xEA: {"nop", imp, 1, 2},
	0xEC: {"cpx", abs, 3, 4},
	0xED: {"sbc", abs, 3, 4},
	0xEE: {"inc", abs, 3, 6},
	0xF0: {"beq", rel, 2, 2},
	0xF1: {"sbc", iny, 2, 5},
	0xF5: {"sbc", zpx, 2, 4},
	0xF6: {"inc", zpx, 2, 6},
	0xF8: {"sed", imp, 1, 2},
	0xF9: {"sbc", aby, 3, 4},
	0xFD: {"sbc", abx, 3, 4},
	0xFE: {"inc", abx, 3, 7},
}

type InstHandler func(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint

var instHandlerTable = map[uint8]InstHandler{
	0x00: execBrk,
	0x01: execOra,
	0x05: execOra,
	0x06: execAsl,
	0x08: execPhp,
	0x09: execOra,
	0x0A: execAsl,
	0x0D: execOra,
	0x0E: execAsl,
	0x10: execBpl,
	0x11: execOra,
	0x15: execOra,
	0x16: execAsl,
	0x18: execClc,
	0x19: execOra,
	0x1D: execOra,
	0x1E: execAsl,
	0x20: execJsr,
	0x21: execAnd,
	0x24: execBit,
	0x25: execAnd,
	0x26: execRol,
	0x28: execPlp,
	0x29: execAnd,
	0x2A: execRol,
	0x2C: execBit,
	0x2D: execAnd,
	0x2E: execRol,
	0x30: execBmi,
	0x31: execAnd,
	0x35: execAnd,
	0x36: execRol,
	0x38: execSec,
	0x39: execAnd,
	0x3D: execAnd,
	0x3E: execRol,
	0x40: execRti,
	0x41: execEor,
	0x45: execEor,
	0x46: execLsr,
	0x48: execPha,
	0x49: execEor,
	0x4A: execLsr,
	0x4C: execJmp,
	0x4D: execEor,
	0x4E: execLsr,
	0x50: execBvc,
	0x51: execEor,
	0x55: execEor,
	0x56: execLsr,
	0x58: execCli,
	0x59: execEor,
	0x5D: execEor,
	0x5E: execLsr,
	0x60: execRts,
	0x61: execAdc,
	0x65: execAdc,
	0x66: execRor,
	0x68: execPla,
	0x69: execAdc,
	0x6A: execRor,
	0x6C: execJmp,
	0x6D: execAdc,
	0x6E: execRor,
	0x70: execBvs,
	0x71: execAdc,
	0x75: execAdc,
	0x76: execRor,
	0x78: execSei,
	0x79: execAdc,
	0x7D: execAdc,
	0x7E: execRor,
	0x81: execSta,
	0x84: execSty,
	0x85: execSta,
	0x86: execStx,
	0x88: execDey,
	0x8A: execTxa,
	0x8C: execSty,
	0x8D: execSta,
	0x8E: execStx,
	0x90: execBcc,
	0x91: execSta,
	0x94: execSty,
	0x95: execSta,
	0x96: execStx,
	0x98: execTya,
	0x99: execSta,
	0x9A: execTxs,
	0x9D: execSta,
	0xA0: execLdy,
	0xA1: execLda,
	0xA2: execLdx,
	0xA4: execLdy,
	0xA5: execLda,
	0xA6: execLdx,
	0xA8: execTay,
	0xA9: execLda,
	0xAA: execTax,
	0xAC: execLdy,
	0xAD: execLda,
	0xAE: execLdx,
	0xB0: execBcs,
	0xB1: execLda,
	0xB4: execLdy,
	0xB5: execLda,
	0xB6: execLdx,
	0xB8: execClv,
	0xB9: execLda,
	0xBA: execTsx,
	0xBC: execLdy,
	0xBD: execLda,
	0xBE: execLdx,
	0xC0: execCpy,
	0xC1: execCmp,
	0xC4: execCpy,
	0xC5: execCmp,
	0xC6: execDec,
	0xC8: execIny,
	0xC9: execCmp,
	0xCA: execDex,
	0xCC: execCpy,
	0xCD: execCmp,
	0xCE: execDec,
	0xD0: execBne,
	0xD1: execCmp,
	0xD5: execCmp,
	0xD6: execDec,
	0xD8: execCld,
	0xD9: execCmp,
	0xDD: execCmp,
	0xDE: execDec,
	0xE0: execCpx,
	0xE1: execSbc,
	0xE4: execCpx,
	0xE5: execSbc,
	0xE6: execInc,
	0xE8: execInx,
	0xE9: execSbc,
	0xEA: execNop,
	0xEC: execCpx,
	0xED: execSbc,
	0xEE: execInc,
	0xF0: execBeq,
	0xF1: execSbc,
	0xF5: execSbc,
	0xF6: execInc,
	0xF8: execSed,
	0xF9: execSbc,
	0xFD: execSbc,
	0xFE: execInc,
}

type Cpu struct {
	a          uint8
	x          uint8
	y          uint8
	s          uint8
	p          uint8
	pc         uint16
	nmiLatched bool
	nes        *Nes
	mem        *MainMemory
}

type Memory interface {
	Read8(uint16) uint8
	Read8NoTrace(uint16) uint8
	Read16(uint16) uint16
	Read16NoTrace(uint16) uint16
	Write8(uint16, uint8)
	Write8NoTrace(uint16, uint8)
	Write16(uint16, uint16)
	Write16NoTrace(uint16, uint16)
}

func (c *Cpu) Reset() {
	c.a = 0
	c.x = 0
	c.y = 0
	c.s = 0xfd
	c.p = 0x34
	c.pc = c.nes.mem.Read16(VEC_RESET)
	Debug("reset vector = %x\n", c.nes.mem.Read16(VEC_RESET))
}

func (c *Cpu) Regdump() {
	Debug("a = %02Xh\n", c.a)
	Debug("x = %02Xh\n", c.x)
	Debug("y = %02Xh\n", c.y)
	Debug("s = %02X\n", c.s)
	Debug("p = %02Xh\n", c.p)
	Debug("pc = %04Xh\n", c.pc)
}

const stackBase = 0x100

func (cpu *Cpu) push8(v uint8) {
	cpu.mem.Write8(stackBase+uint16(cpu.s), v)
	cpu.s--
}

func (cpu *Cpu) pop8() uint8 {
	cpu.s++
	return cpu.mem.Read8NoTrace(stackBase + uint16(cpu.s))
}

func (cpu *Cpu) push16(v uint16) {
	cpu.push8(uint8(v >> 8))
	cpu.push8(uint8(v & 0x0ff))
}

func (cpu *Cpu) pop16() uint16 {
	v := uint16(cpu.pop8())
	v |= uint16(cpu.pop8()) << 8
	return v
}

func getValueAcc(cpu *Cpu) uint8 {
	return cpu.a
}

func setValueAcc(cpu *Cpu, v uint8) {
	cpu.a = v
}

func getOpdstrAcc(mem Memory, pc uint16) string {
	return "A"
}

func getValueImm(cpu *Cpu) uint8 {
	return cpu.mem.Read8NoTrace(cpu.pc + 1)
}

func getOpdstrImm(mem Memory, pc uint16) string {
	return fmt.Sprintf("#$%02X", mem.Read8NoTrace(pc+1))
}

func getOpdstrImp(mem Memory, pc uint16) string {
	return ""
}

func getAddrZrp(cpu *Cpu) uint16 {
	return uint16(getValueImm(cpu))
}

func getValueZrp(cpu *Cpu) uint8 {
	return cpu.mem.Read8(getAddrZrp(cpu))
}

func setValueZrp(cpu *Cpu, v uint8) {
	cpu.mem.Write8(getAddrZrp(cpu), v)
}

func getOpdstrZrp(mem Memory, pc uint16) string {
	return fmt.Sprintf("$%02X", mem.Read8NoTrace(pc+1))
}

func getAddrZpx(cpu *Cpu) uint16 {
	return uint16(getValueImm(cpu) + cpu.x)
}

func getValueZpx(cpu *Cpu) uint8 {
	return cpu.mem.Read8(getAddrZpx(cpu))
}

func setValueZpx(cpu *Cpu, v uint8) {
	cpu.mem.Write8(getAddrZpx(cpu), v)
}

func getOpdstrZpx(mem Memory, pc uint16) string {
	return fmt.Sprintf("$%02X,X", mem.Read8NoTrace(pc+1))
}

func getAddrZpy(cpu *Cpu) uint16 {
	return uint16(getValueImm(cpu) + cpu.y)
}

func getValueZpy(cpu *Cpu) uint8 {
	return cpu.mem.Read8NoTrace(getAddrZpy(cpu))
}

func setValueZpy(cpu *Cpu, v uint8) {
	cpu.mem.Write8(getAddrZpy(cpu), v)
}

func getOpdstrZpy(mem Memory, pc uint16) string {
	return fmt.Sprintf("$%02X,Y", mem.Read8NoTrace(pc+1))
}

func getAddrInd(cpu *Cpu) uint16 {
	a := cpu.mem.Read16NoTrace(cpu.pc + 1)
	lo := cpu.mem.Read8NoTrace(a)
	hi := cpu.mem.Read8NoTrace(a&0xff00 | uint16(uint8(a&0x0ff)+1))
	return uint16(hi)<<8 | uint16(lo)
}

func getOpdstrInd(mem Memory, pc uint16) string {
	return fmt.Sprintf("($%04X)", mem.Read16NoTrace(pc+1))
}

func getAddrAbs(cpu *Cpu) uint16 {
	return cpu.mem.Read16NoTrace(cpu.pc + 1)
}

func getValueAbs(cpu *Cpu) uint8 {
	return cpu.mem.Read8NoTrace(getAddrAbs(cpu))
}

func setValueAbs(cpu *Cpu, v uint8) {
	cpu.mem.Write8(getAddrAbs(cpu), v)
}

func getOpdstrAbs(mem Memory, pc uint16) string {
	return fmt.Sprintf("$%04X", mem.Read16NoTrace(pc+1))
}

func getAddrAbx(cpu *Cpu) uint16 {
	return cpu.mem.Read16NoTrace(cpu.pc+1) + uint16(cpu.x)
}

func getValueAbx(cpu *Cpu) uint8 {
	return cpu.mem.Read8(getAddrAbx(cpu))
}

func setValueAbx(cpu *Cpu, v uint8) {
	cpu.mem.Write8(getAddrAbx(cpu), v)
}

func getOpdstrAbx(mem Memory, pc uint16) string {
	return fmt.Sprintf("$%04X,x", mem.Read16NoTrace(pc+1))
}

func getAddrAby(cpu *Cpu) uint16 {
	return cpu.mem.Read16NoTrace(cpu.pc+1) + uint16(cpu.y)
}

func getValueAby(cpu *Cpu) uint8 {
	return cpu.mem.Read8(getAddrAby(cpu))
}

func setValueAby(cpu *Cpu, v uint8) {
	cpu.mem.Write8(getAddrAby(cpu), v)
}

func getOpdstrAby(mem Memory, pc uint16) string {
	return fmt.Sprintf("$%04X,y", mem.Read16NoTrace(pc+1))
}

func getAddrInx(cpu *Cpu) uint16 {
	a := getValueImm(cpu) + cpu.x
	u := uint16(cpu.mem.Read8NoTrace(uint16(a)))
	u |= uint16(cpu.mem.Read8NoTrace(uint16(a+1))) << 8
	return u
}

func getValueInx(cpu *Cpu) uint8 {
	return cpu.mem.Read8(getAddrInx(cpu))
}

func setValueInx(cpu *Cpu, v uint8) {
	cpu.mem.Write8(getAddrInx(cpu), v)
}

func getOpdstrInx(mem Memory, pc uint16) string {
	return fmt.Sprintf("($%02X,X)", mem.Read8NoTrace(pc+1))
}

func getAddrIny(cpu *Cpu) uint16 {
	a := getValueImm(cpu)
	lo := cpu.mem.Read8NoTrace(uint16(a))
	hi := cpu.mem.Read8NoTrace(uint16(a + 1))
	return (uint16(hi)<<8 | uint16(lo)) + uint16(cpu.y)
}

func getValueIny(cpu *Cpu) uint8 {
	return cpu.mem.Read8(getAddrIny(cpu))
}

func setValueIny(cpu *Cpu, v uint8) {
	cpu.mem.Write8(getAddrIny(cpu), v)
}

func getOpdstrIny(mem Memory, pc uint16) string {
	return fmt.Sprintf("($%02X,Y)", mem.Read8NoTrace(pc+1))
}

func getOpdstrRel(mem Memory, pc uint16) string {
	d := uint16(int(pc+2) + int(int8(mem.Read8NoTrace(pc+1))))
	return fmt.Sprintf("$%04X", d)
}

func updateFlagsNz(v uint8, cpu *Cpu) {
	cpu.p &= ^(P_N | P_Z)
	if v&0x80 != 0 {
		cpu.p |= P_N
	}
	if v == 0 {
		cpu.p |= P_Z
	}
}

func execLda(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.a = modeOpsTable[mode].getValue(cpu)
	updateFlagsNz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execLdx(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.x = modeOpsTable[mode].getValue(cpu)
	updateFlagsNz(cpu.x, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execLdy(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.y = modeOpsTable[mode].getValue(cpu)
	updateFlagsNz(cpu.y, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execJmp(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.pc = modeOpsTable[mode].getAddress(cpu)
	return instTable[opc].cycle
}

func execSta(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	modeOpsTable[mode].setValue(cpu, cpu.a)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execStx(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	modeOpsTable[mode].setValue(cpu, cpu.x)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execSty(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	modeOpsTable[mode].setValue(cpu, cpu.y)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execTax(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.x = cpu.a
	updateFlagsNz(cpu.x, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execTxa(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.a = cpu.x
	updateFlagsNz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execTxs(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.s = cpu.x
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execTsx(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.x = cpu.s
	updateFlagsNz(cpu.x, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execTya(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.a = cpu.y
	updateFlagsNz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execTay(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.y = cpu.a
	updateFlagsNz(cpu.y, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execAdc(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	aPrev := cpu.a
	m := modeOpsTable[mode].getValue(cpu)
	u := uint16(cpu.a) + uint16(m) + uint16(cpu.p&P_C)
	v := int(int8(aPrev)) + int(int8(m)) + int(cpu.p&P_C)
	cpu.a = uint8(u)

	cpu.p &= ^(P_C | P_V)
	if u >= 0x100 {
		cpu.p |= P_C
	}
	if v < -128 || v > 127 {
		cpu.p |= P_V
	}
	updateFlagsNz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execAnd(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	v := modeOpsTable[mode].getValue(cpu)
	cpu.a &= v
	updateFlagsNz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execAsl(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	v := modeOpsTable[mode].getValue(cpu)
	if (v & 0x80) != 0 {
		cpu.p |= P_C
	} else {
		cpu.p &= ^P_C
	}
	v <<= 1
	modeOpsTable[mode].setValue(cpu, v)
	updateFlagsNz(v, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execBit(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p &= ^(P_V | P_N | P_Z)
	v := modeOpsTable[mode].getValue(cpu)
	if v&0x40 != 0 {
		cpu.p |= P_V
	}
	if v&0x80 != 0 {
		cpu.p |= P_N
	}
	if v&cpu.a == 0 {
		cpu.p |= P_Z
	}
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func cmpGen(u uint8, v uint8, cpu *Cpu) {
	cpu.p &= ^(P_C | P_Z | P_N)
	if u >= v {
		cpu.p |= P_C
	}
	if u == v {
		cpu.p |= P_Z
	}
	if (u-v)&0x80 != 0 {
		cpu.p |= P_N
	}
}

func execCmp(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	m := modeOpsTable[mode].getValue(cpu)
	cmpGen(cpu.a, m, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execCld(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p &= ^P_D
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execSed(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p |= P_D
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execCpx(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	v := modeOpsTable[mode].getValue(cpu)
	cmpGen(cpu.x, v, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execCpy(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	v := modeOpsTable[mode].getValue(cpu)
	cmpGen(cpu.y, v, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execDec(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	v := modeOpsTable[mode].getValue(cpu) - 1
	modeOpsTable[mode].setValue(cpu, v)
	updateFlagsNz(v, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execDex(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.x--
	updateFlagsNz(cpu.x, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execDey(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.y--
	updateFlagsNz(cpu.y, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execInc(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	v := modeOpsTable[mode].getValue(cpu) + 1
	modeOpsTable[mode].setValue(cpu, v)
	updateFlagsNz(v, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execInx(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.x++
	updateFlagsNz(cpu.x, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execIny(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.y++
	updateFlagsNz(cpu.y, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execEor(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.a ^= modeOpsTable[mode].getValue(cpu)
	updateFlagsNz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execLsr(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	u := modeOpsTable[mode].getValue(cpu)
	v := u >> 1
	cpu.p &= ^(P_C | P_Z | P_N)
	if u&0x01 != 0 {
		cpu.p |= P_C
	}
	modeOpsTable[mode].setValue(cpu, v)
	updateFlagsNz(v, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execOra(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	v := modeOpsTable[mode].getValue(cpu)
	cpu.a |= v
	updateFlagsNz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execRol(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	oldCarry := cpu.p & P_C
	u := modeOpsTable[mode].getValue(cpu)
	v := u << 1
	if oldCarry != 0 {
		v |= 0x01
	}
	cpu.p &= ^P_C
	if u&0x80 != 0 {
		cpu.p |= P_C
	}
	cpu.p &= ^P_N
	if v&0x80 != 0 {
		cpu.p |= P_N
	}
	modeOpsTable[mode].setValue(cpu, v)
	cpu.p &= ^P_Z
	if cpu.a == 0 {
		cpu.p |= P_Z
	}
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execRor(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	u := modeOpsTable[mode].getValue(cpu)
	v := u >> 1
	if cpu.p&P_C != 0 {
		v |= 0x80
	}
	cpu.p &= ^(P_C | P_N | P_Z)
	if u&0x01 != 0 {
		cpu.p |= P_C
	}
	if v&0x80 != 0 {
		cpu.p |= P_N
	}
	modeOpsTable[mode].setValue(cpu, v)
	if cpu.a == 0 {
		cpu.p |= P_Z
	}
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execSbc(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	aPrev := cpu.a
	m := modeOpsTable[mode].getValue(cpu)
	v := m + (1 - cpu.p&P_C)
	cpu.a = cpu.a - v
	u := int(int8(aPrev)) - int(int8(m)) - int(1-cpu.p&P_C)
	cpu.p &= ^(P_C | P_V)
	if aPrev >= v {
		cpu.p |= P_C
	}
	if u < -128 || u > 127 {
		cpu.p |= P_V
	}
	updateFlagsNz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execPha(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.push8(cpu.a)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execPhp(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.push8(cpu.p)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execPla(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.a = cpu.pop8()
	updateFlagsNz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execPlp(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p = cpu.pop8() | P_R
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execJsr(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.push16(uint16(cpu.pc + 2))
	cpu.pc = getAddrAbs(cpu)
	return instTable[opc].cycle
}

func execRts(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.pc = cpu.pop16()
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execRti(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p = cpu.pop8() | P_R
	cpu.pc = cpu.pop16()
	return instTable[opc].cycle
}

func execBranchGen(cpu *Cpu, opc uint8, mode InstMode, bytes uint, v bool) uint {
	var extracycle uint = 0
	if v {
		extracycle = 1
		nextpc := cpu.pc + uint16(bytes)
		cpu.pc = uint16(int(nextpc) + int(int8(getValueImm(cpu))))
		if nextpc>>8 != cpu.pc>>8 {
			extracycle = 2
		}
	} else {
		cpu.pc += uint16(bytes)
	}
	return instTable[opc].cycle + extracycle
}

func execBmi(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return execBranchGen(cpu, opc, mode, bytes, cpu.p&P_N != 0)
}

func execBcc(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return execBranchGen(cpu, opc, mode, bytes, cpu.p&P_C == 0)
}

func execBcs(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return execBranchGen(cpu, opc, mode, bytes, cpu.p&P_C != 0)
}

func execBeq(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return execBranchGen(cpu, opc, mode, bytes, cpu.p&P_Z != 0)
}

func execBne(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return execBranchGen(cpu, opc, mode, bytes, cpu.p&P_Z == 0)
}

func execBpl(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return execBranchGen(cpu, opc, mode, bytes, cpu.p&P_N == 0)
}

func execBvc(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return execBranchGen(cpu, opc, mode, bytes, cpu.p&P_V == 0)
}

func execBvs(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return execBranchGen(cpu, opc, mode, bytes, cpu.p&P_V != 0)
}

func execClc(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p &= ^P_C
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execCli(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p &= ^P_I
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execClv(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p &= ^P_V
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execSec(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p |= P_C
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execSei(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p |= P_I
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func execBrk(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.push16(cpu.pc)
	cpu.push8(cpu.p)
	cpu.pc = cpu.mem.Read16(VEC_IRQ)
	cpu.p |= P_B
	return instTable[opc].cycle
}

func execNop(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func (cpu *Cpu) setNmi() {
	cpu.nmiLatched = true
}

func (cpu *Cpu) executeInst() uint {
	if cpu.nmiLatched {
		//Debug("NMI latched\n")
		cpu.nmiLatched = false
		cpu.push16(cpu.pc)
		cpu.push8(cpu.p)
		cpu.pc = cpu.mem.Read16(VEC_NMI)
	}

	opc := cpu.mem.Read8NoTrace(cpu.pc)
	mode := instTable[opc].mode
	bytes := instTable[opc].bytes
	if cpu.nes.dbg.trace {
		Debug("%04X: %s %-10s    opc=%02Xh A:%02X X:%02X Y:%02X P:%02X SP:%02X\n",
			cpu.pc, instTable[opc].mnemonic,
			modeOpsTable[mode].getOpdString(cpu.mem, cpu.pc),
			opc, cpu.a, cpu.x, cpu.y, cpu.p, cpu.s)
	}
	cycle := instHandlerTable[opc](cpu, opc, mode, bytes)

	return cycle
}

func GetAsmStr(mem Memory, pc uint16) (error, int, string) {
	opc := mem.Read8NoTrace(pc)
	_, ok := instTable[opc]
	if !ok {
		return errors.New("invalid opcode"), 0, "invalid opcode"
	}

	mode := instTable[opc].mode
	s := fmt.Sprintf("%04X: %s %-10s",
		pc,
		instTable[opc].mnemonic,
		modeOpsTable[mode].getOpdString(mem, pc))
	return nil, int(instTable[opc].bytes), s
}

func NewCpu(nes *Nes) *Cpu {
	cpu := new(Cpu)
	cpu.nes = nes
	cpu.mem = nes.mem
	return cpu
}
