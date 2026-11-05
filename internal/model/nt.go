package model

import (
	"fmt"
	"regexp"
	"strings"
)

// DocType represents a supported Brazilian Electronic Tax Document type (DF-e).
type DocType string

const (
	DocTypeCTE  DocType = "cte"
	DocTypeNFE  DocType = "nfe"
	DocTypeMDFE DocType = "mdfe"
	DocTypeNFCE DocType = "nfce"
)

// ParseDocType normalizes and validates a document type string.
func ParseDocType(s string) (DocType, error) {
	clean := strings.ToLower(strings.TrimSpace(s))
	clean = strings.ReplaceAll(clean, "-", "")
	switch DocType(clean) {
	case DocTypeCTE:
		return DocTypeCTE, nil
	case DocTypeNFE:
		return DocTypeNFE, nil
	case DocTypeMDFE:
		return DocTypeMDFE, nil
	case DocTypeNFCE:
		return DocTypeNFCE, nil
	default:
		return "", fmt.Errorf("document type %q is unsupported (expected: cte, nfe, mdfe, nfce)", s)
	}
}

// NTIdentifier contains parsed and normalized metadata identifying a Technical Note.
type NTIdentifier struct {
	RawCode string  // e.g. "NT 2025.001", "2025.001", "2025001"
	Year    string  // "2025"
	Seq     string  // "001"
	Code    string  // "2025.001"
	Doc     DocType // "cte"
	Version string  // "v1.14a" or empty if unspecified
}

var (
	ntCodeRegex = regexp.MustCompile(`(?i)(?:NT\s*)?(\d{4})[._\s-]?(\d{3})`)
	ntVerRegex  = regexp.MustCompile(`(?i)(?:v|vers[aã]o)?\s*([0-9]+\.[0-9]+[a-z]?)`)
)

// ParseNTIdentifier parses and normalizes user input into an NTIdentifier.
func ParseNTIdentifier(raw string, defaultDoc DocType) (NTIdentifier, error) {
	matches := ntCodeRegex.FindStringSubmatch(raw)
	if len(matches) < 3 {
		return NTIdentifier{}, fmt.Errorf("invalid technical note code format %q (expected format like '2025.001' or 'NT 2025.001')", raw)
	}

	year, seq := matches[1], matches[2]
	code := fmt.Sprintf("%s.%s", year, seq)

	var version string
	verMatches := ntVerRegex.FindStringSubmatch(raw)
	if len(verMatches) >= 2 {
		version = strings.ToLower(strings.TrimSpace(verMatches[1]))
	}

	return NTIdentifier{
		RawCode: raw,
		Year:    year,
		Seq:     seq,
		Code:    code,
		Doc:     defaultDoc,
		Version: version,
	}, nil
}

// FieldKind represents the XML element classification in SEFAZ NT layout tables (G, E, A, CE).
type FieldKind string

const (
	KindGroup     FieldKind = "G"
	KindElement   FieldKind = "E"
	KindAttribute FieldKind = "A"
	KindChoice    FieldKind = "CE"
)

// FieldElement represents a single field or group in the SEFAZ layout table.
type FieldElement struct {
	ID          int       `json:"id"`                    // Order number in table (#)
	Name        string    `json:"name"`                  // Field name / tag name
	Kind        FieldKind `json:"kind"`                  // G, E, A, CE
	Parent      string    `json:"parent"`                // Parent tag name
	XMLPath     string    `json:"xmlPath"`               // Inferred full path: e.g. IMP.IBSCBS.GIBSCBS.VIBS
	DataType    string    `json:"dataType"`              // N, C, D, etc.
	Occurrence  string    `json:"occurrence"`            // 1-1, 0-1, 1-N, 0-N
	Length      string    `json:"length,omitempty"`      // e.g. 13,2 or 1
	Description string    `json:"description,omitempty"` // Context / rule description
}

// Section represents a chapter or section extracted from the Technical Note.
type Section struct {
	Number int            `json:"number"`
	Title  string         `json:"title"`
	Fields []FieldElement `json:"fields,omitempty"`
}

// DocumentSpec represents the structured specification extracted from a Technical Note.
type DocumentSpec struct {
	DocType     DocType        `json:"docType"`
	NTCode      string         `json:"ntCode"`
	Version     string         `json:"version"`
	Title       string         `json:"title"`
	Date        string         `json:"date,omitempty"`
	SourceURL   string         `json:"sourceUrl,omitempty"`
	SourceFile  string         `json:"sourceFile,omitempty"`
	Sections    []Section      `json:"sections"`
	LayoutTable []FieldElement `json:"layoutTable"`
}

