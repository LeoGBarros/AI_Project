# Autenticação Service-to-Service — Client Credentials

> Contexto: [Seção 4.4 — Service-to-Service](../../TECHNICAL_BASE.md#44-service-to-service)

---

## Visão Geral

Comunicação entre serviços backend usa **OAuth 2.0 Client Credentials**. Não há usuário envolvido — cada serviço se autentica com seu próprio `client_id` e `client_secret` no Keycloak. O token é reutilizado até próximo da expiração.

## Diagrama ASCII

```text
Serviço A              Keycloak           Kong            Serviço B
    │                      │                │                  │
    │  POST /token         │                │                  │
    │  grant_type=         │                │                  │
    │  client_credentials  │                │                  │
    │  {client_id, secret} │                │                  │
    │─────────────────────>│                │                  │
    │                      │                │                  │
    │  200 OK (JWT)        │                │                  │
    │  expires_in: 300s    │                │                  │
    │<─────────────────────│                │                  │
    │                      │                │                  │
    │  Cache token local   │                │                  │
    │                      │                │                  │
    │  GET /v1/resource (Bearer JWT)        │                  │
    │──────────────────────────────────────>│                  │
    │                      │                │  Valida JWT      │
    │                      │                │  Forward request │
    │                      │                │─────────────────>│
    │                      │                │                  │
    │  200 OK { data }     │                │  200 OK { data } │
    │<─────────────────────────────────────│<─────────────────│
    │                      │                │                  │
```

## Diagrama Mermaid

```mermaid
sequenceDiagram
    autonumber

    participant SvcA as Serviço A
    participant Keycloak as Keycloak
    participant Kong as Kong
    participant SvcB as Serviço B

    Note over SvcA: Token ausente ou expirado

    SvcA->>Keycloak: POST /token (grant_type=client_credentials, client_id, client_secret)

    alt Client inválido
        Keycloak-->>SvcA: 401 unauthorized_client
    else Autenticado
        Keycloak-->>SvcA: 200 OK (access_token, expires_in: 300s)
        SvcA->>SvcA: Cache token (TTL = expires_in - 30s)
    end

    SvcA->>Kong: GET /v1/resource (Bearer JWT)
    Kong->>Kong: Valida JWT (assinatura, exp, iss)
    Kong->>SvcB: Forward request

    SvcB-->>Kong: 200 OK { data }
    Kong-->>SvcA: 200 OK { data }
```

## Parâmetros

| Parâmetro | Valor | Descrição |
|---|---|---|
| `grant_type` | `client_credentials` | Tipo de grant S2S |
| `client_id` | Exclusivo por serviço | Cada serviço tem seu próprio client_id |
| `client_secret` | Configurado no Keycloak | Secret do service account |
| `TTL do token` | 5 minutos | Vida máxima do token S2S |
| `Cache buffer` | 30 segundos | Renova antes de expirar |

---

> Anterior: [Refresh de Token](auth-token-refresh-flow.md)
> Voltar ao índice: [README](README.md)
