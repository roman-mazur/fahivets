//go:build js

package main

import (
	"bytes"
	_ "embed"
	"image"
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
	frames := make(chan *image.RGBA, 1)
	go runSimulation(m, frames)
	jsSetupAnimationFrames(frames, m.Keyboard)

	var done chan struct{}
	<-done
}

func runSimulation(m *fahivets.Computer, frames chan *image.RGBA) {
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

	var scaledBufEdit, scaleBufSent *image.RGBA

	for {
		const maxSteps = 16_000
		for range maxSteps {
			_, err := m.Step()
			if err != nil {
				log.Println(err)
				return
			}

		}

		frame := m.Display.Image()
		if scaledBufEdit == nil {
			size := frame.Bounds()
			scaledBufEdit = image.NewRGBA(image.Rect(0, 0, size.Dx()*2, size.Dy()*2))
			scaleBufSent = image.NewRGBA(image.Rect(0, 0, size.Dx()*2, size.Dy()*2))
		}
		draw.NearestNeighbor.Scale(scaledBufEdit, scaledBufEdit.Bounds(), frame, frame.Bounds(), draw.Src, nil)
		// Flip the frame buffers.
		select {
		case frames <- scaleBufSent:
			scaledBufEdit, scaleBufSent = scaleBufSent, scaledBufEdit
		default:
		}

		if !rainLoaded {
			copy(m.CPU.Memory[rainProg.StartAddress:], rainProg.Content)
			rainLoaded = true
			m.CPU.Exec(arch.JMP(rainProg.StartAddress))
		}

		time.Sleep(10 * time.Millisecond) // TODO: Replace with proper clock.
	}
}

func jsSetupAnimationFrames(frames chan *image.RGBA, keyboard *devices.Keyboard) {
	var jsHandler js.Func
	jsHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		select {
		case frame := <-frames:
			renderDisplayImage(frame)
		default:
		}

		kbEvent, keyCode, keyState := nextKeyboardEvent()
		if kbEvent {
			log.Println("kb event", keyCode, keyState)
			m.Keyboard.Event(keyCode, keyState)
		}

		js.Global().Call("requestAnimationFrame", jsHandler)
		return nil
	})
	js.Global().Call("requestAnimationFrame", jsHandler)
}

func renderDisplayImage(buf *image.RGBA) {
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
