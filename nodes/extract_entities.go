package nodes

import (
	"context"

	"christiangeorgelucas/nlp-tools/axiom"
	gen "christiangeorgelucas/nlp-tools/gen"
)

// ExtractEntities recognizes named entities in text (people, organizations,
// locations, and similar categories — the exact label set is determined by
// jdkato/prose's bundled Maxent classifier) and returns each entity's text,
// category label, and byte offset range in the original text (see the
// OFFSETS note in messages.proto; a multi-word entity whose original
// spacing/punctuation prevents an exact substring match reports start=-1,
// end=-1 rather than a wrong guess). Text over 2 MiB or not valid UTF-8
// returns a structured error instead of panicking.
func ExtractEntities(ctx context.Context, ax axiom.Context, input *gen.Document) (*gen.EntitiesResult, error) {
	if errMsg := validateText(input.Text); errMsg != "" {
		return &gen.EntitiesResult{Error: errMsg}, nil
	}

	doc, err := newProseDocument(input.Text, true /*tag, required for extraction*/, false /*segment*/, true /*extract*/)
	if err != nil {
		return &gen.EntitiesResult{Error: "entity extraction failed: " + err.Error()}, nil
	}

	return &gen.EntitiesResult{Entities: entitiesToProto(input.Text, doc.Entities())}, nil
}
