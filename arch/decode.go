package arch

import "fmt"

func DecodeBytes(data []byte) (Instruction, int, error) {
	switch cmdByte := data[0]; cmdByte {
	case 0x00:
		return NOP(), 1, nil
	case 0x07:
		return RAL(), 1, nil
	case 0x0F:
		return RAR(), 1, nil
	case 0x17:
		return RLC(), 1, nil
	case 0x1F:
		return RRC(), 1, nil
	case 0x22:
		return SHLD(nextWord(data)), 3, nil
	case 0x27:
		return DAA(), 1, nil
	case 0x2A:
		return LHLD(nextWord(data)), 3, nil
	case 0x2F:
		return CMA(), 1, nil
	case 0x32:
		return STA(nextWord(data)), 3, nil
	case 0x3A:
		return LDA(nextWord(data)), 3, nil
	case 0x3F:
		return CMC(), 1, nil
	case 0x37:
		return STC(), 1, nil
	case 0x76:
		return HLT(), 1, nil
	case 0xC3:
		return JMP(nextWord(data)), 3, nil
	case 0xC6:
		return ADI(data[1]), 2, nil
	case 0xC9:
		return RET(), 1, nil
	case 0xCE:
		return ACI(data[1]), 2, nil
	case 0xCD:
		return CALL(nextWord(data)), 3, nil
	case 0xD3:
		return OUT(data[1]), 2, nil
	case 0xD6:
		return SUI(data[1]), 2, nil
	case 0xDB:
		return IN(data[1]), 2, nil
	case 0xDE:
		return SBI(data[1]), 2, nil
	case 0xE3:
		return XTHL(), 1, nil
	case 0xE6:
		return ANI(data[1]), 2, nil
	case 0xE9:
		return PCHL(), 1, nil
	case 0xEB:
		return XCHG(), 1, nil
	case 0xEE:
		return XRI(data[1]), 2, nil
	case 0xF3:
		return DI(), 1, nil
	case 0xF6:
		return ORI(data[1]), 2, nil
	case 0xF9:
		return SPHL(), 1, nil
	case 0xFB:
		return EI(), 1, nil
	case 0xFE:
		return CPI(data[1]), 2, nil
	default:
		// Addition with register.
		if cmdByte>>4 == 8 {
			r := cmdByte & 0x07
			if mask(cmdByte, 0x08) {
				return ADC(r), 1, nil
			} else {
				return ADD(r), 1, nil
			}
		}
		// And with register.
		if cmdByte>>3 == 0x14 {
			return ANA(cmdByte & 0x07), 1, nil
		}
		// Or with register.
		if cmdByte>>3 == 0x16 {
			return ORA(cmdByte & 0x07), 1, nil
		}
		// Conditional call.
		if cmdByte&0x7 == 0x4 && mask(cmdByte, 0xC0) {
			cnd := ConditionCode((cmdByte & 0x38) >> 3)
			return Ccnd(cnd, nextWord(data)), 3, nil
		}
		// Compare with a register.
		if cmdByte>>3 == 0x17 {
			r := cmdByte & 0x07
			return CMP(r), 1, nil
		}
		// DAD.
		if cmdByte&0x0F == 0x09 && cmdByte>>6 == 0 {
			rp := (cmdByte & 0x30) >> 4
			return DAD(rp), 1, nil
		}
		// Decrements.
		if cmdByte>>6 == 0 {
			if cmdByte&0x07 == 0x05 {
				r := cmdByte & 0x38 >> 3
				return DCR(r), 1, nil
			}
			if cmdByte&0x0F == 0x0B {
				rp := cmdByte & 0x30 >> 4
				return DCX(rp), 1, nil
			}
		}
		if mask(cmdByte, 0xC0) {
			// Pop and push.
			rp := cmdByte & 0x30 >> 4
			if cmdByte&0x0F == 0x01 {
				return POP(rp), 1, nil
			}
			if cmdByte&0x0F == 0x05 {
				return PUSH(rp), 1, nil
			}
			// Conditional return.
			cnd := ConditionCode((cmdByte >> 3) & 0x07)
			if cmdByte&0x07 == 0 {
				return Rcnd(cnd), 1, nil
			}
			// Reset/interrupts.
			if cmdByte&0x07 == 0x07 {
				return RST(byte(cnd)), 1, nil
			}
			// Conditional jump.
			if cmdByte&0x07 == 0x02 {
				return JCnd(cnd, nextWord(data)), 3, nil
			}
		}

		// Subtraction.
		if cmdByte>>4 == 9 {
			r := cmdByte & 0x07
			if mask(cmdByte, 0x08) {
				return SBB(r), 1, nil
			} else {
				return SUB(r), 1, nil
			}
		}

		if cmdByte>>4 == 0xA && mask(cmdByte, 0x08) {
			return XRA(cmdByte & 0x07), 1, nil
		}

		// Increments.
		if cmdByte>>6 == 0 {
			if cmdByte&0x07 == 0x04 {
				r := cmdByte & 0x38 >> 3
				return INR(r), 1, nil
			}
			if cmdByte&0x0F == 0x03 {
				return INX(cmdByte & 0x30 >> 4), 1, nil
			}
		}

		// Load.
		if cmdByte&0x0F == 0x0A && cmdByte>>5 == 0 {
			return LDAX(cmdByte >> 4 & 0x01), 1, nil
		}
		if cmdByte&0x0F == 0x01 && cmdByte>>6 == 0 {
			rp := cmdByte >> 4 & 0x03
			return LXI(rp, nextWord(data)), 3, nil
		}

		// Store.
		if cmdByte&0x0F == 0x02 && cmdByte>>6 == 0 {
			rp := cmdByte >> 4 & 0x03
			return STAX(rp), 1, nil
		}

		// Move.
		switch sel := cmdByte >> 6; sel {
		case 0:
			if cmdByte&0x07 == 0x06 {
				return MVI(cmdByte>>3&0x07, data[1]), 2, nil
			}
		case 1:
			return MOV(cmdByte>>3&0x07, cmdByte&0x07), 1, nil
		}

		return Instruction{}, 0, fmt.Errorf("unknown instruction 0x%02x (%#b)", data[0], data[0])
	}
}

func nextWord(data []byte) uint16 { return uint16(data[1]) | uint16(data[2])<<8 }

func mask(cmdByte, mask byte) bool { return (cmdByte & mask) == mask }
