package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nt-inspect/internal/model"
)

// Store manages disk cache for downloaded NT PDFs and parsed DocumentSpec JSON files.
type Store struct {
	BaseDir string
}

// DefaultStore returns a Store located in standard OS user cache directory (~/.cache/nt-inspect).
func DefaultStore() (*Store, error) {
	userCache, err := os.UserCacheDir()
	if err != nil {
		userCache = os.TempDir()
	}
	dir := filepath.Join(userCache, "nt-inspect")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("could not create cache directory %s: %w", dir, err)
	}
	return &Store{BaseDir: dir}, nil
}

func sanitizeKey(s string) string {
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "/", "_")
	return strings.ToLower(s)
}

// PDFPath returns the filesystem path where the NT PDF is cached.
func (s *Store) PDFPath(doc model.DocType, ntCode, version string) string {
	vPart := version
	if vPart == "" {
		vPart = "latest"
	}
	filename := fmt.Sprintf("%s_%s_%s.pdf", string(doc), sanitizeKey(ntCode), sanitizeKey(vPart))
	return filepath.Join(s.BaseDir, string(doc), filename)
}

// SpecPath returns the filesystem path where the parsed DocumentSpec JSON is cached.
func (s *Store) SpecPath(doc model.DocType, ntCode, version string) string {
	vPart := version
	if vPart == "" {
		vPart = "latest"
	}
	filename := fmt.Sprintf("%s_%s_%s.json", string(doc), sanitizeKey(ntCode), sanitizeKey(vPart))
	return filepath.Join(s.BaseDir, string(doc), filename)
}

// SavePDF saves PDF bytes to cache.
func (s *Store) SavePDF(doc model.DocType, ntCode, version string, data []byte) (string, error) {
	path := s.PDFPath(doc, ntCode, version)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write cached pdf %s: %w", path, err)
	}
	return path, nil
}

// GetPDF reads PDF bytes from cache if present.
func (s *Store) GetPDF(doc model.DocType, ntCode, version string) ([]byte, string, error) {
	path := s.PDFPath(doc, ntCode, version)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	return data, path, nil
}

// SaveSpec serializes and saves a DocumentSpec to cache.
func (s *Store) SaveSpec(spec *model.DocumentSpec) error {
	path := s.SpecPath(spec.DocType, spec.NTCode, spec.Version)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal spec: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// GetSpec loads a cached DocumentSpec if present.
func (s *Store) GetSpec(doc model.DocType, ntCode, version string) (*model.DocumentSpec, error) {
	path := s.SpecPath(doc, ntCode, version)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var spec model.DocumentSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("unmarshal cached spec %s: %w", path, err)
	}
	return &spec, nil
}

