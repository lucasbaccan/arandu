<p align="center">
  <img src="frontend/public/img/arandu-completo.png" alt="Arandu" width="420" />
</p>

# Arandu

**Arandu** significa *"conhecimento"* em Guarani — e é exatamente o coração da plataforma: uma aplicação para conduzir dinâmicas, integrações e quebra-gelos ao vivo, onde o conhecimento move as pessoas.

Em uma única aplicação: coleta de respostas pré-evento, apresentação ao vivo no telão e estatísticas pós-evento.

## ✨ Funcionalidades atuais

- **Autenticação por e-mail/senha** (bcrypt + JWT em cookie httpOnly) — telas de login e criação de conta
- **Eventos**:
  - Criar eventos com **PIN de acesso** (automático de 6 dígitos ou personalizado)
  - Listar eventos no dashboard com status e PIN
  - Editar título, PIN e opção de exibir ranking
- **UI/UX**:
  - Tema escuro com partículas animadas de fundo em todas as telas
  - Layout mobile-first (participantes acessam pelo celular)
  - Toasts de feedback (sucesso/erro) nas ações
  - Datas no formato brasileiro (DD/MM/YYYY)
- **Binário único**: frontend (HTML/CSS/JS) embutido no executável Go via `//go:embed`

## 🛠 Stack

| Camada    | Tecnologia                                   |
| --------- | -------------------------------------------- |
| Backend   | Go (REST API + SPA)                          |
| Frontend  | Svelte + Vite                                |
| Banco     | SQLite (WAL, single-writer)                  |
| Identidade| JWT em cookie httpOnly + bcrypt              |

## 📁 Estrutura

```
├── backend/
│   ├── cmd/server/          # entrada (flags, menu interativo, shutdown)
│   └── internal/
│       ├── api/             # rotas e handlers (auth, eventos, proxy dev)
│       ├── auth/            # senha (bcrypt) e sessão (JWT)
│       ├── config/          # configuração por env vars
│       ├── ids/             # gerador de Snowflake IDs
│       └── store/           # SQLite (users, events)
├── frontend/
│   ├── public/img/          # logo oficial
│   └── src/
│       ├── components/      # Button, Card, Input, Particles, Toast
│       ├── lib/             # api, stores (auth, toast, router)
│       └── views/           # Home, Login, Register, Dashboard, EventCreate, EventEdit
├── Makefile                 # fluxo de desenvolvimento padronizado
└── .air.toml                # auto-reload do backend
```

## 🚀 Como rodar

Pré-requisitos: [Go 1.22+](https://go.dev), [Node 18+](https://nodejs.org) e `make` (no Windows, use WSL).

```bash
make deps          # instala dependências (Go + npm)
make dev           # sobe tudo: backend (:8080) + rebuild automático do frontend
```

Abra **http://localhost:8080** — o Go serve a API e o frontend, e qualquer alteração em `.svelte`/`.go` é aplicada automaticamente (vite build --watch + air).

### Outros comandos

```bash
make test          # testes Go + frontend
make build         # binário completo para o sistema atual (backend/bin/arandu)
make build-windows # binário Windows (backend/bin/arandu.exe)
make run           # roda o binário compilado
make stop          # encerra servidores de dev órfãos
make seed          # popula o banco com dados fake (usuário demo@demo.com / senha demo)
```

### Rodando o binário

```bash
./arandu --yes              # assume os padrões (0.0.0.0:8080, ./data/app.db), sem perguntar
./arandu -port 9090         # configura via flags
./arandu                    # sem flags/env: menu interativo pergunta host, porta e banco
```

## 🐳 Docker

A imagem é multi-stage (frontend → backend → runtime mínimo) e produz um único binário com o frontend embutido.

```bash
make docker                # docker build -t arandu:latest .
docker compose up -d       # (ou apenas) sobe com volume de dados + env
```

Para produção, defina um segredo forte de sessão antes de subir:

```bash
export JWT_SECRET=$(openssl rand -hex 32)
docker run -d --name arandu \
  -p 8080:8080 \
  -e JWT_SECRET="$JWT_SECRET" \
  -e COOKIE_SECURE=true \
  -v arandu_data:/app/data \
  arandu:latest
```

- O banco SQLite e as fotos materializadas ficam em `/app/data` (volume `arandu_data`) — sobrevivem a recriação do container.
- Configuração é toda por env vars (mesmas da tabela abaixo); `HOST`/`PORT`/`DATABASE_PATH` já vêm com defaults próprios do container.
- Rode atrás de um proxy HTTPS (Caddy/Nginx) e use `COOKIE_SECURE=true` + `COOKIE_SAMESITE=none` se o frontend estiver em outro domínio.

### Variáveis de ambiente

| Variável            | Padrão           | Descrição                              |
| ------------------- | ---------------- | -------------------------------------- |
| `HOST`              | `0.0.0.0`        | Endereço de escuta (0.0.0.0 expõe na rede) |
| `PORT`              | `8080`           | Porta do servidor                      |
| `DATABASE_PATH`     | `./data/app.db`  | Arquivo do banco SQLite                |
| `JWT_SECRET`        | (dev)            | Segredo para assinar os tokens         |
| `MIN_PASSWORD_LENGTH` | `3`            | Tamanho mínimo de senha                |
| `SESSION_HOURS`     | `24`             | Duração da sessão em horas             |
| `COOKIE_SECURE`     | `false`          | Cookie somente HTTPS                   |
| `SNOWFLAKE_NODE`    | `0`              | Nó do gerador de IDs                   |

## 🧪 Testes

- **Backend (Go):** rotas, autenticação, eventos, máquina de PIN e proxy de dev
- **Frontend (Vitest + Testing Library):** componentes, formulários, guards de rota e stores

```bash
make test
```

## 🗺 Próximos passos

- Fluxo do participante (acesso via PIN + e-mail contra lista de respondentes)
- Perguntas (Tipo A — em grupo e Tipo B — individuais) com votação ao vivo via WebSocket
- Reações com emojis, timers e pontuação por agilidade
- Estatísticas e exportação de dados (CSV/PDF)

---

<p align="center">
  <sub>Arandu — conhecimento que conecta pessoas. 🗣️</sub>
</p>
