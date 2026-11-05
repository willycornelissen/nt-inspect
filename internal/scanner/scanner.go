package scanner

import (
	"context"
	"nt-inspect/internal/model"
)

// CodeSymbol represents a discovered tag mapping, property, or field in the codebase.
type CodeSymbol struct {
	Name       string // Tag name or property name (e.g. "VIBS", "tpEnteGov")
	Normalized string // Upper-case normalized identifier (e.g. "VIBS", "TPENTEGOV")
	XPath      string // Full XML path if mapped (e.g. "CTEPROC.CTE.INFCTE.IMP.IBSCBS.GIBSCBS.VIBS")
	FilePath   string // Relative file path (e.g. "src/main/java/.../IBSCBSMapper.java")
	LineNumber int    // 1-based line number
	Kind       string // "MAPPER_TAG", "JSON_PROPERTY", "FIELD_DEF"
}

// ScannedCodebase holds all extracted symbols from the parser repository.
type ScannedCodebase struct {
	DocType model.DocType
	RootDir string
	Symbols []CodeSymbol
}

// CodeScanner defines the capability to scan a parser repository for XML/model mappings.
type CodeScanner interface {
	Scan(ctx context.Context, rootDir string) (*ScannedCodebase, error)
}

