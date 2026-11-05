package portal

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"nt-inspect/internal/model"
)

func TestVersionScore(t *testing.T) {
	tests := []struct {
		v1   string
		v2   string
		more bool // v1 > v2
	}{
		{"1.14a", "1.14", true},
		{"1.14", "1.10", true},
		{"1.10", "1.00", true},
		{"v1.14a", "v1.14", true},
		{"versão 1.14a", "1.14", true},
	}

	for _, tt := range tests {
		s1 := VersionScore(tt.v1)
		s2 := VersionScore(tt.v2)
		if (s1 > s2) != tt.more {
			t.Errorf("VersionScore(%s)=%f vs VersionScore(%s)=%f, expected more=%v", tt.v1, s1, tt.v2, s2, tt.more)
		}
	}
}

func TestCTEProviderFindAndDownload(t *testing.T) {
	// Create mock HTTP server serving portal HTML and mock file
	mockHTML := `
	<html>
	<body>
		<a href="/portal/download.aspx?id=101">Nota Técnica 2025.001 v1.00 - Reforma Tributária</a>
		<a href="/portal/download.aspx?id=102">Nota Técnica 2025.001 v1.14 - Atualização</a>
		<a href="/portal/download.aspx?id=103">Nota Técnica 2025.001 v1.14a - Reforma Tributária Final</a>
		<a href="/portal/download.aspx?id=999">Nota Técnica 2024.001 v1.00 - Outra</a>
	</body>
	</html>
	`

	pdfContent := []byte("%PDF-1.4 test data")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/portal/listaConteudo.aspx":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(mockHTML))
		case "/portal/download.aspx":
			// If requested latest id 103, return a zip containing the pdf
			if r.URL.Query().Get("id") == "103" {
				var buf bytes.Buffer
				zw := zip.NewWriter(&buf)
				f, _ := zw.Create("NT2025.001_v1.14a.pdf")
				_, _ = f.Write(pdfContent)
				_ = zw.Close()
				w.Header().Set("Content-Type", "application/zip")
				_, _ = w.Write(buf.Bytes())
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(pdfContent)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := &CTEProvider{
		BaseURL:    server.URL + "/portal/listaConteudo.aspx",
		HTTPClient: server.Client(),
	}

	ntID, _ := model.ParseNTIdentifier("2025.001", model.DocTypeCTE)
	remote, err := provider.FindLatestNT(context.Background(), ntID)
	if err != nil {
		t.Fatalf("FindLatestNT failed: %v", err)
	}

	if remote.Version != "1.14a" {
		t.Fatalf("expected version 1.14a, got %s", remote.Version)
	}

	expectedURL := fmt.Sprintf("%s/portal/download.aspx?id=103", server.URL)
	if remote.DownloadURL != expectedURL {
		t.Fatalf("expected URL %s, got %s", expectedURL, remote.DownloadURL)
	}

	// Download and extract from zip test
	data, err := provider.DownloadPDF(context.Background(), remote)
	if err != nil {
		t.Fatalf("DownloadPDF failed: %v", err)
	}
	if !bytes.Equal(data, pdfContent) {
		t.Fatalf("downloaded data does not match original pdf content")
	}
}

