# Fluxo ROPC — Resource Owner Password Credentials

> Contexto: [Seção 4 — Autenticação e Autorização](../../TECHNICAL_BASE.md#4-autenticação-e-autorização)

---

## Visão Geral

O fluxo ROPC (`grant_type=password`) é utilizado por clientes mobile (iOS/Android) e desktop/CLI (app). O cliente envia username e password diretamente ao `auth-service`, que repassa ao Keycloak via `POST /token`. Clientes mobile e app usam `client_id` público (sem `client_secret`). O Kong **não** participa do fluxo de login — ele apenas valida o token nos requests subsequentes.

## Diagrama ASCII

```text
┌──────────┐                  ┌──────────────┐                  ┌──────────┐
│  Cliente  │                  │ auth-service │                  │ Keycloak │
│(mobile/  │                  │              │                  │          │
│  app)    │                  │              │                  │          │
└────┬─────┘                  └──────┬───────┘                  └────┬─────┘
     │                               │                               │
     │  1. POST /login               │                               │
     │   { username, password,       │                               │
     │     client_type: "mobile" }   │                               │
     │──────────────────────────────►│                               │
     │                               │                               │
     │                               │  2. Valida client_type        │
     │                               │     (deve ser mobile ou app)  │
     │                               │                               │
     │                               │  3. Resolve client_id         │
     │                               │     por client_type           │
     │                               │     (sem client_secret)       │
     │                               │                               │
     │                               │  4. POST /token               │
     │                               │   grant_type=password         │
     │                               │   client_id=...               │
     │                               │   username=...                │
     │                               │   password=...                │
     │                               │   scope=openid                │
     │                               │──────────────────────────────►│
     │                               │                               │
     │                               │  5. Keycloak valida           │
     │                               │     credenciais e realm       │
     │                               │                               │
     │                               │  [Credenciais inválidas]      │
     │                               │◄──────────────────────────────│
     │  401 { error: invalid_grant } │   400 { error: invalid_grant }│
     │◄──────────────────────────────│                               │
     │                               │                               │
     │                               │  [Credenciais válidas]        │
     │                               │◄──────────────────────────────│
     │  200 OK                       │   200 { access_token,         │
     │  { access_token,              │         refresh_token }       │
     │    refresh_token,             │                               │
     │    expires_in }               │                               │
     │◄──────────────────────────────│                               │
     │                               │                               │
```

## 4.1.a — Login (Obtenção do Token)

```mermaid
sequenceDiagram
    autonumber

    box rgb(100, 149, 237) Cliente
        actor Usuario as Usuário (mobile/app)
    end
    box rgb(95, 158, 110) Serviço
        participant AuthSvc as auth-service
    end
    box rgb(154, 165, 70) IAM
        participant Keycloak as Keycloak
    end

    Usuario->>AuthSvc: POST /v1/auth/login { username, password, client_type }

    AuthSvc->>AuthSvc: Valida client_type (deve ser "mobile" ou "app")

    alt client_type inválido (ex: "web")
        AuthSvc-->>Usuario: 400 { error: INVALID_CLIENT_TYPE, message: "must use PKCE flow" }
    end

    AuthSvc->>AuthSvc: Resolve client_id por client_type (sem client_secret para mobile/app)

    AuthSvc->>Keycloak: POST /realms/{realm}/protocol/openid-connect/token
    Note right of Keycloak: grant_type=password<br/>client_id<br/>username, password<br/>scope=openid

    Keycloak->>Keycloak: Valida credenciais e verifica realm

    alt Credenciais inválidas
        Keycloak-->>AuthSvc: 400 { error: invalid_grant }
        AuthSvc-->>Usuario: 401 Unauthorized { error: INVALID_CREDENTIALS }
    else Credenciais válidas
        Keycloak-->>AuthSvc: 200 OK { access_token, refresh_token, expires_in }
        AuthSvc-->>Usuario: 200 OK { access_token, refresh_token, token_type, expires_in }
    end
```

---

## 4.1.b — Request Autenticado

Com o `access_token` obtido, o usuário faz requests à API. O Kong valida o JWT antes de rotear ao microsserviço destino.

```mermaid
sequenceDiagram
    autonumber

    box rgb(100, 149, 237) Cliente
        actor Usuario as Usuário
    end
    box rgb(205, 92, 92) API Gateway
        participant Kong as Kong
    end
    box rgb(154, 165, 70) IAM
        participant Keycloak as Keycloak
    end
    box rgb(95, 158, 110) Serviço
        participant Servico as Microsserviço
    end

    Usuario->>Kong: GET /v1/resource (Authorization: Bearer JWT)

    Kong->>Keycloak: GET /realms/{realm}/protocol/openid-connect/certs
    Note right of Kong: Busca JWKS (chave pública).<br/>Resultado em cache por TTL configurado.
    Keycloak-->>Kong: 200 OK { keys: [...] }
    Kong->>Kong: Valida assinatura, exp, iss, aud

    alt Token ausente
        Kong-->>Usuario: 401 Unauthorized { message: no token provided }
    else Token inválido ou expirado
        Kong-->>Usuario: 401 Unauthorized { message: invalid or expired token }
    else Token válido
        Kong->>Servico: Forward request
        Note right of Kong: Adiciona headers:<br/>X-Consumer-ID<br/>X-Consumer-Username<br/>X-Correlation-ID

        Servico->>Servico: Extrai e verifica roles/scopes do JWT

        alt Permissão insuficiente
            Servico-->>Kong: 403 Forbidden
            Kong-->>Usuario: 403 Forbidden { error: insufficient_permissions }
        else Autorizado
            Servico->>Servico: Processa a requisição
            Servico-->>Kong: 200 OK { data }
            Kong-->>Usuario: 200 OK { data }
        end
    end
```

## Parâmetros / Configuração

| Parâmetro | Valor | Descrição |
|---|---|---|
| `grant_type` | `password` | Tipo de grant ROPC |
| `client_type` | `mobile` ou `app` | Tipo de cliente (determina o `client_id`) |
| `client_id` | Configurado por tipo | Mobile e app usam client_id público (sem secret) |
| `client_secret` | Vazio para mobile/app | Clientes públicos não possuem secret |
| `scope` | `openid` | Scope padrão solicitado ao Keycloak |

---

> Anterior: [Fluxo PKCE (web)](auth-pkce-flow.md)
> Próximo: [Renovação de Token](auth-token-refresh-flow.md)
> Voltar ao índice: [README](README.md)
