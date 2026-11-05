# Plan: Milestone 2 — Inspeção de Conformidade do Parser Local

**Source PRD**: `.claude/prds/nt-inspect.prd.md`
**Selected Milestone**: `2 — Inspeção de conformidade do parser local`
**Complexity**: Medium

## Summary
Implementar no `nt-inspect` o motor de inspeção estática de código-fonte capaz de analisar o repositório do parser fiscal (a começar pela estrutura Java de `cte-parser/`: classes de modelo, anotações `@JsonProperty`, mapeamentos `FIELD_SETTERS` de tags XML e classes de tabela/builder) e cruzar essa base de código contra o `DocumentSpec` da Nota Técnica (gerado no Milestone 1). O motor identifica campos compliant (mapeados corretamente), campos faltantes (novos campos exigidos pela NT ausentes no código) e divergências de nomenclatura ou grupos obsoletos mantidos no parser.

## Patterns to Mirror
| Category | Source | Pattern |
|---|---|---|
| Domain / Parser Model | `cte-parser/src/main/java/br/gov/mg/fazenda/cteparser/mapper/IBSCBSMapper.java:22-89` e `cte-parser/src/main/java/br/gov/mg/fazenda/cteparser/mapper/CteMapper.java:82-86` | Identificadores de caminho XML como strings estáticas (`"CTEPROC.CTE.INFCTE.IMP.IBSCBS.GIBSCBS.VIBS"` e `"IDE.GCOMPRAGOV.TPENTEGOV"`) e setters de campos. |
| Model Classes | `cte-parser/src/main/java/br/gov/mg/fazenda/cteparser/cte/IbsCbs.java:9-217` e `cte-parser/src/main/java/br/gov/mg/fazenda/cteparser/cte/Cte.java:680-695` | Declarações de atributos Java com anotações `@JsonProperty("...")` e tipos (`BigDecimal`, `Integer`, `String`, `List<String>`). |
| Go Patterns | `.agents/skills/golang-patterns/SKILL.md:21-120` | Structs concretas para resultados de inspeção (`InspectionReport`, `FieldMatch`, `ComplianceStatus`), interfaces para scanners de linguagem (`CodeScanner`), wrapping de erros idiomático e table-driven tests com fixtures. |

## Files to Change
| File | Action | Why |
|---|---|---|
| `internal/model/compliance.go` | CREATE | Estruturas de dados de conformidade (`ComplianceStatus`: COMPLIANT / NON_COMPLIANT, `FieldFindingKind`: MATCH, MISSING, OBSOLETE, NAME_MISMATCH, `FieldFinding`, `InspectionReport`). |
| `internal/scanner/scanner.go` | CREATE | Interface `CodeScanner` e orquestrador de scanning de código-fonte do parser. |
| `internal/scanner/java_scanner.go` | CREATE | Scanner especializado em Java para extrair mapeamentos de tags XML (`SETTERS.put("...", ...)`), `@JsonProperty` e membros de classe no projeto alvo. |
| `internal/scanner/java_scanner_test.go` | CREATE | Testes unitários com trechos reais de `IBSCBSMapper.java`, `CteMapper.java` e `IbsCbs.java`. |
| `internal/inspector/inspector.go` | CREATE | Motor de comparação que cruza os campos do `DocumentSpec` da NT com os símbolos/mapeamentos extraídos do código pelo scanner. |
| `internal/inspector/inspector_test.go` | CREATE | Testes unitários comparando `DocumentSpec` simulado com findings esperados (match, missing, obsolete). |
| `internal/cli/root.go` | UPDATE | Conectar o pipeline de inspeção no comando CLI: após obter a NT, invocar o scanner no diretório alvo, gerar o `InspectionReport` e retornar o exit code apropriado (0 para compliant, 1 para non-compliant). |
| `.claude/prds/nt-inspect.prd.md` | UPDATE | Marcar o Milestone 2 como `in-progress` e vincular o plano. |

## Tasks

### Task 1: Modelos de Dados de Conformidade e Status
- **Action**: Criar `internal/model/compliance.go` definindo:
  - `ComplianceStatus`: enum `StatusCompliant`, `StatusNonCompliant`, `StatusError`
  - `FieldFindingKind`: enum `FindingMatch`, `FindingMissing`, `FindingObsolete`, `FindingNameMismatch`
  - `FieldFinding`: detalhe de cada campo (nome da tag, path XML esperado na NT, path ou propriedade encontrada no código, arquivo/linha de referência, mensagem explicativa, severidade)
  - `InspectionReport`: resumo quantitativo (`TotalRequired`, `MatchedCount`, `MissingCount`, `ObsoleteCount`, `ComplianceRate`), lista de findings e metadados da NT/repositório.
