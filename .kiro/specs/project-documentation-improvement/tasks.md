clea# Plano de Implementação: Melhoria da Documentação do Projeto

## Visão Geral

Implementação incremental da documentação do AI Project, seguindo a hierarquia de 3 níveis (README → TECHNICAL_BASE → docs/diagrams). Cada tarefa constrói sobre a anterior, começando pela renomeação de diagramas (base para todos os links), passando pela criação de novos documentos, e finalizando com revisão de coerência e testes de propriedade.

## Tarefas

- [x] 1. Renomear diagramas e atualizar índice
  - [x] 1.1 Renomear arquivos de diagramas existentes com nomes de patterns
    - Renomear `docs/diagrams/architecture-overview.md` → `docs/diagrams/hexagonal-architecture-overview.md`
    - Renomear `docs/diagrams/auth-login-flow.md` → `docs/diagrams/auth-ropc-login-flow.md`
    - Renomear `docs/diagrams/auth-token-refresh.md` → `docs/diagrams/auth-token-refresh-flow.md`
    - Renomear `docs/diagrams/auth-service-to-service.md` → `docs/diagrams/auth-client-credentials-s2s.md`
    - Manter `circuit-breaker-states.md` e `pubsub-event-flow.md` (já contêm nome do pattern)
    - _Requisitos: 8.1, 4.5_
  - [x] 1.2 Atualizar `docs/diagrams/README.md` com novos nomes e tabelas por categoria
    - Reorganizar índice em tabelas agrupadas por categoria (Arquitetura, Autenticação, Mensageria, Resiliência)
    - Colunas: Arquivo | Pattern | Descrição
    - Incluir o novo `auth-pkce-flow.md` (será criado na tarefa 2)
    - _Requisitos: 8.3_
  - [x] 1.3 Atualizar links em `TECHNICAL_BASE.md` para os novos nomes de diagramas
    - Atualizar todos os links na seção 3 (Arquitetura), seção 4 (Autenticação) e seção 8.2 (Circuit Breaker)
    - _Requisitos: 8.2, 9.1_
  - [x] 1.4 Atualizar links em `.cursor/rules/project-base.mdc` para os novos nomes de diagramas
    - Atualizar tabela "Mapa de Diagramas" com novos caminhos
    - _Requisitos: 8.2_

- [x] 2. Criar diagrama PKCE separado e ajustar diagrama ROPC
  - [x] 2.1 Criar `docs/diagrams/auth-pkce-flow.md` com fluxo Authorization Code + PKCE
    - Incluir diagrama Mermaid de sequência: Cliente → auth-service `/authorize` → Keycloak → `/callback` → tokens
    - Incluir diagrama ASCII complementar do fluxo PKCE
    - Seguir template de diagrama definido no design (Visão Geral, ASCII, Mermaid, Parâmetros)
    - Referenciar seção 4 do TECHNICAL_BASE.md
    - _Requisitos: 4.2, 4.5, 1.1, 1.8_
  - [x] 2.2 Ajustar `docs/diagrams/auth-ropc-login-flow.md` para focar apenas no fluxo ROPC
    - Atualizar título para "Fluxo ROPC — Resource Owner Password Credentials"
    - Adicionar diagrama ASCII complementar ao Mermaid existente
    - Garantir que o conteúdo foca apenas em `grant_type=password` (mobile/app)
    - _Requisitos: 4.3, 4.5, 1.1_

- [x] 3. Adicionar diagramas ASCII aos diagramas existentes
  - [x] 3.1 Adicionar diagrama ASCII em `docs/diagrams/hexagonal-architecture-overview.md`
    - Diagrama ASCII da visão macro (clientes → Kong → serviços → dados)
    - Diagrama ASCII da regra de dependência hexagonal (camadas e direção)
    - _Requisitos: 3.1, 3.5, 1.1_
  - [x] 3.2 Adicionar diagrama ASCII em `docs/diagrams/circuit-breaker-states.md`
    - Diagrama ASCII dos estados (Closed → Open → Half-Open) e transições
    - _Requisitos: 1.1_
  - [x] 3.3 Adicionar diagrama ASCII em `docs/diagrams/pubsub-event-flow.md`
    - Diagrama ASCII do fluxo Publisher → Redis → Subscribers
    - _Requisitos: 1.1_
  - [x] 3.4 Adicionar diagrama ASCII em `docs/diagrams/auth-token-refresh-flow.md`
    - Diagrama ASCII do fluxo de refresh token
    - _Requisitos: 1.1_
  - [x] 3.5 Adicionar diagrama ASCII em `docs/diagrams/auth-client-credentials-s2s.md`
    - Diagrama ASCII do fluxo Client Credentials
    - _Requisitos: 1.1_

