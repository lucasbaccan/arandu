.PHONY: dev backend-dev frontend-dev frontend-watch run build build-windows test clean stop tools deps seed

GOBIN := $(shell go env GOBIN 2>/dev/null)
AIR := $(if $(GOBIN),$(GOBIN)/air,$(shell go env GOPATH)/bin/air)

HOST ?= 0.0.0.0
PORT ?= 8888
FRONTEND_PORT ?= 5173

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
	@cd backend && HOST=$(HOST) PORT=$(PORT) FRONTEND_DIR=./web/dist "$(AIR)"

frontend-watch: ## Recompila o frontend a cada mudança (vite build --watch)
	@echo "🏗️  Observando o frontend (vite build --watch)..."
	@cd frontend && BACKEND_PORT=$(PORT) npm run build:watch

frontend-dev: ## Frontend Svelte com HMR (vite; exposto em $(HOST)) — opcional
	@echo "🎨 Frontend com HMR (vite) em http://localhost:$(FRONTEND_PORT) (e http://IP-da-rede:$(FRONTEND_PORT)) — proxy /api -> :$(PORT)"
	@cd frontend && BACKEND_PORT=$(PORT) npm run dev -- --host $(HOST) --port $(FRONTEND_PORT)

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
