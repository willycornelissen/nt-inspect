package cache

import (
	"os"
	"testing"

	"nt-inspect/internal/model"
)

func TestCacheStore(t *testing.T) {
	tmp := t.TempDir()
	store := &Store{BaseDir: tmp}

	// 1. Save and Get PDF
	pdfData := []byte("%PDF-1.4 mock content")
	savedPath, err := store.SavePDF(model.DocTypeCTE, "2025.001", "v1.14a", pdfData)
	if err != nil {
		t.Fatalf("SavePDF failed: %v", err)
	}
	if _, err := os.Stat(savedPath); err != nil {
		t.Fatalf("saved file not found: %v", err)
	}

	loadedData, loadedPath, err := store.GetPDF(model.DocTypeCTE, "2025.001", "v1.14a")
	if err != nil {
		t.Fatalf("GetPDF failed: %v", err)
	}
	if loadedPath != savedPath || string(loadedData) != string(pdfData) {
		t.Fatalf("GetPDF mismatched content or path")
	}

	// 2. Save and Get Spec
	spec := &model.DocumentSpec{
		DocType: model.DocTypeCTE,
		NTCode:  "2025.001",
		Version: "v1.14a",
		Title:   "Nota Tecnica Teste",
		LayoutTable: []model.FieldElement{
			{ID: 1, Name: "vIBS", Kind: model.KindElement, Parent: "gIBSCBS"},
		},
	}

	if err := store.SaveSpec(spec); err != nil {
		t.Fatalf("SaveSpec failed: %v", err)
	}

	loadedSpec, err := store.GetSpec(model.DocTypeCTE, "2025.001", "v1.14a")
	if err != nil {
		t.Fatalf("GetSpec failed: %v", err)
	}
	if loadedSpec.Title != spec.Title || len(loadedSpec.LayoutTable) != 1 {
		t.Fatalf("loaded spec does not match saved spec: %+v", loadedSpec)
	}
}

