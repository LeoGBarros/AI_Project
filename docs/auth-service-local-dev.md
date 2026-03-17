# Auth Service Local Dev

Este guia sobe a stack local do `auth-service` com `Redis` e `Keycloak` e serve como referência para validar login, refresh, logout e o fluxo `PKCE` web.

## 1. Subir a stack

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

## 2. Entrar no Keycloak

Acesse `http://localhost:8081` e use:

- usuário: `admin`
- senha: `admin`

## 3. Criar o realm

Crie um realm chamado `AI-Project`.

## 4. Criar os clients

Crie estes clients no realm `AI-Project`:

- `web-app`
- `mobile-app`
- `backend-app`

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

## 7. Login

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

