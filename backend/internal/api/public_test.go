package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func setupPublicEvent(t *testing.T, h http.Handler) (cookie *http.Cookie, eventID string, groupQID, openQID, optAID, optBID string) {
	t.Helper()
	cookie = registerUser(t, h)
	eventID = createEventAndGetID(t, h, cookie, "Evento Público")

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/questions", map[string]any{
		"title": "Qual sua linguagem favorita?", "type": "GROUP", "options": []string{"Go", "JS"},
	}, []*http.Cookie{cookie})
	var created struct {
		Question questionDTO `json:"question"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode pergunta de grupo: %v", err)
	}
	groupQID = created.Question.ID
	optAID = created.Question.Options[0].ID
	optBID = created.Question.Options[1].ID

	rec = doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/questions", map[string]any{
		"title": "Qual sua comida favorita?", "type": "OPEN_TEXT",
	}, []*http.Cookie{cookie})
	var createdOpen struct {
		Question questionDTO `json:"question"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createdOpen); err != nil {
		t.Fatalf("decode pergunta aberta: %v", err)
	}
	openQID = createdOpen.Question.ID

	rec = doJSON(t, h, http.MethodPatch, "/api/events/"+eventID, map[string]any{
		"title": "Evento Público", "status": "OPEN_FOR_ANSWERS", "allowEdit": true,
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("abrir respostas: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	return cookie, eventID, groupQID, openQID, optAID, optBID
}

func TestPublicGetEventNotFound(t *testing.T) {
	h := newTestAPI(t).Handler()
	rec := doJSON(t, h, http.MethodGet, "/api/public/events/999999", nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status esperado 404, got %d", rec.Code)
	}
}

func TestPublicResolvePINNotFound(t *testing.T) {
	h := newTestAPI(t).Handler()
	rec := doJSON(t, h, http.MethodGet, "/api/public/events/by-pin?pin=NAOEXISTE", nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status esperado 404, got %d", rec.Code)
	}
}

func TestPublicResolvePINEmpty(t *testing.T) {
	h := newTestAPI(t).Handler()
	rec := doJSON(t, h, http.MethodGet, "/api/public/events/by-pin?pin=", nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status esperado 404, got %d", rec.Code)
	}
}

func TestPublicResolvePINFindsEvent(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	rec := doJSON(t, h, http.MethodPost, "/api/events", map[string]string{
		"title": "Evento com PIN", "pinCode": "ACHAME123",
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar evento: status %d: %s", rec.Code, rec.Body.String())
	}
	var created struct {
		Event eventDTO `json:"event"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode evento: %v", err)
	}

	rec = doJSON(t, h, http.MethodGet, "/api/public/events/by-pin?pin=achame123", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode resposta: %v", err)
	}
	if resp.ID != created.Event.ID {
		t.Errorf("id esperado %s, got %s", created.Event.ID, resp.ID)
	}
}

func TestPublicGetEvent(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)

	rec := doJSON(t, h, http.MethodGet, "/api/public/events/"+eventID, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// não deve vazar PIN nem dono
	if body := rec.Body.String(); strings.Contains(body, "pinCode") || strings.Contains(body, "ownerId") {
		t.Errorf("resposta pública não deveria conter pinCode/ownerId: %s", body)
	}

	var resp struct {
		Event     publicEventDTO      `json:"event"`
		Questions []publicQuestionDTO `json:"questions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Event.ID != eventID || resp.Event.Title != "Evento Público" {
		t.Errorf("evento divergente: %+v", resp.Event)
	}
	if len(resp.Questions) != 2 {
		t.Fatalf("esperado 2 perguntas, got %d", len(resp.Questions))
	}
	if resp.Questions[0].ID != groupQID || len(resp.Questions[0].Options) != 2 || resp.Questions[0].Options[0].ID != optAID {
		t.Errorf("primeira pergunta divergente: %+v", resp.Questions[0])
	}
	if resp.Questions[1].ID != openQID || resp.Questions[1].Type != "OPEN_TEXT" || len(resp.Questions[1].Options) != 0 {
		t.Errorf("segunda pergunta divergente: %+v", resp.Questions[1])
	}
	if !resp.Event.AnswersOpen {
		t.Error("answersOpen deveria ser true após abrir as respostas")
	}
}

func TestPublicGetEventAnswersClosedByDefault(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	eventID := createEventAndGetID(t, h, cookie, "Evento recém-criado")

	rec := doJSON(t, h, http.MethodGet, "/api/public/events/"+eventID, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Event publicEventDTO `json:"event"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Event.AnswersOpen {
		t.Error("answersOpen deveria ser false por padrão (evento em PREPARATION)")
	}
}

func TestSubmitAnswersEventNotFound(t *testing.T) {
	h := newTestAPI(t).Handler()
	rec := doJSON(t, h, http.MethodPost, "/api/public/events/999999/submit", map[string]any{
		"email": "a@b.com", "answers": []map[string]string{},
	}, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status esperado 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSubmitAnswersNoQuestions(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	eventID := createEventAndGetID(t, h, cookie, "Evento vazio")

	rec := doJSON(t, h, http.MethodPatch, "/api/events/"+eventID, map[string]any{
		"title": "Evento vazio", "status": "OPEN_FOR_ANSWERS",
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("abrir respostas: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "a@b.com", "answers": []map[string]string{},
	}, nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status esperado 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSubmitAnswersRejectedWhenNotOpen(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie := registerUser(t, h)
	eventID := createEventAndGetID(t, h, cookie, "Evento fechado")

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/questions", map[string]any{
		"title": "P", "type": "GROUP", "options": []string{"A", "B"},
	}, []*http.Cookie{cookie})
	var created struct {
		Question questionDTO `json:"question"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode pergunta: %v", err)
	}

	rec = doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "a@b.com",
		"answers": []map[string]string{
			{"questionId": created.Question.ID, "optionId": created.Question.Options[0].ID},
		},
	}, nil)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status esperado 403 (respostas fechadas), got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSubmitAnswersSuccess(t *testing.T) {
	app := newTestAPI(t)
	h := app.Handler()
	_, eventID, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "Participante@Exemplo.com",
		"name":  "Participante",
		"photo": "",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "  Pizza  "},
		},
	}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	id, err := strconv.ParseInt(eventID, 10, 64)
	if err != nil {
		t.Fatalf("parse eventID: %v", err)
	}
	p, err := app.store.FindParticipantByEventAndEmail(context.Background(), id, "participante@exemplo.com")
	if err != nil {
		t.Fatalf("participante não encontrado: %v", err)
	}
	if p.Email != "participante@exemplo.com" {
		t.Errorf("email deveria ser normalizado, got %s", p.Email)
	}
}

