# Plano de Implementação — user-service

## Status
- `[ ]` Não iniciado
- `[~]` Em progresso
- `[-]` Pendente
- `[x]` Concluído

---

## Tarefas

- [x] 4.3 Escrever testes unitários para `GetProfileUseCase`
  - Cenários: perfil existente, auto-criação, user_id vazio
  - Usar mocks para `UserProfileRepository`
  - _Requisitos: 1.1, 1.2, 1.3, 4.3_

- [~] 4.4 Implementar `user-service/internal/application/update_profile_usecase.go`
  - Receber `UserProfileRepository`, `EventPublisher` e `*zap.Logger` via construtor
  - Validar campos via `domain.UpdateProfileInput.Validate()`
  - Atualizar perfil no repositório; publicar evento `user.updated`; se publicação falhar, logar erro sem reverter
  - _Requisitos: 2.1, 2.2, 2.3, 3.1, 3.5_

- [~] 4.5 Escrever teste de propriedade para atualização com timestamp
  - **Propriedade 4: Atualização reflete mudanças com timestamp atualizado**
  - **Valida: Requisitos 2.2, 2.3**

- [~] 4.6 Escrever teste de propriedade para falha na publicação de evento
  - **Propriedade 6: Falha na publicação de evento não reverte a atualização**
  - **Valida: Requisito 3.5**

- [~] 4.7 Escrever testes unitários para `UpdateProfileUseCase`
  - Cenários: atualização válida, campos inválidos (400), perfil não encontrado (404), falha no publisher
  - Usar mocks para `UserProfileRepository` e `EventPublisher`
  - _Requisitos: 2.1, 2.2, 2.4, 2.5, 3.5_

- [~] 5. Checkpoint — Verificar domain e application
  - Garantir que todos os testes passam, perguntar ao usuário se houver dúvidas.

- [ ] 6. Camada Adapters — implementações concretas
  - [~] 6.1 Criar `user-service/pkg/apierror/errors.go`
    - Reutilizar o mesmo padrão do `auth-service`: funções `WriteError` e `WriteValidationError`
    - Envelope com campos `code`, `message`, `details`, `trace_id`
    - _Requisitos: 12.1, 12.2_

  - [~] 6.2 Criar `user-service/pkg/middleware/` (correlation, logging, tracing)
    - Reutilizar o mesmo padrão do `auth-service`: `CorrelationID`, `Logging`, `Tracing`
    - _Requisitos: 7.1, 7.2, 7.3_

  - [~] 6.3 Implementar `user-service/internal/adapters/http/handler.go`
    - Implementar `ports/input.UserHandler` com chi router
    - Extrair JWT do header `Authorization` (decode sem verificação de assinatura, confiar no Kong)
    - Parsear claim `sub` (user_id), `preferred_username`, `email`
    - Retornar 401 se claim `sub` ausente ou vazia
    - Mapear erros de domínio para HTTP status codes via `pkg/apierror`
    - Registrar rotas: `GET /v1/users/me`, `PUT /v1/users/me`
    - _Requisitos: 1.1, 2.1, 4.1, 4.2, 4.3, 12.3_

  - [~] 6.4 Escrever teste de propriedade para extração de claims JWT
    - **Propriedade 1: Extração de claims JWT (round-trip)**
    - **Valida: Requisitos 1.1, 4.1, 4.2**

  - [~] 6.5 Escrever teste de propriedade para mapeamento de erros
    - **Propriedade 9: Mapeamento de erros de domínio para HTTP**
    - **Valida: Requisitos 12.1, 12.2, 12.3**

  - [~] 6.6 Escrever testes unitários para o handler HTTP
    - Cenários: GET com perfil existente, GET com auto-criação, PUT válido, PUT com validação inválida, PUT com perfil não encontrado, JWT sem claim sub
    - Usar mocks para os use cases
    - _Requisitos: 1.1, 1.2, 2.1, 2.4, 4.3, 12.3_

  - [~] 6.7 Implementar `user-service/internal/adapters/postgres/repository.go`
    - Implementar `ports/output.UserProfileRepository` com `pgx/v5`
    - Queries: `SELECT`, `INSERT`, `UPDATE` na tabela `user_profiles`
    - Instrumentar queries com spans OpenTelemetry (`db.system = postgresql`)
    - _Requisitos: 5.1, 5.3, 7.4, 11.4_

  - [~] 6.8 Implementar `user-service/internal/adapters/redis/publisher.go`
    - Implementar `ports/output.EventPublisher` com `go-redis/v9`
    - Formatar envelope de evento com `event_id`, `event_type`, `source_service`, `timestamp`, `correlation_id`, `version`, `payload`
    - Canal: `user-service.user.updated`
    - Propagar `correlation_id` do contexto
    - _Requisitos: 3.1, 3.2, 3.3, 3.4, 11.5_

  - [~] 6.9 Escrever teste de propriedade para envelope de evento
    - **Propriedade 5: Envelope de evento contém todos os campos obrigatórios**
    - **Valida: Requisitos 3.1, 3.2, 3.3, 3.4**

- [~] 7. Checkpoint — Verificar adapters
  - Garantir que todos os testes passam, perguntar ao usuário se houver dúvidas.

- [ ] 8. Wiring, health checks e containerização
  - [~] 8.1 Implementar `user-service/cmd/server/main.go`
    - Carregar configuração via `config.Load()`
    - Inicializar logger zap, OpenTelemetry tracer
    - Conectar PostgreSQL com `pgxpool` (MaxConns=20, MinConns=5, MaxConnLifetime=30min, MaxConnIdleTime=5min)
    - Validar conexão PostgreSQL no startup (fail fast)
    - Conectar Redis e validar com Ping
    - Instanciar adapters (repository, publisher) e use cases via DI manual
    - Configurar chi router com middlewares (Recoverer, RequestID, CorrelationID, Tracing, Logging)
    - Registrar rotas `/v1/users` e health checks
    - Implementar graceful shutdown com timeout de 10 segundos (SIGINT, SIGTERM)
    - _Requisitos: 5.4, 5.5, 7.3, 8.1, 8.2, 8.3, 8.4_

  - [~] 8.2 Implementar health checks no `main.go`
    - `GET /healthz/live` → 200 `{"status":"ok"}`
    - `GET /healthz/ready` → 200 se PostgreSQL e Redis OK; 503 com `reason` se algum indisponível
    - Sem autenticação, fora do roteamento do Kong
    - _Requisitos: 6.1, 6.2, 6.3, 6.4_

  - [~] 8.3 Escrever teste de propriedade para readiness check
    - **Propriedade 7: Readiness check reflete estado das dependências**
    - **Valida: Requisitos 6.2, 6.3**

  - [~] 8.4 Criar `user-service/Dockerfile`
    - Multi-stage: `golang:1.22-alpine` (builder) → `gcr.io/distroless/static-debian12` (runtime)
    - Seguir o mesmo padrão do `auth-service/Dockerfile`
    - _Requisitos: 9.1, 9.3_

  - [~] 8.5 Adicionar `user-service` ao `docker-compose.yml`
    - Dependências: `redis`, `postgres`
    - Variáveis de ambiente: `USER_SERVICE_PORT`, `USER_SERVICE_DB_URL`, `USER_SERVICE_REDIS_ADDR`, `USER_SERVICE_LOG_LEVEL`
    - Expor porta mapeada (ex: `8083:8080`)
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
