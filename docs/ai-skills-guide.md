# Guia de Skills de IA

> Contexto: Este guia documenta as skills e rules disponíveis para agentes de IA (Cursor) neste projeto.

---

## O que são Skills

Skills são instruções estruturadas que ensinam o agente de IA a executar tarefas recorrentes de desenvolvimento seguindo os padrões do projeto. Cada skill é um arquivo `SKILL.md` armazenado em `.cursor/skills/<nome-da-skill>/`.

```text
.cursor/skills/
├── create-microservice/
│   └── SKILL.md          ← Scaffoldar novo microsserviço
├── create-skill/
│   └── SKILL.md          ← Criar nova skill
├── implement-event/
│   └── SKILL.md          ← Implementar evento Pub/Sub
└── implement-usecase/
    └── SKILL.md          ← Adicionar use case em serviço existente
```

| Aspecto | Detalhe |
|---|---|
| Formato | Markdown com frontmatter YAML (`name` + `description`) |
| Idioma | Português brasileiro |
| Acionamento | O agente de IA seleciona a skill automaticamente com base no pedido do usuário |
| Estrutura obrigatória | Antes de Começar → Informações Necessárias → Passo a Passo → Checklist Final |
| Limite | Máximo 500 linhas por `SKILL.md` (usar progressive disclosure se necessário) |

---

## Skills Disponíveis

| Skill | O que faz | Quando usar |
|---|---|---|
| `create-microservice` | Scaffolda um novo microsserviço Go do zero com arquitetura hexagonal, DI manual, health checks, Dockerfile e OpenAPI base | Quando precisar criar um novo serviço ou repositório de serviço |
| `create-skill` | Cria uma nova skill seguindo os padrões de estrutura, idioma e convenções do projeto | Quando precisar automatizar um workflow recorrente ou documentar um processo repetitivo como skill |
| `implement-event` | Implementa publicação ou consumo de eventos via Redis Pub/Sub com envelope padrão e idempotência | Quando precisar publicar um evento de domínio ou criar um consumer de eventos |
| `implement-usecase` | Adiciona um novo use case em um microsserviço existente, seguindo todas as camadas hexagonais (domain → application → ports → adapters → handler) | Quando precisar implementar uma nova funcionalidade, endpoint ou operação em um serviço já existente |

---

## Rules vs Skills

Rules e skills são complementares mas funcionam de formas diferentes:

| Tipo | Local | Acionamento | Descrição |
|---|---|---|---|
| Rule `project-base` | `.cursor/rules/project-base.mdc` | Automático (sempre ativo) | Regra base do projeto — referência técnica central (`TECHNICAL_BASE.md`), princípios inegociáveis, mapa de diagramas e tabela de skills |
| Rule `go-standards` | `.cursor/rules/go-standards.mdc` | Automático em arquivos `**/*.go` | Padrões de codificação Go — estrutura hexagonal, libs recomendadas, error handling, DI, testes e concorrência |
| Rule `api-standards` | `.cursor/rules/api-standards.mdc` | Automático em `**/handler*.go`, `**/openapi.yaml` | Padrões de API REST — convenções de URL, métodos HTTP, envelopes de resposta, headers e observabilidade |
| Rule `database-standards` | `.cursor/rules/database-standards.mdc` | Automático em `**/*.sql`, `**/migrations/**`, `**/repository.go`, `**/cache.go` | Padrões de banco de dados — nomenclatura PostgreSQL/MongoDB/Redis, migrations e connection pool |
| Skill `create-microservice` | `.cursor/skills/create-microservice/` | Sob demanda (pedido do usuário) | Passo a passo para scaffoldar um novo microsserviço completo |
| Skill `create-skill` | `.cursor/skills/create-skill/` | Sob demanda (pedido do usuário) | Passo a passo para criar uma nova skill seguindo os padrões |
| Skill `implement-event` | `.cursor/skills/implement-event/` | Sob demanda (pedido do usuário) | Passo a passo para implementar publicação ou consumo de eventos Redis Pub/Sub |
| Skill `implement-usecase` | `.cursor/skills/implement-usecase/` | Sob demanda (pedido do usuário) | Passo a passo para adicionar um novo use case em serviço existente |

```text
┌─────────────────────────────────────────────────────────────┐
│                    Agente de IA (Cursor)                    │
│                                                             │
│  ┌───────────────────────┐  ┌────────────────────────────┐  │
│  │   Rules (automáticas) │  │   Skills (sob demanda)     │  │
│  │                       │  │                            │  │
│  │  • project-base       │  │  • create-microservice     │  │
│  │  • go-standards       │  │  • create-skill            │  │
│  │  • api-standards      │  │  • implement-event         │  │
│  │  • database-standards │  │  • implement-usecase       │  │
│  │                       │  │                            │  │
│  │  Aplicadas sempre ou  │  │  Acionadas quando o        │  │
│  │  por tipo de arquivo  │  │  usuário pede uma tarefa   │  │
│  └───────────────────────┘  └────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## Como escrever prompts eficazes

Ao pedir para o agente de IA executar uma tarefa, siga este checklist para obter melhores resultados:

| Critério | Descrição | Exemplo |
|---|---|---|
| Nomear a ação | Comece com um verbo claro que descreva o que você quer | "Crie um novo microsserviço chamado `order-service`" |
| Especificar o contexto | Indique o serviço, camada ou arquivo envolvido | "No `user-service`, na camada de application..." |
| Definir o escopo | Deixe claro o que está incluído e o que não está | "Apenas o use case e o handler, sem migration por enquanto" |
| Informar dependências | Mencione bancos de dados, eventos ou serviços relacionados | "Usa PostgreSQL e publica evento `order.created`" |
| Referenciar padrões | Aponte para seções do TECHNICAL_BASE ou diagramas relevantes | "Seguindo o fluxo do diagrama `pubsub-event-flow.md`" |
| Dar exemplos concretos | Forneça nomes de campos, endpoints ou payloads esperados | "Endpoint `POST /v1/orders` com campos `product_id` e `quantity`" |
| Um pedido por vez | Evite combinar múltiplas tarefas em um único prompt | "Implemente o use case CreateOrder" (não "crie o serviço, use case e deploy") |

---

> Voltar ao índice: [README](../README.md)
