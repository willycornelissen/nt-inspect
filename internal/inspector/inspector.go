package inspector

import (
	"context"
	"fmt"
	"strings"

	"nt-inspect/internal/model"
	"nt-inspect/internal/scanner"
)

// Inspector compares an NT DocumentSpec with scanned code symbols to assess compliance.
type Inspector struct{}

// NewInspector creates an Inspector instance.
func NewInspector() *Inspector {
	return &Inspector{}
}

// Inspect cross-references layout requirements against the codebase.
func (ins *Inspector) Inspect(ctx context.Context, spec *model.DocumentSpec, codebase *scanner.ScannedCodebase) (*model.InspectionReport, error) {
	report := &model.InspectionReport{
		DocType:   spec.DocType,
		NTCode:    spec.NTCode,
		NTVersion: spec.Version,
		TargetDir: codebase.RootDir,
		Findings:  make([]model.FieldFinding, 0, len(spec.LayoutTable)),
	}

	// Index scanned code symbols for efficient matching
	xpathMap := make(map[string][]scanner.CodeSymbol)
	nameMap := make(map[string][]scanner.CodeSymbol)

	for _, s := range codebase.Symbols {
		if s.XPath != "" {
			xpathMap[s.XPath] = append(xpathMap[s.XPath], s)
		}
		nameMap[s.Normalized] = append(nameMap[s.Normalized], s)
	}

	totalRequired := 0
	matchedCount := 0
	missingCount := 0
	mismatchCount := 0
	obsoleteCount := 0

	// 1. Audit fields required by NT
	for _, field := range spec.LayoutTable {
		// Only audit concrete elements and meaningful groups (skip pure abstract containers if needed)
		totalRequired++

		found := false
		var matchSymbol *scanner.CodeSymbol

		// A. Check for matching XPath (full or suffix match)
		expectedSuffix := field.XMLPath
		if idx := strings.Index(field.XMLPath, "."); idx != -1 {
			expectedSuffix = field.XMLPath[idx+1:]
		}

		for xp, syms := range xpathMap {
			if strings.HasSuffix(xp, field.XMLPath) || strings.HasSuffix(xp, expectedSuffix) {
				found = true
				matchSymbol = &syms[0]
				break
			}
		}

		// B. If not matched by XPath, check by normalized field name (tag or @JsonProperty)
		if !found {
			normField := strings.ToUpper(field.Name)
			if syms, ok := nameMap[normField]; ok && len(syms) > 0 {
				found = true
				matchSymbol = &syms[0]
			}
		}

		// C. Record finding
		if found && matchSymbol != nil {
			matchedCount++
			report.Findings = append(report.Findings, model.FieldFinding{
				Kind:         model.FindingMatch,
				FieldName:    field.Name,
				ParentName:   field.Parent,
				ExpectedPath: field.XMLPath,
				FoundPath:    matchSymbol.XPath,
				Location:     fmt.Sprintf("%s:%d", matchSymbol.FilePath, matchSymbol.LineNumber),
				Message:      fmt.Sprintf("Campo %s mapeado corretamente no parser", field.Name),
			})
		} else {
			missingCount++
			report.Findings = append(report.Findings, model.FieldFinding{
				Kind:         model.FindingMissing,
				FieldName:    field.Name,
				ParentName:   field.Parent,
				ExpectedPath: field.XMLPath,
				Message:      fmt.Sprintf("Campo %s (<%s>) exigido pela NT não encontrado no código", field.Name, field.Name),
				Suggestion:   fmt.Sprintf("Adicionar propriedade no modelo e registrar mapeamento de tag para %s", field.XMLPath),
			})
		}
	}

	// 2. Audit obsolete / superseded fields known in NT 2025.001
	// E.g.: gIBSCredPres group or tpCompraGov (renamed to tpEnteGov)
	if strings.Contains(spec.NTCode, "2025.001") {
		// Check for revoked credit group
		for symName, syms := range nameMap {
			if strings.HasPrefix(symName, "GIBSCREDPRES") || strings.HasPrefix(symName, "GCBSCREDPRES") {
				obsoleteCount++
				report.Findings = append(report.Findings, model.FieldFinding{
					Kind:       model.FindingObsolete,
					FieldName:  syms[0].Name,
					Location:   fmt.Sprintf("%s:%d", syms[0].FilePath, syms[0].LineNumber),
					Message:    fmt.Sprintf("Grupo/Campo %s foi revogado/excluído na NT 2025.001 v1.10+", syms[0].Name),
					Suggestion: "Remover ou descontinuar campos de crédito presumido do modelo e mapper",
				})
			}

			// Check for old name tpCompraGov
			if symName == "TPCOMPRAGOV" {
				mismatchCount++
				report.Findings = append(report.Findings, model.FieldFinding{
					Kind:         model.FindingNameMismatch,
					FieldName:    "tpCompraGov",
					ExpectedPath: "IDE.GCOMPRAGOV.TPENTEGOV",
					Location:     fmt.Sprintf("%s:%d", syms[0].FilePath, syms[0].LineNumber),
					Message:      "Campo nomeado como 'tpCompraGov' foi alterado para 'tpEnteGov' na NT v1.14",
					Suggestion:   "Atualizar mapeamento e nome do campo para tpEnteGov",
				})
			}
		}
	}

	// 3. Compute metrics & overall verdict
	rate := 0.0
	if totalRequired > 0 {
		rate = float64(matchedCount) / float64(totalRequired)
	}

	report.Summary = model.ComplianceSummary{
		TotalRequired:  totalRequired,
		MatchedCount:   matchedCount,
		MissingCount:   missingCount,
		ObsoleteCount:  obsoleteCount,
		MismatchCount:  mismatchCount,
		ComplianceRate: rate,
	}

	// Compliant if no missing or mismatch findings
	if missingCount == 0 && mismatchCount == 0 {
		report.Status = model.StatusCompliant
		report.ExitCode = 0
	} else {
		report.Status = model.StatusNonCompliant
		report.ExitCode = 1
	}

	return report, nil
}

