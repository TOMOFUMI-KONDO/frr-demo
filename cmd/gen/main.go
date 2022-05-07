package main

import (
	"flag"

	"github.com/TOMOFUMI-KONDO/frr-demo/gen"
)

var dir string

func init() {
	flag.StringVar(&dir, "dir", ".", "Path of dir dir")
	flag.Parse()
}

func main() {
	if err := gen.Gen(dir); err != nil {
		panic(err)
	}
}
