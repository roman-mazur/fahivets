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

func lookup32(r1, r2 *byte, sp *uint16) int32 {
	if sp != nil {
		return int32(*sp)
	}
	return int32(*r1)<<8 | int32(*r2)
}

func addDst(m *CPU, reg *byte, v int16, doCarry bool) {
	carry := int16(0)
	if doCarry && m.PSW.C {
		carry = 1
		if v < 0 {
			carry = -1
		}
	}
	v0 := int16(*reg)
	sum := v0 + v + carry
	*reg = byte(sum & 0xFF)
	m.setZSPC(sum)
	if v >= 0 {
		m.setAddA(v0, v, carry)
	} else {
		m.setSubA(v0, -v, carry)
	}
}

func incA(m *CPU, v int16, doCarry bool) { addDst(m, &m.Registers.A, v, doCarry) }

// ACI implements the ACI instruction (Add to Accumulator with Carry).
func ACI(data byte) Instruction {
	return Instruction{
		Name:    fmt.Sprintf("ACI 0x%02x", data),
		Size:    2,
		Execute: func(m *CPU) { incA(m, int16(data), true) },
	}
}

// ADC implements the ADC instruction (Add Register or Memory to Accumulator with Carry).
func ADC(r byte) Instruction {
	return Instruction{
		Name:    fmt.Sprintf("ADC %s", RegisterCode(r)),
		Size:    1,
		Execute: func(m *CPU) { incA(m, lookup8(m.selectOperand(r)), true) },
	}
}

// ADD implements the ADD instruction (Add Register or Memory to Accumulator).
func ADD(r byte) Instruction {
	return Instruction{
		Name:    fmt.Sprintf("ADD %s", RegisterCode(r)),
		Size:    1,
		Execute: func(m *CPU) { incA(m, lookup8(m.selectOperand(r)), false) },
	}
}

// ADI implements the ADI instruction (Add Immediate to Accumulator).
func ADI(data byte) Instruction {
	return Instruction{
		Name:    fmt.Sprintf("ADI 0x%02x", data),
		Size:    2,
		Execute: func(m *CPU) { incA(m, int16(data), false) },
	}
}

func andA(m *CPU, v byte) {
	result := m.Registers.A & v
	m.Registers.A = result
	m.setZSPC(int16(result))
	m.PSW.A = false
}

func orA(m *CPU, v byte) {
	result := m.Registers.A | v
	m.Registers.A = result
	m.setZSPC(int16(result))
	m.PSW.A = false
}

// ANA implements the ANA instruction (AND Register or Memory with Accumulator).
func ANA(r byte) Instruction {
	return Instruction{
		Name:    fmt.Sprintf("ANA %s", RegisterCode(r)),
		Size:    1,
		Execute: func(m *CPU) { andA(m, byte(lookup8(m.selectOperand(r)))) },
	}
}

// ANI implements the ANI instruction (AND Immediate with Accumulator).
func ANI(data byte) Instruction {
	return Instruction{
		Name:    fmt.Sprintf("ANI 0x%02x", data),
		Size:    2,
		Execute: func(m *CPU) { andA(m, data) },
		Encode:  func(out []byte) { out[0], out[1] = 0xE6, data },
	}
}

// CALL implements the CALL instruction (Call subroutine).
func CALL(addr uint16) Instruction {
	return Instruction{
		Name: fmt.Sprintf("CALL 0x%04x", addr),
		Size: 3,
		Execute: func(m *CPU) {
			m.push16(m.PC + 3) // Push return address
			m.PC = addr        // Jump to subroutine
		},
	}
}

// Ccnd implements the conditional CALL instruction.
func Ccnd(cnd ConditionCode, addr uint16) Instruction {
	return Instruction{
		Name: fmt.Sprintf("Ccnd %s 0x%04x", cnd, addr),
		Size: 3,
		Execute: func(m *CPU) {
			if cnd.Check(m) {
				m.push16(m.PC + 3)
				m.PC = addr
			}
		},
		Encode: func(out []byte) {
			out[0], out[1], out[2] = 0xC6|(byte(cnd)<<3), byte(addr&0xFF), byte((addr>>8)&0xFF)
		},
	}
}

