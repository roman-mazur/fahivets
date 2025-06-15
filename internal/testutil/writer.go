package testutil

import (
	"bytes"
	"io"
	"testing"
)

// NewTestLogWriter creates an io.Writer that writes to testing.T logs and
// auto-cleans remaining buffer on test cleanup.
func NewTestLogWriter(t *testing.T) io.Writer {
	res := &testWriter{t: t}
	t.Cleanup(func() {
		if res.buf.Len() > 0 {
			res.t.Log(res.buf.String())
		}
	})
	return res
}

type testWriter struct {
	t   *testing.T
	buf bytes.Buffer
}

func (tw *testWriter) Write(p []byte) (n int, err error) {
	if i := bytes.IndexByte(p, '\n'); i != -1 {
		tw.buf.Write(p[:i])
		if tw.buf.Len() > 0 {
			tw.t.Log(tw.buf.String())
			tw.buf.Reset()
		}
		tw.buf.Write(p[i+1:])
	} else {
		tw.buf.Write(p)
	}
	return len(p), nil
}
