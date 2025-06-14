package arch

import "testing"

func TestMemoryMapping(t *testing.T) {
	regStart, regEnd := MemoryMappingRange(MemRegisters2K)
	if regStart != 0xF800 {
		t.Errorf("memory mapping failed: got 0x%X, want 0xF800", regStart)
	}
	if regEnd != 0xFFFF {
		t.Errorf("memory mapping failed: got 0x%X, want 0xFFFF", regEnd)
	}
}
