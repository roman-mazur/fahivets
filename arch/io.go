package arch

import (
	"sync"
)

// IoController represents К580ВВ55 (the analog of Intel 8255 microcontroller) that controls
// interactions with the keyboard and other devices.
// See https://en.wikipedia.org/wiki/Intel_8255
type IoController struct {
	mem []byte

	a, b, cl, ch chan byte

	umu     sync.Mutex // updates mutex
	updates [3]byte
}

const (
	portA = iota
	portB
	portC
	controlFlags
)

func InitIoController(m *CPU) *IoController {
	addr, _ := MemoryMapping(MemRegisters2K)
	res := &IoController{
		mem: m.Memory[addr : addr+4],

		a:  make(chan byte, 1),
		b:  make(chan byte, 1),
		cl: make(chan byte, 1),
		ch: make(chan byte, 1),
	}
	if len(res.mem) != 4 {
		panic("bad memory size")
	}
	return res
}

// Sync propagates values between the [CPU.Memory] and controlled device that integrates with the ports A/B/C.
// It is supposed to be called in the routine that works with attached CPU.
func (c *IoController) Sync() {
	ctl := c.mem[controlFlags]
	if mask(ctl, 0x80) {
		switch ioMode := (ctl >> 5) & 0x3; ioMode {
		case 0:
			c.syncSimpleIO(ctl)
		case 1:
			c.syncStrobedIO(ctl)
		default:
			c.syncStrobedBidirectIO(ctl)
		}
	} else {
		c.syncBSR(ctl)
	}
}

// ReceiveA obtains a value provided by the CPU after Sync is called.
// This method will block until the value is provided.
// The value is not provided until the port is set to the output mode by the CPU.
func (c *IoController) ReceiveA() byte { return <-c.a }

// ReceiveB obtains a value provided by the CPU after Sync is called.
// This method will block until the value is provided.
// The value is not provided until the port is set to the output mode by the CPU.
func (c *IoController) ReceiveB() byte { return <-c.b }

// ReceiveCLow obtains a value provided by the CPU after Sync is called.
// This method will block until the value is provided.
// The value is not provided until the port is set to the output mode by the CPU.
func (c *IoController) ReceiveCLow() byte { return <-c.cl }

// ReceiveCHigh obtains a value provided by the CPU after Sync is called.
// This method will block until the value is provided.
// The value is not provided until the port is set to the output mode by the CPU.
func (c *IoController) ReceiveCHigh() byte { return <-c.ch }

// SendA sets the value that should be visible to the CPU after the next Sync.
// The value is used only if the port is set to the input mode.
func (c *IoController) SendA(v byte) { c.update(portA, v) }

// SendB sets the value that should be visible to the CPU after the next Sync.
// The value is used only if the port is set to the input mode.
func (c *IoController) SendB(v byte) { c.update(portB, v) }

// SendCLow sets the value that should be visible to the CPU after the next Sync.
// The value is used only if the port is set to the input mode.
func (c *IoController) SendCLow(v byte) { c.updateC(true, v) }

// SendCHigh sets the value that should be visible to the CPU after the next Sync.
// The value is used only if the port is set to the input mode.
func (c *IoController) SendCHigh(v byte) { c.updateC(false, v) }

func (c *IoController) update(i int, v byte) {
	c.umu.Lock()
	c.updates[i] = v
	c.umu.Unlock()
}

func (c *IoController) updateC(low bool, v byte) {
	c.umu.Lock()
	if low {
		c.updates[portC] = (c.updates[portC] & 0xF0) | (v & 0x0F)
	} else {
		c.updates[portC] = (c.updates[portC] & 0x0F) | (v << 4)
	}
	c.umu.Unlock()
}

func (c *IoController) syncSimpleIO(ctl byte) {
	c.umu.Lock()
	c.lSyncPortSimple(portA, mask(ctl, 0x10))
	c.lSyncPortSimple(portB, mask(ctl, 0x02))
	c.lSyncPortCSimple(true, mask(ctl, 0x01))
	c.lSyncPortCSimple(false, mask(ctl, 0x08))
	c.umu.Unlock()
}

func (c *IoController) lSyncPortSimple(port int, inputMode bool) {
	if inputMode {
		c.mem[port] = c.updates[port]
	} else {
		var conn chan<- byte
		switch port {
		case portA:
			conn = c.a
		case portB:
			conn = c.b
		default:
			panic("bad port")
		}
		sendOutValue(conn, c.mem[port])
	}
}

func (c *IoController) lSyncPortCSimple(low bool, inputMode bool) {
	if inputMode {
		if low {
			c.mem[portC] = (c.mem[portC] & 0xF0) | (c.updates[portC] & 0x0F)
		} else {
			c.mem[portC] = (c.mem[portC] & 0x0F) | (c.updates[portC] & 0xF0)
		}
	} else {
		if low {
			sendOutValue(c.cl, c.mem[portC]&0x0F)
		} else {
			sendOutValue(c.ch, c.mem[portC]>>4)
		}
	}
}

func (c *IoController) syncStrobedIO(byte) {
	panic("not implemented")
}

func (c *IoController) syncStrobedBidirectIO(byte) {
	panic("not implemented")
}

func (c *IoController) syncBSR(ctl byte) {
	selector := (ctl >> 1) & 0x07
	if mask(ctl, 1) {
		c.mem[portC] |= 1 << selector
	} else {
		c.mem[portC] &^= 1 << selector
	}
	sendOutValue(c.ch, c.mem[portC]>>4)
	sendOutValue(c.cl, c.mem[portC]&0x0F)
}

func sendOutValue(conn chan<- byte, val byte) {
	select {
	case conn <- val:
	default:
	}
}
