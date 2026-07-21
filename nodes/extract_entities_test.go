package nodes_test

import (
	"context"
	"strings"
	"testing"

	gen "christiangeorgelucas/nlp-tools/gen"
	"christiangeorgelucas/nlp-tools/nodes"
)

// TestExtractEntities is the golden test, using a sentence whose correct
// entity classification is verifiable independent of prose's implementation
// by ordinary world knowledge: "Barack Obama" denotes a person and "Hawaii"
// denotes a place, regardless of which library or model is asked.
func TestExtractEntities(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	text := "Barack Obama was born in Hawaii."
	input := &gen.Document{Text: text}

	got, err := nodes.ExtractEntities(ctx, ax, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error: %s", got.Error)
	}

	if len(got.Entities) != 2 {
		t.Fatalf("entity count = %d, want 2 (%v)", len(got.Entities), got.Entities)
	}

	person, place := got.Entities[0], got.Entities[1]
	if person.Text != "Barack Obama" {
		t.Errorf("entities[0].Text = %q, want %q", person.Text, "Barack Obama")
	}
	if person.Label != "PERSON" {
		t.Errorf("entities[0].Label = %q, want %q (a person's name)", person.Label, "PERSON")
	}
	if place.Text != "Hawaii" {
		t.Errorf("entities[1].Text = %q, want %q", place.Text, "Hawaii")
	}
	// prose's label taxonomy for places includes GPE ("geopolitical entity")
	// and LOCATION; assert it landed in a place-shaped category, not PERSON
	// or ORGANIZATION.
	if place.Label != "GPE" && place.Label != "LOCATION" {
		t.Errorf("entities[1].Label = %q, want a place category (GPE/LOCATION)", place.Label)
	}

	for i, ent := range got.Entities {
		if ent.Start < 0 || ent.End < 0 {
			t.Errorf("entities[%d] %q: offsets not recovered", i, ent.Text)
			continue
		}
		if got := text[ent.Start:ent.End]; got != ent.Text {
			t.Errorf("entities[%d]: text[%d:%d] = %q, want %q", i, ent.Start, ent.End, got, ent.Text)
		}
	}
}

// TestExtractEntities_NoEntities confirms plain text with no recognizable
// entities returns an empty, non-error result — no false positives forced.
func TestExtractEntities_NoEntities(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ExtractEntities(ctx, ax, &gen.Document{Text: "The quick brown fox jumps over the lazy dog."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error: %s", got.Error)
	}
	if len(got.Entities) != 0 {
		t.Errorf("expected no entities in a sentence with none, got %v", got.Entities)
	}
}

// TestExtractEntities_OversizedInput is the error-path test.
func TestExtractEntities_OversizedInput(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	huge := strings.Repeat("a ", 2*1024*1024)

	got, err := nodes.ExtractEntities(ctx, ax, &gen.Document{Text: huge})
	if err != nil {
		t.Fatalf("expected a structured error, not a Go error: %v", err)
	}
	if got.Error == "" {
		t.Fatal("expected a structured error for oversized input, got none")
	}
}
