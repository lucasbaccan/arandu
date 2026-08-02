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
	if resp.Event.PINCode != "meu-pin_1" {
		t.Errorf("PIN customizado não preservado, got %q", resp.Event.PINCode)
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
