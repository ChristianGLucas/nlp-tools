package nodes

import (
	"context"

	"christiangeorgelucas/nlp-tools/axiom"
	gen "christiangeorgelucas/nlp-tools/gen"
)

// SegmentSentences splits text into sentences using jdkato/prose's Punkt
// sentence tokenizer (an unsupervised sentence-boundary model that handles
// abbreviations like "Dr." and "U.K." without splitting on their periods),
// returning each sentence's text and byte offset range in the original text
// (see the OFFSETS note in messages.proto). Text over 2 MiB or not valid
// UTF-8 returns a structured error instead of panicking.
func SegmentSentences(ctx context.Context, ax axiom.Context, input *gen.Document) (*gen.SentencesResult, error) {
	if errMsg := validateText(input.Text); errMsg != "" {
		return &gen.SentencesResult{Error: errMsg}, nil
	}

	doc, err := newProseDocument(input.Text, false /*tag*/, true /*segment*/, false /*extract*/)
	if err != nil {
		return &gen.SentencesResult{Error: "sentence segmentation failed: " + err.Error()}, nil
	}

	return &gen.SentencesResult{Sentences: sentencesToProto(input.Text, doc.Sentences())}, nil
}
