package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saifety-org/lab/internal/contextcollection"
)

func TestPlanPinnedOfflineAndNoOverwrite(t *testing.T) {
	b, err := os.ReadFile("../../datasets/contextual/v5/protocol.json")
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "plan.json")
	if run([]string{"plan", "-input", "../../datasets/contextual/v5/protocol.json", "-sha256", strings.Repeat("0", 64), "-out", out}) == nil {
		t.Fatal("unpinned protocol accepted")
	}
	args := []string{"plan", "-input", "../../datasets/contextual/v5/protocol.json", "-sha256", contextcollection.Hash(b), "-out", out}
	if err = run(args); err != nil {
		t.Fatal(err)
	}
	if run(args) == nil {
		t.Fatal("plan overwritten")
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var p contextcollection.Plan
	if err = json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Slots) != 6380 {
		t.Fatalf("wrong quota reservations:%d", len(p.Slots))
	}
	counts := map[string]int{}
	for n, s := range p.Slots {
		if s.Index != n {
			t.Fatal("unstable tickets")
		}
		counts[s.Purpose+"/"+s.Language]++
	}
	for _, lang := range []string{"en", "ru", "es", "zh"} {
		if counts["validation/"+lang] != 600 || counts["confirmation/"+lang] != 600 || counts["train/"+lang] != 175 {
			t.Fatal(counts)
		}
	}
	if run([]string{"collect", "-input", out, "-sha256", strings.Repeat("f", 64), "-out", filepath.Join(t.TempDir(), "nothing")}) == nil {
		t.Fatal("unpinned plan accepted")
	}
	if run([]string{"unknown"}) == nil || run(nil) == nil {
		t.Fatal("unknown operation accepted")
	}
}
