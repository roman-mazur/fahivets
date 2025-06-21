package arch

func EncodeInstructions(program []Instruction, out []byte) {
	dst := out
	for _, cmd := range program {
		if cmd.Encode == nil {
			panic("no encode defined for " + cmd.Name)
		}
		if int(cmd.Size) > len(dst) {
			panic("output memory is not sufficient to fit the program, failed at " + cmd.Name)
		}
		cmd.Encode(dst)
		dst = dst[cmd.Size:]
	}
}
