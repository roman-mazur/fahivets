package fahivets

import (
	"time"

	"rmazur.io/fahivets/arch"
	"rmazur.io/fahivets/devices"
)

type Computer struct {
	CPU      arch.CPU
	Keyboard *devices.Keyboard
	Display  *devices.Display

	ioCtl         *arch.IoController
	portBComposer *devices.PortComposer

	cyclesSinceSleep int
	lastSleep        time.Time
}

func NewComputer() *Computer {
	var c Computer
	c.ioCtl = arch.InitIoController(&c.CPU)

	c.portBComposer = devices.NewPortComposer(c.ioCtl.SendB)

	c.Keyboard = devices.NewKeyboard(&devices.ComposedIoController{
		PortA: c.ioCtl.SendA,
		// Keyboard is not connected to lower 2 bits.
		PortB:    c.portBComposer.MaskedSend(0xFC),
		PortCLow: c.ioCtl.SendCLow,
	})

	c.Display = devices.NewDisplay(&c.CPU)
	return &c
}

func (c *Computer) Shutdown() {
	c.Keyboard.ShutDown()
	c.portBComposer.ShutDown()
}

func (c *Computer) Step() (cmd arch.Instruction, err error) {
	var cycles int
	cmd, cycles, err = c.CPU.Step()
	if err != nil {
		return
	}
	c.ioCtl.Sync()

	c.cyclesSinceSleep += cycles
	return
}

func (c *Computer) SimSleep() {
	const frequency = 2_000_000 // 2MHz

	simDuration := time.Second * time.Duration(c.cyclesSinceSleep) / frequency
	passed := time.Since(c.lastSleep)

	if simDuration < 10*time.Millisecond { // TODO: fix.
		return
	}

	diff := simDuration - passed
	if diff < 0 {
		c.lastSleep = time.Now()
		return
	}
	time.Sleep(diff)
	c.cyclesSinceSleep = 0
	c.lastSleep = time.Now()
}
