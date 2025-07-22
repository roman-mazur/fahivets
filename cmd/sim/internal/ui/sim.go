//go:build js

package main

import (
	"bytes"
	_ "embed"
	"image"
	"log"
	"syscall/js"

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
	rainRks []byte

	m = fahivets.NewComputer()
)

func main() {
	var ui UiWorld

	ui = &jsUiWorld{root: js.Global()}
	newFrames := make(chan image.Image, 1)
	processedFrames := make(chan image.Image)

	go runSimulation(m, newFrames, processedFrames)

	ui.ConsumeDisplayFrames(newFrames, processedFrames)
	ui.ConnectKeyboard(m.Keyboard)

	var done chan struct{}
	<-done
}

type UiWorld interface {
	ConsumeDisplayFrames(newFrames, processedFrames chan image.Image)
	ConnectKeyboard(keyboard *devices.Keyboard)
}

func runSimulation(m *fahivets.Computer, newFrames, processedFrames chan image.Image) {
	// Note: it should be possible to minimize the exe size if we link
	// programs directly to the CPU.Memory.
	romStart := arch.MemoryMapping(arch.MemROM2K)
	copy(m.CPU.Memory[romStart:], bootloader)
	copy(m.CPU.Memory[arch.MemoryMapping(arch.MemROMExtra12K):], monitor)
	m.CPU.PC = uint16(romStart)

	// Make sure bootloader is executed.
	for range 16_000 {
		_, err := m.Step()
		if err != nil {
			log.Println(err)
			return
		}
	}

	// Load the rain game.
	rainProg, err := fahivets.ReadRks(bytes.NewReader(rainRks))
	if err != nil {
		panic(err)
	}
	copy(m.CPU.Memory[rainProg.StartAddress:], rainProg.Content)
	m.CPU.Exec(arch.JMP(rainProg.StartAddress))

	var dp displayPipeline
	for {
		_, err := m.Step()
		if err != nil {
			log.Println(err)
			return
		}
		dp.Advance(newFrames, processedFrames, m.Display)
		m.SimSleep()
	}
}

type displayPipeline struct {
	init bool
	buf  *image.RGBA
}

func (dp *displayPipeline) Advance(newFrames, processedFrames chan image.Image, display *devices.Display) {
	dp.ensureBuffers(display)

	var out, in chan image.Image
	if dp.buf != nil {
		out = newFrames
	} else {
		in = processedFrames
	}
	select {
	case out <- dp.buf:
		dp.buf = nil
	case buf := <-in:
		dp.buf = buf.(*image.RGBA)
		dp.fillBuf(display.Image())
	default:
	}
}

func (dp *displayPipeline) ensureBuffers(display *devices.Display) {
	if !dp.init {
		dp.init = true
		frame := display.Image()
		size := frame.Bounds()
		dp.buf = image.NewRGBA(image.Rect(0, 0, size.Dx()*2, size.Dy()*2))
		dp.fillBuf(frame)
	}
}

func (dp *displayPipeline) fillBuf(frame image.Image) {
	draw.NearestNeighbor.Scale(dp.buf, dp.buf.Bounds(), frame, frame.Bounds(), draw.Src, nil)
}
