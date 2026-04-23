# Implementar Novo Use Case

Adicionar um novo use case em um microsserviço existente, seguindo a arquitetura em camadas (domain → application → ports → adapters → handler).

## Antes de Começar

Leia as seguintes seções do [`TECHNICAL_BASE.md`](../../../TECHNICAL_BASE.md):
- Seção 3.2 — Arquitetura, camadas e regra de dependência
- Seção 5 — Padrões de API REST (URL, HTTP status, envelopes)
- Seção 9.3 — Error handling tipado
- Seção 9.4 — Injeção de dependência

## Informações Necessárias

Antes de implementar, confirme com o usuário:
1. Nome do use case (ex: `CreateUser`, `GetOrderByID`)
2. Operação HTTP (método + rota, ex: `POST /v1/users`)
3. Publica evento? (se sim, qual `event_type`)

## Passo a Passo

Siga esta ordem. Nunca pule camadas.

### 1 — Entidade/VO em `internal/domain/`

O domínio não importa nada externo. Apenas regras de negócio puras:

```go
// internal/domain/user.go
type User struct {
    ID    uuid.UUID
    Name  string
    Email string
}

func NewUser(name, email string) (*User, error) {
    if name == "" { return nil, ErrInvalidName }
    if email == "" { return nil, ErrInvalidEmail }
    return &User{ID: uuid.New(), Name: name, Email: email}, nil
}
```

### 2 — Erros de domínio em `internal/domain/errors.go`

```go
var (
    ErrUserNotFound    = errors.New("user not found")
    ErrEmailDuplicated = errors.New("email already in use")
)
```

### 3 — Interface de repositório em `internal/ports/output/`

```go
type UserRepository interface {
    Save(ctx context.Context, user *domain.User) error
    FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}
```

### 4 — Use case em `internal/application/`

```go
type CreateUserUseCase struct {
    repo   output.UserRepository
    logger *zap.Logger
}

func NewCreateUserUseCase(repo output.UserRepository, logger *zap.Logger) *CreateUserUseCase {
    return &CreateUserUseCase{repo: repo, logger: logger}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
    user, err := domain.NewUser(input.Name, input.Email)
    if err != nil { return nil, fmt.Errorf("create user: %w", err) }
    if err := uc.repo.Save(ctx, user); err != nil { return nil, fmt.Errorf("create user: %w", err) }
    return &CreateUserOutput{ID: user.ID.String(), Name: user.Name, Email: user.Email}, nil
}
```

### 5 — Adaptador de repositório em `internal/adapters/`

Implementar a interface definida no passo 3 com o banco de dados concreto.

### 6 — Handler HTTP em `internal/adapters/http/`

Mapear erros de domínio para HTTP status codes:

```go
switch {
case errors.Is(err, domain.ErrUserNotFound):
    apierror.WriteNotFound(w, err)
case errors.Is(err, domain.ErrEmailDuplicated):
    apierror.WriteConflict(w, err)
default:
    apierror.WriteInternal(w, err)
}
```

### 7 — Registrar no `cmd/server/main.go`

DI manual — instanciar e conectar as dependências.

## Checklist Final

- [ ] Entidade criada em `internal/domain/` sem imports externos
- [ ] Erros de domínio tipados em `internal/domain/errors.go`
- [ ] Interface de repositório em `internal/ports/output/`
- [ ] Use case em `internal/application/` recebe dependências via construtor
- [ ] Adaptador implementa a interface
- [ ] Handler mapeia erros de domínio para HTTP status codes
- [ ] Rota registrada e DI feita em `main.go`
- [ ] Testes unitários do use case com mocks (cobertura ≥ 80%)
- [ ] Endpoint documentado em `api/openapi.yaml`
