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
