package arch

import "fmt"

func lookup8(r *byte, mem []byte) int16 {
	if r != nil {
		return int16(*r)
	}
	return int16(mem[0])
}

func ref8(r *byte, mem []byte) *byte {
	if r != nil {
		return r
	}
	return &mem[0]
}

func lookup16(r1, r2 *byte, sp *uint16) int16 {
	if sp != nil {
		return int16(*sp)
	}
	return int16(*r1)<<8 | int16(*r2)
}

func addDst(m *Machine, reg *byte, v int16, doCarry bool) {
	carry := int16(0)
	if doCarry && m.PSW.C {
		carry = 1
	}
	v0 := int16(*reg)
	sum := v0 + v + carry
	*reg = byte(sum & 0xFF)
	m.setZSPC(sum)
	m.setAddA(v0, v, carry)
}

func incA(m *Machine, v int16, doCarry bool) { addDst(m, &m.Registers.A, v, doCarry) }

// ACI implements the ACI instruction (Add to Accumulator with Carry).
func ACI(data byte) Instruction {
	return Instruction{
		Size:    2,
		Execute: func(m *Machine) { incA(m, int16(data), true) },
	}
}

// ADC implements the ADC instruction (Add Register or Memory to Accumulator with Carry).
func ADC(r byte) Instruction {
	return Instruction{
		Size:    1,
		Execute: func(m *Machine) { incA(m, lookup8(m.selectOperand(r)), true) },
	}
}

// ADD implements the ADD instruction (Add Register or Memory to Accumulator).
func ADD(r byte) Instruction {
	return Instruction{
		Size:    1,
		Execute: func(m *Machine) { incA(m, lookup8(m.selectOperand(r)), false) },
	}
}

// ADI implements the ADI instruction (Add Immediate to Accumulator).
func ADI(data byte) Instruction {
	return Instruction{
		Size:    2,
		Execute: func(m *Machine) { incA(m, int16(data), false) },
	}
}

func andA(m *Machine, v byte) {
	result := m.Registers.A & v
	m.Registers.A = result
	m.setZSPC(int16(result))
	m.PSW.A = false
}

// ANA implements the ANA instruction (AND Register or Memory with Accumulator).
func ANA(r byte) Instruction {
	return Instruction{
		Size:    1,
		Execute: func(m *Machine) { andA(m, byte(lookup8(m.selectOperand(r)))) },
	}
}

// ANI implements the ANI instruction (AND Immediate with Accumulator).
func ANI(data byte) Instruction {
	return Instruction{
		Size:    2,
		Execute: func(m *Machine) { andA(m, data) },
	}
}

// CALL implements the CALL instruction (Call subroutine).
func CALL(addr uint16) Instruction {
	return Instruction{
		Size: 3,
		Execute: func(m *Machine) {
			m.push16(m.PC + 3) // Push return address
			m.PC = addr        // Jump to subroutine
		},
	}
}

func condition(cnd byte) func(*Machine) bool {
	switch cnd {
	case 0:
		return func(m *Machine) bool { return !m.PSW.Z }
	case 1:
		return func(m *Machine) bool { return m.PSW.Z }
	case 2:
		return func(m *Machine) bool { return !m.PSW.C }
	case 3:
		return func(m *Machine) bool { return m.PSW.C }
	case 4:
		return func(m *Machine) bool { return !m.PSW.P }
	case 5:
		return func(m *Machine) bool { return m.PSW.P }
	case 6:
		return func(m *Machine) bool { return !m.PSW.S }
	case 7:
		return func(m *Machine) bool { return m.PSW.S }
	default:
		panic(fmt.Errorf("invalid condition: 0x%02x", cnd))
	}
}

// Ccnd implements the conditional CALL instruction.
func Ccnd(cnd byte, addr uint16) Instruction {
	return Instruction{
		Size: 3,
		Execute: func(m *Machine) {
			if condition(cnd)(m) {
				m.push16(m.PC + 3)
				m.PC = addr
			}
		},
	}
}

// CMA implements the CMA instruction (Complement Accumulator).
func CMA() Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			m.Registers.A = ^m.Registers.A
		},
	}
}

// CMC implements the CMC instruction (Complement Carry).
func CMC() Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			m.PSW.C = !m.PSW.C
		},
	}
}

func cmpA(m *Machine, v int16) {
	d := int16(m.Registers.A) - v
	m.setZSPC(d)
	m.PSW.C = d < 0
	m.setSubA(int16(m.Registers.A), v, 0)
}

