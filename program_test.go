package fahivets_test

import (
	"bytes"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rmazur.io/fahivets"
	"rmazur.io/fahivets/arch"
	"rmazur.io/fahivets/devices"
	"rmazur.io/fahivets/internal/testutil"
)

func readData(t *testing.T, name string) []byte {
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
	bootProg := readData(t, "progs/bootloader.rom")
	monitorProg := readData(t, "progs/monitor.rom")

	romStart := arch.MemoryMapping(arch.MemROM2K)

	m := fahivets.NewComputer()
	copy(m.CPU.Memory[romStart:], bootProg)
	copy(m.CPU.Memory[romStart+0x830:], monitorProg)
	m.CPU.PC = uint16(romStart)

	tOut := testutil.NewTestLogWriter(t)

	const (
		debug    = false
		maxSteps = 16_000
	)

	for i := range maxSteps {
		addr := m.CPU.PC
		cmd, err := m.Step()
		if err != nil {
			t.Logf("%05d 0x%04x:\t%s", i, addr, &m.CPU)
			_ = m.CPU.Memory.DumpSparse(tOut, 0, len(m.CPU.Memory))
			t.Fatal(err)
		}
		if debug || strings.Contains(cmd.Name, "0xff") || m.CPU.Registers.H == 0xFF || i > maxSteps-100 {
			t.Logf("%05d 0x%04x: %s\t%s", i, addr, cmd.Name, &m.CPU)
		}
	}

	t.Log("IO")
	_ = m.CPU.Memory.Dump(tOut, arch.MemoryIoCtrl, arch.MemoryIoCtrl+64)

	t.Log("DISPLAY")
	displayStart, displayEnd := arch.MemoryMappingRange(arch.MemDisplay12K)
	_ = m.CPU.Memory.DumpSparse(tOut, displayStart, displayEnd+1)

	var outImg bytes.Buffer
	err := png.Encode(&outImg, devices.NewDisplay(&m.CPU).Image())
	if err != nil {
		t.Fatal(err)
	}

	sample := readData(t, "display-sample.png")
	if bytes.Compare(sample, outImg.Bytes()) != 0 {
		t.Error("output image is different from the sample")
		out, err := os.Create("error.png")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = out.Close() })
		_, err = io.Copy(out, &outImg)
		if err != nil {
			t.Fatal(err)
		}
	}
}
