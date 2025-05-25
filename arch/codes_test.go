package arch

import "testing"

func TestInstructionSet_DecodeAndExecute(t *testing.T) {
	is := &InstructionSet{}

	testCases := []struct {
		name          string
		input         []byte
		initialState  Machine
		expectedState Machine
		expectedSize  int
		expectedError bool
	}{
		{
			name:  "NOP",
			input: []byte{0x00},
			initialState: Machine{
				PC: 0x1000,
			},
			expectedState: Machine{
				PC: 0x1001, // PC should increment by instruction size
			},
			expectedSize: 1,
		},
		{
			name:  "CMC",
			input: []byte{0x3F},
			initialState: Machine{
				PC: 0x1000,
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
			},
			expectedState: Machine{
				PC: 0x1001,
				PSW: PSW{
					C: true, // Carry flag should be set
				},
			},
			expectedSize: 1,
		},
		{
			name:  "ADD B",
			input: []byte{0x80},
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					A: 0x01,
					B: 0x02,
				},
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					A: 0x01,
				},
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					A: 0x01,
				},
				PSW: PSW{
					C: true, //Carry is set
				},
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					A: 0x02,
					B: 0x01,
				},
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					A: 0x02,
				},
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					A: 0x51,
				},
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					A: 0x0A,
					E: 0x05,
				},
				PSW: PSW{C: true}, // Initial Carry flag
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					A: 0x02,
					E: 0x05,
				},
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					A: 0x0A,
					E: 0x0A,
				},
				PSW: PSW{C: true}, // Initial Carry flag
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					A: 0x0A,
				},
				PSW: PSW{C: true},
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					B: 0x01,
				},
			},
			expectedState: Machine{
				PC: 0x1001,
				Registers: Registers{
					B: 0x02,
				},
				PSW: PSW{
					P: true,
				},
			},
			expectedSize: 1,
		},
		{
			name:  "DCR B",
			input: []byte{0x05},
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					B: 0x01,
				},
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					A: 0x03,
					B: 0x01,
				},
				PSW: PSW{
					C: true,
				},
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					A: 0x03,
				},
				PSW: PSW{
					C: true,
				},
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				Registers: Registers{
					A: 0x0A,
					B: 0x0A,
				},
			},
			expectedState: Machine{
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
			input:         []byte{0xFF},
			expectedSize:  0,
			expectedError: true,
		},
		{
			name:  "CALL 0x2050",
			input: []byte{0xCD, 0x50, 0x20}, // CALL 0x2050
			initialState: Machine{
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC:  0x1000,
				SP:  0x2000,
				PSW: PSW{C: true}, // Carry flag is set
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				SP: 0x2000,
				// Carry flag is not set
			},
			expectedState: Machine{
				PC: 0x1003, // PC should increment to the next instruction
				SP: 0x2000, // SP should not change
				// Memory should not change
			},
			expectedSize: 3,
		},
		{
			name:  "CM 0x2050 (Sign Set)",
			input: []byte{0xFC, 0x50, 0x20}, // CM 0x2050
			initialState: Machine{
				PC:  0x1000,
				SP:  0x2000,
				PSW: PSW{S: true}, // Sign flag is set
			},
			expectedState: Machine{
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
			initialState: Machine{
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: Machine{
				PC: 0x1003, // PC should increment to the next instruction
				SP: 0x2000, // SP should not change
			},
			expectedSize: 3,
		},
		{
			name:  "DAA (A = 0x9B)",
			input: []byte{0x27}, // DAA
			initialState: Machine{
				PC:        0x1000,
				Registers: Registers{A: 0x9B},
				PSW:       PSW{S: true},
			},
			expectedState: Machine{
				PC:        0x1001,
				Registers: Registers{A: 0x01},
				PSW:       PSW{C: true, P: true, A: true},
			},
			expectedSize: 1,
		},
		{
			name:  "DAD D (H:L = 0x0101, D:E = 0x0100)",
			input: []byte{0x19}, // DAD D
			initialState: Machine{
				Registers: Registers{
					H: 0x01,
					L: 0x01,
					D: 0x01,
					E: 0x00,
				},
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: Machine{
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
			initialState: Machine{
				Registers: Registers{
					H: 0xA1,
					L: 0x7B,
					B: 0x60,
					C: 0x00,
				},
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: Machine{
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
			initialState: Machine{
				Registers: Registers{
					H: 0x3A,
					L: 0x7C,
				},
				PC:     0x1000,
				SP:     0x2000,
				Memory: [65536]byte{0x3A7C: 0x02}, // Memory must be initialized
			},
			expectedState: Machine{
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
			initialState: Machine{
				Registers: Registers{
					A: 42,
				},
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: Machine{
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
			initialState: Machine{
				Registers: Registers{
					H: 0x40,
					L: 0x00,
				},
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: Machine{
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
			initialState: Machine{
				Registers: Registers{
					B: 0x12,
					C: 0x34,
				},
				SP: 0x2000,
				PC: 0x1000,
			},
			expectedState: Machine{
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
			initialState: Machine{
				SP:     0x1FFE,
				PC:     0x1000,
				Memory: [65536]byte{0x1FFE: 0x56, 0x1FFF: 0x78},
			},
			expectedState: Machine{
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
			initialState: Machine{
				Registers: Registers{
					A: 0x85, // 10000101
				},
				PC:  0x1000,
				SP:  0x2000,
				PSW: PSW{C: true},
			},
			expectedState: Machine{
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
			initialState: Machine{
				Registers: Registers{
					A: 0x81, // 10000001
				},
				PC:  0x1000,
				SP:  0x2000,
				PSW: PSW{C: true},
			},
			expectedState: Machine{
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
			initialState: Machine{
				Registers: Registers{
					A: 0x55, // 01010101
				},
				PSW: PSW{C: false},
				PC:  0x1000,
				SP:  0x2000,
			},
			expectedState: Machine{
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
			initialState: Machine{
				Registers: Registers{
					A: 0x55, // 01010101
				},
				PSW: PSW{C: true},
				PC:  0x1000,
				SP:  0x2000,
			},
			expectedState: Machine{
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
			initialState: Machine{
				SP:     0x1000,
				PC:     0x2000,
				Memory: [65536]byte{0x1000: 0x50, 0x1001: 0x20},
			},
			expectedState: Machine{
				SP:     0x1002,
				PC:     0x2050,
				Memory: [65536]byte{0x1000: 0x50, 0x1001: 0x20},
			},
			expectedSize: 1,
		},
		{
			name:  "RC (C = 1, SP = 0x1000, MEM(0x1000) = 0x50, MEM(0x1001) = 0x20)",
			input: []byte{0xD8}, // RC
			initialState: Machine{
				SP:     0x1000,
				PC:     0x2000,
				PSW:    PSW{C: true},
				Memory: [65536]byte{0x1000: 0x50, 0x1001: 0x20},
			},
			expectedState: Machine{
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
			initialState: Machine{
				SP:     0x1000,
				PC:     0x2000,
				PSW:    PSW{C: false},
				Memory: [65536]byte{0x1000: 0x50, 0x1001: 0x20},
			},
			expectedState: Machine{
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
			initialState: Machine{
				SP: 0x1000,
				PC: 0x2000,
			},
			expectedState: Machine{
				SP:     0x0FFE,
				PC:     0x0000,
				Memory: [65536]byte{0x0FFF: 0x20, 0x0FFE: 0x01},
			},
			expectedSize: 1,
		},
		{
			name:  "RST 1 (PC = 0x2000, SP = 0x1000)",
			input: []byte{0xCF}, // RST 1
			initialState: Machine{
				SP: 0x1000,
				PC: 0x2000,
			},
			expectedState: Machine{
				SP:     0x0FFE,
				PC:     0x0008,
				Memory: [65536]byte{0x0FFF: 0x20, 0x0FFE: 0x01},
			},
			expectedSize: 1,
		},
		{
			name:  "SBB B (A = 0x3E, B = 0x2A, C = 0)",
			input: []byte{0x98}, // SBB B
			initialState: Machine{
				Registers: Registers{
					A: 0x3E,
					B: 0x2A,
				},
				PSW: PSW{C: false},
				PC:  0x1000,
				SP:  0x2000,
			},
			expectedState: Machine{
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
			initialState: Machine{
				Registers: Registers{
					A: 0x3E,
					B: 0x3E,
				},
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: Machine{
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
			name:  "SBB B (A = 0x05, B = 0x10, C = 0)",
			input: []byte{0x98}, // SBB B
			initialState: Machine{
				Registers: Registers{
					A: 0x05,
					B: 0x10,
				},
				PC: 0x1000,
				SP: 0x2000,
			},
			expectedState: Machine{
				Registers: Registers{
					A: 0xF5,
					B: 0x10,
				},
				PSW: PSW{C: true, S: true, Z: false, P: true, A: true},
				PC:  0x1001,
				SP:  0x2000,
			},
			expectedSize: 1,
		},
		{
			name:  "SBI (A = 0x3E, data = 0x2A, C = 0)",
			input: []byte{0xDE, 0x2A}, // SBI 0x2A
			initialState: Machine{
				Registers: Registers{
					A: 0x3E,
				},
				PSW: PSW{C: false},
				PC:  0x1000,
				SP:  0x2000,
			},
			expectedState: Machine{
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
			initialState: Machine{
				Registers: Registers{
					A: 0x05,
				},
				PSW: PSW{C: false},
				PC:  0x1000,
				SP:  0x2000,
			},
			expectedState: Machine{
				Registers: Registers{
					A: 0xF5,
				},
				PSW: PSW{C: true, S: true, Z: false, P: false, A: true},
				PC:  0x1002,
				SP:  0x2000,
			},
			expectedSize: 2,
		},
		{
			name:  "SBI (A = 0x05, data = 0x04, C = 1)",
			input: []byte{0xDE, 0x04}, // SBI 0x04
			initialState: Machine{
				Registers: Registers{
					A: 0x05,
				},
				PSW: PSW{C: true},
				PC:  0x1000,
				SP:  0x2000,
			},
			expectedState: Machine{
				Registers: Registers{
					A: 0x00,
				},
				PSW: PSW{C: false, S: false, Z: true, P: true, A: false},
				PC:  0x1002,
				SP:  0x2000,
			},
			expectedSize: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.initialState

			instruction, size, err := is.DecodeBytes(tc.input)

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

func dumpMemory(t *testing.T, m *Machine) {
	t.Helper()
	t.Log("memory:")
	_ = m.Memory.DumpSparse((*testWriter)(t))
}

type testWriter testing.T

func (tw *testWriter) Write(p []byte) (n int, err error) {
	tw.Logf("%s", string(p))
	return len(p), nil
}
