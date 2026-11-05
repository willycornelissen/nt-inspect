package extractor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// TextExtractor abstracts extracting plain text from a PDF buffer.
type TextExtractor interface {
	ExtractText(ctx context.Context, pdfData []byte) (string, error)
}

// PopplerTextExtractor invokes the system pdftotext utility.
type PopplerTextExtractor struct{}

// NewPopplerTextExtractor checks if pdftotext is available and returns an extractor.
func NewPopplerTextExtractor() (*PopplerTextExtractor, error) {
	_, err := exec.LookPath("pdftotext")
	if err != nil {
		return nil, fmt.Errorf("pdftotext utility not found in PATH (please install poppler-utils): %w", err)
	}
	return &PopplerTextExtractor{}, nil
}

// ExtractText pipes pdfData into pdftotext -layout - - and returns plain text.
func (p *PopplerTextExtractor) ExtractText(ctx context.Context, pdfData []byte) (string, error) {
	cmd := exec.CommandContext(ctx, "pdftotext", "-layout", "-", "-")
	cmd.Stdin = bytes.NewReader(pdfData)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftotext failed (%s): %w", stderr.String(), err)
	}

	return stdout.String(), nil
}