// CMP implements the CMP instruction (Compare Register or Memory with Accumulator).
func CMP(r byte) Instruction {
	return Instruction{
		Size:    1,
		Execute: func(m *Machine) { cmpA(m, lookup8(m.selectOperand(r))) },
	}
}

// CPI implements the CPI instruction (Compare Immediate with Accumulator).
func CPI(data byte) Instruction {
	return Instruction{
		Size:    2,
		Execute: func(m *Machine) { cmpA(m, int16(data)) },
	}
}

// DAA implements the DAA instruction (Decimal Adjust Accumulator).
func DAA() Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			if (m.Registers.A&0x0F) > 9 || m.PSW.A {
				incA(m, int16(6), false)
			}
			oldA := m.PSW.A
			if ((m.Registers.A>>4)&0x0F) > 9 || m.PSW.C {
				incA(m, int16(0x60), false)
				m.PSW.A = oldA
			}
		},
	}
}

// DAD implements the DAD instruction (Double Add).
func DAD(rp byte) Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			v1 := lookup16(m.selectDoubleOperand(2)) // HL registers.
			v2 := lookup16(m.selectDoubleOperand(rp))

			result := int32(v1) + int32(v2)
			m.Registers.H = byte((result >> 8) & 0xFF)
			m.Registers.L = byte(result & 0xFF)
			m.PSW.C = result > 0xFFFF
		},
	}
}

// DCR implements the DCR instruction (Decrement Register or Memory).
func DCR(r byte) Instruction {
	return Instruction{
		Size:    1,
		Execute: func(m *Machine) { addDst(m, ref8(m.selectOperand(r)), -1, false) },
	}
}

// DCX implements the DCX instruction (Decrement Register Pair).
func DCX(rp byte) Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {

			switch rp {
			case 0: // B and C
				BC := uint16(m.Registers.B)<<8 | uint16(m.Registers.C)
				BC--
				m.Registers.B = byte((BC >> 8) & 0xFF)
				m.Registers.C = byte(BC & 0xFF)
			case 1: // D and E
				DE := uint16(m.Registers.D)<<8 | uint16(m.Registers.E)
				DE--
				m.Registers.D = byte((DE >> 8) & 0xFF)
				m.Registers.E = byte(DE & 0xFF)
			case 2: // H and L
				HL := uint16(m.Registers.H)<<8 | uint16(m.Registers.L)
				HL--
				m.Registers.H = byte((HL >> 8) & 0xFF)
				m.Registers.L = byte(HL & 0xFF)
			case 3: // SP
				m.SP--
			}
		},
	}
}

// LXI implements the LXI instruction (Load Register Pair Immediate).
func LXI(rp byte, data uint16) Instruction {
	return Instruction{
		Size: 3,
		Execute: func(m *Machine) {
			switch rp {
			case 0:
				m.Registers.B = byte((data >> 8) & 0xFF)
				m.Registers.C = byte(data & 0xFF)
			case 1:
				m.Registers.D = byte((data >> 8) & 0xFF)
				m.Registers.E = byte(data & 0xFF)
			case 2:
				m.Registers.H = byte((data >> 8) & 0xFF)
				m.Registers.L = byte(data & 0xFF)
			case 3:
				m.SP = data
			}
		},
	}
}

// POP implements the POP instruction (Pop Data onto Register Pair)
func POP(rp byte) Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			h, l, sp := m.selectDoubleOperand(rp)
			if sp != nil {
				h = &m.Registers.A
			}
			if l == nil {
				m.setPSW(m.Memory[m.SP])
			} else {
				*l = m.Memory[m.SP]
			}
			*h = m.Memory[m.SP+1]
			m.SP += 2
		},
	}
}

// PUSH implements the PUSH instruction (Push Register Pair onto Stack)
func PUSH(rp byte) Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			var (
				h, l *byte
				sp   *uint16
			)
			h, l, sp = m.selectDoubleOperand(rp)
			if sp != nil {
				// Use PSW data.
				h = &m.Registers.A
				psw := m.psw()
				l = &psw
			}

			m.push8(*h)
			m.push8(*l)
		},
	}
}

// RAL implements the RAL instruction (Rotate Accumulator Left through Carry)
func RAL() Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			carry := byte(0)
			if m.PSW.C {
				carry = 1
			}
			m.PSW.C = (m.Registers.A >> 7) == 1
			m.Registers.A = (m.Registers.A << 1) | carry
		},
	}
}

// RAR implements the RAR instruction (Rotate Accumulator Right through Carry)
func RAR() Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			c := byte(0)
			if m.PSW.C {
				c = 0x80
			}
			m.PSW.C = m.Registers.A&1 == 1
			m.Registers.A = (m.Registers.A >> 1) | c
		},
	}
}

