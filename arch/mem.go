package arch

import (
	"fmt"
	"io"
)

type Memory [65536]byte

type MemSection byte

const (
	MemUser16K MemSection = iota
	MemReserved16K
	MemUser4K
	MemDisplay12K
	MemROM2K
	MemROMExtra12K
	MemRegisters2K

	memSectionsCnt
)

var memoryMapping = [memSectionsCnt]uint16{
	MemUser16K:     0,
	MemReserved16K: 0x4000,
	MemUser4K:      0x8000,
	MemDisplay12K:  0x9000,
	MemROM2K:       0xC000,
	MemROMExtra12K: 0xC800,
	MemRegisters2K: 0xF800,
}

// MemoryMappingRange returns the start and end address of a particular memory section.
func MemoryMappingRange(s MemSection) (start uint16, end uint16) {
	return memoryMapping[s], memoryMapping[(s+1)%memSectionsCnt] - 1
}

// MemoryMapping returns the start address of the selected memory section.
func MemoryMapping(s MemSection) uint16 { return memoryMapping[s] }

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

func (m *Memory) Dump(out io.Writer, start, end uint16) error {
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
