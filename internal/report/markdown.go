package report

import (
	"fmt"
	"io"
	"strings"
	"time"

	"nt-inspect/internal/model"
)

// GenerateMarkdown writes a complete formatted compliance audit report in Markdown to w.
func GenerateMarkdown(w io.Writer, report *model.InspectionReport) error {
	totalGaps := report.Summary.MissingCount + report.Summary.MismatchCount + report.Summary.ObsoleteCount

	statusBadge := "**EM CONFORMIDADE**"
	if report.Status != model.StatusCompliant {
		statusBadge = fmt.Sprintf("**NÃO CONFORME (%d GAPS IDENTIFICADOS)**", totalGaps)
	}

	// 1. Header & Metadata
	_, _ = fmt.Fprintf(w, "# Relatório de Conformidade — NT %s\n\n", report.NTCode)
	_, _ = fmt.Fprintf(w, "- **Status Geral:** %s\n", statusBadge)
	_, _ = fmt.Fprintf(w, "- **Documento Fiscal:** %s\n", strings.ToUpper(string(report.DocType)))
	_, _ = fmt.Fprintf(w, "- **Nota Técnica:** NT %s (Versão: %s)\n", report.NTCode, report.NTVersion)
	_, _ = fmt.Fprintf(w, "- **Diretório Inspecionado:** `%s`\n", report.TargetDir)
	_, _ = fmt.Fprintf(w, "- **Data da Inspeção:** %s\n\n", time.Now().Format("02/01/2006 15:04:05"))

	// 2. Executive Summary Table
	_, _ = fmt.Fprintln(w, "## 1. Resumo Executivo")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "| Métrica | Quantidade | Observação |")
	_, _ = fmt.Fprintln(w, "|---|---|---|")
	_, _ = fmt.Fprintf(w, "| Total de Campos da NT | %d | Total de grupos e campos da especificação |\n", report.Summary.TotalRequired)
	_, _ = fmt.Fprintf(w, "| Campos em Conformidade | %d | Mapeados corretamente no parser |\n", report.Summary.MatchedCount)
	_, _ = fmt.Fprintf(w, "| Campos Faltantes | %d | Novos campos exigidos pela NT ausentes no código |\n", report.Summary.MissingCount)
	_, _ = fmt.Fprintf(w, "| Nomes Divergentes | %d | Tags ou campos renomeados na NT |\n", report.Summary.MismatchCount)
	_, _ = fmt.Fprintf(w, "| Campos Obsoletos Detectados | %d | Grupos ou campos revogados ainda mantidos no código |\n", report.Summary.ObsoleteCount)
	_, _ = fmt.Fprintf(w, "| **Taxa de Aderência** | **%.1f%%** | Proporção de conformidade com a NT |\n\n", report.Summary.ComplianceRate*100.0)

	// Filter findings by kind
	var missing []model.FieldFinding
	var mismatches []model.FieldFinding
	var obsolete []model.FieldFinding
	var matched []model.FieldFinding

	for _, f := range report.Findings {
		switch f.Kind {
		case model.FindingMissing:
			missing = append(missing, f)
		case model.FindingNameMismatch:
			mismatches = append(mismatches, f)
		case model.FindingObsolete:
			obsolete = append(obsolete, f)
		case model.FindingMatch:
			matched = append(matched, f)
		}
	}

	// 3. Gaps Analysis
	_, _ = fmt.Fprintln(w, "## 2. Gaps e Ações de Adequação")
	_, _ = fmt.Fprintln(w, "")

	if totalGaps == 0 {
		_, _ = fmt.Fprintln(w, "Nenhum gap encontrado. O código do parser está 100% em conformidade com as regras avaliadas da NT.")
		_, _ = fmt.Fprintln(w, "")
	} else {
		// 3.1 Missing Fields
		if len(missing) > 0 {
			_, _ = fmt.Fprintln(w, "### 2.1 Campos Faltantes no Código")
			_, _ = fmt.Fprintln(w, "")
			_, _ = fmt.Fprintln(w, "| Campo / Tag | Elemento Pai | Caminho XML Esperado | Ação Sugerida |")
			_, _ = fmt.Fprintln(w, "|---|---|---|---|")
			for _, f := range missing {
				_, _ = fmt.Fprintf(w, "| `%s` | `%s` | `%s` | %s |\n", f.FieldName, f.ParentName, f.ExpectedPath, f.Suggestion)
			}
			_, _ = fmt.Fprintln(w, "")
		}

		// 3.2 Name Mismatches
		if len(mismatches) > 0 {
			_, _ = fmt.Fprintln(w, "### 2.2 Nomes Divergentes ou Alterados")
			_, _ = fmt.Fprintln(w, "")
			_, _ = fmt.Fprintln(w, "| Nome no Código | Caminho Esperado na NT | Localização | Ação Sugerida |")
			_, _ = fmt.Fprintln(w, "|---|---|---|---|")
			for _, f := range mismatches {
				_, _ = fmt.Fprintf(w, "| `%s` | `%s` | `%s` | %s |\n", f.FieldName, f.ExpectedPath, f.Location, f.Suggestion)
			}
			_, _ = fmt.Fprintln(w, "")
		}

		// 3.3 Obsolete Fields
		if len(obsolete) > 0 {
			_, _ = fmt.Fprintln(w, "### 2.3 Campos Obsoletos ou Revogados pela NT")
			_, _ = fmt.Fprintln(w, "")
			_, _ = fmt.Fprintln(w, "| Símbolo no Código | Localização | Diagnóstico |")
			_, _ = fmt.Fprintln(w, "|---|---|---|")
			seen := make(map[string]bool)
			for _, f := range obsolete {
				if !seen[f.FieldName] {
					seen[f.FieldName] = true
					_, _ = fmt.Fprintf(w, "| `%s` | `%s` | %s |\n", f.FieldName, f.Location, f.Message)
				}
			}
			_, _ = fmt.Fprintln(w, "")
		}
	}

	// 4. Compliant Fields Sample
	_, _ = fmt.Fprintln(w, "## 3. Campos em Conformidade")
	_, _ = fmt.Fprintln(w, "")
	if len(matched) == 0 {
		_, _ = fmt.Fprintln(w, "Nenhum campo mapeado foi associado a esta NT.")
	} else {
		_, _ = fmt.Fprintf(w, "Foram validados **%d** campos mapeados corretamente no parser.\n\n", len(matched))
		_, _ = fmt.Fprintln(w, "| Campo | Caminho XML | Mapeado em |")
		_, _ = fmt.Fprintln(w, "|---|---|---|")
		maxShow := 30
		if len(matched) < maxShow {
			maxShow = len(matched)
		}
		for i := 0; i < maxShow; i++ {
			f := matched[i]
			_, _ = fmt.Fprintf(w, "| `%s` | `%s` | `%s` |\n", f.FieldName, f.ExpectedPath, f.Location)
		}
		if len(matched) > maxShow {
			_, _ = fmt.Fprintf(w, "\n*... e mais %d campos validados com sucesso.*\n", len(matched)-maxShow)
		}
	}

	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "---")
	_, _ = fmt.Fprintln(w, "*Relatório gerado automaticamente por `nt-inspect`.*")

	return nil
}

