package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestFlatten(t *testing.T) {
	tts := []struct {
		in  string
		out []string
	}{
		{`null`, []string{`json	null`}},
		{`true`, []string{`json	true`}},
		{`false`, []string{`json	false`}},
		{`42`, []string{`json	42`}},
		{`55.99999`, []string{`json	55.99999`}},
		{`-0.55`, []string{`json	-0.55`}},
		{`"hi"`, []string{`json	"hi"`}},
		{`{}`, []string{}},
		{`{"key":"val"}`, []string{`json.key	"val"`}},
		{`{"!@#$%^&*()-_=+|{}привет,<>/?":1,"x x":2,"x][x":3,"x.x":4}`, []string{
			`json.!@#$%^&*()-_=+|{}привет,<>/?	1`,
			`json["x x"]	2`,
			`json["x][x"]	3`,
			`json["x.x"]	4`,
		}},
		{`{"key":{"key":"val"}}`, []string{`json.key.key	"val"`}},
		{`{"key":["val"]}`, []string{`json.key[0]	"val"`}},
		{
			`{"key":[{"key":"val"},{"key":["val"]},42]}`,
			[]string{
				`json.key[0].key	"val"`,
				`json.key[1].key[0]	"val"`,
				`json.key[2]	42`,
			},
		},
		{
			`[{"key":["val",42,null]}]`,
			[]string{
				`json[0].key[0]	"val"`,
				`json[0].key[1]	42`,
				`json[0].key[2]	null`,
			},
		},
		{
			`[1,"hi",true,null,{},[]]`,
			[]string{
				`json[0]	1`,
				`json[1]	"hi"`,
				`json[2]	true`,
				`json[3]	null`,
			},
		},
	}
	for _, tt := range tts {
		t.Run(tt.in, func(t *testing.T) {
			var (
				is  = is.New(t)
				out strings.Builder
			)

			flatten(bytes.NewReader([]byte(tt.in)), &out, false)

			lines := strings.Split(out.String(), "\n")
			lines = lines[:len(lines)-1] // discard last empty line
			is.Equal(tt.out, lines)
		})
	}
}

func BenchmarkFlatten(b *testing.B) {
	var (
		f, _      = os.Open("testdata/big.json")
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
		flatten(buf, io.Discard, false)
	}
}
