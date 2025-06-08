package arch

import (
	"testing"
)

func TestACI(t *testing.T) {
	var m Machine
	m.Registers.A = 2
	m.PSW.C = true

	ACI(1).Execute(&m)

	if m.Registers.A != 4 {
		t.Errorf("ACI: A = %d, want 4", m.Registers.A)
	}
}

func TestADC(t *testing.T) {
	var m Machine
	m.Registers.A = 2
	m.Registers.B = 1
	m.PSW.C = true

	ADC(0).Execute(&m) // r/m = B

	if m.Registers.A != 4 {
		t.Errorf("ADC: A = %d, want 4", m.Registers.A)
	}
}

func TestADD(t *testing.T) {
	var m Machine
	m.Registers.A = 2
	m.Registers.C = 1
	m.PSW.C = true

	ADD(1).Execute(&m) // r/m = C

	if m.Registers.A != 3 {
		t.Errorf("ADD: A = %d, want 3", m.Registers.A)
	}
}

func TestADI(t *testing.T) {
	var m Machine
	m.Registers.A = 42
	m.PSW.C = false

	ADI(3).Execute(&m)

	if m.Registers.A != 45 {
		t.Errorf("ADI: A = %d, want 45", m.Registers.A)
	}
}

func TestANA(t *testing.T) {
	var m Machine
	m.Registers.A = 2
	m.Registers.D = 1

	ANA(2).Execute(&m) // r/m = D

	if m.Registers.A != 0 {
		t.Errorf("ANA: A = %d, want 0", m.Registers.A)
	}

	m.Registers.A = 2
	m.Registers.D = 1

	ANA(2).Execute(&m) // r/m = D

	if m.Registers.A != 0 {
		t.Errorf("ANA: A = %d, want 0", m.Registers.A)
	}

	m.Registers.A = 2
	m.Registers.D = 1
	ANA(2).Execute(&m) // r/m = D

	if m.Registers.A != 0 {
		t.Errorf("ANA: A = %d, want 0", m.Registers.A)
	}
}

func TestANI(t *testing.T) {
	var m Machine
	m.Registers.A = 2

	ANI(1).Execute(&m)

	if m.Registers.A != 0 {
		t.Errorf("ANI: A = %d, want 0", m.Registers.A)
	}
}

func TestCMA(t *testing.T) {
	var m Machine
	m.Registers.A = 0x51

	CMA().Execute(&m)

	if m.Registers.A != 0xAE {
		t.Errorf("CMA: A = %x, want %x", m.Registers.A, 0xAE)
	}
}

func TestCMC(t *testing.T) {
	var m Machine
	m.PSW.C = false

	CMC().Execute(&m)

	if m.PSW.C != true {
		t.Errorf("CMC: fC = %v, want true", m.PSW.C)
	}

	var m2 Machine
	m2.PSW.C = true
	CMC().Execute(&m2)

	if m2.PSW.C != false {
		t.Errorf("CMC: fC = %v, want false", m2.PSW.C)
	}
}

func TestCMP(t *testing.T) {
	var m Machine
	m.Registers.A = 0x0A
	m.Registers.E = 0x05
	m.PSW.C = true

	CMP(3).Execute(&m) // r/m = E

	if m.PSW.C {
		t.Errorf("CMP(A=%x, E=%x): fC = %v, want false", m.Registers.A, m.Registers.E, m.PSW.C)
	}

	var m2 Machine
	m2.Registers.A = 0x02
	m2.Registers.E = 0x05
	m2.PSW.C = false
	t.Log(&m2)

	CMP(3).Execute(&m2) // r/m = E
	t.Log(&m2)

	if !m2.PSW.C {
		t.Errorf("CMP(A=%x, E=%x): fC = %v, want true", m2.Registers.A, m2.Registers.E, m2.PSW.C)
	}

	var m3 Machine
	m3.Registers.A = 0x0A
	m3.Registers.E = 0x0A
	m3.PSW.Z = false

	CMP(3).Execute(&m3) // r/m = E

	if !m3.PSW.Z {
		t.Errorf("CMP(A=%x, E=%x): fZ = %v, want true", m3.Registers.A, m3.Registers.E, m3.PSW.C)
	}

	var m4 Machine
	m4.Registers.A = 0xEB // -0x15
	m4.Registers.E = 0x05
	m4.PSW.C = true

	CMP(3).Execute(&m4) // r/m = E

	if m4.PSW.C {
		t.Errorf("CMP(A=%x, E=%x): fC = %v, want false", m4.Registers.A, m4.Registers.E, m4.PSW.C)
	}
}

