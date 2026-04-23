---
inclusion: fileMatch
fileMatchPattern: "**/redis/**,**/cache.go,**/state_store.go,**/publisher.go"
---

# Padrões de Banco de Dados — Redis

Referência completa: [`TECHNICAL_BASE.md` — seção 7](../../TECHNICAL_BASE.md#7-padrões-de-banco-de-dados)

## Namespacing de Chaves (Obrigatório)

Formato: `{servico}:{entidade}:{id}:{campo_opcional}`

```
auth-service:pkce-state:uuid-456
auth-service:session:uuid-123
```

Nunca criar chaves sem namespace — colisões entre serviços são críticas.

## TTL (Obrigatório)

Toda chave deve ter TTL definido explicitamente. Nunca criar chaves sem expiração em produção.

| Tipo de dado | TTL recomendado |
|---|---|
| Cache de objeto de domínio | 5–15 minutos |
| Cache de lista/query | 1–5 minutos |
| Sessão de usuário | Igual ao tempo de vida do JWT |
| Rate limiting counter | Janela do rate limit (ex: 60s) |
| Token de service account | Até 80% do TTL do token |
| Estado PKCE | 5 minutos |

## Pub/Sub — Canais

- Formato: `{servico}.{evento}` (ex: `user-service.user.created`)
- Envelope obrigatório: ver [`docs/diagrams/pubsub-event-flow.md`](../../docs/diagrams/pubsub-event-flow.md)

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

- Consumidores devem ser idempotentes: processar o mesmo `event_id` múltiplas vezes deve produzir o mesmo resultado
