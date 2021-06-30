package main

import (
	"fmt"
	"io"

	json "github.com/pitr/jsontokenizer"
	"github.com/valyala/bytebufferpool"
)

var (
	tk = json.NewWithSize(nil, 4*1024)

	bTrue   = []byte("true\n")
	bFalse  = []byte("false\n")
	bNull   = []byte("null\n")
	bDot    = []byte{'.'}
	bPrefix = []byte("json")
	bSep    = []byte(" = ")
)

func flatten(in io.Reader, out io.Writer, stream bool) {
	var (
		keys        []*bytebufferpool.ByteBuffer // stack of object keys (or nil)
		arrs        []int                        // stack of array indices
		key, keyRaw *bytebufferpool.ByteBuffer
		err         error
		t           json.TokType
		last        int
	)

	if stream {
		keys = append(keys, nil)
		arrs = append(arrs, 0)
	}

	tk.Reset(in)

	for {
		t, err = tk.Token()
		if err == io.EOF {
			return
		}
		fail(err)
		switch t {
		case json.TokArrayOpen:
			keys = append(keys, nil)
			arrs = append(arrs, 0)
		case json.TokObjectOpen:
			keys = append(keys, nil)
			arrs = append(arrs, -1)
		case json.TokArrayClose, json.TokObjectClose:
			last = len(keys) - 1
			if last < 0 {
				fail(fmt.Errorf("invalid JSON"))
			}
			keys = keys[:last]
			arrs = arrs[:last]
			endObjArr(keys, arrs)
		case json.TokTrue, json.TokFalse:
			printPrefix(out, keys, arrs)
			if t == json.TokTrue {
				_, err = out.Write(bTrue)
			} else {
				_, err = out.Write(bFalse)
			}
			fail(err)

			endObjArr(keys, arrs)
		case json.TokNumber:
			printPrefix(out, keys, arrs)
			_, err = tk.ReadNumber(out)
			fail(err)
			_, err = out.Write(nl)
			fail(err)

			endObjArr(keys, arrs)
		case json.TokString:
			last = len(keys) - 1
			if last >= 0 && keys[last] == nil && arrs[last] == -1 {
				key = bytebufferpool.Get()
				_, err = tk.ReadString(key)
				fail(err)

				if !validKey(key) {
					key, keyRaw = bytebufferpool.Get(), key
					err = key.WriteByte('[')
					fail(err)
					err = key.WriteByte('"')
					fail(err)
					_, err = key.Write(keyRaw.B)
					fail(err)
					err = key.WriteByte('"')
					fail(err)
					err = key.WriteByte(']')
					fail(err)
					bytebufferpool.Put(keyRaw)
				}

				keys[last] = key

			} else {
				printPrefix(out, keys, arrs)
				_, err = out.Write(quote)
				fail(err)
				_, err = tk.ReadString(out)
				fail(err)
				_, err = out.Write(quote)
				fail(err)
				_, err = out.Write(nl)
				fail(err)
				endObjArr(keys, arrs)
			}
		case json.TokNull:
			printPrefix(out, keys, arrs)
			_, err = out.Write(bNull)
			fail(err)
			endObjArr(keys, arrs)
		}
	}
}

func endObjArr(keys []*bytebufferpool.ByteBuffer, arrs []int) {
	last := len(keys) - 1
	if last < 0 {
		return
	}
	key := keys[last]

	if key != nil {
		bytebufferpool.Put(key)
		keys[last] = nil
	} else {
		arrs[last]++
	}
}

func printPrefix(out io.Writer, keys []*bytebufferpool.ByteBuffer, arrs []int) {
	_, err := out.Write(bPrefix)
	fail(err)

	for ix, key := range keys {
		switch {
		case key == nil:
			_, err = fmt.Fprintf(out, "[%d]", arrs[ix])
			fail(err)
		case key.B[0] == '[':
			_, err = out.Write(key.B)
			fail(err)
		default:
			_, err = out.Write(bDot)
			fail(err)
			_, err = out.Write(key.B)
			fail(err)
		}
	}
	_, err = out.Write(bSep)
	fail(err)
}

func validKey(key *bytebufferpool.ByteBuffer) bool {
	if len(key.B) == 0 {
		return false
	}
	for _, r := range key.String() {
		if r == ' ' || r == '.' || r == '[' || r == ']' {
			return false
		}
	}
	return true
}