// CMA implements the CMA instruction (Complement Accumulator).
func CMA() Instruction {
	return Instruction{
		Name: "CMA",
		Size: 1,
		Execute: func(m *CPU) {
			m.Registers.A = ^m.Registers.A
		},
	}
}

// CMC implements the CMC instruction (Complement Carry).
func CMC() Instruction {
	return Instruction{
		Name: "CMC",
		Size: 1,
		Execute: func(m *CPU) {
			m.PSW.C = !m.PSW.C
		},
	}
}

func cmpA(m *CPU, v int16) {
	d := int16(m.Registers.A) - v
	m.setZSPC(d)
	m.PSW.C = d < 0
	m.setSubA(int16(m.Registers.A), v, 0)
}

// CMP implements the CMP instruction (Compare Register or Memory with Accumulator).
func CMP(r byte) Instruction {
	return Instruction{
		Name:    fmt.Sprintf("CMP %s", RegisterCode(r)),
		Size:    1,
		Execute: func(m *CPU) { cmpA(m, lookup8(m.selectOperand(r))) },
	}
}

// CPI implements the CPI instruction (Compare Immediate with Accumulator).
func CPI(data byte) Instruction {
	return Instruction{
		Name:    fmt.Sprintf("CPI 0x%02x", data),
		Size:    2,
		Execute: func(m *CPU) { cmpA(m, int16(data)) },
		Encode:  func(out []byte) { out[0], out[1] = 0xFE, data },
	}
}

