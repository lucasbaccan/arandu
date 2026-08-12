package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestListResponsesEmpty(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, _, _, _, _ := setupPublicEvent(t, h)

	rec := doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ParticipantCount int                      `json:"participantCount"`
		Participants     []participantResponseDTO `json:"participants"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ParticipantCount != 0 || len(resp.Participants) != 0 {
		t.Errorf("esperado 0 participantes, got %+v", resp)
	}
}

func TestListResponses(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, groupQID, openQID, optAID, optBID := setupPublicEvent(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Pizza"},
		},
	}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("envio de ana: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "bia@exemplo.com",
		"name":  "Bia",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optBID},
			{"questionId": openQID, "text": "Sushi"},
		},
	}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("envio de bia: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ParticipantCount int                      `json:"participantCount"`
		Participants     []participantResponseDTO `json:"participants"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ParticipantCount != 2 || len(resp.Participants) != 2 {
		t.Fatalf("esperado 2 participantes, got %+v", resp)
	}

	ana := resp.Participants[0]
	if ana.Email != "ana@exemplo.com" || len(ana.Answers) != 2 {
		t.Fatalf("participante ana divergente: %+v", ana)
	}
	for _, a := range ana.Answers {
		if a.QuestionID == groupQID {
			if a.QuestionTitle != "Qual sua linguagem favorita?" || a.OptionText != "Go" {
				t.Errorf("resposta de grupo divergente: %+v", a)
			}
		}
		if a.QuestionID == openQID {
			if a.Text != "Pizza" || a.OptionID != "" {
				t.Errorf("resposta aberta divergente: %+v", a)
			}
		}
	}
}

func TestListResponsesReflectsCurrentQuestionOrder(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)

	doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Pizza"},
		},
	}, nil)

	// reordena: pergunta aberta passa a vir primeiro
	rec := doJSON(t, h, http.MethodPut, "/api/events/"+eventID+"/questions/order", map[string]any{
		"questionIds": []string{openQID, groupQID},
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("reordenar perguntas: status esperado 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Participants []participantResponseDTO `json:"participants"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	answers := resp.Participants[0].Answers
	if len(answers) != 2 {
		t.Fatalf("esperado 2 respostas, got %d", len(answers))
	}
	if answers[0].QuestionID != openQID || answers[1].QuestionID != groupQID {
		t.Errorf("ordem das respostas deveria refletir a nova ordem das perguntas, got %+v", answers)
	}
}

func TestListResponsesRequiresOwnership(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, _, _, _, _ := setupPublicEvent(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/auth/register", map[string]string{
		"name": "Bia", "email": "bia-owner@exemplo.com", "password": "segredo",
	}, nil)
	otherCookie := sessionCookie(t, rec)

	rec = doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{otherCookie})
	if rec.Code != http.StatusNotFound {
		t.Errorf("status esperado 404, got %d", rec.Code)
	}

	rec = doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("sem autenticação: status esperado 401, got %d", rec.Code)
	}
}

