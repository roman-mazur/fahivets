package fahivets

import (
	"rmazur.io/fahivets/arch"
	"rmazur.io/fahivets/devices"
)

type Computer struct {
	CPU      arch.CPU
	Keyboard *devices.Keyboard
	Display  *devices.Display

	ioCtl         *arch.IoController
	portBComposer *devices.PortComposer
}

func NewComputer() *Computer {
	var c Computer
	c.ioCtl = arch.InitIoController(&c.CPU)

	c.portBComposer = devices.NewPortComposer(c.ioCtl.SendB)

	c.Keyboard = devices.NewKeyboard(&devices.ComposedIoController{
		PortA: c.ioCtl.SendA,
		// Keyboard is not connected to lower 2 bites.
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
	cmd, err = c.CPU.Step()
	if err != nil {
		return
	}
	c.ioCtl.Sync()
	return
}
