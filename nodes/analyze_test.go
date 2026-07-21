package nodes_test

import (
	"context"
	"strings"
	"testing"

	gen "christiangeorgelucas/nlp-tools/gen"
	"christiangeorgelucas/nlp-tools/nodes"
)

// TestAnalyze is the golden test for the combined pipeline: it must produce
// real, correct tokens, entities, and sentences in one call — not an empty
// or partially-populated result.
func TestAnalyze(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	text := "Barack Obama was born in Hawaii. He later became president."
	input := &gen.Document{Text: text}

	got, err := nodes.Analyze(ctx, ax, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error: %s", got.Error)
	}

	if len(got.Sentences) != 2 {
		t.Fatalf("sentence count = %d, want 2 (%v)", len(got.Sentences), got.Sentences)
	}
	if len(got.Entities) == 0 {
		t.Fatal("expected at least one entity (Barack Obama / Hawaii)")
	}
	if len(got.Tokens) == 0 {
		t.Fatal("expected tokens")
	}

	var foundPerson bool
	for _, e := range got.Entities {
		if e.Text == "Barack Obama" && e.Label == "PERSON" {
			foundPerson = true
		}
	}
	if !foundPerson {
		t.Errorf("expected entity {Barack Obama, PERSON} among %v", got.Entities)
	}

	// Every token's tag must be populated (Analyze always tags, unlike a
	// hypothetical tokenization-only mode).
	for i, tok := range got.Tokens {
		if tok.Tag == "" {
			t.Errorf("tokens[%d] %q: expected a non-empty POS tag", i, tok.Text)
		}
	}
}

// TestAnalyze_ConsistentWithIndividualNodes cross-checks Analyze's tokens
// against Tokenize's on the same input — the combined node must not diverge
// from the single-purpose nodes it composes.
func TestAnalyze_ConsistentWithIndividualNodes(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	text := "Apple is looking at buying a U.K. startup for $1 billion."
	input := &gen.Document{Text: text}

	analyzed, err := nodes.Analyze(ctx, ax, input)
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}
	tokenized, err := nodes.Tokenize(ctx, ax, input)
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}

	if len(analyzed.Tokens) != len(tokenized.Tokens) {
		t.Fatalf("token count mismatch: Analyze=%d Tokenize=%d", len(analyzed.Tokens), len(tokenized.Tokens))
	}
	for i := range analyzed.Tokens {
		a, tk := analyzed.Tokens[i], tokenized.Tokens[i]
		if a.Text != tk.Text || a.Tag != tk.Tag || a.Start != tk.Start || a.End != tk.End {
			t.Errorf("token[%d] mismatch: Analyze=%+v Tokenize=%+v", i, a, tk)
		}
	}
}

// TestAnalyze_OversizedInput is the error-path test.
func TestAnalyze_OversizedInput(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	huge := strings.Repeat("a ", 2*1024*1024)

	got, err := nodes.Analyze(ctx, ax, &gen.Document{Text: huge})
	if err != nil {
		t.Fatalf("expected a structured error, not a Go error: %v", err)
	}
	if got.Error == "" {
		t.Fatal("expected a structured error for oversized input, got none")
	}
}
