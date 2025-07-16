package fahivets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadRks(t *testing.T) {
	const progsPath = "testdata/progs"
	progsList, err := os.ReadDir(progsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, progsEntry := range progsList {
		if strings.HasSuffix(progsEntry.Name(), ".rks") {
			t.Run(progsEntry.Name(), func(t *testing.T) {
				f, err := os.Open(filepath.Join(progsPath, progsEntry.Name()))
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = f.Close() })
				data, err := ReadRks(f)
				if err != nil {
					t.Error(err)
				}
				if data.EndAddress <= data.StartAddress {
					t.Errorf("data.EndAddress(%x) <= data.StartAddress(%x)", data.EndAddress, data.StartAddress)
				}
				if len(data.Content) == 0 {
					t.Errorf("len(data.Content) == 0")
				}
			})
		}
	}
}