- [x] 4. Checkpoint — Verificar diagramas
  - Garantir que todos os diagramas renomeados e novos estão corretos, links internos funcionam. Perguntar ao usuário se há dúvidas.

- [x] 5. Atualizar TECHNICAL_BASE.md com novas seções
  - [x] 5.1 Adicionar tabela de autenticação por tipo de cliente na seção 4
    - Tabela com colunas: Tipo de Cliente | Fluxo OAuth 2.0 | Endpoint | Descrição
    - Incluir Web (PKCE), Mobile (ROPC), Desktop/CLI (ROPC), Service-to-Service (Client Credentials)
    - _Requisitos: 4.1, 4.2, 4.3, 4.4_
  - [x] 5.2 Adicionar tabela de validação em duas camadas (Kong + Microsserviço)
    - Tabela com colunas: Validação | Onde | Obrigatória | O que verifica
    - Incluir diagrama ASCII do fluxo Kong (obrigatória) → Microsserviço (opcional/RBAC)
    - _Requisitos: 5.1, 5.2, 5.3, 5.4_
  - [x] 5.3 Adicionar seção dedicada ao Kong API Gateway
    - Tabela de funcionalidades: Funcionalidade | Descrição | Plugin/Config
    - Diagrama ASCII do fluxo de request pelo Kong
    - Diagrama Mermaid complementar do fluxo JWT no Kong
    - Tabela de rate limiting: Tipo | Limite | Janela
    - _Requisitos: 6.1, 6.2, 6.3, 6.4, 6.5_
  - [x] 5.4 Adicionar diagrama ASCII de arquitetura macro simplificado na seção 3
    - Diagrama ASCII mostrando apenas componentes principais (clientes, Kong, Keycloak, microsserviços, bancos)
    - Diagrama Mermaid complementar simplificado (sem subgraphs de observabilidade)
    - Tabela de características de microsserviços: Característica | Descrição
    - Tabela de microsserviços existentes: Serviço | Pattern | Banco de Dados
    - _Requisitos: 3.1, 3.2, 3.3, 3.4_
  - [x] 5.5 Atualizar estrutura documentada do auth-service para refletir código real
    - Documentar que auth-service usa `keycloak/` e `redis/` como adapters (não `postgres/` nem `mongo/`)
    - _Requisitos: 9.2, 9.5_