func TestListResponsesIncludesEditToken(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)

	doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Pizza"},
		},
	}, nil)

	rec := doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	var resp struct {
		Participants []participantResponseDTO `json:"participants"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Participants[0].EditToken == "" {
		t.Error("editToken não deveria vir vazio na listagem do organizador")
	}
}

func TestUpdateParticipantPhoto(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)

	doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Pizza"},
		},
	}, nil)

	rec := doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	var resp struct {
		Participants []participantResponseDTO `json:"participants"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	participantID := resp.Participants[0].ID

	rec = doJSON(t, h, http.MethodPatch,
		"/api/events/"+eventID+"/responses/"+participantID+"/photo",
		map[string]any{"photo": "data:image/jpeg;base64,Zm9vCg=="}, []*http.Cookie{cookie},
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Participants[0].Photo != "data:image/jpeg;base64,Zm9vCg==" {
		t.Errorf("foto não atualizada: %+v", resp.Participants[0])
	}
}

func TestUpdateParticipantPhotoValidation(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)

	doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Pizza"},
		},
	}, nil)

	rec := doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	var resp struct {
		Participants []participantResponseDTO `json:"participants"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	participantID := resp.Participants[0].ID

	cases := []struct {
		name          string
		participantID string
		body          map[string]any
		cookie        *http.Cookie
		want          int
	}{
		{"formato inválido", participantID, map[string]any{"photo": "não-é-data-url"}, cookie, http.StatusBadRequest},
		{"participante inexistente", "999999", map[string]any{"photo": "data:image/jpeg;base64,Zm9vCg=="}, cookie, http.StatusNotFound},
		{"sem autenticação", participantID, map[string]any{"photo": "data:image/jpeg;base64,Zm9vCg=="}, nil, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var cookies []*http.Cookie
			if tc.cookie != nil {
				cookies = []*http.Cookie{tc.cookie}
			}
			rec := doJSON(t, h, http.MethodPatch,
				"/api/events/"+eventID+"/responses/"+tc.participantID+"/photo",
				tc.body, cookies,
			)
			if rec.Code != tc.want {
				t.Errorf("status esperado %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestUpdateParticipantPhotoRequiresOwnership(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)

	doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Pizza"},
		},
	}, nil)

	rec := doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	var resp struct {
		Participants []participantResponseDTO `json:"participants"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	participantID := resp.Participants[0].ID

	rec = doJSON(t, h, http.MethodPost, "/api/auth/register", map[string]string{
		"name": "Bia", "email": "bia-owner2@exemplo.com", "password": "segredo",
	}, nil)
	otherCookie := sessionCookie(t, rec)

	rec = doJSON(t, h, http.MethodPatch,
		"/api/events/"+eventID+"/responses/"+participantID+"/photo",
		map[string]any{"photo": "data:image/jpeg;base64,Zm9vCg=="}, []*http.Cookie{otherCookie},
	)
	if rec.Code != http.StatusNotFound {
		t.Errorf("dono diferente: esperado 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateAnswerOptionQuestion(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, groupQID, openQID, optAID, optBID := setupPublicEvent(t, h)

	doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Pizza"},
		},
	}, nil)

	rec := doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	var resp struct {
		Participants []participantResponseDTO `json:"participants"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	participantID := resp.Participants[0].ID

	rec = doJSON(t, h, http.MethodPatch,
		"/api/events/"+eventID+"/responses/"+participantID+"/answers/"+groupQID,
		map[string]any{"optionId": optBID}, []*http.Cookie{cookie},
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	found := false
	for _, a := range resp.Participants[0].Answers {
		if a.QuestionID == groupQID {
			found = true
			if a.OptionID != optBID || a.OptionText != "JS" {
				t.Errorf("resposta não atualizada corretamente: %+v", a)
			}
		}
	}
	if !found {
		t.Fatal("resposta da pergunta de grupo não encontrada após atualização")
	}
}

func TestUpdateAnswerOpenTextCensor(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)

	doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Texto ofensivo"},
		},
	}, nil)

	rec := doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	var resp struct {
		Participants []participantResponseDTO `json:"participants"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	participantID := resp.Participants[0].ID

	rec = doJSON(t, h, http.MethodPatch,
		"/api/events/"+eventID+"/responses/"+participantID+"/answers/"+openQID,
		map[string]any{"text": "[removido pelo organizador]"}, []*http.Cookie{cookie},
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	for _, a := range resp.Participants[0].Answers {
		if a.QuestionID == openQID && a.Text != "[removido pelo organizador]" {
			t.Errorf("texto não censurado: %+v", a)
		}
	}
}

func TestUpdateAnswerValidation(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)

	doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email": "ana@exemplo.com",
		"name":  "Ana",
		"answers": []map[string]string{
			{"questionId": groupQID, "optionId": optAID},
			{"questionId": openQID, "text": "Pizza"},
		},
	}, nil)

	rec := doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	var resp struct {
		Participants []participantResponseDTO `json:"participants"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	participantID := resp.Participants[0].ID

	cases := []struct {
		name          string
		participantID string
		questionID    string
		body          map[string]any
		cookie        *http.Cookie
		want          int
	}{
		{"opção inválida", participantID, groupQID, map[string]any{"optionId": "999999"}, cookie, http.StatusBadRequest},
		{"opção faltando", participantID, groupQID, map[string]any{}, cookie, http.StatusBadRequest},
		{"participante inexistente", "999999", groupQID, map[string]any{"optionId": optAID}, cookie, http.StatusNotFound},
		{"pergunta inexistente", participantID, "999999", map[string]any{"optionId": optAID}, cookie, http.StatusNotFound},
		{"sem autenticação", participantID, groupQID, map[string]any{"optionId": optAID}, nil, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var cookies []*http.Cookie
			if tc.cookie != nil {
				cookies = []*http.Cookie{tc.cookie}
			}
			rec := doJSON(t, h, http.MethodPatch,
				"/api/events/"+eventID+"/responses/"+tc.participantID+"/answers/"+tc.questionID,
				tc.body, cookies,
			)
			if rec.Code != tc.want {
				t.Errorf("status esperado %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}
