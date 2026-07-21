package nodes_test

import (
	"context"
	"strings"
	"testing"

	gen "christiangeorgelucas/nlp-tools/gen"
	"christiangeorgelucas/nlp-tools/nodes"
)

// TestSegmentSentences is the golden test. The expected boundaries are an
// independent oracle: any competent English reader agrees this passage is
// exactly two sentences, with the "Dr." abbreviation NOT ending the first
// one — that is the specific capability (abbreviation-aware segmentation,
// not naive "split on every period") this node exists to provide.
func TestSegmentSentences(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	text := "Dr. Smith works at Acme Corp. He lives in San Francisco."
	input := &gen.Document{Text: text}

	got, err := nodes.SegmentSentences(ctx, ax, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error: %s", got.Error)
	}

	want := []string{
		"Dr. Smith works at Acme Corp.",
		"He lives in San Francisco.",
	}
	if len(got.Sentences) != len(want) {
		t.Fatalf("sentence count = %d, want %d (%v)", len(got.Sentences), len(want), got.Sentences)
	}
	for i, s := range got.Sentences {
		if s.Text != want[i] {
			t.Errorf("sentences[%d].Text = %q, want %q", i, s.Text, want[i])
		}
		if s.Start < 0 || s.End < 0 {
			t.Errorf("sentences[%d] %q: offsets not recovered", i, s.Text)
			continue
		}
		if got := text[s.Start:s.End]; got != s.Text {
			t.Errorf("sentences[%d]: text[%d:%d] = %q, want %q", i, s.Start, s.End, got, s.Text)
		}
	}
}

// TestSegmentSentences_Single confirms a single-sentence input is not
// over-split.
func TestSegmentSentences_Single(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	text := "The quick brown fox jumps over the lazy dog."

	got, err := nodes.SegmentSentences(ctx, ax, &gen.Document{Text: text})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Sentences) != 1 {
		t.Fatalf("sentence count = %d, want 1 (%v)", len(got.Sentences), got.Sentences)
	}
	if got.Sentences[0].Text != text {
		t.Errorf("sentences[0].Text = %q, want %q", got.Sentences[0].Text, text)
	}
}

// TestSegmentSentences_OversizedInput is the error-path test.
func TestSegmentSentences_OversizedInput(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	huge := strings.Repeat("a ", 2*1024*1024)

	got, err := nodes.SegmentSentences(ctx, ax, &gen.Document{Text: huge})
	if err != nil {
		t.Fatalf("expected a structured error, not a Go error: %v", err)
	}
	if got.Error == "" {
		t.Fatal("expected a structured error for oversized input, got none")
	}
}