- **Mirror**: Padrões de modelo do Milestone 1 em `internal/model/nt.go`.
- **Validate**: `go test ./internal/model/...`

### Task 2: Scanner de Código-Fonte Java (`internal/scanner`)
- **Action**: Implementar `internal/scanner/java_scanner.go` que percorre a árvore de código do repositório alvo (ex.: `cte-parser/src/main/java`) e extrai:
  - Tags XML literais mapeadas em dicionários de setters (ex.: `"CTEPROC.CTE.INFCTE.IMP.IBSCBS.GIBSCBS.VIBS"` e `"VTAR"`, `"CLASSDUTO"`).
  - Anotações `@JsonProperty("...")` nas classes de modelo.
  - Nomes de campos e métodos getters/setters em classes de modelo (`IbsCbs`, `Cte`, `ModalDutoviario`).
  - Associação entre tag XML encontrada e arquivo/linha no código-fonte.
- **Mirror**: Convenções de navegação do `internal/detector/detector.go` e caminhos de `cte-parser/src/main/java/...`.
- **Validate**: `go test -v ./internal/scanner/...`

### Task 3: Motor de Cruzamento e Auditoria de Conformidade (`internal/inspector`)
- **Action**: Implementar `internal/inspector/inspector.go` que recebe o `DocumentSpec` da NT e os artefatos escaneados do repositório:
  - Para cada `FieldElement` da NT:
    - Verifica se o caminho XML completo ou relativo (ex.: `IMP.IBSCBS.GIBSCBS.VIBS` ou sufixo `GIBSCBS.VIBS`) está mapeado em setters/handlers XML.
    - Se não encontrado pelo caminho XML, verifica se o nome da tag existe via `@JsonProperty` ou atributo em classes de modelo correlatas.
    - Se não existir, registra finding `FindingMissing` (campo novo da NT que o parser ainda não implementa).
  - Para os campos do repositório que pertencem aos grupos afetados pela NT:
    - Se a NT suprimiu/revogou o grupo ou campo (ex.: `gIBSCredPres` na v1.14), registra finding `FindingObsolete`.
    - Se o nome foi alterado (ex.: `tpCompraGov` -> `tpEnteGov`), registra finding `FindingNameMismatch`.
  - Calcula a taxa de conformidade e o veredito final (`StatusCompliant` se 0 missing/mismatch críticos, senão `StatusNonCompliant`).
- **Mirror**: Casos de auditoria descritos no histórico do `cte-parser` (`nt/PLANO-CONFORMIDADE-NT-2025.001.md`).
- **Validate**: `go test -v ./internal/inspector/...`

### Task 4: Integração CLI e Veredito Ponta a Ponta
- **Action**: Atualizar `internal/cli/root.go` para encadear a etapa de inspeção logo após a obtenção do `DocumentSpec`. Exibir a análise de conformidade no terminal ou em JSON:
  - Definir exit code `0` quando compliant e `1` quando gaps forem detectados.
  - Permitir a inspeção local com `--pdf` ou via download automático.
- **Mirror**: Formatos de saída e códigos de retorno combinados na resposta anterior.
- **Validate**: Executar `nt-inspect 2025.001 --dir cte-parser` e testar saídas em modo humano e `--json`.

## Validation
```bash
go vet ./...
go test -race -cover ./...
go build -o bin/nt-inspect ./cmd/nt-inspect
./bin/nt-inspect 2025.001 --dir cte-parser
```

## Risks
| Risk | Likelihood | Mitigation |
|---|---|---|
| Variação no prefixo do caminho XML no código (`CTEPROC.CTE.INFCTE...` vs `INFCTE...` vs tag simples `VIBS`) | Medium | Normalizar os caminhos XML comparando sufixos hierárquicos (ex.: `hasSuffix` ou casamento de árvore `Parent/Tag`) para não gerar falsos positivos |
| Identificação de campos obsoletos em grupos não tocados pela NT | Low | Restringir a checagem de obsolescência estritamente aos grupos que a NT alterou (ex.: `IBSCBS`, `gCompraGov`, `Duto`) |
| Código Java multilinha ou quebras em `@JsonProperty` | Low | Usar regex e scanner de tokens com suporte a whitespace flexível nos fontes Java |

## Acceptance
- [x] All tasks complete
- [x] Validation passes
- [x] Patterns mirrored, not reinvented

