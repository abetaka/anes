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
	0x00: exec_brk,
	0x01: exec_ora,
	0x05: exec_ora,
	0x06: exec_asl,
	0x08: exec_php,
	0x09: exec_ora,
	0x0A: exec_asl,
	0x0D: exec_ora,
	0x0E: exec_asl,
	0x10: exec_bpl,
	0x11: exec_ora,
	0x15: exec_ora,
	0x16: exec_asl,
	0x18: exec_clc,
	0x19: exec_ora,
	0x1D: exec_ora,
	0x1E: exec_asl,
	0x20: exec_jsr,
	0x21: exec_and,
	0x24: exec_bit,
	0x25: exec_and,
	0x26: exec_rol,
	0x28: exec_plp,
	0x29: exec_and,
	0x2A: exec_rol,
	0x2C: exec_bit,
	0x2D: exec_and,
	0x2E: exec_rol,
	0x30: exec_bmi,
	0x31: exec_and,
	0x35: exec_and,
	0x36: exec_rol,
	0x38: exec_sec,
	0x39: exec_and,
	0x3D: exec_and,
	0x3E: exec_rol,
	0x40: exec_rti,
	0x41: exec_eor,
	0x45: exec_eor,
	0x46: exec_lsr,
	0x48: exec_pha,
	0x49: exec_eor,
	0x4A: exec_lsr,
	0x4C: exec_jmp,
	0x4D: exec_eor,
	0x4E: exec_lsr,
	0x50: exec_bvc,
	0x51: exec_eor,
	0x55: exec_eor,
	0x56: exec_lsr,
	0x58: exec_cli,
	0x59: exec_eor,
	0x5D: exec_eor,
	0x5E: exec_lsr,
	0x60: exec_rts,
	0x61: exec_adc,
	0x65: exec_adc,
	0x66: exec_ror,
	0x68: exec_pla,
	0x69: exec_adc,
	0x6A: exec_ror,
	0x6C: exec_jmp,
	0x6D: exec_adc,
	0x6E: exec_ror,
	0x70: exec_bvs,
	0x71: exec_adc,
	0x75: exec_adc,
	0x76: exec_ror,
	0x78: exec_sei,
	0x79: exec_adc,
	0x7D: exec_adc,
	0x7E: exec_ror,
	0x81: exec_sta,
	0x84: exec_sty,
	0x85: exec_sta,
	0x86: exec_stx,
	0x88: exec_dey,
	0x8A: exec_txa,
	0x8C: exec_sty,
	0x8D: exec_sta,
	0x8E: exec_stx,
	0x90: exec_bcc,
	0x91: exec_sta,
	0x94: exec_sty,
	0x95: exec_sta,
	0x96: exec_stx,
	0x98: exec_tya,
	0x99: exec_sta,
	0x9A: exec_txs,
	0x9D: exec_sta,
	0xA0: exec_ldy,
	0xA1: exec_lda,
	0xA2: exec_ldx,
	0xA4: exec_ldy,
	0xA5: exec_lda,
	0xA6: exec_ldx,
	0xA8: exec_tay,
	0xA9: exec_lda,
	0xAA: exec_tax,
	0xAC: exec_ldy,
	0xAD: exec_lda,
	0xAE: exec_ldx,
	0xB0: exec_bcs,
	0xB1: exec_lda,
	0xB4: exec_ldy,
	0xB5: exec_lda,
	0xB6: exec_ldx,
	0xB8: exec_clv,
	0xB9: exec_lda,
	0xBA: exec_tsx,
	0xBC: exec_ldy,
	0xBD: exec_lda,
	0xBE: exec_ldx,
	0xC0: exec_cpy,
	0xC1: exec_cmp,
	0xC4: exec_cpy,
	0xC5: exec_cmp,
	0xC6: exec_dec,
	0xC8: exec_iny,
	0xC9: exec_cmp,
	0xCA: exec_dex,
	0xCC: exec_cpy,
	0xCD: exec_cmp,
	0xCE: exec_dec,
	0xD0: exec_bne,
	0xD1: exec_cmp,
	0xD5: exec_cmp,
	0xD6: exec_dec,
	0xD8: exec_cld,
	0xD9: exec_cmp,
	0xDD: exec_cmp,
	0xDE: exec_dec,
	0xE0: exec_cpx,
	0xE1: exec_sbc,
	0xE4: exec_cpx,
	0xE5: exec_sbc,
	0xE6: exec_inc,
	0xE8: exec_inx,
	0xE9: exec_sbc,
	0xEA: exec_nop,
	0xEC: exec_cpx,
	0xED: exec_sbc,
	0xEE: exec_inc,
	0xF0: exec_beq,
	0xF1: exec_sbc,
	0xF5: exec_sbc,
	0xF6: exec_inc,
	0xF8: exec_sed,
	0xF9: exec_sbc,
	0xFD: exec_sbc,
	0xFE: exec_inc,
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

