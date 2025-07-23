package arch

import "testing"

func TestPSW_String(t *testing.T) {
	var psw PSW
	psw.Z = true
	psw.A = true
	if psw.String() != "ACPSZ: 10001" {
		t.Fail()
	}
}

func TestParityBit(t *testing.T) {
	var psw PSW

	psw.SetZSPC(0)
	if !psw.P {
		t.Error("no parity for 0")
	}

	psw.SetZSPC(1)
	if psw.P {
		t.Error("parity for 1")
	}
	psw.SetZSPC(2)
	if psw.P {
		t.Error("parity for 2")
	}

	psw.SetZSPC(3)
	if !psw.P {
		t.Error("no parity for 3")
	}
}
