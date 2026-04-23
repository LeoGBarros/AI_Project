# Guia de Agent Hooks — Kiro

Este guia documenta os hooks configurados no projeto, como foram criados, como usar e como verificar se estão funcionando.

---

## O que são Hooks?

Hooks são automações que disparam ações do agente quando eventos específicos acontecem no IDE. Ficam em `.kiro/hooks/` como arquivos `.kiro.hook` (JSON).

Cada hook tem:
- Um evento que o dispara (quando)
- Uma ação que ele executa (o quê)

---

## Hooks do Projeto

### 1. Verificação pré-escrita

| Campo | Valor |
|---|---|
| Arquivo | `.kiro/hooks/pre-write-check.kiro.hook` |
| Evento | `preToolUse` (antes de escrever arquivos) |
| Tipo de ação | `askAgent` |
| Acionamento | Automático — roda toda vez que o agente vai escrever/modificar um arquivo |

O que verifica:
- Regra de dependência entre camadas (domain ← application ← adapters)
- Error handling com `errors.Is`/`errors.As` e `fmt.Errorf` com `%w`
- `context.Context` como primeiro parâmetro em funções de I/O
- Ausência de variáveis globais de dependência

Como saber se está funcionando: quando o agente gera código, você verá uma pausa antes de cada escrita onde ele valida os padrões. Se encontrar problemas, ele corrige antes de salvar.

---

### 2. Review pós-tarefa

| Campo | Valor |
|---|---|
| Arquivo | `.kiro/hooks/post-task-review.kiro.hook` |
| Evento | `postTaskExecution` (após completar uma tarefa de spec) |
| Tipo de ação | `askAgent` |
| Acionamento | Automático — roda quando uma tarefa de spec é marcada como concluída |

O que verifica:
- Segurança: exposição de secrets, falta de validação, endpoints sem autenticação
- Arquitetura: violação de camadas, imports incorretos, lógica fora do lugar
- Padrões: aderência ao TECHNICAL_BASE.md (error handling, DI, logs com zap)
- Testes: cobertura dos use cases, mocks para interfaces

Saída: lista de problemas com severidade (CRÍTICO / ALERTA / SUGESTÃO) e sugestões de correção.

Como saber se está funcionando: após cada tarefa de spec completada, o agente automaticamente faz um review e lista os achados no chat.

---

### 3. Review de PR

| Campo | Valor |
|---|---|
| Arquivo | `.kiro/hooks/pr-review-check.kiro.hook` |
| Evento | `userTriggered` (acionado manualmente) |
| Tipo de ação | `askAgent` |
| Acionamento | Manual — você clica no botão do hook quando quiser |

O que verifica:
- Segurança: secrets hardcoded, inputs não validados, endpoints sem JWT, stack traces expostos
- Arquitetura: regra de dependência, lógica no lugar certo, interfaces nos ports
- Padrões: error handling, DI manual, context.Context, logs zap, health checks
- Testes: unitários nos use cases, mocks, cobertura

Saída: veredito final (APROVADO / APROVADO COM RESSALVAS / MUDANÇAS NECESSÁRIAS) com lista de itens por severidade.

Como usar:
1. Termine sua implementação e faça commit
2. Na aba "Agent Hooks" do Kiro, encontre "Review de PR"
3. Clique no botão de play para acionar
4. O agente analisa o diff e dá o veredito
5. Se aprovado, abra o PR para review humano

---

## Como acessar os hooks

Na barra lateral do Kiro, abra a seção "Agent Hooks". Lá você vê todos os hooks, pode ativar/desativar e acionar os manuais.

Alternativamente, use o Command Palette e busque "Open Kiro Hook UI".

---

## Como criar um novo hook

Duas formas:
1. Pela interface: Command Palette → "Open Kiro Hook UI" → criar novo
2. Pedindo ao Kiro: "Crie um hook que roda os testes quando eu salvar um arquivo .go"

O hook é salvo como JSON em `.kiro/hooks/<id>.kiro.hook`.

---

## Tipos de evento disponíveis

| Evento | Quando dispara |
|---|---|
| `fileEdited` | Quando um arquivo é salvo |
| `fileCreated` | Quando um arquivo é criado |
| `fileDeleted` | Quando um arquivo é deletado |
| `userTriggered` | Quando o usuário clica manualmente |
| `promptSubmit` | Quando uma mensagem é enviada ao agente |
| `agentStop` | Quando o agente termina uma execução |
| `preToolUse` | Antes de uma ferramenta ser executada |
| `postToolUse` | Depois de uma ferramenta ser executada |
| `preTaskExecution` | Antes de uma tarefa de spec começar |
| `postTaskExecution` | Depois de uma tarefa de spec ser concluída |

## Tipos de ação

| Ação | O que faz |
|---|---|
| `askAgent` | Envia um prompt ao agente para ele executar |
| `runCommand` | Executa um comando shell |

---

## Fluxo de desenvolvimento com hooks

```text
Você pede uma implementação
    │
    ▼
Hook "Verificação pré-escrita" (automático)
    → Valida padrões antes de cada arquivo ser salvo
    │
    ▼
Hook "Review pós-tarefa" (automático)
    → Review de segurança e arquitetura ao completar tarefa
    │
    ▼
Você faz commit e prepara o PR
    │
    ▼
Hook "Review de PR" (manual)
    → Analisa o diff completo
    → Dá veredito: APROVADO / MUDANÇAS NECESSÁRIAS
    │
    ▼
Se aprovado → abre PR para review humano
```

---

> Voltar ao índice: [README](../README.md)
