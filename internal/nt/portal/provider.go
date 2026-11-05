package portal

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"nt-inspect/internal/model"
)

// RemoteNT represents a Technical Note entry discovered on a portal.
type RemoteNT struct {
	Code        string // "2025.001"
	Version     string // "v1.14a"
	Title       string
	Date        string
	DownloadURL string
}

// Provider represents a remote repository / portal for a specific fiscal document type.
type Provider interface {
	DocType() model.DocType
	FindLatestNT(ctx context.Context, ntID model.NTIdentifier) (*RemoteNT, error)
	DownloadPDF(ctx context.Context, remote *RemoteNT) ([]byte, error)
}

// CleanPortalURL removes whitespace, linebreaks, and XML entities (like &#xD; &#xA;) from portal URLs.
func CleanPortalURL(rawURL string) string {
	cleaned := rawURL
	// Replace XML encoded entities often found inside ASP.NET HTML hrefs
	cleaned = strings.ReplaceAll(cleaned, "&#xD;", "")
	cleaned = strings.ReplaceAll(cleaned, "&#xA;", "")
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	cleaned = strings.ReplaceAll(cleaned, "\r", "")
	cleaned = strings.ReplaceAll(cleaned, "\n", "")
	cleaned = strings.ReplaceAll(cleaned, "\t", "")

	// Remove unencoded whitespace
	parts := strings.Fields(cleaned)
	cleaned = strings.Join(parts, "")

	return cleaned
}

// ValidatePDFOrZip verifies that the downloaded buffer starts with %PDF or PK zip headers.
func ValidatePDFOrZip(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) < 4 {
		return fmt.Errorf("arquivo baixado está vazio ou truncado (%d bytes)", len(data))
	}

	if bytes.HasPrefix(trimmed, []byte("%PDF")) || bytes.HasPrefix(trimmed, []byte("PK\x03\x04")) {
		return nil
	}

	// If it starts with HTML or DOCTYPE
	if bytes.HasPrefix(trimmed, []byte("<!DOCTYPE")) || bytes.HasPrefix(trimmed, []byte("<html")) || bytes.Contains(trimmed[:min(len(trimmed), 500)], []byte("erro.aspx")) {
		if bytes.Contains(trimmed, []byte("Sistema temporariamente indisponível")) {
			return fmt.Errorf("o portal oficial da SEFAZ retornou página de erro ('Sistema temporariamente indisponível'). Por favor, utilize a flag --pdf <arquivo> para inspecionar com a NT local")
		}
		return fmt.Errorf("o link do portal oficial retornou uma página HTML em vez de um arquivo PDF/ZIP. Utilize a flag --pdf <arquivo> para fornecer o PDF da NT diretamente")
	}

	return fmt.Errorf("o conteúdo recebido não é um PDF nem ZIP válido (início dos dados: %q)", string(trimmed[:min(len(trimmed), 30)]))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// VersionScore computes a comparable float for sorting revisions like "1.00", "1.14", "1.14a".
func VersionScore(verStr string) float64 {
	clean := strings.ToLower(strings.TrimSpace(verStr))

	re := regexp.MustCompile(`(\d+)(?:\.(\d+))?([a-z])?`)
	matches := re.FindStringSubmatch(clean)
	if len(matches) == 0 {
		return 0.0
	}

	major, _ := strconv.Atoi(matches[1])
	minor := 0
	if matches[2] != "" {
		minor, _ = strconv.Atoi(matches[2])
	}

	letterScore := 0.0
	if len(matches) > 3 && matches[3] != "" {
		char := matches[3][0]
		letterScore = float64(char-'a'+1) * 0.001
	}

	return float64(major)*1000.0 + float64(minor) + letterScore
}

// DefaultHTTPClient returns a resilient HTTP client with timeouts.
func DefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 45 * time.Second,
	}
}

// ExtractPDFFromZipIfApplicable inspects data, and if it's a zip archive, extracts the first .pdf found.
func ExtractPDFFromZipIfApplicable(data []byte) ([]byte, error) {
	if !bytes.HasPrefix(data, []byte("PK\x03\x04")) {
		// Not a zip file
		return data, nil
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("read zip archive: %w", err)
	}

	for _, f := range zr.File {
		if strings.EqualFold(filepath.Ext(f.Name), ".pdf") {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("open zip file entry %s: %w", f.Name, err)
			}
			defer rc.Close()
			buf, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("read zip file entry %s: %w", f.Name, err)
			}
			return buf, nil
		}
	}

	return nil, fmt.Errorf("no PDF file found inside downloaded zip archive")
}
