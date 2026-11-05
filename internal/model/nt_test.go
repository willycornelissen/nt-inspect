package model

import (
	"testing"
)

func TestParseDocType(t *testing.T) {
	tests := []struct {
		input   string
		want    DocType
		wantErr bool
	}{
		{"cte", DocTypeCTE, false},
		{"CT-e", DocTypeCTE, false},
		{"CTE", DocTypeCTE, false},
		{"nfe", DocTypeNFE, false},
		{"NF-E", DocTypeNFE, false},
		{"mdfe", DocTypeMDFE, false},
		{"nfce", DocTypeNFCE, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		got, err := ParseDocType(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseDocType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseDocType(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseNTIdentifier(t *testing.T) {
	tests := []struct {
		raw      string
		defDoc   DocType
		wantCode string
		wantYear string
		wantSeq  string
		wantErr  bool
	}{
		{"2025.001", DocTypeCTE, "2025.001", "2025", "001", false},
		{"NT 2025.001", DocTypeCTE, "2025.001", "2025", "001", false},
		{"nt2025.001", DocTypeCTE, "2025.001", "2025", "001", false},
		{"NT 2025_001", DocTypeCTE, "2025.001", "2025", "001", false},
		{"2025001", DocTypeCTE, "2025.001", "2025", "001", false},
		{"invalid", DocTypeCTE, "", "", "", true},
	}

	for _, tt := range tests {
		got, err := ParseNTIdentifier(tt.raw, tt.defDoc)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseNTIdentifier(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			continue
		}
		if !tt.wantErr {
			if got.Code != tt.wantCode || got.Year != tt.wantYear || got.Seq != tt.wantSeq {
				t.Errorf("ParseNTIdentifier(%q) = %+v, want code %s", tt.raw, got, tt.wantCode)
			}
		}
	}
}

