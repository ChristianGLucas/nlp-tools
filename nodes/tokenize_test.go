package nodes_test

import (
	"context"
	"strings"
	"testing"

	gen "christiangeorgelucas/nlp-tools/gen"
	"christiangeorgelucas/nlp-tools/nodes"
)

// TestTokenize is the golden test: a fixed sentence with Penn Treebank
// part-of-speech tags that are independently verifiable against the Penn
// Treebank tagging standard (not just "whatever prose happens to output") —
// e.g. "is" is a 3rd-person-singular-present verb (VBZ), "a" is a
// determiner (DT), "." is punctuation ("."). This is the independent-oracle
// check for this node: the expected tags come from the published tagging
// convention, not from re-running the library under test.
func TestTokenize(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	text := "Apple is looking at buying a U.K. startup for $1 billion."
	input := &gen.Document{Text: text}

	got, err := nodes.Tokenize(ctx, ax, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error: %s", got.Error)
	}

	wantTexts := []string{"Apple", "is", "looking", "at", "buying", "a", "U.K.", "startup", "for", "$", "1", "billion", "."}
	wantTags := []string{"NNP", "VBZ", "VBG", "IN", "VBG", "DT", "NNP", "NN", "IN", "$", "CD", "CD", "."}

	if len(got.Tokens) != len(wantTexts) {
		t.Fatalf("token count = %d, want %d (%v)", len(got.Tokens), len(wantTexts), got.Tokens)
	}
	for i, tok := range got.Tokens {
		if tok.Text != wantTexts[i] {
			t.Errorf("token[%d].Text = %q, want %q", i, tok.Text, wantTexts[i])
		}
		if tok.Tag != wantTags[i] {
			t.Errorf("token[%d].Tag = %q (text %q), want %q", i, tok.Tag, tok.Text, wantTags[i])
		}
		// Offset invariant, independent of prose's own logic: the recovered
		// span must literally reproduce the token's text from the source.
		if tok.Start < 0 || tok.End < 0 {
			t.Errorf("token[%d] %q: offsets not recovered (start=%d end=%d)", i, tok.Text, tok.Start, tok.End)
			continue
		}
		if got := text[tok.Start:tok.End]; got != tok.Text {
			t.Errorf("token[%d]: text[%d:%d] = %q, want %q", i, tok.Start, tok.End, got, tok.Text)
		}
	}

	// "Apple" is the sentence's grammatical subject and is proper-noun
	// tagged (NNP) regardless of whether the entity extractor's category
	// label for it is exactly right — POS tagging and NER are independent
	// passes over the same tokens.
	if got.Tokens[0].Tag != "NNP" {
		t.Errorf("first token Apple should be tagged NNP (proper noun), got %q", got.Tokens[0].Tag)
	}
}

// TestTokenize_Empty confirms empty input is not an error — it is
// trivially zero tokens, not a malformed request.
func TestTokenize_Empty(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.Tokenize(ctx, ax, &gen.Document{Text: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error on empty input: %s", got.Error)
	}
	if len(got.Tokens) != 0 {
		t.Errorf("expected zero tokens for empty input, got %d", len(got.Tokens))
	}
}

// TestTokenize_OversizedInput is the error-path test: input over the 2 MiB
// cap must return a structured error, not crash or hang.
func TestTokenize_OversizedInput(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	huge := strings.Repeat("a ", 2*1024*1024) // > 2 MiB

	got, err := nodes.Tokenize(ctx, ax, &gen.Document{Text: huge})
	if err != nil {
		t.Fatalf("expected a structured error, not a Go error: %v", err)
	}
	if got.Error == "" {
		t.Fatal("expected a structured error for oversized input, got none")
	}
	if len(got.Tokens) != 0 {
		t.Errorf("expected no tokens alongside an error, got %d", len(got.Tokens))
	}
}

// TestTokenize_InvalidUTF8 is the error-path test for malformed encoding.
func TestTokenize_InvalidUTF8(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	bad := "hello \xff\xfe world"

	got, err := nodes.Tokenize(ctx, ax, &gen.Document{Text: bad})
	if err != nil {
		t.Fatalf("expected a structured error, not a Go error: %v", err)
	}
	if got.Error == "" {
		t.Fatal("expected a structured error for invalid UTF-8, got none")
	}
}

// TestTokenize_Determinism invokes the node twice with the same input and
// requires an identical result — every node in this package is claimed to
// be deterministic.
func TestTokenize_Determinism(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	input := &gen.Document{Text: "The quick brown fox jumps over the lazy dog."}

	got1, err1 := nodes.Tokenize(ctx, ax, input)
	got2, err2 := nodes.Tokenize(ctx, ax, input)
	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v / %v", err1, err2)
	}
	if len(got1.Tokens) != len(got2.Tokens) {
		t.Fatalf("nondeterministic token count: %d vs %d", len(got1.Tokens), len(got2.Tokens))
	}
	for i := range got1.Tokens {
		if got1.Tokens[i].Text != got2.Tokens[i].Text || got1.Tokens[i].Tag != got2.Tokens[i].Tag {
			t.Errorf("nondeterministic token[%d]: %+v vs %+v", i, got1.Tokens[i], got2.Tokens[i])
		}
	}
}
