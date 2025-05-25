package arch

import "fmt"

type InstructionSet struct {
}

func (is *InstructionSet) DecodeBytes(data []byte) (Instruction, int, error) {
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
	case 0x27:
		return DAA(), 1, nil
	case 0x2F:
		return CMA(), 1, nil
	case 0x3F:
		return CMC(), 1, nil
	case 0x37:
		return STC(), 1, nil
	case 0xC6:
		return ADI(data[1]), 2, nil
	case 0xC9:
		return RET(), 1, nil
	case 0xCE:
		return ACI(data[1]), 2, nil
	case 0xCD:
		return CALL(nextWord(data)), 3, nil
	case 0xDE:
		return SBI(data[1]), 2, nil
	case 0xE6:
		return ANI(data[1]), 2, nil
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
		// Conditional call.
		if cmdByte&0x7 == 0x4 && mask(cmdByte, 0xC0) {
			cnd := (cmdByte & 0x38) >> 3
			return Ccnd(cnd, nextWord(data)), 3, nil
		}
		// Compare with a register.
		if cmdByte>>3 == 0x17 {
			r := cmdByte & 0x07
			return CMP(r), 1, nil
		}
		// DAD.
		if cmdByte&0x0F == 0x09 && !mask(cmdByte, 0xC0) {
			rp := (cmdByte & 0x30) >> 4
			return DAD(rp), 1, nil
		}
		// Decrements.
		if !mask(cmdByte, 0xC0) {
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
			cnd := (cmdByte & 0x38) >> 3
			if cmdByte&0x07 == 0 {
				return Rcnd(cnd), 1, nil
			}
			// Reset/interrupts.
			if cmdByte&0x07 == 0x07 {
				return RST(cnd), 1, nil
			}
		}

		// Subtraction.
		if cmdByte>>4 == 9 && mask(cmdByte, 0x08) {
			return SBB(cmdByte & 0x07), 1, nil
		}
		return Instruction{}, 0, fmt.Errorf("unknown instruction 0x%02x", data[0])
	}
}

func nextWord(data []byte) uint16 { return uint16(data[1]) | uint16(data[2])<<8 }

func mask(cmdByte, mask byte) bool { return (cmdByte & mask) == mask }
