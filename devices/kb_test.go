package devices

import (
	"testing"
	"time"

	"rmazur.io/fahivets/arch"
	"rmazur.io/fahivets/internal/testutil"
)

func TestKeyboard(t *testing.T) {
	var (
		cpu    arch.CPU
		ioCtrl = arch.InitIoController(&cpu)
	)

	portBComposer := NewPortComposer(ioCtrl.SendB)
	t.Cleanup(portBComposer.ShutDown)

	kbController := ComposedIoController{
		PortA:    ioCtrl.SendA,
		PortB:    portBComposer.MaskedSend(0xFC), // Ignore lower 2 bits.
		PortCLow: ioCtrl.SendCLow,
	}

	kb := NewKeyboard(&kbController)
	t.Cleanup(kb.ShutDown)

	someKey := MatrixKeyCode(1, 5)

	keyEventFinished := make(chan struct{})
	go func() {
		kb.Event(someKey, KeyStateDown)
		close(keyEventFinished)
	}()

	userMemStart, userMemEnd := arch.MemoryMappingRange(arch.MemUser16K)

	program := arch.Program{
		// Program the IO controller.
		arch.LXI(arch.RegisterPairHL, arch.MemoryMapping(arch.MemRegisters2K)+3), // Control byte.
		arch.MVI(arch.RegisterSelMemory, 0x93),

		// Check columns.
		arch.LXI(arch.RegisterPairHL, arch.MemoryMapping(arch.MemRegisters2K)), // Port A.
		arch.MOV(arch.RegisterSelA, arch.RegisterSelMemory),
		arch.CPI(0xDF),
		arch.JCnd(arch.ConditionCodeNZ, 0), // Loop if no match.
		arch.LXI(arch.RegisterPairHL, arch.MemoryMapping(arch.MemRegisters2K)+2), // Port C.
		arch.MOV(arch.RegisterSelA, arch.RegisterSelMemory),
		arch.ANI(0x0F), // Lower part only.
		arch.CPI(0x0F),
		arch.JCnd(arch.ConditionCodeNZ, 0), // Loop if no match.

		// Check rows.
		arch.LXI(arch.RegisterPairHL, arch.MemoryMapping(arch.MemRegisters2K)+1), // Port B.
		arch.MOV(arch.RegisterSelA, arch.RegisterSelMemory),
		arch.ANI(0xFC), // Ignore first 2 bits.
		arch.CPI(0xBC),
		arch.JCnd(arch.ConditionCodeNZ, 0), // Loop if no match.

		// Indicate the success.
		arch.LXI(arch.RegisterPairHL, userMemEnd-1),
		arch.MVI(arch.RegisterSelMemory, 1),

		// NOP loop.
		arch.NOP(),
		arch.JMP(0), // Placeholder.
	}
	program[len(program)-1] = arch.JMP(uint16(len(program) - 2))
	t.Logf("Program (%d instructions):\n%s", len(program), program)

	arch.EncodeInstructions(program, cpu.Memory[userMemStart:userMemEnd/2])

	ioFinished := false

	for n := range 100 {
		if ioFinished && cpu.Memory[userMemEnd-1] == 1 {
			t.Logf("Reached the expected state after %d steps", n+1)
			return
		}

		cmd, err := cpu.Step()
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("after %s: %s", cmd.Name, &cpu)
		ioCtrl.Sync()

		select {
		case <-keyEventFinished:
			ioFinished = true
		default:
			// Continue CPU cycle. Avoid real load.
			time.Sleep(10 * time.Millisecond)
		}
	}

	t.Error("program didn't complete as expected")
	t.Log("CPU")
	t.Log(&cpu)
	t.Log("MEMORY")
	_ = cpu.Memory.Dump(testutil.NewTestLogWriter(t), userMemEnd-16, userMemEnd)
}

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

func TestMatrixKeyCode(t *testing.T) {
	r, c := MatrixKeyCode(4, 2).matrix()
	if r != 4 || c != 2 {
		t.Errorf("MatrixKeyCode(4, 2).matrix() = (%d, %d); want (%d, %d)", r, c, 4, 2)
	}
}
