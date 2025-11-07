# Tenant CRUD - Template Multi-tenant

Este repositório fornece um ponto de partida completo para aplicações SaaS multi-tenant escritas em Go. Ele combina uma API REST com segregação de dados por tenant, autenticação baseada em JWT e ferramentas de linha de comando para orquestrar o ciclo de vida da aplicação. O objetivo é servir como _boilerplate_ para plataformas que desejam iniciar rapidamente com multi-tenant.

## Visão Geral

- **Linguagem:** Go 1.22+
- **Framework HTTP:** [Gin](https://github.com/gin-gonic/gin)
- **ORM:** [GORM](https://gorm.io/)
- **Banco de dados:** PostgreSQL
- **Autenticação:** JWT (access/refresh)
- **Documentação:** Swagger (exposta em `/doc/index.html`)
- **Comandos CLI:** `./tenant-crud --help`

A aplicação segue uma arquitetura modular com camadas bem definidas para domínio, aplicação e infraestrutura. Cada domínio (por exemplo, _tenant_ ou _user_) possui seus próprios _controllers_, _services_, _repositories_ e DTOs dentro de `internal/iam`.

## Estrutura do Projeto

```
├── cmd
│   ├── bootstrap        # Inicialização do ambiente, DI e servidor HTTP
│   ├── cli              # Interface de linha de comando (start/stop/migrations)
│   └── server           # HTTP server e definição das rotas
├── configs.json.example # Configuração de ambiente (copie para configs.json)
├── docs                 # Artefatos Swagger gerados pelo swaggo
├── internal
│   ├── iam              # Domínio de identidade e acesso multi-tenant
│   │   ├── application  # Casos de uso (aplicação) e serviços de autenticação
│   │   ├── di           # Contêiner de injeção de dependências (domain/application)
│   │   ├── domain       # Entidades, DTOs, services e controllers por agregado
│   │   └── middleware   # Middlewares de autorização e resolução de tenant
│   ├── infra            # Implementações de infraestrutura (DB, JWT, sistema)
│   └── pkg              # Utilidades reutilizáveis (erros, respostas HTTP etc.)
├── go.mod / go.sum      # Dependências do Go
└── main.go              # Entrada da aplicação, chama o CLI
```

### Fluxo do Servidor

1. `main.go` executa `cmd/cli.Execute()`.
2. O CLI lê os _flags_ e pode iniciar o servidor HTTP (`--start`), executar migrations ou operar no banco.
3. `cmd/bootstrap.New()` carrega `configs.json`, inicializa o gerador de tokens JWT, a conexão com PostgreSQL e monta o contêiner de dependências (`internal/iam/di`).
4. `cmd/server.NewHTTPServer()` cria o servidor HTTP com Gin, registra a documentação Swagger e aplica middlewares multi-tenant antes de encaminhar para os _controllers_.

### Domínio Multi-tenant

- **Middleware de autorização:** `internal/iam/middleware` resolve o tenant a partir do token JWT e injeta o contexto no `*gin.Context`.
- **Controllers:** `internal/iam/domain/<aggregate>/controller` expõem rotas registradas via `RegisterRoutes`.
- **Services:** `internal/iam/domain/<aggregate>/service` aplicam regras de negócio por tenant.
- **Repositories:** `internal/iam/domain/<aggregate>/repository` encapsulam operações GORM filtrando pelo `tenant_id`.

Essa separação facilita a criação de novos agregados multi-tenant replicando a estrutura existente para `tenant` e `user`.

## Configuração do Ambiente

1. **Copie o arquivo de exemplo:**
   ```bash
   cp configs.json.example configs.json
   ```
2. **Atualize os campos necessários:**
   - `app.env`: `dev` ou `prod` (controla o modo do Gin).
   - `security.jwt_*`: segredos e expiração do token.
   - `server.http.port`: porta exposta pelo servidor.
   - `databases.postgres`: credenciais e host do banco multi-tenant.
3. **Variáveis adicionais:** Você pode sobrescrever qualquer chave de `configs.json` com variáveis de ambiente (Viper segue a convenção `APP_ENV`, `SECURITY_JWT_ACCESS_SECRET`, etc.).

## Comandos da CLI

```bash
./tenant-crud --start                 # Inicia o servidor HTTP
./tenant-crud --stop                  # Finaliza o servidor em execução
./tenant-crud --migration-update      # Executa migrations de atualização
./tenant-crud --migration-seed        # Popular dados base
./tenant-crud --db-check              # Verifica tabelas disponíveis
./tenant-crud --db-delete             # Remove todas as tabelas (CUIDADO)
./tenant-crud --db-backup --local=/caminho/do/backup
```

- Os comandos que interagem com o banco inicializam `internal/infra/database/postgres` automaticamente.
- O arquivo PID do servidor é salvo em `run/server.pid` para permitir `--stop`.

## Executando a Aplicação

1. **Instale as dependências do Go:**
   ```bash
   go mod download
   ```
2. **Execute migrations (opcional):**
   ```bash
   go run main.go --migration-update
   go run main.go --migration-seed
   ```
3. **Inicie o servidor:**
   ```bash
   go run main.go --start
   ```
4. **Acesse a documentação:**
   - Swagger: `http://localhost:8080/doc/index.html`
   - API base: `http://localhost:8080/api/v1`

## Convenções de Desenvolvimento

- **Injeção de dependências:** Use o contêiner de `internal/iam/di` para registrar novos domínios e serviços. Evite instanciar dependências diretamente nos _controllers_.
- **Separação por camadas:** Replique o padrão `controller -> service -> repository` para novos módulos.
- **DTOs e validação:** Utilize os pacotes `dto` e `errors` existentes como referência para mensagens consistentes.
- **Middlewares:** Centralize regras de autorização/acesso em `internal/iam/middleware` para manter a aplicação multi-tenant segura.
- **Testes:** Recomendado utilizar `testing` com `suite` para serviços e `httptest` para controllers.

## Próximos Passos / Customização

- Adicione observabilidade (metrics/logs) integrando com `internal/pkg`.
- Implemente _feature flags_ específicos por tenant adicionando colunas nas tabelas e validando na camada de serviço.
- Ajuste as políticas de expiração de tokens em `internal/infra/jwt` conforme a necessidade do produto.
- Para ambientes de produção, considere habilitar TLS na frente do servidor ou utilizar um _reverse proxy_.

---

Sinta-se à vontade para adaptar este template às necessidades da sua plataforma multi-tenant. Pull requests e sugestões são bem-vindos!
