package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

// registerUserNamed cadastra mais um usuário (além do primeiro, que sempre
// vira super admin — ver TestFirstUserIsSuperAdmin) e devolve o cookie de
// sessão dele.
func registerUserNamed(t *testing.T, h http.Handler, name, email string) *http.Cookie {
	t.Helper()
	rec := doJSON(t, h, http.MethodPost, "/api/conta/criar-conta", map[string]string{
		"name": name, "email": email, "password": "segredo",
	}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("registrar %s: status %d: %s", name, rec.Code, rec.Body.String())
	}
	return sessionCookie(t, rec)
}

func meRole(t *testing.T, h http.Handler, cookie *http.Cookie) string {
	t.Helper()
	rec := doJSON(t, h, http.MethodGet, "/api/conta/eu", nil, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("/api/conta/eu: status %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		User userDTO `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp.User.Role
}

func TestFirstUserIsSuperAdmin(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	if role := meRole(t, h, cookie); role != "super_admin" {
		t.Errorf("papel esperado super_admin para o primeiro cadastro, got %s", role)
	}
}

func TestSecondUserIsOrganizer(t *testing.T) {
	h := newTestAPI(t).Handler()
	registerUser(t, h)
	bia := registerUserNamed(t, h, "Bia", "bia@exemplo.com")
	if role := meRole(t, h, bia); role != "organizer" {
		t.Errorf("papel esperado organizer para o segundo cadastro, got %s", role)
	}
}

func TestAdminRoutesRequireSuperAdmin(t *testing.T) {
	h := newTestAPI(t).Handler()
	registerUser(t, h) // super admin, não usado neste teste
	bia := registerUserNamed(t, h, "Bia", "bia@exemplo.com")

	paths := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/admin/usuarios"},
		{http.MethodGet, "/api/admin/eventos"},
		{http.MethodGet, "/api/admin/configuracoes"},
	}
	for _, p := range paths {
		t.Run(p.path, func(t *testing.T) {
			rec := doJSON(t, h, p.method, p.path, nil, []*http.Cookie{bia})
			if rec.Code != http.StatusForbidden {
				t.Errorf("status esperado 403 para organizador comum, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestAdminListarUsuarios(t *testing.T) {
	h := newTestAPI(t).Handler()
	admin := registerUser(t, h)
	bia := registerUserNamed(t, h, "Bia", "bia@exemplo.com")
	createEventAndGetID(t, h, bia, "Evento da Bia")

	rec := doJSON(t, h, http.MethodGet, "/api/admin/usuarios", nil, []*http.Cookie{admin})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Users []adminUserDTO `json:"users"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Users) != 2 {
		t.Fatalf("esperado 2 usuários, got %d", len(resp.Users))
	}
	if resp.Users[0].Role != "super_admin" {
		t.Errorf("primeiro usuário deveria ser super_admin, got %s", resp.Users[0].Role)
	}
	if resp.Users[1].EventCount != 1 {
		t.Errorf("Bia deveria ter 1 evento, got %d", resp.Users[1].EventCount)
	}
}

func TestAdminExcluirUsuario(t *testing.T) {
	h := newTestAPI(t).Handler()
	admin := registerUser(t, h)
	bia := registerUserNamed(t, h, "Bia", "bia@exemplo.com")
	eventID := createEventAndGetID(t, h, bia, "Evento da Bia")

	var biaID string
	{
		rec := doJSON(t, h, http.MethodGet, "/api/admin/usuarios", nil, []*http.Cookie{admin})
		var resp struct {
			Users []adminUserDTO `json:"users"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		biaID = resp.Users[1].ID
	}

	// Super admin não pode se auto-excluir por aqui.
	var adminID string
	{
		rec := doJSON(t, h, http.MethodGet, "/api/conta/eu", nil, []*http.Cookie{admin})
		var resp struct {
			User userDTO `json:"user"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		adminID = resp.User.ID
	}
	rec := doJSON(t, h, http.MethodDelete, "/api/admin/usuarios/"+adminID, nil, []*http.Cookie{admin})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("auto-exclusão deveria ser bloqueada, status %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, h, http.MethodDelete, "/api/admin/usuarios/"+biaID, nil, []*http.Cookie{admin})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// O evento de Bia some junto.
	rec = doJSON(t, h, http.MethodGet, "/api/eventos/"+eventID, nil, []*http.Cookie{admin})
	if rec.Code != http.StatusNotFound {
		t.Errorf("evento deveria ter sido removido com a conta, status %d", rec.Code)
	}

	rec = doJSON(t, h, http.MethodDelete, "/api/admin/usuarios/"+biaID, nil, []*http.Cookie{admin})
	if rec.Code != http.StatusNotFound {
		t.Errorf("excluir de novo deveria dar 404, got %d", rec.Code)
	}
}

func TestSuperAdminAcessaEventoDeOutroOrganizador(t *testing.T) {
	h := newTestAPI(t).Handler()
	admin := registerUser(t, h)
	bia := registerUserNamed(t, h, "Bia", "bia@exemplo.com")
	eventID := createEventAndGetID(t, h, bia, "Evento da Bia")

	rec := doJSON(t, h, http.MethodGet, "/api/eventos/"+eventID, nil, []*http.Cookie{admin})
	if rec.Code != http.StatusOK {
		t.Fatalf("super admin deveria acessar evento de outro dono, status %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, h, http.MethodPatch, "/api/eventos/"+eventID, map[string]any{"title": "Renomeado pelo admin"}, []*http.Cookie{admin})
	if rec.Code != http.StatusOK {
		t.Fatalf("super admin deveria editar evento de outro dono, status %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Event eventDTO `json:"event"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Event.Title != "Renomeado pelo admin" {
		t.Errorf("título não foi atualizado: %+v", resp.Event)
	}
	if resp.Event.OwnerID == "" {
		t.Error("ownerId deveria continuar preenchido")
	}
}

func TestAdminListarEventos(t *testing.T) {
	h := newTestAPI(t).Handler()
	admin := registerUser(t, h)
	bia := registerUserNamed(t, h, "Bia", "bia@exemplo.com")
	createEventAndGetID(t, h, admin, "Evento do admin")
	createEventAndGetID(t, h, bia, "Evento da Bia")

	rec := doJSON(t, h, http.MethodGet, "/api/admin/eventos", nil, []*http.Cookie{admin})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Events []adminEventDTO `json:"events"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Events) != 2 {
		t.Fatalf("esperado 2 eventos (de todos os donos), got %d", len(resp.Events))
	}
	if resp.Events[0].OwnerEmail == "" {
		t.Error("ownerEmail deveria vir preenchido")
	}
}

func TestAdminConfiguracoesRegistro(t *testing.T) {
	h := newTestAPI(t).Handler()
	admin := registerUser(t, h)

	// Padrão: registro habilitado.
	rec := doJSON(t, h, http.MethodGet, "/api/conta/configuracao", nil, nil)
	var cfg struct {
		RegistrationEnabled bool `json:"registrationEnabled"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &cfg)
	if !cfg.RegistrationEnabled {
		t.Error("registro deveria começar habilitado")
	}

	rec = doJSON(t, h, http.MethodPatch, "/api/admin/configuracoes", map[string]any{"registrationEnabled": false, "emailEnabled": true}, []*http.Cookie{admin})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, h, http.MethodPost, "/api/conta/criar-conta", map[string]string{
		"name": "Bia", "email": "bia@exemplo.com", "password": "segredo",
	}, nil)
	if rec.Code != http.StatusForbidden {
		t.Errorf("cadastro deveria estar bloqueado, status %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, h, http.MethodGet, "/api/conta/configuracao", nil, nil)
	_ = json.Unmarshal(rec.Body.Bytes(), &cfg)
	if cfg.RegistrationEnabled {
		t.Error("registro deveria refletir desabilitado")
	}

	// Reabilitando, o cadastro volta a funcionar.
	rec = doJSON(t, h, http.MethodPatch, "/api/admin/configuracoes", map[string]any{"registrationEnabled": true, "emailEnabled": true}, []*http.Cookie{admin})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, h, http.MethodPost, "/api/conta/criar-conta", map[string]string{
		"name": "Bia", "email": "bia@exemplo.com", "password": "segredo",
	}, nil)
	if rec.Code != http.StatusCreated {
		t.Errorf("cadastro deveria voltar a funcionar, status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminConfiguracoesEmail(t *testing.T) {
	h := newTestAPI(t).Handler()
	admin := registerUser(t, h)

	// Padrão: e-mail habilitado.
	rec := doJSON(t, h, http.MethodGet, "/api/conta/configuracao", nil, nil)
	var cfg struct {
		EmailEnabled bool `json:"emailEnabled"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &cfg)
	if !cfg.EmailEnabled {
		t.Error("e-mail deveria começar habilitado")
	}

	rec = doJSON(t, h, http.MethodPatch, "/api/admin/configuracoes", map[string]any{"registrationEnabled": true, "emailEnabled": false}, []*http.Cookie{admin})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, h, http.MethodGet, "/api/conta/configuracao", nil, nil)
	_ = json.Unmarshal(rec.Body.Bytes(), &cfg)
	if cfg.EmailEnabled {
		t.Error("e-mail deveria refletir desabilitado")
	}

	// Não é super admin nem está logado: a rota de admin continua negando.
	rec = doJSON(t, h, http.MethodPatch, "/api/admin/configuracoes", map[string]any{"registrationEnabled": true, "emailEnabled": true}, nil)
	if rec.Code != http.StatusForbidden && rec.Code != http.StatusUnauthorized {
		t.Errorf("sem sessão deveria negar, status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestFluxoRedefinicaoDeSenhaPorLink(t *testing.T) {
	h := newTestAPI(t).Handler()
	admin := registerUser(t, h)
	bia := registerUserNamed(t, h, "Bia", "bia@exemplo.com")

	var biaID string
	{
		rec := doJSON(t, h, http.MethodGet, "/api/admin/usuarios", nil, []*http.Cookie{admin})
		var resp struct {
			Users []adminUserDTO `json:"users"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		biaID = resp.Users[1].ID
	}
	_ = bia

	rec := doJSON(t, h, http.MethodPost, "/api/admin/usuarios/"+biaID+"/link-redefinicao", nil, []*http.Cookie{admin})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var linkResp struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expiresAt"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &linkResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if linkResp.Token == "" {
		t.Fatal("token vazio")
	}

	// Token válido: a checagem pública confirma.
	rec = doJSON(t, h, http.MethodGet, "/api/publico/redefinir-senha?token="+linkResp.Token, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("validar token: status %d: %s", rec.Code, rec.Body.String())
	}

	// Redefine a senha usando o token.
	rec = doJSON(t, h, http.MethodPost, "/api/publico/redefinir-senha", map[string]string{
		"token": linkResp.Token, "novaSenha": "nova-senha-123",
	}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("redefinir: status %d: %s", rec.Code, rec.Body.String())
	}

	// A senha antiga não funciona mais; a nova sim.
	rec = doJSON(t, h, http.MethodPost, "/api/conta/entrar", map[string]string{
		"email": "bia@exemplo.com", "password": "segredo",
	}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("senha antiga deveria falhar, status %d", rec.Code)
	}
	rec = doJSON(t, h, http.MethodPost, "/api/conta/entrar", map[string]string{
		"email": "bia@exemplo.com", "password": "nova-senha-123",
	}, nil)
	if rec.Code != http.StatusOK {
		t.Errorf("senha nova deveria funcionar, status %d: %s", rec.Code, rec.Body.String())
	}

	// Token de uso único: usar de novo falha.
	rec = doJSON(t, h, http.MethodPost, "/api/publico/redefinir-senha", map[string]string{
		"token": linkResp.Token, "novaSenha": "outra-senha-456",
	}, nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("reuso do token deveria falhar, status %d", rec.Code)
	}
}

func TestRedefinirSenhaTokenInvalido(t *testing.T) {
	h := newTestAPI(t).Handler()

	rec := doJSON(t, h, http.MethodGet, "/api/publico/redefinir-senha?token=inexistente", nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status esperado 404, got %d", rec.Code)
	}

	rec = doJSON(t, h, http.MethodPost, "/api/publico/redefinir-senha", map[string]string{
		"token": "inexistente", "novaSenha": "qualquer-coisa",
	}, nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status esperado 400, got %d", rec.Code)
	}
}
