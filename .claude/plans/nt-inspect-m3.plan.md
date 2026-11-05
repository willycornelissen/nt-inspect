# Plan: Milestone 3 — Geração Persistente de Relatório de Gaps

**Source PRD**: `.claude/prds/nt-inspect.prd.md`
**Selected Milestone**: `3 — Relatório confiável de gaps`
**Complexity**: Small

## Summary
Adicionar ao `nt-inspect` a capacidade de gerar automaticamente relatórios de conformidade persistentes em Markdown (`--output` ou padrão `./compliance-report-nt<codigo>.md`), detalhando a auditoria completa da NT (título, versão, data, resumo estatístico, tabela de campos aderentes e detalhamento acionável de cada gap com arquivo/linha de código e sugestão de correção), permitindo que o desenvolvedor versionador registre o plano de adequação no próprio repositório auditado (assim como feito historicamente em `cte-parser/nt/PLANO-CONFORMIDADE-NT-2025.001.md`).

## Patterns to Mirror
| Category | Source | Pattern |
|---|---|---|
| Historical Reference | `cte-parser/nt/PLANO-CONFORMIDADE-NT-2025.001.md` (no git commit `bc1fee5`) | Estrutura de relatório contendo resumo executivo, identificação da NT, tabelas por grupo (IBS/CBS, Compras Gov, Modais) com colunas `Campo | Situação | Ação necessária`. |
| Go CLI Output | `.agents/skills/golang-patterns/SKILL.md:21-65` | Gerador puro e desacoplado (`internal/report`) que recebe `InspectionReport` e escreve em `io.Writer` ou arquivo, com testes table-driven. |

## Files to Change
| File | Action | Why |
|---|---|---|
| `internal/report/markdown.go` | CREATE | Renderizador de Markdown formatando o `InspectionReport` em documento profissional com tabelas e checklist de ações. |
| `internal/report/markdown_test.go` | CREATE | Testes unitários validando a geração do Markdown a partir de relatórios mockados com e sem gaps. |
| `internal/cli/root.go` | UPDATE | Adicionar flags `--output` / `-o` (caminho personalizado do arquivo de relatório) e `--save-report` (ou geração padrão de `.md` no diretório alvo), notificando o usuário sobre o arquivo criado. |
| `.claude/prds/nt-inspect.prd.md` | UPDATE | Marcar o Milestone 3 como `in-progress` e vincular este plano. |

## Tasks

### Task 1: Renderizador de Relatório em Markdown (`internal/report`)
- **Action**: Implementar `internal/report/markdown.go` que converte um `model.InspectionReport` em Markdown padronizado:
  - Cabeçalho com metadados (Documento fiscal, NT, versão, data de execução, diretório do parser).
  - Veredito de conformidade e badge/resumo estatístico (`Total de campos`, `Mapeados`, `Faltantes`, `Obsoletos`, `% Aderência`).
  - Seção detalhada dos **Gaps de Conformidade**:
    - Tabela de campos faltantes com `Tag`, `Pai`, `Path XML`, `Tipo/Ocorrência` e `Ação recomendada`.
    - Tabela de nomes divergentes com `Nome no Código`, `Esperado pela NT`, `Localização` e `Ação recomendada`.
    - Tabela de campos obsoletos/revogados com `Campo`, `Localização` e `Motivo`.
  - Seção dos **Campos em Conformidade** (resumo dos itens validados).
- **Mirror**: Estrutura de seções de `cte-parser/nt/PLANO-CONFORMIDADE-NT-2025.001.md`.
- **Validate**: `go test -v ./internal/report/...`

### Task 2: Integração de Salvamento e Flag de Saída na CLI
- **Action**: Atualizar `internal/cli/root.go` para suportar:
  - `--output` / `-o <caminho>`: caminho explícito para salvar o relatório Markdown.
  - Se `--output` não for informado e não estiver em modo `--json`, gerar automaticamente `./compliance-report-nt<codigo>.md` no diretório de execução atual.
  - Exibir no terminal a mensagem informando a criação do relatório persistente:
    `[+] Relatório de conformidade salvo em: ./compliance-report-nt2025.001.md`
- **Mirror**: Padrões de flags do Cobra já utilizados no Milestone 1 e 2.
- **Validate**: `go test -v ./internal/cli/... && go build -o bin/nt-inspect ./cmd/nt-inspect`

### Task 3: Validação Integrada do Milestone 3
- **Action**: Executar o CLI contra o repositório `cte-parser` com salvamento de relatório:
  - Validar a criação do arquivo Markdown.
  - Verificar a integridade do conteúdo e consistência com os testes de conformidade.
- **Mirror**: Verificação ponta a ponta dos marcos anteriores.
- **Validate**: `./bin/nt-inspect 2025.001 --dir cte-parser -o /tmp/report-test.md && head -n 40 /tmp/report-test.md`

## Validation
```bash
go vet ./...
go test -race -cover ./...
go build -o bin/nt-inspect ./cmd/nt-inspect
./bin/nt-inspect 2025.001 --dir cte-parser -o /tmp/report-test.md
```

## Risks
| Risk | Likelihood | Mitigation |
|---|---|---|
| Permissão de escrita no diretório alvo ao gerar o `.md` padrão | Low | Em caso de falha de I/O no arquivo, emitir aviso claro no `stderr` sem abortar a exibição do resumo no terminal |

## Acceptance
- [x] All tasks complete
- [x] Validation passes
- [x] Patterns mirrored, not reinvented

