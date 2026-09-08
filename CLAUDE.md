# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Arandu ("conhecimento" in Guarani) — a single-binary app for running live icebreaker/team-building
dynamics: pre-event answer collection, live presentation on a big screen, and post-event stats.
Go backend embeds the built Svelte frontend (`//go:embed`) into one executable.

## Commands

```bash
make deps          # go mod download + npm install
make dev           # backend (air, auto-reload) + frontend (vite build --watch), serves on :8080
make frontend-dev  # optional: Svelte dev server with HMR on :5173 (proxies /api to :8080)
make build          # production build: frontend then backend binary -> backend/bin/arandu
make build-windows  # cross-compile for Windows (run from WSL)
make run            # run the compiled binary
make stop           # kill orphaned dev servers (air/vite/backend/tmp/arandu)
make test           # go test ./... + npm test (vitest run)
```

Single test, backend (from `backend/`):
```bash
go test ./internal/api/...
go test ./internal/api/ -run TestHandleCreateEvent
```

Single test, frontend (from `frontend/`):
```bash
npx vitest run src/views/EventEdit.test.js
npx vitest run -t "nome do teste"
```

## Releases (Conventional Commits + semantic-release)

O projeto segue **Conventional Commits** (`feat:`, `fix:`, `feat!:`/`BREAKING CHANGE:` etc.), validado
no CI (`.github/workflows/ci.yml`) via commitlint. Nada de pre-commit hook local: o repo é Go na
raiz e o pacote npm em `frontend/`, então o husky (que espera o `.git` junto do `package.json`) não
se aplica — a validação fica no CI das PRs.

**A versão (fonte única) está em `frontend/package.json`** e é a mesma exibida na UI
(`buildInfo`, injetada pelo Vite `define` em `vite.config.js`). Ao dar push em `main`, o workflow
`.github/workflows/release.yml` roda o `semantic-release`:

- lê os Conventional Commits desde a última tag;
- calcula o próximo semver (`feat` → minor, `fix` → patch, `feat!`/breaking → major);
- faz bump no `frontend/package.json` (+ lock), atualiza `frontend/CHANGELOG.md`;
- cria o commit da release, a tag (`vX.Y.Z`) e o Release no GitHub, e faz push de tudo
  (commit + tag) com `GITHUB_TOKEN` (`contents: write`).

Para liberar localmente (para testar sem o CI): `npm --prefix frontend run release` (equivale a
`npx semantic-release`). O link "GitHub" + versão/commit aparecem no "Como funciona" e no popover
de ajuda (`frontend/src/lib/buildInfo.js`).

Running the compiled binary directly:
```bash
./arandu --yes              # non-interactive, defaults to 0.0.0.0:8080, ./data/app.db
./arandu -port 9090          # flags override env/defaults
./arandu                     # no flags/env -> interactive menu (host, port, db path)
```

Config is env-var driven (see `.env.example` / `backend/internal/config/config.go`): `HOST`, `PORT`,
`DATABASE_PATH`, `PHOTOS_DIR`, `JWT_SECRET`, `MIN_PASSWORD_LENGTH`, `SESSION_HOURS`, `COOKIE_SECURE`,
`SNOWFLAKE_NODE`. `VITE_DEV_URL`/`FRONTEND_DIR` are dev-only overrides used by `make dev`/`frontend-dev`
so the Go server can proxy to or serve from an unbuilt frontend instead of the embedded `dist`.

**Participant photos are stored in the DB as base64 data URLs but never sent in JSON payloads.**
`backend/internal/photos` materializes them as files under `PHOTOS_DIR` (`./data/photos`) and serves
them via `GET /api/fotos/{participantId}` — DTOs (`responses`, public participant, live snapshots)
carry only that URL (`a.photoURL(p)`). The first request for a photo decodes the DB value and writes
the file (slower); later requests serve straight from disk with `Cache-Control`
`public, max-age=3600, must-revalidate` (`photos.ImageMaxAge` = 1h, then revalidation via
If-Modified-Since → 304). Updating
or removing a photo invalidates the cached file (`photos.Remove`). `cmd/server/main.go` runs a daily
cleanup loop (`RunCleanupLoop`, 24h tick) that deletes files older than 7 days (`PHOTOS_DIR` files are
recreated on demand from the DB, so deletion is safe).

