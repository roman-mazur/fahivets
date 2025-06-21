package devices

import "fmt"

type IoController interface {
	SendA(byte)
	SendB(byte)
	SendCLow(byte)
	SendCHigh(byte)
}

type IoSendFunc func(byte)

type ComposedIoController struct {
	PortA, PortB, PortCLow, PortCHigh IoSendFunc
}

func (cic *ComposedIoController) SendA(v byte)     { sendOrPanicUnused(cic.PortA, "A", v) }
func (cic *ComposedIoController) SendB(v byte)     { sendOrPanicUnused(cic.PortB, "B", v) }
func (cic *ComposedIoController) SendCLow(v byte)  { sendOrPanicUnused(cic.PortCLow, "CLow", v) }
func (cic *ComposedIoController) SendCHigh(v byte) { sendOrPanicUnused(cic.PortCHigh, "CHigh", v) }

func sendOrPanicUnused(send IoSendFunc, name string, v byte) {
	if send == nil {
		panic(fmt.Errorf("port %s is supposed to be unused", name))
	}
	send(v)
}

// PortComposer exists to combine multiple devices connected to the same port of the IO controller.
// It's not a real device, but an extra layer we need as IoController interface allows working with the whole port only,
// not individual pins.
type PortComposer struct {
	dstSend IoSendFunc
	c       chan maskedValue
}

// NewPortComposer creates a new PortComposer.
func NewPortComposer(dst IoSendFunc) *PortComposer {
	pc := &PortComposer{
		dstSend: dst,
		c:       make(chan maskedValue),
	}
	go pc.compose()
	return pc
}

func (pc *PortComposer) ShutDown() {
	close(pc.c)
}

func (pc *PortComposer) MaskedSend(mask byte) IoSendFunc {
	return func(value byte) {
		pc.c <- maskedValue{v: value, mask: mask}
	}
}

type maskedValue struct{ v, mask byte }

func (pc *PortComposer) compose() {
	var composedValue byte
	for data := range pc.c {
		composedValue = (composedValue &^ data.mask) | (data.v & data.mask)
		pc.dstSend(composedValue)
	}
}
