package api

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"devopsconecta/backend/internal/auth"
	"devopsconecta/backend/internal/config"
	"devopsconecta/backend/internal/ids"
	"devopsconecta/backend/internal/live"
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
	cfg   config.Config
	store *store.Store
	ids   *ids.Generator
	live  *live.Manager
}

func New(cfg config.Config, st *store.Store, gen *ids.Generator, liveManager *live.Manager) *API {
	// Liga a persistência do estado ao vivo (pergunta atual, revelação,
	// blank/aviso/ocultar) — sem isso, um restart do servidor (deploy, crash,
	// `air` recompilando em dev) apagava tudo, mesmo já tendo sido salvo
	// durante a apresentação. Ver live.Manager.SetLoader.
	liveManager.SetLoader(func(eventID int64) live.EventState {
		ctx := context.Background()
		revealed, err := st.ListRevealedByEvent(ctx, eventID)
		if err != nil {
			log.Printf("api: carregar revelações salvas do evento %d: %v", eventID, err)
			revealed = make(map[int64]map[int64]bool)
		}
		liveState, err := st.GetEventLiveState(ctx, eventID)
		if err != nil {
			log.Printf("api: carregar estado ao vivo salvo do evento %d: %v", eventID, err)
		}
		return live.EventState{
			CurrentQuestionID: liveState.CurrentQuestionID,
			Revealed:          revealed,
			Blanked:           liveState.Blanked,
			Message:           liveState.Message,
			AnswersHidden:     liveState.AnswersHidden,
			NamesHidden:       liveState.NamesHidden,
		}
	})
	return &API{cfg: cfg, store: st, ids: gen, live: liveManager}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/auth/register", a.handleRegister)
	mux.HandleFunc("POST /api/auth/login", a.handleLogin)
	mux.HandleFunc("POST /api/auth/logout", a.handleLogout)
	mux.HandleFunc("GET /api/auth/me", a.requireAuth(a.handleMe))
	mux.HandleFunc("GET /api/auth/config", a.handleConfig)
	mux.HandleFunc("POST /api/events", a.requireAuth(a.handleCreateEvent))
	mux.HandleFunc("GET /api/events", a.requireAuth(a.handleListEvents))
	mux.HandleFunc("GET /api/events/{id}", a.requireAuth(a.handleGetEvent))
	mux.HandleFunc("PATCH /api/events/{id}", a.requireAuth(a.handleUpdateEvent))
	mux.HandleFunc("GET /api/events/{id}/questions", a.requireAuth(a.handleListQuestions))
	mux.HandleFunc("POST /api/events/{id}/questions", a.requireAuth(a.handleCreateQuestion))
	mux.HandleFunc("PUT /api/events/{id}/questions/order", a.requireAuth(a.handleReorderQuestions))
	mux.HandleFunc("PATCH /api/events/{id}/questions/{questionId}", a.requireAuth(a.handleUpdateQuestion))
	mux.HandleFunc("DELETE /api/events/{id}/questions/{questionId}", a.requireAuth(a.handleDeleteQuestion))
	mux.HandleFunc("GET /api/events/{id}/responses", a.requireAuth(a.handleListResponses))
	mux.HandleFunc("PATCH /api/events/{id}/responses/{participantId}/answers/{questionId}", a.requireAuth(a.handleUpdateAnswer))
	mux.HandleFunc("PATCH /api/events/{id}/responses/{participantId}/photo", a.requireAuth(a.handleUpdateParticipantPhoto))
	mux.HandleFunc("POST /api/events/{id}/live/question", a.requireAuth(a.handleLiveSetQuestion))
	mux.HandleFunc("POST /api/events/{id}/live/reveal", a.requireAuth(a.handleLiveReveal))
	mux.HandleFunc("POST /api/events/{id}/live/unreveal", a.requireAuth(a.handleLiveUnreveal))
	mux.HandleFunc("POST /api/events/{id}/live/reveal-all", a.requireAuth(a.handleLiveRevealAll))
	mux.HandleFunc("POST /api/events/{id}/live/reset", a.requireAuth(a.handleLiveReset))
	mux.HandleFunc("POST /api/events/{id}/live/reset-all", a.requireAuth(a.handleLiveResetAll))
	mux.HandleFunc("POST /api/events/{id}/live/blank", a.requireAuth(a.handleLiveSetBlanked))
	mux.HandleFunc("POST /api/events/{id}/live/hide-answers", a.requireAuth(a.handleLiveSetAnswersHidden))
	mux.HandleFunc("POST /api/events/{id}/live/hide-names", a.requireAuth(a.handleLiveSetNamesHidden))
	mux.HandleFunc("POST /api/events/{id}/live/message", a.requireAuth(a.handleLiveSetMessage))
	mux.HandleFunc("POST /api/events/{id}/live/interactions", a.requireAuth(a.handleLiveSetInteractions))
	mux.HandleFunc("POST /api/events/{id}/live/qa/{messageId}/dismiss", a.requireAuth(a.handleLiveDismissQA))
	mux.HandleFunc("GET /api/events/{id}/live/state", a.requireAuth(a.handleLiveAdminState))
	mux.HandleFunc("GET /api/events/{id}/live/stream", a.requireAuth(a.handleLiveAdminStream))
	mux.HandleFunc("GET /api/events/{id}/live/presentation/state", a.requireAuth(a.handleLivePresentationState))
	mux.HandleFunc("GET /api/events/{id}/live/presentation/stream", a.requireAuth(a.handleLivePresentationStream))
	mux.HandleFunc("GET /api/public/events/by-pin", a.handlePublicResolvePIN)
	mux.HandleFunc("GET /api/public/events/{id}", a.handlePublicGetEvent)
	mux.HandleFunc("GET /api/public/events/{id}/participant", a.handlePublicGetParticipant)
	mux.HandleFunc("POST /api/public/events/{id}/submit", a.handleSubmitAnswers)
	mux.HandleFunc("POST /api/public/events/{id}/live/join", a.handleLiveJoin)
	mux.HandleFunc("GET /api/public/events/{id}/live/state", a.handleLiveState)
	mux.HandleFunc("GET /api/public/events/{id}/live/stream", a.handleLiveStream)
	mux.HandleFunc("POST /api/public/events/{id}/live/react", a.handleLiveReact)
	mux.HandleFunc("POST /api/public/events/{id}/live/qa", a.handleLiveSubmitQA)
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
		if strings.HasPrefix(o, "*.") {
			suffixes = append(suffixes, strings.TrimPrefix(o, "*"))
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

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
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

	u, err := a.store.CreateUser(r.Context(), store.User{
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

	a.setSession(w, u.ID)
	writeJSON(w, http.StatusCreated, map[string]any{"user": toUserDTO(u)})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
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

	u, err := a.store.FindUserByEmail(r.Context(), req.Email)
	if errors.Is(err, store.ErrNotFound) || (err == nil && !auth.CheckPassword(u.PasswordHash, req.Password)) {
		writeError(w, http.StatusUnauthorized, "E-mail ou senha inválidos.")
		return
	}
	if err != nil {
		log.Printf("api: login: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao fazer login.")
		return
	}

	a.setSession(w, u.ID)
	writeJSON(w, http.StatusOK, map[string]any{"user": toUserDTO(u)})
}

func (a *API) handleLogout(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleMe(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	u, err := a.store.FindUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Sessão inválida.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": toUserDTO(u)})
}

func (a *API) handleConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"minPasswordLength": a.cfg.MinPasswordLength,
	})
}

func (a *API) handleAPI404(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "Rota não encontrada.")
}

// --- Sessão (cookie httpOnly + JWT) ---

func (a *API) setSession(w http.ResponseWriter, userID int64) {
	token, err := auth.NewSessionToken(a.cfg.JWTSecret, userID, a.cfg.SessionHours)
	if err != nil {
		log.Printf("api: gerar token: %v", err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   a.cfg.CookieSecure,
		SameSite: a.sameSiteMode(),
		MaxAge:   a.cfg.SessionHours * 3600,
	})
}

func clearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
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
			clearSession(w)
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
		if path != "" && !fileExists(static, path) {
			r.URL.Path = "/"
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
