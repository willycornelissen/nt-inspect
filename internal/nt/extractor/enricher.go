package extractor

import (
	"context"

	"nt-inspect/internal/model"
)

// SemanticEnricher allows plugging an optional LLM or remote service to refine complex rules.
type SemanticEnricher interface {
	EnrichSpec(ctx context.Context, spec *model.DocumentSpec, fullText string) error
}

// NoopEnricher is the default enricher that performs no external operations.
type NoopEnricher struct{}

func (n *NoopEnricher) EnrichSpec(ctx context.Context, spec *model.DocumentSpec, fullText string) error {
	return nil
}

