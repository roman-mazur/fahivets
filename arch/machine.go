package arch

import (
	"bytes"
	"fmt"
)

type PSW struct {
	Z bool // Zero flag
	S bool // Sign flag
	P bool // Parity flag
	C bool // Carry flag
	A bool // Auxiliary carry flag for decimal arithmetics
}

func (psw *PSW) String() string {
	v := 0
	if psw.Z {
		v |= 1
	}
	if psw.S {
		v |= 2
	}
	if psw.P {
		v |= 4
	}
	if psw.C {
		v |= 8
	}
	if psw.A {
		v |= 0x10
	}
	return fmt.Sprintf("ACPSZ: %05b", v)
}

type Registers struct {
	A, B, C, D, E, H, L byte // 8-bit general-purpose registers
}

func (r *Registers) String() string {
	return fmt.Sprintf("A:%02x B:%02x C:%02x D:%02x E:%02x H:%02x L:%02x",
		r.A, r.B, r.C, r.D, r.E, r.H, r.L)
}

type Ports [256]byte

type CPU struct {
	Registers Registers

	PSW PSW
	PC  uint16 // Program Counter
	SP  uint16 // Stack Pointer

	Memory     Memory // 64KB memory space
	Interrupts bool
	In, Out    Ports
}

func (m *CPU) String() string {
	return fmt.Sprintf("%s %s PC:%04x SP:%04x", &m.Registers, &m.PSW, m.PC, m.SP)
}

func (m *CPU) Exec(ins Instruction) int {
	pc := m.PC
	cycles := ins.Execute(m)
	if pc == m.PC {
		m.PC += uint16(ins.Size)
	}
	return cycles
}

func (m *CPU) Step() (Instruction, int, error) {
	cmd, _, err := DecodeBytes(m.Memory[m.PC:])
	if err != nil {
		return cmd, 0, err
	}
	c := m.Exec(cmd)
	return cmd, c, nil
}

func (m *CPU) psw() byte {
	res := byte(2) // Bit 0 is always 1, bits 3 and 5 are always 0.
	if m.PSW.C {
		res |= 1
	}
	if m.PSW.P {
		res |= 0x04
	}
	if m.PSW.A {
		res |= 0x10
	}
	if m.PSW.Z {
		res |= 0x40
	}
	if m.PSW.S {
		res |= 0x80
	}
	return res
}

func (m *CPU) setPSW(v byte) {
	if v&0x02 != 0x02 {
		panic(fmt.Errorf("invalid PSW value %02x", v))
	}
	m.PSW.C = v&1 == 1
	m.PSW.P = v&0x04 == 0x04
	m.PSW.A = v&0x10 == 0x10
	m.PSW.Z = v&0x40 == 0x40
	m.PSW.S = v&0x80 == 0x80
}

func (m *CPU) setAddA(v1, v2, c int16) {
	m.PSW.A = ((v1 & 0x0F) + (v2 & 0x0F) + (c & 0x0F)) > 0x0F
}

func (m *CPU) setSubA(v1, v2, c int16) {
	m.PSW.A = ((v1 & 0x0F) - (v2 & 0x0F) - (c & 0x0F)) < 0
}

func (m *CPU) setZSPC(v int16) {
	m.PSW.Z = v == 0
	m.PSW.S = v&0x0080 == 0x0080
	m.PSW.P = v&1 == 1
	m.PSW.C = v < -0xFF || v > 0xFF
}

const (
	RegisterSelB = iota
	RegisterSelC
	RegisterSelD
	RegisterSelE
	RegisterSelH
	RegisterSelL
	RegisterSelMemory // (Memory reference through address in H:L)
	RegisterSelA
)

func (m *CPU) selectOperand(s byte) (reg *byte, mem []byte) {
	switch s {
	case RegisterSelA:
		reg = &m.Registers.A
	case RegisterSelB:
		reg = &m.Registers.B
	case RegisterSelC:
		reg = &m.Registers.C
	case RegisterSelD:
		reg = &m.Registers.D
	case RegisterSelE:
		reg = &m.Registers.E
	case RegisterSelH:
		reg = &m.Registers.H
	case RegisterSelL:
		reg = &m.Registers.L
	case RegisterSelMemory:
		addr := uint16(m.Registers.H)<<8 | uint16(m.Registers.L)
		if addr > 0xFFFF {
			panic(fmt.Errorf("memory address out of range %04x", addr))
		}
		mem = m.Memory[addr:]
	default:
		panic(fmt.Errorf("invalid selector %02x", s))
	}
	return
}

