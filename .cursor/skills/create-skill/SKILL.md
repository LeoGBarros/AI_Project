---
name: create-skill
description: Criar uma nova skill para o projeto, seguindo os padrões de estrutura, linguagem e convenções já estabelecidos. Use quando o usuário pedir para criar uma nova skill, automatizar um workflow recorrente ou documentar um processo repetitivo como skill.
---

# Criar Nova Skill

## Antes de Começar

Leia as skills existentes como referência de formato e estilo:
- [`.cursor/skills/create-microservice/SKILL.md`](../create-microservice/SKILL.md)
- [`.cursor/skills/implement-usecase/SKILL.md`](../implement-usecase/SKILL.md)
- [`.cursor/skills/implement-event/SKILL.md`](../implement-event/SKILL.md)

Todas seguem o mesmo padrão descrito na seção "Padrão do Projeto" abaixo.

## Informações Necessárias

Antes de criar, confirme com o usuário:
1. **Nome da skill** (ex: `deploy-service`, `create-migration`) — minúsculas, separado por hífens, max 64 chars
2. **Objetivo** — qual tarefa recorrente esta skill resolve?
3. **Escopo** — projeto (`.cursor/skills/`) ou pessoal (`~/.cursor/skills/`)?
4. **Referências** — precisa referenciar seções do `TECHNICAL_BASE.md` ou diagramas?
5. **Templates de código** — a skill inclui exemplos de código Go ou outros artefatos?
6. **Arquivos de suporte** — precisa de arquivos auxiliares (reference.md, examples.md, scripts)?

Use a ferramenta `AskQuestion` para coletar estas informações de forma estruturada quando disponível.

---

## Padrão do Projeto

Toda skill deste projeto segue estas convenções:

### Idioma
- Escrita em **português brasileiro**

### Frontmatter YAML obrigatório
```yaml
---
name: nome-da-skill
description: Descrição concisa do que a skill faz e quando usar. Use quando o usuário pedir para...
---
```
- `name`: minúsculas, apenas letras, números e hífens, max 64 chars
- `description`: max 1024 chars, terceira pessoa, incluir **O QUE** faz e **QUANDO** usar

### Seções obrigatórias

| Seção | Conteúdo |
|---|---|
| **Antes de Começar** | Referências a documentos/diagramas que devem ser lidos antes de implementar |
| **Informações Necessárias** | Lista numerada do que confirmar com o usuário antes de começar |
| **Passo a Passo** | Instruções sequenciais e concretas, com exemplos de código quando aplicável |
| **Checklist Final** | Lista de verificação com `- [ ]` para validar que tudo foi feito |

### Regras de conteúdo

- Máximo de **500 linhas** no `SKILL.md`
- **Exemplos concretos** de código sobre descrições abstratas
- **Terminologia consistente** — escolha um termo e use em todo o documento
- Se o conteúdo exceder 500 linhas, use **progressive disclosure**: coloque detalhes em arquivos auxiliares (`reference.md`, `examples.md`) referenciados a partir do `SKILL.md`
- Referências a arquivos auxiliares devem ser de **um nível de profundidade** (link direto do SKILL.md)

---

## Passo a Passo

### 1. Coletar Requisitos

Utilize a ferramenta `AskQuestion` (quando disponível) para confirmar:

```
Perguntas sugeridas:
- "Qual o nome da skill?" com campo livre
- "Onde deve ser armazenada?" com opções ["Projeto (.cursor/skills/)", "Pessoal (~/.cursor/skills/)"]
- "A skill precisa de arquivos de suporte?" com opções ["Sim", "Não"]
```

Se o `AskQuestion` não estiver disponível, pergunte conversacionalmente.

### 2. Criar o Diretório

```bash
SKILL_NAME=<nome-da-skill>
mkdir -p .cursor/skills/$SKILL_NAME
```

### 3. Escrever o `SKILL.md`

Use o template abaixo como base, adaptando para o caso específico:

```markdown
---
name: <nome-da-skill>
description: <Descrição em português. O QUE faz + QUANDO usar.>
---

# <Título da Skill>

## Antes de Começar

Leia os seguintes documentos antes de começar:
- [`TECHNICAL_BASE.md` — seção X.X](../../../TECHNICAL_BASE.md) — <quando aplicável>
- [`docs/diagrams/<diagrama>.md`](../../../docs/diagrams/<diagrama>.md) — <quando aplicável>

## Informações Necessárias

Antes de implementar, confirme com o usuário:
1. **Campo 1** (ex: valor esperado)
2. **Campo 2** (ex: valor esperado)

---

## Passo a Passo

### Passo 1 — <Título do passo>

<Instruções concretas com exemplos de código quando aplicável>

### Passo 2 — <Título do passo>

<Instruções concretas>

---

## Checklist Final

- [ ] Item de verificação 1
- [ ] Item de verificação 2
- [ ] Skill registrada na tabela de skills em `.cursor/rules/project-base.mdc`
```

### 4. Criar Arquivos de Suporte (se necessário)

Se a skill precisar de documentação extensa ou scripts:

```
<nome-da-skill>/
├── SKILL.md              # Obrigatório — instruções principais
├── reference.md          # Opcional — documentação detalhada
├── examples.md           # Opcional — exemplos de uso
└── scripts/              # Opcional — scripts utilitários
    └── helper.sh
```

### 5. Registrar na Tabela de Skills

Adicione a nova skill na tabela `Skills Disponíveis` em `.cursor/rules/project-base.mdc`:

```markdown
| `<nome-da-skill>` | <Descrição curta de quando usar> |
```

A tabela fica na seção `## Skills Disponíveis` do arquivo `.cursor/rules/project-base.mdc`.

---

## Anti-padrões

| Evitar | Fazer |
|---|---|
| Escrever em inglês | Manter todo o conteúdo em português brasileiro |
| Skill genérica sem contexto do projeto | Referenciar TECHNICAL_BASE.md e diagramas quando aplicável |
| SKILL.md com mais de 500 linhas | Usar progressive disclosure com arquivos auxiliares |
| Descrições vagas no frontmatter ("ajuda com coisas") | Incluir O QUE faz + QUANDO usar com termos de trigger |
| Nomes genéricos (`helper`, `utils`) | Nomes descritivos (`create-migration`, `deploy-service`) |
| Múltiplos termos para o mesmo conceito | Escolher um termo e usar consistentemente |
| Não registrar na tabela de skills | Sempre adicionar em `project-base.mdc` ao final |

---

## Checklist Final

- [ ] Diretório criado em `.cursor/skills/<nome>/`
- [ ] `SKILL.md` com frontmatter YAML válido (`name` + `description`)
- [ ] `description` escrita em terceira pessoa com O QUE + QUANDO
- [ ] Seções obrigatórias presentes: "Antes de Começar", "Informações Necessárias", "Passo a Passo", "Checklist Final"
- [ ] Conteúdo em português brasileiro
- [ ] SKILL.md com menos de 500 linhas
- [ ] Terminologia consistente em todo o documento
- [ ] Referências a arquivos auxiliares com no máximo um nível de profundidade
- [ ] Skill registrada na tabela de skills em `.cursor/rules/project-base.mdc`
