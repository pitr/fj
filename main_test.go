package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/matryer/is"
)

func TestE2E(t *testing.T) {
	is := is.New(t)
	buf := new(bytes.Buffer)
	out := new(bytes.Buffer)

	f, err := os.Open("testdata/big.json")
	is.NoErr(err)

	flatten(f, buf, false)
	unflatten(buf, out, false)

	_, err = f.Seek(0, io.SeekStart)
	is.NoErr(err)

	original, err := io.ReadAll(f)
	is.NoErr(err)

	buf.Reset()
	err = json.Compact(buf, original)
	is.NoErr(err)

	is.Equal(buf.String(), out.String())
}
