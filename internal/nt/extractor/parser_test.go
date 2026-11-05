package extractor

import (
	"context"
	"testing"

	"nt-inspect/internal/model"
)

func TestLayoutParser(t *testing.T) {
	sampleText := `
Projeto Conhecimento de Transporte Eletrônico
Nota Técnica 2025.001 – Reforma Tributária do Consumo
Versão 1.14 a de 11 de março de 2026

Sumário
1 Introdução
3 Alteração do leiaute do imposto do CTe
6 Criação do grupo de compras governamentais no grupo ide

1 Introdução
Texto de introdução aqui.

3 Alteração do leiaute do imposto do CTe
1   CST   E   IBSCBS   N   1-1   3   Código da Situação Tributária
2   cClassTrib   E   IBSCBS   N   1-1   6   Código da Classificação Tributária
3   indDoacao   E   IBSCBS   N   0-1   1   Informar quando doação
4   gIBSCBS   G   IBSCBS   -   0-1   -   Grupo de informações específicas do IBS/CBS
5   vBC   E   gIBSCBS   N   1-1   13,2   Valor da Base de cálculo comum a IBS/CBS
6   vIBS   E   gIBSCBS   N   1-1   13,2   Valor do IBS

6 Criação do grupo de compras governamentais no grupo ide
7   tpEnteGov   E   gCompraGov   N   1-1   1   Tipo de Ente Governamental
8   tpOperGov   E   gCompraGov   N   0-1   1   Tipo de Operação Governamental
`

	parser := NewLayoutParser(nil)
	spec, err := parser.Parse(context.Background(), model.DocTypeCTE, "2025.001", sampleText)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if spec.Version != "1.14" && spec.Version != "1.14 a" {
		t.Logf("Parsed version: %q", spec.Version)
	}
	if spec.Date != "11 de março de 2026" {
		t.Errorf("Expected date '11 de março de 2026', got %q", spec.Date)
	}

	// Verify sections
	if len(spec.Sections) < 2 {
		t.Fatalf("Expected at least 2 sections, got %d", len(spec.Sections))
	}

	// Verify layout fields
	if len(spec.LayoutTable) == 0 {
		t.Fatalf("Expected layout fields, got 0")
	}

	fieldMap := make(map[string]model.FieldElement)
	for _, f := range spec.LayoutTable {
		fieldMap[f.Name] = f
	}

	vIBS, ok := fieldMap["vIBS"]
	if !ok {
		t.Fatalf("Expected vIBS to be extracted")
	}
	if vIBS.Parent != "gIBSCBS" {
		t.Errorf("Expected vIBS parent to be gIBSCBS, got %s", vIBS.Parent)
	}
	// Verify XML Path resolution: IBSCBS.GIBSCBS.VIBS
	if vIBS.XMLPath != "IBSCBS.GIBSCBS.VIBS" {
		t.Errorf("Expected XMLPath 'IBSCBS.GIBSCBS.VIBS', got %q", vIBS.XMLPath)
	}

	tpEnteGov, ok := fieldMap["tpEnteGov"]
	if !ok {
		t.Fatalf("Expected tpEnteGov to be extracted")
	}
	if tpEnteGov.Parent != "gCompraGov" {
		t.Errorf("Expected tpEnteGov parent gCompraGov, got %s", tpEnteGov.Parent)
	}
}

