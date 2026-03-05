# Índice de Diagramas

Este diretório contém todos os diagramas de fluxo e sequência do projeto, organizados por contexto. Cada arquivo é autossuficiente e pode ser referenciado individualmente por agentes, skills e documentação técnica.

---

## Arquitetura

| Arquivo | Descrição |
|---|---|
| [architecture-overview.md](architecture-overview.md) | Visão macro do sistema, regra de dependência hexagonal e fluxo ponta a ponta de um request |

## Autenticação e Autorização

| Arquivo | Descrição |
|---|---|
| [auth-login-flow.md](auth-login-flow.md) | Fluxo de login do usuário e request autenticado via Kong + Keycloak |
| [auth-token-refresh.md](auth-token-refresh.md) | Fluxo de renovação de token JWT (refresh token) |
| [auth-service-to-service.md](auth-service-to-service.md) | Autenticação entre serviços (OAuth 2.0 Client Credentials) |

## Mensageria

| Arquivo | Descrição |
|---|---|
| [pubsub-event-flow.md](pubsub-event-flow.md) | Fluxo de publicação e consumo de eventos via Redis Pub/Sub com idempotência |

## Resiliência

| Arquivo | Descrição |
|---|---|
| [circuit-breaker-states.md](circuit-breaker-states.md) | Estados e transições do Circuit Breaker |

---

> Referência principal: [TECHNICAL_BASE.md](../../TECHNICAL_BASE.md)
