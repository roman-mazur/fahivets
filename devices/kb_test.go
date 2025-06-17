package devices

import "testing"

func TestReverseBits(t *testing.T) {
	for _, tc := range []struct {
		x, y byte
	}{
		{0, 0},
		{0xFF, 0xFF},
		{0x01, 0x80},
		{0x10, 0x08},
		{0x80, 0x01},
		{0x66, 0x66},
		{0x60, 0x06},
		{0x03, 0xC0},
	} {
		if got, want := reverseBits(tc.x), tc.y; got != want {
			t.Errorf("reverseBits(%08b) = %08b; want %08b", tc.x, got, want)
		}
	}
}
