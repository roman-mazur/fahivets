package fahivets_test

import (
	"bytes"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"runtime"
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
	monitorStart := arch.MemoryMapping(arch.MemROMExtra12K)

	m := fahivets.NewComputer()
	copy(m.CPU.Memory[romStart:], bootProg)
	copy(m.CPU.Memory[monitorStart:], monitorProg)
	m.CPU.PC = uint16(romStart)

	tOut := testutil.NewTestLogWriter(t)

	const maxSteps = 16_000

	advance(t, m, maxSteps, false)

	t.Log("IO")
	_ = m.CPU.Memory.Dump(tOut, arch.MemoryIoCtrl, arch.MemoryIoCtrl+64)

	t.Log("DISPLAY")
	displayStart, displayEnd := arch.MemoryMappingRange(arch.MemDisplay12K)
	_ = m.CPU.Memory.DumpSparse(tOut, displayStart, displayEnd+1)

	var outImg bytes.Buffer
	err := png.Encode(&outImg, m.Display.Image())
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

	t.Log("Running monitor")
	m.CPU.PC = uint16(monitorStart)

	advance(t, m, maxSteps*5, false)
	captureDisplay(t, m, "test.png")

	// 0xc269
	t.Log("check keypress subroutine")
	m.CPU.PC = 0xc269
	keyCode := devices.MatrixKeyCode(3, 3)
	m.Keyboard.Event(keyCode, devices.KeyStateDown)
	runtime.Gosched()
	advance(t, m, 9, true)
	m.Keyboard.Event(keyCode, devices.KeyStateUp)
	runtime.Gosched()
	t.Log("key is up")
	advance(t, m, 255*80, true)
}

func advance(t *testing.T, m *fahivets.Computer, steps int, debug bool) {
	t.Helper()
	t.Logf("advancing by %d steps", steps)
	tOut := testutil.NewTestLogWriter(t)
	for i := range steps {
		addr := m.CPU.PC
		cmd, err := m.Step()
		if err != nil {
			t.Logf("%05d 0x%04x:\t%s", i, addr, &m.CPU)
			_ = m.CPU.Memory.DumpSparse(tOut, 0, len(m.CPU.Memory))
			t.Fatal(err)
		}
		if debug {
			t.Logf("%05d 0x%04x: %s\t%s", i, addr, cmd.Name, &m.CPU)
		}
	}
}

func captureDisplay(t *testing.T, m *fahivets.Computer, name string) {
	t.Helper()
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error("cannot close the output file:", err)
		}
	})

	err = png.Encode(f, m.Display.Image())
	if err != nil {
		t.Error(err)
	}
	t.Log("display captured in", name)
}
