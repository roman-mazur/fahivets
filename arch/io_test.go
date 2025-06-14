package arch

import (
	"testing"
)

func TestIoController(t *testing.T) {
	var (
		cpu CPU
		ioc = InitIoController(&cpu)
	)

	baseAddress := MemoryMapping(MemRegisters2K)

	type commValue struct {
		name string
		val  byte
	}

	type portInfo struct {
		offset uint16
		recvF  func() byte
		sendF  func(byte)
	}

	ports := map[string]portInfo{
		"A": {offset: 0, recvF: ioc.ReceiveA, sendF: ioc.SendA},
		"B": {offset: 1, recvF: ioc.ReceiveB, sendF: ioc.SendB},

		"C":  {offset: 2},
		"CL": {recvF: ioc.ReceiveCLow, sendF: ioc.SendCLow},
		"CH": {recvF: ioc.ReceiveCHigh, sendF: ioc.SendCHigh},
	}

	port := func(name string) portInfo {
		t.Helper()
		res, ok := ports[name]
		if !ok {
			t.Fatalf("port %s not found", name)
		}
		return res
	}

	for _, tc := range []struct {
		name      string
		ctl       byte
		cpuWrites []commValue
		ioSend    []commValue
		ioRecv    []commValue
		cpuReads  []commValue
	}{
		{
			name: "io/B=input/other=output",
			ctl:  0x82,
			cpuWrites: []commValue{
				{name: "A", val: 0x55},
				{name: "C", val: 0xFF},
			},
			ioSend: []commValue{
				{name: "B", val: 0x42},
			},
			ioRecv: []commValue{
				{name: "A", val: 0x55},
				{name: "CL", val: 0x0F},
				{name: "CH", val: 0x0F},
			},
			cpuReads: []commValue{
				{name: "B", val: 0x42},
			},
		},
		{
			name: "bsr/simple",
			ctl:  0x0B, // Set 5th bit of C.
			cpuWrites: []commValue{
				{name: "C", val: 1},
			},
			ioRecv: []commValue{
				{name: "CL", val: 1},
				{name: "CH", val: 2},
			},
			cpuReads: []commValue{
				{name: "C", val: 0x21},
			},
		},
		{
			name: "io/all=input",
			ctl:  0b10011011,
			cpuWrites: []commValue{ // Will be overwritten.
				{name: "A", val: 0xFF},
				{name: "B", val: 0xFF},
				{name: "C", val: 0xFF},
			},
			ioSend: []commValue{
				{name: "A", val: 0x11},
				{name: "B", val: 0x22},
				{name: "CL", val: 0x13}, // Higher bits ignored.
				{name: "CH", val: 0x14},
			},
			cpuReads: []commValue{
				{name: "A", val: 0x11},
				{name: "B", val: 0x22},
				{name: "C", val: 0x43},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cpu.Memory[baseAddress+3] = tc.ctl

			for _, w := range tc.cpuWrites {
				cpu.Memory[baseAddress+port(w.name).offset] = w.val
			}

			for _, s := range tc.ioSend {
				port(s.name).sendF(s.val)
			}

			ioc.Sync()

			for _, r := range tc.ioRecv {
				val := port(r.name).recvF()
				if val != r.val {
					t.Errorf("got %v receiving from %s; want %v", val, r.name, r.val)
				}
			}
			for _, r := range tc.cpuReads {
				val := cpu.Memory[baseAddress+port(r.name).offset]
				if val != r.val {
					t.Errorf("got %v reading for %s; want %v", val, r.name, r.val)
				}
			}
		})
	}
}