func TestSubmitAnswersReusesParticipant(t *testing.T) {
	app := newTestAPI(t)
	h := app.Handler()
	_, eventID, groupQID, openQID, optAID, optBID := setupPublicEvent(t, h)

	body := func(optionID string) map[string]any {
		return map[string]any{
			"email": "ana@exemplo.com",
			"name":  "Ana",
			"answers": []map[string]string{
				{"questionId": groupQID, "optionId": optionID},
				{"questionId": openQID, "text": "Pizza"},
			},
		}
	}

	rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", body(optAID), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("primeiro envio: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	id, err := strconv.ParseInt(eventID, 10, 64)
	if err != nil {
		t.Fatalf("parse eventID: %v", err)
	}
	first, err := app.store.FindParticipantByEventAndEmail(context.Background(), id, "ana@exemplo.com")
	if err != nil {
		t.Fatalf("participante não encontrado após primeiro envio: %v", err)
	}

	rec = doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", body(optBID), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("reenvio: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	second, err := app.store.FindParticipantByEventAndEmail(context.Background(), id, "ana@exemplo.com")
	if err != nil {
		t.Fatalf("participante não encontrado após reenvio: %v", err)
	}
	if first.ID != second.ID {
		t.Errorf("reenvio deveria reaproveitar o mesmo participante, got %d e %d", first.ID, second.ID)
	}
}

