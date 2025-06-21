package devices

import (
	"image"
	"image/color"

	"rmazur.io/fahivets/arch"
)

type Display struct {
	mem []byte
}

func NewDisplay(cpu *arch.CPU) *Display {
	displayStart, displayEnd := arch.MemoryMappingRange(arch.MemDisplay12K)
	return &Display{
		mem: cpu.Memory[displayStart : displayEnd+1],
	}
}

func (c *Display) Image() image.Image {
	const height = 256
	width := len(c.mem) / height * 8
	img := image.NewGray(image.Rect(0, 0, width, height))
	for i, b := range c.mem {
		x, y := i/height, i%height
		for s := range 8 {
			img.Set(x*8+7-s, y, color.Gray{Y: (b >> s) & 1 * 255})
		}
	}
	return img
}
