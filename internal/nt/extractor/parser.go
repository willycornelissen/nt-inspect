package extractor

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"nt-inspect/internal/model"
)

// LayoutParser extracts metadata, chapters, and layout table lines from NT text.
type LayoutParser struct {
	Enricher SemanticEnricher
}

// NewLayoutParser returns an initialized parser.
func NewLayoutParser(enricher SemanticEnricher) *LayoutParser {
	if enricher == nil {
		enricher = &NoopEnricher{}
	}
	return &LayoutParser{Enricher: enricher}
}

// Parse extracts a structured model.DocumentSpec from the full text of an NT.
func (p *LayoutParser) Parse(ctx context.Context, docType model.DocType, ntCode string, text string) (*model.DocumentSpec, error) {
	spec := &model.DocumentSpec{
		DocType: docType,
		NTCode:  ntCode,
	}

	// 1. Extract Title, Version, and Date
	p.extractHeaderMetadata(spec, text)

	// 2. Extract Sections / Chapters
	spec.Sections = p.extractSections(text)

	// 3. Extract Layout Table fields
	spec.LayoutTable = p.extractLayoutFields(text)

	// 4. Resolve full XML paths (Pai -> Campo)
	p.resolveXMLPaths(spec.LayoutTable)

	// 5. Optional semantic enrichment (LLM hook)
	if err := p.Enricher.EnrichSpec(ctx, spec, text); err != nil {
		return nil, fmt.Errorf("semantic enrichment: %w", err)
	}

	return spec, nil
}

func (p *LayoutParser) extractHeaderMetadata(spec *model.DocumentSpec, text string) {
	titleRegex := regexp.MustCompile(`(?i)(?:Projeto\s+)?([^\n\r]+Nota\s+T[ée]cnica\s+[^\n\r]+)`)
	if m := titleRegex.FindStringSubmatch(text); len(m) > 1 {
		spec.Title = strings.TrimSpace(m[1])
	} else {
		spec.Title = fmt.Sprintf("Nota Técnica %s", spec.NTCode)
	}

	verRegex := regexp.MustCompile(`(?i)Vers[aã]o\s+([0-9]+\.[0-9]+[a-z]?)`)
	if m := verRegex.FindStringSubmatch(text); len(m) > 1 {
		spec.Version = m[1]
	}

	dateRegex := regexp.MustCompile(`(?i)(?:de\s+)?(\d{1,2}\s+de\s+[a-zA-ZçÇ]+\s+de\s+\d{4})`)
	if m := dateRegex.FindStringSubmatch(text); len(m) > 1 {
		spec.Date = m[1]
	}
}

func (p *LayoutParser) extractSections(text string) []model.Section {
	var sections []model.Section
	// Pattern like: "3   Alteração do leiaute do imposto do CTe" or "1 Introdução"
	secRegex := regexp.MustCompile(`(?m)^([0-9]{1,2})\s+([A-ZÀ-Ú][^\n\r]{4,80})$`)
	matches := secRegex.FindAllStringSubmatch(text, -1)

	for _, m := range matches {
		num, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		title := strings.TrimSpace(m[2])
		// Avoid false positives like table row numbers
		if strings.HasPrefix(title, "CST") || strings.HasPrefix(title, "vBC") {
			continue
		}
		sections = append(sections, model.Section{
			Number: num,
			Title:  title,
		})
	}
	return sections
}

