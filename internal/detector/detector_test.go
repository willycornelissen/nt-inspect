package detector

import (
	"os"
	"path/filepath"
	"testing"

	"nt-inspect/internal/model"
)

func TestDetectDocType(t *testing.T) {
	// Test 1: Explicit flag override
	dt, err := DetectDocType("", "cte")
	if err != nil || dt != model.DocTypeCTE {
		t.Fatalf("expected cte, got %v (err: %v)", dt, err)
	}

	// Test 2: Existing cte-parser directory in workspace
	cwd, _ := os.Getwd()
	cteDir := filepath.Join(filepath.Dir(cwd), "nt-inspect", "cte-parser")
	if _, statErr := os.Stat(cteDir); statErr == nil {
		detected, detErr := DetectDocType(cteDir, "")
		if detErr != nil {
			t.Fatalf("failed to detect cte-parser: %v", detErr)
		}
		if detected != model.DocTypeCTE {
			t.Fatalf("expected cte for cte-parser dir, got %v", detected)
		}
	}

	// Test 3: Temporary mock directory with pom.xml for NFe
	tmpDir := t.TempDir()
	pomContent := `<project><artifactId>nfe-parser</artifactId><name>nfe-parser</name></project>`
	_ = os.WriteFile(filepath.Join(tmpDir, "pom.xml"), []byte(pomContent), 0644)

	detectedNFe, errNFe := DetectDocType(tmpDir, "")
	if errNFe != nil || detectedNFe != model.DocTypeNFE {
		t.Fatalf("expected nfe for temp dir with nfe-parser pom, got %v (err: %v)", detectedNFe, errNFe)
	}

	// Test 4: Ambiguous directory with nothing
	emptyDir := t.TempDir()
	_, errEmpty := DetectDocType(emptyDir, "")
	if errEmpty == nil {
		t.Fatalf("expected error for empty directory, got nil")
	}
}