- [x] 6. Criar README.md na raiz do projeto
  - [x] 6.1 Criar `README.md` com todas as seções obrigatórias
    - Visão geral (máximo 3 frases)
    - Diagrama ASCII de arquitetura macro (bloco ` ```text `)
    - Diagrama Mermaid complementar simplificado
    - Seção "Quick Start" com comandos `docker compose` (bloco ` ```bash `)
    - Tabela de pré-requisitos: Ferramenta | Versão Mínima
    - Tabela de stack: Componente | Tecnologia | Papel
    - Tabela de microsserviços: Serviço | Descrição | Stack | Status
    - Árvore ASCII da estrutura de pastas (bloco ` ```text `)
    - Tabela de documentação: Documento | Descrição | Link
    - Seguir Progressive_Disclosure: links para TECHNICAL_BASE.md e docs/ sem duplicar conteúdo
    - _Requisitos: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 2.9, 2.10_

- [x] 7. Criar guia de skills de IA
  - [x] 7.1 Criar `docs/ai-skills-guide.md`
    - Seção explicando conceito de skills: o que são, onde ficam, como são acionadas
    - Tabela de skills disponíveis: Skill | O que faz | Quando usar (baseado em `.cursor/skills/`)
    - Tabela comparativa rules vs skills: Tipo | Local | Acionamento | Descrição
    - Seção de como escrever prompts eficazes: tabela Critério | Descrição | Exemplo
    - _Requisitos: 7.1, 7.2, 7.3, 7.4_

- [x] 8. Checkpoint — Verificar documentos novos
  - Garantir que README.md, ai-skills-guide.md e TECHNICAL_BASE.md estão corretos. Verificar renderização de tabelas e diagramas. Perguntar ao usuário se há dúvidas.

- [x] 9. Revisão de coerência entre documentação e código
  - [x] 9.1 Verificar correspondência de endpoints: `api/openapi.yaml` vs `handler.go`
    - Confirmar que todos os paths do OpenAPI (`/login`, `/authorize`, `/callback`, `/refresh`, `/logout`, `/healthz/live`, `/healthz/ready`) existem como rotas no handler
    - _Requisitos: 9.3_
  - [x] 9.2 Verificar correspondência de variáveis de ambiente: `docker-compose.yml` vs `config/config.go`
    - Confirmar que todas as variáveis do docker-compose existem no config.go e vice-versa
    - _Requisitos: 9.4_
  - [x] 9.3 Verificar correspondência de estrutura: `TECHNICAL_BASE.md` vs filesystem do auth-service
    - Confirmar que diretórios documentados existem no auth-service real
    - Documentar divergências encontradas e corrigir na documentação
    - _Requisitos: 9.2, 9.5_
  - [x] 9.4 Verificar que todos os links internos em todos os documentos Markdown apontam para arquivos existentes
    - Iterar sobre todos os .md do projeto e validar links relativos
    - _Requisitos: 9.1, 8.2_

- [x] 10. Testes de propriedade e validação estrutural
  - [x] 10.1 Criar estrutura de testes em Go para validação de documentação
    - Criar diretório `docs/tests/` com `go.mod` e dependência de `gopter`
    - Criar arquivo base `docs/tests/doc_test.go` com helpers para iterar sobre arquivos Markdown
    - _Requisitos: 1.1, 1.4_
  - [x] 10.2 Escrever teste de propriedade: representações visuais em blocos de código
    - **Property 1: Representações visuais usam blocos de código apropriados**
    - Iterar sobre todos os .md, verificar que diagramas ASCII estão em ` ```text ` e Mermaid em ` ```mermaid `
    - **Valida: Requisitos 1.1, 1.5**
  - [x] 10.3 Escrever teste de propriedade: hierarquia de headings consistente
    - **Property 2: Hierarquia de headings é consistente**
    - Iterar sobre todos os .md, parsear headings e verificar hierarquia válida (sem saltos de nível)
    - **Valida: Requisitos 1.4**
  - [x] 10.4 Escrever teste de propriedade: syntax highlighting em blocos de código
    - **Property 3: Blocos de código possuem syntax highlighting**
    - Iterar sobre todos os .md, verificar que blocos com conteúdo reconhecível têm tag de linguagem
    - **Valida: Requisitos 1.6**
  - [x] 10.5 Escrever teste de propriedade: links internos usam caminhos relativos
    - **Property 4: Links internos usam caminhos relativos**
    - Iterar sobre todos os .md, extrair links internos e verificar que são relativos
    - **Valida: Requisitos 1.9**
  - [x] 10.6 Escrever teste de propriedade: links internos resolvem para arquivos existentes
    - **Property 5: Links internos resolvem para arquivos existentes**
    - Iterar sobre todos os .md, extrair links internos e verificar que o arquivo destino existe
    - **Valida: Requisitos 8.2, 9.1**
  - [x] 10.7 Escrever teste de propriedade: nomes de diagramas contêm nome do pattern
    - **Property 6: Nomes de arquivos de diagramas contêm o nome do pattern**
    - Iterar sobre `docs/diagrams/`, verificar que cada arquivo (exceto README.md) contém nome de pattern
    - **Valida: Requisitos 4.5, 8.1**

- [x] 11. Checkpoint final — Validação completa
  - Garantir que todos os testes passam, todos os links funcionam, todos os documentos renderizam corretamente no GitHub. Perguntar ao usuário se há dúvidas.

## Notas

- Tarefas marcadas com `*` são opcionais e podem ser puladas para um MVP mais rápido
- Cada tarefa referencia requisitos específicos para rastreabilidade
- Checkpoints garantem validação incremental
- Testes de propriedade usam Go + `gopter` para validar estrutura dos Markdown
- Quando houver divergência entre documentação e código, o código é a fonte da verdade
