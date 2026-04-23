---
inclusion: fileMatch
fileMatchPattern: "**/*.go"
---

# Padrões de Codificação Go

Referência completa: [`TECHNICAL_BASE.md` — seção 9](../../TECHNICAL_BASE.md#9-padrões-de-codificação-go)

## Estrutura de Diretórios

```
service-name/
├── cmd/server/main.go              # Único ponto de wiring de DI
├── internal/
│   ├── domain/                     # Entidades, VOs, erros. Zero imports externos.
│   ├── application/                # Use cases com interfaces injetadas
│   ├── ports/
│   │   ├── input/handler.go        # Interface do handler HTTP
│   │   └── output/                 # Interfaces de repositório e publisher
│   └── adapters/                   # Implementações concretas
├── pkg/                            # Pacotes compartilhados
├── config/config.go
├── api/openapi.yaml
├── Dockerfile
└── go.mod
```

## Regra de Dependência

```
domain  ←  application  ←  adapters  ←  cmd/main.go
```

- `domain`: zero imports externos ao pacote
- `application`: importa apenas `domain` e `ports/output`
- `adapters`: implementam as interfaces de `ports`
- `cmd/main.go`: único lugar onde dependências concretas são instanciadas

## Bibliotecas Recomendadas

| Propósito | Biblioteca |
|---|---|
| HTTP server | `net/http` + `github.com/go-chi/chi/v5` |
| Validação | `github.com/go-playground/validator/v10` |
| Redis | `github.com/redis/go-redis/v9` |
| JWT | `github.com/golang-jwt/jwt/v5` |
| Configuração | `github.com/spf13/viper` |
| Logging | `go.uber.org/zap` |
| OpenTelemetry | `go.opentelemetry.io/otel` |
| Testes | `github.com/stretchr/testify` |

## Error Handling

- Nunca ignore erros com `_`
- Erros de domínio verificáveis via `errors.Is` / `errors.As`
- Encapsule com contexto: `fmt.Errorf("operacao: %w", err)`
- Na camada HTTP, mapeie erros de domínio para HTTP status codes

## Injeção de Dependência

- DI manual, sem frameworks
- `main.go` é o único lugar onde implementações concretas são instanciadas
- Dependências sempre via construtor, nunca variáveis globais

## context.Context

- Sempre o primeiro parâmetro em funções que fazem I/O
- Nunca armazene context em structs

## Nomeação e Estilo

- Pacotes em `lowercase` sem underscore
- Interfaces definidas onde são usadas, não onde são implementadas
- Linhas com no máximo 120 caracteres
- Comentários de exportação obrigatórios em funções e tipos públicos

## Testes

- Cobertura mínima: 80% das funções de `domain` e `application`
- Testes unitários sem dependências externas (use mocks)
- Formato: `Test<FuncName>_<Scenario>_<ExpectedResult>`
- Prefira table-driven tests