func TestCPI(t *testing.T) {
	var m Machine
	m.Registers.A = 0x0A
	m.PSW.C = true

	CPI(0x05).Execute(&m)

	if m.PSW.C != false {
		t.Errorf("CPI: fC = %v, want false", m.PSW.C)
	}
}

func TestDAA(t *testing.T) {
	var m Machine
	m.Registers.A = 0x9B
	DAA().Execute(&m)

	if m.Registers.A != 0x01 {
		t.Errorf("DAA: A = %v, want 1", m.Registers.A)
	}

	if !m.PSW.C {
		t.Errorf("DAA: Carry = %v, want true", m.PSW.C)
	}

	if !m.PSW.A {
		t.Errorf("DAA: Aux Carry = %v, want true", m.PSW.A)
	}
}

func TestDAD(t *testing.T) {
	var m Machine
	m.Registers.D = 1
	m.Registers.E = 0
	m.Registers.H = 1
	m.Registers.L = 1

	DAD(1).Execute(&m) //rp=b01

	if m.Registers.H != 2 {
		t.Errorf("DAD: H = %v, want 2", m.Registers.H)
	}

	if m.Registers.L != 1 {
		t.Errorf("DAD: L = %v, want 1", m.Registers.L)
	}

	var m2 Machine
	m2.Registers.B = 0x33
	m2.Registers.C = 0x9F
	m2.Registers.H = 0xA1
	m2.Registers.L = 0x7B

	DAD(0).Execute(&m2) //rp=b00
	if m2.Registers.H != 0xD5 {
		t.Errorf("DAD: H = %v, want 0xD5", m2.Registers.H)
	}

	if m2.Registers.L != 0x1A {
		t.Errorf("DAD: L = %v, want 0x1A", m2.Registers.L)
	}
}

func TestDCR(t *testing.T) {
	var m Machine
	m.Registers.H = 0x3A
	m.Registers.L = 0x7C
	addr := uint16(m.Registers.H)<<8 | uint16(m.Registers.L)
	m.Memory[addr] = 2
	DCR(6).Execute(&m) // r/m = b110

	if m.Memory[addr] != 1 {
		t.Errorf("DCR: MEM(0x3A7C) = %v, want 1", m.Memory[addr])
	}

	var m2 Machine
	m2.Registers.A = 42
	DCR(7).Execute(&m2) // r/m = b111

	if m2.Registers.A != 41 {
		t.Errorf("DCR: A = %v, want 41", m2.Registers.A)
	}
}

func TestDCX(t *testing.T) {
	var m Machine
	m.Registers.H = 0x98
	m.Registers.L = 0x00

	DCX(2).Execute(&m) // rp = b10

	if m.Registers.H != 0x97 {
		t.Errorf("DCX: H = %v, want 0x97", m.Registers.H)
	}

	if m.Registers.L != 0xFF {
		t.Errorf("DCX: L = %v, want 0xFF", m.Registers.L)
	}
}

func TestPOP(t *testing.T) {
	var m Machine
	m.PSW.Z = true
	m.Registers.A = 1
	m.SP = 0
	m.Memory[m.SP] = 0x02

	POP(0b11).Execute(&m)

	if m.PSW.Z || m.Registers.A != 0 {
		t.Errorf("POP: fZ = %v, A = %d, want to be reset", m.PSW.Z, m.Registers.A)
	}

	var m2 Machine
	m2.Registers.B = 1
	m2.Registers.C = 1
	m2.SP = 0

	POP(0).Execute(&m2)

	if m2.Registers.B != 0 || m2.Registers.C != 0 {
		t.Errorf("POP: B = %v, C = %v, want to be reset", m2.Registers.B, m2.Registers.C)
	}
}

