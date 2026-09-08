# syntax=docker/dockerfile:1
#
# Imagem de produção do Arandu: um único binário Go com o frontend Svelte
# embutido (//go:embed). Três estágios: build do frontend → build do backend →
# imagem mínima de runtime com os dados em volume.

###############################################################################
# Estágio 1 — frontend (Svelte + Vite): gera os assets estáticos em
# backend/web/dist, que o binário Go embute.
###############################################################################
FROM node:26-alpine AS frontend

WORKDIR /app

# Commit curto do repositório, injetado no build (o .git não é copiado pra
# dentro da imagem). Vem do build-arg GIT_COMMIT; sem ele, cai em dev.
ARG GIT_COMMIT=dev
ENV VITE_APP_COMMIT=$GIT_COMMIT

# Dependências primeiro (camada com cache: só invalida quando o lockfile muda).
COPY frontend/package.json frontend/package-lock.json ./frontend/
RUN cd frontend && npm ci

# Código-fonte e build. O vite grava em ../backend/web/dist (ver vite.config.js).
COPY frontend ./frontend
RUN cd frontend && npm run build

###############################################################################
# Estágio 2 — backend (Go): compila o binário único.
# O SQLite é pure-Go (modernc.org/sqlite), então não precisa de cgo.
###############################################################################
FROM golang:1.26-alpine AS backend

WORKDIR /app

# Cache de módulos Go (só rebaixa quando go.mod/go.sum mudam).
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download

# Código do backend + dist buildado no estágio anterior. backend/web/dist fica
# fora do contexto da build (ver .dockerignore); o que vale é o COPY abaixo.
COPY backend ./backend
COPY --from=frontend /app/backend/web/dist ./backend/web/dist

RUN cd backend && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /arandu ./cmd/server \
    && mkdir -p /app/data

###############################################################################
# Estágio 3 — runtime: imagem mínima distroless (sem shell, sem apk, já vem
# com CA certificates) rodando como usuário não-root. As datas são formatadas
# no navegador (frontend), então o container não precisa de tzdata.
###############################################################################
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=backend /arandu /arandu
COPY --from=backend --chown=65532:65532 /app/data /app/data

# Configuração por env vars (o binário lê env antes de qualquer menu interativo).
ENV HOST=0.0.0.0 \
    PORT=8080 \
    DATABASE_PATH=/app/data/app.db \
    PHOTOS_DIR=/app/data/photos

EXPOSE 8080
VOLUME ["/app/data"]

CMD ["/arandu"]
