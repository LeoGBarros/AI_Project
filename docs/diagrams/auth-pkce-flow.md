# Fluxo PKCE — Authorization Code + PKCE

> Contexto: [Seção 4 — Autenticação e Autorização](../../TECHNICAL_BASE.md#4-autenticação-e-autorização)

---

## Visão Geral

Fluxo usado por clientes web (browser/SPA). O `auth-service` gera o `code_verifier`/`code_challenge`, armazena o estado em Redis (TTL 5 min), e troca o authorization code por tokens no callback — sem expor `client_secret` ao browser.

## Diagrama ASCII

```text
Browser           auth-service         Redis          Keycloak
   │                   │                 │                │
   │  GET /authorize   │                 │                │
   │──────────────────>│                 │                │
   │                   │  Gera state +   │                │
   │                   │  code_verifier  │                │
   │                   │  + challenge    │                │
   │                   │─── Salva state ─>                │
   │                   │                 │                │
   │  302 Redirect     │                 │                │
   │<──────────────────│                 │                │
   │                   │                 │                │
   │  Login no Keycloak (com code_challenge)             │
   │─────────────────────────────────────────────────────>│
   │                   │                 │                │
   │  302 → /callback?code=xxx&state=yyy                 │
   │<────────────────────────────────────────────────────│
   │                   │                 │                │
   │  GET /callback    │                 │                │
   │──────────────────>│                 │                │
   │                   │── Busca state ──>                │
   │                   │<────────────────│                │
   │                   │                 │                │
   │                   │  POST /token (code + verifier)  │
   │                   │─────────────────────────────────>│
   │                   │<────────────────────────────────│
   │                   │                 │                │
   │                   │── Deleta state ─>                │
   │  200 OK (tokens)  │                 │                │
   │<──────────────────│                 │                │
```

## Diagrama Mermaid

```mermaid
sequenceDiagram
    autonumber

    actor Cliente as Cliente Web (SPA)
    participant AuthSvc as auth-service
    participant Redis as Redis
    participant Keycloak as Keycloak

    Cliente->>AuthSvc: GET /v1/auth/authorize?redirect_uri=...&client_id=...

    AuthSvc->>AuthSvc: Gera state, code_verifier, code_challenge (S256)
    AuthSvc->>Redis: Salva PKCEState (TTL 5 min)

    AuthSvc-->>Cliente: 302 → Keycloak /auth?code_challenge=...&state=...

    Cliente->>Keycloak: Login (formulário)
    Keycloak-->>Cliente: 302 → /callback?code=AUTH_CODE&state=STATE

    Cliente->>AuthSvc: GET /v1/auth/callback?code=...&state=...

    AuthSvc->>Redis: Busca PKCEState por state
    Redis-->>AuthSvc: code_verifier, redirect_uri

    AuthSvc->>Keycloak: POST /token (code + code_verifier)

    alt Inválido
        Keycloak-->>AuthSvc: 400 invalid_grant
        AuthSvc-->>Cliente: 401 Unauthorized
    else Válido
        Keycloak-->>AuthSvc: 200 OK (access_token, refresh_token)
        AuthSvc->>Redis: Deleta PKCEState
        AuthSvc-->>Cliente: 200 OK (tokens)
    end
```

## Parâmetros

| Parâmetro | Valor | Descrição |
|---|---|---|
| `code_verifier` | 64 bytes, URL-safe base64 | Segredo temporário gerado pelo auth-service |
| `code_challenge` | SHA-256(code_verifier) | Hash enviado ao Keycloak |
| `code_challenge_method` | `S256` | Método de derivação |
| `state` | 32 bytes, URL-safe base64 | Proteção CSRF |
| `PKCEState TTL` | 5 minutos | Tempo máximo entre /authorize e /callback |
| `grant_type` | `authorization_code` | Tipo de grant na troca |

---

> Anterior: [Login ROPC (mobile/app)](auth-ropc-login-flow.md)
> Próximo: [Renovação de Token](auth-token-refresh-flow.md)
> Voltar ao índice: [README](README.md)