func (p *LayoutParser) extractLayoutFields(text string) []model.FieldElement {
	var fields []model.FieldElement

	lines := strings.Split(text, "\n")
	tokens := make([]string, 0, len(lines))
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		// Filter out page headers/footers
		if strings.HasPrefix(trimmed, "Página ") || strings.HasPrefix(trimmed, "Projeto") || strings.HasPrefix(trimmed, "Conhecimento de") || strings.HasPrefix(trimmed, "NT 2025") {
			continue
		}
		if trimmed != "" {
			tokens = append(tokens, trimmed)
		}
	}

	// In pdftotext -layout or standard vertical flow, table columns appear either on single rows or sequentially:
	// # \n Campo \n Ele \n Pai \n Tipo \n Ocor. \n Tam. \n Descrição
	// Or tab/space delimited: "1   CST   E   IBSCBS   N   1-1   3   Código..."
	singleLineRegex := regexp.MustCompile(`^(\d+)\s+([a-zA-Z0-9_]+)\s+([GEA]|CE)\s+([a-zA-Z0-9_]+)\s+([NCD\-])\s+([01]-[01Nn]|\-)\s+([0-9,\-vV]+)?\s*(.*)$`)

	seenIDs := make(map[string]bool)

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]

		// 1. Try single-line table row
		if m := singleLineRegex.FindStringSubmatch(t); len(m) > 5 {
			id, _ := strconv.Atoi(m[1])
			name := m[2]
			key := fmt.Sprintf("%d-%s", id, name)
			if !seenIDs[key] {
				seenIDs[key] = true
				fields = append(fields, model.FieldElement{
					ID:          id,
					Name:        name,
					Kind:        model.FieldKind(m[3]),
					Parent:      m[4],
					DataType:    m[5],
					Occurrence:  m[6],
					Length:      strings.TrimSpace(m[7]),
					Description: strings.TrimSpace(m[8]),
				})
			}
			continue
		}

		// 2. Try vertical sequential tokens:
		// [i]   = # (e.g. "1")
		// [i+1] = Campo (e.g. "CST")
		// [i+2] = Ele (e.g. "E" or "G")
		// [i+3] = Pai (e.g. "IBSCBS")
		if _, err := strconv.Atoi(t); err == nil && i+4 < len(tokens) {
			id, _ := strconv.Atoi(t)
			name := tokens[i+1]
			kind := tokens[i+2]
			parent := tokens[i+3]

			if (kind == "E" || kind == "G" || kind == "A" || kind == "CE") &&
				len(name) >= 2 && len(parent) >= 2 &&
				!strings.Contains(name, " ") && !strings.Contains(parent, " ") {

				dataType := ""
				occur := ""
				length := ""
				desc := ""

				// Look ahead for Tipo, Ocor, etc.
				idx := i + 4
				if idx < len(tokens) {
					dataType = tokens[idx]
					idx++
				}
				if idx < len(tokens) {
					occur = tokens[idx]
					idx++
				}
				if idx < len(tokens) {
					length = tokens[idx]
					idx++
				}
				if idx < len(tokens) {
					desc = tokens[idx]
				}

				key := fmt.Sprintf("%d-%s", id, name)
				if !seenIDs[key] {
					seenIDs[key] = true
					fields = append(fields, model.FieldElement{
						ID:          id,
						Name:        name,
						Kind:        model.FieldKind(kind),
						Parent:      parent,
						DataType:    dataType,
						Occurrence:  occur,
						Length:      length,
						Description: desc,
					})
					i += 4 // advance token index
				}
			}
		}
	}

	return fields
}

// resolveXMLPaths builds hierarchical tag paths (e.g. IMP.IBSCBS.GIBSCBS.VIBS).
func (p *LayoutParser) resolveXMLPaths(fields []model.FieldElement) {
	parentMap := make(map[string]string) // child -> parent

	for _, f := range fields {
		if f.Parent != "" && f.Parent != "-" {
			parentMap[f.Name] = f.Parent
		}
	}

	for i := range fields {
		curr := fields[i].Name
		pathParts := []string{curr}

		visited := make(map[string]bool)
		for {
			visited[curr] = true
			parent, ok := parentMap[curr]
			if !ok || parent == "" || parent == "-" || visited[parent] {
				break
			}
			pathParts = append([]string{parent}, pathParts...)
			curr = parent
		}

		fields[i].XMLPath = strings.ToUpper(strings.Join(pathParts, "."))
	}
}

