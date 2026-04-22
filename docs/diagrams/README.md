# Índice de Diagramas

Este diretório contém todos os diagramas de fluxo e sequência do projeto, organizados por contexto. Cada arquivo é autossuficiente e pode ser referenciado individualmente por agentes, skills e documentação técnica.

---

## Arquitetura

| Arquivo | Pattern | Descrição |
|---|---|---|
| [hexagonal-architecture-overview.md](hexagonal-architecture-overview.md) | Hexagonal Architecture | Visão macro do sistema, regra de dependência hexagonal e fluxo ponta a ponta de um request |

## Autenticação e Autorização

| Arquivo | Pattern | Descrição |
|---|---|---|
| [auth-pkce-flow.md](auth-pkce-flow.md) | PKCE (Authorization Code + PKCE) | Fluxo de autenticação para clientes web (browser) via `/authorize` e `/callback` |
| [auth-ropc-login-flow.md](auth-ropc-login-flow.md) | ROPC (Resource Owner Password Credentials) | Fluxo de login direto com username/password para clientes mobile e desktop |
| [auth-token-refresh-flow.md](auth-token-refresh-flow.md) | Token Refresh | Fluxo de renovação de token JWT (refresh token) |
| [auth-client-credentials-s2s.md](auth-client-credentials-s2s.md) | Client Credentials | Autenticação entre serviços (OAuth 2.0 Client Credentials) |

## Mensageria

| Arquivo | Pattern | Descrição |
|---|---|---|
| [pubsub-event-flow.md](pubsub-event-flow.md) | Pub/Sub | Fluxo de publicação e consumo de eventos via Redis Pub/Sub com idempotência |

## Resiliência

| Arquivo | Pattern | Descrição |
|---|---|---|
| [circuit-breaker-states.md](circuit-breaker-states.md) | Circuit Breaker | Estados e transições do Circuit Breaker |

---

> Referência principal: [TECHNICAL_BASE.md](../../TECHNICAL_BASE.md)
