package nodes

import (
	"fmt"
	"strings"
	"unicode/utf8"

	gen "christiangeorgelucas/nlp-tools/gen"
	"github.com/jdkato/prose/v2"
)

// maxDocumentBytes bounds the size of input text every node will process.
// prose's segmentation/tagging pipeline is roughly linear in input size but
// unbounded input still means unbounded per-request cost — cap it, the same
// way christiangeorgelucas/pdf-tools caps page count and ocr-tools caps image
// bytes/megapixels on untrusted input.
const maxDocumentBytes = 2 * 1024 * 1024 // 2 MiB

// validateText checks a Document's text against the package's hard size and
// encoding bounds, returning a non-empty, human-readable error string when
// the input should be rejected (never panics or lets a downstream library
// misbehave on invalid input).
func validateText(text string) string {
	if len(text) > maxDocumentBytes {
		return fmt.Sprintf("text exceeds the %d byte limit (got %d bytes)", maxDocumentBytes, len(text))
	}
	if !utf8.ValidString(text) {
		return "text is not valid UTF-8"
	}
	return ""
}

// newProseDocument runs prose's pipeline with the given stages enabled. It
// is the single place every node builds a *prose.Document, so tagging,
// segmentation, and extraction options stay consistent across nodes.
func newProseDocument(text string, tag, segment, extract bool) (*prose.Document, error) {
	return prose.NewDocument(
		text,
		prose.WithTokenization(true),
		prose.WithTagging(tag),
		prose.WithSegmentation(segment),
		prose.WithExtraction(extract),
	)
}

// cursor tracks a monotonically-advancing search position into a fixed
// source text, used to recover byte offsets for a left-to-right sequence of
// substrings (tokens, sentences, or entities) that prose reports as text
// only, without positions. Each sequence (tokens vs. sentences vs. entities)
// must use its own cursor, since they are independent left-to-right passes
// over the SAME source text, not one merged pass.
type cursor struct {
	text string
	pos  int
}

// locate finds the next occurrence of piece at or after the cursor's current
// position and advances the cursor past it. Returns (-1, -1) without moving
// the cursor if piece is empty or cannot be found — callers must treat that
// as "position unknown", never guess.
func (c *cursor) locate(piece string) (int32, int32) {
	if piece == "" || c.pos > len(c.text) {
		return -1, -1
	}
	idx := strings.Index(c.text[c.pos:], piece)
	if idx < 0 {
		return -1, -1
	}
	start := c.pos + idx
	end := start + len(piece)
	c.pos = end
	return int32(start), int32(end)
}

// tokensToProto converts prose tokens into gen.Token, recovering offsets
// with a dedicated cursor over the original text.
func tokensToProto(text string, toks []prose.Token) []*gen.Token {
	c := &cursor{text: text}
	out := make([]*gen.Token, 0, len(toks))
	for _, t := range toks {
		start, end := c.locate(t.Text)
		out = append(out, &gen.Token{
			Text:  t.Text,
			Tag:   t.Tag,
			Label: t.Label,
			Start: start,
			End:   end,
		})
	}
	return out
}

// entitiesToProto converts prose entities into gen.Entity, recovering
// offsets with a dedicated cursor over the original text.
func entitiesToProto(text string, ents []prose.Entity) []*gen.Entity {
	c := &cursor{text: text}
	out := make([]*gen.Entity, 0, len(ents))
	for _, e := range ents {
		start, end := c.locate(e.Text)
		out = append(out, &gen.Entity{
			Text:  e.Text,
			Label: e.Label,
			Start: start,
			End:   end,
		})
	}
	return out
}

// sentencesToProto converts prose sentences into gen.Sentence, recovering
// offsets with a dedicated cursor over the original text.
func sentencesToProto(text string, sents []prose.Sentence) []*gen.Sentence {
	c := &cursor{text: text}
	out := make([]*gen.Sentence, 0, len(sents))
	for _, s := range sents {
		start, end := c.locate(s.Text)
		out = append(out, &gen.Sentence{
			Text:  s.Text,
			Start: start,
			End:   end,
		})
	}
	return out
}
