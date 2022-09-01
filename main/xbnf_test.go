package main

import (
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	os.Args = os.Args[0:1]
	os.Args = append(os.Args, "-xbnf")
	os.Args = append(os.Args, "../samples/dgen/dgen.xbnf")
	os.Args = append(os.Args, "-file")
	os.Args = append(os.Args, "../samples/dgen/sample1.0.dgen")
	main()
}
