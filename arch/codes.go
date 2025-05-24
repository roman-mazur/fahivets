package arch

import "fmt"

type InstructionSet struct {
}

func (is *InstructionSet) DecodeBytes(data []byte) (Instruction, int, error) {
	switch cmdByte := data[0]; cmdByte {
	case 0x3F:
		return CMC(), 1, nil
	case 0x37:
		return STC(), 1, nil
	default:

		return Instruction{}, 0, fmt.Errorf("unknown instruction %02x", data[0])
	}
}
