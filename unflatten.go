package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/valyala/bytebufferpool"
)

var (
	bComma      = []byte{','}
	bQuote      = []byte{'"'}
	bQuoteColon = []byte{'"', ':'}
	bArrOpen    = []byte{'['}
	bArrClose   = []byte{']'}
	bObjOpen    = []byte{'{'}
	bObjClose   = []byte{'}'}

	breaks = [256]bool{' ': true, '.': true, '[': true, ']': true}
)

func unflatten(in io.Reader, out io.Writer, stream bool) {
	var (
		err error

		// stack of object keys or array indices
		keys []*bytebufferpool.ByteBuffer
		// same stack but denotes if array or object
		isarr []bool
		line  []byte
		// ix of char in line, of end of current field, of current field
		i, j, field int
		onarr       bool
		prev        byte
	)

	scan := bufio.NewScanner(in)

	for scan.Scan() {
		line = scan.Bytes()
		field = -1
		onarr = false

		if !bytes.HasPrefix(line, []byte("json")) {
			fail(fmt.Errorf("bad line %q", line))
		}

	CHARSCAN:
		for i = 4; i < len(line); i++ { // skip "json" prefix
			switch line[i] {
			case '.':
				field++
				onarr = false
			case ']':
			case '[':
				field++
				onarr = line[i+1] != '"'
			case ' ':
				// compute if things should be closed
				closeFields(out, isarr[field+1:])
				for _, key := range keys[field+1:] {
					bytebufferpool.Put(key)
				}
				keys = keys[:field+1]
				isarr = isarr[:field+1]
				_, err = out.Write(line[i+3:])
				fail(err)
				break CHARSCAN

			default: // a-z, 0-9 etc

				if line[i] == '"' { // json["x x"]

					prev = 0
					for j = i + 1; line[j] != '"' && prev != '\\'; j++ {
						prev = line[j]
					}
					j++

				} else {
					// read field name
					for j = i; !breaks[line[j]]; j++ {
					}
				}

				switch {
				case field < 0:
					fail(fmt.Errorf("bad line %q", line))
				case field >= len(keys): // go deeper
					openField(out, onarr)
					if !onarr {
						objKey(out, line[i:j])
					}

					key := bytebufferpool.Get()
					_, err = key.Write(line[i:j])
					fail(err)

					keys = append(keys, key)
					isarr = append(isarr, onarr)

				case onarr != isarr[field]:
					fail(fmt.Errorf("type switch between array and object: %s", line))

				case !bytes.Equal(keys[field].B, line[i:j]): // new field name
					closeFields(out, isarr[field+1:])
					for _, key := range keys[field:] {
						bytebufferpool.Put(key)
					}
					keys = keys[:field]
					isarr = isarr[:field+1]

					_, err = out.Write(bComma)
					fail(err)
					if !onarr {
						objKey(out, line[i:j])
					}
					key := bytebufferpool.Get()
					_, err = key.Write(line[i:j])
					fail(err)
					keys = append(keys, key)
					// isarr stays the same, no change to types

				default: // same key
				}
				i = j - 1
			}

		}
	}

	closeFields(out, isarr)
	fail(scan.Err())
}

func closeFields(out io.Writer, isarr []bool) {
	var err error
	for i := len(isarr) - 1; i >= 0; i-- {
		if isarr[i] {
			_, err = out.Write(bArrClose)
		} else {
			_, err = out.Write(bObjClose)
		}
		fail(err)
	}
}

func openField(out io.Writer, onarr bool) {
	var err error
	if onarr {
		_, err = out.Write(bArrOpen)
		fail(err)
	} else {
		_, err = out.Write(bObjOpen)
		fail(err)
	}
}

func objKey(out io.Writer, key []byte) {
	var err error
	_, err = out.Write(bQuote)
	fail(err)
	if key[0] == '"' {
		key = key[1 : len(key)-1]
	}
	_, err = out.Write(key)
	fail(err)
	_, err = out.Write(bQuoteColon)
	fail(err)
}
