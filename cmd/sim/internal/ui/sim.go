//go:build js

package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"log"
	"syscall/js"
	"time"
	"unsafe"

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
	// Note: it should be possible to minimize the exe size if we link
	// programs directly to the CPU.Memory.
	romStart := arch.MemoryMapping(arch.MemROM2K)
	copy(m.CPU.Memory[romStart:], bootloader)
	copy(m.CPU.Memory[arch.MemoryMapping(arch.MemROMExtra12K):], monitor)
	m.CPU.PC = uint16(romStart)

	rainProg, err := fahivets.ReadRks(bytes.NewReader(rainRks))
	if err != nil {
		panic(err)
	}
	rainLoaded := false

	var scaledBuf *image.RGBA

	for {
		const maxSteps = 16_000
		for range maxSteps {
			_, err := m.Step()
			if err != nil {
				log.Println(err)
				return
			}

			kbEvent, keyCode, keyState := nextKeyboardEvent()
			if kbEvent {
				m.Keyboard.Event(keyCode, keyState)
			}
		}

		frame := m.Display.Image()
		if scaledBuf == nil {
			size := frame.Bounds()
			scaledBuf = image.NewRGBA(image.Rect(0, 0, size.Dx()*2, size.Dy()*2))
		}
		renderDisplayImage(scaledBuf, frame)

		if !rainLoaded {
			copy(m.CPU.Memory[rainProg.StartAddress:], rainProg.Content)
			rainLoaded = true
			m.CPU.Exec(arch.JMP(rainProg.StartAddress))
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func convertRGBA(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	res := image.NewRGBA(bounds)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			res.Set(x, y, color.RGBAModel.Convert(src.At(x, y)))
		}
	}
	return res
}

func renderDisplayImage(buf *image.RGBA, src image.Image) {
	draw.NearestNeighbor.Scale(buf, buf.Bounds(), src, src.Bounds(), draw.Src, nil)
	ptr := uintptr(unsafe.Pointer(&buf.Pix[0]))
	size := buf.Bounds().Size()
	js.Global().Call("renderDisplay", ptr, len(buf.Pix), size.X, size.Y)
}

func nextKeyboardEvent() (present bool, keyCode devices.KeyCode, state devices.KeyState) {
	const bufferName = "kbEventsBuffer"
	buffer := js.Global().Get(bufferName)
	if buffer.Length() == 0 {
		return
	}
	event := buffer.Call("shift")
	code, down := event.Get("code").String(), event.Get("down").Bool()
	if keyCode, present = jsKeyCodes[code]; present {
		if down {
			state = devices.KeyStateDown
		}
	}
	return
}
