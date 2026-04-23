# Implementar Evento Redis Pub/Sub

Implementar publicação ou consumo de eventos via Redis Pub/Sub, seguindo o envelope padrão e idempotência.

## Antes de Começar

- [`docs/diagrams/pubsub-event-flow.md`](../../../docs/diagrams/pubsub-event-flow.md) — fluxo completo
- [`TECHNICAL_BASE.md` — seção 3.4](../../../TECHNICAL_BASE.md#34-comunicação-entre-serviços) — padrões de mensageria

## Informações Necessárias

1. Tipo do evento (ex: `user.created`)
2. Serviço publicador (ex: `user-service`)
3. Serviço(s) consumidor(es)
4. Payload do evento
5. Implementando: publicação, consumo, ou ambos?

## Envelope de Evento (Obrigatório)

```json
{
  "event_id": "uuid-v4",
  "event_type": "user.created",
  "source_service": "user-service",
  "timestamp": "2026-03-05T12:00:00Z",
  "correlation_id": "uuid-v4",
  "version": "1",
  "payload": {}
}
```

Canal: `{source_service}.{event_type}` → ex: `user-service.user.created`

## Publicador

1. Definir interface `EventPublisher` em `ports/output/publisher.go`
2. Implementar em `adapters/redis/publisher.go` com envelope completo
3. Chamar no use case após operação bem-sucedida

## Consumidor

1. Criar consumer em `adapters/redis/consumer.go`
2. Implementar deduplicação via `event_id` (chave Redis com TTL 24h)
3. Propagar `correlation_id` no contexto
4. Iniciar consumer em goroutine com shutdown via context

## Checklist Final

Publicador:
- [ ] Interface `EventPublisher` em `ports/output/`
- [ ] Implementação com envelope completo
- [ ] `event_id` gerado como UUID v4
- [ ] `correlation_id` propagado do contexto
- [ ] Canal no formato `{servico}.{evento}`

Consumidor:
- [ ] Deduplicação via `event_id` (Redis TTL 24h)
- [ ] `correlation_id` propagado no contexto
- [ ] Handler idempotente
- [ ] Consumer com shutdown via context
