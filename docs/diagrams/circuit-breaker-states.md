# Circuit Breaker — Estados e Transições

> Contexto: [Seção 8.2 — Escalabilidade](../../TECHNICAL_BASE.md#82-escalabilidade)

---

## Visão Geral

O Circuit Breaker é um padrão de resiliência que protege um serviço de realizar chamadas repetidas a uma dependência que está falhando. Ele "abre o circuito" após um limiar de falhas, evitando sobrecarga e permitindo recuperação.

Threshold padrão deste projeto: **5 falhas consecutivas em 30 segundos**.

---

## Diagrama ASCII — Estados e Transições

```text
                    ┌─────────────────────────────────┐
                    │           CLOSED                 │
                    │      (operação normal)           │
                    │                                  │
                    │  Sucesso → contador zerado       │
                    │  Falha   → incrementa contador   │
                    └──────────────┬──────────────────┘
                                   │
                          5 falhas em 30s
                                   │
                                   ▼
                    ┌─────────────────────────────────┐
                    │            OPEN                  │
                    │    (circuito aberto)             │
                    │                                  │
                    │  Todas as requisições            │
                    │  rejeitadas imediatamente        │
                    └──────────────┬──────────────────┘
                                   │
                         timeout 60s expirado
                                   │
                                   ▼
                    ┌─────────────────────────────────┐
                    │         HALF-OPEN                │
                    │    (teste de recuperação)        │
                    │                                  │
                    │  Permite UMA requisição teste    │
                    └───────┬─────────────┬───────────┘
                            │             │
                   Teste OK │             │ Teste falhou
                            │             │
                            ▼             ▼
                        CLOSED          OPEN
                     (retoma normal)  (reinicia timeout)
```

## Diagrama de Estados

```mermaid
stateDiagram-v2
    [*] --> Closed

    Closed --> Open : 5 falhas consecutivas em 30s
    Open --> HalfOpen : Timeout de espera expirado (60s)
    HalfOpen --> Closed : Requisição de teste bem-sucedida
    HalfOpen --> Open : Requisição de teste falhou

    state Closed {
        [*] --> Monitorando
        Monitorando --> Monitorando : Requisição com sucesso (contador zerado)
        Monitorando --> ContandoFalhas : Requisição falhou
        ContandoFalhas --> Monitorando : Requisição com sucesso (contador zerado)
    }

    state Open {
        [*] --> Bloqueado
        Bloqueado --> Bloqueado : Todas as requisições rejeitadas imediatamente
    }

    state HalfOpen {
        [*] --> AguardandoTeste
        AguardandoTeste --> AguardandoTeste : Requisições adicionais bloqueadas
    }
```

---

## Diagrama de Sequência — Comportamento por Estado

### Estado Closed (circuito fechado — operação normal)

```mermaid
sequenceDiagram
    autonumber

    box rgb(100, 149, 237) Caller
        participant SvcA as Serviço A
    end
    box rgb(196, 164, 70) Resiliência
        participant CB as Circuit Breaker
    end
    box rgb(95, 158, 110) Dependência
        participant SvcB as Serviço B
    end

    Note over CB: Estado: CLOSED (operação normal)

    SvcA->>+CB: Solicita chamada a Serviço B
    CB->>+SvcB: Repassa chamada (circuito fechado)
    SvcB-->>-CB: 200 OK
    CB-->>-SvcA: 200 OK

    Note over CB: Sucesso. Contador de falhas: 0
```

### Estado Open (circuito aberto — falhas acima do threshold)

```mermaid
sequenceDiagram
    autonumber

    box rgb(100, 149, 237) Caller
        participant SvcA as Serviço A
    end
    box rgb(196, 164, 70) Resiliência
        participant CB as Circuit Breaker
    end
    box rgb(95, 158, 110) Dependência
        participant SvcB as Serviço B
    end

    Note over CB: Estado: OPEN (circuito aberto)<br/>Serviço B com falhas recentes.

    SvcA->>+CB: Solicita chamada a Serviço B
    Note over CB, SvcB: Serviço B NAO é contactado
    CB-->>-SvcA: Falha imediata (erro padrão ou fallback)

    Note over CB: Aguardando timeout (60s)<br/>para transitar para Half-Open.
```

### Estado Half-Open (teste de recuperação)

```mermaid
sequenceDiagram
    autonumber

    box rgb(100, 149, 237) Caller
        participant SvcA as Serviço A
    end
    box rgb(196, 164, 70) Resiliência
        participant CB as Circuit Breaker
    end
    box rgb(95, 158, 110) Dependência
        participant SvcB as Serviço B
    end

    Note over CB: Estado: HALF-OPEN<br/>Timeout expirado. Permitindo requisição de teste.

    SvcA->>CB: Solicita chamada a Serviço B
    CB->>SvcB: Permite UMA requisição de teste

    alt Teste bem-sucedido
        SvcB-->>CB: 200 OK
        CB-->>SvcA: 200 OK
        Note over CB: Transição para CLOSED.<br/>Operação normal retomada.
    else Teste falhou
        SvcB-->>CB: Erro / timeout
        CB-->>SvcA: Falha
        Note over CB: Transição para OPEN.<br/>Reinicia timeout de espera.
    end
```

---

## Parâmetros de Configuração

| Parâmetro | Valor padrão | Descrição |
|---|---|---|
| `failure_threshold` | 5 | Número de falhas consecutivas para abrir o circuito |
| `failure_window` | 30s | Janela de tempo para contabilizar falhas |
| `open_timeout` | 60s | Tempo aguardado no estado Open antes de testar Half-Open |
| `success_threshold` | 1 | Sucessos consecutivos em Half-Open para fechar o circuito |

---

## Comportamento de Fallback Recomendado

Quando o circuito está **Open**, o serviço deve retornar uma resposta degradada aceitável (fallback), nunca simplesmente propagar o erro ao cliente:

- Retornar dados do cache (se disponível no Redis)
- Retornar resposta padrão/vazia com status `503 Service Unavailable`
- Enfileirar a operação para retry posterior via Redis Pub/Sub

---

> Voltar ao índice: [README](README.md)
