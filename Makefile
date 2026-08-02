.PHONY: dev backend-dev frontend-dev run build test clean tools deps

deps: ## Instala dependências (Go + npm)
	@cd backend && go mod download
	@cd frontend && npm install

dev: ## Sobe backend (auto-reload) + frontend (HMR) juntos
	@trap 'kill 0' INT TERM EXIT; \
	$(MAKE) backend-dev & \
	$(MAKE) frontend-dev

backend-dev: ## Backend Go com auto-reload (air)
	@cd backend && air

frontend-dev: ## Frontend Svelte com HMR (vite)
	@cd frontend && npm run dev

run: ## Roda o binário compilado (API + frontend buildado)
	@cd backend && ./bin/server

build: ## Compila backend e frontend
	@cd frontend && npm run build
	@cd backend && go build -o bin/server ./cmd/server

test: ## Testes de backend e frontend
	@cd backend && go test ./...
	@cd frontend && npm test

tools: ## Instala ferramentas de desenvolvimento (air)
	@go install github.com/air-verse/air@latest

clean:
	@rm -rf backend/bin backend/tmp frontend/node_modules frontend/dist
