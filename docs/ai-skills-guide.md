# Guia de Skills e Steering — Kiro

> Este guia documenta as skills e steering files disponíveis para o Kiro neste projeto.
> Para um guia detalhado de como usar, verificar resultados e troubleshooting, veja [kiro-skills-guide.md](kiro-skills-guide.md).

---

## Steering Files (Regras Automáticas)

Steering files ficam em `.kiro/steering/` e são carregados automaticamente pelo Kiro.

| Arquivo | Tipo | Quando é carregado | O que faz |
|---|---|---|---|
| `project-base.md` | always | Toda interação | Princípios, stack, mapa de diagramas, arquitetura |
| `api-standards.md` | fileMatch | `handler*.go`, `openapi.yaml` | Padrões de API REST, HTTP status, envelopes |
| `go-standards.md` | fileMatch | `*.go` | Padrões Go, libs, error handling, DI, testes |
| `database-standards.md` | fileMatch | `redis/`, `cache.go`, `state_store.go` | Padrões Redis, namespacing, TTL, Pub/Sub |

---

## Skills (Sob Demanda)

Skills ficam em `.kiro/skills/<nome>/SKILL.md` e são acionadas quando a tarefa se encaixa.

| Skill | O que faz | Quando usar |
|---|---|---|
| `implement-usecase` | Adiciona um novo use case seguindo as camadas (domain → handler) | "Implementa o use case X no auth-service" |
| `implement-event` | Implementa publicação ou consumo de eventos Redis Pub/Sub | "Adiciona publicação do evento user.created" |
| `create-microservice` | Scaffolda um novo microsserviço Go do zero | "Cria um novo serviço chamado X" |

---

## Steering vs Skills

| Aspecto | Steering | Skills |
|---|---|---|
| Local | `.kiro/steering/` | `.kiro/skills/<nome>/SKILL.md` |
| Acionamento | Automático (always ou fileMatch) | Quando a tarefa se encaixa |
| Propósito | Regras e padrões gerais | Passo-a-passo para tarefas específicas |

```text
┌─────────────────────────────────────────────────────────┐
│                       Kiro                              │
│                                                         │
│  ┌─────────────────────┐  ┌──────────────────────────┐  │
│  │ Steering (automático)│  │  Skills (sob demanda)    │  │
│  │                     │  │                          │  │
│  │  • project-base     │  │  • implement-usecase     │  │
│  │  • go-standards     │  │  • implement-event       │  │
│  │  • api-standards    │  │  • create-microservice   │  │
│  │  • database-standards│ │                          │  │
│  └─────────────────────┘  └──────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## Como escrever prompts eficazes

| Critério | Descrição | Exemplo |
|---|---|---|
| Nomear a ação | Comece com um verbo claro | "Crie um novo microsserviço chamado `order-service`" |
| Especificar o contexto | Indique o serviço ou camada | "No `auth-service`, na camada de application..." |
| Definir o escopo | O que está incluído e o que não | "Apenas o use case e o handler, sem migration" |
| Informar dependências | Bancos, eventos ou serviços | "Publica evento `order.created`" |
| Um pedido por vez | Evite combinar múltiplas tarefas | "Implemente o use case CreateOrder" |

---

> Voltar ao índice: [README](../README.md)
