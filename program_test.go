package fahivets_test

import (
	"bytes"
	"fmt"
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
	t.Cleanup(func() { _ = f.Close() })
	res, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func readRks(t *testing.T, name string) fahivets.RksData {
	data := readData(t, name)
	res, err := fahivets.ReadRks(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func initWithBootloader(t *testing.T) *fahivets.Computer {
	bootProg := readData(t, "progs/bootloader.rom")
	monitorProg := readData(t, "progs/monitor.rom")

	romStart := arch.MemoryMapping(arch.MemROM2K)
	monitorStart := arch.MemoryMapping(arch.MemROMExtra12K)

	m := fahivets.NewComputer()
	copy(m.CPU.Memory[romStart:], bootProg)
	copy(m.CPU.Memory[monitorStart:], monitorProg)
	m.CPU.PC = uint16(romStart)

	const maxSteps = 16_000
	advance(t, m, maxSteps, false)
	return m
}

func TestBootloader(t *testing.T) {
	monitorStart := arch.MemoryMapping(arch.MemROMExtra12K)

	m := initWithBootloader(t)
	tOut := testutil.NewTestLogWriter(t)

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

	advance(t, m, 16_000*5, false)
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

func TestGames(t *testing.T) {
	rainGame := readRks(t, "progs/rain.rks")
	chessGame := readRks(t, "progs/chess4.rks")

	run := func(t *testing.T, data fahivets.RksData, start int, steps int, dirName, prefix string) *fahivets.Computer {
		m := initWithBootloader(t)
		copy(m.CPU.Memory[data.StartAddress:], data.Content)
		m.CPU.PC = uint16(start)
		advance(t, m, steps, false)
		captureDisplay(t, m, fmt.Sprintf("testdata/%s/%s-test-%d.png", dirName, prefix, start))
		return m
	}

	t.Run("chess/run", func(t *testing.T) {
		run(t, chessGame, 0, 10000, "examples", "chess")
	})

	t.Run("rain/run", func(t *testing.T) {
		m := run(t, rainGame, 48, 32000, "examples", "rain")
		digit1 := devices.MatrixKeyCode(4, 10)
		m.Keyboard.Event(digit1, devices.KeyStateDown)
		runtime.Gosched()
		advance(t, m, 256, true)
		m.Keyboard.Event(digit1, devices.KeyStateUp)
		runtime.Gosched()
		advance(t, m, 16000, false)
	})
}

func advance(t *testing.T, m *fahivets.Computer, steps int, debug bool) {
	t.Helper()
	t.Logf("advancing by %d steps", steps)
	defer func() {
		e := recover()
		if e != nil {
			t.Errorf("panic: %v", e)
		}
	}()

	tOut := testutil.NewTestLogWriter(t)
	for i := range steps {
		addr := m.CPU.PC
		cmd, err := m.Step()
		if err != nil {
			t.Logf("%05d 0x%04x:\t%s", i, addr, &m.CPU)
			if debug {
				_ = m.CPU.Memory.DumpSparse(tOut, 0, len(m.CPU.Memory))
			}
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

func TestPrograms(t *testing.T) {
	for _, tc := range []struct {
		name         string
		offset       int
		startAddress int
	}{
		{"bootloader.rom", 0, arch.MemoryMapping(arch.MemROM2K)},
		{"rain.rks", 4, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := readData(t, filepath.Join("progs", tc.name))
			prog, n, err := arch.DecodeBytesAll(data[tc.offset:])
			prog.StartAddress = tc.startAddress
			t.Log(prog, n, err)
		})
	}
}
