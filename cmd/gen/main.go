package main

import (
	"flag"

	"github.com/TOMOFUMI-KONDO/frr-demo/gen"
)

var base string
var cfg string

func init() {
	flag.StringVar(&base, "base", ".", "Path of base dir")
	flag.StringVar(&cfg, "cfg", "hosts.yaml", "Path of hosts config file from base dir")
	flag.Parse()
}

func main() {
	if err := gen.GenHost(base, cfg); err != nil {
		panic(err)
	}
}
