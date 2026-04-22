# Documento de Requisitos — Melhoria da Documentação do Projeto

## Introdução

Este documento define os requisitos para a melhoria da documentação do AI Project. O objetivo é tornar a documentação clara, visual e de fácil navegação — especialmente para quem acessa o repositório pelo GitHub. A documentação deve seguir o princípio de "minimal prose, maximum structure": priorizar tabelas, diagramas ASCII, hierarquia visual e progressive disclosure em vez de parágrafos longos. Cobre: estilo de documentação GitHub-friendly, README.md como porta de entrada visual, simplificação da arquitetura com diagramas ASCII e Mermaid, especificação dos padrões de autenticação por tipo de cliente, documentação do Kong API Gateway, melhoria das skills de IA com padrão "O que faz / Quando usar", reorganização de pastas e diagramas com nomes dos patterns, e garantia de coerência entre documentação e código.

## Glossário

- **Documentação**: Conjunto de arquivos Markdown que descrevem a arquitetura, padrões, fluxos e guias do projeto (TECHNICAL_BASE.md, docs/, README.md).
- **README**: Arquivo `README.md` na raiz do repositório que serve como ponto de entrada visual para novos desenvolvedores.
- **Diagrama_ASCII**: Representação visual de fluxos e arquitetura usando caracteres de texto puro (box-drawing characters, setas, pipes), renderizável em qualquer visualizador Markdown sem dependência de plugins.
- **Diagrama**: Arquivo Markdown em `docs/diagrams/` contendo diagramas Mermaid e Diagrama_ASCII que ilustram fluxos e arquitetura.
- **Progressive_Disclosure**: Padrão de documentação onde a visão geral é apresentada primeiro (README, tabelas resumo) e detalhes são acessíveis via links para documentos específicos, evitando sobrecarga de informação.
- **Skill**: Arquivo `SKILL.md` em `.cursor/skills/` que documenta um processo recorrente de desenvolvimento para uso com agentes de IA.
- **Kong**: API Gateway responsável por roteamento, validação JWT, rate limiting e SSL termination.
- **Keycloak**: Servidor de identidade (IAM) responsável por emissão de tokens JWT, gerenciamento de usuários e RBAC.
- **Pattern**: Padrão de design ou arquitetura utilizado no projeto (ex: Hexagonal Architecture, PKCE, ROPC, Client Credentials).
- **ROPC**: Resource Owner Password Credentials — fluxo OAuth 2.0 usado por clientes mobile e desktop.
- **PKCE**: Proof Key for Code Exchange — fluxo OAuth 2.0 usado por clientes web (browser).
- **Client_Credentials**: Fluxo OAuth 2.0 usado para autenticação entre microsserviços (service-to-service).
- **Arquitetura_Hexagonal**: Padrão Ports & Adapters que separa lógica de negócio de infraestrutura.
- **Microsserviço**: Serviço independente com domínio bem definido, seguindo Arquitetura_Hexagonal.
- **GitHub_Markdown**: Variante de Markdown renderizada pela interface web do GitHub, com suporte a tabelas, blocos de código com syntax highlighting, badges, emojis e diagramas Mermaid.
- **Mermaid**: Linguagem de diagramação baseada em texto, suportada nativamente pelo GitHub para renderização de diagramas em arquivos Markdown.
- **Navegação_GitHub**: Experiência de leitura e exploração de um repositório diretamente pela interface web do GitHub, onde o README.md é exibido automaticamente e links relativos permitem navegar entre documentos.
- **Tabela_Referência_Rápida**: Tabela Markdown usada para apresentar informações de forma compacta e escaneável, seguindo o padrão de colunas consistentes (ex: Nome | Descrição | Quando Usar).

## Requisitos

### Requisito 1: Estilo de documentação GitHub-friendly

**User Story:** Como desenvolvedor que acessa o repositório pelo GitHub, eu quero que toda a documentação siga um estilo visual, estruturado e de fácil navegação, para que eu possa compreender o projeto rapidamente sem precisar ler parágrafos longos.

#### Critérios de Aceitação

