// Command context-corpus validates and prepares contextual research fixtures.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/saifety-org/lab/internal/contextcorpus"
	"os"
)

func run(args []string) error {
	flags := flag.NewFlagSet("context-corpus", flag.ContinueOnError)
	dir := flags.String("source", "datasets/contextual/v1", "corpus directory")
	out := flags.String("out", "", "optional prepared artifact directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	c, err := contextcorpus.Load(*dir)
	if err != nil {
		return err
	}
	if *out != "" {
		if err = contextcorpus.Prepare(c, *out); err != nil {
			return err
		}
	}
	return json.NewEncoder(os.Stdout).Encode(c.Manifest())
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
