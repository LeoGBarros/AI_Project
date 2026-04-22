# Auth Service Local Dev

Este guia sobe a stack local do `auth-service` com `Redis` e `Keycloak` e serve como referência para validar login, refresh, logout e o fluxo `PKCE` web.

## 0. Subir todo o ambiente

Na raiz do repositório (é necessário ter o Docker em execução e permissão para usar o daemon):

```bash
docker compose up --build
```

Para rodar em segundo plano:

```bash
docker compose up --build -d
```

Aguarde até os três serviços estarem saudáveis. Em seguida use o **painel do Keycloak** e a **chamada de login** abaixo.

---

## 1. Subir a stack (referência)

Na raiz do repositório:

```bash
docker compose up --build
```

Serviços expostos:

- `auth-service`: `http://localhost:8082`
- `Keycloak`: `http://localhost:8081`
- `Redis`: `localhost:6379`

O `auth-service` já expõe:

- `GET /healthz/live`
- `GET /healthz/ready`
- `POST /v1/auth/login`
- `GET /v1/auth/authorize`
- `GET /v1/auth/callback`
- `POST /v1/auth/refresh`
- `POST /v1/auth/logout`

## 2. Painel do Keycloak

| Item | Valor |
|------|--------|
| **URL** | **http://localhost:8081** |
| **Usuário** | `admin` |
| **Senha** | `admin` |

1. Abra no navegador: **http://localhost:8081**
2. Clique em **Administration Console**
3. Faça login com `admin` / `admin`
4. Crie o realm e os clients conforme as seções abaixo (se ainda não existirem)
5. Crie um **usuário de teste** no realm `AI-Project` (Users → Add user → username/senha; em Credentials defina a senha e desmarque "Temporary") para usar na chamada de login

## 3. Criar o realm

Crie um realm chamado `AI-Project`.

## 4. Criar os clients

Crie estes clients no realm `AI-Project` (Clients → Create client):

- `web-app`
- `mobile-app`
- `backend-app`

Para **mobile-app** e **backend-app** (usados no login ROPC): em *Capability config*, habilite **Direct access grants** (Resource Owner Password Credentials). Para **web-app**, use o fluxo Authorization Code com PKCE (não habilite Direct access grants).

Regras práticas do ambiente atual:

- `web-app` usa o fluxo `PKCE`.
- `mobile-app` e `backend-app` usam o fluxo `ROPC` quando aplicável.
- O estado temporário do fluxo `PKCE` fica no `Redis`, então a readiness depende dele.

Para o client `web-app`, copie o secret gerado e substitua o valor de:

```env
AUTH_SERVICE_KEYCLOAK_CLIENT_SECRET_WEB=...
```

## 5. Configurar o auth-service

As variáveis já estão no `docker-compose.yml`, mas você pode ajustar se necessário:

```env
AUTH_SERVICE_KEYCLOAK_BASE_URL=http://keycloak:8080
AUTH_SERVICE_KEYCLOAK_REALM=AI-Project
AUTH_SERVICE_KEYCLOAK_CLIENT_ID_WEB=web-app
AUTH_SERVICE_KEYCLOAK_CLIENT_SECRET_WEB=secret-web
AUTH_SERVICE_KEYCLOAK_CLIENT_ID_MOBILE=mobile-app
AUTH_SERVICE_KEYCLOAK_CLIENT_ID_APP=backend-app
AUTH_SERVICE_REDIS_ADDR=redis:6379
```

Essas variáveis refletem o wiring atual do serviço:

- `Keycloak` roda no container `keycloak` e é acessado internamente em `http://keycloak:8080`.
- `Redis` roda no container `redis` e é acessado em `redis:6379`.
- O `auth-service` sobe na porta `8080` dentro do container e é publicado na porta `8082` do host.

## 6. Validar a API

Quando a stack estiver pronta, teste:

- `GET http://localhost:8082/healthz/live`
- `GET http://localhost:8082/healthz/ready`

Se o `ready` falhar, verifique primeiro a conectividade com `Redis`.

## 7. Obter um token válido (chamada de login)

Com o realm `AI-Project`, os clients (`mobile-app`, `backend-app`) e um **usuário de teste** já criados no Keycloak, use o `auth-service` para obter um token:

```bash
curl -s -X POST http://localhost:8082/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "SEU_USUARIO_KEYCLOAK",
    "password": "SUA_SENHA",
    "client_type": "mobile"
  }'
```

Exemplo com usuário `testuser` e senha `test123`:

```bash
curl -s -X POST http://localhost:8082/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123","client_type":"mobile"}'
```

Resposta esperada (200 OK):

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI...",
  "token_type": "Bearer",
  "expires_in": 300,
  "refresh_expires_in": 1800
}
```

Use o `access_token` no header `Authorization: Bearer <access_token>` em chamadas a APIs protegidas. Para renovar, use `POST /v1/auth/refresh` com o `refresh_token`.

## 8. Login (referência Keycloak direta)

O fluxo de login usa o endpoint padrão do Keycloak:

```bash
POST /realms/AI-Project/protocol/openid-connect/token
```

O `auth-service` documenta os fluxos suportados em `api/openapi.yaml` e os casos principais são:

- `login` para `mobile` e `app`
- `authorize` para iniciar o `PKCE` web
- `callback` para concluir o `PKCE`
- `refresh` para renovar tokens
- `logout` para revogar o `refresh_token`

Se quiser, depois eu também posso te entregar um `realm export` pronto com:

- realm `AI-Project`
- clients
- roles
- um usuário de teste