1. THE Documentação SHALL utilizar Diagrama_ASCII dentro de blocos de código (` ```text `) para representar fluxos e arquitetura, garantindo renderização correta em qualquer visualizador GitHub_Markdown.
2. THE Documentação SHALL utilizar tabelas Markdown como formato principal para apresentar listas de componentes, padrões, comandos e configurações, em vez de listas com marcadores ou parágrafos descritivos.
3. THE Documentação SHALL seguir o padrão de Progressive_Disclosure: cada documento apresenta uma visão geral compacta primeiro, com links para documentos detalhados quando necessário.
4. THE Documentação SHALL utilizar hierarquia visual clara com headings Markdown (h2, h3) para que o sumário automático do GitHub facilite a navegação.
5. THE Documentação SHALL representar a estrutura de pastas do projeto como árvore ASCII dentro de blocos de código (` ```text `), não como listas com marcadores.
6. THE Documentação SHALL utilizar blocos de código com syntax highlighting (` ```bash `, ` ```yaml `, ` ```go `) para todos os comandos e exemplos de configuração.
7. THE Documentação SHALL priorizar estrutura sobre prosa: usar tabelas, diagramas e listas numeradas em vez de parágrafos explicativos longos.
8. THE Documentação SHALL garantir que todos os elementos visuais (tabelas, diagramas Mermaid, blocos de código, Diagrama_ASCII) renderizem corretamente na interface web do GitHub sem dependência de extensões ou plugins externos.
9. THE Documentação SHALL utilizar links relativos entre documentos Markdown para que a navegação funcione diretamente na interface web do GitHub.

### Requisito 2: Adicionar README.md na raiz do projeto

**User Story:** Como desenvolvedor novo no projeto, eu quero encontrar um README.md visual e bem estruturado na raiz do repositório, para que eu possa entender o propósito do projeto, a arquitetura macro e como começar — tudo visível ao abrir o repositório no GitHub.

#### Critérios de Aceitação

1. THE README SHALL conter uma seção de visão geral com uma descrição concisa (máximo 3 frases) do propósito do AI Project.
2. THE README SHALL conter um Diagrama_ASCII de arquitetura macro mostrando o fluxo: Clientes → Kong → Microsserviços → Bancos de Dados, dentro de um bloco de código ` ```text `.
3. THE README SHALL conter uma Tabela_Referência_Rápida de microsserviços com colunas: Serviço | Descrição | Stack | Status.
4. THE README SHALL conter uma seção "Quick Start" posicionada logo após a visão geral e o diagrama de arquitetura, com os comandos para subir o ambiente local via Docker Compose em blocos de código ` ```bash `.
5. THE README SHALL conter uma Tabela_Referência_Rápida da stack de tecnologia com colunas: Componente | Tecnologia | Papel.
6. THE README SHALL conter a estrutura de pastas do repositório como árvore ASCII dentro de um bloco de código ` ```text `.
7. THE README SHALL conter uma seção de links para Documentação detalhada, organizada como tabela com colunas: Documento | Descrição | Link.
8. THE README SHALL conter uma seção de pré-requisitos listando as ferramentas necessárias (Docker, Go 1.22+) como tabela com colunas: Ferramenta | Versão Mínima.
9. THE README SHALL incluir um diagrama Mermaid de arquitetura macro simplificado como complemento ao Diagrama_ASCII, renderizável nativamente na interface web do GitHub.
10. THE README SHALL seguir o padrão de Progressive_Disclosure: apresentar visão geral compacta com links para documentos detalhados, sem duplicar conteúdo extenso do TECHNICAL_BASE.md.

### Requisito 3: Simplificar a documentação de arquitetura com representação visual

**User Story:** Como desenvolvedor, eu quero que a documentação de arquitetura seja visual e direta, para que eu possa entender a topologia do sistema olhando diagramas em vez de ler documentos extensos.

#### Critérios de Aceitação

1. THE Documentação SHALL incluir um Diagrama_ASCII de arquitetura macro simplificado mostrando apenas os componentes principais (clientes, Kong, Keycloak, microsserviços, bancos de dados) dentro de um bloco de código ` ```text `.
2. THE Documentação SHALL incluir um diagrama Mermaid complementar ao Diagrama_ASCII, sem subgraphs de observabilidade, que renderize corretamente na interface web do GitHub.
3. THE Documentação SHALL descrever a abordagem de microsserviços em uma Tabela_Referência_Rápida com colunas: Característica | Descrição (ex: domínio independente, deploy independente, banco próprio).
4. THE Documentação SHALL incluir uma Tabela_Referência_Rápida mapeando cada Microsserviço existente ao Pattern arquitetural utilizado, com colunas: Serviço | Pattern | Banco de Dados.
5. THE Documentação SHALL representar a regra de dependência da Arquitetura_Hexagonal como Diagrama_ASCII mostrando as camadas e direção das dependências.
6. WHEN um novo Microsserviço for adicionado ao projeto, THE Documentação SHALL ser atualizada para incluir o novo Microsserviço na Tabela_Referência_Rápida.

### Requisito 4: Especificar padrões de autenticação por tipo de cliente

**User Story:** Como desenvolvedor, eu quero saber qual Pattern de autenticação OAuth 2.0 é usado para cada tipo de cliente, para que eu possa implementar o fluxo correto sem ambiguidade.

#### Critérios de Aceitação

1. THE Documentação SHALL conter uma Tabela_Referência_Rápida mapeando cada tipo de cliente ao Pattern OAuth 2.0 correspondente, com colunas: Tipo de Cliente | Fluxo OAuth 2.0 | Endpoint | Descrição.
2. THE Documentação SHALL explicar que o fluxo PKCE é utilizado por clientes web (browser) através dos endpoints `/authorize` e `/callback`.
3. THE Documentação SHALL explicar que o fluxo ROPC é utilizado por clientes mobile e app/desktop através do endpoint `/login`.
4. THE Documentação SHALL explicar que o fluxo Client_Credentials é utilizado para comunicação service-to-service, onde cada Microsserviço possui client_id e client_secret próprios no Keycloak.
5. THE Documentação SHALL nomear cada diagrama de autenticação com o Pattern correspondente (ex: "Fluxo PKCE — Authorization Code + PKCE", "Fluxo ROPC — Resource Owner Password Credentials", "Fluxo Client Credentials — Service-to-Service").

### Requisito 5: Documentar autenticação no backend como opcional

**User Story:** Como desenvolvedor de microsserviços, eu quero que a documentação deixe claro que a validação de JWT no backend (dentro do Microsserviço) é opcional quando o Kong já valida o token, para que eu possa decidir o nível de segurança adequado para cada endpoint.

#### Critérios de Aceitação

1. THE Documentação SHALL descrever dois níveis de validação de autenticação em uma Tabela_Referência_Rápida com colunas: Validação | Onde | Obrigatória | O que verifica.
2. THE Documentação SHALL explicar que a validação no Kong verifica assinatura, expiração (exp), emissor (iss) e audiência (aud) do token JWT.
3. THE Documentação SHALL explicar que a validação adicional no Microsserviço é recomendada para verificação de roles e permissões específicas do domínio (RBAC).
4. THE Documentação SHALL incluir um Diagrama_ASCII mostrando o fluxo de validação em duas camadas: Kong (obrigatória) → Microsserviço (opcional/RBAC).

### Requisito 6: Documentar funcionalidade do Kong API Gateway

**User Story:** Como desenvolvedor, eu quero entender o papel e as funcionalidades do Kong no projeto, para que eu possa configurar rotas, plugins e entender como os requests são processados antes de chegar aos microsserviços.

#### Critérios de Aceitação

1. THE Documentação SHALL conter uma seção dedicada ao Kong explicando o papel do Kong como ponto de entrada único para todos os requests externos.
2. THE Documentação SHALL listar as funcionalidades do Kong como Tabela_Referência_Rápida com colunas: Funcionalidade | Descrição | Plugin/Config.
3. THE Documentação SHALL incluir um Diagrama_ASCII mostrando o fluxo de um request passando pelo Kong até o Microsserviço destino, incluindo as etapas de validação.
4. THE Documentação SHALL incluir um diagrama Mermaid complementar mostrando o fluxo de verificação JWT no Kong, renderizável na interface web do GitHub.
5. THE Documentação SHALL documentar as configurações de rate limiting padrão como Tabela_Referência_Rápida com colunas: Tipo | Limite | Janela.

### Requisito 7: Melhorar documentação das skills de IA

**User Story:** Como desenvolvedor que utiliza agentes de IA (Cursor), eu quero que as skills existentes sejam documentadas de forma visual e padronizada, para que eu possa encontrar e utilizar a skill certa rapidamente.

#### Critérios de Aceitação

1. THE Documentação SHALL conter uma seção explicando o conceito de skills no contexto do projeto: o que são, onde ficam armazenadas (.cursor/skills/) e como são acionadas pelo agente de IA.
2. THE Documentação SHALL listar todas as skills disponíveis em uma Tabela_Referência_Rápida com colunas: Skill | O que faz | Quando usar, seguindo o padrão de anatomia consistente do repositório agent-skills.
3. THE Documentação SHALL explicar a relação entre rules (.cursor/rules/) e skills (.cursor/skills/) em uma Tabela_Referência_Rápida com colunas: Tipo | Local | Acionamento | Descrição.
4. THE Documentação SHALL conter uma seção explicando como escrever prompts eficazes para criação de skills, apresentada como tabela de checklist com colunas: Critério | Descrição | Exemplo.

### Requisito 8: Organizar pastas e diagramas com nomes dos patterns

**User Story:** Como desenvolvedor, eu quero que as pastas e diagramas do projeto tenham nomes que reflitam os patterns utilizados, para que eu possa navegar pela documentação de forma intuitiva e entender o contexto de cada arquivo pelo nome.

#### Critérios de Aceitação

1. THE Documentação SHALL nomear cada arquivo de diagrama incluindo o nome do Pattern que o diagrama descreve (ex: `auth-pkce-flow.md`, `auth-ropc-login-flow.md`, `auth-client-credentials-s2s.md`, `hexagonal-architecture-overview.md`, `circuit-breaker-states.md`, `pubsub-event-flow.md`).
2. THE Documentação SHALL atualizar todas as referências internas (links em TECHNICAL_BASE.md, project-base.mdc, README.md e outros documentos) para refletir os novos nomes de arquivos de diagramas.
3. THE Documentação SHALL manter o arquivo `docs/diagrams/README.md` como índice atualizado de todos os diagramas, organizado como Tabela_Referência_Rápida agrupada por categoria (Arquitetura, Autenticação, Mensageria, Resiliência) com colunas: Arquivo | Pattern | Descrição.
4. THE Documentação SHALL organizar a estrutura de pastas de forma que a hierarquia reflita a separação de responsabilidades do projeto (docs/ para documentação, .cursor/ para configuração de IA, cada serviço em pasta própria na raiz).

### Requisito 9: Revisar coerência da estrutura atual

**User Story:** Como desenvolvedor, eu quero que a estrutura do projeto seja revisada e validada quanto à coerência entre documentação, código e diagramas, para que eu possa confiar que a documentação reflete o estado real do projeto.

#### Critérios de Aceitação

1. THE Documentação SHALL garantir que todos os diagramas referenciados no TECHNICAL_BASE.md existam em `docs/diagrams/` e que os links estejam funcionais.
2. THE Documentação SHALL garantir que a estrutura de diretórios do auth-service documentada no TECHNICAL_BASE.md corresponda à estrutura real do código-fonte.
3. THE Documentação SHALL garantir que os endpoints documentados no `api/openapi.yaml` do auth-service correspondam aos endpoints implementados no código.
4. THE Documentação SHALL garantir que as variáveis de ambiente documentadas correspondam às variáveis utilizadas no `docker-compose.yml` e no `config/config.go` do auth-service.
5. IF uma inconsistência for encontrada entre Documentação e código, THEN THE Documentação SHALL ser atualizada para refletir o estado real do código.
