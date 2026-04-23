# Plano de Implementação — user-service

## Status
- `[ ]` Não iniciado
- `[~]` Em progresso
- `[-]` Pendente
- `[x]` Concluído

---

## Tarefas

- [x] 4.3 Escrever testes unitários para `GetProfileUseCase`
  - Cenários: perfil existente, auto-criação, user_id vazio, erro inesperado do repositório
  - Usar mocks para `UserProfileRepository`
  - Verificar que auto-criação usa claims JWT (`PreferredUsername`, `Email`) conforme design.md
  - Verificar que `user_id` vazio retorna `domain.ErrInvalidToken` (verificável via `errors.Is`)
  - Formato: `TestGetProfile_<Scenario>_<ExpectedResult>` (table-driven tests)
  - _Requisitos: 1.1, 1.2, 1.3, 4.3_

- [x] 4.4 Implementar `user-service/internal/application/update_profile_usecase.go`
  - Receber `UserProfileRepository`, `EventPublisher` e `*zap.Logger` via construtor (DI manual)
  - Struct `UpdateProfileInput` conforme design.md: campo nomeado `Fields domain.UpdateProfileInput` (NÃO usar embedding)
  - Validar campos via `input.Fields.Validate()`; se inválido, retornar erro que carregue o `map[string]string` de detalhes por campo (conforme Requisito 2.4)
  - Considerar criar `ValidationError` em `domain/errors.go` com `Details map[string]string`, compatível com `errors.Is(err, domain.ErrValidation)`
  - Atualizar perfil no repositório; publicar evento `user.updated`; se publicação falhar, logar erro sem reverter (fire-and-forget, Requisito 3.5)
  - Adicionar span OpenTelemetry: `otel.Tracer("user-service").Start(ctx, "usecase.UpdateProfile")` (conforme padrão do auth-service)
  - _Requisitos: 2.1, 2.2, 2.3, 3.1, 3.5_
  - _Referência: design.md seção UpdateProfileUseCase_

- [~] 4.5 Escrever teste de propriedade para atualização com timestamp
  - **Propriedade 4: Atualização reflete mudanças com timestamp atualizado**
  - Mock de `UserProfileRepository` que retorna perfil com `UpdatedAt` atualizado
  - Gerar `UpdateProfileInput` aleatório com pelo menos um campo não-nil via `rapid`
  - Verificar que `UpdatedAt >= timestamp antes da operação` e campos não incluídos permanecem inalterados
  - Mínimo 100 iterações
  - **Valida: Requisitos 2.2, 2.3**

- [~] 4.6 Escrever teste de propriedade para falha na publicação de evento
  - **Propriedade 6: Falha na publicação de evento não reverte a atualização**
  - Mock de `EventPublisher` que sempre retorna erro + mock de repo que retorna sucesso
  - Verificar que `Execute` retorna sucesso mesmo quando publisher falha
  - Mínimo 100 iterações
  - **Valida: Requisito 3.5**

- [~] 4.7 Escrever testes unitários para `UpdateProfileUseCase`
  - Cenários (table-driven tests):
    - Atualização válida com todos os campos → perfil atualizado + evento publicado
    - `UserID` vazio → `domain.ErrInvalidToken`
    - Campos inválidos → erro de validação COM detalhes por campo (acessíveis via `errors.As`)
    - Perfil não encontrado → erro wrapping `domain.ErrUserNotFound` (via `errors.Is`)
    - Falha no publisher → sucesso com perfil atualizado (fire-and-forget)
  - Usar mocks para `UserProfileRepository` e `EventPublisher`
  - _Requisitos: 2.1, 2.2, 2.4, 2.5, 3.5_

- [~] 5. Checkpoint — Verificar domain e application
  - Garantir que todos os testes passam, perguntar ao usuário se houver dúvidas.

