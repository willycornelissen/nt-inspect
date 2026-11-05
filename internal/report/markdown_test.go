package report

import (
	"bytes"
	"strings"
	"testing"

	"nt-inspect/internal/model"
)

func TestGenerateMarkdown(t *testing.T) {
	rep := &model.InspectionReport{
		Status:    model.StatusNonCompliant,
		ExitCode:  1,
		DocType:   model.DocTypeCTE,
		NTCode:    "2025.001",
		NTVersion: "1.14",
		TargetDir: "/projects/cte-parser",
		Summary: model.ComplianceSummary{
			TotalRequired:  10,
			MatchedCount:   8,
			MissingCount:   1,
			MismatchCount:  1,
			ObsoleteCount:  1,
			ComplianceRate: 0.8,
		},
		Findings: []model.FieldFinding{
			{
				Kind:         model.FindingMatch,
				FieldName:    "vIBS",
				ExpectedPath: "IMP.IBSCBS.GIBSCBS.VIBS",
				Location:     "IBSCBSMapper.java:42",
			},
			{
				Kind:         model.FindingMissing,
				FieldName:    "missingField",
				ParentName:   "gParent",
				ExpectedPath: "IMP.GPARENT.MISSINGFIELD",
				Suggestion:   "Mapear campo",
			},
			{
				Kind:         model.FindingNameMismatch,
				FieldName:    "tpCompraGov",
				ExpectedPath: "IDE.GCOMPRAGOV.TPENTEGOV",
				Location:     "Cte.java:682",
				Suggestion:   "Atualizar nome",
			},
			{
				Kind:      model.FindingObsolete,
				FieldName: "gIBSCredPres_vCredPres",
				Location:  "IbsCbs.java:140",
				Message:   "Campo revogado",
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateMarkdown(&buf, rep)
	if err != nil {
		t.Fatalf("GenerateMarkdown failed: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "Relatório de Conformidade — NT 2025.001") {
		t.Errorf("expected title with NT code")
	}
	if !strings.Contains(out, "NÃO CONFORME") {
		t.Errorf("expected NON_COMPLIANT status in report")
	}
	if !strings.Contains(out, "missingField") {
		t.Errorf("expected missing field in table")
	}
	if !strings.Contains(out, "tpCompraGov") {
		t.Errorf("expected mismatch field in table")
	}
	if !strings.Contains(out, "gIBSCredPres_vCredPres") {
		t.Errorf("expected obsolete field in table")
	}
}

