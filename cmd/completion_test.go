package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestFishCompletionValidSyntax(t *testing.T) {
	root := Lefthook()
	var buf bytes.Buffer
	root.Writer = &buf

	if err := root.Run(context.Background(), []string{"lefthook", "completion", "fish"}); err != nil {
		t.Fatalf("completion fish: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Fatal("expected fish completion script")
	}

	// Regression for evilmartians/lefthook#1430 / urfave/cli#2285: older
	// generators emitted Go default-value syntax such as (string=lefthook),
	// which fish rejects when sourcing vendor_completions.d scripts.
	if strings.Contains(out, "(string=") {
		t.Fatalf("fish completion contains invalid Go default syntax:\n%s", out)
	}

	if !strings.Contains(out, "complete -c lefthook") {
		t.Fatalf("expected fish complete directives:\n%s", out)
	}
}
