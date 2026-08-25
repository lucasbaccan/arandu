package api

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"devopsconecta/backend/internal/auth"
	"devopsconecta/backend/internal/config"
	"devopsconecta/backend/internal/ids"
	"devopsconecta/backend/internal/live"
	"devopsconecta/backend/internal/photos"
	"devopsconecta/backend/internal/store"
	"devopsconecta/backend/web"
)

const (
	authProviderEmail = "email"
	maxNameLength     = 100
	maxPasswordLength = 72
)

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]+$`)

type userDTO struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	AuthProvider string `json:"authProvider"`
	CreatedAt    string `json:"createdAt"`
}

func toUserDTO(u store.User) userDTO {
	return userDTO{
		ID:           strconv.FormatInt(u.ID, 10),
		Email:        u.Email,
		Name:         u.Name,
		AuthProvider: u.AuthProvider,
		CreatedAt:    u.CreatedAt.Format(time.RFC3339),
	}
}

type API struct {
	cfg    config.Config
	store  *store.Store
	ids    *ids.Generator
	live   *live.Manager
	photos *photos.Store
}

func New(cfg config.Config, st *store.Store, gen *ids.Generator, liveManager *live.Manager, photoStore *photos.Store) *API {
	// Liga a persistência do estado ao vivo (pergunta atual, revelação,
	// blank/aviso/ocultar) — sem isso, um restart do servidor (deploy, crash,
	// `air` recompilando em dev) apagava tudo, mesmo já tendo sido salvo
	// durante a apresentação. Ver live.Manager.SetLoader.
	liveManager.SetLoader(func(eventID int64) live.EventState {
		ctx := context.Background()
		revealed, err := st.ListarReveladosPorEvento(ctx, eventID)
		if err != nil {
			log.Printf("api: carregar revelações salvas do evento %d: %v", eventID, err)
			revealed = make(map[int64]map[int64]bool)
		}
		liveState, err := st.BuscarEstadoAoVivoDoEvento(ctx, eventID)
		if err != nil {
			log.Printf("api: carregar estado ao vivo salvo do evento %d: %v", eventID, err)
		}
		return live.EventState{
			CurrentQuestionID:  liveState.CurrentQuestionID,
			Revealed:           revealed,
			Blanked:            liveState.Blanked,
			Message:            liveState.Message,
			AnswersHidden:      liveState.AnswersHidden,
			NamesHidden:        liveState.NamesHidden,
			PresentDensityMode: liveState.PresentDensityMode,
		}
	})
	return &API{cfg: cfg, store: st, ids: gen, live: liveManager, photos: photoStore}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/conta/criar-conta", a.handleCriarConta)
	mux.HandleFunc("POST /api/conta/entrar", a.handleEntrar)
	mux.HandleFunc("POST /api/conta/sair", a.handleSair)
	mux.HandleFunc("GET /api/conta/eu", a.requireAuth(a.handleEu))
	mux.HandleFunc("GET /api/conta/configuracao", a.handleConfiguracao)
	mux.HandleFunc("POST /api/eventos", a.requireAuth(a.handleCriarEvento))
	mux.HandleFunc("GET /api/eventos", a.requireAuth(a.handleListarEventos))
	mux.HandleFunc("GET /api/eventos/{id}", a.requireAuth(a.handleBuscarEvento))
	mux.HandleFunc("PATCH /api/eventos/{id}", a.requireAuth(a.handleAtualizarEvento))
	mux.HandleFunc("GET /api/eventos/{id}/perguntas", a.requireAuth(a.handleListarPerguntas))
	mux.HandleFunc("POST /api/eventos/{id}/perguntas", a.requireAuth(a.handleCriarPergunta))
	mux.HandleFunc("PUT /api/eventos/{id}/perguntas/ordem", a.requireAuth(a.handleReordenarPerguntas))
	mux.HandleFunc("PATCH /api/eventos/{id}/perguntas/{questionId}", a.requireAuth(a.handleAtualizarPergunta))
	mux.HandleFunc("DELETE /api/eventos/{id}/perguntas/{questionId}", a.requireAuth(a.handleRemoverPergunta))
	mux.HandleFunc("GET /api/eventos/{id}/respostas", a.requireAuth(a.handleListarRespostas))
	mux.HandleFunc("PATCH /api/eventos/{id}/respostas/{participantId}/resposta/{questionId}", a.requireAuth(a.handleAtualizarResposta))
	mux.HandleFunc("PATCH /api/eventos/{id}/respostas/{participantId}/foto", a.requireAuth(a.handleAtualizarFotoParticipante))
	mux.HandleFunc("POST /api/eventos/{id}/ao-vivo/pergunta", a.requireAuth(a.handleAoVivoDefinirPergunta))
	mux.HandleFunc("POST /api/eventos/{id}/ao-vivo/revelar", a.requireAuth(a.handleAoVivoRevelar))
	mux.HandleFunc("POST /api/eventos/{id}/ao-vivo/ocultar", a.requireAuth(a.handleAoVivoOcultar))
	mux.HandleFunc("POST /api/eventos/{id}/ao-vivo/revelar-todos", a.requireAuth(a.handleAoVivoRevelarTodos))
	mux.HandleFunc("POST /api/eventos/{id}/ao-vivo/reiniciar", a.requireAuth(a.handleAoVivoReiniciar))
	mux.HandleFunc("POST /api/eventos/{id}/ao-vivo/reiniciar-tudo", a.requireAuth(a.handleAoVivoReiniciarTudo))
	mux.HandleFunc("POST /api/eventos/{id}/ao-vivo/em-branco", a.requireAuth(a.handleAoVivoDefinirEmBranco))
	mux.HandleFunc("POST /api/eventos/{id}/ao-vivo/ocultar-respostas", a.requireAuth(a.handleAoVivoOcultarRespostas))
	mux.HandleFunc("POST /api/eventos/{id}/ao-vivo/ocultar-nomes", a.requireAuth(a.handleAoVivoOcultarNomes))
	mux.HandleFunc("POST /api/eventos/{id}/ao-vivo/modo-densidade", a.requireAuth(a.handleAoVivoDefinirModoDensidade))
	mux.HandleFunc("POST /api/eventos/{id}/ao-vivo/mensagem", a.requireAuth(a.handleAoVivoDefinirMensagem))
	mux.HandleFunc("POST /api/eventos/{id}/ao-vivo/interacoes", a.requireAuth(a.handleAoVivoDefinirInteracoes))
	mux.HandleFunc("POST /api/eventos/{id}/ao-vivo/perguntas/{messageId}/dispensar", a.requireAuth(a.handleAoVivoDispensarPergunta))
	mux.HandleFunc("GET /api/eventos/{id}/ao-vivo/estado", a.requireAuth(a.handleAoVivoEstadoAdmin))
	mux.HandleFunc("GET /api/eventos/{id}/ao-vivo/fluxo", a.requireAuth(a.handleAoVivoFluxoAdmin))
	mux.HandleFunc("GET /api/eventos/{id}/ao-vivo/apresentacao/estado", a.requireAuth(a.handleAoVivoEstadoApresentacao))
	mux.HandleFunc("GET /api/eventos/{id}/ao-vivo/apresentacao/fluxo", a.requireAuth(a.handleAoVivoFluxoApresentacao))
	mux.HandleFunc("GET /api/publico/eventos/por-pin", a.handlePublicoResolverPIN)
	mux.HandleFunc("GET /api/publico/eventos/{id}", a.handlePublicoBuscarEvento)
	mux.HandleFunc("GET /api/publico/eventos/{id}/participante", a.handlePublicoBuscarParticipante)
	mux.HandleFunc("POST /api/publico/eventos/{id}/enviar", a.handleEnviarRespostas)
	mux.HandleFunc("POST /api/publico/eventos/{id}/ao-vivo/entrar", a.handleAoVivoEntrar)
	mux.HandleFunc("GET /api/publico/eventos/{id}/ao-vivo/estado", a.handleAoVivoEstado)
	mux.HandleFunc("GET /api/publico/eventos/{id}/ao-vivo/fluxo", a.handleAoVivoFluxo)
	mux.HandleFunc("POST /api/publico/eventos/{id}/ao-vivo/reagir", a.handleAoVivoReagir)
	mux.HandleFunc("POST /api/publico/eventos/{id}/ao-vivo/perguntas", a.handleAoVivoEnviarPergunta)
	mux.HandleFunc("DELETE /api/publico/eventos/{id}/ao-vivo/perguntas/{messageId}", a.handleAoVivoRemoverPergunta)
	mux.HandleFunc("GET /api/fotos/{participantId}", a.handleFotoParticipante)
	mux.HandleFunc("/api/", a.handleAPI404)

	if a.cfg.ViteDevURL != "" {
		if p := devProxy(a.cfg.ViteDevURL); p != nil {
			mux.Handle("/", p)
		}
	} else if a.cfg.FrontendDir != "" {
		mux.Handle("/", spaHandler(os.DirFS(a.cfg.FrontendDir)))
	} else if spa := a.spa(); spa != nil {
		mux.Handle("/", spa)
	}

	return withLogging(a.cors(mux))
}

// cors libera origens configuradas em CORS_ORIGINS (separadas por vírgula),
// necessário quando o frontend roda em outro domínio (ex: Vercel). Como o
// backend usa cookies de sessão, Allow-Credentials é sempre true e o Origin
// é ecoado (nunca "*"). Entradas com prefixo "*." (ex: *.vercel.app) liberam
// qualquer subdomínio daquele sufixo — útil para previews com URL dinâmica.
// Sem CORS_ORIGINS, nenhum header CORS é emitido.
func (a *API) cors(next http.Handler) http.Handler {
	allowed := map[string]bool{}
	var suffixes []string
	for _, o := range strings.Split(a.cfg.CORSOrigins, ",") {
		if o = strings.TrimSpace(o); o == "" {
			continue
		}
		if strings.Contains(o, "*.") {
			// "*." pode vir com esquema (https://*.vercel.app) — o sufixo é tudo após o "*"
			suffixes = append(suffixes, o[strings.Index(o, "*")+1:])
		} else {
			allowed[o] = true
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		ok := origin != "" && (allowed["*"] || allowed[origin])
		if !ok {
			for _, s := range suffixes {
				if strings.HasSuffix(origin, s) {
					ok = true
					break
				}
			}
		}
		if ok {
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Add("Vary", "Origin")
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// sameSiteMode converte COOKIE_SAMESITE (lax|none|strict) para http.SameSite.
// "none" é o necessário para o frontend em outro domínio (Vercel) receber o
// cookie de sessão — e exige COOKIE_SECURE=true (navegadores rejeitam
// SameSite=None sem Secure).
func (a *API) sameSiteMode() http.SameSite {
	switch strings.ToLower(a.cfg.CookieSameSite) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}

// cookieAttrs decide Secure/SameSite do cookie de sessão. Com COOKIE_SECURE ou
// COOKIE_SAMESITE definidos, manda a configuração. Sem elas, decide pela própria
// requisição, porque as duas combinações válidas dependem de como a página foi
// aberta e errar deixa a pessoa sem sessão:
//   - HTTP puro (dev em localhost): Lax sem Secure — o navegador descarta um
//     cookie Secure vindo de http://.
//   - HTTPS mesmo site: Lax com Secure.
//   - HTTPS cross-site (frontend em outro domínio, ex: Vercel): None + Secure —
//     é a única combinação que o navegador envia numa requisição cross-site.
func (a *API) cookieAttrs(r *http.Request) (bool, http.SameSite) {
	if !a.cfg.CookieAuto {
		return a.cfg.CookieSecure, a.sameSiteMode()
	}
	if !requestIsHTTPS(r) {
		return false, http.SameSiteLaxMode
	}
	if requestIsCrossSite(r) {
		return true, http.SameSiteNoneMode
	}
	return true, http.SameSiteLaxMode
}

// requestIsHTTPS considera o proxy reverso à frente (Traefik, nginx): para o Go
// a conexão chega em texto puro, e só o X-Forwarded-Proto conta a verdade.
func requestIsHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	proto := r.Header.Get("X-Forwarded-Proto")
	if i := strings.IndexByte(proto, ','); i >= 0 {
		proto = proto[:i]
	}
	return strings.EqualFold(strings.TrimSpace(proto), "https")
}

// requestIsCrossSite: o Origin só vem preenchido e diferente do host quando a
// página que fez a chamada mora em outro domínio.
func requestIsCrossSite(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := r.Host
	if fwd := r.Header.Get("X-Forwarded-Host"); fwd != "" {
		host = fwd
	}
	return !strings.EqualFold(u.Host, host)
}

func (a *API) spa() http.Handler {
	sub, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		log.Printf("api: embed do frontend indisponível: %v", err)
		return nil
	}
	return spaHandler(sub)
}

// --- Handlers ---

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *API) handleCriarConta(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Informe seu nome.")
		return
	}
	if len(req.Name) > maxNameLength {
		writeError(w, http.StatusBadRequest, "Nome muito longo.")
		return
	}
	if !isValidEmail(req.Email) {
		writeError(w, http.StatusBadRequest, "Informe um e-mail válido.")
		return
	}
	if len(req.Password) < a.cfg.MinPasswordLength {
		writeError(w, http.StatusBadRequest, "A senha deve ter pelo menos "+strconv.Itoa(a.cfg.MinPasswordLength)+" caracteres.")
		return
	}
	if len(req.Password) > maxPasswordLength {
		writeError(w, http.StatusBadRequest, "A senha é muito longa (máximo "+strconv.Itoa(maxPasswordLength)+" caracteres).")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("api: hash de senha: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao criar a conta.")
		return
	}

	u, err := a.store.CriarUsuario(r.Context(), store.User{
		ID:           a.ids.NextID(),
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: hash,
		AuthProvider: authProviderEmail,
	})
	if errors.Is(err, store.ErrEmailTaken) {
		writeError(w, http.StatusConflict, "Este e-mail já está cadastrado.")
		return
	}
	if err != nil {
		log.Printf("api: criar usuário: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao criar a conta.")
		return
	}

	a.setSession(w, r, u.ID)
	writeJSON(w, http.StatusCreated, map[string]any{"user": toUserDTO(u)})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *API) handleEntrar(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Informe e-mail e senha.")
		return
	}

	u, err := a.store.BuscarUsuarioPorEmail(r.Context(), req.Email)
	if errors.Is(err, store.ErrNotFound) || (err == nil && !auth.CheckPassword(u.PasswordHash, req.Password)) {
		writeError(w, http.StatusUnauthorized, "E-mail ou senha inválidos.")
		return
	}
	if err != nil {
		log.Printf("api: login: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao fazer login.")
		return
	}

	a.setSession(w, r, u.ID)
	writeJSON(w, http.StatusOK, map[string]any{"user": toUserDTO(u)})
}

func (a *API) handleSair(w http.ResponseWriter, r *http.Request) {
	a.clearSession(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleEu(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	u, err := a.store.BuscarUsuarioPorID(r.Context(), userID)
	if err != nil {
		// Token assinado corretamente, mas o dono não existe mais neste banco
		// (conta removida, ou o servidor trocou de DATABASE_PATH). Sem limpar o
		// cookie aqui, o navegador reenvia o mesmo token para sempre e a pessoa
		// fica presa em "Sessão inválida" sem nada que ela possa fazer na tela.
		a.clearSession(w, r)
		writeError(w, http.StatusUnauthorized, "Sessão inválida.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": toUserDTO(u)})
}

func (a *API) handleConfiguracao(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"minPasswordLength": a.cfg.MinPasswordLength,
	})
}

func (a *API) handleAPI404(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "Rota não encontrada.")
}

// --- Sessão (cookie httpOnly + JWT) ---

func (a *API) setSession(w http.ResponseWriter, r *http.Request, userID int64) {
	token, err := auth.NewSessionToken(a.cfg.JWTSecret, userID, a.cfg.SessionHours)
	if err != nil {
		log.Printf("api: gerar token: %v", err)
		return
	}
	secure, sameSite := a.cookieAttrs(r)
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   a.cfg.SessionHours * 3600,
	})
}

// Apagar precisa repetir Secure/SameSite de quando o cookie foi criado: o
// navegador só substitui um cookie por outro de mesmo nome/caminho/domínio, e um
// Set-Cookie de expiração com atributos incompatíveis é ignorado em silêncio.
func (a *API) clearSession(w http.ResponseWriter, r *http.Request) {
	secure, sameSite := a.cookieAttrs(r)
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   -1,
	})
}

type ctxKey string

const ctxUserID ctxKey = "userID"

func userIDFromContext(ctx context.Context) int64 {
	id, _ := ctx.Value(ctxUserID).(int64)
	return id
}

func (a *API) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.SessionCookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Não autenticado.")
			return
		}
		userID, err := auth.VerifySessionToken(a.cfg.JWTSecret, cookie.Value)
		if err != nil {
			a.clearSession(w, r)
			writeError(w, http.StatusUnauthorized, "Sessão inválida.")
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserID, userID)
		next(w, r.WithContext(ctx))
	}
}

// --- Helpers ---

func isValidEmail(email string) bool {
	if !emailRe.MatchString(email) {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return errors.New("JSON inválido.")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("api: encode json: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

// spaHandler serve o frontend buildado com fallback para index.html.
func spaHandler(static fs.FS) http.Handler {
	fileServer := http.FileServerFS(static)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		isIndex := path == "" || !fileExists(static, path)
		if isIndex {
			r.URL.Path = "/"
			// O index.html referencia os arquivos hasheados do build atual (ex:
			// index-BDSztN80.js); um index.html em cache no navegador aponta pra um
			// hash que pode não existir mais depois de um novo build (dev com
			// rebuild automático), servindo uma tela quebrada sem erro visível.
			// Os arquivos em /assets/* têm hash no nome, então esses continuam
			// cacheáveis à vontade.
			w.Header().Set("Cache-Control", "no-store")
		}
		fileServer.ServeHTTP(w, r)
	})
}

func fileExists(fsys fs.FS, path string) bool {
	f, err := fsys.Open(path)
	if err != nil {
		return false
	}
	f.Close()
	return true
}
