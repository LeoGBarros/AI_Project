---
name: implement-usecase
description: Adicionar um novo use case em um microsserviço existente, seguindo a arquitetura hexagonal (domain → application → ports → adapters → handler). Use quando o usuário pedir para implementar uma nova funcionalidade, endpoint ou operação em um serviço já existente.
---

# Implementar Novo Use Case (Hexagonal)

## Antes de Começar

Leia as seguintes seções do [`TECHNICAL_BASE.md`](../../../TECHNICAL_BASE.md):
- **Seção 3.2** — Arquitetura hexagonal, camadas e regra de dependência
- **Seção 5** — Padrões de API REST (URL, HTTP status, envelopes)
- **Seção 9.3** — Error handling tipado
- **Seção 9.4** — Injeção de dependência

## Informações Necessárias

Antes de implementar, confirme com o usuário:
1. **Nome do use case** (ex: `CreateUser`, `GetOrderByID`, `CancelOrder`)
2. **Operação HTTP** (método + rota, ex: `POST /v1/users`)
3. **Banco de dados** envolvido (PostgreSQL, MongoDB)
4. **Publica evento?** (se sim, qual `event_type`)
5. **Há mudança de schema?** (se sim, criar migration)

---

## Passo a Passo

Siga esta ordem. Nunca pule camadas — cada passo depende do anterior.

### Passo 1 — Definir a Entidade/VO em `internal/domain/`

O domínio não importa nada externo. Defina apenas regras de negócio puras:

```go
// internal/domain/user.go
package domain

import "github.com/google/uuid"

type User struct {
    ID    uuid.UUID
    Name  string
    Email string
}

// Regra de negócio: validação no domínio
func NewUser(name, email string) (*User, error) {
    if name == "" {
        return nil, ErrInvalidName
    }
    if email == "" {
        return nil, ErrInvalidEmail
    }
    return &User{ID: uuid.New(), Name: name, Email: email}, nil
}
```

### Passo 2 — Definir Erros de Domínio em `internal/domain/errors.go`

```go
// internal/domain/errors.go
package domain

import "errors"

var (
    ErrUserNotFound    = errors.New("user not found")
    ErrEmailDuplicated = errors.New("email already in use")
    ErrInvalidName     = errors.New("name is required")
    ErrInvalidEmail    = errors.New("email is required")
)
```

- Erros de domínio são verificáveis via `errors.Is` / `errors.As`
- Nunca retorne strings cruas de erro — sempre use variáveis tipadas

### Passo 3 — Definir Interface de Repositório em `internal/ports/output/`

```go
// internal/ports/output/repository.go
package output

import (
    "context"
    "github.com/google/uuid"
    "<org>/<service>/internal/domain"
)

type UserRepository interface {
    Save(ctx context.Context, user *domain.User) error
    FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
}
```

- A interface é definida aqui (onde é **usada** pelo use case), não no adaptador
- Nenhum tipo específico de banco de dados nesta interface

### Passo 4 — Implementar o Use Case em `internal/application/`

```go
// internal/application/create_user.go
package application

import (
    "context"
    "fmt"

    "go.uber.org/zap"
    "<org>/<service>/internal/domain"
    "output "<org>/<service>/internal/ports/output"
)

type CreateUserUseCase struct {
    repo      output.UserRepository
    publisher output.EventPublisher // opcional, se publicar evento
    logger    *zap.Logger
}

func NewCreateUserUseCase(
    repo output.UserRepository,
    publisher output.EventPublisher,
    logger *zap.Logger,
) *CreateUserUseCase {
    return &CreateUserUseCase{repo: repo, publisher: publisher, logger: logger}
}

type CreateUserInput struct {
    Name  string
    Email string
}

type CreateUserOutput struct {
    ID    string
    Name  string
    Email string
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
    // 1. Verificar duplicata antes de criar
    existing, _ := uc.repo.FindByEmail(ctx, input.Email)
    if existing != nil {
        return nil, domain.ErrEmailDuplicated
    }

    // 2. Criar entidade via construtor do domínio (validação embutida)
    user, err := domain.NewUser(input.Name, input.Email)
    if err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }

    // 3. Persistir
    if err := uc.repo.Save(ctx, user); err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }

    // 4. Publicar evento (opcional)
    // uc.publisher.Publish(ctx, "user-service.user.created", ...)

    uc.logger.Info("usuário criado", zap.String("user_id", user.ID.String()))

    return &CreateUserOutput{ID: user.ID.String(), Name: user.Name, Email: user.Email}, nil
}
```

