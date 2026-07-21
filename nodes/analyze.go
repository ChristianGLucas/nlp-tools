package nodes

import (
	"context"

	"christiangeorgelucas/nlp-tools/axiom"
	gen "christiangeorgelucas/nlp-tools/gen"
)

// Analyze runs the full jdkato/prose pipeline over text in one call —
// tokenization with part-of-speech tagging, named-entity extraction, and
// sentence segmentation — and returns all three result sets together. This
// is the efficient one-shot alternative to invoking Tokenize,
// ExtractEntities, and SegmentSentences separately when a caller wants
// everything; each field carries the same byte-offset semantics documented
// on Tokenize/ExtractEntities/SegmentSentences. Text over 2 MiB or not valid
// UTF-8 returns a structured error instead of panicking.
func Analyze(ctx context.Context, ax axiom.Context, input *gen.Document) (*gen.AnalyzeResult, error) {
	if errMsg := validateText(input.Text); errMsg != "" {
		return &gen.AnalyzeResult{Error: errMsg}, nil
	}

	doc, err := newProseDocument(input.Text, true /*tag*/, true /*segment*/, true /*extract*/)
	if err != nil {
		return &gen.AnalyzeResult{Error: "analysis failed: " + err.Error()}, nil
	}

	return &gen.AnalyzeResult{
		Tokens:    tokensToProto(input.Text, doc.Tokens()),
		Entities:  entitiesToProto(input.Text, doc.Entities()),
		Sentences: sentencesToProto(input.Text, doc.Sentences()),
	}, nil
}