- [ ] 6. Camada Adapters — implementações concretas
  - [~] 6.1 Criar `user-service/pkg/apierror/errors.go`
    - Reutilizar o mesmo padrão do `auth-service/pkg/apierror/errors.go`: funções `WriteError` e `WriteValidationError`
    - `WriteError(w, statusCode, code, message, traceID)` — envelope com `code`, `message`, `trace_id`
    - `WriteValidationError(w, traceID, details)` — envelope com `code: "VALIDATION_ERROR"`, `details` (array de `{field, message}`)
    - Nunca expor stack traces em erros 500 (Requisito 12.2)
    - _Requisitos: 12.1, 12.2_
    - _Referência: `auth-service/pkg/apierror/errors.go`_

  - [~] 6.2 Criar `user-service/pkg/middleware/` (correlation, logging, tracing)
    - Copiar e adaptar do `auth-service/pkg/middleware/`: `CorrelationID`, `Logging`, `Tracing`
    - `CorrelationID`: extrair `X-Correlation-ID` do header, gerar UUID se ausente, injetar no context
    - `Logging`: campos obrigatórios conforme TECHNICAL_BASE.md seção 6.3
    - `Tracing`: span raiz por request com atributos `http.method`, `http.route`, `http.status_code`
    - _Requisitos: 7.1, 7.2, 7.3_
    - _Referência: `auth-service/pkg/middleware/`_

  - [~] 6.3 Implementar `user-service/internal/adapters/http/handler.go`
    - Implementar `ports/input.UserHandler` com chi router
    - Receber `GetProfileUseCase`, `UpdateProfileUseCase` e `*zap.Logger` via construtor
    - Extrair JWT sem verificar assinatura (confiar no Kong — design.md decisão 1)
    - Parsear claims: `sub` → `user_id`, `preferred_username`, `email`
    - Retornar 401 via `apierror.WriteError` se claim `sub` ausente ou vazia
    - Para PUT: montar `UpdateProfileInput{UserID, Fields: domain.UpdateProfileInput{...}}`
    - Mapear erros conforme design.md "Fluxo de Erros no Handler":
      - `ErrUserNotFound` → 404, `ErrInvalidToken` → 401, `*ValidationError` (via `errors.As`) → 400 com detalhes, outros → 500
    - _Requisitos: 1.1, 2.1, 4.1, 4.2, 4.3, 12.3_
    - _Referência: design.md seção "Fluxo de Erros no Handler"_

  - [~] 6.4 Escrever teste de propriedade para extração de claims JWT
    - **Propriedade 1: Extração de claims JWT (round-trip)**
    - Gerar JWTs aleatórios com `sub` não-vazio via `rapid`; verificar round-trip
    - Mínimo 100 iterações
    - **Valida: Requisitos 1.1, 4.1, 4.2**

  - [~] 6.5 Escrever teste de propriedade para mapeamento de erros
    - **Propriedade 9: Mapeamento de erros de domínio para HTTP**
    - Gerar erros aleatórios; verificar mapeamento correto e que erros 500 não expõem detalhes
    - Mínimo 100 iterações
    - **Valida: Requisitos 12.1, 12.2, 12.3**

  - [~] 6.6 Escrever testes unitários para o handler HTTP
    - Cenários (table-driven): GET existente, GET auto-criação, GET sem sub, PUT válido, PUT inválido (400 com detalhes), PUT not found (404), PUT body inválido
    - Usar mocks para use cases
    - Formato: `TestHandler_<Method>_<Scenario>_<ExpectedResult>`
    - _Requisitos: 1.1, 1.2, 2.1, 2.4, 4.3, 12.3_

  - [~] 6.7 Implementar `user-service/internal/adapters/postgres/repository.go`
    - Implementar `ports/output.UserProfileRepository` com `pgx/v5`; receber `*pgxpool.Pool` via construtor
    - `FindByID`: SELECT com `WHERE deleted_at IS NULL`; retornar `domain.ErrUserNotFound` se não encontrar
    - `Update`: query dinâmica com campos não-nil; sempre atualizar `updated_at = NOW()`; retornar `ErrUserNotFound` se 0 rows
    - Instrumentar com spans OpenTelemetry (`db.system = "postgresql"`)
    - Usar parametrized queries — nunca concatenar SQL
    - _Requisitos: 5.1, 5.3, 7.4, 11.4_

  - [~] 6.8 Implementar `user-service/internal/adapters/redis/publisher.go`
    - Implementar `ports/output.EventPublisher` com `go-redis/v9`; receber `*redis.Client` via construtor
    - Envelope conforme TECHNICAL_BASE.md seção 3.4: `event_id` (UUID v4), `event_type: "user.updated"`, `source_service: "user-service"`, `version: "1"`, `timestamp` (ISO 8601), `correlation_id` (do context), `payload`
    - Canal: `user-service.user.updated`
    - Instrumentar com span OpenTelemetry
    - _Requisitos: 3.1, 3.2, 3.3, 3.4, 11.5_

  - [~] 6.9 Escrever teste de propriedade para envelope de evento
    - **Propriedade 5: Envelope de evento contém todos os campos obrigatórios**
    - Gerar dados aleatórios; verificar presença e formato de todos os campos do envelope
    - Mínimo 100 iterações
    - **Valida: Requisitos 3.1, 3.2, 3.3, 3.4**