// DAA implements the DAA instruction (Decimal Adjust Accumulator).
func DAA() Instruction {
	return Instruction{
		Name: "DAA",
		Size: 1,
		Execute: func(m *CPU) {
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

func storeDoubleAdd(h, l *byte, v1, v2 int32) int32 {
	result := v1 + v2
	*h = byte((result >> 8) & 0xFF)
	*l = byte(result & 0xFF)
	return result
}

// DAD implements the DAD instruction (Double Add).
func DAD(rp byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("DAD %s", RegisterPairCode(rp)),
		Size: 1,
		Execute: func(m *CPU) {
			v1 := lookup32(m.selectDoubleOperand(2)) // HL registers.
			v2 := lookup32(m.selectDoubleOperand(rp))

			res := storeDoubleAdd(&m.Registers.H, &m.Registers.L, v1, v2)
			m.PSW.C = res > 0xFFFF
		},
	}
}

// DCR implements the DCR instruction (Decrement Register or Memory).
func DCR(r byte) Instruction {
	return Instruction{
		Name:    fmt.Sprintf("DCR %s", RegisterCode(r)),
		Size:    1,
		Execute: func(m *CPU) { addDst(m, ref8(m.selectOperand(r)), -1, false) },
	}
}

// DCX implements the DCX instruction (Decrement Register Pair).
func DCX(rp byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("DCX %s", RegisterPairCode(rp)),
		Size: 1,
		Execute: func(m *CPU) {
			h, l, sp := m.selectDoubleOperand(rp)
			if sp != nil {
				*sp--
				return
			}
			v := uint16(*h)<<8 | uint16(*l)
			v--
			*h = byte((v >> 8) & 0xFF)
			*l = byte(v & 0xFF)
		},
	}
}

// LXI implements the LXI instruction (Load Register Pair Immediate).
func LXI(rp byte, data uint16) Instruction {
	return Instruction{
		Name: fmt.Sprintf("LXI %s 0x%04x", RegisterPairCode(rp), data),
		Size: 3,
		Execute: func(m *CPU) {
			h, l, sp := m.selectDoubleOperand(rp)
			if sp != nil {
				*sp = data
			} else {
				*h = byte((data >> 8) & 0xFF)
				*l = byte(data & 0xFF)
			}
		},
		Encode: func(out []byte) {
			out[0], out[1], out[2] = 0x01|(rp<<4), byte(data&0xFF), byte(data>>8)
		},
	}
}

// POP implements the POP instruction (Pop Data onto Register Pair)
func POP(rp byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("POP %s", RegisterPairCode(rp)),
		Size: 1,
		Execute: func(m *CPU) {
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
		Name: fmt.Sprintf("PUSH %s", RegisterPairCode(rp)),
		Size: 1,
		Execute: func(m *CPU) {
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
		Name: "RAL",
		Size: 1,
		Execute: func(m *CPU) {
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
		Name: "RAR",
		Size: 1,
		Execute: func(m *CPU) {
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
		Name:    "STC",
		Size:    1,
		Execute: func(m *CPU) { m.PSW.C = true },
	}
}

// RLC implements the RLC instruction (Rotate Accumulator Left)
func RLC() Instruction {
	return Instruction{
		Name: "RLC",
		Size: 1,
		Execute: func(m *CPU) {
			carry := m.Registers.A >> 7
			m.Registers.A = (m.Registers.A << 1) | carry
			m.PSW.C = carry == 1
		},
	}
}

// RRC implements the RRC instruction (Rotate Accumulator Right)
func RRC() Instruction {
	return Instruction{
		Name: "RRC",
		Size: 1,
		Execute: func(m *CPU) {
			carry := m.Registers.A & 0x01
			m.Registers.A = (m.Registers.A >> 1) | (carry << 7)
			m.PSW.C = carry == 1
		},
	}
}

// Rcnd implements the conditional return instruction.
func Rcnd(cnd ConditionCode) Instruction {
	return Instruction{
		Name: fmt.Sprintf("Rcnd %s", cnd),
		Size: 1,
		Execute: func(m *CPU) {
			if cnd.Check(m) {
				m.PC = m.pop16()
			}
		},
	}
}

// RET implements the RET instruction (Return from subroutine).
func RET() Instruction {
	return Instruction{
		Name:    "RET",
		Size:    1,
		Execute: func(m *CPU) { m.PC = m.pop16() },
	}
}

// RST implements the RST instruction (Restart).
func RST(n byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("RST %d", n),
		Size: 1,
		Execute: func(m *CPU) {
			m.push16(m.PC + 1)
			m.PC = uint16(n << 3)
		},
	}
}

// SBB implements the SBB instruction (Subtract Register or Memory from Accumulator with Borrow).
func SBB(r byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("SBB %s", RegisterCode(r)),
		Size: 1,
		Execute: func(m *CPU) {
			reg, mem := m.selectOperand(r)
			var val int16
			if mem != nil {
				val = int16(mem[0])
			} else {
				val = int16(*reg)
			}
			addDst(m, &m.Registers.A, -val, true)
		},
	}
}

// SBI implements the SBI instruction (Subtract Immediate from Accumulator with Borrow).
func SBI(data byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("SBI 0x%02x", data),
		Size: 2,
		Execute: func(m *CPU) {
			addDst(m, &m.Registers.A, -int16(data), true)
		},
	}
}

// SHLD implements the SHLD instruction (Store H and L Directly).
func SHLD(addr uint16) Instruction {
	return Instruction{
		Name: fmt.Sprintf("SHLD 0x%04x", addr),
		Size: 3,
		Execute: func(m *CPU) {
			m.Memory[addr] = m.Registers.L
			m.Memory[addr+1] = m.Registers.H
		},
	}
}

// SPHL implements the SPHL instruction (Move HL to SP).
func SPHL() Instruction {
	return Instruction{
		Name: "SPHL",
		Size: 1,
		Execute: func(m *CPU) {
			m.SP = uint16(m.Registers.H)<<8 | uint16(m.Registers.L)
		},
	}
}

// STA implements the STA instruction (Store Accumulator Directly).
func STA(addr uint16) Instruction {
	return Instruction{
		Name: fmt.Sprintf("STA 0x%04x", addr),
		Size: 3,
		Execute: func(m *CPU) {
			m.Memory[addr] = m.Registers.A
		},
	}
}

// STAX implements the STAX instruction (Store Accumulator Indirectly).
func STAX(rp byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("STAX %s", RegisterPairCode(rp)),
		Size: 1,
		Execute: func(m *CPU) {
			h, l, sp := m.selectDoubleOperand(rp)
			if sp != nil {
				panic("STAX with SP")
			}
			m.Memory[uint16(*h)<<8|uint16(*l)] = m.Registers.A
		},
	}
}

// SUB implements the SUB instruction (Subtract Register or Memory from Accumulator).
func SUB(r byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("SUB %s", RegisterCode(r)),
		Size: 1,
		Execute: func(m *CPU) {
			val := lookup8(m.selectOperand(r))
			addDst(m, &m.Registers.A, -val, false)
		},
	}
}

