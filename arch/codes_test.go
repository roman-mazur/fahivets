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
				PSW: PSW{
					C: true,
				},
			},
			expectedState: Machine{
				PC: 0x1002,
				Registers: Registers{
					A: 0x0A,
				},
				PSW: PSW{
					C: false,
				},
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
