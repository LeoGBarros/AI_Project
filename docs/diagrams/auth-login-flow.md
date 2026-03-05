# Fluxo de Autenticação do Usuário

> Contexto: [Seção 4.1 — Autenticação e Autorização](../../TECHNICAL_BASE.md#4-autenticação-e-autorização)

---

## 4.1.a — Login (Obtenção do Token)

O usuário autentica-se diretamente no Keycloak. O Kong **não** participa do fluxo de login — ele apenas valida o token nos requests subsequentes.

```mermaid
sequenceDiagram
    autonumber

    box rgb(100, 149, 237) Cliente
        actor Usuario as Usuário
    end
    box rgb(154, 165, 70) IAM
        participant Keycloak as Keycloak
    end

    Usuario->>Keycloak: POST /realms/{realm}/protocol/openid-connect/token
    Note right of Keycloak: grant_type=password<br/>client_id, username, password
    Keycloak->>Keycloak: Valida credenciais e verifica realm

    alt Credenciais inválidas
        Keycloak-->>Usuario: 401 Unauthorized { error: invalid_grant }
    else Credenciais válidas
        Keycloak-->>Usuario: 200 OK
        Note left of Keycloak: access_token (JWT, curta duração)<br/>refresh_token (longa duração)<br/>expires_in (segundos)
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

---

> Próximo: [Renovação de Token](auth-token-refresh.md)
> Voltar ao índice: [README](README.md)
