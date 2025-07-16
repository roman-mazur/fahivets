package fahivets

import (
	"encoding/binary"
	"fmt"
	"io"
)

type RksData struct {
	StartAddress, EndAddress uint16
	Checksum                 uint16
	Content                  []byte
}

func ReadRks(in io.Reader) (data RksData, err error) {
	var (
		header [4]byte
		n      int
	)
	n, err = in.Read(header[:])
	if err != nil {
		err = fmt.Errorf("failed to read the hedader: %w", err)
		return
	}
	if n != 4 {
		err = fmt.Errorf("failed to read the hedader: expected 4 bytes, got %d", n)
		return
	}

	data.StartAddress = binary.LittleEndian.Uint16(header[0:2])
	data.EndAddress = binary.LittleEndian.Uint16(header[2:])

	data.Content, err = io.ReadAll(in)
	if err != nil {
		err = fmt.Errorf("failed to read the content: %w", err)
		return
	}
	if len(data.Content) < 2 {
		err = fmt.Errorf("not enoung data for the cheksum")
		return
	}
	data.Checksum = binary.LittleEndian.Uint16(data.Content[len(data.Content)-2:])
	data.Content = data.Content[:len(data.Content)-2]
	expectedLen := int(data.EndAddress-data.StartAddress) + 1
	if len(data.Content) != expectedLen {
		err = fmt.Errorf("content length does not match the addresses")
	}
	return
}