func TestSubmitAnswersReturnsEditToken(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Pizza"},
		},
	}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		OK        bool   `json:"ok"`
		EditToken string `json:"editToken"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.EditToken == "" {
		t.Error("editToken não deveria vir vazio no envio inicial")
	}
}

func TestSubmitAnswersWithEditTokenUpdatesSameParticipant(t *testing.T) {
	app := newTestAPI(t)
	h := app.Handler()
	_, eventID, groupQID, openQID, optAID, optBID := setupPublicEvent(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Pizza"},
		},
	}, nil)
	var first struct {
		EditToken string `json:"editToken"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &first); err != nil {
		t.Fatalf("decode: %v", err)
	}

	id, err := strconv.ParseInt(eventID, 10, 64)
	if err != nil {
		t.Fatalf("parse eventID: %v", err)
	}
	before, err := app.store.FindParticipantByEventAndEmail(context.Background(), id, "ana@exemplo.com")
	if err != nil {
		t.Fatalf("participante não encontrado: %v", err)
	}

	// reenvio via token, sem informar e-mail: deve atualizar o mesmo participante.
	rec = doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"editToken": first.EditToken,
		"name":      "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optBID},
			{"questionId": openQID, "text": "Sushi"},
		},
	}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("reenvio via token: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	after, err := app.store.FindParticipantByEventAndEmail(context.Background(), id, "ana@exemplo.com")
	if err != nil {
		t.Fatalf("participante não encontrado após reenvio: %v", err)
	}
	if before.ID != after.ID {
		t.Errorf("reenvio via token deveria reaproveitar o mesmo participante, got %d e %d", before.ID, after.ID)
	}

	answers, err := app.store.ListAnswersByParticipant(context.Background(), after.ID)
	if err != nil {
		t.Fatalf("listar respostas: %v", err)
	}
	optB, _ := strconv.ParseInt(optBID, 10, 64)
	found := false
	for _, a := range answers {
		if strconv.FormatInt(a.QuestionID, 10) == groupQID {
			found = true
			if a.OptionID != optB {
				t.Errorf("resposta não atualizada pelo reenvio via token: %+v", a)
			}
		}
	}
	if !found {
		t.Fatal("resposta de grupo não encontrada após reenvio via token")
	}
}

