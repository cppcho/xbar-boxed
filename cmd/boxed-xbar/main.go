package main

import (
	"os"
	"time"

	boxed "github.com/cppcho/boxed/internal/boxed"
)

func main() {
	home, _ := os.UserHomeDir()
	p := boxed.DefaultPaths()

	x := &boxed.XbarApp{
		Paths:    p,
		Runner:   boxed.RealRunner{},
		NowFunc:  time.Now,
		Stdout:   os.Stdout,
		BoxedBin: home + "/bin/boxed",
	}
	x.Run()
}
