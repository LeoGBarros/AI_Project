---
inclusion: fileMatch
fileMatchPattern: "**/handler*.go,**/openapi.yaml,**/openapi.yml"
---

# Padrões de API REST

Referência completa: [`TECHNICAL_BASE.md` — seção 5](../../TECHNICAL_BASE.md#5-padrões-de-api-rest)

## Convenções de URL

| Regra | Correto | Incorreto |
|---|---|---|
| Recursos em plural | `/users` | `/user` |
| Minúsculas com hífen | `/user-profiles` | `/userProfiles` |
| Hierarquia com recurso pai | `/users/{id}/orders` | `/getUserOrders` |
| Sem verbos na URL | `GET /orders` | `GET /getOrders` |
| Versionamento no path | `/v1/users` | `/users?version=1` |

## Métodos HTTP e Idempotência

| Método | Uso | Idempotente | Body |
|---|---|---|---|
| `GET` | Leitura | Sim | Não |
| `POST` | Criação | Não | Sim |
| `PUT` | Substituição completa | Sim | Sim |
| `PATCH` | Atualização parcial | Não | Sim |
| `DELETE` | Remoção | Sim | Não |

## Códigos de Status HTTP

| Código | Situação |
|---|---|
| `200 OK` | Operação bem-sucedida com body |
| `201 Created` | Recurso criado — incluir `Location` header |
| `204 No Content` | Sucesso sem body (ex: DELETE) |
| `400 Bad Request` | Dados de entrada inválidos |
| `401 Unauthorized` | Token ausente ou inválido |
| `403 Forbidden` | Token válido mas sem permissão |
| `404 Not Found` | Recurso não encontrado |
| `409 Conflict` | Conflito de estado (ex: duplicata) |
| `422 Unprocessable Entity` | Validação de negócio falhou |
| `429 Too Many Requests` | Rate limit atingido |
| `500 Internal Server Error` | Erro interno inesperado |

## Envelopes de Resposta (Obrigatórios)

### Erro

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "O campo 'email' é obrigatório.",
    "details": [
      { "field": "email", "message": "campo obrigatório" }
    ],
    "trace_id": "abc123def456"
  }
}
```

### Lista paginada

```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 150,
    "total_pages": 8
  }
}
```

## Headers Obrigatórios

| Header | Direção | Descrição |
|---|---|---|
| `Authorization: Bearer <token>` | Request | Token JWT |
| `Content-Type: application/json` | Request/Response | Formato do body |
| `X-Request-ID` | Request | UUID v4 gerado pelo cliente ou Kong |
| `X-Correlation-ID` | Request/Response | ID de correlação para rastreamento distribuído |

## Health Check (Obrigatório por Serviço)

| Endpoint | Comportamento |
|---|---|
| `GET /healthz/live` | Retorna `200 {"status":"ok"}` se o processo está vivo |
| `GET /healthz/ready` | Retorna `200` se conexões com BD e cache estão OK; `503` caso contrário |
