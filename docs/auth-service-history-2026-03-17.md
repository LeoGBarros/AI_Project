# Histórico do `auth-service` - 2026-03-17

Este documento registra o que foi construído e consolidado hoje no `auth-service`, além das decisões técnicas que orientaram o trabalho e os motivos de cada escolha.

## Resumo do que foi entregue

- Estrutura hexagonal do `auth-service` com `domain`, `application`, `ports` e `adapters`.
- Fluxos de autenticação para `login`, `authorize`, `callback`, `refresh` e `logout`.
- Integração com `Keycloak` para emissão, renovação e revogação de tokens.
- Persistência do estado `PKCE` em `Redis` para suportar o fluxo web com `state` e `code_verifier`.
- Inicialização de `OpenTelemetry`, `logger` estruturado e middlewares de correlação/rastreabilidade.
- Endpoints de health check em `/healthz/live` e `/healthz/ready`.
- Contrato OpenAPI documentando os fluxos suportados pelo serviço.
- Ambiente local documentado com `docker compose`, `Keycloak` e `Redis`.

## Decisões tomadas e motivos

### 1. `main.go` como ponto único de wiring

Decisão:

- Manter a montagem das dependências em `cmd/server/main.go`.

Motivo:

- Isso preserva a arquitetura hexagonal e evita acoplamento entre camadas.
- Facilita entender o grafo de dependências do serviço em um único lugar.
- Permite trocar adapters sem alterar a lógica de negócio.

### 2. `Keycloak` como fonte de identidade

Decisão:

- Usar `Keycloak` para login, refresh, logout e autorização baseada em tokens.

Motivo:

- Centraliza autenticação e reduz duplicação de lógica de IAM no serviço.
- Mantém o `auth-service` como orquestrador de fluxo, não como provedor de identidade.
- Alinha o serviço com a base técnica do projeto para autenticação centralizada.

### 3. `ROPC` restrito a `mobile` e `app`

Decisão:

- Permitir `ROPC` apenas para clientes `mobile` e `app`.
- Exigir `PKCE` para o cliente `web`.

Motivo:

- `ROPC` não é o fluxo mais apropriado para navegador.
- `PKCE` oferece proteção melhor para aplicações web públicas.
- A divisão reduz risco de exposição de credenciais em clientes browser-based.

### 4. Estado `PKCE` armazenado em `Redis`

Decisão:

- Armazenar `state`, `code_verifier`, `client_id` e `redirect_uri` em `Redis`.

Motivo:

- O fluxo de callback precisa validar o `state` com segurança.
- O armazenamento temporário em `Redis` permite expiração automática e simples limpeza de estado consumido.
- Evita persistir dados transitórios em banco relacional sem necessidade.

### 5. `OpenTelemetry` e logs estruturados desde o startup

Decisão:

- Inicializar tracing e logger no bootstrap do serviço.

Motivo:

- Observabilidade é requisito da base técnica, não uma melhoria posterior.
- Permite rastrear falhas no login, callback e refresh desde o primeiro request.
- Garante correlação entre request, span e logs usando `trace_id` e `correlation_id`.

### 6. Health checks fora do roteamento autenticado

Decisão:

- Expor `/healthz/live` e `/healthz/ready` diretamente no serviço.

Motivo:

- Liveness precisa responder sem dependência externa.
- Readiness deve refletir a disponibilidade do `Redis`, que é essencial para o fluxo `PKCE`.
- Esses endpoints não devem depender de autenticação nem do Kong.

### 7. API-first com `OpenAPI`

Decisão:

- Manter `auth-service/api/openapi.yaml` como contrato do serviço.

Motivo:

- O contrato documenta os fluxos suportados e reduz ambiguidade para integração.
- Ajuda a alinhar implementação, documentação e validação de comportamento.

## Soluções aplicadas

- Estrutura de projeto separada por responsabilidade.
- Middleware para `correlation_id`, tracing e logging.
- Handler HTTP dedicado para mapear erros de domínio em respostas HTTP coerentes.
- Use cases separados para cada operação de autenticação.
- Modelagem explícita de tipos de entrada e saída por caso de uso.
- Documentação local para subir `Redis` e `Keycloak` com `docker compose`.

## Observações relevantes para manutenção

- O fluxo web depende do armazenamento do estado `PKCE` e da limpeza do estado consumido.
- O `refresh_token` deve ser tratado como credencial sensível e nunca logado.
- O contrato OpenAPI precisa permanecer sincronizado com os handlers e com os casos de uso.
- O guia local deve ser atualizado sempre que o layout de portas, variáveis ou dependências mudar.

## Arquivos de referência

- [`auth-service/api/openapi.yaml`](../auth-service/api/openapi.yaml)
- [`auth-service/cmd/server/main.go`](../auth-service/cmd/server/main.go)
- [`auth-service/internal/adapters/http/handler.go`](../auth-service/internal/adapters/http/handler.go)
- [`auth-service/internal/application/login_usecase.go`](../auth-service/internal/application/login_usecase.go)
- [`auth-service/internal/application/callback_usecase.go`](../auth-service/internal/application/callback_usecase.go)
- [`auth-service/internal/application/refresh_usecase.go`](../auth-service/internal/application/refresh_usecase.go)
- [`auth-service/internal/application/logout_usecase.go`](../auth-service/internal/application/logout_usecase.go)
- [`docs/auth-service-local-dev.md`](auth-service-local-dev.md)
