package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devopsconecta/backend/internal/auth"
	"devopsconecta/backend/internal/config"
	"devopsconecta/backend/internal/ids"
	"devopsconecta/backend/internal/store"
)

func newTestAPI(t *testing.T) *API {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("abrir store: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrar store: %v", err)
	}

	cfg := config.Config{
		Port:              "8080",
		JWTSecret:         "test-secret",
		SessionHours:      24,
		MinPasswordLength: 3,
		DatabasePath:      "",
		CookieSecure:      false,
		SnowflakeNode:     1,
	}
	return New(cfg, store.New(db), ids.NewGenerator(1))
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func sessionCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	cookies := rec.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "session" {
			return c
		}
	}
	t.Fatalf("cookie de sessão não encontrado")
	return nil
}

func registerUser(t *testing.T, h http.Handler) *http.Cookie {
	t.Helper()
	rec := doJSON(t, h, http.MethodPost, "/api/auth/register", map[string]string{
		"name": "Ana", "email": "ana@exemplo.com", "password": "segredo",
	}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register: status %d: %s", rec.Code, rec.Body.String())
	}
	return sessionCookie(t, rec)
}

func TestRegisterSuccess(t *testing.T) {
	h := newTestAPI(t).Handler()
	rec := doJSON(t, h, http.MethodPost, "/api/auth/register", map[string]string{
		"name": "Ana", "email": "Ana@Exemplo.com", "password": "abc",
	}, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if sessionCookie(t, rec) == nil {
		t.Error("cookie de sessão esperado")
	}

	var resp struct {
		User userDTO `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode resposta: %v", err)
	}
	if resp.User.ID == "" {
		t.Error("id serializado como string esperado")
	}
	if resp.User.Email != "ana@exemplo.com" {
		t.Errorf("email deveria ser normalizado, got %s", resp.User.Email)
	}
	if resp.User.AuthProvider != "email" {
		t.Errorf("authProvider esperado email, got %s", resp.User.AuthProvider)
	}
}

func TestRegisterValidation(t *testing.T) {
	h := newTestAPI(t).Handler()

	cases := []struct {
		name string
		body map[string]string
		want int
	}{
		{"sem nome", map[string]string{"email": "a@b.com", "password": "123"}, http.StatusBadRequest},
		{"email inválido", map[string]string{"name": "A", "email": "invalido", "password": "123"}, http.StatusBadRequest},
		{"senha curta", map[string]string{"name": "A", "email": "a@b.com", "password": "12"}, http.StatusBadRequest},
		{"json inválido", nil, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doJSON(t, h, http.MethodPost, "/api/auth/register", tc.body, nil)
			if rec.Code != tc.want {
				t.Errorf("status esperado %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
			var resp map[string]string
			_ = json.Unmarshal(rec.Body.Bytes(), &resp)
			if resp["error"] == "" {
				t.Error("mensagem de erro esperada")
			}
		})
	}
}

func TestRegisterPasswordTooLong(t *testing.T) {
	h := newTestAPI(t).Handler()
	long := strings.Repeat("a", 73)
	rec := doJSON(t, h, http.MethodPost, "/api/auth/register", map[string]string{
		"name": "A", "email": "a@b.com", "password": long,
	}, nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status esperado 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	h := newTestAPI(t).Handler()
	registerUser(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/auth/register", map[string]string{
		"name": "Outra", "email": "ana@exemplo.com", "password": "segredo",
	}, nil)
	if rec.Code != http.StatusConflict {
		t.Errorf("status esperado 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLoginSuccess(t *testing.T) {
	h := newTestAPI(t).Handler()
	registerUser(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/auth/login", map[string]string{
		"email": "ana@exemplo.com", "password": "segredo",
	}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if sessionCookie(t, rec) == nil {
		t.Error("cookie de sessão esperado")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	h := newTestAPI(t).Handler()
	registerUser(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/auth/login", map[string]string{
		"email": "ana@exemplo.com", "password": "errada",
	}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status esperado 401, got %d", rec.Code)
	}
}

func TestLoginUnknownUser(t *testing.T) {
	h := newTestAPI(t).Handler()
	rec := doJSON(t, h, http.MethodPost, "/api/auth/login", map[string]string{
		"email": "ninguem@exemplo.com", "password": "qualquer",
	}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status esperado 401, got %d", rec.Code)
	}
}

func TestLoginValidation(t *testing.T) {
	h := newTestAPI(t).Handler()

	rec := doJSON(t, h, http.MethodPost, "/api/auth/login", map[string]string{
		"email": "", "password": "",
	}, nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("campos vazios: status esperado 400, got %d", rec.Code)
	}

	rec = doJSON(t, h, http.MethodPost, "/api/auth/login", map[string]string{
		"email": "", "password": "123",
	}, nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("email vazio: status esperado 400, got %d", rec.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("json inválido: status esperado 400, got %d", rec2.Code)
	}
}

func TestMeInvalidSession(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)

	cookie.Value = cookie.Value + "x"
	rec := doJSON(t, h, http.MethodGet, "/api/auth/me", nil, []*http.Cookie{cookie})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("token adulterado: status esperado 401, got %d", rec.Code)
	}
}

func TestMeUnknownUser(t *testing.T) {
	h := newTestAPI(t).Handler()
	token, err := auth.NewSessionToken("test-secret", 999999, 1)
	if err != nil {
		t.Fatalf("gerar token: %v", err)
	}
	rec := doJSON(t, h, http.MethodGet, "/api/auth/me", nil, []*http.Cookie{{Name: "session", Value: token}})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("usuário inexistente: status esperado 401, got %d", rec.Code)
	}
}

func TestMe(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)

	rec := doJSON(t, h, http.MethodGet, "/api/auth/me", nil, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		User userDTO `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.User.Name != "Ana" {
		t.Errorf("nome esperado Ana, got %s", resp.User.Name)
	}
}

func TestMeUnauthenticated(t *testing.T) {
	h := newTestAPI(t).Handler()
	rec := doJSON(t, h, http.MethodGet, "/api/auth/me", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status esperado 401, got %d", rec.Code)
	}
}

func TestLogout(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/auth/logout", nil, []*http.Cookie{cookie})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status esperado 204, got %d", rec.Code)
	}

	expired := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == "session" && c.MaxAge < 0 {
			expired = true
		}
	}
	if !expired {
		t.Error("cookie de sessão deveria ser expirado no logout")
	}
}

