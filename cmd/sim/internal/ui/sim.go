//go:build js

package main

import (
	_ "embed"
	"image"
	"image/color"
	"log"
	"syscall/js"
	"time"
	"unsafe"

	"rmazur.io/fahivets"
	"rmazur.io/fahivets/arch"
	"rmazur.io/fahivets/devices"
)

var (
	//go:embed progs/bootloader.rom
	bootloader []byte
	//go:embed progs/monitor.rom
	monitor []byte

	m = fahivets.NewComputer()
)

func main() {
	// Note: it should be possible to minimize the exe size if we link
	// programs directly to the CPU.Memory.
	romStart := arch.MemoryMapping(arch.MemROM2K)
	copy(m.CPU.Memory[romStart:], bootloader)
	copy(m.CPU.Memory[arch.MemoryMapping(arch.MemROMExtra12K):], monitor)
	m.CPU.PC = uint16(romStart)

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

		img := convertRGBA(devices.NewDisplay(&m.CPU).Image())
		renderDisplayImage(img)
		time.Sleep(100 * time.Millisecond)
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

func renderDisplayImage(img *image.RGBA) {
	size := img.Bounds().Size()
	ptr := uintptr(unsafe.Pointer(&img.Pix[0]))
	js.Global().Call("renderDisplay", ptr, len(img.Pix), size.X, size.Y)
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
