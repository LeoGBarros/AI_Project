# Guia de Skills e Steering — Kiro

Este guia explica como as skills e steering files do projeto funcionam, como usar e como verificar se estão sendo aplicadas.

---

## O que são Steering Files?

Steering files são arquivos Markdown em `.kiro/steering/` que fornecem contexto e regras ao Kiro automaticamente. Funcionam como "instruções de fundo" que guiam o agente sem precisar repetir regras a cada conversa.

### Steering Files do Projeto

| Arquivo | Tipo | Quando é carregado | O que faz |
|---|---|---|---|
| `project-base.md` | always | Toda interação | Define princípios, stack, mapa de diagramas e arquitetura |
| `api-standards.md` | fileMatch | Ao editar `handler*.go` ou `openapi.yaml` | Padrões de API REST, HTTP status, envelopes, headers |
| `go-standards.md` | fileMatch | Ao editar qualquer `.go` | Padrões Go, libs, error handling, DI, testes |
| `database-standards.md` | fileMatch | Ao editar arquivos em `redis/`, `cache.go`, `state_store.go` | Padrões Redis, namespacing, TTL, Pub/Sub |

### Como saber se estão sendo usadas?

- `project-base.md` (always): aparece em toda interação. Você verá o Kiro seguindo os princípios e referenciando diagramas automaticamente.
- Os demais (fileMatch): são carregados quando você abre ou edita um arquivo que bate com o padrão. Por exemplo, ao editar `handler.go`, o `api-standards.md` e `go-standards.md` são injetados no contexto.

### Como verificar?

Observe se o Kiro:
- Segue convenções de URL ao sugerir endpoints (api-standards)
- Usa `errors.Is`/`errors.As` ao lidar com erros (go-standards)
- Aplica TTL em chaves Redis (database-standards)
- Referencia diagramas ao explicar fluxos (project-base)

Se o Kiro não seguir alguma regra, pode ser que o arquivo não foi carregado (verifique o `fileMatchPattern`) ou que a instrução precisa ser mais explícita.

---

## O que são Skills?

Skills são instruções passo-a-passo em `.kiro/skills/<nome>/SKILL.md` que o Kiro usa quando você pede para executar uma tarefa específica. Diferente do steering (que é passivo), skills são acionadas ativamente quando o contexto da conversa bate com a descrição.

### Skills do Projeto

| Skill | O que faz | Quando usar |
|---|---|---|
| `implement-usecase` | Guia para adicionar um novo use case seguindo as camadas | "Implementa o use case X no auth-service" |
| `implement-event` | Guia para publicar/consumir eventos Redis Pub/Sub | "Adiciona publicação do evento user.created" |
| `create-microservice` | Scaffold de um novo microsserviço do zero | "Cria um novo serviço chamado X" |

### Como usar uma skill?

Basta pedir ao Kiro algo que se encaixe na descrição da skill. Exemplos:

- "Implementa um novo use case de criação de pedido no order-service" → aciona `implement-usecase`
- "Adiciona consumo do evento order.created no notification-service" → aciona `implement-event`
- "Cria um novo microsserviço chamado payment-service" → aciona `create-microservice`

Você também pode referenciar a skill diretamente no chat usando `#`.

### Como ver o resultado?

Após o Kiro executar uma skill, verifique:

1. Os arquivos foram criados nas camadas corretas? (domain → application → ports → adapters)
2. O checklist final da skill foi atendido? (cada skill tem um checklist no final)
3. Os testes passam? (`go test ./...`)
4. O código segue os padrões do steering? (error handling, DI, etc.)

### Como saber se a skill certa foi usada?

Observe o comportamento do Kiro:
- Se seguiu a ordem de camadas (domain primeiro, handler por último) → `implement-usecase` foi aplicada
- Se criou envelope de evento com `event_id`, `correlation_id` → `implement-event` foi aplicada
- Se criou a árvore completa com health checks e Dockerfile → `create-microservice` foi aplicada

Se o Kiro não seguir o passo-a-passo da skill, você pode pedir explicitamente: "Siga a skill implement-usecase para isso".

---

## Steering vs Skills — Resumo

| Aspecto | Steering | Skills |
|---|---|---|
| Local | `.kiro/steering/` | `.kiro/skills/<nome>/SKILL.md` |
| Acionamento | Automático (always ou fileMatch) | Quando a tarefa se encaixa na descrição |
| Propósito | Regras e padrões gerais | Passo-a-passo para tarefas específicas |
| Exemplo | "Use errors.Is para erros de domínio" | "Para criar um use case, siga estes 7 passos" |

---

## Origem

Estes arquivos foram migrados do Cursor (`.cursor/rules/` e `.cursor/skills/`) para o formato Kiro. A skill `create-skill` (meta-skill para criar novas skills) não foi migrada por ser específica do Cursor.
