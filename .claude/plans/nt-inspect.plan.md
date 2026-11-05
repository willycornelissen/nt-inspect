# Plan: Milestone 1 — Obtenção e Extração Estruturada da Nota Técnica pelo Código

**Source PRD**: `.claude/prds/nt-inspect.prd.md`
**Selected Milestone**: `1 — Obtenção da NT pelo código`
**Complexity**: Medium

## Summary
Inicializa o aplicativo CLI portátil `nt-inspect` em Go (`spf13/cobra`) capaz de receber o código de uma Nota Técnica (ex.: `2025.001`), inferir automaticamente o modelo de documento fiscal a partir do diretório atual (começando por CT-e, como em `cte-parser/`, com flag `--doc` extensível e fallback `--pdf`), localizar e baixar a versão mais recente da NT no portal oficial com cache em disco, e extrair deterministicamente suas seções e tabelas padronizadas de leiaute XML (com interface pronta para enriquecimento opcional via LLM).

## Patterns to Mirror
*Nota: Não existe código Go prévio neste workspace (`nt-inspect` é um módulo Go greenfield na raiz). As convenções de código Go espelham os guias locais de `.agents/skills/golang-patterns/SKILL.md`, enquanto as convenções de domínio fiscal e detecção de repositório espelham os artefatos reais de `cte-parser/`.*

| Category | Source | Pattern |
|---|---|---|
| Naming (Go) | `.agents/skills/golang-patterns/SKILL.md:21-45` | Pacotes curtos em lowercase (`detector`, `portal`, `extractor`, `cli`); nenhum código Go prévio existe no repositório |
| Errors (Go) | `.agents/skills/golang-patterns/SKILL.md:96-120` | Erros tipados de domínio + wrapping contextual com `fmt.Errorf("context: %w", err)`; sem panics em fluxo normal |
| Logging / Output | `Nenhum código CLI prévio (greenfield)` | `stderr` para progresso/diagnóstico estruturado e `stdout` limpo para resumo humano ou `--json` |
| Data access & Cache | `.agents/skills/golang-patterns/SKILL.md:74-92` | Aceitar interfaces (`Provider`, `HTTPClient`, `SemanticEnricher`) e retornar structs concretas (`DocumentSpec`, `FieldSpec`); cache em `os.UserCacheDir()/nt-inspect` |
| Domain / Target Repo | `cte-parser/pom.xml:6-9` e `cte-parser/src/main/java/br/gov/mg/fazenda/cteparser/mapper/IBSCBSMapper.java:22-89` | Identificação do parser CT-e por `pom.xml` (`cte-parser`) / pacote `cteparser` e representação hierárquica de caminhos de tags XML (`IMP.IBSCBS.GIBSCBS.VIBS`) |
| Tests | `cte-parser/src/test/resources/Cte/` e `.agents/skills/golang-patterns/SKILL.md` | Table-driven tests idiomáticos em Go (`*_test.go` ao lado do pacote) usando fixtures reais de HTML de portal e texto/PDF de NT em `testdata/` |

