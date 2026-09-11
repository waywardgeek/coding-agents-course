// Command grade is the Chapter 1 auto-grader.
//
//	grade ./solutions/ch01      # build and grade a package directory
//	grade ./some/prebuilt-bin   # grade an existing executable
//	grade -json ./solutions/ch01
//
// It needs no API key and reaches no network: the submission is pointed at a
// fake Anthropic server on localhost via ANTHROPIC_BASE_URL.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/waywardgeek/coding-agents-course/internal/grade"
)

func main() {
	asJSON := flag.Bool("json", false, "emit the report as JSON")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: grade [-json] <package-dir|executable>\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	bin, cleanup, err := grade.Build(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "grade: %v\n", err)
		os.Exit(2)
	}
	defer cleanup()

	res, err := grade.Run(bin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grade: %v\n", err)
		os.Exit(2)
	}

	report := grade.NewReport(grade.Evaluate(res), res.Stderr)
	if *asJSON {
		if err := report.WriteJSON(os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "grade: %v\n", err)
			os.Exit(2)
		}
	} else {
		report.WriteText(os.Stdout)
	}
	if !report.Passed {
		os.Exit(1)
	}
}
