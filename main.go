package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	nl    = []byte{'\n'}
	quote = []byte{'"'}
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, strings.Join([]string{
			"Flatten JSON into greppable list",
			"",
			"Usage:",
			"  fj [OPTIONS] [FILE|-]",
			"",
			"Options:",
			"  -u, --unflatten  Turn flattened list back into JSON",
			"  -s, --stream     Treat JSON/flattened as stream",
			"  -h, --help       Print this information",
		}, "\n"))
	}

	var (
		unflattenFlag bool
		streamFlag    bool
	)

	flag.BoolVar(&unflattenFlag, "u", false, "")
	flag.BoolVar(&unflattenFlag, "unflatten", false, "")
	flag.BoolVar(&streamFlag, "s", false, "")
	flag.BoolVar(&streamFlag, "stream", false, "")

	flag.Parse()

	ins := make([]io.Reader, 0, flag.NArg())

	for _, file := range flag.Args() {
		if file == "-" {
			ins = append(ins, os.Stdin)
		} else {
			f, err := os.Open(file)
			fail(err)
			ins = append(ins, f)
		}
	}

	if len(ins) == 0 {
		ins = append(ins, os.Stdin)
	}

	in := io.MultiReader(ins...)
	out := bufio.NewWriter(os.Stdout)

	if unflattenFlag {
		unflatten(in, out, streamFlag)
	} else {
		flatten(in, out, streamFlag)
	}

	fail(out.Flush())
}

func fail(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
