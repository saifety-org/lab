// Command context-corpus validates and prepares contextual research fixtures.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/saifety-org/lab/internal/contextcorpus"
)

func run(args []string) error {
	return execute(args, os.Stdout)
}

func execute(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("context-corpus", flag.ContinueOnError)
	dir := flags.String("source", "datasets/contextual/v2", "corpus directory")
	out := flags.String("out", "", "optional prepared artifact directory")
	reviewOut := flags.String("review-out", "", "new directory for blinded human review packet")
	reviewKey := flags.String("review-key-out", "", "new unblinded answer key file for post-blind reconciliation; keep separate from reviewers")
	reviews := flags.String("reviews", "", "human review ledger to validate")
	requireReviewed := flags.Bool("require-reviewed", false, "fail unless declared human/model review is complete (requires -reviews)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if *reviewKey != "" && (*reviewOut != "" || *reviews != "" || *out != "" || *requireReviewed) {
		return fmt.Errorf("answer key generation must be a separate operation")
	}
	if *requireReviewed && *reviews == "" {
		return fmt.Errorf("-require-reviewed requires -reviews")
	}
	if *reviewOut != "" && *reviews != "" {
		return fmt.Errorf("packet generation and review validation are separate operations")
	}
	if *out != "" && (*reviewOut != "" || *reviews != "") {
		return fmt.Errorf("preparation and review are separate operations")
	}
	c, err := contextcorpus.Load(*dir)
	if err != nil {
		return err
	}
	if *reviewKey != "" {
		return c.WriteReviewKey(*reviewKey)
	}
	if *out != "" {
		if err = contextcorpus.Prepare(c, *out); err != nil {
			return err
		}
	}
	if *reviewOut != "" || *reviews != "" {
		p, l := c.BuildReview()
		if *reviews != "" {
			l, err = contextcorpus.LoadReview(*reviews)
			if err != nil {
				return err
			}
		}
		result, err := c.CheckReview(p, l)
		if err != nil {
			return err
		}
		if *reviewOut != "" {
			if err = contextcorpus.WriteReview(p, l, *reviewOut); err != nil {
				return err
			}
		}
		if err = json.NewEncoder(output).Encode(result); err != nil {
			return err
		}
		if *requireReviewed && result.Status != "complete" {
			return fmt.Errorf("review incomplete (%d issues)", len(result.Issues))
		}
		return nil
	}
	return json.NewEncoder(output).Encode(c.Manifest())
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