// STC implements the STC instruction (Set Carry)
func STC() Instruction {
	return Instruction{
		Size:    1,
		Execute: func(m *Machine) { m.PSW.C = true },
	}
}

// RLC implements the RLC instruction (Rotate Accumulator Left)
func RLC() Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			carry := m.Registers.A >> 7
			m.Registers.A = (m.Registers.A << 1) | carry
			m.PSW.C = carry == 1
		},
	}
}

// RRC implements the RRC instruction (Rotate Accumulator Right)
func RRC() Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			carry := m.Registers.A & 0x01
			m.Registers.A = (m.Registers.A >> 1) | (carry << 7)
			m.PSW.C = carry == 1
		},
	}
}

// Rcnd implements the conditional return instruction.
func Rcnd(cnd func(*Machine) bool) Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			if cnd(m) {
				m.PC = m.pop16()
			}
		},
	}
}

// RET implements the RET instruction (Return from subroutine).
func RET() Instruction {
	return Instruction{
		Size:    1,
		Execute: func(m *Machine) { m.PC = m.pop16() },
	}
}

// RST implements the RST instruction (Restart).
func RST(n byte) Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			m.push16(m.PC + 1)
			m.PC = uint16(n << 3)
		},
	}
}

// SBB implements the SBB instruction (Subtract Register or Memory from Accumulator with Borrow).
func SBB(r byte) Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			var operand byte
			switch r {
			case 0:
				operand = m.Registers.B
			case 1:
				operand = m.Registers.C
			case 2:
				operand = m.Registers.D
			case 3:
				operand = m.Registers.E
			case 4:
				operand = m.Registers.H
			case 5:
				operand = m.Registers.L
			case 6:
				addr := uint16(m.Registers.H)<<8 | uint16(m.Registers.L)
				operand = m.Memory[addr]
			case 7:
				operand = m.Registers.A
			}

			borrow := int16(0)
			if m.PSW.C {
				borrow = 1
			}

			sum := uint16(m.Registers.A) - uint16(operand) - uint16(borrow)
			m.setSubA(int16(m.Registers.A), int16(operand), borrow)
			m.Registers.A = byte(sum & 0xFF)
			m.setZSPC(int16(sum & 0xFF))
			m.PSW.C = sum > 0xFF
		},
	}
}

// SBI implements the SBI instruction (Subtract Immediate from Accumulator with Borrow).
func SBI(data byte) Instruction {
	return Instruction{
		Size: 2,
		Execute: func(m *Machine) {
			borrow := int16(0)
			if m.PSW.C {
				borrow = 1
			}
			sum := uint16(m.Registers.A) - uint16(data) - uint16(borrow)
			m.setSubA(int16(m.Registers.A), int16(data), borrow)
			m.Registers.A = byte(sum & 0xFF)
			m.setZSPC(int16(sum & 0xFF))
			m.PSW.C = sum > 0xFF
		},
	}
}

// SHLD implements the SHLD instruction (Store H and L Directly).
func SHLD(addr uint16) Instruction {
	return Instruction{
		Size: 3,
		Execute: func(m *Machine) {
			m.Memory[addr] = m.Registers.L
			m.Memory[addr+1] = m.Registers.H
		},
	}
}

// SPHL implements the SPHL instruction (Move HL to SP).
func SPHL() Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			m.SP = uint16(m.Registers.H)<<8 | uint16(m.Registers.L)
		},
	}
}

// STA implements the STA instruction (Store Accumulator Directly).
func STA(addr uint16) Instruction {
	return Instruction{
		Size: 3,
		Execute: func(m *Machine) {
			m.Memory[addr] = m.Registers.A
		},
	}
}

// STAX implements the STAX instruction (Store Accumulator Indirectly).
func STAX(r byte) Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			var addr uint16
			switch r {
			case 0: // B and C
				addr = uint16(m.Registers.B)<<8 | uint16(m.Registers.C)
			case 1: // D and E
				addr = uint16(m.Registers.D)<<8 | uint16(m.Registers.E)
			}
			m.Memory[addr] = m.Registers.A
		},
	}
}