## Architecture

**Single binary, two runtime frontend sources.** `backend/internal/api/server.go`'s `Handler()` picks,
in priority order, how to serve `/`: proxy to `ViteDevURL` (HMR dev), serve `FrontendDir` from disk (air
dev, `web/dist`), or serve the `//go:embed`-ed `backend/web/dist` (`backend/web/embed.go`) for the
production binary. All three exist so `make dev` gets auto-reload on both sides while `make build`
produces one self-contained executable. Everything under `/api/` is a REST/SSE API on the same
`http.ServeMux`; everything else falls through to the SPA handler with `index.html` fallback.

**Two parallel identity systems, both JWT-in-cookie/query, deliberately not unified:**
- Owners/organizers: email+password (bcrypt) -> session JWT in an httpOnly cookie
  (`backend/internal/auth/session.go`), checked by `requireAuth` middleware. This is who manages events,
  questions, and drives the live presentation.
- Live viewers (participants/observers, no login): join a running event via PIN + email, get a
  short-lived `LiveClaims` JWT (`backend/internal/auth/ao_vivo.go`) that's passed as a query-string `token`
  (not a cookie, since these are cross-device links/QR codes) to the `/api/publico/eventos/{id}/ao-vivo/*`
  endpoints.

**Live presentation state lives in memory, not SQLite** (`backend/internal/live/gerenciador.go`). Which
question is on screen, which participants have been "revealed", blanked/message overlay state, and
answers-hidden state are all per-event, mutex-protected, in-process maps — deliberately not persisted,
since it resets on restart and only one process runs at a time. State changes are pushed to viewers via
Server-Sent Events: a coalescing `Subscribe` channel (buffer 1, "something changed, go re-fetch a
snapshot") drives the main live state, while `SubscribeReactions` is a separate non-coalescing channel
because individual emoji reactions must each render (dropping under backpressure is acceptable — it's
cosmetic). `handleAoVivoFluxoAdmin` / `handleAoVivoFluxo` in `backend/internal/api/ao_vivo.go` are the SSE
endpoints for organizer and public views respectively.

**SQLite is single-writer.** `store.Open` sets `SetMaxOpenConns(1)` with WAL mode — see
`backend/internal/store/store.go`. Schema migrations are plain idempotent `CREATE TABLE IF NOT EXISTS` /
`ALTER TABLE ... ADD COLUMN` (swallowing "duplicate column" errors) run at startup in `Migrate()`, not a
versioned migration tool — new columns get added there.

**IDs are Snowflake, not autoincrement** (`backend/internal/ids/snowflake.go`), one `Generator` shared
across the app (per-instance mutex, not a DB sequence). They're serialized as JSON strings (`id,string`
convention / manual `strconv.FormatInt`) everywhere to avoid precision loss in JS.

**Frontend has no framework router or build-time routing** — `frontend/src/lib/router.js` is a ~15-line
hand-rolled pathname store (pushState + popstate), and `App.svelte` matches routes with a handful of
regexes/equality checks in one `{#if}` chain. Auth-gating redirects also live inline in `App.svelte`'s
reactive block, keyed off `authStore.js`'s `authReady`/`user` stores — there's no route-guard
abstraction.