- [~] 7. Checkpoint — Verificar adapters
  - Garantir que todos os testes passam, perguntar ao usuário se houver dúvidas.

- [ ] 8. Wiring, health checks e containerização
  - [~] 8.1 Implementar `user-service/cmd/server/main.go`
    - Carregar configuração via `config.Load()`; se erro, `log.Fatal`
    - Inicializar logger `zap.NewProduction()` com `service: "user-service"`; nível via `USER_SERVICE_LOG_LEVEL`
    - Inicializar OpenTelemetry tracer provider com exporter OTLP; defer shutdown
    - Conectar PostgreSQL com `pgxpool` (MaxConns=20, MinConns=5, MaxConnLifetime=30min, MaxConnIdleTime=5min)
    - Validar conexão PostgreSQL no startup com `pool.Ping(ctx)` (fail fast, Requisito 5.5)
    - Conectar Redis e validar com Ping
    - Instanciar adapters e use cases via DI manual (TECHNICAL_BASE.md seção 9.4)
    - Configurar chi router: `Recoverer`, `RequestID`, `CorrelationID`, `Tracing`, `Logging`
    - Registrar rotas `/v1/users` e health checks `/healthz`
    - Graceful shutdown: `SIGINT`/`SIGTERM`, timeout 10s
    - _Requisitos: 5.4, 5.5, 7.3, 8.1, 8.2, 8.3, 8.4_
    - _Referência: `auth-service/cmd/server/main.go`_

  - [~] 8.2 Implementar health checks no `main.go`
    - `GET /healthz/live` → 200 `{"status":"ok"}`
    - `GET /healthz/ready` → verificar `pool.Ping` e `redis.Ping`:
      - Ambos OK → 200; PostgreSQL falhou → 503 `"postgresql unavailable"`; Redis falhou → 503 `"redis unavailable"`; ambos → 503 `"postgresql and redis unavailable"`
    - Sem autenticação, fora do Kong (Requisito 6.4)
    - _Requisitos: 6.1, 6.2, 6.3, 6.4_

  - [~] 8.3 Escrever teste de propriedade para readiness check
    - **Propriedade 7: Readiness check reflete estado das dependências**
    - Gerar combinações de estados via `rapid`; verificar 200 vs 503 com reason correto
    - Mínimo 100 iterações
    - **Valida: Requisitos 6.2, 6.3**

  - [~] 8.4 Criar `user-service/Dockerfile`
    - Multi-stage conforme TECHNICAL_BASE.md seção 9.9
    - Builder: `golang:1.22-alpine`; Runtime: `gcr.io/distroless/static-debian12`
    - _Requisitos: 9.1, 9.3_
    - _Referência: `auth-service/Dockerfile`_

  - [~] 8.5 Adicionar `user-service` ao `docker-compose.yml`
    - `build: ./user-service`; `depends_on: redis, postgres`
    - Env: `USER_SERVICE_PORT=8080`, `USER_SERVICE_DB_URL`, `USER_SERVICE_REDIS_ADDR=redis:6379`, `USER_SERVICE_LOG_LEVEL=debug`
    - Porta: `8083:8080`; não alterar serviços existentes
    - _Requisitos: 9.2, 9.3_

- [~] 9. Checkpoint final — Validação completa
  - Garantir que todos os testes passam, perguntar ao usuário se houver dúvidas.

---

## Notas

- Tarefas marcadas com `*` são opcionais e podem ser puladas para um MVP mais rápido
- Cada tarefa referencia requisitos específicos para rastreabilidade
- Checkpoints garantem validação incremental
- Testes de propriedade validam propriedades universais de corretude
- Testes unitários validam exemplos específicos e edge cases
