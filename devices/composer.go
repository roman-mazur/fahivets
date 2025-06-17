package devices

type IoController interface {
	SendA(byte)
	SendB(byte)
	SendCLow(byte)
	SendCHigh(byte)
}

type IoSendFunc func(byte)

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
