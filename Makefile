.PHONY: dev backend-dev frontend-dev frontend-watch run build test clean tools deps

GOBIN := $(shell go env GOBIN 2>/dev/null)
AIR := $(if $(GOBIN),$(GOBIN)/air,$(shell go env GOPATH)/bin/air)

deps: ## Instala dependências (Go + npm)
	@echo "📦 Baixando dependências do Go..."
	@cd backend && go mod download
	@echo "📦 Instalando dependências do frontend (npm)..."
	@cd frontend && npm install
	@echo "✅ Dependências instaladas!"

dev: ## Sobe backend (serve o frontend buildado) + rebuild automático do frontend
	@echo "🚀 Ambiente de desenvolvimento — abra http://localhost:8080 (Go serve API + frontend)"
	@trap 'kill 0' INT TERM EXIT; \
	$(MAKE) backend-dev & \
	$(MAKE) frontend-watch

backend-dev: ## Backend Go com auto-reload (air; serve o frontend do dist em dev)
	@test -x "$(AIR)" || { echo "🛠️  Instalando air..."; go install github.com/air-verse/air@latest; }
	@echo "⚙️  Backend com auto-reload em http://localhost:8080 (frontend servido pelo Go)"
	@cd backend && FRONTEND_DIR=./web/dist "$(AIR)"

frontend-watch: ## Recompila o frontend a cada mudança (vite build --watch)
	@echo "🏗️  Observando o frontend (vite build --watch)..."
	@cd frontend && npm run build:watch

frontend-dev: ## Frontend Svelte com HMR (vite; exposto em 0.0.0.0) — opcional
	@echo "🎨 Frontend com HMR (vite) em http://localhost:5173 (e http://IP-da-rede:5173)"
	@cd frontend && npm run dev -- --host 0.0.0.0

run: ## Roda o binário compilado (API + frontend buildado)
	@echo "▶️  Rodando o servidor compilado em http://localhost:8080"
	@cd backend && ./bin/server

build: ## Compila backend e frontend
	@echo "🏗️  Compilando frontend..."
	@cd frontend && npm run build
	@echo "🏗️  Compilando backend..."
	@cd backend && go build -o bin/server ./cmd/server
	@echo "✅ Build concluído! (backend/bin/server + frontend embutido)"

test: ## Testes de backend e frontend
	@echo "🧪 Rodando testes do backend..."
	@cd backend && go test ./...
	@echo "🧪 Rodando testes do frontend..."
	@cd frontend && npm test
	@echo "🎉 Todos os testes passaram!"

tools: ## Instala ferramentas de desenvolvimento (air)
	@echo "🛠️  Instalando air..."
	@go install github.com/air-verse/air@latest
	@echo "✅ air instalado!"

clean:
	@echo "🧹 Limpando artefatos de build..."
	@rm -rf backend/bin backend/tmp frontend/node_modules frontend/dist
	@echo "🧹 Limpeza concluída!"
