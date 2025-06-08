package arch

import (
	"io"
	"os"
	"path/filepath"
	"testing"
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

	const (
		videoStart   = 0x9000
		romStart     = 0xC000
		monitorStart = 0xC830
	)

	var m Machine
	copy(m.Memory[romStart:], bootProg)
	copy(m.Memory[monitorStart:], monitorProg)
	m.SP = romStart
	m.PC = romStart

	tOut := newTestWriter(t)

	for i := 0; i < 10000; i++ {
		cmd, err := m.Step()
		if err != nil {
			t.Logf("step %d: %s", i, &m)
			_ = m.Memory.DumpSparse(tOut)
			t.Fatal(err)
		}
		t.Log(i, cmd.Name)
	}
	//_ = m.Memory.Dump(tOut, videoStart, romStart)
}
