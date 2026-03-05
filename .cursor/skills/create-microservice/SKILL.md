---
name: create-microservice
description: Scaffoldar um novo microsserviço Go do zero, seguindo a arquitetura hexagonal e todos os padrões definidos no TECHNICAL_BASE. Use quando o usuário pedir para criar um novo serviço ou um novo repositório de serviço.
---

# Criar Novo Microsserviço

## Antes de Começar

Leia as seguintes seções do [`TECHNICAL_BASE.md`](../../../TECHNICAL_BASE.md) antes de criar qualquer arquivo:
- **Seção 3.2** — Estrutura hexagonal e regra de dependência
- **Seção 9** — Padrões de codificação Go
- **Seção 8.4** — Health check obrigatório

Diagrama de arquitetura de referência: [`docs/diagrams/architecture-overview.md`](../../../docs/diagrams/architecture-overview.md)

## Informações Necessárias

Antes de scaffoldar, confirme com o usuário:
1. **Nome do serviço** (ex: `user-service`, `order-service`)
2. **Banco de dados principal** (PostgreSQL, MongoDB ou ambos)
3. **Publica eventos?** (sim/não — se sim, qual evento inicial)
4. **Consome eventos?** (sim/não — se sim, de qual canal)

---

## Passo a Passo

### 1. Criar a Árvore de Diretórios

```bash
SERVICE=<nome-do-servico>

mkdir -p $SERVICE/{cmd/server,internal/{domain,application,ports/{input,output},adapters/{http,postgres,mongo,redis}},pkg/{middleware,apierror},config,migrations,api}
```

Resultado esperado:
```
<service-name>/
├── cmd/server/
├── internal/
│   ├── domain/
│   ├── application/
│   ├── ports/input/
│   ├── ports/output/
│   └── adapters/{http,postgres,mongo,redis}/
├── pkg/{middleware,apierror}/
├── config/
├── migrations/
└── api/
```

### 2. Inicializar o Módulo Go

```bash
cd $SERVICE
go mod init github.com/<org>/$SERVICE
```

Adicionar as libs obrigatórias (conforme TECHNICAL_BASE seção 2.2):

```bash
go get github.com/go-chi/chi/v5
go get github.com/go-playground/validator/v10
go get github.com/jackc/pgx/v5          # se usar PostgreSQL
go get go.mongodb.org/mongo-driver/v2   # se usar MongoDB
go get github.com/redis/go-redis/v9
go get github.com/golang-jwt/jwt/v5
go get github.com/spf13/viper
go get go.uber.org/zap
go get go.opentelemetry.io/otel
go get github.com/golang-migrate/migrate/v4
go get github.com/stretchr/testify
```

### 3. Criar `config/config.go`

Leitura de variáveis de ambiente com viper. Validar obrigatórias no startup:

```go
package config

import (
    "fmt"
    "github.com/spf13/viper"
)

type Config struct {
    ServiceName string
    Port        string
    LogLevel    string
    DBUrl       string
    RedisAddr   string
}

func Load() (*Config, error) {
    viper.AutomaticEnv()

    cfg := &Config{
        ServiceName: viper.GetString("SERVICE_NAME"),
        Port:        viper.GetString("PORT"),
        LogLevel:    viper.GetString("LOG_LEVEL"),
        DBUrl:       viper.GetString("DB_URL"),
        RedisAddr:   viper.GetString("REDIS_ADDR"),
    }

    if cfg.DBUrl == "" {
        return nil, fmt.Errorf("DB_URL é obrigatório")
    }
    if cfg.RedisAddr == "" {
        return nil, fmt.Errorf("REDIS_ADDR é obrigatório")
    }

    return cfg, nil
}
```

Nomenclatura de variáveis: `{SERVICO}_{COMPONENTE}_{CAMPO}` (ex: `USER_SERVICE_DB_URL`)

### 4. Criar `cmd/server/main.go`

Este é o único arquivo onde dependências concretas são instanciadas e conectadas (DI manual):

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os/signal"
    "syscall"

    "github.com/go-chi/chi/v5"
    "go.uber.org/zap"

    "<org>/<service>/config"
    // importar adapters, use cases e handlers aqui
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("configuração inválida: %v", err)
    }

    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // --- Wiring de dependências ---
    // repo := postgres.NewUserRepository(pool, logger)
    // publisher := redis.NewEventPublisher(redisClient, logger)
    // uc := application.NewCreateUserUseCase(repo, publisher, logger)
    // handler := httpAdapter.NewUserHandler(uc, logger)

    r := chi.NewRouter()
    // r.Mount("/v1/users", handler.Routes())

    // Health checks (sem autenticação, sem Kong)
    r.Get("/healthz/live", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    })
    r.Get("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
        // TODO: verificar conexão com BD e Redis
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    })

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}

    go func() {
        logger.Info("servidor iniciado", zap.String("port", cfg.Port))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("erro no servidor", zap.Error(err))
        }
    }()

    <-ctx.Done()
    logger.Info("encerrando servidor...")
    srv.Shutdown(context.Background())
}
```

### 5. Criar `Dockerfile`

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/service /service
EXPOSE 8080
ENTRYPOINT ["/service"]
```

### 6. Criar `api/openapi.yaml` (Contrato Base)

```yaml
openapi: "3.0.3"
info:
  title: <Service Name> API
  version: "1.0.0"
servers:
  - url: /v1
paths:
  /healthz/live:
    get:
      summary: Liveness check
      responses:
        "200":
          description: Serviço está vivo
components:
  schemas:
    ErrorResponse:
      type: object
      properties:
        error:
          type: object
          properties:
            code:
              type: string
            message:
              type: string
            details:
              type: array
              items:
                type: object
                properties:
                  field:
                    type: string
                  message:
                    type: string
            trace_id:
              type: string
```

### 7. Criar Primeira Migration (se usar PostgreSQL)

```sql
-- migrations/000001_initial_schema.up.sql
-- (adicionar tabelas conforme o domínio do serviço)

-- migrations/000001_initial_schema.down.sql
-- DROP TABLE IF EXISTS ...;
```

### 8. Checklist Final

Antes de considerar o scaffold concluído, verificar:

- [ ] Árvore de diretórios criada conforme seção 3.2 do TECHNICAL_BASE
- [ ] `go.mod` com todas as libs obrigatórias
- [ ] `config.go` valida variáveis obrigatórias no startup
- [ ] `main.go` com DI manual e shutdown gracioso
- [ ] `GET /healthz/live` e `GET /healthz/ready` implementados
- [ ] `Dockerfile` multi-stage
- [ ] `api/openapi.yaml` com schema base e `ErrorResponse`
- [ ] Primeira migration criada (se usar PostgreSQL)
- [ ] Nenhuma variável global de dependência (somente construtores)