const (
	RegisterPairBC = iota // 00=BC   (B:C as 16 bit register)
	RegisterPairDE        // 01=DE   (D:E as 16 bit register)
	RegisterPairHL        // 10=HL   (H:L as 16 bit register)
	RegisterPairSP        // 11=SP   (Stack pointer, refers to PSW (FLAGS:A) for PUSH/POP)
)

func (m *CPU) selectDoubleOperand(s byte) (r1, r2 *byte, sp *uint16) {
	switch s {
	case RegisterPairBC:
		r1 = &m.Registers.B
		r2 = &m.Registers.C
	case RegisterPairDE:
		r1 = &m.Registers.D
		r2 = &m.Registers.E
	case RegisterPairHL:
		r1 = &m.Registers.H
		r2 = &m.Registers.L
	case RegisterPairSP:
		sp = &m.SP
	default:
		panic(fmt.Errorf("invalid double selector %02x", s))
	}
	return
}

func (m *CPU) push8(v byte) {
	m.SP--
	m.Memory[m.SP] = v
}

func (m *CPU) push16(v uint16) {
	m.Memory[m.SP-1] = byte(v >> 8)
	m.Memory[m.SP-2] = byte(v & 0xFF)
	m.SP -= 2
}

func (m *CPU) pop8() byte {
	r := m.Memory[m.SP]
	m.SP++
	return r
}

func (m *CPU) pop16() uint16 {
	r := uint16(m.Memory[m.SP]) | uint16(m.Memory[m.SP+1])<<8
	m.SP += 2
	return r
}

type Instruction struct {
	Name    string
	Size    byte
	Execute func(m *CPU) int
	Encode  func(out []byte)
}

type Program struct {
	Instructions []Instruction
	StartAddress int
}

func (p Program) String() string {
	var (
		out  bytes.Buffer
		addr = p.StartAddress
	)
	for _, cmd := range p.Instructions {
		out.WriteString(fmt.Sprintf("%04x ", addr))
		out.WriteString(cmd.Name)
		out.WriteByte('\n')
		addr += int(cmd.Size)
	}
	return out.String()
}

type RegisterCode byte

func (rc RegisterCode) String() string { return registerNames[rc : rc+1] }

type RegisterPairCode byte

func (rpc RegisterPairCode) String() string { return pairNames[rpc] }

type ConditionCode byte

const (
	ConditionCodeNZ     ConditionCode = iota // 000=NZ 'Z' Z=0
	ConditionCodeZ                           // 001=Z  'z'
	ConditionCodeNC                          // 010=NC 'C'
	ConditionCodeC                           // 011=C  'c'
	ConditionCodeP0                          // 100=P0 'P' P=0
	ConditionCodeP1                          // 101=P1 'p' P=1
	ConditionCodeSPlus                       // 110=+  'S'
	ConditionCodeSMinus                      // 111=-  's'
)

func (cc ConditionCode) String() string { return conditionNames[cc : cc+1] }

func (cc ConditionCode) Check(m *CPU) bool {
	switch cc {
	case ConditionCodeNZ:
		return !m.PSW.Z
	case ConditionCodeZ:
		return m.PSW.Z
	case ConditionCodeNC:
		return !m.PSW.C
	case ConditionCodeC:
		return m.PSW.C
	case ConditionCodeP0:
		return !m.PSW.P
	case ConditionCodeP1:
		return m.PSW.P
	case ConditionCodeSPlus:
		return !m.PSW.S
	case ConditionCodeSMinus:
		return m.PSW.S
	default:
		panic(fmt.Errorf("invalid condition: 0x%02x", cc))
	}
}

const (
	registerNames  = "BCDEHLMA"
	conditionNames = "ZzCcPpSs"
)

var pairNames = [4]string{"BC", "DE", "HL", "SP"}