func TestPUSH(t *testing.T) {
	var m Machine
	m.Registers.D = 1
	m.Registers.E = 2

	PUSH(0b01).Execute(&m)

	if m.Memory[m.SP+1] != m.Registers.D {
		t.Errorf("PUSH: stack[SP+1] = %d, want %d", m.Memory[m.SP-1], m.Registers.D)
	}

	if m.Memory[m.SP] != m.Registers.E {
		t.Errorf("PUSH: stack[SP] = %d, want %d", m.Memory[m.SP-2], m.Registers.E)
	}

	m.Registers.A = 42
	m.PSW.C = true
	t.Log("PSW: ", m.PSW)

	PUSH(0b11).Execute(&m) // rp=b11 for A+PSW

	hardcodedExpectedPSW := byte(0b11) // (C=1)

	if m.Memory[m.SP+1] != m.Registers.A {
		t.Errorf("PUSH A+PSW: stack[SP+1] = %d, want %d", m.Memory[m.SP+1], m.Registers.A)
	}

	if m.Memory[m.SP] != hardcodedExpectedPSW {
		t.Errorf("PUSH A+PSW: stack[SP] = %d, want PSW %d", m.Memory[m.SP], hardcodedExpectedPSW)
	}
}

func TestRLC(t *testing.T) {
	var m Machine
	m.Registers.A = 0b10011001

	RLC().Execute(&m)

	if m.Registers.A != 0b00110011 {
		t.Errorf("RLC: A = %08b, want 00110011", m.Registers.A)
	}

	if !m.PSW.C {
		t.Errorf("RLC: Carry = %v, want true", m.PSW.C)
	}
}

func TestRRC(t *testing.T) {
	var m Machine
	m.Registers.A = 0b10011001

	RRC().Execute(&m)

	if m.Registers.A != 0b11001100 {
		t.Errorf("RRC: A = %08b, want 11001100", m.Registers.A)
	}

	if !m.PSW.C {
		t.Errorf("RRC: Carry = %v, want true", m.PSW.C)
	}
}

func TestRAL(t *testing.T) {
	var m Machine
	m.Registers.A = 0b10011001
	m.PSW.C = false

	RAL().Execute(&m)

	if m.Registers.A != 0b00110010 {
		t.Errorf("RAL: A = %08b, want 00110010", m.Registers.A)
	}

	if !m.PSW.C {
		t.Errorf("RAL: Carry = %v, want true", m.PSW.C)
	}
}

func TestRAR(t *testing.T) {
	var m Machine
	m.Registers.A = 0b10011001
	m.PSW.C = true

	RAR().Execute(&m)

	if m.Registers.A != 0b11001100 {
		t.Errorf("RAR: A = %08b, want 11001100", m.Registers.A)
	}

	if m.PSW.C != true {
		t.Errorf("RAR: Carry = %v, want true", m.PSW.C)
	}
}

func TestSTA(t *testing.T) {
	var m Machine
	m.Registers.A = 0x56
	addr := uint16(0x1234)

	STA(addr).Execute(&m)

	if m.Memory[addr] != m.Registers.A {
		t.Errorf("STA: Memory[0x1234] = %X, want %X", m.Memory[addr], m.Registers.A)
	}
}

func TestLDA(t *testing.T) {
	var m Machine
	addr := uint16(0x1234)
	m.Memory[addr] = 0x7F

	LDA(addr).Execute(&m)

	if m.Registers.A != 0x7F {
		t.Errorf("LDA: A = %X, want %X", m.Registers.A, 0x7F)
	}
}

func TestJMP(t *testing.T) {
	var m Machine
	targetAddr := uint16(0x4567)

	JMP(targetAddr).Execute(&m)

	if m.PC != targetAddr {
		t.Errorf("JMP: PC = %X, want PC to be %X", m.PC, targetAddr)
	}
}
