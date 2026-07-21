package nodes

import (
	"context"

	"christiangeorgelucas/nlp-tools/axiom"
	gen "christiangeorgelucas/nlp-tools/gen"
)

// Tokenize splits text into word/punctuation tokens and tags each with its
// Penn Treebank part-of-speech tag, using the MIT-licensed jdkato/prose
// tokenizer and averaged-perceptron tagger. Each token also carries its byte
// offset range in the original text (see the OFFSETS note in messages.proto)
// and, when the token participates in a named entity, an IOB label — the
// same classification ExtractEntities groups into whole entities. Text over
// 2 MiB or not valid UTF-8 returns a structured error instead of panicking.
func Tokenize(ctx context.Context, ax axiom.Context, input *gen.Document) (*gen.TokensResult, error) {
	if errMsg := validateText(input.Text); errMsg != "" {
		return &gen.TokensResult{Error: errMsg}, nil
	}

	doc, err := newProseDocument(input.Text, true /*tag*/, false /*segment*/, true /*extract, to populate Label*/)
	if err != nil {
		return &gen.TokensResult{Error: "tokenization failed: " + err.Error()}, nil
	}

	return &gen.TokensResult{Tokens: tokensToProto(input.Text, doc.Tokens())}, nil
}
