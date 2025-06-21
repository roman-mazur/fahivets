package arch

import (
	"reflect"
	"runtime"
	"testing"

	"rmazur.io/fahivets/internal/testutil"
)

func TestDecodeAndExecute(t *testing.T) {
	testCases := []struct {
		name          string
		input         []byte
		initialState  CPU
		expectedState CPU
		expectedSize  int
		expectedError bool
	}{
		{
			name:  "NOP",
			input: []byte{0x00},
			initialState: CPU{
				PC: 0x1000,
			},
			expectedState: CPU{
				PC: 0x1001, // PC should increment by instruction size
			},
			expectedSize: 1,
		},
		{
			name:  "CMC",
			input: []byte{0x3F},
			initialState: CPU{
				PC: 0x1000,
			},
			expectedState: CPU{
				PC: 0x1001,
				PSW: PSW{
					C: true, // Carry flag should be toggled
				},
			},
			expectedSize: 1,
		},
		{
			name:  "STC",
			input: []byte{0x37},
			initialState: CPU{
				PC: 0x1000,
			},
			expectedState: CPU{
				PC: 0x1001,
				PSW: PSW{
					C: true, // Carry flag should be set
				},
			},
			expectedSize: 1,
		},
		{
			name:  "ADC B (A = A + B + Carry, no carry)",
			input: []byte{0x88}, // ADC B
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x01,
					B: 0x02,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					A: 0x03,
					B: 0x02,
				},
				PSW: PSW{P: true},
			},
			expectedSize: 1,
		},
		{
			name:  "ADD B",
			input: []byte{0x80},
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x01,
					B: 0x02,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					A: 0x03, // A = A + B
					B: 0x02,
				},
				PSW: PSW{P: true},
			},
			expectedSize: 1,
		},
		{
			name:  "ADI data",
			input: []byte{0xC6, 0x05},
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x01,
				},
			},
			expectedState: CPU{
				PC: 0x1002,
				Registers: Registers{
					A: 0x06, // A = A + 5
				},
			},
			expectedSize: 2,
		},
		{
			name:  "ACI data",
			input: []byte{0xCE, 0x05},
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x01,
				},
				PSW: PSW{
					C: true, //Carry is set
				},
			},
			expectedState: CPU{
				PC: 0x1002,
				Registers: Registers{
					A: 0x07, // A = A + 5 + Carry
				},
				PSW: PSW{P: true},
			},
			expectedSize: 2,
		},
		{
			name:  "ANA B",
			input: []byte{0xA0},
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x02,
					B: 0x01,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					A: 0x00,
					B: 0x01,
				},
				PSW: PSW{
					Z: true,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "ANI data",
			input: []byte{0xE6, 0x01},
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x02,
				},
			},
			expectedState: CPU{
				PC: 0x1002,
				Registers: Registers{
					A: 0x00,
				},
				PSW: PSW{
					Z: true,
				},
			},
			expectedSize: 2,
		},
		{
			name:  "CMA",
			input: []byte{0x2F},
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x51,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					A: 0xAE,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "CMP E (A > E)",
			input: []byte{0xBB}, // CMP E
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x0A,
					E: 0x05,
				},
				PSW: PSW{C: true}, // Initial Carry flag
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					A: 0x0A,
					E: 0x05,
				},
				PSW: PSW{P: true},
			},
			expectedSize: 1,
		},
		{
			name:  "CMP E (A < E)",
			input: []byte{0xBB}, // CMP E
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x02,
					E: 0x05,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					A: 0x02,
					E: 0x05,
				},
				PSW: PSW{C: true, S: true, P: true, A: true},
			},
			expectedSize: 1,
		},
		{
			name:  "CMP E (A == E)",
			input: []byte{0xBB}, // CMP E
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x0A,
					E: 0x0A,
				},
				PSW: PSW{C: true}, // Initial Carry flag
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					A: 0x0A,
					E: 0x0A,
				},
				PSW: PSW{Z: true}, // Carry flag reset, Zero flag set
			},
			expectedSize: 1,
		},
		{
			name:  "CPI data",
			input: []byte{0xFE, 0x05},
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x0A,
				},
				PSW: PSW{C: true},
			},
			expectedState: CPU{
				PC: 0x1002,
				Registers: Registers{
					A: 0x0A,
				},
				PSW: PSW{P: true},
			},
			expectedSize: 2,
		},
		{
			name:  "INR B",
			input: []byte{0x04},
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					B: 0x01,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					B: 0x02,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "DCR B",
			input: []byte{0x05},
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					B: 0x01,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					B: 0x00,
				},
				PSW: PSW{
					Z: true,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "SUB B",
			input: []byte{0x90},
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x03,
					B: 0x01,
				},
				PSW: PSW{
					C: true,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					A: 0x02,
					B: 0x01,
				},
				PSW: PSW{
					C: false,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "SUI data",
			input: []byte{0xD6, 0x01},
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x03,
				},
				PSW: PSW{
					C: true,
				},
			},
			expectedState: CPU{
				PC: 0x1002,
				Registers: Registers{
					A: 0x02,
				},
				PSW: PSW{
					C: false,
				},
			},
			expectedSize: 2,
		},
		{
			name:  "XRA B",
			input: []byte{0xA8},
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x0A,
					B: 0x0A,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					A: 0x00,
					B: 0x0A,
				},
				PSW: PSW{
					Z: true,
				},
			},
			expectedSize: 1,
		},
		{
			name:          "Unknown instruction",
			input:         []byte{0x08},
			expectedSize:  0,
			expectedError: true,
		},
		{
			name:  "CALL 0x2050",
			input: []byte{0xCD, 0x50, 0x20}, // CALL 0x2050
			initialState: CPU{
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: CPU{
				PC: 0x2050, // PC should point to the called address
				SP: 0x1FFE, // SP should be decremented by 2
				Memory: Memory{
					0x1FFE: 0x03,
					0x1FFF: 0x10,
				},
			},
			expectedSize: 3,
		},
		{
			name:  "CC 0x2050 (Carry Set)",
			input: []byte{0xDC, 0x50, 0x20}, // CC 0x2050
			initialState: CPU{
				PC:  0x1000,
				SP:  0x2000,
				PSW: PSW{C: true}, // Carry flag is set
			},
			expectedState: CPU{
				PC:  0x2050, // PC should point to the called address
				SP:  0x1FFE, // SP should be decremented by 2
				PSW: PSW{C: true},
				Memory: Memory{
					0x1FFE: 0x03, // Low byte of return address (0x1003)
					0x1FFF: 0x10, // High byte of return address (0x1003)
				},
			},
			expectedSize: 3,
		},
		{
			name:  "CC 0x2050 (Carry Not Set)",
			input: []byte{0xDC, 0x50, 0x20}, // CC 0x2050
			initialState: CPU{
				PC: 0x1000,
				SP: 0x2000,
				// Carry flag is not set
			},
			expectedState: CPU{
				PC: 0x1003, // PC should increment to the next instruction
				SP: 0x2000, // SP should not change
				// Memory should not change
			},
			expectedSize: 3,
		},
		{
			name:  "CM 0x2050 (Sign Set)",
			input: []byte{0xFC, 0x50, 0x20}, // CM 0x2050
			initialState: CPU{
				PC:  0x1000,
				SP:  0x2000,
				PSW: PSW{S: true}, // Sign flag is set
			},
			expectedState: CPU{
				PC:  0x2050, // PC should point to the called address
				SP:  0x1FFE, // SP should be decremented by 2
				PSW: PSW{S: true},
				Memory: Memory{
					0x1FFE: 0x03, // Low byte of return address (0x1003)
					0x1FFF: 0x10, // High byte of return address (0x1003)
				},
			},
			expectedSize: 3,
		},
		{
			name:  "CM 0x2050 (Sign Not Set)",
			input: []byte{0xFC, 0x50, 0x20}, // CM 0x2050
			initialState: CPU{
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: CPU{
				PC: 0x1003, // PC should increment to the next instruction
				SP: 0x2000, // SP should not change
			},
			expectedSize: 3,
		},
		{
			name:  "DAA (A = 0x9B)",
			input: []byte{0x27}, // DAA
			initialState: CPU{
				PC:        0x1000,
				Registers: Registers{A: 0x9B},
				PSW:       PSW{S: true},
			},
			expectedState: CPU{
				PC:        0x1001,
				Registers: Registers{A: 0x01},
				PSW:       PSW{C: true, P: true, A: true},
			},
			expectedSize: 1,
		},
		{
			name:  "DAD D (H:L = 0x0101, D:E = 0x0100)",
			input: []byte{0x19}, // DAD D
			initialState: CPU{
				Registers: Registers{
					H: 0x01,
					L: 0x01,
					D: 0x01,
					E: 0x00,
				},
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: CPU{
				Registers: Registers{
					H: 0x02,
					L: 0x01,
					D: 0x01,
					E: 0x00,
				},
				PC: 0x1001,
				SP: 0x2000,
			},
			expectedSize: 1,
		},
		{
			name:  "DAD B (H:L = 0xA17B, B:C = 0x6000, overflow)",
			input: []byte{0x09}, // DAD B
			initialState: CPU{
				Registers: Registers{
					H: 0xA1,
					L: 0x7B,
					B: 0x60,
					C: 0x00,
				},
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: CPU{
				Registers: Registers{
					H: 0x01,
					L: 0x7B,
					B: 0x60,
					C: 0x00,
				},
				PSW: PSW{C: true}, // Carry flag is set
				PC:  0x1001,
				SP:  0x2000,
			},
			expectedSize: 1,
		},
		{
			name:  "DCR M (H:L points to memory location 0x3A7C, MEM(0x3A7C) = 0x02)",
			input: []byte{0x35}, // DCR M
			initialState: CPU{
				Registers: Registers{
					H: 0x3A,
					L: 0x7C,
				},
				PC:     0x1000,
				SP:     0x2000,
				Memory: [65536]byte{0x3A7C: 0x02}, // Memory must be initialized
			},
			expectedState: CPU{
				Registers: Registers{
					H: 0x3A,
					L: 0x7C,
				},
				PC:     0x1001,
				SP:     0x2000,
				PSW:    PSW{P: true},
				Memory: [65536]byte{0x3A7C: 0x01}, // Memory decremented
			},
			expectedSize: 1,
		},
		{
			name:  "DCR A (A = 42)",
			input: []byte{0x3D}, // DCR A
			initialState: CPU{
				Registers: Registers{
					A: 42,
				},
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: CPU{
				Registers: Registers{
					A: 41,
				},
				PC:  0x1001,
				SP:  0x2000,
				PSW: PSW{P: true},
			},
			expectedSize: 1,
		},
		{
			name:  "DCX H (H:L = 0x4000)",
			input: []byte{0x2B}, // DCX H
			initialState: CPU{
				Registers: Registers{
					H: 0x40,
					L: 0x00,
				},
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: CPU{
				Registers: Registers{
					H: 0x3F,
					L: 0xFF,
				},
				PC: 0x1001,
				SP: 0x2000,
			},
			expectedSize: 1,
		},
		{
			name:  "PUSH B (B:C = 0x1234, SP = 0x2000)",
			input: []byte{0xC5}, // PUSH B
			initialState: CPU{
				Registers: Registers{
					B: 0x12,
					C: 0x34,
				},
				SP: 0x2000,
				PC: 0x1000,
			},
			expectedState: CPU{
				Registers: Registers{
					B: 0x12,
					C: 0x34,
				},
				SP:     0x1FFE,
				PC:     0x1001,
				Memory: [65536]byte{0x1FFF: 0x12, 0x1FFE: 0x34},
			},
			expectedSize: 1,
		},
		{
			name:  "POP B (SP = 0x1FFE, MEM(0x1FFE) = 0x56, MEM(0x1FFF) = 0x78)",
			input: []byte{0xC1}, // POP B
			initialState: CPU{
				SP:     0x1FFE,
				PC:     0x1000,
				Memory: [65536]byte{0x1FFE: 0x56, 0x1FFF: 0x78},
			},
			expectedState: CPU{
				Registers: Registers{
					B: 0x78,
					C: 0x56,
				},
				SP:     0x2000,
				PC:     0x1001,
				Memory: [65536]byte{0x1FFE: 0x56, 0x1FFF: 0x78},
			},
			expectedSize: 1,
		},
		{
			name:  "RLC (A = 0x85)",
			input: []byte{0x07}, // RLC
			initialState: CPU{
				Registers: Registers{
					A: 0x85, // 10000101
				},
				PC:  0x1000,
				SP:  0x2000,
				PSW: PSW{C: true},
			},
			expectedState: CPU{
				Registers: Registers{
					A: 0x0B, // 00001011
				},
				PSW: PSW{C: true},
				PC:  0x1001,
				SP:  0x2000,
			},
			expectedSize: 1,
		},
		{
			name:  "RRC (A = 0x81)",
			input: []byte{0x0F}, // RRC
			initialState: CPU{
				Registers: Registers{
					A: 0x81, // 10000001
				},
				PC:  0x1000,
				SP:  0x2000,
				PSW: PSW{C: true},
			},
			expectedState: CPU{
				Registers: Registers{
					A: 0xC0, // 11000000
				},
				PSW: PSW{C: true},
				PC:  0x1001,
				SP:  0x2000,
			},
			expectedSize: 1,
		},
		{
			name:  "RAL (A = 0x55, C = 0)",
			input: []byte{0x17}, // RAL
			initialState: CPU{
				Registers: Registers{
					A: 0x55, // 01010101
				},
				PSW: PSW{C: false},
				PC:  0x1000,
				SP:  0x2000,
			},
			expectedState: CPU{
				Registers: Registers{
					A: 0xAA, // 10101010
				},
				PSW: PSW{C: false},
				PC:  0x1001,
				SP:  0x2000,
			},
			expectedSize: 1,
		},
		{
			name:  "RAR (A = 0x55, C = 1)",
			input: []byte{0x1F}, // RAR
			initialState: CPU{
				Registers: Registers{
					A: 0x55, // 01010101
				},
				PSW: PSW{C: true},
				PC:  0x1000,
				SP:  0x2000,
			},
			expectedState: CPU{
				Registers: Registers{
					A: 0xAA, // 10101010
				},
				PSW: PSW{C: true},
				PC:  0x1001,
				SP:  0x2000,
			},
			expectedSize: 1,
		},
		{
			name:  "RET (SP = 0x1000, MEM(0x1000) = 0x50, MEM(0x1001) = 0x20)",
			input: []byte{0xC9}, // RET
			initialState: CPU{
				SP:     0x1000,
				PC:     0x2000,
				Memory: [65536]byte{0x1000: 0x50, 0x1001: 0x20},
			},
			expectedState: CPU{
				SP:     0x1002,
				PC:     0x2050,
				Memory: [65536]byte{0x1000: 0x50, 0x1001: 0x20},
			},
			expectedSize: 1,
		},
		{
			name:  "RC (C = 1, SP = 0x1000, MEM(0x1000) = 0x50, MEM(0x1001) = 0x20)",
			input: []byte{0xD8}, // RC
			initialState: CPU{
				SP:     0x1000,
				PC:     0x2000,
				PSW:    PSW{C: true},
				Memory: [65536]byte{0x1000: 0x50, 0x1001: 0x20},
			},
			expectedState: CPU{
				SP:     0x1002,
				PC:     0x2050,
				PSW:    PSW{C: true},
				Memory: [65536]byte{0x1000: 0x50, 0x1001: 0x20},
			},
			expectedSize: 1,
		},
		{
			name:  "RC (C = 0, SP = 0x1000, MEM(0x1000) = 0x50, MEM(0x1001) = 0x20)",
			input: []byte{0xD8}, // RC
			initialState: CPU{
				SP:     0x1000,
				PC:     0x2000,
				PSW:    PSW{C: false},
				Memory: [65536]byte{0x1000: 0x50, 0x1001: 0x20},
			},
			expectedState: CPU{
				SP:     0x1000,
				PC:     0x2001,
				PSW:    PSW{C: false},
				Memory: [65536]byte{0x1000: 0x50, 0x1001: 0x20},
			},
			expectedSize: 1,
		},
		{
			name:  "RST 0 (PC = 0x2000, SP = 0x1000)",
			input: []byte{0xC7}, // RST 0
			initialState: CPU{
				SP: 0x1000,
				PC: 0x2000,
			},
			expectedState: CPU{
				SP:     0x0FFE,
				PC:     0x0000,
				Memory: [65536]byte{0x0FFF: 0x20, 0x0FFE: 0x01},
			},
			expectedSize: 1,
		},
		{
			name:  "RST 1 (PC = 0x2000, SP = 0x1000)",
			input: []byte{0xCF}, // RST 1
			initialState: CPU{
				SP: 0x1000,
				PC: 0x2000,
			},
			expectedState: CPU{
				SP:     0x0FFE,
				PC:     0x0008,
				Memory: [65536]byte{0x0FFF: 0x20, 0x0FFE: 0x01},
			},
			expectedSize: 1,
		},
		{
			name:  "SBB B (A = 0x3E, B = 0x2A, C = 0)",
			input: []byte{0x98}, // SBB B
			initialState: CPU{
				Registers: Registers{
					A: 0x3E,
					B: 0x2A,
				},
				PSW: PSW{C: false},
				PC:  0x1000,
				SP:  0x2000,
			},
			expectedState: CPU{
				Registers: Registers{
					A: 0x14,
					B: 0x2A,
				},
				PSW: PSW{C: false, S: false, Z: false, P: false, A: false},
				PC:  0x1001,
				SP:  0x2000,
			},
			expectedSize: 1,
		},
		{
			name:  "SBB B (A = 0x3E, B = 0x3E, C = 0)",
			input: []byte{0x98}, // SBB B
			initialState: CPU{
				Registers: Registers{
					A: 0x3E,
					B: 0x3E,
				},
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: CPU{
				Registers: Registers{
					A: 0x00,
					B: 0x3E,
				},
				PC:  0x1001,
				SP:  0x2000,
				PSW: PSW{Z: true},
			},
			expectedSize: 1,
		},
		{
			name:  "SBB L (A = 0x04, L = 0x02, C = 1)",
			input: []byte{0x9D}, // SBB L
			initialState: CPU{
				Registers: Registers{
					A: 0x04,
					L: 0x02,
				},
				PSW: PSW{C: true},
				PC:  0x1000,
				SP:  0x2000,
			},
			expectedState: CPU{
				Registers: Registers{
					A: 0x01,
					L: 0x02,
				},
				PSW: PSW{A: true, P: true},
				PC:  0x1001,
				SP:  0x2000,
			},
			expectedSize: 1,
		},
		{
			name:  "SBI (A = 0x3E, data = 0x2A, C = 0)",
			input: []byte{0xDE, 0x2A}, // SBI 0x2A
			initialState: CPU{
				Registers: Registers{
					A: 0x3E,
				},
				PSW: PSW{C: false},
				PC:  0x1000,
				SP:  0x2000,
			},
			expectedState: CPU{
				Registers: Registers{
					A: 0x14,
				},
				PSW: PSW{C: false, S: false, Z: false, P: false, A: false},
				PC:  0x1002,
				SP:  0x2000,
			},
			expectedSize: 2,
		},
		{
			name:  "SBI (A = 0x05, data = 0x10, C = 0)",
			input: []byte{0xDE, 0x10}, // SBI 0x10
			initialState: CPU{
				Registers: Registers{
					A: 0x05,
				},
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: CPU{
				Registers: Registers{
					A: 0xF5,
				},
				PSW: PSW{S: true, P: true},
				PC:  0x1002,
				SP:  0x2000,
			},
			expectedSize: 2,
		},
		{
			name:  "SBI (A = 0x05, data = 0x04, C = 1)",
			input: []byte{0xDE, 0x04}, // SBI 0x04
			initialState: CPU{
				Registers: Registers{
					A: 0x05,
				},
				PSW: PSW{C: true},
				PC:  0x1000,
				SP:  0x2000,
			},
			expectedState: CPU{
				// Zero registers.
				PSW: PSW{Z: true, A: true},
				PC:  0x1002,
				SP:  0x2000,
			},
			expectedSize: 2,
		},
		{
			name:  "SHLD",
			input: []byte{0x22, 0x22, 0x11}, // SHLD 0x1122
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					H: 0x02,
					L: 0x03,
				},
			},
			expectedState: CPU{
				PC: 0x1003,
				Registers: Registers{
					H: 0x02,
					L: 0x03,
				},
				Memory: Memory{0x1122: 0x03, 0x1123: 0x02},
			},
			expectedSize: 3,
		},
		{
			name:  "SPHL",
			input: []byte{0xF9},
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					H: 0x02,
					L: 0x03,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				SP: 0x0203,
				Registers: Registers{
					H: 0x02,
					L: 0x03,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "STA",
			input: []byte{0x32, 0x44, 0x33}, // STA 0x3344
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x2A,
				},
			},
			expectedState: CPU{
				PC: 0x1003,
				Registers: Registers{
					A: 0x2A,
				},
				Memory: Memory{0x3344: 0x2A},
			},
			expectedSize: 3,
		},
		{
			name:  "SUB",
			input: []byte{0x90}, // SUB B
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x0A,
					B: 0x05,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					A: 0x05,
					B: 0x05,
				},
				PSW: PSW{
					P: true,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "SUI",
			input: []byte{0xD6, 0x05}, // SUI 0x05
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0x0A,
				},
			},
			expectedState: CPU{
				PC: 0x1002,
				Registers: Registers{
					A: 0x05,
				},
				PSW: PSW{
					P: true,
				},
			},
			expectedSize: 2,
		},
		{
			name:  "XCHG (Exchange H-L with D-E)",
			input: []byte{0xEB}, // XCHG
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					H: 0x12,
					L: 0x34,
					D: 0x56,
					E: 0x78,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					H: 0x56,
					L: 0x78,
					D: 0x12,
					E: 0x34,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "XRA A (A = A ^ A, result is 0)",
			input: []byte{0xAF}, // XRA A
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0xFF,
				},
				PSW: PSW{},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					A: 0x00,
				},
				PSW: PSW{Z: true, P: false},
			},
			expectedSize: 1,
		},
		{
			name:  "XRI",
			input: []byte{0xEE, 0x55}, // XRI 0x55
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0xAA,
				},
				PSW: PSW{},
			},
			expectedState: CPU{
				PC: 0x1002,
				Registers: Registers{
					A: 0xFF,
				},
				PSW: PSW{P: true, S: true},
			},
			expectedSize: 2,
		},
		{
			name:  "XTHL (Exchange H-L with top of stack)",
			input: []byte{0xE3}, // XTHL
			initialState: CPU{
				PC:        0x1000,
				SP:        0x2002,
				Registers: Registers{H: 0x12, L: 0x34},
				Memory:    Memory{0x2002: 0x56, 0x2003: 0x78},
			},
			expectedState: CPU{
				PC:        0x1001,
				SP:        0x2002,
				Registers: Registers{H: 0x78, L: 0x56},
				Memory:    Memory{0x2002: 0x34, 0x2003: 0x12},
			},
			expectedSize: 1,
		},

		{
			name:  "EI (Enable Interrupts)",
			input: []byte{0xFB}, // EI
			initialState: CPU{
				PC: 0x1000,
			},
			expectedState: CPU{
				PC:         0x1001,
				Interrupts: true,
			},
			expectedSize: 1,
		},
		{
			name:  "DI (Disable Interrupts)",
			input: []byte{0xF3}, // DI
			initialState: CPU{
				PC:         0x1000,
				Interrupts: true,
			},
			expectedState: CPU{
				PC:         0x1001,
				Interrupts: false,
			},
			expectedSize: 1,
		},
		{
			name:  "HLT (Halt Execution)",
			input: []byte{0x76}, // HLT
			initialState: CPU{
				PC: 0x1000,
			},
			expectedState: CPU{
				PC: 0,
			},
			expectedSize: 1,
		},
		{
			name:  "IN (Input from port)",
			input: []byte{0xDB, 0x10}, // IN 0x10
			initialState: CPU{
				PC: 0x1000,
				In: Ports{0x10: 0x7F}, // Port 0x10 holds the value 0x7F
			},
			expectedState: CPU{
				PC: 0x1002,
				In: Ports{0x10: 0x7F}, // Port state remains unchanged
				Registers: Registers{
					A: 0x7F, // Value from port 0x10 loaded into the accumulator (A)
				},
			},
			expectedSize: 2,
		},
		{
			name:  "OUT (Output to port)",
			input: []byte{0xD3, 0x10}, // OUT 0x10
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					A: 0xAB,
				},
			},
			expectedState: CPU{
				PC:  0x1002,
				Out: Ports{0x10: 0xAB},
				Registers: Registers{
					A: 0xAB,
				},
			},
			expectedSize: 2,
		},
		{
			name:  "INX H (Increment HL Pair)",
			input: []byte{0x23}, // INX H
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					H: 0x12,
					L: 0x34,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					H: 0x12, // High byte remains 0x12
					L: 0x35, // Low byte incremented by 1
				},
			},
			expectedSize: 1,
		},
		{
			name:  "INX H (Increment HL Pair with Carry)",
			input: []byte{0x23}, // INX H
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					H: 0x12,
					L: 0xFF,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					H: 0x13, // High byte incremented due to carry
					L: 0x00, // Low byte becomes 0x00
				},
			},
			expectedSize: 1,
		},
		{
			name:  "JMP (Unconditional Jump)",
			input: []byte{0xC3, 0x00, 0x20}, // JMP 0x2000
			initialState: CPU{
				PC: 0x1000,
			},
			expectedState: CPU{
				PC: 0x2000, // Program counter jumps to 0x2000
			},
			expectedSize: 3,
		},
		{
			name:  "JC (Jump if Carry - Carry Set)",
			input: []byte{0xDA, 0x00, 0x20}, // JC 0x2000
			initialState: CPU{
				PC:  0x1000,
				PSW: PSW{C: true}, // Carry flag is set
			},
			expectedState: CPU{
				PC:  0x2000,       // Program counter jumps to 0x2000
				PSW: PSW{C: true}, // Carry flag remains unchanged
			},
			expectedSize: 3,
		},
		{
			name:  "JC (Jump if Carry - Carry Not Set)",
			input: []byte{0xDA, 0x00, 0x20}, // JC 0x2000
			initialState: CPU{
				PC:  0x1000,
				PSW: PSW{C: false}, // Carry flag is not set
			},
			expectedState: CPU{
				PC:  0x1003,        // Program counter just increments
				PSW: PSW{C: false}, // Carry flag remains unchanged
			},
			expectedSize: 3,
		},
		{
			name:  "JZ (Jump if Zero - Zero Flag Set)",
			input: []byte{0xCA, 0x00, 0x20}, // JZ 0x2000
			initialState: CPU{
				PC:  0x1000,
				PSW: PSW{Z: true}, // Zero flag is set
			},
			expectedState: CPU{
				PC:  0x2000,       // Program counter jumps to 0x2000
				PSW: PSW{Z: true}, // Zero flag remains unchanged
			},
			expectedSize: 3,
		},
		{
			name:  "JZ (Jump if Zero - Zero Flag Not Set)",
			input: []byte{0xCA, 0x00, 0x20}, // JZ 0x2000
			initialState: CPU{
				PC:  0x1000,
				PSW: PSW{Z: false}, // Zero flag is not set
			},
			expectedState: CPU{
				PC:  0x1003,        // Program counter just increments
				PSW: PSW{Z: false}, // Zero flag remains unchanged
			},
			expectedSize: 3,
		},
		{
			name:  "LDA (Load Accumulator Direct Address)",
			input: []byte{0x3A, 0x10, 0x20}, // LDA 0x2010
			initialState: CPU{
				PC: 0x1000,
				Memory: Memory{
					0x2010: 0xAB,
				},
			},
			expectedState: CPU{
				PC: 0x1003,
				Registers: Registers{
					A: 0xAB,
				},
				Memory: Memory{
					0x2010: 0xAB,
				},
			},
			expectedSize: 3,
		},
		{
			name:  "LDAX (Load Accumulator Indirect Address from BC)",
			input: []byte{0x0A}, // LDAX BC
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					B: 0x20,
					C: 0x10,
				},
				Memory: Memory{
					0x2010: 0xAB,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					A: 0xAB,
					B: 0x20,
					C: 0x10,
				},
				Memory: Memory{
					0x2010: 0xAB,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "LDAX (Load Accumulator Indirect Address from DE)",
			input: []byte{0x1A}, // LDAX D
			initialState: CPU{
				PC: 0x1000,
				Registers: Registers{
					D: 0x30,
					E: 0x20,
				},
				Memory: Memory{
					0x3020: 0xCD,
				},
			},
			expectedState: CPU{
				PC: 0x1001,
				Registers: Registers{
					A: 0xCD,
					D: 0x30,
					E: 0x20,
				},
				Memory: Memory{
					0x3020: 0xCD,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "LHLD (Load H and L Direct Address)",
			input: []byte{0x2A, 0x10, 0x20}, // LHLD 0x2010
			initialState: CPU{
				PC: 0x1000,
				Memory: Memory{
					0x2010: 0x34,
					0x2011: 0x12,
				},
			},
			expectedState: CPU{
				PC: 0x1003,
				Registers: Registers{
					L: 0x34,
					H: 0x12,
				},
				Memory: Memory{
					0x2010: 0x34,
					0x2011: 0x12,
				},
			},
			expectedSize: 3,
		},
		{
			name:  "LXI (Load Register Pair Immediate - BC)",
			input: []byte{0x01, 0x34, 0x12}, // LXI B, 0x1234
			initialState: CPU{
				PC: 0x1000,
			},
			expectedState: CPU{
				PC: 0x1003,
				Registers: Registers{
					B: 0x12,
					C: 0x34,
				},
			},
			expectedSize: 3,
		},
		{
			name:  "LXI (Load Register Pair Immediate - DE)",
			input: []byte{0x11, 0x78, 0x56}, // LXI D, 0x5678
			initialState: CPU{
				PC: 0x2000,
			},
			expectedState: CPU{
				PC: 0x2003,
				Registers: Registers{
					D: 0x56,
					E: 0x78,
				},
			},
			expectedSize: 3,
		},
		{
			name:  "LXI (Load Register Pair Immediate - HL)",
			input: []byte{0x21, 0x9A, 0xBC}, // LXI H, 0xBC9A
			initialState: CPU{
				PC: 0x3000,
			},
			expectedState: CPU{
				PC: 0x3003,
				Registers: Registers{
					H: 0xBC,
					L: 0x9A,
				},
			},
			expectedSize: 3,
		},
		{
			name:  "LXI (Load Stack Pointer Immediate)",
			input: []byte{0x31, 0xEF, 0xCD}, // LXI SP, 0xCDEF
			initialState: CPU{
				PC: 0x4000,
			},
			expectedState: CPU{
				PC: 0x4003,
				SP: 0xCDEF,
			},
			expectedSize: 3,
		},
		{
			name:  "MOV (Move Register B to Register A)",
			input: []byte{0x78}, // MOV A, B
			initialState: CPU{
				PC: 0x5000,
				Registers: Registers{
					B: 0x56,
				},
			},
			expectedState: CPU{
				PC: 0x5001,
				Registers: Registers{
					A: 0x56,
					B: 0x56,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "MOV (Move Memory to Register L)",
			input: []byte{0x6E}, // MOV L, M
			initialState: CPU{
				PC: 0x6000,
				Registers: Registers{
					H: 0x20,
					L: 0x10,
				},
				Memory: Memory{
					0x2010: 0x89,
				},
			},
			expectedState: CPU{
				PC: 0x6001,
				Registers: Registers{
					H: 0x20,
					L: 0x89,
				},
				Memory: Memory{
					0x2010: 0x89,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "MVI (Move Immediate to Register B)",
			input: []byte{0x06, 0xAB}, // MVI B, 0xAB
			initialState: CPU{
				PC: 0x7000,
			},
			expectedState: CPU{
				PC: 0x7002,
				Registers: Registers{
					B: 0xAB,
				},
			},
			expectedSize: 2,
		},
		{
			name:  "MVI (Move Immediate to Memory Location Indexed by HL)",
			input: []byte{0x36, 0xCD}, // MVI M, 0xCD
			initialState: CPU{
				PC: 0x8000,
				Registers: Registers{
					H: 0x22,
					L: 0x10,
				},
				Memory: Memory{},
			},
			expectedState: CPU{
				PC: 0x8002,
				Registers: Registers{
					H: 0x22,
					L: 0x10,
				},
				Memory: Memory{
					0x2210: 0xCD,
				},
			},
			expectedSize: 2,
		},
		{
			name:  "ORA (Logical OR Accumulator with Register B)",
			input: []byte{0xB0}, // ORA B
			initialState: CPU{
				PC: 0x9000,
				Registers: Registers{
					A: 0x0F,
					B: 0xF0,
				},
			},
			expectedState: CPU{
				PC: 0x9001,
				Registers: Registers{
					A: 0xFF,
					B: 0xF0,
				},
				PSW: PSW{P: true, S: true},
			},
			expectedSize: 1,
		},
		{
			name:  "ORI (Logical OR Accumulator with Immediate)",
			input: []byte{0xF6, 0x3C}, // ORI 0x3C
			initialState: CPU{
				PC: 0xA000,
				Registers: Registers{
					A: 0xC3,
				},
			},
			expectedState: CPU{
				PC: 0xA002,
				Registers: Registers{
					A: 0xFF,
				},
				PSW: PSW{P: true, S: true},
			},
			expectedSize: 2,
		},
		{
			name:  "PCHL (Load HL into PC)",
			input: []byte{0xE9}, // PCHL
			initialState: CPU{
				PC: 0xB000,
				Registers: Registers{
					H: 0x12,
					L: 0x34,
				},
			},
			expectedState: CPU{
				PC: 0x1234,
				Registers: Registers{
					H: 0x12,
					L: 0x34,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "STAX (Store Accumulator in Memory Addressed by BC)",
			input: []byte{0x02}, // STAX B
			initialState: CPU{
				PC: 0xC000,
				Registers: Registers{
					A: 0x5A,
					B: 0x20,
					C: 0x10,
				},
				Memory: Memory{},
			},
			expectedState: CPU{
				PC: 0xC001,
				Registers: Registers{
					A: 0x5A,
					B: 0x20,
					C: 0x10,
				},
				Memory: Memory{
					0x2010: 0x5A,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "STAX (Store Accumulator in Memory Addressed by DE)",
			input: []byte{0x12}, // STAX D
			initialState: CPU{
				PC: 0xC100,
				Registers: Registers{
					A: 0x7E,
					D: 0x30,
					E: 0x20,
				},
				Memory: Memory{},
			},
			expectedState: CPU{
				PC: 0xC101,
				Registers: Registers{
					A: 0x7E,
					D: 0x30,
					E: 0x20,
				},
				Memory: Memory{
					0x3020: 0x7E,
				},
			},
			expectedSize: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.initialState

			instruction, size, err := DecodeBytes(tc.input)

			if err != nil {
				t.Logf("error: %s", err)
			}
			if (err != nil) != tc.expectedError {
				t.Fatalf("expected error: %v, got: %v", tc.expectedError, err)
			}

			if size != tc.expectedSize {
				t.Errorf("expected size: %d, got: %d", tc.expectedSize, size)
			}

			if err == nil {
				m.Exec(instruction)

				if m != tc.expectedState {
					t.Errorf("expected state:\n%s\ngot:\n%s\ninitial:\n%s", &tc.expectedState, &m, &tc.initialState)
					dumpMemory(t, &m)
				}
			}
		})
	}
}

func dumpMemory(t *testing.T, m *CPU) {
	t.Helper()
	t.Log("memory:")
	_ = m.Memory.DumpSparse(testutil.NewTestLogWriter(t), 0, len(m.Memory))
}

func TestAllInstructions(t *testing.T) {
	data := [3]byte{0, 1, 2}
	invalid := [256]bool{
		0x08: true,
		0x10: true, // TODO: Currently noticed in the monitor program.
		0x18: true,
		0x20: true,
		0x28: true,
		0x30: true,
		0x38: true,
		0xCB: true,
		0xD9: true,
		0xDD: true,
		0xED: true,
		0xFD: true,
	}

	for i := range 256 {
		data[0] = byte(i)
		cmd, n, err := DecodeBytes(data[:])
		if invalid[i] {
			if err == nil {
				name := runtime.FuncForPC(reflect.ValueOf(cmd.Execute).Pointer()).Name()
				t.Errorf("expected error decoding instruction %02x, got %d bytes %s", i, n, name)
			}
			continue
		}

		if err != nil {
			t.Errorf("error decoding instruction %02x: %s", i, err)
			continue
		}

		if n < 1 || n > 3 {
			t.Errorf("invalid instruction size for %02x: %d", i, n)
		}
		if cmd.Execute == nil {
			t.Errorf("no exec func for %02x", i)
		}
	}
}