## Files to Change
| File | Action | Why |
|---|---|---|
| `.gitignore` | CREATE | Ignorar binários compilados (`nt-inspect`, `bin/`), perfis de cobertura e caches temporários sem afetar `cte-parser/` |
| `go.mod` | CREATE | Inicializar módulo Go do CLI `nt-inspect` e dependência `github.com/spf13/cobra` |
| `cmd/nt-inspect/main.go` | CREATE | Entrypoint enxuto repassando contexto/sinais de sistema operacional e código de saída para o pacote CLI |
| `internal/model/nt.go` | CREATE | Tipos de domínio da Nota Técnica (`DocType`, `NTIdentifier`, `DocumentSpec`, `Section`, `FieldElement`, hierarquia `Pai -> Campo`) |
| `internal/detector/detector.go` | CREATE | Inferir automaticamente o `DocType` (`cte`, etc.) inspecionando o diretório de execução atual (`pom.xml`, nomes de pacotes/arquivos) |
| `internal/detector/detector_test.go` | CREATE | Testes table-driven validando detecção automática sobre a estrutura de `cte-parser` e diretórios ambíguos |
| `internal/nt/cache/cache.go` | CREATE | Gerenciar cache local de PDFs e especificações extraídas (`~/.cache/nt-inspect/<doc>/`) com suporte a `--refresh` |
| `internal/nt/cache/cache_test.go` | CREATE | Testes unitários de leitura, escrita e invalidação de cache usando `t.TempDir()` |
| `internal/nt/portal/provider.go` | CREATE | Interface `Provider` extensível por `DocType`, normalizador de código de NT (`2025.001`, `NT 2025.001`) e seletor da versão mais recente (ex.: `v1.14a` > `v1.10`) |
| `internal/nt/portal/cte.go` | CREATE | Implementação do resolvedor/downloader do Portal Nacional do CT-e (listagem oficial de Notas Técnicas + extração de link PDF/ZIP) |
| `internal/nt/portal/cte_test.go` | CREATE | Testes determinísticos com `httptest.Server` e fixtures HTML simulando múltiplas versões publicadas e falhas de rede |
| `internal/nt/extractor/pdf.go` | CREATE | Adaptador de extração de texto de PDF invocando `pdftotext` (disponível no ambiente) com verificação clara de dependência |
| `internal/nt/extractor/parser.go` | CREATE | Parser determinístico de metadados, seções e tabelas de leiaute DF-e (`#`, `Campo`, `Ele`, `Pai`, `Tipo`, `Ocor.`, `Tam.`, `Descrição`) reconstruindo os caminhos XML |
| `internal/nt/extractor/enricher.go` | CREATE | Interface opcional `SemanticEnricher` (modo híbrido: no-op por padrão determinístico, plugável para provedor LLM) |
| `internal/nt/extractor/parser_test.go` | CREATE | Testes table-driven com trechos reais de texto de NT (ex.: grupos `IBSCBS`, `gCompraGov`, `duto`, `infDCe`) garantindo extração fiel de campos e pais |
| `internal/cli/root.go` | CREATE | Comando raiz `nt-inspect <codigo-nt>` com flags `--doc`, `--pdf`, `--dir`, `--refresh`, `--json` e exibição estruturada da NT obtida |
| `internal/cli/root_test.go` | CREATE | Testes de ponta a ponta da CLI usando `--pdf` / provider mockado validando códigos de saída e formato do output |
| `.claude/prds/nt-inspect.prd.md` | UPDATE | Marcar o Milestone 1 como `in-progress` e vincular este plano |

## Tasks

### Task 1: Inicialização do Módulo Go e Modelos de Domínio da NT
- **Action**: Criar `.gitignore`, `go.mod` e `internal/model/nt.go` contendo os tipos `DocType` (`cte`, e constantes reservadas para `nfe`, `mdfe`, `nfce`), `NTIdentifier` (normalização de entradas como `2025.001`, `NT2025.001`, versão opcional) e `DocumentSpec` (metadados, lista de `Section`, `FieldElement` com `ID`, `Name`, `Kind` G/E/A/CE, `Parent`, `XMLPath`, `DataType`, `Occurrences`, `Length`, `Description`).
- **Mirror**: `.agents/skills/golang-patterns/SKILL.md:21-72` (tipos simples com zero-value seguro e funções puras de normalização).
- **Validate**: `go test ./internal/model/...`

### Task 2: Detector Automático do Tipo de Parser Fiscal no Diretório Atual
- **Action**: Implementar `internal/detector/detector.go` que analisa o diretório alvo (`--dir` ou `os.Getwd()`), detectando `DocTypeCTE` ao encontrar marcadores como `artifactId` `cte-parser` em `pom.xml`, pacotes `cteparser` ou schemas/fixtures CT-e, permitindo sobrescrita explícita via `--doc` e retornando erro orientativo caso o diretório seja desconhecido.
- **Mirror**: `cte-parser/pom.xml:6-9` (identificadores reais do projeto alvo).
- **Validate**: `go test -v ./internal/detector/...`

