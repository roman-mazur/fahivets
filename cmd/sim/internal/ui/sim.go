package main

import (
	"bytes"
	_ "embed"
	"image"
	"log"

	"golang.org/x/image/draw"
	"rmazur.io/fahivets"
	"rmazur.io/fahivets/arch"
	"rmazur.io/fahivets/devices"
)

var (
	//go:embed progs/bootloader.rom
	bootloader []byte
	//go:embed progs/monitor.rom
	monitor []byte

	//go:embed progs/rain.rks
	programRks []byte

	m = fahivets.NewComputer()
)

func main() {
	ui := makeUiWorld()

	prepareSimulation(m)

	frame := m.Display.Image()
	size := frame.Bounds()
	frameBuf := image.NewRGBA(image.Rect(0, 0, size.Dx()*2, size.Dy()*2))
	fillBuf(frameBuf, frame)

	ui.ConsumeDisplayFrames(func() image.Image {
		const (
			cpuFreq     = 2_000_000 // 2 MHz
			refreshRate = 60        // Hz
			perFrame    = cpuFreq / refreshRate
		)

		n := 0
		for n < perFrame {
			_, c, err := m.Step()
			if err != nil {
				log.Println("step error:", err)
				break
			}
			n += c
		}
		fillBuf(frameBuf, m.Display.Image())
		return frameBuf
	})

	ui.ConnectKeyboard(m.Keyboard)

	var done chan struct{}
	<-done
}

type UiWorld interface {
	ConsumeDisplayFrames(frameF func() image.Image)
	ConnectKeyboard(keyboard *devices.Keyboard)
}

func prepareSimulation(m *fahivets.Computer) {
	// Note: it should be possible to minimize the exe size if we link
	// programs directly to the CPU.Memory.
	romStart := arch.MemoryMapping(arch.MemROM2K)
	copy(m.CPU.Memory[romStart:], bootloader)
	copy(m.CPU.Memory[arch.MemoryMapping(arch.MemROMExtra12K):], monitor)
	m.CPU.PC = uint16(romStart)

	// Make sure bootloader is executed.
	for range 16_000 {
		_, _, err := m.Step()
		if err != nil {
			log.Println(err)
			return
		}
	}

	// Load the program.
	program, err := fahivets.ReadRks(bytes.NewReader(programRks))
	if err != nil {
		panic(err)
	}
	copy(m.CPU.Memory[program.StartAddress:], program.Content)
	m.CPU.Exec(arch.JMP(program.StartAddress))
}

func fillBuf(buf *image.RGBA, frame image.Image) {
	draw.NearestNeighbor.Scale(buf, buf.Bounds(), frame, frame.Bounds(), draw.Src, nil)
}