func TestConfig(t *testing.T) {
	h := newTestAPI(t).Handler()
	rec := doJSON(t, h, http.MethodGet, "/api/auth/config", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d", rec.Code)
	}
	var resp struct {
		MinPasswordLength int `json:"minPasswordLength"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.MinPasswordLength != 3 {
		t.Errorf("minPasswordLength esperado 3, got %d", resp.MinPasswordLength)
	}
}

func TestUnknownAPI404(t *testing.T) {
	h := newTestAPI(t).Handler()
	rec := doJSON(t, h, http.MethodGet, "/api/nao-existe", nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status esperado 404, got %d", rec.Code)
	}
}

func TestSpaServesIndex(t *testing.T) {
	h := newTestAPI(t).Handler()

	for _, path := range []string{"/", "/login", "/dashboard"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("%s: status esperado 200, got %d", path, rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
			t.Errorf("%s: content-type esperado html, got %s", path, ct)
		}
		if !strings.Contains(rec.Body.String(), "DevOps Conecta") {
			t.Errorf("%s: resposta deveria conter o app", path)
		}
	}
}

func TestServesFrontendFromDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("FS-FRONT-PAGE"), 0o644); err != nil {
		t.Fatalf("criar index.html: %v", err)
	}

	cfg := config.Config{
		JWTSecret:         "test-secret",
		SessionHours:      24,
		MinPasswordLength: 3,
		FrontendDir:       dir,
	}
	app := New(cfg, store.New(nil), ids.NewGenerator(1))
	h := app.Handler()

	for _, path := range []string{"/", "/login"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Body.String() != "FS-FRONT-PAGE" {
			t.Errorf("%s: esperado página do filesystem, got %q", path, rec.Body.String())
		}
	}
}

func TestDevProxyForwardsFrontend(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("VITE-LIVE-PAGE"))
	}))
	defer upstream.Close()

	cfg := config.Config{
		JWTSecret:         "test-secret",
		SessionHours:      24,
		MinPasswordLength: 3,
		ViteDevURL:        upstream.URL,
	}
	app := New(cfg, store.New(nil), ids.NewGenerator(1))
	h := app.Handler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Body.String() != "VITE-LIVE-PAGE" {
		t.Errorf("esperado página do Vite, got %q", rec.Body.String())
	}
}

func TestDevProxyFallsBackToNextUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("FALLBACK-UP"))
	}))
	defer upstream.Close()

	// primeiro upstream aponta para uma porta sem ninguém escutando
	dead := "http://127.0.0.1:1"

	cfg := config.Config{
		JWTSecret:         "test-secret",
		SessionHours:      24,
		MinPasswordLength: 3,
		ViteDevURL:        dead + "," + upstream.URL,
	}
	app := New(cfg, store.New(nil), ids.NewGenerator(1))
	h := app.Handler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Body.String() != "FALLBACK-UP" {
		t.Errorf("esperado resposta do segundo upstream, got %q", rec.Body.String())
	}
}

func TestDevProxyAllDown(t *testing.T) {
	cfg := config.Config{
		JWTSecret:         "test-secret",
		SessionHours:      24,
		MinPasswordLength: 3,
		ViteDevURL:        "http://127.0.0.1:1",
	}
	app := New(cfg, store.New(nil), ids.NewGenerator(1))
	h := app.Handler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Errorf("status esperado 502, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "make frontend-dev") {
		t.Errorf("mensagem amigável esperada, got %q", rec.Body.String())
	}
}
