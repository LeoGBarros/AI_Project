# Base do Projeto

## Fonte da Verdade

O arquivo [`TECHNICAL_BASE.md`](../../TECHNICAL_BASE.md) é a única fonte de verdade técnica deste projeto. Antes de implementar qualquer código, leia as seções relevantes deste documento. Qualquer desvio dos padrões definidos deve ser explicitamente justificado.

## Princípios Inegociáveis

| Princípio | Implicação prática |
|---|---|
| Separação de responsabilidades | Cada serviço tem domínio único; não acesse internals de outros serviços |
| Design para falha | Sempre implemente circuit breaker, retry com backoff e timeout em chamadas externas |
| Observabilidade desde o início | Logs JSON estruturados, spans OpenTelemetry e métricas são obrigatórios |
| Segurança por padrão | Todo endpoint protegido com JWT; nunca expor stack traces em produção |
| API-first | O contrato `api/openapi.yaml` é definido antes da implementação |
| Imutabilidade de infra | Toda mudança de schema via migration versionada; nunca alterar migrations já aplicadas |

## Mapa de Diagramas

| Diagrama | Quando usar |
|---|---|
| [`docs/diagrams/architecture-overview.md`](../../docs/diagrams/architecture-overview.md) | Topologia do sistema e estrutura dos microsserviços |
| [`docs/diagrams/auth-pkce-flow.md`](../../docs/diagrams/auth-pkce-flow.md) | Fluxo PKCE para clientes web |
| [`docs/diagrams/auth-ropc-login-flow.md`](../../docs/diagrams/auth-ropc-login-flow.md) | Login ROPC para mobile/desktop |
| [`docs/diagrams/auth-token-refresh-flow.md`](../../docs/diagrams/auth-token-refresh-flow.md) | Renovação de access token via refresh token |
| [`docs/diagrams/auth-client-credentials-s2s.md`](../../docs/diagrams/auth-client-credentials-s2s.md) | Chamadas autenticadas entre microsserviços |
| [`docs/diagrams/pubsub-event-flow.md`](../../docs/diagrams/pubsub-event-flow.md) | Publicação ou consumo de eventos Redis Pub/Sub |
| [`docs/diagrams/circuit-breaker-states.md`](../../docs/diagrams/circuit-breaker-states.md) | Resiliência em chamadas a serviços externos |

## Stack e Tecnologias

| Camada | Tecnologia |
|---|---|
| API Gateway | Kong 3.x |
| IAM / Identidade | Keycloak 24.x |
| Linguagem | Go 1.22+ |
| Cache / Fila | Redis 7.x |
| Observabilidade | OpenTelemetry (SDK Go 1.x) |
| Containerização | Docker 24.x |

## Arquitetura de Cada Serviço

Todo microsserviço segue organização em camadas:

```
internal/
├── domain/          # Entidades, Value Objects, regras de negócio. SEM dependências externas.
├── application/     # Use cases. Importa apenas domain e ports/output.
├── ports/
│   ├── input/       # Interfaces de entrada (HTTP handlers, consumers)
│   └── output/      # Interfaces de saída (repositórios, publishers)
└── adapters/        # Implementações concretas (Keycloak, Redis, HTTP client)
```

Regra de dependência: `domain` ← `application` ← `adapters`. O `domain` nunca importa nada externo.
