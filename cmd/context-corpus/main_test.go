package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/saifety-org/lab/internal/contextcorpus"
)

func TestReviewCommandDoesNotCertifyTemplate(t *testing.T) {
	source := "../../datasets/contextual/v1"
	out := filepath.Join(t.TempDir(), "review")
	var b bytes.Buffer
	if err := execute([]string{"-source", source, "-review-out", out}, &b); err != nil {
		t.Fatal(err)
	}
	var result contextcorpus.ReviewResult
	if err := json.Unmarshal(b.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "pending" || result.TotalRows != 252 || result.ReviewedRows != 0 {
		t.Fatalf("unexpected review status: %+v", result)
	}
	b.Reset()
	if err := execute([]string{"-source", source, "-reviews", filepath.Join(out, "review-template.json"), "-require-reviewed"}, &b); err == nil {
		t.Fatal("incomplete review passed acceptance gate")
	}
	if err := json.Unmarshal(b.Bytes(), &result); err != nil {
		t.Fatal("missing machine-readable failure result")
	}
	if err := execute([]string{"-source", source, "-review-out", out}, &b); err == nil {
		t.Fatal("review work overwritten")
	}
}

func TestReviewFlagErrors(t *testing.T) {
	for _, args := range [][]string{
		{"-require-reviewed"},
		{"-out", "out", "-review-out", "review"},
		{"-review-out", "review", "-reviews", "ledger.json"},
		{"-review-out", "review", "-review-key-out", "key"},
		{"-review-key-out", "key", "-require-reviewed"},
		{"unexpected"},
	} {
		if err := execute(args, &bytes.Buffer{}); err == nil {
			t.Fatalf("invalid flags accepted: %v", args)
		}
	}
}
