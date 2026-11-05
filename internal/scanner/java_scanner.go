package scanner

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"nt-inspect/internal/model"
)

// JavaScanner inspects Java projects (Maven/Gradle) looking for tag mappers and model fields.
type JavaScanner struct {
	DocType model.DocType
}

// NewJavaScanner creates a scanner configured for Java parsers.
func NewJavaScanner(doc model.DocType) *JavaScanner {
	return &JavaScanner{DocType: doc}
}

var (
	// Matches: SETTERS.put("CTEPROC.CTE.INFCTE.IMP.IBSCBS.GIBSCBS.VIBS", ...) or .put("VTAR", ...)
	setterRegex = regexp.MustCompile(`\.put\s*\(\s*["']([^"']+)["']`)

	// Matches: @JsonProperty("tpEnteGov") or @JsonProperty(value = "tpEnteGov")
	jsonPropertyRegex = regexp.MustCompile(`@JsonProperty\s*\(\s*(?:value\s*=\s*)?["']([^"']+)["']\s*\)`)

	// Matches: private BigDecimal vIBS; or private Integer tpEnteGov;
	fieldDefRegex = regexp.MustCompile(`(?:private|protected|public)\s+[\w<>,.?\s]+\s+(\w+)\s*;`)
)

func (s *JavaScanner) Scan(ctx context.Context, rootDir string) (*ScannedCodebase, error) {
	absDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("resolve rootDir %s: %w", rootDir, err)
	}

	codebase := &ScannedCodebase{
		DocType: s.DocType,
		RootDir: absDir,
		Symbols: make([]CodeSymbol, 0, 200),
	}

	err = filepath.WalkDir(absDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return filepath.SkipDir
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "target" || name == "build" || name == ".idea" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".java") {
			return nil
		}

		relPath, _ := filepath.Rel(absDir, path)
		if scanErr := s.scanJavaFile(path, relPath, codebase); scanErr != nil {
			return scanErr
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("scan java codebase in %s: %w", absDir, err)
	}

	return codebase, nil
}

func (s *JavaScanner) scanJavaFile(fullPath, relPath string, codebase *ScannedCodebase) error {
	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("open file %s: %w", fullPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// 1. Check for setter tag mappings (.put("TAGPATH", ...))
		if setterMatches := setterRegex.FindAllStringSubmatch(line, -1); len(setterMatches) > 0 {
			for _, sm := range setterMatches {
				tagOrPath := strings.TrimSpace(sm[1])
				lastPart := tagOrPath
				if idx := strings.LastIndex(tagOrPath, "."); idx != -1 {
					lastPart = tagOrPath[idx+1:]
				}

				codebase.Symbols = append(codebase.Symbols, CodeSymbol{
					Name:       lastPart,
					Normalized: strings.ToUpper(lastPart),
					XPath:      strings.ToUpper(tagOrPath),
					FilePath:   relPath,
					LineNumber: lineNum,
					Kind:       "MAPPER_TAG",
				})
			}
		}

		// 2. Check for @JsonProperty("...")
		if jsonMatches := jsonPropertyRegex.FindAllStringSubmatch(line, -1); len(jsonMatches) > 0 {
			for _, jm := range jsonMatches {
				propName := strings.TrimSpace(jm[1])
				codebase.Symbols = append(codebase.Symbols, CodeSymbol{
					Name:       propName,
					Normalized: strings.ToUpper(propName),
					FilePath:   relPath,
					LineNumber: lineNum,
					Kind:       "JSON_PROPERTY",
				})
			}
		}

		// 3. Check for field definitions in model classes
		if strings.Contains(relPath, "/cte/") || strings.Contains(relPath, "/table/") || strings.Contains(relPath, "/model/") {
			if fieldMatches := fieldDefRegex.FindAllStringSubmatch(trimmed, -1); len(fieldMatches) > 0 {
				for _, fm := range fieldMatches {
					fName := strings.TrimSpace(fm[1])
					codebase.Symbols = append(codebase.Symbols, CodeSymbol{
						Name:       fName,
						Normalized: strings.ToUpper(fName),
						FilePath:   relPath,
						LineNumber: lineNum,
						Kind:       "FIELD_DEF",
					})
				}
			}
		}
	}

	return scanner.Err()
}

