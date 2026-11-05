package detector

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"nt-inspect/internal/model"
)

type pomProject struct {
	ArtifactID string `xml:"artifactId"`
	Name       string `xml:"name"`
}

// DetectDocType analyzes the directory to infer the fiscal document type (e.g. cte).
// If explicit is provided and valid, it takes precedence.
func DetectDocType(targetDir string, explicit string) (model.DocType, error) {
	if explicit != "" {
		return model.ParseDocType(explicit)
	}

	if targetDir == "" {
		var err error
		targetDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("could not get current working directory: %w", err)
		}
	}

	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		return "", fmt.Errorf("invalid target directory %s: %w", targetDir, err)
	}

	// 1. Inspect Maven pom.xml if present
	pomPath := filepath.Join(absDir, "pom.xml")
	if data, err := os.ReadFile(pomPath); err == nil {
		var pom pomProject
		if xml.Unmarshal(data, &pom) == nil {
			art := strings.ToLower(pom.ArtifactID)
			name := strings.ToLower(pom.Name)
			if strings.Contains(art, "cte") || strings.Contains(name, "cte") {
				return model.DocTypeCTE, nil
			}
			if strings.Contains(art, "nfe") || strings.Contains(name, "nfe") {
				return model.DocTypeNFE, nil
			}
			if strings.Contains(art, "mdfe") || strings.Contains(name, "mdfe") {
				return model.DocTypeMDFE, nil
			}
		}
	}

	// 2. Inspect directory name itself
	base := strings.ToLower(filepath.Base(absDir))
	if strings.Contains(base, "cte") {
		return model.DocTypeCTE, nil
	}
	if strings.Contains(base, "nfe") {
		return model.DocTypeNFE, nil
	}
	if strings.Contains(base, "mdfe") {
		return model.DocTypeMDFE, nil
	}

	// 3. Scan top-level filenames and package paths
	cteFound := false
	nfeFound := false
	_ = filepath.WalkDir(absDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return filepath.SkipDir
		}
		// Limit depth to avoid walking entire node_modules, target, etc.
		rel, _ := filepath.Rel(absDir, path)
		depth := strings.Count(rel, string(filepath.Separator))
		if depth > 4 {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() && (d.Name() == "target" || d.Name() == ".git" || d.Name() == "node_modules") {
			return filepath.SkipDir
		}

		lower := strings.ToLower(d.Name())
		if strings.Contains(lower, "cteparser") || strings.Contains(lower, "cte") {
			cteFound = true
		}
		if strings.Contains(lower, "nfeparser") || strings.Contains(lower, "nfe") {
			nfeFound = true
		}
		return nil
	})

	if cteFound && !nfeFound {
		return model.DocTypeCTE, nil
	}
	if nfeFound && !cteFound {
		return model.DocTypeNFE, nil
	}

	return "", fmt.Errorf("could not automatically detect fiscal document type in %s. Please specify with --doc <cte|nfe|mdfe|nfce>", targetDir)
}

