# Criar Novo Microsserviço

Scaffoldar um novo microsserviço Go do zero, seguindo a arquitetura em camadas e padrões do projeto.

> Nota: Atualmente só existe o `auth-service`. Esta skill é para quando um novo serviço for necessário.

## Antes de Começar

- [`TECHNICAL_BASE.md` — seção 3.2](../../../TECHNICAL_BASE.md#32-padrão-arquitetural-por-serviço-hexagonal--ddd) — Estrutura e regra de dependência
- [`TECHNICAL_BASE.md` — seção 9](../../../TECHNICAL_BASE.md#9-padrões-de-codificação-go) — Padrões Go
- [`docs/diagrams/architecture-overview.md`](../../../docs/diagrams/architecture-overview.md) — Arquitetura do sistema

## Informações Necessárias

1. Nome do serviço (ex: `user-service`, `order-service`)
2. Banco de dados principal (PostgreSQL, MongoDB ou nenhum)
3. Publica eventos? (sim/não)
4. Consome eventos? (sim/não)

## Passo a Passo

### 1 — Criar árvore de diretórios

```
<service-name>/
├── cmd/server/main.go
├── internal/
│   ├── domain/
│   ├── application/
│   ├── ports/input/ e output/
│   └── adapters/http/ e redis/
├── pkg/middleware/ e apierror/
├── config/config.go
└── api/openapi.yaml
```

### 2 — Inicializar módulo Go

```bash
go mod init github.com/<org>/<service-name>
```

### 3 — Criar `config/config.go`

Variáveis de ambiente com viper. Validar obrigatórias no startup.

### 4 — Criar `cmd/server/main.go`

DI manual, chi router, health checks, graceful shutdown.

### 5 — Criar `Dockerfile`

Multi-stage: golang:1.22-alpine → distroless.

### 6 — Criar `api/openapi.yaml`

Contrato base com health check e schema de erro.

## Checklist Final

- [ ] Árvore de diretórios conforme padrão
- [ ] `go.mod` com libs obrigatórias
- [ ] `config.go` valida variáveis no startup
- [ ] `main.go` com DI manual e graceful shutdown
- [ ] `GET /healthz/live` e `GET /healthz/ready` implementados
- [ ] `Dockerfile` multi-stage
- [ ] `api/openapi.yaml` com schema base
