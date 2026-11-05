package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"nt-inspect/internal/detector"
	"nt-inspect/internal/inspector"
	"nt-inspect/internal/model"
	"nt-inspect/internal/nt/cache"
	"nt-inspect/internal/nt/extractor"
	"nt-inspect/internal/nt/portal"
	"nt-inspect/internal/report"
	"nt-inspect/internal/scanner"
)

type RootOptions struct {
	DocType    string
	PDFPath    string
	TargetDir  string
	OutputPath string
	Refresh    bool
	JSONOutput bool
	NoSave     bool
}

// NewRootCmd constructs the primary CLI command for nt-inspect.
func NewRootCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	opts := &RootOptions{}

	cmd := &cobra.Command{
		Use:   "nt-inspect <codigo-nt>",
		Short: "Inspeciona conformidade de parsers de documentos fiscais contra Notas Técnicas",
		Long: `nt-inspect resolve a Nota Técnica indicada (ex.: 2025.001), baixa e extrai
as especificações de leiaute oficial e inspeciona se o código do parser no diretório
está em conformidade com as alterações, gerando um relatório detalhado.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if out == nil {
				out = cmd.OutOrStdout()
			}
			if errOut == nil {
				errOut = cmd.ErrOrStderr()
			}
			return runInspect(cmd.Context(), args[0], opts, out, errOut)
		},
	}

	cmd.Flags().StringVarP(&opts.DocType, "doc", "d", "", "Tipo de documento fiscal (cte, nfe, mdfe, nfce)")
	cmd.Flags().StringVar(&opts.PDFPath, "pdf", "", "Caminho direto para arquivo PDF da NT local (contingência offline)")
	cmd.Flags().StringVar(&opts.TargetDir, "dir", "", "Diretório do parser a inspecionar (padrão: diretório atual)")
	cmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "", "Caminho do arquivo Markdown de relatório (padrão: ./compliance-report-nt<codigo>.md)")
	cmd.Flags().BoolVarP(&opts.Refresh, "refresh", "r", false, "Ignora cache e busca novamente a NT no portal oficial")
	cmd.Flags().BoolVar(&opts.JSONOutput, "json", false, "Emite o resultado estruturado em formato JSON")
	cmd.Flags().BoolVar(&opts.NoSave, "no-save", false, "Desativa a geração do arquivo de relatório Markdown")

	return cmd
}

func runInspect(ctx context.Context, rawNTCode string, opts *RootOptions, out io.Writer, errOut io.Writer) error {
	// 1. Detect or validate DocType and target directory
	targetDir := opts.TargetDir
	if targetDir == "" {
		var err error
		targetDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("obter diretório atual: %w", err)
		}
	}

	docType, err := detector.DetectDocType(targetDir, opts.DocType)
	if err != nil {
		return err
	}

	// 2. Parse NT Identifier
	ntID, err := model.ParseNTIdentifier(rawNTCode, docType)
	if err != nil {
		return err
	}

	if !opts.JSONOutput {
		_, _ = fmt.Fprintf(errOut, "[+] Documento fiscal detectado: %s\n", docType)
	}

	// 3. Setup Cache and Extractor
	cacheStore, err := cache.DefaultStore()
	if err != nil {
		return fmt.Errorf("setup cache: %w", err)
	}

	textExtractor, err := extractor.NewPopplerTextExtractor()
	if err != nil {
		return err
	}

	layoutParser := extractor.NewLayoutParser(nil)

	var spec *model.DocumentSpec
	var pdfBytes []byte
	var sourceDesc string

	// Check if cached spec exists and refresh is not requested
	if !opts.Refresh && opts.PDFPath == "" {
		if cachedSpec, err := cacheStore.GetSpec(docType, ntID.Code, ntID.Version); err == nil && cachedSpec != nil {
			spec = cachedSpec
			sourceDesc = "cache local"
		}
	}

	// If not retrieved from spec cache:
	if spec == nil {
		if opts.PDFPath != "" {
			// Local PDF provided via flag
			pdfBytes, err = os.ReadFile(opts.PDFPath)
			if err != nil {
				return fmt.Errorf("read local pdf %s: %w", opts.PDFPath, err)
			}
			sourceDesc = fmt.Sprintf("arquivo local (%s)", opts.PDFPath)
		} else {
			// Check if a local NT file exists in targetDir (e.g. targetDir/nt/NT2025.001.pdf or nt/cte.pdf)
			localCandidates := []string{
				filepath.Join(targetDir, "nt", fmt.Sprintf("NT%s.pdf", ntID.Code)),
				filepath.Join(targetDir, "nt", fmt.Sprintf("nt_%s.pdf", ntID.Code)),
				filepath.Join(targetDir, "nt", fmt.Sprintf("%s.pdf", ntID.Code)),
				filepath.Join(targetDir, "nt", "cte.pdf"),
				filepath.Join(targetDir, fmt.Sprintf("NT%s.pdf", ntID.Code)),
			}

			var foundLocal string
			for _, c := range localCandidates {
				if fi, statErr := os.Stat(c); statErr == nil && !fi.IsDir() && fi.Size() > 1000 {
					foundLocal = c
					break
				}
			}

			if foundLocal != "" {
				pdfBytes, err = os.ReadFile(foundLocal)
				if err != nil {
					return fmt.Errorf("ler arquivo local de NT em %s: %w", foundLocal, err)
				}
				sourceDesc = fmt.Sprintf("arquivo local do projeto (%s)", foundLocal)
			} else {
				// Query portal
				if !opts.JSONOutput {
					_, _ = fmt.Fprintf(errOut, "[+] Consultando portal oficial de %s para NT %s...\n", docType, ntID.Code)
				}

				var prov portal.Provider
				switch docType {
				case model.DocTypeCTE:
					prov = portal.NewCTEProvider(nil)
				default:
					return fmt.Errorf("portal provider for doc type %s is not yet implemented (use --pdf)", docType)
				}

				remote, err := prov.FindLatestNT(ctx, ntID)
				if err != nil {
					return fmt.Errorf("localizar NT no portal: %w", err)
				}

				if !opts.JSONOutput {
					_, _ = fmt.Fprintf(errOut, "[+] Baixando %s (%s)...\n", remote.Title, remote.DownloadURL)
				}

				pdfBytes, err = prov.DownloadPDF(ctx, remote)
				if err != nil {
					return fmt.Errorf("download da NT: %w", err)
				}

				// Save to cache
				_, _ = cacheStore.SavePDF(docType, ntID.Code, remote.Version, pdfBytes)
				sourceDesc = fmt.Sprintf("portal oficial (%s)", remote.Version)
			}
		}

		// Extract text from PDF
		plainText, err := textExtractor.ExtractText(ctx, pdfBytes)
		if err != nil {
			return fmt.Errorf("extrair texto do PDF: %w", err)
		}

		spec, err = layoutParser.Parse(ctx, docType, ntID.Code, plainText)
		if err != nil {
			return fmt.Errorf("parsear estrutura da NT: %w", err)
		}
		spec.SourceFile = sourceDesc

		// Save parsed spec to cache
		_ = cacheStore.SaveSpec(spec)
	}

	if !opts.JSONOutput {
		_, _ = fmt.Fprintf(errOut, "[+] NT %s obtida via %s\n", spec.NTCode, sourceDesc)
		_, _ = fmt.Fprintf(errOut, "[+] Inspecionando código-fonte em %s...\n\n", targetDir)
	}

	// 4. Scan codebase
	javaScanner := scanner.NewJavaScanner(docType)
	codebase, err := javaScanner.Scan(ctx, targetDir)
	if err != nil {
		return fmt.Errorf("inspeção do código-fonte: %w", err)
	}

	// 5. Inspect compliance against NT specification
	complianceInspector := inspector.NewInspector()
	inspectionRep, err := complianceInspector.Inspect(ctx, spec, codebase)
	if err != nil {
		return fmt.Errorf("cruzamento de conformidade: %w", err)
	}

	// 6. Save persistent Markdown report if not disabled and not JSON
	if !opts.JSONOutput && !opts.NoSave {
		saveReportPath := opts.OutputPath
		if saveReportPath == "" {
			cleanCode := ntID.Code
			saveReportPath = filepath.Join(targetDir, fmt.Sprintf("compliance-report-nt%s.md", cleanCode))
		}

		if repErr := saveMarkdownReport(saveReportPath, inspectionRep); repErr == nil {
			_, _ = fmt.Fprintf(errOut, "[+] Relatório de conformidade salvo em: %s\n\n", saveReportPath)
		} else {
			_, _ = fmt.Fprintf(errOut, "[!] Aviso: não foi possível salvar o relatório em %s: %v\n\n", saveReportPath, repErr)
		}
	}

	// 7. Render Output
	if opts.JSONOutput {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(inspectionRep); err != nil {
			return err
		}
	} else {
		renderComplianceReport(out, inspectionRep)
	}

	if inspectionRep.Status != model.StatusCompliant {
		// Exit code 1 for non-compliant
		os.Exit(1)
	}

	return nil
}

func saveMarkdownReport(filePath string, rep *model.InspectionReport) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return report.GenerateMarkdown(f, rep)
}

func renderComplianceReport(out io.Writer, report *model.InspectionReport) {
	if report.Status == model.StatusCompliant {
		_, _ = fmt.Fprintf(out, "STATUS: EM CONFORMIDADE (%s)\n", report.Status)
	} else {
		totalGaps := report.Summary.MissingCount + report.Summary.MismatchCount + report.Summary.ObsoleteCount
		_, _ = fmt.Fprintf(out, "STATUS: NÃO CONFORME (%s) — %d GAPS ENCONTRADOS\n", report.Status, totalGaps)
	}
	_, _ = fmt.Fprintln(out, "------------------------------------------------------------")
	_, _ = fmt.Fprintf(out, "Nota Técnica:    %s (versão: %s)\n", report.NTCode, report.NTVersion)
	_, _ = fmt.Fprintf(out, "Escopo avaliado: %d campos de leiaute\n", report.Summary.TotalRequired)
	_, _ = fmt.Fprintf(out, "Conformidade:    %.1f%% aderente (%d/%d campos)\n\n",
		report.Summary.ComplianceRate*100.0, report.Summary.MatchedCount, report.Summary.TotalRequired)

	// Group findings
	var missing []model.FieldFinding
	var mismatches []model.FieldFinding
	var obsolete []model.FieldFinding

	for _, f := range report.Findings {
		switch f.Kind {
		case model.FindingMissing:
			missing = append(missing, f)
		case model.FindingNameMismatch:
			mismatches = append(mismatches, f)
		case model.FindingObsolete:
			obsolete = append(obsolete, f)
		}
	}

	if len(missing) > 0 {
		_, _ = fmt.Fprintf(out, "[CAMPOS FALTANDO NO CÓDIGO]\n")
		for _, f := range missing {
			_, _ = fmt.Fprintf(out, "  ✖ %-18s (Pai: %s, Path: %s)\n", f.FieldName, f.ParentName, f.ExpectedPath)
			if f.Suggestion != "" {
				_, _ = fmt.Fprintf(out, "    Sugestão: %s\n", f.Suggestion)
			}
		}
		_, _ = fmt.Fprintln(out)
	}

	if len(mismatches) > 0 {
		_, _ = fmt.Fprintf(out, "[NOMES DIVERGENTES / ALTERADOS NA NT]\n")
		for _, f := range mismatches {
			_, _ = fmt.Fprintf(out, "  ⚠ %s (%s)\n", f.FieldName, f.Location)
			_, _ = fmt.Fprintf(out, "    Mensagem: %s\n", f.Message)
			if f.Suggestion != "" {
				_, _ = fmt.Fprintf(out, "    Sugestão: %s\n", f.Suggestion)
			}
		}
		_, _ = fmt.Fprintln(out)
	}

	if len(obsolete) > 0 {
		_, _ = fmt.Fprintf(out, "[CAMPOS OBSOLETOS OU REVOGADOS PELA NT DETECTADOS NO CÓDIGO]\n")
		seen := make(map[string]bool)
		for _, f := range obsolete {
			if !seen[f.FieldName] {
				seen[f.FieldName] = true
				_, _ = fmt.Fprintf(out, "  ⚠ %s (%s)\n", f.FieldName, f.Location)
				_, _ = fmt.Fprintf(out, "    Mensagem: %s\n", f.Message)
			}
		}
		_, _ = fmt.Fprintln(out)
	}

	if report.Status == model.StatusCompliant {
		_, _ = fmt.Fprintf(out, "✔ Todos os campos exigidos pela NT foram mapeados com sucesso no parser.\n")
	}
}
