package arch

import (
	"fmt"
	"io"
)

type PSW struct {
	Z bool // Zero flag
	S bool // Sign flag
	P bool // Parity flag
	C bool // Carry flag
	A bool // Auxiliary carry flag for decimal arithmetics
}

type Registers struct {
	A, B, C, D, E, H, L byte // 8-bit general-purpose registers
}

type Memory [65536]byte

func (m *Memory) DumpSparse(out io.Writer) error {
	for i := range m {
		if m[i] != 0 {
			_, err := fmt.Fprintf(out, "%04x: 0x%02x\n", i, m[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Memory) Dump(out io.Writer, start, end int) error {
	for i := start; i < end; i++ {
		if (i-start)%16 == 0 {
			_, err := fmt.Fprintf(out, "\n%04x:", i)
			if err != nil {
				return err
			}
		}
		_, err := fmt.Fprintf(out, " %02x", m[i])
		if err != nil {
			return err
		}
		if (i-start)%16 == 15 {
			_, err := fmt.Fprintf(out, "\n")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type Ports [256]byte

type Machine struct {
	Registers Registers

	PSW PSW
	PC  uint16 // Program Counter
	SP  uint16 // Stack Pointer

	Memory     Memory // 64KB memory space
	Interrupts bool
	In, Out    Ports
}

func (m *Machine) String() string {
	return fmt.Sprintf("A: %02x B: %02x C: %02x D: %02x E: %02x H: %02x L: %02x PC: %04x SP: %04x ZSPCA: %v",
		m.Registers.A, m.Registers.B, m.Registers.C, m.Registers.D, m.Registers.E, m.Registers.H, m.Registers.L, m.PC, m.SP, m.PSW)
}

func (m *Machine) Exec(ins Instruction) {
	pc := m.PC
	ins.Execute(m)
	if pc == m.PC {
		m.PC += uint16(ins.Size)
	}
}

func (m *Machine) Step() (Instruction, error) {
	cmd, _, err := DecodeBytes(m.Memory[m.PC:])
	if err != nil {
		return cmd, err
	}
	m.Exec(cmd)
	return cmd, nil
}

func (m *Machine) psw() byte {
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

func (m *Machine) setPSW(v byte) {
	if v&0x02 != 0x02 {
		panic(fmt.Errorf("invalid PSW value %02x", v))
	}
	m.PSW.C = v&1 == 1
	m.PSW.P = v&0x04 == 0x04
	m.PSW.A = v&0x10 == 0x10
	m.PSW.Z = v&0x40 == 0x40
	m.PSW.S = v&0x80 == 0x80
}

func (m *Machine) setAddA(v1, v2, c int16) {
	m.PSW.A = ((v1 & 0x0F) + (v2 & 0x0F) + (c & 0x0F)) > 0x0F
}

func (m *Machine) setSubA(v1, v2, c int16) {
	m.PSW.A = ((v1 & 0x0F) - (v2 & 0x0F) - (c & 0x0F)) < 0
}

func (m *Machine) setZSPC(v int16) {
	m.PSW.Z = v == 0
	m.PSW.S = v&0x0080 == 0x0080
	m.PSW.P = v&1 == 1
	m.PSW.C = v < -0xFF || v > 0xFF
}

func (m *Machine) selectOperand(s byte) (reg *byte, mem []byte) {
	switch s {
	case 7:
		reg = &m.Registers.A
	case 0:
		reg = &m.Registers.B
	case 1:
		reg = &m.Registers.C
	case 2:
		reg = &m.Registers.D
	case 3:
		reg = &m.Registers.E
	case 4:
		reg = &m.Registers.H
	case 5:
		reg = &m.Registers.L
	case 6:
		addr := uint16(m.Registers.H)<<8 | uint16(m.Registers.L)
		if addr >= 0xFFFF {
			panic(fmt.Errorf("memory address out of range %04x", addr))
		}
		mem = m.Memory[addr:]
	default:
		panic(fmt.Errorf("invalid selector %02x", s))
	}
	return
}

func (m *Machine) selectDoubleOperand(s byte) (r1, r2 *byte, sp *uint16) {
	switch s {
	case 0:
		r1 = &m.Registers.B
		r2 = &m.Registers.C
	case 1:
		r1 = &m.Registers.D
		r2 = &m.Registers.E
	case 2:
		r1 = &m.Registers.H
		r2 = &m.Registers.L
	case 3:
		sp = &m.SP
	default:
		panic(fmt.Errorf("invalid double selector %02x", s))
	}
	return
}

func (m *Machine) push8(v byte) {
	m.SP--
	m.Memory[m.SP] = v
}

func (m *Machine) push16(v uint16) {
	m.Memory[m.SP-1] = byte(v >> 8)
	m.Memory[m.SP-2] = byte(v & 0xFF)
	m.SP -= 2
}

func (m *Machine) pop8() byte {
	r := m.Memory[m.SP]
	m.SP++
	return r
}

func (m *Machine) pop16() uint16 {
	r := uint16(m.Memory[m.SP]) | uint16(m.Memory[m.SP+1])<<8
	m.SP += 2
	return r
}

type Instruction struct {
	Name    string
	Size    byte
	Execute func(m *Machine)
}

type RegisterCode byte

func (rc RegisterCode) String() string { return registerNames[rc : rc+1] }

type RegisterPairCode byte

func (rpc RegisterPairCode) String() string { return pairNames[rpc] }

type ConditionCode byte

func (cc ConditionCode) String() string { return conditionNames[cc : cc+1] }

func (cc ConditionCode) Check(m *Machine) bool {
	switch cc {
	case 0:
		return !m.PSW.Z
	case 1:
		return m.PSW.Z
	case 2:
		return !m.PSW.C
	case 3:
		return m.PSW.C
	case 4:
		return !m.PSW.P
	case 5:
		return m.PSW.P
	case 6:
		return !m.PSW.S
	case 7:
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