### Passo 5 — Implementar o Adaptador de Repositório

**PostgreSQL** (`internal/adapters/postgres/user_repository.go`):

```go
package postgres

import (
    "context"
    "errors"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "go.uber.org/zap"
    "<org>/<service>/internal/domain"
)

type UserRepository struct {
    pool   *pgxpool.Pool
    logger *zap.Logger
}

func NewUserRepository(pool *pgxpool.Pool, logger *zap.Logger) *UserRepository {
    return &UserRepository{pool: pool, logger: logger}
}

func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
    _, err := r.pool.Exec(ctx,
        `INSERT INTO users (id, name, email, created_at, updated_at)
         VALUES ($1, $2, $3, NOW(), NOW())`,
        user.ID, user.Name, user.Email,
    )
    return err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    row := r.pool.QueryRow(ctx,
        `SELECT id, name, email FROM users WHERE id = $1 AND deleted_at IS NULL`,
        id,
    )
    var u domain.User
    if err := row.Scan(&u.ID, &u.Name, &u.Email); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, domain.ErrUserNotFound
        }
        return nil, err
    }
    return &u, nil
}
```

### Passo 6 — Implementar o Handler HTTP em `internal/adapters/http/`

```go
// internal/adapters/http/user_handler.go
package http

import (
    "encoding/json"
    "errors"
    "net/http"

    "github.com/go-chi/chi/v5"
    "go.uber.org/zap"
    "<org>/<service>/internal/application"
    "<org>/<service>/internal/domain"
    "<org>/<service>/pkg/apierror"
)

type UserHandler struct {
    createUser *application.CreateUserUseCase
    logger     *zap.Logger
}

func NewUserHandler(createUser *application.CreateUserUseCase, logger *zap.Logger) *UserHandler {
    return &UserHandler{createUser: createUser, logger: logger}
}

func (h *UserHandler) Routes() http.Handler {
    r := chi.NewRouter()
    r.Post("/", h.CreateUser)
    return r
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var body struct {
        Name  string `json:"name"  validate:"required"`
        Email string `json:"email" validate:"required,email"`
    }

    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        apierror.WriteBadRequest(w, "body inválido", nil)
        return
    }

    out, err := h.createUser.Execute(r.Context(), application.CreateUserInput{
        Name:  body.Name,
        Email: body.Email,
    })
    if err != nil {
        switch {
        case errors.Is(err, domain.ErrEmailDuplicated):
            apierror.WriteConflict(w, err.Error())
        default:
            apierror.WriteInternal(w, err.Error())
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(out)
}
```

### Passo 7 — Registrar no `cmd/server/main.go`

```go
// Instanciar e conectar (DI manual)
userRepo := postgres.NewUserRepository(pool, logger)
publisher := redis.NewEventPublisher(redisClient, logger)
createUserUC := application.NewCreateUserUseCase(userRepo, publisher, logger)
userHandler := httpAdapter.NewUserHandler(createUserUC, logger)

r.Mount("/v1/users", userHandler.Routes())
```

### Passo 8 — Criar Migration (se houver mudança de schema)

```sql
-- migrations/000002_create_users_table.up.sql
CREATE TABLE users (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    email       VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_users_email ON users(email) WHERE deleted_at IS NULL;
```

```sql
-- migrations/000002_create_users_table.down.sql
DROP TABLE IF EXISTS users;
```

---

## Checklist Final

- [ ] Entidade/VO criada em `internal/domain/` sem imports externos
- [ ] Erros de domínio tipados em `internal/domain/errors.go`
- [ ] Interface de repositório definida em `internal/ports/output/`
- [ ] Use case em `internal/application/` recebe dependências via construtor
- [ ] Adaptador em `internal/adapters/<banco>/` implementa a interface
- [ ] Handler HTTP mapeia erros de domínio para HTTP status codes corretos
- [ ] Rota registrada e dependências injetadas em `main.go`
- [ ] Migration criada se houve mudança de schema
- [ ] Testes unitários do use case com mocks das interfaces (cobertura ≥ 80%)
- [ ] Endpoint documentado em `api/openapi.yaml`
