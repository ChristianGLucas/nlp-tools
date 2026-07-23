# christiangeorgelucas/nlp-tools

Composable [Axiom](https://axiomide.com) nodes for structured natural-language
processing over plain text: tokenization with part-of-speech tagging, named-entity
recognition, and sentence segmentation. Built for the Axiom marketplace.

Wraps [`jdkato/prose`](https://github.com/jdkato/prose) (v2, MIT license), a pure-Go
NLP library whose trained models (an averaged-perceptron POS tagger and a Maxent
named-entity classifier) are compiled directly into the binary — no runtime model
download, no external service call, fully offline and deterministic.

## Nodes

All nodes take a single `Document { string text }` input. That field name
deliberately matches the `text` field already used by
[`ocr-tools`](https://github.com/ChristianGLucas/ocr-tools)'s `TextOut`,
[`pdf-tools`](https://github.com/ChristianGLucas/pdf-tools)'s `TextResult`, and
[`html-tools`](https://github.com/ChristianGLucas/html-tools)'s `TextResult`, so any
of those packages' extracted text flows straight into these nodes in a flow with a
trivial one-field mapping.

- **Tokenize** — split text into word/punctuation tokens, each tagged with its Penn
  Treebank part-of-speech tag (e.g. `NNP`, `VBZ`) and its byte-offset span in the
  original text.
- **ExtractEntities** — recognize named entities (people, organizations, locations,
  and similar categories) with their category label and byte-offset span.
- **SegmentSentences** — split text into sentences using an abbreviation-aware Punkt
  segmenter (correctly keeps "Dr. Smith" from splitting on the period), with each
  sentence's byte-offset span.
- **Analyze** — run all three passes in one call and return tokens, entities, and
  sentences together — the efficient option when a caller wants everything.

## Offsets

`jdkato/prose` reports token/entity/sentence *text* but not character positions.
Every node here recovers byte offsets deterministically with a left-to-right
substring search over the original text. This is exact for every token (prose only
splits text, it never rewrites characters) and for single-word entities/sentences. A
multi-word entity is reconstructed by prose with single-space joins between its
tokens; if the original text had different spacing there, the offset cannot be
recovered exactly and the node reports `start = -1, end = -1` rather than guessing.

## Bounds

Input text is capped at 2 MiB and must be valid UTF-8; oversized or invalid input
returns a structured `error` field rather than crashing.

## License

MIT — see [LICENSE](./LICENSE). `jdkato/prose` and its full runtime dependency
closure (`golang-set`, `commonregex`, `gonum`, `neurosnap/sentences`) are all
MIT/BSD-3-Clause; none of it is copyleft.