// SUB implements the SUB instruction (Subtract Register or Memory from Accumulator).
func SUB(r byte) Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			var operand byte
			switch r {
			case 0:
				operand = m.Registers.B
			case 1:
				operand = m.Registers.C
			case 2:
				operand = m.Registers.D
			case 3:
				operand = m.Registers.E
			case 4:
				operand = m.Registers.H
			case 5:
				operand = m.Registers.L
			case 6:
				addr := uint16(m.Registers.H)<<8 | uint16(m.Registers.L)
				operand = m.Memory[addr]
			case 7:
				operand = m.Registers.A
			}
			sum := uint16(m.Registers.A) - uint16(operand)
			m.setSubA(int16(m.Registers.A), int16(operand), 0)
			m.Registers.A = byte(sum & 0xFF)
			m.setZSPC(int16(sum & 0xFF))
			m.PSW.C = sum > 0xFF
		},
	}
}

// SUI implements the SUI instruction (Subtract Immediate from Accumulator).
func SUI(data byte) Instruction {
	return Instruction{
		Size: 2,
		Execute: func(m *Machine) {
			sum := uint16(m.Registers.A) - uint16(data)
			m.setSubA(int16(m.Registers.A), int16(data), 0)
			m.Registers.A = byte(sum & 0xFF)
			m.setZSPC(int16(sum & 0xFF))
			m.PSW.C = sum > 0xFF
		},
	}
}

// XCHG implements the XCHG instruction (Exchange H&L with D&E).
func XCHG() Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			m.Registers.H, m.Registers.D = m.Registers.D, m.Registers.H
			m.Registers.L, m.Registers.E = m.Registers.E, m.Registers.L
		},
	}
}

// XRA implements the XRA instruction (Exclusive OR Register or Memory with Accumulator).
func XRA(r byte) Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			var operand byte
			switch r {
			case 0:
				operand = m.Registers.B
			case 1:
				operand = m.Registers.C
			case 2:
				operand = m.Registers.D
			case 3:
				operand = m.Registers.E
			case 4:
				operand = m.Registers.H
			case 5:
				operand = m.Registers.L
			case 6:
				addr := uint16(m.Registers.H)<<8 | uint16(m.Registers.L)
				operand = m.Memory[addr]
			case 7:
				operand = m.Registers.A
			}
			m.Registers.A ^= operand
			m.setZSPC(int16(m.Registers.A))
			m.PSW.C = false
		},
	}
}

// LHLD implements the LHLD instruction (Load H and L Directly).
func LHLD(addr uint16) Instruction {
	return Instruction{
		Size: 3,
		Execute: func(m *Machine) {
			m.Registers.L = m.Memory[addr]
			m.Registers.H = m.Memory[addr+1]
		},
	}
}

// INX implements the INX instruction (Increment Register Pair).
func INX(rp byte) Instruction {
	return Instruction{
		Size: 1,
		Execute: func(m *Machine) {
			switch rp {
			case 0: // B and C
				BC := uint16(m.Registers.B)<<8 | uint16(m.Registers.C)
				BC++
				m.Registers.B = byte((BC >> 8) & 0xFF)
				m.Registers.C = byte(BC & 0xFF)
			case 1: // D and E
				DE := uint16(m.Registers.D)<<8 | uint16(m.Registers.E)
				DE++
				m.Registers.D = byte((DE >> 8) & 0xFF)
				m.Registers.E = byte(DE & 0xFF)
			case 2: // H and L
				HL := uint16(m.Registers.H)<<8 | uint16(m.Registers.L)
				HL++
				m.Registers.H = byte((HL >> 8) & 0xFF)
				m.Registers.L = byte(HL & 0xFF)
			case 3: // SP
				m.SP++
			}
		},
	}
}

// LDA implements the LDA instruction (Load Accumulator Directly).
func LDA(addr uint16) Instruction {
	return Instruction{
		Size: 3,
		Execute: func(m *Machine) {
			m.Registers.A = m.Memory[addr]
		},
	}
}

// JMP implements the JMP instruction (Jump Unconditionally).
func JMP(addr uint16) Instruction {
	return Instruction{
		Size: 3,
		Execute: func(m *Machine) {
			m.PC = addr
		},
	}
}

// CZ implements the CZ instruction (Call if Zero Flag is Set).
func CZ(addr uint16) Instruction {
	return Instruction{
		Size: 3,
		Execute: func(m *Machine) {
			if m.PSW.Z {
				m.Memory[m.SP-1] = byte((m.PC >> 8) & 0xFF) // Store high byte of PC
				m.Memory[m.SP-2] = byte(m.PC & 0xFF)        // Store low byte of PC
				m.SP -= 2                                   // Update SP
				m.PC = addr                                 // Jump to the address
			}
		},
	}
}

func NOP() Instruction {
	return Instruction{
		Size:    1,
		Execute: func(m *Machine) {},
	}
}
