.PHONY: dev backend-dev frontend-dev frontend-watch frontend-clean run build build-windows docker test clean stop tools deps seed

GOBIN := $(shell go env GOBIN 2>/dev/null)
AIR := $(if $(GOBIN),$(GOBIN)/air,$(shell go env GOPATH)/bin/air)

HOST ?= 0.0.0.0
PORT ?= 8080
FRONTEND_PORT ?= 5173
# Vazio de propósito: sem COOKIE_SECURE/COOKIE_SAMESITE o backend decide a
# política do cookie por requisição (ver cookieAttrs em internal/api/server.go) —
# Lax sem Secure em http://localhost, None+Secure quando a página vem de outro
# domínio por HTTPS. Fixar um dos dois aqui quebra o outro cenário em silêncio:
# Secure sobre HTTP puro é descartado pelo navegador, e Lax não é enviado
# cross-site. Para forçar: make backend-dev COOKIE_SECURE=true COOKIE_SAMESITE=none
COOKIE_SECURE ?=
COOKIE_SAMESITE ?=
# Origens liberadas no CORS (vírgula; sufixo "*." libera subdomínios; "*" = qualquer)
CORS_ORIGINS ?= *

deps: ## Instala dependências (Go + npm)
	@echo "📦 Baixando dependências do Go..."
	@cd backend && go mod download
	@echo "📦 Instalando dependências do frontend (npm)..."
	@cd frontend && npm install
	@echo "✅ Dependências instaladas!"

dev: ## Sobe backend (serve o frontend buildado) + rebuild automático do frontend
	@echo "🚀 Ambiente de desenvolvimento — abra http://localhost:$(PORT) (Go serve API + frontend)"
	@trap 'kill 0' INT TERM EXIT; \
	$(MAKE) backend-dev & \
	$(MAKE) frontend-watch

backend-dev: ## Backend Go com auto-reload (air; serve o frontend do dist em dev)
	@test -x "$(AIR)" || { echo "🛠️  Instalando air..."; go install github.com/air-verse/air@latest; }
	@echo "⚙️  Backend com auto-reload em http://$(HOST):$(PORT) (frontend servido pelo Go)"
	@set -a; [ -f ./.env ] && . ./.env; set +a; cd backend && HOST="$${HOST:-$(HOST)}" PORT="$${PORT:-$(PORT)}" FRONTEND_DIR=./web/dist COOKIE_SECURE="$(COOKIE_SECURE)" COOKIE_SAMESITE="$(COOKIE_SAMESITE)" CORS_ORIGINS="$(CORS_ORIGINS)" "$(AIR)"

frontend-watch: ## Recompila o frontend a cada mudança (vite build --watch)
	@echo "🏗️  Observando o frontend (vite build --watch)..."
	@cd frontend && BACKEND_PORT=$(PORT) npm run build:watch

frontend-dev: ## Frontend Svelte com HMR (vite; exposto em $(HOST)) — opcional
	@echo "🎨 Frontend com HMR (vite) em http://localhost:$(FRONTEND_PORT) (e http://IP-da-rede:$(FRONTEND_PORT)) — proxy /api -> :$(PORT)"
	@cd frontend && BACKEND_PORT=$(PORT) npm run dev -- --host $(HOST) --port $(FRONTEND_PORT)

frontend-clean: ## Força a recompilação do frontend do zero (limpa cache do Vite e o dist)
	@echo "🧹 Parando o dev server do frontend (se estiver rodando)..."
	@-pkill -f "vite" 2>/dev/null; sleep 1
	@echo "🧹 Removendo cache do Vite (frontend/node_modules/.vite)..."
	@rm -rf frontend/node_modules/.vite
	@echo "🧹 Removendo o dist embutido (backend/web/dist)..."
	@rm -rf backend/web/dist
	@echo "🏗️  Recompilando o frontend do zero (npm run build)..."
	@cd frontend && npm run build
	@echo "✅ Frontend recompilado do zero em backend/web/dist"
	@echo "💡 Próximos passos:"
	@echo "   • dev:              make dev            (ou make frontend-dev para HMR)"
	@echo "   • binário compilado: make build && make run"

run: ## Roda o binário compilado (API + frontend buildado)
	@echo "▶️  Rodando o servidor compilado em http://$(HOST):$(PORT)"
	@cd backend && HOST=$(HOST) PORT=$(PORT) ./bin/arandu

build: ## Compila backend e frontend (binário para o sistema atual)
	@echo "🏗️  Compilando frontend..."
	@cd frontend && npm run build
	@echo "🏗️  Compilando backend..."
	@cd backend && go build -o bin/arandu ./cmd/server
	@echo "✅ Build concluído! (backend/bin/arandu + frontend embutido)"

build-windows: ## Compila o frontend e o backend para Windows (arandu.exe) — rode no WSL
	@echo "🏗️  Compilando frontend..."
	@cd frontend && npm run build
	@echo "🏗️  Compilando backend para Windows (GOOS=windows)..."
	@cd backend && GOOS=windows GOARCH=amd64 go build -o bin/arandu.exe ./cmd/server
	@echo "✅ arandu.exe gerado em backend/bin/arandu.exe"
	@echo "💡 Para rodar no Windows: backend\\bin\\arandu.exe  (ou os comandos manuais:)"
	@echo "   cd frontend && npm run build"
	@echo "   cd backend && GOOS=windows GOARCH=amd64 go build -o bin/arandu.exe ./cmd/server"

docker: ## Builda a imagem Docker de produção (arandu:latest)
	@echo "🐳 Buildando a imagem Docker (arandu:latest)..."
	@docker build -t arandu:latest .
	@echo "✅ Imagem pronta! Rode com: docker run -p 8080:8080 -v arandu_data:/app/data arandu:latest"

test: ## Testes de backend e frontend
	@echo "🧪 Rodando testes do backend..."
	@cd backend && go test ./...
	@echo "🧪 Rodando testes do frontend..."
	@cd frontend && npm test
	@echo "🎉 Todos os testes passaram!"

seed: ## Popula o banco com dados fake pra explorar o app (usuário demo@demo.com / senha demo)
	@echo "🌱 Gerando dados de demonstração (usuário demo@demo.com / senha demo)..."
	@cd backend && go run ./cmd/seed
	@echo "✅ Dados de demonstração prontos!"

tools: ## Instala ferramentas de desenvolvimento (air)
	@echo "🛠️  Instalando air..."
	@go install github.com/air-verse/air@latest
	@echo "✅ air instalado!"

clean:
	@echo "🧹 Limpando artefatos de build..."
	@rm -rf backend/bin backend/tmp frontend/node_modules frontend/dist
	@echo "🧹 Limpeza concluída!"

stop: ## Encerra servidores de dev órfãos (quando o air não está mais rodando)
	@echo "⏹️  Encerrando servidores órfãos..."
	@-pkill -f "arandu" 2>/dev/null; pkill -f "air" 2>/dev/null; pkill -f "vite" 2>/dev/null
	@-fuser -k -TERM $(PORT)/tcp $(FRONTEND_PORT)/tcp 2>/dev/null
	@sleep 1
	@-fuser -k -KILL $(PORT)/tcp $(FRONTEND_PORT)/tcp 2>/dev/null
	@echo "✅ Portas liberadas ($(PORT) e $(FRONTEND_PORT)). Rode 'make dev' para subir de novo."
