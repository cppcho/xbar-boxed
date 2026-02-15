package main

import (
	"fmt"
	"os"
	"time"

	boxed "github.com/cppcho/boxed/internal/boxed"
)

const usage = `Boxed - A timeboxing timer for your menu bar.

Usage:
    boxed start <duration in minutes> <task name...>
    boxed stop
    boxed again`

func main() {
	p := boxed.DefaultPaths()
	boxed.EnsureConfig(p)

	app := &boxed.App{
		Paths:   p,
		Runner:  boxed.RealRunner{},
		NowFunc: time.Now,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "start":
		err = app.CmdStart(os.Args[2:])
	case "stop":
		err = app.CmdStop(os.Args[2:])
	case "again":
		err = app.CmdAgain(os.Args[2:])
	case "complete":
		err = app.CmdComplete(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}

	if err != nil {
		os.Exit(1)
	}
}
