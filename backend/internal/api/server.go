package api

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"devopsconecta/backend/internal/auth"
	"devopsconecta/backend/internal/config"
	"devopsconecta/backend/internal/ids"
	"devopsconecta/backend/internal/store"
	"devopsconecta/backend/web"
)

const (
	authProviderEmail = "email"
	maxNameLength     = 100
	maxPasswordLength = 72
)

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

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
	cfg  config.Config
	store *store.Store
	ids  *ids.Generator
}

func New(cfg config.Config, st *store.Store, gen *ids.Generator) *API {
	return &API{cfg: cfg, store: st, ids: gen}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/auth/register", a.handleRegister)
	mux.HandleFunc("POST /api/auth/login", a.handleLogin)
	mux.HandleFunc("POST /api/auth/logout", a.handleLogout)
	mux.HandleFunc("GET /api/auth/me", a.requireAuth(a.handleMe))
	mux.HandleFunc("GET /api/auth/config", a.handleConfig)
	mux.HandleFunc("/api/", a.handleAPI404)

	spa := a.spa()
	if spa != nil {
		mux.Handle("/", spa)
	}

	return withLogging(mux)
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
		SameSite: http.SameSiteLaxMode,
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
