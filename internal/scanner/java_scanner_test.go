package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"nt-inspect/internal/model"
)

func TestJavaScanner(t *testing.T) {
	// Scan actual cte-parser project if available
	cwd, _ := os.Getwd()
	cteDir := filepath.Join(cwd, "..", "..", "cte-parser")
	if _, err := os.Stat(cteDir); err != nil {
		cteDir = filepath.Join(cwd, "cte-parser")
	}

	if _, err := os.Stat(cteDir); err != nil {
		t.Skip("cte-parser not found, skipping real repo test")
	}

	sc := NewJavaScanner(model.DocTypeCTE)
	cb, err := sc.Scan(context.Background(), cteDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(cb.Symbols) == 0 {
		t.Fatalf("expected extracted symbols from cte-parser, got 0")
	}

	// Verify key symbols from NT 2025.001
	symbolMap := make(map[string]bool)
	xpathMap := make(map[string]bool)
	for _, s := range cb.Symbols {
		symbolMap[s.Normalized] = true
		if s.XPath != "" {
			xpathMap[s.XPath] = true
		}
	}

	// Check if VIBS mapper was found
	if !symbolMap["VIBS"] {
		t.Errorf("expected symbol 'VIBS' to be scanned")
	}

	// Check if full path was captured
	if !xpathMap["CTEPROC.CTE.INFCTE.IMP.IBSCBS.GIBSCBS.VIBS"] {
		t.Errorf("expected full xpath 'CTEPROC.CTE.INFCTE.IMP.IBSCBS.GIBSCBS.VIBS' to be captured")
	}

	// Check if tpEnteGov was found
	if !symbolMap["TPENTEGOV"] {
		t.Errorf("expected symbol 'TPENTEGOV' to be scanned")
	}
}