### Task 3: Cache Local e Resolvedor do Portal Oficial do CT-e
- **Action**: Implementar `internal/nt/cache` e `internal/nt/portal` (`provider.go` e `cte.go`) para consultar a página oficial de Notas Técnicas do Portal CT-e, localizar todas as entradas correspondentes ao código solicitado (ex.: `2025.001`), ordenar semanticamente revisões (`1.00`, `1.10`, `1.14`, `1.14a`) escolhendo a mais recente por padrão, baixar o arquivo PDF (extraindo de `.zip` em memória se empacotado pelo portal) e persistir no cache local. Incluir suporte direto a arquivo local via `--pdf` para operação offline/contingência.
- **Mirror**: `.agents/skills/golang-patterns/SKILL.md:74-113` (`http.Client` com timeout configurável, `context.Context` e erros encapsulados com `%w`).
- **Validate**: `go test -v ./internal/nt/cache/... ./internal/nt/portal/...`

### Task 4: Extrator Determinístico de PDF/Tabelas DF-e + Hook Híbrido LLM
- **Action**: Implementar `internal/nt/extractor/pdf.go` (execução segura de `pdftotext` sem shell) e `internal/nt/extractor/parser.go` para extrair título, versão, histórico resumido, seções e todas as linhas das tabelas de leiaute de 8 colunas (`# | Campo | Ele | Pai | Tipo | Ocor. | Tam. | Descrição`), montando automaticamente a árvore `Pai -> Campo` (ex.: `imp/IBSCBS/gIBSCBS/vIBS`). Definir `SemanticEnricher` em `enricher.go` para suportar enriquecimento opcional via LLM sem acoplar o modo determinístico a chaves de API.
- **Mirror**: Estrutura de caminhos hierárquicos de `cte-parser/src/main/java/br/gov/mg/fazenda/cteparser/mapper/IBSCBSMapper.java:22-89`.
- **Validate**: `go test -v ./internal/nt/extractor/...`

### Task 5: Interface CLI Cobra (`nt-inspect <codigo-nt>`) e Integração do Milestone 1
- **Action**: Implementar `cmd/nt-inspect/main.go` e `internal/cli/root.go` usando `spf13/cobra`, orquestrando: detecção do `DocType` no diretório atual -> resolução/cache da NT (ou leitura de `--pdf`) -> extração estruturada -> renderização clara no terminal (metadados da NT, origem cache/portal, seções identificadas e grupos/campos extraídos) ou JSON (`--json`).
- **Mirror**: `.agents/skills/golang-patterns/SKILL.md` (tratamento explícito de erros, saída limpa e testabilidade via `cobra.Command.SetOut`).
- **Validate**: `go test -race ./... && go build -o bin/nt-inspect ./cmd/nt-inspect`

## Validation
```bash
go vet ./...
go test -race -cover ./...
go build -o bin/nt-inspect ./cmd/nt-inspect
./bin/nt-inspect --help
```

## Risks
| Risk | Likelihood | Mitigation |
|---|---|---|
| Mudança de HTML, CAPTCHA ou bloqueio temporário no Portal Nacional do CT-e | Medium | Usar `User-Agent` explícito, cache persistente após o primeiro download e flag `--pdf <arquivo>` como contingência imediata sem bloquear a auditoria (respeitando estritamente a diretriz de nunca tentar burlar CAPTCHAs) |
| Quebras de página e colunas desalinhadas em PDFs da SEFAZ durante a extração de texto | High | Projetar o parser de tabelas tolerante a cabeçalhos repetidos de página (`Página X / Y`) e descrições multilinha, validado com fixtures reais de tabelas da NT 2025.001 |
| Notas Técnicas distribuídas em pacotes `.zip` em vez de `.pdf` direto no portal | Medium | Detectar content-type/extensão `.zip` no downloader e descompactar em memória via `archive/zip` da biblioteca padrão do Go |
| Ausência do binário `pdftotext` em outra máquina onde o CLI for copiado | Low | Verificar `exec.LookPath("pdftotext")` com mensagem de instrução clara (`poppler-utils`) e isolar a extração atrás da interface `TextExtractor` |

## Acceptance
- [x] All tasks complete
- [x] Validation passes
- [x] Patterns mirrored, not reinvented

