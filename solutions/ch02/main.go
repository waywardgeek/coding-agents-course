package main

// Three surfaces, one conversation implementation behind all of them.
//
//	ch02              grader mode — stdio JSON-lines (the default)
//	ch02 chat         the REPL
//	ch02 render LOG   play LOG -> context -> render, print the request, exit
//
// `render` is the one worth staring at. It is the whole pipeline —
// EventLog -> Context -> vendor request — as a pure function on the command
// line, with no network anywhere near it. If that subcommand is awkward to
// write, the context is not actually separate from the transport.

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	switch {
	case len(args) == 0:
		NewEngine(false).Run()

	case args[0] == "chat":
		fmt.Fprintln(os.Stderr, "Type to talk. Type while it is working to steer it mid-turn.")
		NewEngine(true).Run()

	case args[0] == "render":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: ch02 render LOG")
			os.Exit(2)
		}
		if err := renderLog(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "render: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "usage: ch02 [chat | render LOG]\n")
		os.Exit(2)
	}
}

// renderLog replays a log and prints the request it would produce.
//
// Nothing here consults the clock, the network, a random source, or a Go map.
// Those are the four ways non-determinism gets into a renderer, and a log is a
// fixed input: two runs over a fixed input must agree byte for byte.
func renderLog(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	log, err := LoadLog(data)
	if err != nil {
		return err
	}
	ctx := Play(log)
	req := Render(ctx, NewClient().Model, MaxTokens)
	out, err := MarshalPretty(req)
	if err != nil {
		return err
	}
	os.Stdout.Write(out)
	os.Stdout.Write([]byte("\n"))
	return nil
}
