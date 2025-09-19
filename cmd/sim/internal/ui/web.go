//go:build js

package main

import (
	"image"
	"log"
	"syscall/js"
	"unsafe"

	"rmazur.io/fahivets/devices"
)

func makeUiWorld() UiWorld {
	return &jsUiWorld{root: js.Global()}
}

type jsUiWorld struct {
	root js.Value
}

func (w *jsUiWorld) ConsumeDisplayFrames(f func() image.Image) {
	const callName = "requestAnimationFrame"

	var jsHandler js.Func
	jsHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		renderDisplayImage(imageToRGBA(f()))

		w.root.Call(callName, jsHandler)
		return nil
	})

	w.root.Call(callName, jsHandler)
}

func (w *jsUiWorld) ConnectKeyboard(keyboard *devices.Keyboard) {
	log.Println("Connecting keyboard...")

	docEl := w.root.Get("document").Get("documentElement")
	const callName = "addEventListener"
	docEl.Call(callName, "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		handleKeyboardEvent(jsEventCode(args), true, keyboard)
		return nil
	}))
	docEl.Call(callName, "keyup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		handleKeyboardEvent(jsEventCode(args), false, keyboard)
		return nil
	}))
}

func jsEventCode(args []js.Value) string { return args[0].Get("code").String() }

func handleKeyboardEvent(code string, down bool, keyboard *devices.Keyboard) {
	if keyCode, present := jsKeyCodes[code]; present {
		state := devices.KeyStateUp
		if down {
			state = devices.KeyStateDown
		}
		log.Printf("code %s, keycode %v, state %v", code, keyCode, state)
		keyboard.Event(keyCode, state)
	} else if down {
		log.Println("no keyboard mapping for", code)
	}
}

func renderDisplayImage(buf *image.RGBA) {
	ptr := uintptr(unsafe.Pointer(&buf.Pix[0]))
	size := buf.Bounds().Size()
	js.Global().Call("renderDisplay", ptr, len(buf.Pix), size.X, size.Y)
}

func imageToRGBA(img image.Image) *image.RGBA {
	rgba, ok := img.(*image.RGBA)
	if ok {
		return rgba
	}

	rgba = image.NewRGBA(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	return rgba
}

var jsKeyCodes = map[string]devices.KeyCode{
	"F1":  devices.MatrixKeyCode(5, 11),
	"F2":  devices.MatrixKeyCode(5, 10),
	"F3":  devices.MatrixKeyCode(5, 9),
	"F4":  devices.MatrixKeyCode(5, 8),
	"F5":  devices.MatrixKeyCode(5, 7),
	"F6":  devices.MatrixKeyCode(5, 6),
	"F7":  devices.MatrixKeyCode(5, 5),
	"F8":  devices.MatrixKeyCode(5, 4),
	"F9":  devices.MatrixKeyCode(5, 3),
	"F10": devices.MatrixKeyCode(5, 2),
	"F11": devices.MatrixKeyCode(5, 1),
	"F12": devices.MatrixKeyCode(5, 0),

	"IntlBackslash": devices.MatrixKeyCode(4, 11),
	"Digit1":        devices.MatrixKeyCode(4, 10),
	"Digit2":        devices.MatrixKeyCode(4, 9),
	"Digit3":        devices.MatrixKeyCode(4, 8),
	"Digit4":        devices.MatrixKeyCode(4, 7),
	"Digit5":        devices.MatrixKeyCode(4, 6),
	"Digit6":        devices.MatrixKeyCode(4, 5),
	"Digit7":        devices.MatrixKeyCode(4, 4),
	"Digit8":        devices.MatrixKeyCode(4, 3),
	"Digit9":        devices.MatrixKeyCode(4, 2),
	"Digit0":        devices.MatrixKeyCode(4, 1),
	"Equal":         devices.MatrixKeyCode(4, 0),

	"KeyQ":         devices.MatrixKeyCode(3, 11),
	"KeyW":         devices.MatrixKeyCode(3, 10),
	"KeyE":         devices.MatrixKeyCode(3, 9),
	"KeyR":         devices.MatrixKeyCode(3, 8),
	"KeyT":         devices.MatrixKeyCode(3, 7),
	"KeyY":         devices.MatrixKeyCode(3, 6),
	"KeyU":         devices.MatrixKeyCode(3, 5),
	"KeyI":         devices.MatrixKeyCode(3, 4),
	"KeyO":         devices.MatrixKeyCode(3, 3),
	"KeyP":         devices.MatrixKeyCode(3, 2),
	"BracketLeft":  devices.MatrixKeyCode(3, 1),
	"BracketRight": devices.MatrixKeyCode(3, 0),

	"KeyA":      devices.MatrixKeyCode(2, 11),
	"KeyS":      devices.MatrixKeyCode(2, 10),
	"KeyD":      devices.MatrixKeyCode(2, 9),
	"KeyF":      devices.MatrixKeyCode(2, 8),
	"KeyG":      devices.MatrixKeyCode(2, 7),
	"KeyH":      devices.MatrixKeyCode(2, 6),
	"KeyJ":      devices.MatrixKeyCode(2, 5),
	"KeyK":      devices.MatrixKeyCode(2, 4),
	"KeyL":      devices.MatrixKeyCode(2, 3),
	"Semicolon": devices.MatrixKeyCode(2, 2),
	"Quote":     devices.MatrixKeyCode(2, 1),
	"Backslash": devices.MatrixKeyCode(2, 0),

	"KeyZ":      devices.MatrixKeyCode(1, 11),
	"KeyX":      devices.MatrixKeyCode(1, 10),
	"KeyC":      devices.MatrixKeyCode(1, 9),
	"KeyV":      devices.MatrixKeyCode(1, 8),
	"KeyB":      devices.MatrixKeyCode(1, 7),
	"KeyN":      devices.MatrixKeyCode(1, 6),
	"KeyM":      devices.MatrixKeyCode(1, 5),
	"Comma":     devices.MatrixKeyCode(1, 4),
	"Period":    devices.MatrixKeyCode(1, 3),
	"Slash":     devices.MatrixKeyCode(1, 2),
	"Backquote": devices.MatrixKeyCode(1, 1),
	"Backspace": devices.MatrixKeyCode(1, 0),

	"ShiftLeft":  devices.MatrixKeyCode(0, 11),
	"MetaLeft":   devices.MatrixKeyCode(0, 10),
	"ArrowUp":    devices.MatrixKeyCode(0, 9),
	"ArrowDown":  devices.MatrixKeyCode(0, 8),
	"Space":      devices.MatrixKeyCode(0, 5),
	"ArrowLeft":  devices.MatrixKeyCode(0, 4),
	"Enter":      devices.MatrixKeyCode(0, 3),
	"ArrowRight": devices.MatrixKeyCode(0, 2),
	"MetaRight":  devices.MatrixKeyCode(0, 1),
	"AltRight":   devices.MatrixKeyCode(0, 0),
}
