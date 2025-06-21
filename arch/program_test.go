package arch_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"rmazur.io/fahivets/arch"
	"rmazur.io/fahivets/internal/testutil"
)

func readProgram(t *testing.T, name string) []byte {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	res, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestBootloader(t *testing.T) {
	bootProg := readProgram(t, "bootloader.rom")
	monitorProg := readProgram(t, "monitor.rom")

	romStart := arch.MemoryMapping(arch.MemROM2K)

	var m arch.CPU
	copy(m.Memory[romStart:], bootProg)
	copy(m.Memory[romStart+0x830:], monitorProg)
	m.PC = romStart

	tOut := testutil.NewTestLogWriter(t)

	for i := 0; i < 16000; i++ {
		addr := m.PC
		cmd, err := m.Step()
		if err != nil {
			t.Logf("0x%04x: %s", addr, &m)
			_ = m.Memory.DumpSparse(tOut)
			t.Fatal(err)
		}
		t.Logf("0x%04x: %s\t%s", addr, cmd.Name, &m)
	}
	_ = m.Memory.Dump(tOut, 0xff00, 0xffff)
}
