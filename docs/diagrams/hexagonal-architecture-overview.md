# Arquitetura do Sistema

> Contexto: [Seção 3 — Arquitetura](../../TECHNICAL_BASE.md#3-arquitetura)

---

## Visão Macro

Diagrama completo mostrando a comunicação entre clientes, API Gateway, IAM, microsserviços, bancos de dados e observabilidade.

### Diagrama ASCII — Visão Macro

```text
┌──────────┐  ┌──────────┐  ┌──────────┐
│  Web App │  │ Mobile   │  │ Terceiros│
└────┬─────┘  └────┬─────┘  └────┬─────┘
     │             │              │
     └──────┬──────┘──────┬───────┘
            │  HTTPS       │
            ▼              ▼
     ┌──────────────────────────┐
     │      Kong API Gateway    │
     │  (JWT, Rate Limit, SSL)  │
     └──────┬───────────────────┘
            │
   ┌────────┼────────────────┐
   │        │                │
   │   ┌────┴────┐           │
   │   │Keycloak │           │
   │   │  (IAM)  │           │
   │   └─────────┘           │
   │                         │
   ▼         ▼               ▼
┌─────────┐ ┌─────────┐ ┌─────────┐
│  Svc A  │ │  Svc B  │ │  Svc N  │
│(Hexagon)│ │(Hexagon)│ │(Hexagon)│
└────┬────┘ └────┬────┘ └────┬────┘
     │           │            │
┌────┴────┐ ┌────┴────┐ ┌────┴────┐
│PostgreSQL│ │ MongoDB │ │PostgreSQL│
│ + Redis  │ │ + Redis │ │ + Redis  │
└──────────┘ └─────────┘ └──────────┘
```

### Diagrama Mermaid — Visão Macro

```mermaid
graph TD
    subgraph clients [Clientes Externos]
        Web["Web App"]
        Mobile["Mobile App"]
        ThirdParty["Terceiros / APIs"]
    end

    subgraph gateway [API Gateway]
        Kong["Kong<br/>Rate Limiting | JWT Validation | Routing | SSL"]
    end

    subgraph iam [Identity & Access Management]
        Keycloak["Keycloak<br/>Emissão de Tokens | RBAC | Scopes"]
    end

    subgraph services [Microsserviços]
        SvcA["Microsserviço A<br/>(Domain + Hexagonal)"]
        SvcB["Microsserviço B<br/>(Domain + Hexagonal)"]
        SvcN["Microsserviço N<br/>(Domain + Hexagonal)"]
    end

    subgraph data [Camada de Dados]
        PostgreSQL["PostgreSQL<br/>Dados Transacionais"]
        MongoDB["MongoDB<br/>Documentos Flexíveis"]
        Redis["Redis<br/>Cache + Pub/Sub"]
    end

    subgraph observability [Observabilidade]
        OTel["OpenTelemetry Collector"]
        Traces["Backend de Traces"]
        Metrics["Backend de Métricas"]
        Logs["Backend de Logs"]
    end

    Web -->|"HTTPS"| Kong
    Mobile -->|"HTTPS"| Kong
    ThirdParty -->|"HTTPS"| Kong

    Kong -->|"Valida JWT"| Keycloak
    Kong -->|"Route"| SvcA
    Kong -->|"Route"| SvcB
    Kong -->|"Route"| SvcN

    SvcA -->|"SQL"| PostgreSQL
    SvcA -->|"Cache"| Redis
    SvcB -->|"Documentos"| MongoDB
    SvcB -->|"Cache"| Redis
    SvcN -->|"SQL / Docs"| PostgreSQL
    SvcN -->|"Cache"| Redis

    SvcA -->|"PUBLISH evento"| Redis
    Redis -->|"SUBSCRIBE"| SvcB
    Redis -->|"SUBSCRIBE"| SvcN

    SvcA -.->|"OTLP"| OTel
    SvcB -.->|"OTLP"| OTel
    SvcN -.->|"OTLP"| OTel

    OTel --> Traces
    OTel --> Metrics
    OTel --> Logs
```

---

## Regra de Dependência — Arquitetura Hexagonal

As dependências fluem de fora para dentro. A camada de domínio é o núcleo e não conhece nada externo. As interfaces (Ports) definem contratos que os Adapters implementam.

### Diagrama ASCII — Regra de Dependência

```text
                    Mundo Externo
        ┌───────────────────────────────────┐
        │  HTTP/gRPC   DB    Cache   Queue  │
        └───────┬───────┬──────┬──────┬─────┘
                │       │      │      │
                ▼       ▼      ▼      ▼        Dependências
        ┌───────────────────────────────────┐  fluem de
        │         Adapters (concretos)      │  fora para
        │  HTTP Handler │ Repo │ Publisher  │  dentro
        └───────────────┬───────────────────┘      │
                        │ implementa               │
                        ▼                          ▼
        ┌───────────────────────────────────┐
        │      Ports (interfaces)           │
        │   Input Ports  │  Output Ports    │
        └───────────────┬───────────────────┘
                        │
                        ▼
        ┌───────────────────────────────────┐
        │     Application (Use Cases)       │
        │   Orquestra o domínio via ports   │
        └───────────────┬───────────────────┘
                        │
                        ▼
        ┌───────────────────────────────────┐
        │     Domain (Regras de Negócio)    │
        │  Entidades │ Value Objects │ Erros│
        │     *** NÃO CONHECE NADA ***     │
        │     ***    EXTERNO       ***     │
        └───────────────────────────────────┘
```

### Diagrama Mermaid — Regra de Dependência

```mermaid
graph BT
    subgraph external [Mundo Externo]
        HTTP["HTTP / gRPC"]
        DB["PostgreSQL / MongoDB"]
        Cache["Redis Cache"]
        Queue["Redis Pub/Sub"]
    end

    subgraph adapters [Adapters — Implementações Concretas]
        HTTPAdapter["HTTP Handler<br/>(chi router)"]
        DBAdapter["Repository<br/>(pgx / mongo-driver)"]
        CacheAdapter["Cache Adapter<br/>(go-redis)"]
        QueueAdapter["Publisher / Subscriber<br/>(go-redis)"]
    end

    subgraph ports [Ports — Interfaces]
        InputPort["Ports Input<br/>(Handler Interface)"]
        OutputPort["Ports Output<br/>(Repository / Publisher Interface)"]
    end

    subgraph app [Application — Use Cases]
        UseCase["Use Case / Service<br/>Orquestra o domínio"]
    end

    subgraph domain [Domain — Regras de Negócio]
        Entity["Entidades + Value Objects"]
        DomainErr["Erros de Domínio"]
    end

    HTTP --> HTTPAdapter
    DB --> DBAdapter
    Cache --> CacheAdapter
    Queue --> QueueAdapter

    HTTPAdapter -->|"implementa"| InputPort
    DBAdapter -->|"implementa"| OutputPort
    CacheAdapter -->|"implementa"| OutputPort
    QueueAdapter -->|"implementa"| OutputPort

    InputPort --> UseCase
    UseCase -->|"usa interface"| OutputPort
    UseCase --> Entity
    UseCase --> DomainErr
```

---

## Fluxo de um Request (visão ponta a ponta)

Exemplo de um request HTTP desde o cliente até a persistência no banco.

```mermaid
sequenceDiagram
    autonumber

    box rgb(100, 149, 237) Cliente
        actor Cliente as Cliente
    end
    box rgb(205, 92, 92) API Gateway
        participant Kong as Kong
    end
    box rgb(154, 165, 70) IAM
        participant Keycloak as Keycloak
    end
    box rgb(95, 158, 110) Adapter In
        participant Handler as HTTP Handler
    end
    box rgb(147, 112, 185) Application
        participant UseCase as Use Case
    end
    box rgb(80, 162, 162) Domain
        participant Domain as Domain
    end
    box rgb(175, 135, 80) Adapter Out
        participant Repo as Repository
    end
    box rgb(218, 155, 65) Banco de Dados
        participant DB as PostgreSQL
    end

    Cliente->>+Kong: GET /v1/users/123 (Bearer JWT)
    Kong->>+Keycloak: Valida JWT (JWKS em cache)
    Keycloak-->>-Kong: OK (token válido)
    Kong->>+Handler: Forward + headers de contexto

    Handler->>+UseCase: GetUser(ctx, userID)
    UseCase->>+Domain: Valida regras de negócio
    Domain-->>-UseCase: OK

    UseCase->>+Repo: FindByID(ctx, userID)
    Repo->>+DB: SELECT * FROM users WHERE id = $1
    DB-->>-Repo: Row
    Repo-->>-UseCase: User entity

    UseCase-->>-Handler: User DTO
    Handler-->>-Kong: 200 OK { user }
    Kong-->>-Cliente: 200 OK { user }
```

---

> Voltar ao índice: [README](README.md)
