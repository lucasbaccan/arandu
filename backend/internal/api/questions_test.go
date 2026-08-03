package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func createEventForQuestions(t *testing.T, h http.Handler, cookie *http.Cookie) string {
	t.Helper()
	return createEventAndGetID(t, h, cookie, "Evento com perguntas")
}

func TestCreateQuestion(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventForQuestions(t, h, cookie)

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+id+"/questions", map[string]any{
		"title":   "Qual é o seu prato favorito?",
		"type":    "GROUP",
		"options": []string{"Pizza", "Sushi", "Churrasco"},
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Question questionDTO `json:"question"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	q := resp.Question
	if q.ID == "" || q.EventID != id {
		t.Errorf("ids divergentes: %+v", q)
	}
	if q.Title != "Qual é o seu prato favorito?" || q.Type != "GROUP" {
		t.Errorf("pergunta divergente: %+v", q)
	}
	if q.LayoutView != "TIMELINE" {
		t.Errorf("layout padrão esperado TIMELINE, got %s", q.LayoutView)
	}
	if len(q.Options) != 3 || q.Options[0].Text != "Pizza" || q.Options[0].ID == "" {
		t.Errorf("opções divergentes: %+v", q.Options)
	}
	if q.OrderIndex != 0 {
		t.Errorf("primeira pergunta deveria ter order 0, got %d", q.OrderIndex)
	}
}

func TestCreateQuestionOrdering(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventForQuestions(t, h, cookie)

	doJSON(t, h, http.MethodPost, "/api/events/"+id+"/questions", map[string]any{
		"title": "Primeira", "type": "GROUP", "options": []string{"A", "B"},
	}, []*http.Cookie{cookie})
	rec := doJSON(t, h, http.MethodPost, "/api/events/"+id+"/questions", map[string]any{
		"title": "Segunda", "type": "INDIVIDUAL", "options": []string{"Única"},
	}, []*http.Cookie{cookie})

	var resp struct {
		Question questionDTO `json:"question"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Question.OrderIndex != 1 {
		t.Errorf("segunda pergunta deveria ter order 1, got %d", resp.Question.OrderIndex)
	}
}

func TestCreateQuestionValidation(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventForQuestions(t, h, cookie)

	cases := []struct {
		name string
		body map[string]any
		want int
	}{
		{"sem título", map[string]any{"type": "GROUP", "options": []string{"A", "B"}}, http.StatusBadRequest},
		{"título vazio", map[string]any{"title": "  ", "type": "GROUP", "options": []string{"A", "B"}}, http.StatusBadRequest},
		{"tipo inválido", map[string]any{"title": "P", "type": "OUTRO", "options": []string{"A", "B"}}, http.StatusBadRequest},
		{"layout inválido", map[string]any{"title": "P", "type": "GROUP", "layoutView": "XYZ", "options": []string{"A", "B"}}, http.StatusBadRequest},
		{"sem opções", map[string]any{"title": "P", "type": "GROUP", "options": []string{}}, http.StatusBadRequest},
		{"opção vazia", map[string]any{"title": "P", "type": "GROUP", "options": []string{"A", ""}}, http.StatusBadRequest},
		{"opções duplicadas", map[string]any{"title": "P", "type": "GROUP", "options": []string{"A", "A"}}, http.StatusBadRequest},
		{"group com uma opção", map[string]any{"title": "P", "type": "GROUP", "options": []string{"A"}}, http.StatusBadRequest},
		{"opção longa demais", map[string]any{"title": "P", "type": "GROUP", "options": []string{"A", "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"}}, http.StatusBadRequest},
		{"muitas opções", map[string]any{"title": "P", "type": "INDIVIDUAL", "options": []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}}, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doJSON(t, h, http.MethodPost, "/api/events/"+id+"/questions", tc.body, []*http.Cookie{cookie})
			if rec.Code != tc.want {
				t.Errorf("status esperado %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCreateQuestionOpenText(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventForQuestions(t, h, cookie)

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+id+"/questions", map[string]any{
		"title": "Qual sua comida favorita?", "type": "OPEN_TEXT",
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Question questionDTO `json:"question"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Question.Type != "OPEN_TEXT" {
		t.Errorf("tipo divergente: %+v", resp.Question)
	}
	if len(resp.Question.Options) != 0 {
		t.Errorf("pergunta de resposta aberta não deveria ter opções: %+v", resp.Question.Options)
	}
}

func TestCreateQuestionOpenTextIgnoresOptions(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventForQuestions(t, h, cookie)

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+id+"/questions", map[string]any{
		"title": "Comentário livre", "type": "OPEN_TEXT", "options": []string{"Ignorada"},
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Question questionDTO `json:"question"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Question.Options) != 0 {
		t.Errorf("opções enviadas para OPEN_TEXT deveriam ser ignoradas: %+v", resp.Question.Options)
	}
}

func TestCreateQuestionOwnership(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookieAna := registerUser(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/auth/register", map[string]string{
		"name": "Bia", "email": "bia@exemplo.com", "password": "segredo",
	}, nil)
	cookieBia := sessionCookie(t, rec)

	id := createEventForQuestions(t, h, cookieAna)

	cases := []struct {
		name   string
		path   string
		cookie *http.Cookie
		want   int
	}{
		{"evento de outro dono", "/api/events/" + id + "/questions", cookieBia, http.StatusNotFound},
		{"evento inexistente", "/api/events/999999/questions", cookieAna, http.StatusNotFound},
		{"id de evento inválido", "/api/events/abc/questions", cookieAna, http.StatusBadRequest},
		{"sem autenticação", "/api/events/" + id + "/questions", nil, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doJSON(t, h, http.MethodPost, tc.path, map[string]any{
				"title": "P", "type": "GROUP", "options": []string{"A", "B"},
			}, []*http.Cookie{tc.cookie})
			if rec.Code != tc.want {
				t.Errorf("status esperado %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestListQuestions(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventForQuestions(t, h, cookie)

	doJSON(t, h, http.MethodPost, "/api/events/"+id+"/questions", map[string]any{
		"title": "Primeira", "type": "GROUP", "options": []string{"A", "B"},
	}, []*http.Cookie{cookie})
	doJSON(t, h, http.MethodPost, "/api/events/"+id+"/questions", map[string]any{
		"title": "Segunda", "type": "INDIVIDUAL", "options": []string{"Única"},
	}, []*http.Cookie{cookie})

	rec := doJSON(t, h, http.MethodGet, "/api/events/"+id+"/questions", nil, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Questions []questionDTO `json:"questions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Questions) != 2 {
		t.Fatalf("esperado 2 perguntas, got %d", len(resp.Questions))
	}
	if resp.Questions[0].Title != "Primeira" || resp.Questions[1].Title != "Segunda" {
		t.Errorf("ordem divergente: %+v", resp.Questions)
	}
	if len(resp.Questions[0].Options) != 2 || resp.Questions[1].Type != "INDIVIDUAL" {
		t.Errorf("perguntas divergentes: %+v", resp.Questions)
	}
}

func TestListQuestionsEmpty(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventForQuestions(t, h, cookie)

	rec := doJSON(t, h, http.MethodGet, "/api/events/"+id+"/questions", nil, []*http.Cookie{cookie})
	var resp struct {
		Questions []questionDTO `json:"questions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Questions == nil || len(resp.Questions) != 0 {
		t.Errorf("esperado lista vazia, got %+v", resp.Questions)
	}
}

func TestDeleteQuestion(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventForQuestions(t, h, cookie)

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+id+"/questions", map[string]any{
		"title": "Para remover", "type": "GROUP", "options": []string{"A", "B"},
	}, []*http.Cookie{cookie})
	var created struct {
		Question questionDTO `json:"question"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	rec = doJSON(t, h, http.MethodDelete, "/api/events/"+id+"/questions/"+created.Question.ID, nil, []*http.Cookie{cookie})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status esperado 204, got %d: %s", rec.Code, rec.Body.String())
	}

	list := doJSON(t, h, http.MethodGet, "/api/events/"+id+"/questions", nil, []*http.Cookie{cookie})
	var resp struct {
		Questions []questionDTO `json:"questions"`
	}
	_ = json.Unmarshal(list.Body.Bytes(), &resp)
	if len(resp.Questions) != 0 {
		t.Errorf("pergunta deveria ter sido removida: %+v", resp.Questions)
	}
}

func TestDeleteQuestionErrors(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventForQuestions(t, h, cookie)

	cases := []struct {
		name string
		path string
		want int
	}{
		{"pergunta inexistente", "/api/events/" + id + "/questions/999999", http.StatusNotFound},
		{"id de pergunta inválido", "/api/events/" + id + "/questions/abc", http.StatusBadRequest},
		{"evento inexistente", "/api/events/999999/questions/1", http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doJSON(t, h, http.MethodDelete, tc.path, nil, []*http.Cookie{cookie})
			if rec.Code != tc.want {
				t.Errorf("status esperado %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func createQuestionsForReorder(t *testing.T, h http.Handler, cookie *http.Cookie, id string) []string {
	t.Helper()
	ids := make([]string, 0, 3)
	for _, title := range []string{"Primeira", "Segunda", "Terceira"} {
		rec := doJSON(t, h, http.MethodPost, "/api/events/"+id+"/questions", map[string]any{
			"title": title, "type": "GROUP", "options": []string{"A", "B"},
		}, []*http.Cookie{cookie})
		var resp struct {
			Question questionDTO `json:"question"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode pergunta: %v", err)
		}
		ids = append(ids, resp.Question.ID)
	}
	return ids
}

func listQuestionTitles(t *testing.T, h http.Handler, cookie *http.Cookie, id string) []string {
	t.Helper()
	rec := doJSON(t, h, http.MethodGet, "/api/events/"+id+"/questions", nil, []*http.Cookie{cookie})
	var resp struct {
		Questions []questionDTO `json:"questions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	titles := make([]string, 0, len(resp.Questions))
	for _, q := range resp.Questions {
		titles = append(titles, q.Title)
	}
	return titles
}

func TestUpdateQuestion(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventForQuestions(t, h, cookie)

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+id+"/questions", map[string]any{
		"title": "Antes", "type": "GROUP", "layoutView": "CENTER", "options": []string{"A", "B"},
	}, []*http.Cookie{cookie})
	var created struct {
		Question questionDTO `json:"question"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	rec = doJSON(t, h, http.MethodPatch, "/api/events/"+id+"/questions/"+created.Question.ID, map[string]any{
		"title": "Depois", "options": []string{"X", "Y", "Z"},
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Question questionDTO `json:"question"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Question.Title != "Depois" || resp.Question.Type != "GROUP" || resp.Question.LayoutView != "CENTER" {
		t.Errorf("pergunta atualizada divergente: %+v", resp.Question)
	}
	if len(resp.Question.Options) != 3 || resp.Question.Options[0].Text != "X" || resp.Question.Options[2].Text != "Z" {
		t.Errorf("opções divergentes: %+v", resp.Question.Options)
	}
	if resp.Question.OrderIndex != created.Question.OrderIndex {
		t.Errorf("ordem deveria ser preservada, got %d", resp.Question.OrderIndex)
	}

	got := listQuestionTitles(t, h, cookie, id)
	if len(got) != 1 || got[0] != "Depois" {
		t.Errorf("alteração não persistida: %v", got)
	}
}

func TestUpdateQuestionOpenText(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventForQuestions(t, h, cookie)

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+id+"/questions", map[string]any{
		"title": "Antes", "type": "OPEN_TEXT",
	}, []*http.Cookie{cookie})
	var created struct {
		Question questionDTO `json:"question"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	rec = doJSON(t, h, http.MethodPatch, "/api/events/"+id+"/questions/"+created.Question.ID, map[string]any{
		"title": "Depois",
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Question questionDTO `json:"question"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Question.Title != "Depois" || resp.Question.Type != "OPEN_TEXT" || len(resp.Question.Options) != 0 {
		t.Errorf("pergunta atualizada divergente: %+v", resp.Question)
	}
}

func TestUpdateQuestionValidation(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventForQuestions(t, h, cookie)

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+id+"/questions", map[string]any{
		"title": "Minha", "type": "GROUP", "options": []string{"A", "B"},
	}, []*http.Cookie{cookie})
	var created struct {
		Question questionDTO `json:"question"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	otherID := createEventForQuestions(t, h, cookie)

	cases := []struct {
		name string
		path string
		body map[string]any
		want int
	}{
		{"sem título", "/api/events/" + id + "/questions/" + created.Question.ID, map[string]any{"title": "", "options": []string{"A", "B"}}, http.StatusBadRequest},
		{"título vazio", "/api/events/" + id + "/questions/" + created.Question.ID, map[string]any{"title": "  ", "options": []string{"A", "B"}}, http.StatusBadRequest},
		{"menos de 2 opções", "/api/events/" + id + "/questions/" + created.Question.ID, map[string]any{"title": "P", "options": []string{"A"}}, http.StatusBadRequest},
		{"opções duplicadas", "/api/events/" + id + "/questions/" + created.Question.ID, map[string]any{"title": "P", "options": []string{"A", "A"}}, http.StatusBadRequest},
		{"pergunta inexistente", "/api/events/" + id + "/questions/999999", map[string]any{"title": "P", "options": []string{"A", "B"}}, http.StatusNotFound},
		{"pergunta de outro evento", "/api/events/" + otherID + "/questions/" + created.Question.ID, map[string]any{"title": "P", "options": []string{"A", "B"}}, http.StatusNotFound},
		{"id inválido", "/api/events/" + id + "/questions/abc", map[string]any{"title": "P", "options": []string{"A", "B"}}, http.StatusBadRequest},
		{"sem autenticação", "/api/events/" + id + "/questions/" + created.Question.ID, map[string]any{"title": "P", "options": []string{"A", "B"}}, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var cookies []*http.Cookie
			if tc.want != http.StatusUnauthorized {
				cookies = []*http.Cookie{cookie}
			}
			rec := doJSON(t, h, http.MethodPatch, tc.path, tc.body, cookies)
			if rec.Code != tc.want {
				t.Errorf("status esperado %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}

	// título original deve permanecer após as falhas
	got := listQuestionTitles(t, h, cookie, id)
	if len(got) != 1 || got[0] != "Minha" {
		t.Errorf("pergunta não deveria mudar após validações: %v", got)
	}
}

func TestReorderQuestions(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventForQuestions(t, h, cookie)
	ids := createQuestionsForReorder(t, h, cookie, id)

	rec := doJSON(t, h, http.MethodPut, "/api/events/"+id+"/questions/order", map[string]any{
		"questionIds": []string{ids[2], ids[0], ids[1]},
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status esperado 204, got %d: %s", rec.Code, rec.Body.String())
	}

	got := listQuestionTitles(t, h, cookie, id)
	want := []string{"Terceira", "Primeira", "Segunda"}
	for i, title := range want {
		if got[i] != title {
			t.Errorf("ordem esperada %v, got %v", want, got)
			break
		}
	}
}

func TestReorderQuestionsValidation(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	id := createEventForQuestions(t, h, cookie)
	ids := createQuestionsForReorder(t, h, cookie, id)

	// pergunta de outro evento
	otherID := createEventForQuestions(t, h, cookie)
	otherQuestion := createQuestionsForReorder(t, h, cookie, otherID)

	cases := []struct {
		name string
		body map[string]any
		want int
	}{
		{"lista vazia com perguntas existentes", map[string]any{"questionIds": []string{}}, http.StatusBadRequest},
		{"contagem errada", map[string]any{"questionIds": []string{ids[0]}}, http.StatusBadRequest},
		{"id repetido", map[string]any{"questionIds": []string{ids[0], ids[0], ids[1]}}, http.StatusBadRequest},
		{"id de outro evento", map[string]any{"questionIds": []string{otherQuestion[0], ids[0], ids[1]}}, http.StatusBadRequest},
		{"id não numérico", map[string]any{"questionIds": []string{"abc", ids[0], ids[1]}}, http.StatusBadRequest},
		{"id inexistente", map[string]any{"questionIds": []string{"999999", ids[0], ids[1]}}, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doJSON(t, h, http.MethodPut, "/api/events/"+id+"/questions/order", tc.body, []*http.Cookie{cookie})
			if rec.Code != tc.want {
				t.Errorf("status esperado %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}

	// ordem original deve permanecer após as falhas
	got := listQuestionTitles(t, h, cookie, id)
	if got[0] != "Primeira" || got[1] != "Segunda" || got[2] != "Terceira" {
		t.Errorf("ordem não deveria mudar após validações, got %v", got)
	}
}

func TestReorderQuestionsOwnershipAndAuth(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookieAna := registerUser(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/auth/register", map[string]string{
		"name": "Bia", "email": "bia@exemplo.com", "password": "segredo",
	}, nil)
	cookieBia := sessionCookie(t, rec)

	id := createEventForQuestions(t, h, cookieAna)
	ids := createQuestionsForReorder(t, h, cookieAna, id)

	cases := []struct {
		name   string
		path   string
		cookie *http.Cookie
		want   int
	}{
		{"evento de outro dono", "/api/events/" + id + "/questions/order", cookieBia, http.StatusNotFound},
		{"evento inexistente", "/api/events/999999/questions/order", cookieAna, http.StatusNotFound},
		{"sem autenticação", "/api/events/" + id + "/questions/order", nil, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doJSON(t, h, http.MethodPut, tc.path, map[string]any{"questionIds": ids}, []*http.Cookie{tc.cookie})
			if rec.Code != tc.want {
				t.Errorf("status esperado %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}