func exec_lda(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.a = modeOpsTable[mode].getValue(cpu)
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_ldx(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.x = modeOpsTable[mode].getValue(cpu)
	update_flags_nz(cpu.x, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_ldy(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.y = modeOpsTable[mode].getValue(cpu)
	update_flags_nz(cpu.y, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_jmp(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.pc = modeOpsTable[mode].getAddress(cpu)
	return instTable[opc].cycle
}

func exec_sta(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	modeOpsTable[mode].setValue(cpu, cpu.a)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_stx(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	modeOpsTable[mode].setValue(cpu, cpu.x)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_sty(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	modeOpsTable[mode].setValue(cpu, cpu.y)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_tax(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.x = cpu.a
	update_flags_nz(cpu.x, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_txa(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.a = cpu.x
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_txs(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.s = cpu.x
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_tsx(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.x = cpu.s
	update_flags_nz(cpu.x, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_tya(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.a = cpu.y
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_tay(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.y = cpu.a
	update_flags_nz(cpu.y, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_adc(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
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
	return instTable[opc].cycle
}

func exec_and(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	v := modeOpsTable[mode].getValue(cpu)
	cpu.a &= v
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_asl(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
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
	return instTable[opc].cycle
}

func exec_bit(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
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

func exec_cmp(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	v := modeOpsTable[mode].getValue(cpu)
	cmp_gen(cpu.a, v, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_cld(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p &= ^P_D
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_sed(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p |= P_D
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_cpx(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	v := modeOpsTable[mode].getValue(cpu)
	cmp_gen(cpu.x, v, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_cpy(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	v := modeOpsTable[mode].getValue(cpu)
	cmp_gen(cpu.y, v, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_dec(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	v := modeOpsTable[mode].getValue(cpu) - 1
	modeOpsTable[mode].setValue(cpu, v)
	update_flags_nz(v, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_dex(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.x--
	update_flags_nz(cpu.x, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_dey(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.y--
	update_flags_nz(cpu.y, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_inc(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	v := modeOpsTable[mode].getValue(cpu) + 1
	modeOpsTable[mode].setValue(cpu, v)
	update_flags_nz(v, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_inx(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.x++
	update_flags_nz(cpu.x, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_iny(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.y++
	update_flags_nz(cpu.y, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_eor(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.a |= modeOpsTable[mode].getValue(cpu)
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_lsr(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	u := modeOpsTable[mode].getValue(cpu)
	v := u >> 1
	cpu.p &= ^(P_C | P_Z | P_N)
	if u&0x01 != 0 {
		cpu.p |= P_C
	}
	modeOpsTable[mode].setValue(cpu, v)
	update_flags_nz(v, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_ora(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	v := modeOpsTable[mode].getValue(cpu)
	cpu.a |= v
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_rol(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
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
	return instTable[opc].cycle
}

func exec_ror(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
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

func exec_sbc(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
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
	return instTable[opc].cycle
}

func exec_pha(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.push8(cpu.a)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_php(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.push8(cpu.p)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_pla(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.a = cpu.pop8()
	update_flags_nz(cpu.a, cpu)
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_plp(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p = cpu.pop8()
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_jsr(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.push16(uint16(cpu.pc + 2))
	cpu.pc = get_addr_abs(cpu)
	return instTable[opc].cycle
}

func exec_rts(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.pc = cpu.pop16()
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_rti(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p = cpu.pop8()
	cpu.pc = cpu.pop16()
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_branch_gen(cpu *Cpu, opc uint8, mode InstMode, bytes uint, v bool) uint {
	var extracycle uint = 0
	if v {
		extracycle = 1
		nextpc := cpu.pc + uint16(bytes)
		cpu.pc = uint16(int(nextpc) + int(int8(get_value_imm(cpu))))
		if nextpc>>8 != cpu.pc>>8 {
			extracycle = 2
		}
	} else {
		cpu.pc += uint16(bytes)
	}
	return instTable[opc].cycle + extracycle
}

func exec_bmi(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return exec_branch_gen(cpu, opc, mode, bytes, cpu.p&P_N != 0)
}

func exec_bcc(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return exec_branch_gen(cpu, opc, mode, bytes, cpu.p&P_C == 0)
}

func exec_bcs(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return exec_branch_gen(cpu, opc, mode, bytes, cpu.p&P_C != 0)
}

func exec_beq(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return exec_branch_gen(cpu, opc, mode, bytes, cpu.p&P_Z != 0)
}

func exec_bne(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return exec_branch_gen(cpu, opc, mode, bytes, cpu.p&P_Z == 0)
}

func exec_bpl(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return exec_branch_gen(cpu, opc, mode, bytes, cpu.p&P_N == 0)
}

func exec_bvc(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return exec_branch_gen(cpu, opc, mode, bytes, cpu.p&P_V == 0)
}

func exec_bvs(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	return exec_branch_gen(cpu, opc, mode, bytes, cpu.p&P_V != 0)
}

func exec_clc(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p &= ^P_C
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_cli(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p &= ^P_I
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_clv(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p &= ^P_V
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_sec(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p |= P_C
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_sei(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.p |= P_I
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func exec_brk(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.push16(cpu.pc)
	cpu.push8(cpu.p)
	cpu.pc = cpu.mem.Read16(VEC_IRQ)
	cpu.p |= P_B
	return instTable[opc].cycle
}

func exec_nop(cpu *Cpu, opc uint8, mode InstMode, bytes uint) uint {
	cpu.pc += uint16(bytes)
	return instTable[opc].cycle
}

func (cpu *Cpu) setNmi() {
	cpu.nmiLatched = true
}

func (cpu *Cpu) executeInst() uint {
	if cpu.nmiLatched {
		Debug("NMI latched\n")
		cpu.nmiLatched = false
		cpu.push16(cpu.pc - 1)
		cpu.push8(cpu.p)
		cpu.pc = cpu.mem.Read16(VEC_NMI)
	}

	pc := cpu.pc
	opc := cpu.mem.Read8(cpu.pc)
	mode := instTable[opc].mode
	bytes := instTable[opc].bytes
	Debug("%08X: %s %-10s    opc=%02Xh\n", pc, instTable[opc].mnemonic, modeOpsTable[mode].getOpdString(cpu, pc), opc)
	cycle := instHandlerTable[opc](cpu, opc, mode, bytes)

	return cycle
}

func NewCpu(nes *Nes) *Cpu {
	cpu := new(Cpu)
	cpu.nes = nes
	cpu.mem = nes.mem
	return cpu
}
