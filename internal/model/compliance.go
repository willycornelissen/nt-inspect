package model

// ComplianceStatus represents the overall verdict of the compliance inspection.
type ComplianceStatus string

const (
	StatusCompliant    ComplianceStatus = "COMPLIANT"
	StatusNonCompliant ComplianceStatus = "NON_COMPLIANT"
	StatusError        ComplianceStatus = "ERROR"
)

// FieldFindingKind represents the classification of a specific field inspection result.
type FieldFindingKind string

const (
	FindingMatch        FieldFindingKind = "MATCH"
	FindingMissing      FieldFindingKind = "MISSING"
	FindingObsolete     FieldFindingKind = "OBSOLETE"
	FindingNameMismatch FieldFindingKind = "NAME_MISMATCH"
)

// FieldFinding details an individual finding for a field or group.
type FieldFinding struct {
	Kind         FieldFindingKind `json:"kind"`
	FieldName    string           `json:"fieldName"`
	ParentName   string           `json:"parentName,omitempty"`
	ExpectedPath string           `json:"expectedPath,omitempty"`
	FoundPath    string           `json:"foundPath,omitempty"`
	Location     string           `json:"location,omitempty"` // e.g. "IBSCBSMapper.java:24"
	Message      string           `json:"message"`
	Suggestion   string           `json:"suggestion,omitempty"`
}

// ComplianceSummary aggregates statistical counters of the inspection.
type ComplianceSummary struct {
	TotalRequired  int     `json:"totalRequired"`
	MatchedCount   int     `json:"matchedCount"`
	MissingCount   int     `json:"missingCount"`
	ObsoleteCount  int     `json:"obsoleteCount"`
	MismatchCount  int     `json:"mismatchCount"`
	ComplianceRate float64 `json:"complianceRate"` // 0.0 to 1.0 (e.g. 0.95 = 95%)
}

// InspectionReport represents the complete audit report produced by nt-inspect.
type InspectionReport struct {
	Status       ComplianceStatus  `json:"status"`
	ExitCode     int               `json:"exitCode"` // 0 for compliant, 1 for non-compliant, 2 for error
	DocType      DocType           `json:"docType"`
	NTCode       string            `json:"ntCode"`
	NTVersion    string            `json:"ntVersion,omitempty"`
	TargetDir    string            `json:"targetDir"`
	Summary      ComplianceSummary `json:"summary"`
	Findings     []FieldFinding    `json:"findings"`
}

