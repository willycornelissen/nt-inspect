# nt-inspect — CLI de Inspeção de Conformidade Fiscal

## Problem
O desenvolvedor responsável por manter parsers de documentos fiscais enfrenta alta demanda para verificar se o código de cada parser está em conformidade com as Notas Técnicas (NTs) vigentes. Hoje, tanto a busca da documentação de compliance quanto a conferência dos campos e regras no código-fonte são feitas de forma 100% manual, consumindo horas de análise repetitiva e sujeita a falhas humanas a cada nova demanda.

## Evidence
- Relato direto do mantenedor/usuário: a localização da documentação da Nota Técnica e a verificação no código-fonte hoje só conseguem ser feitas manualmente sob alta demanda.
- Assumption — precisa de validação via cronometragem e uso do protótipo em inspeções reais (não há métricas históricas exatas de tempo gasto por auditoria ou taxa de retrabalho).

## Users
- **Primary**: Desenvolvedor/mantenedor de parsers de documentos fiscais que executa o CLI diretamente no diretório do parser para verificar aderência a uma Nota Técnica específica.
- **Not for**: Sistemas ou fluxos que esperam modificação/refatoração imediata e automática do código-fonte sem revisão humana, ou usuários finais validando arquivos fiscais emitidos em produção.

## Hypothesis
We believe **um aplicativo CLI que recebe o código da Nota Técnica como parâmetro, localiza suas exigências de compliance e inspeciona o código do parser no diretório atual** will **eliminar o gargalo da auditoria manual de conformidade** for **o desenvolvedor/mantenedor de parsers de documentos fiscais**.
We'll know we're right when **o usuário conseguir gerar um relatório confiável de gaps de conformidade em minutos em vez de horas**.

## Success Metrics
| Metric | Target | How measured |
|---|---|---|
| Tempo para gerar relatório de gaps por NT | Minutos em vez de horas (`TBD — meta numérica exata precisa de validação via teste prático do MVP`) | Tempo ponta a ponta entre a invocação do CLI com o código da NT e a entrega do relatório |
| Confiabilidade do relatório de gaps | `TBD — precisa de validação via comparação com auditoria manual em NTs conhecidas` | Proporção de divergências reais apontadas corretamente sem omissões críticas |

## Scope
**MVP** — Aplicativo CLI invocado no diretório do parser de documento fiscal recebendo o **código da Nota Técnica** como parâmetro, capaz de obter a especificação correspondente, inspecionar a conformidade do parser local frente à NT e emitir um relatório confiável dos gaps encontrados.

**Out of scope**
- Alterar imediatamente o código-fonte do parser — o foco inicial é exclusivamente inspecionar e relatar os gaps com confiabilidade, sem modificar o repositório alvo.
- Interface gráfica ou painel web — operação restrita à linha de comando no diretório do projeto inspecionado.

## Delivery Milestones
<!-- Business outcomes, not engineering tasks. /plan turns each into a plan. -->
<!-- Status: pending | in-progress | complete -->

| # | Milestone | Outcome | Status | Plan |
|---|---|---|---|---|
| 1 | Obtenção da NT pelo código | Ao informar o código da Nota Técnica no CLI, o usuário obtém automaticamente o conteúdo/exigências da NT sem precisar buscar a documentação manualmente | complete | `.claude/plans/nt-inspect.plan.md` |
| 2 | Inspeção de conformidade do parser local | Ao executar o CLI na pasta de um parser fiscal com o código da NT, o sistema cruza as exigências da NT com o parser presente no diretório atual | complete | `.claude/plans/nt-inspect-m2.plan.md` |
| 3 | Relatório confiável de gaps | O usuário recebe em minutos um relatório claro indicando o que está aderente e quais gaps de conformidade existem em relação à NT informada | complete | `.claude/plans/nt-inspect-m3.plan.md` |

## Open Questions
- [ ] Quais tipos de documentos fiscais (ex.: CT-e, NF-e, MDF-e, NFC-e) devem ser cobertos prioritariamente na resolução automática pelo código da NT (`TBD — precisa de validação via definição na etapa /plan`)?
- [ ] Como desambiguar quando o mesmo código numérico de NT existir em portais de documentos fiscais distintos (ex.: inferir pelo parser do diretório atual ou receber flag complementar)?
- [ ] Qual formato de visualização do relatório de gaps melhor atende ao fluxo diário (saída direta no terminal, arquivo salvo, ou ambos)?
- [ ] Quando uma mesma NT tiver múltiplas versões publicadas (ex.: v1.00, v1.10), o CLI deve sempre considerar a versão mais recente disponível?

## Risks
| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Instabilidade, bloqueio ou mudança de estrutura nos portais oficiais ao buscar a NT pelo código | Média | Alto | Prever mensagens claras de falha na obtenção da documentação e opção futura de cache/fonte local caso o portal esteja indisponível |
| Interpretação imprecisa de regras e tabelas complexas das NTs gerando falsos positivos ou omissões no relatório | Alta | Alto | Validar a saída do MVP comparando o relatório gerado contra uma auditoria manual de referência antes de escalar o uso |
| Ambiguidade de escopo sobre linguagens/estruturas de parsers suportadas | Média | Médio | Delimitar claramente o perfil de parser alvo do primeiro marco durante o `/plan` |

---
*Status: DRAFT — requirements only. Implementation planning pending via /plan.*

