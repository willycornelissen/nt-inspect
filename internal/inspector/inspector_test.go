package inspector

import (
	"context"
	"testing"

	"nt-inspect/internal/model"
	"nt-inspect/internal/scanner"
)

func TestInspector(t *testing.T) {
	spec := &model.DocumentSpec{
		DocType: model.DocTypeCTE,
		NTCode:  "2025.001",
		Version: "v1.14a",
		LayoutTable: []model.FieldElement{
			{Name: "vIBS", Parent: "gIBSCBS", XMLPath: "IMP.IBSCBS.GIBSCBS.VIBS"},
			{Name: "tpEnteGov", Parent: "gCompraGov", XMLPath: "IDE.GCOMPRAGOV.TPENTEGOV"},
			{Name: "missingTag", Parent: "gTest", XMLPath: "IMP.GTEST.MISSINGTAG"},
		},
	}

	codebase := &scanner.ScannedCodebase{
		DocType: model.DocTypeCTE,
		RootDir: "/tmp/mock-parser",
		Symbols: []scanner.CodeSymbol{
			{
				Name:       "vIBS",
				Normalized: "VIBS",
				XPath:      "CTEPROC.CTE.INFCTE.IMP.IBSCBS.GIBSCBS.VIBS",
				FilePath:   "IBSCBSMapper.java",
				LineNumber: 42,
			},
			{
				Name:       "tpEnteGov",
				Normalized: "TPENTEGOV",
				XPath:      "CTEPROC.CTE.INFCTE.IDE.GCOMPRAGOV.TPENTEGOV",
				FilePath:   "CteMapper.java",
				LineNumber: 84,
			},
		},
	}

	ins := NewInspector()
	report, err := ins.Inspect(context.Background(), spec, codebase)
	if err != nil {
		t.Fatalf("Inspect failed: %v", err)
	}

	if report.Status != model.StatusNonCompliant {
		t.Errorf("expected NON_COMPLIANT due to missingTag, got %s", report.Status)
	}

	if report.Summary.MatchedCount != 2 {
		t.Errorf("expected 2 matched fields, got %d", report.Summary.MatchedCount)
	}

	if report.Summary.MissingCount != 1 {
		t.Errorf("expected 1 missing field, got %d", report.Summary.MissingCount)
	}
}

