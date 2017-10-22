package nespkg

import (
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

type InstHandler func(cpu *Cpu, mode InstMode, bytes uint)

type InstParams struct {
	mnemonic string
	mode     InstMode
	handler  InstHandler
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
	getOpdString func(cpu *Cpu, pc uint16) string
}

var modeOpsTable = map[InstMode]ModeOps{
	abs: {get_addr_abs, get_value_abs, set_value_abs, get_opdstr_abs},
	abx: {get_addr_abx, get_value_abx, set_value_abx, get_opdstr_abx},
	aby: {get_addr_aby, get_value_aby, set_value_aby, get_opdstr_aby},
	acc: {nil, get_value_acc, set_value_acc, get_opdstr_acc},
	imm: {nil, get_value_imm, nil, get_opdstr_imm},
	imp: {nil, nil, nil, get_opdstr_imp},
	ind: {get_addr_ind, nil, nil, get_opdstr_ind},
	inx: {get_addr_inx, get_value_inx, set_value_inx, get_opdstr_inx},
	iny: {get_addr_iny, get_value_iny, set_value_iny, get_opdstr_iny},
	rel: {nil, get_value_imm, nil, get_opdstr_rel},
	zrp: {get_addr_zrp, get_value_zrp, set_value_zrp, get_opdstr_zrp},
	zpx: {get_addr_zpx, get_value_zpx, set_value_zpx, get_opdstr_zpx},
	zpy: {get_addr_zpy, get_value_zpy, set_value_zpy, get_opdstr_zpy},
}

var instTable = map[uint8]InstParams{
	0x00: {"brk", imp, exec_brk, 1, 7},
	0x01: {"ora", inx, exec_ora, 2, 6},
	0x05: {"ora", zrp, exec_ora, 2, 3},
	0x06: {"asl", zrp, exec_asl, 2, 5},
	0x08: {"php", imp, exec_php, 1, 3},
	0x09: {"ora", imm, exec_ora, 2, 2},
	0x0A: {"asl", imp, exec_asl, 1, 2},
	0x0D: {"ora", abs, exec_ora, 3, 4},
	0x0E: {"asl", abs, exec_asl, 3, 6},
	0x10: {"bpl", rel, exec_bpl, 2, 2},
	0x11: {"ora", iny, exec_ora, 2, 5},
	0x15: {"ora", zpx, exec_ora, 2, 4},
	0x16: {"asl", zpx, exec_asl, 2, 6},
	0x18: {"clc", imp, exec_clc, 1, 2},
	0x19: {"ora", aby, exec_ora, 3, 4},
	0x1D: {"ora", abx, exec_ora, 3, 4},
	0x1E: {"asl", abx, exec_asl, 3, 7},
	0x20: {"jsr", abs, exec_jsr, 3, 6},
	0x21: {"and", inx, exec_and, 2, 6},
	0x24: {"bit", zrp, exec_bit, 2, 3},
	0x25: {"and", zrp, exec_and, 2, 3},
	0x26: {"rol", zrp, exec_rol, 2, 5},
	0x28: {"plp", imp, exec_plp, 1, 4},
	0x29: {"and", imm, exec_and, 2, 2},
	0x2A: {"rol", acc, exec_rol, 1, 2},
	0x2C: {"bit", abs, exec_bit, 3, 4},
	0x2D: {"and", abs, exec_and, 3, 4},
	0x2E: {"rol", abs, exec_rol, 3, 6},
	0x30: {"bmi", rel, exec_bmi, 2, 2},
	0x31: {"and", iny, exec_and, 2, 5},
	0x35: {"and", zpx, exec_and, 2, 4},
	0x36: {"rol", zpx, exec_rol, 2, 6},
	0x38: {"sec", imp, exec_sec, 1, 2},
	0x39: {"and", aby, exec_and, 3, 4},
	0x3D: {"and", abx, exec_and, 3, 4},
	0x3E: {"rol", abx, exec_rol, 3, 7},
	0x40: {"rti", imp, exec_rti, 1, 6},
	0x41: {"eor", inx, exec_eor, 2, 6},
	0x45: {"eor", zrp, exec_eor, 2, 3},
	0x46: {"lsr", zrp, exec_lsr, 2, 5},
	0x48: {"pha", imp, exec_pha, 1, 3},
	0x49: {"eor", imm, exec_eor, 2, 2},
	0x4A: {"lsr", acc, exec_lsr, 1, 2},
	0x4C: {"jmp", abs, exec_jmp, 3, 3},
	0x4D: {"eor", abs, exec_eor, 3, 4},
	0x4E: {"lsr", abs, exec_lsr, 3, 6},
	0x50: {"bvc", rel, exec_bvc, 2, 2},
	0x51: {"eor", iny, exec_eor, 2, 5},
	0x55: {"eor", zpx, exec_eor, 2, 4},
	0x56: {"lsr", zpx, exec_lsr, 2, 6},
	0x58: {"cli", imp, exec_cli, 1, 2},
	0x59: {"eor", aby, exec_eor, 3, 4},
	0x5D: {"eor", abx, exec_eor, 3, 4},
	0x5E: {"lsr", abx, exec_lsr, 3, 7},
	0x60: {"rts", imp, exec_rts, 1, 6},
	0x61: {"adc", inx, exec_adc, 2, 6},
	0x65: {"adc", zrp, exec_adc, 2, 3},
	0x66: {"ror", zrp, exec_ror, 2, 5},
	0x68: {"pla", imp, exec_pla, 1, 4},
	0x69: {"adc", imm, exec_adc, 2, 2},
	0x6A: {"ror", acc, exec_ror, 1, 2},
	0x6C: {"jmp", ind, exec_jmp, 3, 5},
	0x6D: {"adc", abs, exec_adc, 3, 4},
	0x6E: {"ror", abs, exec_ror, 3, 6},
	0x70: {"bvs", rel, exec_bvs, 2, 2},
	0x71: {"adc", iny, exec_adc, 2, 5},
	0x75: {"adc", zpx, exec_adc, 2, 4},
	0x76: {"ror", zpx, exec_ror, 2, 6},
	0x78: {"sei", imp, exec_sei, 1, 2},
	0x79: {"adc", aby, exec_adc, 3, 4},
	0x7D: {"adc", abx, exec_adc, 3, 4},
	0x7E: {"ror", abx, exec_ror, 3, 7},
	0x81: {"sta", inx, exec_sta, 2, 6},
	0x84: {"sty", zrp, exec_sty, 2, 3},
	0x85: {"sta", zrp, exec_sta, 2, 3},
	0x86: {"stx", zrp, exec_stx, 2, 3},
	0x88: {"dey", imp, exec_dey, 1, 2},
	0x8A: {"txa", imp, exec_txa, 1, 2},
	0x8C: {"sty", abs, exec_sty, 3, 4},
	0x8D: {"sta", abs, exec_sta, 3, 4},
	0x8E: {"stx", abs, exec_stx, 3, 4},
	0x90: {"bcc", rel, exec_bcc, 2, 2},
	0x91: {"sta", iny, exec_sta, 2, 6},
	0x94: {"sty", zpx, exec_sty, 2, 4},
	0x95: {"sta", zpx, exec_sta, 2, 4},
	0x96: {"stx", zpy, exec_stx, 2, 4},
	0x98: {"tya", imp, exec_tya, 1, 2},
	0x99: {"sta", aby, exec_sta, 3, 5},
	0x9A: {"txs", imp, exec_txs, 1, 2},
	0x9D: {"sta", abx, exec_sta, 3, 5},
	0xA0: {"ldy", imm, exec_ldy, 2, 2},
	0xA1: {"lda", inx, exec_lda, 2, 6},
	0xA2: {"ldx", imm, exec_ldx, 2, 2},
	0xA4: {"ldy", zrp, exec_ldy, 2, 3},
	0xA5: {"lda", zrp, exec_lda, 2, 3},
	0xA6: {"ldx", zrp, exec_ldx, 2, 3},
	0xA8: {"tay", imp, exec_tay, 1, 2},
	0xA9: {"lda", imm, exec_lda, 2, 2},
	0xAA: {"tax", imp, exec_tax, 1, 2},
	0xAC: {"ldy", abs, exec_ldy, 3, 4},
	0xAD: {"lda", abs, exec_lda, 3, 4},
	0xAE: {"ldx", abs, exec_ldx, 3, 4},
	0xB0: {"bcs", rel, exec_bcs, 2, 2},
	0xB1: {"lda", iny, exec_lda, 2, 5},
	0xB4: {"ldy", zpx, exec_ldy, 2, 4},
	0xB5: {"lda", zpx, exec_lda, 2, 4},
	0xB6: {"ldx", zpy, exec_ldx, 2, 4},
	0xB8: {"clv", imp, exec_clv, 1, 2},
	0xB9: {"lda", aby, exec_lda, 3, 4},
	0xBA: {"tsx", imp, exec_tsx, 1, 2},
	0xBC: {"ldy", abx, exec_ldy, 3, 4},
	0xBD: {"lda", abx, exec_lda, 3, 4},
	0xBE: {"ldx", aby, exec_ldx, 3, 4},
	0xC0: {"cpy", imm, exec_cpy, 2, 2},
	0xC1: {"cmp", inx, exec_cmp, 2, 6},
	0xC4: {"cpy", zrp, exec_cpy, 2, 3},
	0xC5: {"cmp", zrp, exec_cmp, 2, 3},
	0xC6: {"dec", zrp, exec_dec, 2, 5},
	0xC8: {"iny", imp, exec_iny, 1, 2},
	0xC9: {"cmp", imm, exec_cmp, 2, 2},
	0xCA: {"dex", imp, exec_dex, 1, 2},
	0xCC: {"cpy", abs, exec_cpy, 3, 4},
	0xCD: {"cmp", abs, exec_cmp, 3, 4},
	0xCE: {"dec", abs, exec_dec, 3, 6},
	0xD0: {"bne", rel, exec_bne, 2, 2},
	0xD1: {"cmp", iny, exec_cmp, 2, 5},
	0xD5: {"cmp", zpx, exec_cmp, 2, 4},
	0xD6: {"dec", zpx, exec_dec, 2, 6},
	0xD8: {"cld", imp, exec_cld, 1, 2},
	0xD9: {"cmp", aby, exec_cmp, 3, 4},
	0xDD: {"cmp", abx, exec_cmp, 3, 4},
	0xDE: {"dec", abx, exec_dec, 3, 7},
	0xE0: {"cpx", imm, exec_cpx, 2, 2},
	0xE1: {"sbc", inx, exec_sbc, 2, 6},
	0xE4: {"cpx", zrp, exec_cpx, 2, 3},
	0xE5: {"sbc", zrp, exec_sbc, 2, 3},
	0xE6: {"inc", zrp, exec_inc, 2, 5},
	0xE8: {"inx", imp, exec_inx, 1, 2},
	0xE9: {"sbc", imm, exec_sbc, 2, 2},
	0xEA: {"nop", imp, exec_nop, 1, 2},
	0xEC: {"cpx", abs, exec_cpx, 3, 4},
	0xED: {"sbc", abs, exec_sbc, 3, 4},
	0xEE: {"inc", abs, exec_inc, 3, 6},
	0xF0: {"beq", rel, exec_beq, 2, 2},
	0xF1: {"sbc", iny, exec_sbc, 2, 5},
	0xF5: {"sbc", zpx, exec_sbc, 2, 4},
	0xF6: {"inc", zpx, exec_inc, 2, 6},
	0xF8: {"sed", imp, exec_sed, 1, 2},
	0xF9: {"sbc", aby, exec_sbc, 3, 4},
	0xFD: {"sbc", abx, exec_sbc, 3, 4},
	0xFE: {"inc", abx, exec_inc, 3, 7},
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

func (c *Cpu) Reset() {
	c.a = 0
	c.x = 0
	c.y = 0
	c.s = 0xfa
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
	v := cpu.mem.Read8(stackBase + uint16(cpu.s))
	return v
}

func (cpu *Cpu) push16(v uint16) {
	cpu.mem.Write16(stackBase+uint16(cpu.s), v)
	cpu.s -= 2
}

func (cpu *Cpu) pop16() uint16 {
	cpu.s += 2
	v := cpu.mem.Read16(stackBase + uint16(cpu.s))
	return v
}

func get_value_acc(cpu *Cpu) uint8 {
	return cpu.a
}

func set_value_acc(cpu *Cpu, v uint8) {
	cpu.a = v
}

func get_opdstr_acc(cpu *Cpu, pc uint16) string {
	return "A"
}

func get_value_imm(cpu *Cpu) uint8 {
	return cpu.mem.Read8(cpu.pc + 1)
}

func get_opdstr_imm(cpu *Cpu, pc uint16) string {
	return fmt.Sprintf("#$%02X", cpu.mem.Read8(pc+1))
}

func get_opdstr_imp(cpu *Cpu, pc uint16) string {
	return ""
}

func get_addr_zrp(cpu *Cpu) uint16 {
	return uint16(get_value_imm(cpu))
}

func get_value_zrp(cpu *Cpu) uint8 {
	return cpu.mem.Read8(get_addr_zrp(cpu))
}

func set_value_zrp(cpu *Cpu, v uint8) {
	cpu.mem.Write8(get_addr_zrp(cpu), v)
}

func get_opdstr_zrp(cpu *Cpu, pc uint16) string {
	return fmt.Sprintf("$%02X", cpu.mem.Read8(pc+1))
}

func get_addr_zpx(cpu *Cpu) uint16 {
	return uint16(get_value_imm(cpu) + cpu.x)
}

func get_value_zpx(cpu *Cpu) uint8 {
	return cpu.mem.Read8(get_addr_zpx(cpu))
}

func set_value_zpx(cpu *Cpu, v uint8) {
	cpu.mem.Write8(get_addr_zpx(cpu), v)
}

func get_opdstr_zpx(cpu *Cpu, pc uint16) string {
	return fmt.Sprintf("$%02X,X", cpu.mem.Read8(pc+1))
}

func get_addr_zpy(cpu *Cpu) uint16 {
	return uint16(get_value_imm(cpu) + cpu.y)
}

func get_value_zpy(cpu *Cpu) uint8 {
	return cpu.mem.Read8(get_addr_zpy(cpu))
}

func set_value_zpy(cpu *Cpu, v uint8) {
	cpu.mem.Write8(get_addr_zpy(cpu), v)
}

func get_opdstr_zpy(cpu *Cpu, pc uint16) string {
	return fmt.Sprintf("$%02X,Y", cpu.mem.Read8(pc+1))
}

func get_addr_ind(cpu *Cpu) uint16 {
	a := cpu.mem.Read16(cpu.pc + 1)
	return cpu.mem.Read16(a)
}

func get_opdstr_ind(cpu *Cpu, pc uint16) string {
	return fmt.Sprintf("($%04X)", cpu.mem.Read16(pc+1))
}

func get_addr_abs(cpu *Cpu) uint16 {
	return cpu.mem.Read16(cpu.pc + 1)
}

func get_value_abs(cpu *Cpu) uint8 {
	return cpu.mem.Read8(get_addr_abs(cpu))
}

func set_value_abs(cpu *Cpu, v uint8) {
	cpu.mem.Write8(get_addr_abs(cpu), v)
}

func get_opdstr_abs(cpu *Cpu, pc uint16) string {
	return fmt.Sprintf("$%04X", cpu.mem.Read16(pc+1))
}

func get_addr_abx(cpu *Cpu) uint16 {
	return cpu.mem.Read16(cpu.pc+1) + uint16(cpu.x)
}

func get_value_abx(cpu *Cpu) uint8 {
	return cpu.mem.Read8(get_addr_abx(cpu))
}

func set_value_abx(cpu *Cpu, v uint8) {
	cpu.mem.Write8(get_addr_abx(cpu), v)
}

func get_opdstr_abx(cpu *Cpu, pc uint16) string {
	return fmt.Sprintf("$%04X,x", cpu.mem.Read16(pc+1))
}

func get_addr_aby(cpu *Cpu) uint16 {
	return cpu.mem.Read16(cpu.pc+1) + uint16(cpu.y)
}

func get_value_aby(cpu *Cpu) uint8 {
	return cpu.mem.Read8(get_addr_aby(cpu))
}

func set_value_aby(cpu *Cpu, v uint8) {
	cpu.mem.Write8(get_addr_aby(cpu), v)
}

func get_opdstr_aby(cpu *Cpu, pc uint16) string {
	return fmt.Sprintf("$%04X,y", cpu.mem.Read16(pc+1))
}

func get_addr_inx(cpu *Cpu) uint16 {
	return cpu.mem.Read16(get_addr_zpx(cpu))
}

func get_value_inx(cpu *Cpu) uint8 {
	return cpu.mem.Read8(get_addr_inx(cpu))
}

func set_value_inx(cpu *Cpu, v uint8) {
	cpu.mem.Write8(get_addr_inx(cpu), v)
}

func get_opdstr_inx(cpu *Cpu, pc uint16) string {
	return fmt.Sprintf("($%02X,X)", cpu.mem.Read8(pc+1))
}

func get_addr_iny(cpu *Cpu) uint16 {
	return cpu.mem.Read16(get_addr_zrp(cpu)) + uint16(cpu.y)
}

func get_value_iny(cpu *Cpu) uint8 {
	return cpu.mem.Read8(get_addr_iny(cpu))
}

func set_value_iny(cpu *Cpu, v uint8) {
	cpu.mem.Write8(get_addr_iny(cpu), v)
}

func get_opdstr_iny(cpu *Cpu, pc uint16) string {
	return fmt.Sprintf("($%02X,Y)", cpu.mem.Read8(pc+1))
}

func get_opdstr_rel(cpu *Cpu, pc uint16) string {
	return fmt.Sprintf("$%02X", int8(cpu.mem.Read8(pc+1)))
}

func update_flags_nz(v uint8, cpu *Cpu) {
	cpu.p &= ^(P_N | P_Z)
	if v&0x80 != 0 {
		cpu.p |= P_N
	}
	if v == 0 {
		cpu.p |= P_Z
	}
}

func exec_bmi(cpu *Cpu, mode InstMode, bytes uint) {
	if cpu.p&P_N != 0 {
		cpu.pc += uint16(get_value_imm(cpu))
	} else {
		cpu.pc += uint16(bytes)
	}
}

func exec_lda(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.a = modeOpsTable[mode].getValue(cpu)
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
}

func exec_ldx(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.x = modeOpsTable[mode].getValue(cpu)
	update_flags_nz(cpu.x, cpu)
	cpu.pc += uint16(bytes)
}

func exec_ldy(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.y = modeOpsTable[mode].getValue(cpu)
	update_flags_nz(cpu.y, cpu)
	cpu.pc += uint16(bytes)
}

func exec_jmp(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.pc = modeOpsTable[mode].getAddress(cpu)
}

func exec_sta(cpu *Cpu, mode InstMode, bytes uint) {
	modeOpsTable[mode].setValue(cpu, cpu.a)
	cpu.pc += uint16(bytes)
}

func exec_stx(cpu *Cpu, mode InstMode, bytes uint) {
	modeOpsTable[mode].setValue(cpu, cpu.x)
	cpu.pc += uint16(bytes)
}

func exec_sty(cpu *Cpu, mode InstMode, bytes uint) {
	modeOpsTable[mode].setValue(cpu, cpu.y)
	cpu.pc += uint16(bytes)
}

func exec_tax(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.x = cpu.a
	update_flags_nz(cpu.x, cpu)
	cpu.pc += uint16(bytes)
}

func exec_txa(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.a = cpu.x
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
}

func exec_txs(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.s = cpu.x
	cpu.pc += uint16(bytes)
}

func exec_tsx(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.x = cpu.s
	update_flags_nz(cpu.x, cpu)
	cpu.pc += uint16(bytes)
}

func exec_tya(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.a = cpu.y
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
}

func exec_tay(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.y = cpu.a
	update_flags_nz(cpu.y, cpu)
	cpu.pc += uint16(bytes)
}

func exec_adc(cpu *Cpu, mode InstMode, bytes uint) {
	a_prev := cpu.a
	c := uint16(0)
	if cpu.p&P_C != 0 {
		c = 1
	}
	m := modeOpsTable[mode].getValue(cpu)
	u := uint16(cpu.a) + uint16(m) + c
	cpu.a = uint8(u)

	cpu.p &= ^(P_C | P_V)
	if u >= 0x100 {
		cpu.p |= P_C
	}
	if a_prev&0x80 != cpu.a&0x80 {
		cpu.p |= P_V
	}
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
}

func exec_and(cpu *Cpu, mode InstMode, bytes uint) {
	v := modeOpsTable[mode].getValue(cpu)
	cpu.a &= v
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
}

func exec_asl(cpu *Cpu, mode InstMode, bytes uint) {
	v := modeOpsTable[mode].getValue(cpu)
	if (v & 0x80) != 0 {
		cpu.p |= P_C
	} else {
		cpu.p &= ^P_C
	}
	v <<= 1
	modeOpsTable[mode].setValue(cpu, v)
	update_flags_nz(v, cpu)
	cpu.pc += uint16(bytes)
}

func exec_bit(cpu *Cpu, mode InstMode, bytes uint) {
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
}

func cmp_gen(u uint8, v uint8, cpu *Cpu) {
	cpu.p &= ^(P_C | P_Z | P_N)
	if u >= v {
		cpu.p |= P_C
	}
	if u == v {
		cpu.p |= P_Z
	}
	if (v-u)&0x80 != 0 {
		cpu.p |= P_N
	}
}

func exec_cmp(cpu *Cpu, mode InstMode, bytes uint) {
	v := modeOpsTable[mode].getValue(cpu)
	cmp_gen(cpu.a, v, cpu)
	cpu.pc += uint16(bytes)
}

func exec_cld(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.p &= ^P_D
	cpu.pc += uint16(bytes)
}

func exec_sed(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.p |= P_D
	cpu.pc += uint16(bytes)
}

func exec_cpx(cpu *Cpu, mode InstMode, bytes uint) {
	v := modeOpsTable[mode].getValue(cpu)
	cmp_gen(cpu.x, v, cpu)
	cpu.pc += uint16(bytes)
}

func exec_cpy(cpu *Cpu, mode InstMode, bytes uint) {
	v := modeOpsTable[mode].getValue(cpu)
	cmp_gen(cpu.y, v, cpu)
	cpu.pc += uint16(bytes)
}

func exec_dec(cpu *Cpu, mode InstMode, bytes uint) {
	v := modeOpsTable[mode].getValue(cpu) - 1
	modeOpsTable[mode].setValue(cpu, v)
	update_flags_nz(v, cpu)
	cpu.pc += uint16(bytes)
}

func exec_dex(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.x--
	update_flags_nz(cpu.y, cpu)
	cpu.pc += uint16(bytes)
}

func exec_dey(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.y--
	update_flags_nz(cpu.y, cpu)
	cpu.pc += uint16(bytes)
}

func exec_inc(cpu *Cpu, mode InstMode, bytes uint) {
	v := modeOpsTable[mode].getValue(cpu) + 1
	modeOpsTable[mode].setValue(cpu, v)
	update_flags_nz(v, cpu)
	cpu.pc += uint16(bytes)
}

func exec_inx(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.x++
	update_flags_nz(cpu.x, cpu)
	cpu.pc += uint16(bytes)
}

func exec_iny(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.y++
	update_flags_nz(cpu.y, cpu)
	cpu.pc += uint16(bytes)
}

func exec_eor(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.a |= modeOpsTable[mode].getValue(cpu)
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
}

func exec_lsr(cpu *Cpu, mode InstMode, bytes uint) {
	u := modeOpsTable[mode].getValue(cpu)
	v := u >> 1
	cpu.p &= ^(P_C | P_Z | P_N)
	if u&0x01 != 0 {
		cpu.p |= P_C
	}
	modeOpsTable[mode].setValue(cpu, v)
	update_flags_nz(v, cpu)
	cpu.pc += uint16(bytes)
}

func exec_ora(cpu *Cpu, mode InstMode, bytes uint) {
	v := modeOpsTable[mode].getValue(cpu)
	cpu.a |= v
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
}

func exec_rol(cpu *Cpu, mode InstMode, bytes uint) {
	oldCarry := cpu.p & P_C
	u := modeOpsTable[mode].getValue(cpu)
	v := u << 1
	if oldCarry != 0 {
		v |= 0x01
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
}

func exec_ror(cpu *Cpu, mode InstMode, bytes uint) {
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
}

func exec_sbc(cpu *Cpu, mode InstMode, bytes uint) {
	orig_a := cpu.a
	m := modeOpsTable[mode].getValue(cpu)
	c := uint8(0)
	if cpu.p&P_C != 0 {
		c = 1
	}
	cpu.a = cpu.a - m - c

	cpu.p &= ^(P_C | P_V)
	if orig_a <= cpu.a {
		cpu.p |= P_C
	}
	if orig_a&0x80 != cpu.a&0x80 {
		cpu.p |= P_V
	}

	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
}

func exec_pha(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.push8(cpu.a)
	cpu.pc += uint16(bytes)
}

func exec_php(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.push8(cpu.p)
	cpu.pc += uint16(bytes)
}

func exec_pla(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.a = cpu.pop8()
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
}

func exec_plp(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.p = cpu.pop8()
	cpu.pc += uint16(bytes)
}

func exec_jsr(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.push16(uint16(cpu.pc + 2))
	cpu.pc = get_addr_abs(cpu)
}

func exec_rts(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.pc = cpu.pop16()
	cpu.pc += uint16(bytes)
}

func exec_rti(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.p = cpu.pop8()
	cpu.pc = cpu.pop16()
	cpu.pc += uint16(bytes)
}

func exec_bcc(cpu *Cpu, mode InstMode, bytes uint) {
	if cpu.p&P_C == 0 {
		cpu.pc += uint16(get_value_imm(cpu))
	}
	cpu.pc += uint16(bytes)
}

func exec_bcs(cpu *Cpu, mode InstMode, bytes uint) {
	if cpu.p&P_C != 0 {
		cpu.pc += uint16(get_value_imm(cpu))
	}
	cpu.pc += uint16(bytes)
}

func exec_beq(cpu *Cpu, mode InstMode, bytes uint) {
	if cpu.p&P_Z != 0 {
		cpu.pc += uint16(get_value_imm(cpu))
	}
	cpu.pc += uint16(bytes)
}

func exec_bne(cpu *Cpu, mode InstMode, bytes uint) {
	if cpu.p&P_Z == 0 {
		cpu.pc = uint16(int(cpu.pc) + int(int8(get_value_imm(cpu))))
	}
	cpu.pc += uint16(bytes)
}

func exec_bpl(cpu *Cpu, mode InstMode, bytes uint) {
	if cpu.p&P_N == 0 {
		cpu.pc += uint16(get_value_imm(cpu))
	}
	cpu.pc += uint16(bytes)
}

func exec_bvc(cpu *Cpu, mode InstMode, bytes uint) {
	if cpu.p&P_V == 0 {
		cpu.pc += uint16(get_value_imm(cpu))
	}
	cpu.pc += uint16(bytes)
}

func exec_bvs(cpu *Cpu, mode InstMode, bytes uint) {
	if cpu.p&P_V != 0 {
		cpu.pc += uint16(get_value_imm(cpu))
	}
	cpu.pc += uint16(bytes)
}

func exec_clc(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.p &= ^P_C
	cpu.pc += uint16(bytes)
}

func exec_cli(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.p &= ^P_I
	cpu.pc += uint16(bytes)
}

func exec_clv(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.p &= ^P_V
	cpu.pc += uint16(bytes)
}

func exec_sec(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.p |= P_C
	cpu.pc += uint16(bytes)
}

func exec_sei(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.p |= P_I
	cpu.pc += uint16(bytes)
}

func exec_brk(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.push16(cpu.pc)
	cpu.push8(cpu.p)
	cpu.pc = cpu.mem.Read16(VEC_IRQ)
	cpu.p |= P_B
}

func exec_nop(cpu *Cpu, mode InstMode, bytes uint) {
	cpu.pc += uint16(bytes)
	return
}

func (cpu *Cpu) setNmi() {
	cpu.nmiLatched = true
}

func (cpu *Cpu) executeInst() uint {
	if cpu.nmiLatched {
		cpu.push16(cpu.pc)
		cpu.pc = cpu.mem.Read16(VEC_NMI)
	}

	pc := cpu.pc
	opc := cpu.mem.Read8(cpu.pc)
	mode := instTable[opc].mode
	bytes := instTable[opc].bytes
	instTable[opc].handler(cpu, mode, bytes)
	cycle := instTable[opc].cycle

	Debug("%08X: %s %-10s    opc=%02Xh cycle=%d\n", pc, instTable[opc].mnemonic, modeOpsTable[mode].getOpdString(cpu, pc), opc, cycle)

	return cycle
}

func NewCpu(nes *Nes) *Cpu {
	cpu := new(Cpu)
	cpu.nes = nes
	cpu.mem = nes.mem
	return cpu
}
