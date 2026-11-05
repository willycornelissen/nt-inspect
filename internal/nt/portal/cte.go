package portal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"nt-inspect/internal/model"
)

// CTEProvider queries the Portal Nacional do CT-e for Technical Notes.
type CTEProvider struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewCTEProvider returns a configured CTEProvider.
func NewCTEProvider(client *http.Client) *CTEProvider {
	if client == nil {
		client = DefaultHTTPClient()
	}
	return &CTEProvider{
		BaseURL:    "https://www.cte.fazenda.gov.br/portal/listaConteudo.aspx?tipoConteudo=Y0nErnoZpsg=",
		HTTPClient: client,
	}
}

func (p *CTEProvider) DocType() model.DocType {
	return model.DocTypeCTE
}

// FindLatestNT scrapes or searches the CT-e documentation page for matching Technical Notes.
func (p *CTEProvider) FindLatestNT(ctx context.Context, ntID model.NTIdentifier) (*RemoteNT, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.BaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create portal request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) nt-inspect/1.0")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query CT-e portal %s: %w", p.BaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CT-e portal returned HTTP status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read CT-e portal response: %w", err)
	}

	remoteList, err := p.parsePortalHTML(string(body), ntID)
	if err != nil {
		return nil, err
	}

	if len(remoteList) == 0 {
		return nil, fmt.Errorf("no technical note found matching code %s on CT-e portal", ntID.Code)
	}

	// Sort matching revisions descending by VersionScore
	sort.Slice(remoteList, func(i, j int) bool {
		return VersionScore(remoteList[i].Version) > VersionScore(remoteList[j].Version)
	})

	return remoteList[0], nil
}

// parsePortalHTML extracts links and titles for the targeted NT code from HTML.
func (p *CTEProvider) parsePortalHTML(html string, ntID model.NTIdentifier) ([]*RemoteNT, error) {
	var results []*RemoteNT

	// Pattern matches links referencing downloads/files with NT titles
	linkRegex := regexp.MustCompile(`(?i)<a[^>]+href=["']([^"']+)["'][^>]*>([\s\S]*?)<\/a>`)
	matches := linkRegex.FindAllStringSubmatch(html, -1)

	baseParsed, _ := url.Parse(p.BaseURL)

	for _, m := range matches {
		href := strings.TrimSpace(m[1])
		rawTitle := strings.TrimSpace(regexp.MustCompile(`<[^>]+>`).ReplaceAllString(m[2], " "))
		rawTitle = regexp.MustCompile(`\s+`).ReplaceAllString(rawTitle, " ")

		// Check if this link refers to our target NT code (e.g. "2025.001" or "2025_001")
		matchCode := strings.Contains(rawTitle, ntID.Code) ||
			strings.Contains(rawTitle, fmt.Sprintf("%s_%s", ntID.Year, ntID.Seq)) ||
			strings.Contains(href, ntID.Code) ||
			strings.Contains(href, fmt.Sprintf("%s_%s", ntID.Year, ntID.Seq))

		if !matchCode {
			continue
		}

		// Clean and resolve absolute URL
		cleanedHref := CleanPortalURL(href)
		relURL, err := url.Parse(cleanedHref)
		if err != nil {
			continue
		}
		absURL := baseParsed.ResolveReference(relURL).String()

		// Extract version if present in title or URL (e.g., v1.14a)
		verMatch := regexp.MustCompile(`(?i)v(?:ers[aã]o)?\s*([0-9]+\.[0-9]+[a-z]?)`).FindStringSubmatch(rawTitle)
		version := ""
		if len(verMatch) >= 2 {
			version = verMatch[1]
		} else {
			verMatchHref := regexp.MustCompile(`(?i)v([0-9]+\.[0-9]+[a-z]?)`).FindStringSubmatch(href)
			if len(verMatchHref) >= 2 {
				version = verMatchHref[1]
			}
		}

		results = append(results, &RemoteNT{
			Code:        ntID.Code,
			Version:     version,
			Title:       rawTitle,
			DownloadURL: absURL,
		})
	}

	return results, nil
}

// DownloadPDF downloads the binary from DownloadURL and unzips if needed.
func (p *CTEProvider) DownloadPDF(ctx context.Context, remote *RemoteNT) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, remote.DownloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create download request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) nt-inspect/1.0")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download from %s: %w", remote.DownloadURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned HTTP %d for %s", resp.StatusCode, remote.DownloadURL)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read download content: %w", err)
	}

	if valErr := ValidatePDFOrZip(data); valErr != nil {
		return nil, fmt.Errorf("validação do download de %s: %w", remote.DownloadURL, valErr)
	}

	return ExtractPDFFromZipIfApplicable(data)
}

