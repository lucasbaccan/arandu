package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

var digitsRe = regexp.MustCompile(`^\d{6}$`)

func createEvent(t *testing.T, h http.Handler, cookie *http.Cookie, body map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	return doJSON(t, h, http.MethodPost, "/api/events", body, []*http.Cookie{cookie})
}

func TestCreateEventAutoPIN(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)

	rec := createEvent(t, h, cookie, map[string]string{"title": "Conecta DevOps"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Event eventDTO `json:"event"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Event.ID == "" {
		t.Error("id serializado como string esperado")
	}
	if !digitsRe.MatchString(resp.Event.PINCode) {
		t.Errorf("PIN automático deveria ter 6 dígitos, got %q", resp.Event.PINCode)
	}
	if resp.Event.Status != "PREPARATION" {
		t.Errorf("status esperado PREPARATION, got %s", resp.Event.Status)
	}
}

func TestCreateEventCustomPIN(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)

	rec := createEvent(t, h, cookie, map[string]string{"title": "Retro", "pinCode": "meu-pin_1"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Event eventDTO `json:"event"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Event.PINCode != "MEU-PIN_1" {
		t.Errorf("PIN customizado deveria ser normalizado para maiúsculas, got %q", resp.Event.PINCode)
	}
}

func TestCreateEventValidation(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)

	cases := []struct {
		name string
		body map[string]string
		want int
	}{
		{"sem título", map[string]string{}, http.StatusBadRequest},
		{"título vazio", map[string]string{"title": "  "}, http.StatusBadRequest},
		{"pin inválido", map[string]string{"title": "A", "pinCode": "espaço inválido"}, http.StatusBadRequest},
		{"pin longo demais", map[string]string{"title": "A", "pinCode": "12345678901234567890123456"}, http.StatusBadRequest},
		{"pin com acento", map[string]string{"title": "A", "pinCode": "café"}, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := createEvent(t, h, cookie, tc.body)
			if rec.Code != tc.want {
				t.Errorf("status esperado %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCreateEventUnauthenticated(t *testing.T) {
	h := newTestAPI(t).Handler()
	rec := doJSON(t, h, http.MethodPost, "/api/events", map[string]string{"title": "A"}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status esperado 401, got %d", rec.Code)
	}
}

func TestCreateEventCustomPINTaken(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)

	rec := createEvent(t, h, cookie, map[string]string{"title": "A", "pinCode": "ocupado"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("primeiro evento deveria ser criado, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = createEvent(t, h, cookie, map[string]string{"title": "B", "pinCode": "ocupado"})
	if rec.Code != http.StatusConflict {
		t.Errorf("PIN duplicado: status esperado 409, got %d", rec.Code)
	}
}

func TestCreateEventCustomPINTakenCaseInsensitive(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)

	rec := createEvent(t, h, cookie, map[string]string{"title": "A", "pinCode": "dev"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("primeiro evento deveria ser criado, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = createEvent(t, h, cookie, map[string]string{"title": "B", "pinCode": "DEV"})
	if rec.Code != http.StatusConflict {
		t.Errorf("PIN duplicado (case-insensitive): status esperado 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListEventsOnlyOwned(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookieAna := registerUser(t, h)

	// segunda conta
	rec := doJSON(t, h, http.MethodPost, "/api/auth/register", map[string]string{
		"name": "Bia", "email": "bia@exemplo.com", "password": "segredo",
	}, nil)
	cookieBia := sessionCookie(t, rec)

	createEvent(t, h, cookieAna, map[string]string{"title": "Evento da Ana"})
	createEvent(t, h, cookieBia, map[string]string{"title": "Evento da Bia"})

	rec = doJSON(t, h, http.MethodGet, "/api/events", nil, []*http.Cookie{cookieAna})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d", rec.Code)
	}
	var resp struct {
		Events []eventDTO `json:"events"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Events) != 1 || resp.Events[0].Title != "Evento da Ana" {
		t.Errorf("esperado apenas o evento da Ana, got %+v", resp.Events)
	}
}

func TestListEventsUnauthenticated(t *testing.T) {
	h := newTestAPI(t).Handler()
	rec := doJSON(t, h, http.MethodGet, "/api/events", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status esperado 401, got %d", rec.Code)
	}
}

func createEventAndGetID(t *testing.T, h http.Handler, cookie *http.Cookie, title string) string {
	t.Helper()
	rec := createEvent(t, h, cookie, map[string]string{"title": title})
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar evento: status %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Event eventDTO `json:"event"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp.Event.ID
}

func TestGetEvent(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventAndGetID(t, h, cookie, "Meu evento")

	rec := doJSON(t, h, http.MethodGet, "/api/events/"+id, nil, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Event eventDTO `json:"event"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Event.Title != "Meu evento" {
		t.Errorf("título esperado Meu evento, got %s", resp.Event.Title)
	}
}

func TestGetEventErrors(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookieAna := registerUser(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/auth/register", map[string]string{
		"name": "Bia", "email": "bia@exemplo.com", "password": "segredo",
	}, nil)
	cookieBia := sessionCookie(t, rec)

	id := createEventAndGetID(t, h, cookieAna, "Da Ana")

	cases := []struct {
		name   string
		path   string
		cookie *http.Cookie
		want   int
	}{
		{"evento de outro dono", "/api/events/" + id, cookieBia, http.StatusNotFound},
		{"evento inexistente", "/api/events/999999", cookieAna, http.StatusNotFound},
		{"id inválido", "/api/events/abc", cookieAna, http.StatusBadRequest},
		{"sem autenticação", "/api/events/" + id, nil, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doJSON(t, h, http.MethodGet, tc.path, nil, []*http.Cookie{tc.cookie})
			if rec.Code != tc.want {
				t.Errorf("status esperado %d, got %d", tc.want, rec.Code)
			}
		})
	}
}

func TestUpdateEvent(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventAndGetID(t, h, cookie, "Antes")

	rec := doJSON(t, h, http.MethodPatch, "/api/events/"+id, map[string]any{
		"title": "Depois", "pinCode": "novo-pin", "configShowRanking": true,
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Event eventDTO `json:"event"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Event.Title != "Depois" || resp.Event.PINCode != "NOVO-PIN" || !resp.Event.ShowRanking {
		t.Errorf("evento atualizado divergente: %+v", resp.Event)
	}
}

func TestUpdateEventKeepsPINWhenEmpty(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventAndGetID(t, h, cookie, "Evento")

	rec := doJSON(t, h, http.MethodPatch, "/api/events/"+id, map[string]any{
		"title": "Título novo", "pinCode": "", "configShowRanking": false,
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Event eventDTO `json:"event"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if !digitsRe.MatchString(resp.Event.PINCode) {
		t.Errorf("PIN deveria ser mantido (6 dígitos), got %q", resp.Event.PINCode)
	}
}

func TestUpdateEventErrors(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventAndGetID(t, h, cookie, "Evento")

	cases := []struct {
		name string
		body map[string]any
		want int
	}{
		{"sem título", map[string]any{"title": "", "pinCode": "", "configShowRanking": false}, http.StatusBadRequest},
		{"pin inválido", map[string]any{"title": "A", "pinCode": "com espaço", "configShowRanking": false}, http.StatusBadRequest},
		{"evento inexistente", map[string]any{"title": "A", "pinCode": "", "configShowRanking": false}, http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := "/api/events/" + id
			if tc.body["title"] == "A" && tc.want == http.StatusNotFound {
				path = "/api/events/999999"
			}
			rec := doJSON(t, h, http.MethodPatch, path, tc.body, []*http.Cookie{cookie})
			if rec.Code != tc.want {
				t.Errorf("status esperado %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestUpdateEventPINTaken(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)

	rec := createEvent(t, h, cookie, map[string]string{"title": "Primeiro", "pinCode": "mesmo-pin"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar primeiro evento: %d", rec.Code)
	}
	id2 := createEventAndGetID(t, h, cookie, "Segundo")

	rec = doJSON(t, h, http.MethodPatch, "/api/events/"+id2, map[string]any{
		"title": "Segundo", "pinCode": "mesmo-pin", "configShowRanking": false,
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusConflict {
		t.Errorf("PIN duplicado: status esperado 409, got %d: %s", rec.Code, rec.Body.String())
	}
}
