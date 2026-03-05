# Fluxo de Renovação de Token (Refresh)

> Contexto: [Seção 4.2 — Token JWT](../../TECHNICAL_BASE.md#42-token-jwt)

---

## Visão Geral

O `access_token` JWT tem vida curta (tipicamente 5–15 minutos). Quando ele expira, o cliente usa o `refresh_token` para obter um novo par de tokens sem exigir que o usuário refaça o login.

O `refresh_token` tem vida mais longa (horas ou dias) e é configurado no realm do Keycloak.

---

## Diagrama de Sequência

```mermaid
sequenceDiagram
    autonumber

    box rgb(100, 149, 237) Cliente
        actor Usuario as Usuário / Cliente
    end
    box rgb(205, 92, 92) API Gateway
        participant Kong as Kong
    end
    box rgb(154, 165, 70) IAM
        participant Keycloak as Keycloak
    end

    Note over Usuario: access_token expirado detectado<br/>(resposta 401 ou verificação local de exp)

    Usuario->>+Keycloak: POST /realms/{realm}/protocol/openid-connect/token
    Note right of Keycloak: grant_type=refresh_token<br/>client_id<br/>refresh_token=‹valor›

    alt Refresh token válido e não expirado
        Keycloak->>Keycloak: Invalida refresh_token anterior (rotação)
        Keycloak-->>Usuario: 200 OK
        Note left of Keycloak: novo access_token<br/>novo refresh_token<br/>expires_in

        Usuario->>Kong: Repete request original com novo access_token
        Kong-->>-Usuario: 200 OK { data }

    else Refresh token expirado ou inválido
        Keycloak-->>Usuario: 400 Bad Request { error: invalid_grant }
        Note over Usuario: Sessão encerrada.<br/>Usuário deve fazer login novamente.

        Usuario->>Keycloak: POST /token (grant_type=password)
        Keycloak-->>Usuario: 200 OK (novos tokens)
    end
```

---

## Configurações Relevantes no Keycloak

| Parâmetro | Valor recomendado | Descrição |
|---|---|---|
| `Access Token Lifespan` | 5–15 min | Tempo de vida do access_token |
| `SSO Session Idle` | 30 min | Refresh token expira se não usado |
| `SSO Session Max` | 8–24h | Duração máxima absoluta da sessão |
| `Refresh Token Rotation` | Habilitado | Cada uso do refresh_token emite um novo |

---

> Anterior: [Login do Usuário](auth-login-flow.md)
> Próximo: [Service-to-Service](auth-service-to-service.md)
> Voltar ao índice: [README](README.md)
