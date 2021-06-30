package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestUnflatten(t *testing.T) {
	tts := []struct {
		in  string
		out string
	}{
		{"json.key = 42", `{"key":42}`},
		{"json[11] = 42", `[42]`},
		{`json["11"] = 42`, `{"11":42}`},
		{"json.k1.k2[11] = 42", `{"k1":{"k2":[42]}}`},
		{"json.k1 = \"41\"\njson.k2 = 42\njson.k3 = 43", `{"k1":"41","k2":42,"k3":43}`},
		{"json.k1.sub = 41\njson.k2.sub = 42\njson.k2.sub2 = 422\njson.k3 = 43", `{"k1":{"sub":41},"k2":{"sub":42,"sub2":422},"k3":43}`},
		{"json[1].sub = 41\njson[2].sub = 42\njson[2].sub2 = 422\njson[33] = 43", `[{"sub":41},{"sub":42,"sub2":422},43]`},
		{
			strings.Join([]string{
				`json.!@#$%^&*()-_=+|{}привет,<>/? = 1`,
				`json["x x"] = 2`,
				`json["x][x"] = 3`,
				`json["x.x"] = 4`,
			}, "\n"),
			`{"!@#$%^&*()-_=+|{}привет,<>/?":1,"x x":2,"x][x":3,"x.x":4}`,
		},
	}
	for _, tt := range tts {
		t.Run(tt.out, func(t *testing.T) {
			var (
				is = is.New(t)
				w  strings.Builder
			)
			unflatten(strings.NewReader(tt.in), &w, false)
			is.Equal(tt.out, w.String())
		})
	}
}

func BenchmarkUnflatten(b *testing.B) {
	var (
		f, _      = os.Open("testdata/big.fj")
		data, err = io.ReadAll(f)
		buf       = bytes.NewReader(data)
	)

	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = buf.Seek(0, io.SeekStart)
		unflatten(buf, io.Discard, false)
	}
}
