package main

import (
	"os"
	"time"

	"github.com/cppcho/boxed/internal/boxed"
)

func main() {
	p := boxed.DefaultPaths()

	x := &boxed.XbarApp{
		Paths:   p,
		Runner:  boxed.RealRunner{},
		NowFunc: time.Now,
		Stdout:  os.Stdout,
	}
	x.Run()
}
