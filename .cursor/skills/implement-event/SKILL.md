---
name: implement-event
description: Implementar publicação ou consumo de eventos via Redis Pub/Sub, seguindo o envelope padrão, idempotência e os padrões definidos no TECHNICAL_BASE. Use quando o usuário pedir para publicar um evento de domínio ou para criar um consumer de eventos.
---

# Implementar Evento Redis Pub/Sub

## Antes de Começar

Leia os seguintes documentos:
- [`docs/diagrams/pubsub-event-flow.md`](../../../docs/diagrams/pubsub-event-flow.md) — fluxo completo com idempotência
- [`TECHNICAL_BASE.md` — seção 3.3](../../../TECHNICAL_BASE.md#33-comunicação-entre-serviços) — padrões de mensageria
- [`TECHNICAL_BASE.md` — seção 7.3](../../../TECHNICAL_BASE.md#73-redis) — namespacing e canal Pub/Sub

## Informações Necessárias

Antes de implementar, confirme:
1. **Tipo do evento** (ex: `user.created`, `order.cancelled`)
2. **Serviço publicador** (ex: `user-service`)
3. **Serviço(s) consumidor(es)** (ex: `notification-service`)
4. **Payload** do evento (campos relevantes)
5. **Implementando:** publicação, consumo, ou ambos?

---

## Envelope de Evento (Obrigatório)

Todo evento publicado deve seguir este envelope — sem exceções:

```json
{
  "event_id":      "uuid-v4",
  "event_type":    "user.created",
  "source_service": "user-service",
  "timestamp":     "2026-03-05T12:00:00Z",
  "correlation_id": "uuid-v4",
  "version":       "1",
  "payload":       {}
}
```

- `event_id`: UUID v4 único por evento — usado para deduplicação
- `event_type`: `<entidade>.<acao>` em minúsculas
- `source_service`: nome do serviço que publicou
- `correlation_id`: ID propagado do request original (`X-Correlation-ID`)
- `version`: versão do schema do payload (iniciar em `"1"`)
- Canal: `{source_service}.{event_type}` → ex: `user-service.user.created`

---

## Passo a Passo — Publicador

### 1. Definir Struct do Envelope em `internal/domain/` ou `pkg/`

```go
// pkg/event/envelope.go
package event

import "time"

type Envelope struct {
    EventID       string    `json:"event_id"`
    EventType     string    `json:"event_type"`
    SourceService string    `json:"source_service"`
    Timestamp     time.Time `json:"timestamp"`
    CorrelationID string    `json:"correlation_id"`
    Version       string    `json:"version"`
    Payload       any       `json:"payload"`
}
```

### 2. Definir Interface do Publisher em `internal/ports/output/publisher.go`

```go
// internal/ports/output/publisher.go
package output

import "context"

type EventPublisher interface {
    Publish(ctx context.Context, channel string, eventType string, payload any) error
}
```

### 3. Implementar em `internal/adapters/redis/publisher.go`

```go
// internal/adapters/redis/publisher.go
package redis

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
    "<org>/<service>/pkg/event"
)

type EventPublisher struct {
    client      *redis.Client
    serviceName string
    logger      *zap.Logger
}

func NewEventPublisher(client *redis.Client, serviceName string, logger *zap.Logger) *EventPublisher {
    return &EventPublisher{client: client, serviceName: serviceName, logger: logger}
}

func (p *EventPublisher) Publish(ctx context.Context, channel string, eventType string, payload any) error {
    // Extrair correlation_id do contexto (propagado pelo middleware)
    correlationID, _ := ctx.Value("correlation_id").(string)

    envelope := event.Envelope{
        EventID:       uuid.New().String(),
        EventType:     eventType,
        SourceService: p.serviceName,
        Timestamp:     time.Now().UTC(),
        CorrelationID: correlationID,
        Version:       "1",
        Payload:       payload,
    }

    data, err := json.Marshal(envelope)
    if err != nil {
        return fmt.Errorf("publish: marshal: %w", err)
    }

    if err := p.client.Publish(ctx, channel, data).Err(); err != nil {
        return fmt.Errorf("publish: redis: %w", err)
    }

    p.logger.Info("evento publicado",
        zap.String("channel", channel),
        zap.String("event_type", eventType),
        zap.String("event_id", envelope.EventID),
        zap.String("correlation_id", correlationID),
    )

    return nil
}
```

### 4. Chamar no Use Case

```go
// internal/application/create_user.go
func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
    // ... lógica de negócio ...

    if err := uc.publisher.Publish(
        ctx,
        "user-service.user.created",   // canal: {servico}.{evento}
        "user.created",                 // event_type
        map[string]string{             // payload específico do evento
            "user_id": user.ID.String(),
            "email":   user.Email,
        },
    ); err != nil {
        // Logar mas não falhar o use case por falha de publicação (decisão de negócio)
        uc.logger.Error("falha ao publicar evento", zap.Error(err))
    }

    return output, nil
}
```

---

## Passo a Passo — Consumidor

### 1. Criar o Consumer em `internal/adapters/redis/consumer.go`

```go
// internal/adapters/redis/consumer.go
package redis

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
    "<org>/<service>/pkg/event"
)

type EventConsumer struct {
    client  *redis.Client
    logger  *zap.Logger
    handler func(ctx context.Context, envelope event.Envelope) error
    dedupStore DedupStore // interface para verificar idempotência
}

func NewEventConsumer(
    client *redis.Client,
    handler func(ctx context.Context, envelope event.Envelope) error,
    dedupStore DedupStore,
    logger *zap.Logger,
) *EventConsumer {
    return &EventConsumer{client: client, handler: handler, dedupStore: dedupStore, logger: logger}
}

func (c *EventConsumer) Subscribe(ctx context.Context, channels ...string) error {
    sub := c.client.Subscribe(ctx, channels...)
    defer sub.Close()

    ch := sub.Channel()
    for {
        select {
        case msg, ok := <-ch:
            if !ok {
                return nil
            }
            if err := c.processMessage(ctx, msg); err != nil {
                c.logger.Error("erro ao processar mensagem", zap.Error(err), zap.String("channel", msg.Channel))
            }
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (c *EventConsumer) processMessage(ctx context.Context, msg *redis.Message) error {
    var envelope event.Envelope
    if err := json.Unmarshal([]byte(msg.Payload), &envelope); err != nil {
        return fmt.Errorf("unmarshal: %w", err)
    }

    // Idempotência: verificar se event_id já foi processado
    already, err := c.dedupStore.HasProcessed(ctx, envelope.EventID)
    if err != nil {
        return fmt.Errorf("dedup check: %w", err)
    }
    if already {
        c.logger.Info("evento duplicado ignorado", zap.String("event_id", envelope.EventID))
        return nil
    }

    // Propagar correlation_id no contexto
    ctx = context.WithValue(ctx, "correlation_id", envelope.CorrelationID)

    // Processar
    if err := c.handler(ctx, envelope); err != nil {
        return fmt.Errorf("handler: %w", err)
    }

    // Marcar como processado (TTL: 24h para deduplicação)
    return c.dedupStore.MarkProcessed(ctx, envelope.EventID)
}
```

### 2. Implementar a Deduplicação via Redis

```go
// internal/adapters/redis/dedup.go
package redis

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

type DedupStore interface {
    HasProcessed(ctx context.Context, eventID string) (bool, error)
    MarkProcessed(ctx context.Context, eventID string) error
}

type RedisDedupStore struct {
    client      *redis.Client
    serviceName string
}

func NewRedisDedupStore(client *redis.Client, serviceName string) *RedisDedupStore {
    return &RedisDedupStore{client: client, serviceName: serviceName}
}

func (d *RedisDedupStore) HasProcessed(ctx context.Context, eventID string) (bool, error) {
    key := fmt.Sprintf("%s:processed-events:%s", d.serviceName, eventID)
    exists, err := d.client.Exists(ctx, key).Result()
    return exists > 0, err
}

func (d *RedisDedupStore) MarkProcessed(ctx context.Context, eventID string) error {
    key := fmt.Sprintf("%s:processed-events:%s", d.serviceName, eventID)
    return d.client.Set(ctx, key, "1", 24*time.Hour).Err()
}
```

### 3. Iniciar o Consumer no `cmd/server/main.go`

```go
// Iniciar consumer em goroutine com shutdown via context
dedupStore := redisAdapter.NewRedisDedupStore(redisClient, "notification-service")
consumer := redisAdapter.NewEventConsumer(
    redisClient,
    func(ctx context.Context, env event.Envelope) error {
        // Chamar use case específico conforme event_type
        switch env.EventType {
        case "user.created":
            return sendWelcomeEmailUC.Execute(ctx, env.Payload)
        }
        return nil
    },
    dedupStore,
    logger,
)

go func() {
    if err := consumer.Subscribe(ctx, "user-service.user.created"); err != nil && err != context.Canceled {
        logger.Error("consumer encerrado com erro", zap.Error(err))
    }
}()
```

---

## Checklist Final

**Publicador:**
- [ ] Interface `EventPublisher` definida em `ports/output/publisher.go`
- [ ] Implementação em `adapters/redis/publisher.go` com envelope completo
- [ ] `event_id` gerado como UUID v4
- [ ] `correlation_id` extraído e propagado do contexto
- [ ] Canal no formato `{servico}.{evento}`
- [ ] Log de publicação com `event_id` e `correlation_id`

**Consumidor:**
- [ ] Deduplicação via `event_id` implementada (Redis com TTL 24h)
- [ ] `correlation_id` propagado no contexto para logs e spans
- [ ] Handler idempotente: processar o mesmo evento N vezes produz o mesmo resultado
- [ ] Consumer iniciado em goroutine com shutdown via context cancelation
- [ ] Erros de processamento logados sem derrubar o consumer