// SUI implements the SUI instruction (Subtract Immediate from Accumulator).
func SUI(data byte) Instruction {
	return Instruction{
		Name:    fmt.Sprintf("SUI 0x%02x", data),
		Size:    2,
		Execute: func(m *CPU) { addDst(m, &m.Registers.A, -int16(data), false) },
	}
}

// XCHG implements the XCHG instruction (Exchange H&L with D&E).
func XCHG() Instruction {
	return Instruction{
		Name: "XCHG",
		Size: 1,
		Execute: func(m *CPU) {
			m.Registers.H, m.Registers.D = m.Registers.D, m.Registers.H
			m.Registers.L, m.Registers.E = m.Registers.E, m.Registers.L
		},
	}
}

// XTHL implements the XTHL instruction (Exchange Top of Stack with H and L).
func XTHL() Instruction {
	return Instruction{
		Name: "XTHL",
		Size: 1,
		Execute: func(m *CPU) {
			top := m.Memory[m.SP]
			next := m.Memory[m.SP+1]
			m.Memory[m.SP], m.Registers.L = m.Registers.L, top
			m.Memory[m.SP+1], m.Registers.H = m.Registers.H, next
		},
	}
}

// XRI implements the XRI instruction (Exclusive OR Immediate with Accumulator).
func XRI(data byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("XRI 0x%02x", data),
		Size: 2,
		Execute: func(m *CPU) {
			m.Registers.A ^= data
			m.setZSPC(int16(m.Registers.A))
			m.PSW.C = false
		},
	}
}

// XRA implements the XRA instruction (Exclusive OR Register or Memory with Accumulator).
func XRA(r byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("XRA %s", RegisterCode(r)),
		Size: 1,
		Execute: func(m *CPU) {
			m.Registers.A ^= byte(lookup8(m.selectOperand(r)))
			m.setZSPC(int16(m.Registers.A))
			m.PSW.C = false
		},
	}
}

// LHLD implements the LHLD instruction (Load H and L Directly).
func LHLD(addr uint16) Instruction {
	return Instruction{
		Name: fmt.Sprintf("LHLD 0x%04x", addr),
		Size: 3,
		Execute: func(m *CPU) {
			m.Registers.L = m.Memory[addr]
			m.Registers.H = m.Memory[addr+1]
		},
		Encode: func(out []byte) {
			out[0], out[1], out[2] = 0x2A, byte(addr&0xFF), byte(addr>>8)
		},
	}
}

// INR implements the INR instruction (Increment Register or Memory).
func INR(r byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("INR %s", RegisterCode(r)),
		Size: 1,
		Execute: func(m *CPU) {
			reg, mem := m.selectOperand(r)
			if mem != nil {
				addDst(m, &mem[0], 1, true)
			} else {
				addDst(m, reg, 1, true)
			}
		},
	}
}

// INX implements the INX instruction (Increment Register Pair).
func INX(rp byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("INX %s", RegisterPairCode(rp)),
		Size: 1,
		Execute: func(m *CPU) {
			h, l, sp := m.selectDoubleOperand(rp)
			if sp != nil {
				*sp++
				return
			}
			storeDoubleAdd(h, l, int32(*h)<<8|int32(*l), 1)
		},
	}
}

// LDA implements the LDA instruction (Load Accumulator Directly).
func LDA(addr uint16) Instruction {
	return Instruction{
		Name:    fmt.Sprintf("LDA 0x%04x", addr),
		Size:    3,
		Execute: func(m *CPU) { m.Registers.A = m.Memory[addr] },
	}
}

// LDAX implements the LDAX instruction (Load Accumulator Indirectly from Register Pair).
func LDAX(rp byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("LDAX %s", RegisterPairCode(rp)),
		Size: 1,
		Execute: func(m *CPU) {
			h, l, sp := m.selectDoubleOperand(rp)
			if sp != nil {
				panic("LDAX with SP")
			}
			m.Registers.A = m.Memory[uint16(*h)<<8|uint16(*l)]
		},
	}
}

// JMP implements the JMP instruction (Jump Unconditionally).
func JMP(addr uint16) Instruction {
	return Instruction{
		Name:    fmt.Sprintf("JMP 0x%04x", addr),
		Size:    3,
		Execute: func(m *CPU) { m.PC = addr },
		Encode:  func(out []byte) { out[0], out[1], out[2] = 0xC3, byte(addr&0xFF), byte(addr>>8) },
	}
}

