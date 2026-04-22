# Fluxo de Eventos Pub/Sub (Redis)

> Contexto: [Seção 3.4 — Comunicação entre Serviços](../../TECHNICAL_BASE.md#34-comunicação-entre-serviços)

---

## Visão Geral

Comunicação assíncrona entre serviços usa **Redis Pub/Sub**. O Publisher publica um evento em um canal nomeado. Todos os Subscribers inscritos naquele canal recebem o evento de forma independente.

Regras obrigatórias:
- Todo evento deve seguir o **envelope padrão** (ver `TECHNICAL_BASE.md` seção 3.4)
- Consumidores devem ser **idempotentes**: o mesmo `event_id` processado mais de uma vez deve produzir o mesmo resultado
- O canal segue o padrão: `{servico}.{entidade}.{acao}` (ex: `user-service.user.created`)

---

## Diagrama ASCII — Fluxo Pub/Sub

```text
┌─────────────┐         ┌──────────────────┐         ┌─────────────┐
│  Serviço A  │         │  Redis Pub/Sub   │         │  Serviço B  │
│ (Publisher) │         │                  │         │(Subscriber) │
└──────┬──────┘         │  Canal:          │         └──────▲──────┘
       │                │  user-service.   │                │
       │  PUBLISH       │  user.created    │   Mensagem     │
       │  ─────────────►│                  │───────────────►│
       │                │  ┌────────────┐  │                │
       │   Envelope:    │  │  Entrega   │  │         ┌──────▲──────┐
       │   {event_id,   │  │ simultânea │  │         │  Serviço C  │
       │    event_type,  │  │  a todos   │  │         │(Subscriber) │
       │    source,      │  │subscribers │  │         └─────────────┘
       │    timestamp,   │  └────────────┘  │                │
       │    correlation, │                  │   Mensagem     │
       │    payload}     │                  │───────────────►│
       │                └──────────────────┘                │
       │                                                    │
       │                                                    │
       │            Cada subscriber verifica:               │
       │            ┌──────────────────────────┐            │
       │            │ event_id já processado?  │            │
       │            │  SIM → descarta (WARN)   │            │
       │            │  NÃO → processa + registra│           │
       │            └──────────────────────────┘            │
```

## Diagrama de Sequência — Publicação e Consumo

```mermaid
sequenceDiagram
    autonumber

    box rgb(100, 149, 237) Publisher
        participant SvcA as Serviço A
    end
    box rgb(210, 105, 55) Mensageria
        participant Redis as Redis Pub/Sub
    end
    box rgb(95, 158, 110) Subscriber
        participant SvcB as Serviço B
    end
    box rgb(147, 112, 185) Subscriber
        participant SvcC as Serviço C
    end

    Note over SvcA: Evento de domínio gerado<br/>após operação bem-sucedida

    SvcA->>+Redis: PUBLISH user-service.user.created
    Note right of SvcA: envelope: {<br/>  event_id, event_type,<br/>  source_service, timestamp,<br/>  correlation_id, payload<br/>}

    par Entrega simultânea aos subscribers
        Redis-->>SvcB: Mensagem entregue
    and
        Redis-->>SvcC: Mensagem entregue
    end
    deactivate Redis

    SvcB->>SvcB: Extrai event_id do envelope
    SvcB->>SvcB: Verifica idempotência (event_id já processado?)

    alt event_id já processado
        SvcB-->>SvcB: Descarta evento (log de WARN)
    else event_id novo
        SvcB->>SvcB: Processa lógica de negócio
        SvcB->>SvcB: Registra event_id como processado
        Note over SvcB: Ex: tabela processed_events<br/>ou chave Redis com TTL
    end

    Note over SvcC: Processamento independente<br/>(mesmo fluxo de idempotência)
```

---

## Diagrama de Sequência — Falha no Consumidor

Redis Pub/Sub **não garante entrega** se o subscriber estiver offline. Para cenários que exigem garantia de entrega, avaliar o uso de **Redis Streams** como alternativa.

```mermaid
sequenceDiagram
    autonumber

    box rgb(100, 149, 237) Publisher
        participant SvcA as Serviço A
    end
    box rgb(210, 105, 55) Mensageria
        participant Redis as Redis Pub/Sub
    end
    box rgb(95, 158, 110) Subscriber
        participant SvcB as Serviço B
    end

    Note over SvcB: Serviço B offline / reiniciando

    SvcA->>+Redis: PUBLISH user-service.user.created { event_id }
    Note over Redis: Nenhum subscriber ativo no canal.<br/>Mensagem descartada pelo Redis.
    deactivate Redis

    Note over SvcB: Serviço B volta online
    SvcB->>Redis: SUBSCRIBE user-service.user.created
    Note over SvcB: Mensagem anterior NAO é recebida.<br/>Apenas mensagens futuras são entregues.
```

> **Atenção:** Para garantia de entrega (at-least-once), use **Redis Streams** com consumer groups em vez de Pub/Sub puro.

---

## Namespacing de Canais

| Padrão | Exemplo |
|---|---|
| `{servico}.{entidade}.{acao}` | `user-service.user.created` |
| `{servico}.{entidade}.{acao}` | `order-service.order.status_changed` |
| `{servico}.{entidade}.{acao}` | `payment-service.payment.confirmed` |

---

> Voltar ao índice: [README](README.md)
