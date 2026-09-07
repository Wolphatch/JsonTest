package main

import (
	"fmt"
	"os"

	"github.com/example/json-test/internal/jsontest"
)

var version = "dev"

func main() { os.Exit(realMain(os.Args[1:])) }

func realMain(args []string) int {
	if len(args) == 1 && args[0] == "version" {
		fmt.Printf("json-test %s\n", version)
		return 0
	}
	if len(args) != 2 || (args[0] != "run" && args[0] != "validate") {
		fmt.Fprintln(os.Stderr, "usage: json-test {run|validate} <manifest.yaml> | json-test version")
		return 2
	}
	m, err := jsontest.Load(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid configuration: %v\n", err)
		return 2
	}
	if args[0] == "validate" {
		fmt.Printf("valid: %d test case(s)\n", len(m.Tests))
		return 0
	}
	r := jsontest.Run(m)
	if m.Report == "json" {
		if err := jsontest.WriteJSON(os.Stdout, r); err != nil {
			fmt.Fprintf(os.Stderr, "report: %v\n", err)
			return 1
		}
	} else {
		jsontest.WriteText(os.Stdout, r)
	}
	if !r.Matched {
		return 1
	}
	return 0
}