func JCnd(cnd ConditionCode, addr uint16) Instruction {
	return Instruction{
		Name: fmt.Sprintf("JCnd %s 0x%04x", cnd, addr),
		Size: 3,
		Execute: func(m *CPU) {
			if cnd.Check(m) {
				m.PC = addr
			}
		},
		Encode: func(out []byte) {
			out[0], out[1], out[2] = 0xC2|(byte(cnd)<<3), byte(addr&0xFF), byte(addr>>8)
		},
	}
}

func NOP() Instruction {
	return Instruction{
		Name:    "NOP",
		Size:    1,
		Execute: func(m *CPU) {},
		Encode:  func(out []byte) { out[0] = 0 },
	}
}

// EI implements the EI instruction (Enable Interrupts).
func EI() Instruction {
	return Instruction{
		Name:    "EI",
		Size:    1,
		Execute: func(m *CPU) { m.Interrupts = true },
	}
}

// DI implements the DI instruction (Disable Interrupts).
func DI() Instruction {
	return Instruction{
		Name:    "DI",
		Size:    1,
		Execute: func(m *CPU) { m.Interrupts = false },
	}
}

// HLT implements the HLT instruction (Halt Execution).
func HLT() Instruction {
	return Instruction{
		Name:    "HLT",
		Size:    1,
		Execute: func(m *CPU) { m.PC = 0 },
	}
}

// IN implements the IN instruction (Input from Port to Accumulator).
func IN(port byte) Instruction {
	return Instruction{
		Name:    fmt.Sprintf("IN 0x%02x", port),
		Size:    2,
		Execute: func(m *CPU) { m.Registers.A = m.In[port] },
	}
}

// MOV implements the MOV instruction (Move Data from Source to Destination Register or Memory).
func MOV(dst byte, src byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("MOV %s, %s", RegisterCode(dst), RegisterCode(src)),
		Size: 1,
		Execute: func(m *CPU) {
			srcR, srcMem := m.selectOperand(src)
			var val byte
			if srcMem != nil {
				val = srcMem[0]
			} else {
				val = *srcR
			}
			dstR, dstMem := m.selectOperand(dst)
			if dstMem != nil {
				dstMem[0] = val
			} else {
				*dstR = val
			}
		},
		Encode: func(out []byte) {
			out[0] = 0x60 | dst<<3 | src
		},
	}
}

// MVI implements the MVI instruction (Move Immediate to Register or Memory).
func MVI(dst byte, data byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("MVI %s, 0x%02x", RegisterCode(dst), data),
		Size: 2,
		Execute: func(m *CPU) {
			dstR, dstMem := m.selectOperand(dst)
			if dstMem != nil {
				dstMem[0] = data
			} else {
				*dstR = data
			}
		},
		Encode: func(out []byte) { out[0], out[1] = 0x06|(dst<<3), data },
	}
}

// ORA implements the ORA instruction (Logical OR Register or Memory with Accumulator).
func ORA(r byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("ORA %s", RegisterCode(r)),
		Size: 1,
		Execute: func(m *CPU) {
			r, mem := m.selectOperand(r)
			if mem != nil {
				orA(m, mem[0])
			} else {
				orA(m, *r)
			}
		},
	}
}

// ORI implements the ORI instruction (Logical OR Immediate with Accumulator).
func ORI(data byte) Instruction {
	return Instruction{
		Name:    fmt.Sprintf("ORI 0x%02x", data),
		Size:    2,
		Execute: func(m *CPU) { orA(m, data) },
	}
}

// PCHL implements the PCHL instruction (Load HL into Program Counter).
func PCHL() Instruction {
	return Instruction{
		Name: "PCHL",
		Size: 1,
		Execute: func(m *CPU) {
			m.PC = uint16(m.Registers.H)<<8 | uint16(m.Registers.L)
		},
	}
}

// OUT implements the OUT instruction (Output Accumulator to Port).
func OUT(port byte) Instruction {
	return Instruction{
		Name: fmt.Sprintf("OUT 0x%02x", port),
		Size: 2,
		Execute: func(m *CPU) {
			m.Out[port] = m.Registers.A
		},
	}
}
