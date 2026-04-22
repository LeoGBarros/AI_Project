# Fluxo PKCE — Authorization Code + PKCE

> Contexto: [Seção 4 — Autenticação e Autorização](../../TECHNICAL_BASE.md#4-autenticação-e-autorização)

---

## Visão Geral

O fluxo Authorization Code + PKCE é utilizado por clientes web (browser/SPA). O `auth-service` atua como intermediário: gera o `code_verifier` e `code_challenge` no backend, armazena o estado em Redis com TTL de 5 minutos, e troca o authorization code por tokens no callback — sem expor `client_secret` ao browser.

## Diagrama ASCII

```text
┌──────────┐                  ┌──────────────┐          ┌───────────┐          ┌──────────┐
│  Cliente  │                  │ auth-service │          │   Redis   │          │ Keycloak │
│  (Web)   │                  │              │          │           │          │          │
└────┬─────┘                  └──────┬───────┘          └─────┬─────┘          └────┬─────┘
     │                               │                        │                     │
     │  1. GET /authorize            │                        │                     │
     │   ?redirect_uri=...           │                        │                     │
     │   &client_id=...             │                        │                     │
     │──────────────────────────────►│                        │                     │
     │                               │                        │                     │
     │                               │  2. Gera state_id,     │                     │
     │                               │     code_verifier,     │                     │
     │                               │     code_challenge     │                     │
     │                               │     (SHA-256 / S256)   │                     │
     │                               │                        │                     │
     │                               │  3. Salva PKCEState    │                     │
     │                               │     (TTL 5 min)        │                     │
     │                               │───────────────────────►│                     │
     │                               │                        │                     │
     │  4. HTTP 302 Redirect         │                        │                     │
     │   → Keycloak /auth            │                        │                     │
     │   ?code_challenge=...         │                        │                     │
     │   &state=...                  │                        │                     │
     │◄──────────────────────────────│                        │                     │
     │                               │                        │                     │
     │  5. Usuário autentica         │                        │                     │
     │─────────────────────────────────────────────────────────────────────────────►│
     │                               │                        │                     │
     │  6. Keycloak redireciona      │                        │                     │
     │   → /callback?code=...&state=...                       │                     │
     │◄────────────────────────────────────────────────────────────────────────────│
     │                               │                        │                     │
     │  7. GET /callback             │                        │                     │
     │   ?code=...&state=...         │                        │                     │
     │──────────────────────────────►│                        │                     │
     │                               │                        │                     │
     │                               │  8. Busca PKCEState    │                     │
     │                               │     por state_id       │                     │
     │                               │───────────────────────►│                     │
     │                               │◄───────────────────────│                     │
     │                               │                        │                     │
     │                               │  9. Verifica expiração │                     │
     │                               │     do state           │                     │
     │                               │                        │                     │
     │                               │  10. POST /token       │                     │
     │                               │   grant_type=          │                     │
     │                               │    authorization_code  │                     │
     │                               │   code=...             │                     │
     │                               │   code_verifier=...    │                     │
     │                               │   redirect_uri=...     │                     │
     │                               │───────────────────────────────────────────►│
     │                               │                        │                     │
     │                               │  11. Keycloak valida   │                     │
     │                               │   code_verifier vs     │                     │
     │                               │   code_challenge       │                     │
     │                               │◄───────────────────────────────────────────│
     │                               │                        │                     │
     │                               │  12. Deleta PKCEState  │                     │
     │                               │   (previne replay)     │                     │
     │                               │───────────────────────►│                     │
     │                               │                        │                     │
     │  13. 200 OK                   │                        │                     │
     │   { access_token,             │                        │                     │
     │     refresh_token }           │                        │                     │
     │◄──────────────────────────────│                        │                     │
     │                               │                        │                     │
```

## Diagrama Mermaid

```mermaid
sequenceDiagram
    autonumber

    box rgb(100, 149, 237) Cliente
        actor Cliente as Cliente Web (SPA)
    end
    box rgb(95, 158, 110) Serviço
        participant AuthSvc as auth-service
    end
    box rgb(180, 120, 60) Cache
        participant Redis as Redis
    end
    box rgb(154, 165, 70) IAM
        participant Keycloak as Keycloak
    end

    Cliente->>AuthSvc: GET /v1/auth/authorize?redirect_uri=...&client_id=...

    AuthSvc->>AuthSvc: Gera state_id (32 bytes, URL-safe base64)
    AuthSvc->>AuthSvc: Gera code_verifier (64 bytes, URL-safe base64)
    AuthSvc->>AuthSvc: Deriva code_challenge = SHA-256(code_verifier)

    AuthSvc->>Redis: Salva PKCEState { state_id, code_verifier, redirect_uri, client_id, expires_at }
    Note right of Redis: Chave: auth-service:pkce-state:{state_id}<br/>TTL: 5 minutos

    AuthSvc-->>Cliente: HTTP 302 → Keycloak /auth?response_type=code&client_id=...&redirect_uri=...&state=...&code_challenge=...&code_challenge_method=S256&scope=openid

    Cliente->>Keycloak: Usuário autentica (login form)
    Keycloak-->>Cliente: HTTP 302 → /v1/auth/callback?code=AUTH_CODE&state=STATE_ID

    Cliente->>AuthSvc: GET /v1/auth/callback?code=AUTH_CODE&state=STATE_ID

    AuthSvc->>Redis: Busca PKCEState por state_id
    Redis-->>AuthSvc: PKCEState { code_verifier, redirect_uri, client_id }

    AuthSvc->>AuthSvc: Verifica se state não expirou

    AuthSvc->>Keycloak: POST /realms/{realm}/protocol/openid-connect/token
    Note right of Keycloak: grant_type=authorization_code<br/>client_id, client_secret<br/>code=AUTH_CODE<br/>redirect_uri=...<br/>code_verifier=...

    Keycloak->>Keycloak: Valida code_verifier vs code_challenge (S256)

    alt Código ou verifier inválido
        Keycloak-->>AuthSvc: 400 { error: invalid_grant }
        AuthSvc-->>Cliente: 401 Unauthorized
    else Válido
        Keycloak-->>AuthSvc: 200 OK { access_token, refresh_token, expires_in }
        AuthSvc->>Redis: Deleta PKCEState (previne replay)
        AuthSvc-->>Cliente: 200 OK { access_token, refresh_token, token_type, expires_in }
    end
```

## Parâmetros / Configuração

| Parâmetro | Valor | Descrição |
|---|---|---|
| `code_verifier` | 64 bytes, URL-safe base64 | String aleatória gerada pelo auth-service |
| `code_challenge` | SHA-256(code_verifier) | Hash enviado ao Keycloak no `/auth` |
| `code_challenge_method` | `S256` | Método de derivação do challenge |
| `state` | 32 bytes, URL-safe base64 | Proteção CSRF, vincula `/authorize` ao `/callback` |
| `PKCEState TTL` | 5 minutos | Tempo máximo entre `/authorize` e `/callback` |
| `grant_type` | `authorization_code` | Tipo de grant usado na troca de código |
| `scope` | `openid` | Scope padrão solicitado ao Keycloak |
| `Redis key` | `auth-service:pkce-state:{state_id}` | Chave de armazenamento do estado PKCE |

---

> Anterior: [Login ROPC (mobile/app)](auth-ropc-login-flow.md)
> Próximo: [Renovação de Token](auth-token-refresh-flow.md)
> Voltar ao índice: [README](README.md)