func TestSubmitAnswersBlockedWhenAllowEditDisabled(t *testing.T) {
	app := newTestAPI(t)
	h := app.Handler()
	cookie, eventID, groupQID, openQID, optAID, optBID := setupPublicEvent(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Pizza"},
		},
	}, nil)
	var first struct {
		EditToken string `json:"editToken"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &first); err != nil {
		t.Fatalf("decode: %v", err)
	}

	rec = doJSON(t, h, http.MethodPatch, "/api/events/"+eventID, map[string]any{
		"title": "Evento Público", "status": "OPEN_FOR_ANSWERS", "allowEdit": false,
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("desligar allowEdit: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// reenvio via token de edição: deve ser bloqueado.
	rec = doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"editToken": first.EditToken,
		"name":      "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optBID},
			{"questionId": openQID, "text": "Sushi"},
		},
	}, nil)
	if rec.Code != http.StatusForbidden {
		t.Errorf("reenvio via token com allowEdit=false: status esperado 403, got %d: %s", rec.Code, rec.Body.String())
	}

	// reenvio com o mesmo e-mail, sem token: também deve ser bloqueado.
	rec = doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optBID},
			{"questionId": openQID, "text": "Sushi"},
		},
	}, nil)
	if rec.Code != http.StatusForbidden {
		t.Errorf("reenvio por e-mail com allowEdit=false: status esperado 403, got %d: %s", rec.Code, rec.Body.String())
	}

	// primeiro envio de outra pessoa continua permitido: allowEdit só afeta reenvios.
	rec = doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "bia@exemplo.com",
		"name":  "Bia",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Massa"},
		},
	}, nil)
	if rec.Code != http.StatusOK {
		t.Errorf("primeiro envio com allowEdit=false: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSubmitAnswersWithInvalidEditToken(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"editToken": "token-que-nao-existe",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Pizza"},
		},
	}, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status esperado 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPublicGetParticipantByToken(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"photo": "data:image/jpeg;base64,Zm9vCg==",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Pizza"},
		},
	}, nil)
	var submitResp struct {
		EditToken string `json:"editToken"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &submitResp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	rec = doJSON(t, h, http.MethodGet, "/api/public/events/"+eventID+"/participant?token="+submitResp.EditToken, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Participant publicParticipantDTO `json:"participant"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Participant.Email != "ana@exemplo.com" {
		t.Errorf("email divergente: %+v", resp.Participant)
	}
	if !strings.HasPrefix(resp.Participant.Photo, "/api/photos/") {
		t.Errorf("foto deveria ser uma URL de arquivo, got %q", resp.Participant.Photo)
	}
	if len(resp.Participant.Answers) != 2 {
		t.Fatalf("esperado 2 respostas, got %+v", resp.Participant.Answers)
	}
}

func TestPublicGetParticipantInvalidToken(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, _, _, _, _ := setupPublicEvent(t, h)

	rec := doJSON(t, h, http.MethodGet, "/api/public/events/"+eventID+"/participant?token=inexistente", nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status esperado 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPublicGetParticipantWrongEvent(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID1, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)
	eventID2 := createEventAndGetID(t, h, cookie, "Segundo evento")

	rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID1+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Pizza"},
		},
	}, nil)
	var submitResp struct {
		EditToken string `json:"editToken"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &submitResp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	rec = doJSON(t, h, http.MethodGet, "/api/public/events/"+eventID2+"/participant?token="+submitResp.EditToken, nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("token de outro evento deveria ser 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSubmitAnswersValidation(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)

	valid := func() map[string]any {
		return map[string]any{
			"email": "ana@exemplo.com",
			"name":  "Ana",
			"answers": []map[string]string{
				{"questionId": groupQID, "optionId": optAID},
				{"questionId": openQID, "text": "Pizza"},
			},
		}
	}

	cases := []struct {
		name   string
		mutate func(m map[string]any)
		want   int
	}{
		{"email vazio", func(m map[string]any) { m["email"] = "" }, http.StatusBadRequest},
		{"email inválido", func(m map[string]any) { m["email"] = "não-é-email" }, http.StatusBadRequest},
		{"nome vazio", func(m map[string]any) { m["name"] = "" }, http.StatusBadRequest},
		{"faltando resposta", func(m map[string]any) {
			m["answers"] = []map[string]string{{"questionId": groupQID, "optionId": optAID}}
		}, http.StatusBadRequest},
		{"opção inválida", func(m map[string]any) {
			m["answers"] = []map[string]string{
				{"questionId": groupQID, "optionId": "999999"},
				{"questionId": openQID, "text": "Pizza"},
			}
		}, http.StatusBadRequest},
		{"texto aberto vazio (só espaços)", func(m map[string]any) {
			m["answers"] = []map[string]string{
				{"questionId": groupQID, "optionId": optAID},
				{"questionId": openQID, "text": "   "},
			}
		}, http.StatusBadRequest},
		{"resposta duplicada para mesma pergunta", func(m map[string]any) {
			m["answers"] = []map[string]string{
				{"questionId": groupQID, "optionId": optAID},
				{"questionId": groupQID, "optionId": optAID},
			}
		}, http.StatusBadRequest},
		{"foto com formato inválido", func(m map[string]any) { m["photo"] = "não-é-data-url" }, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := valid()
			tc.mutate(body)
			rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", body, nil)
			if rec.Code != tc.want {
				t.Errorf("status esperado %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}