**Every screen wears one of two shells** (padronização das "telas padronizadas"). Organizer screens:
`components/TopBar.svelte` (56px — only identity and account: logo, area, theme, help, avatar menu; never
changes from screen to screen) plus `components/CrumbBar.svelte` (44px — carries the path, the status chip
and *that screen's* primary action). Public screens: `components/PublicShell.svelte` — particles + gradient,
full logo at 168px, a 440px column, and the fixed corner controls (back / theme / help) that replace the
topbar where there is none. Anything shared between them lives in `app.css` tokens: fields are white with a
1.5px border and radius 10, buttons are radius 10 with exactly four variants
(`primary`/`secondary`/`ghost`/`danger` — one primary action per screen), and errors always carry an icon
plus weight 700 while hints stay neutral text. `components/Chip.svelte` is *the* chip (status, question
kind, option), colored from `lib/eventStatus.js` — screens never pick those colors themselves.
`lib/themeStore.js` drives light/dark via `<html data-theme>`; it's the only thing the app puts in
localStorage. `/tela` (`Tela.svelte`) is the living reference for all of it.

**`frontend/src/lib/api.js`** is a single hand-written client mirroring every backend route 1:1 (grouped
`eventos`/`publico.eventos` namespaces) — there's no codegen from the Go handlers, so adding a backend route
means adding the matching entry here by hand.

> **Whenever a new screen (view) is created, update this file**: add it to the "Screens" list below with
> a one-line description, and add its route to `App.svelte`'s match/render logic if not already covered
> above.

## Screens (`frontend/src/views/`)

Routes are matched by hand in `App.svelte` (see "no framework router" above); most take an `id` prop
parsed from the URL.

| Route | View | Description |
| --- | --- | --- |
| `/` | `Inicio.svelte` | Public landing page. Participants type an event PIN to join (resolves PIN -> event id, then `navigate` to `/plateia/{id}`); organizers get "Entrar"/"Criar conta" links to `/entrar`/`/criar-conta`. |
| `/entrar` | `Entrar.svelte` | Organizer email/password login. |
| `/criar-conta` | `CriarConta.svelte` | Organizer account creation. |
| `/painel` | `Painel.svelte` | Authenticated organizer home: lists their events (status, PIN), links to create/edit. |
| `/evento/novo` | `EventoNovo.svelte` | Form to create a new event (title, PIN). |
| `/evento/{id}` | `EventoEditar.svelte` | Main organizer event-management screen: edit title/PIN/ranking toggle, manage questions (`QuestionForm`), view participant responses (`ResponsesPanel`). |
| `/responder/{id}` | `Responder.svelte` | Public pre-event form: participant identifies with name/email, answers the event's questions, optionally uploads/crops a photo (`AvatarCropper`). No login — a private edit link lets them come back and change answers/photo later. |
| `/palco/{id}` | `Palco.svelte` | Organizer's live-presentation control panel — drives the big-screen show: pick current question, reveal/unreveal/reveal-all participants, blank the screen, hide answers, push a message, toggle audience interactions, watch Q&A submissions. |
| `/palco/{id}/apresentar` | `PalcoApresentar.svelte` | Projection window opened by Palco's "Modo apresentação" — the big-screen show, mirrored over SSE. |
| `/plateia/{id}` | `Plateia.svelte` | Public live view — one unified screen for everyone: top menu (logo, event title, PIN, theme toggle), emoji reactions animating in the background (`ReactionBurstLayer background`) and a single-line bottom dock with compact emoji reactions + a question box for the presenter (💬). It shows **no board content** (no pending queue, questions or answers — the presenter drives those on `/palco` and `/apresentar`); it only reflects the organizer's live state over SSE (`snapshot.message`, `blanked`, `interactionsEnabled`). Session lives only in the URL (`?pin=`) — no login, no localStorage. |
| `/como-funciona` | `ComoFunciona.svelte` | Public tutorial page — expands on `HelpButton`'s popover ("Como funciona o Arandu") with a participant walkthrough and an organizer walkthrough. Linked from `HelpButton`'s "Saiba mais" and footed with a link to `/privacidade`. |
| `/privacidade` | `Privacidade.svelte` | Public privacy policy page — the product-level version of the per-event privacy explainer inline in `Responder.svelte`'s "privacy" step. Linked from `ComoFunciona.svelte`'s footer. |
| `/tela` | `Tela.svelte` | Internal component/design-system showcase (buttons, inputs, selects, switches, cards, toasts, etc.) — not part of the product flow. |
| `/mapa-do-site` | `MapaDoSite.svelte` | Internal site map — every URL of the app grouped by flow; static routes are clickable links, dynamic ones (`{id}`) shown as patterns. Linked from `HelpButton`'s "Mapa do site" and from `/debug`. |
| `/debug` | `Debug.svelte` | Internal developer screen — quick links to `/mapa-do-site` (site map) and `/tela` (design reference), plus current app state (route, theme, user, window). |
| `/opcoes` | `Opcoes.svelte` | Internal decision page — multiple UI alternatives side by side in mini cards (icons/chevrons, buttons, the telão pending-row "show all names" options, etc.), comparing web/mobile renderings. |
