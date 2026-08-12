package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"devopsconecta/backend/internal/auth"
)

var liveEventCounter int

// setupLiveEvent cria um evento com PIN conhecido, uma pergunta de grupo
// (Go/JS) e dois participantes que já responderam (ana -> Go, bia -> JS).
// Dono e PIN são únicos a cada chamada, pra poder criar mais de um evento no
// mesmo teste sem colidir (e-mail de dono repetido / PIN já em uso).
func setupLiveEvent(t *testing.T, h http.Handler) (cookie *http.Cookie, eventID, questionID, optAID, optBID, pin, anaID, biaID string) {
	t.Helper()
	liveEventCounter++
	n := liveEventCounter

	regRec := doJSON(t, h, http.MethodPost, "/api/auth/register", map[string]string{
		"name": "Organizador", "email": fmt.Sprintf("organizador-live-%d@exemplo.com", n), "password": "segredo",
	}, nil)
	if regRec.Code != http.StatusCreated {
		t.Fatalf("registrar dono do evento: status %d: %s", regRec.Code, regRec.Body.String())
	}
	cookie = sessionCookie(t, regRec)
	pin = fmt.Sprintf("LIVEPIN%d", n)

	rec := createEvent(t, h, cookie, map[string]string{"title": "Evento Live", "pinCode": pin})
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar evento: status %d: %s", rec.Code, rec.Body.String())
	}
	var created struct {
		Event eventDTO `json:"event"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode evento: %v", err)
	}
	eventID = created.Event.ID

	rec = doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/questions", map[string]any{
		"title": "Qual sua linguagem favorita?", "type": "GROUP", "options": []string{"Go", "JS"},
	}, []*http.Cookie{cookie})
	var q struct {
		Question questionDTO `json:"question"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &q); err != nil {
		t.Fatalf("decode pergunta: %v", err)
	}
	questionID = q.Question.ID
	optAID = q.Question.Options[0].ID
	optBID = q.Question.Options[1].ID

	rec = doJSON(t, h, http.MethodPatch, "/api/events/"+eventID, map[string]any{
		"title": "Evento Live", "status": "OPEN_FOR_ANSWERS",
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("abrir respostas: status %d: %s", rec.Code, rec.Body.String())
	}

	doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email":   "ana@exemplo.com",
		"name":    "Ana",
		"answers": []map[string]string{{"questionId": questionID, "optionId": optAID}},
	}, nil)
	doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
		"email":   "bia@exemplo.com",
		"name":    "Bia",
		"answers": []map[string]string{{"questionId": questionID, "optionId": optBID}},
	}, nil)

	rec = doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	var resp struct {
		Participants []participantResponseDTO `json:"participants"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode responses: %v", err)
	}
	for _, p := range resp.Participants {
		switch p.Email {
		case "ana@exemplo.com":
			anaID = p.ID
		case "bia@exemplo.com":
			biaID = p.ID
		}
	}
	if anaID == "" || biaID == "" {
		t.Fatalf("participantes nao encontrados: %+v", resp.Participants)
	}
	return
}

func liveJoin(t *testing.T, h http.Handler, eventID, pin, email string) (token, role string, code int) {
	t.Helper()
	rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/live/join", map[string]string{
		"pinCode": pin, "email": email,
	}, nil)
	if rec.Code != http.StatusOK {
		return "", "", rec.Code
	}
	var resp struct {
		Token string `json:"token"`
		Role  string `json:"role"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode join: %v", err)
	}
	return resp.Token, resp.Role, rec.Code
}

func getLiveState(t *testing.T, h http.Handler, eventID, token string) liveSnapshotDTO {
	t.Helper()
	rec := doJSON(t, h, http.MethodGet, "/api/public/events/"+eventID+"/live/state?token="+token, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("state: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var snap liveSnapshotDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	return snap
}

func TestLiveJoinParticipant(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, _, _, _, pin, anaID, _ := setupLiveEvent(t, h)

	token, role, code := liveJoin(t, h, eventID, pin, "ana@exemplo.com")
	if code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d", code)
	}
	if role != auth.LiveRoleParticipant {
		t.Errorf("role esperado participante, got %s", role)
	}
	claims, err := auth.VerifyLiveViewerToken("test-secret", token)
	if err != nil {
		t.Fatalf("verificar token: %v", err)
	}
	if claims.ParticipantID != anaID {
		t.Errorf("participantId esperado %s, got %s", anaID, claims.ParticipantID)
	}

	// e-mail com espaço/maiúscula diferente ainda deve reconhecer o mesmo participante
	_, role2, code2 := liveJoin(t, h, eventID, pin, "  ANA@Exemplo.com  ")
	if code2 != http.StatusOK || role2 != auth.LiveRoleParticipant {
		t.Errorf("join com e-mail variando maiuscula/espaco deveria reconhecer participante, got role=%s code=%d", role2, code2)
	}
}

func TestLiveJoinObserverWhenEmailNotAParticipant(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)

	token, role, code := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")
	if code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d", code)
	}
	if role != auth.LiveRoleObserver {
		t.Errorf("role esperado observador, got %s", role)
	}
	claims, err := auth.VerifyLiveViewerToken("test-secret", token)
	if err != nil {
		t.Fatalf("verificar token: %v", err)
	}
	if claims.ParticipantID != "" {
		t.Errorf("observador nao deveria ter participantId, got %s", claims.ParticipantID)
	}
}

func TestLiveJoinAsGuestWithoutEmail(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)

	token, role, code := liveJoin(t, h, eventID, pin, "")
	if code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d", code)
	}
	if role != auth.LiveRoleObserver {
		t.Errorf("role esperado observador, got %s", role)
	}
	claims, err := auth.VerifyLiveViewerToken("test-secret", token)
	if err != nil {
		t.Fatalf("verificar token: %v", err)
	}
	if claims.ParticipantID != "" {
		t.Errorf("convidado nao deveria ter participantId, got %s", claims.ParticipantID)
	}
}

func getLiveAdminState(t *testing.T, h http.Handler, eventID string, cookie *http.Cookie) liveAdminSnapshotDTO {
	t.Helper()
	rec := doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/live/state", nil, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("admin state: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var snap liveAdminSnapshotDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
		t.Fatalf("decode admin state: %v", err)
	}
	return snap
}

func TestLiveSetBlankedAppearsInPublicSnapshot(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)
	token, _, _ := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/blank", map[string]bool{"blanked": true}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("blank: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if snap := getLiveState(t, h, eventID, token); !snap.Blanked {
		t.Fatalf("esperava blanked=true no snapshot público, got %+v", snap)
	}

	rec = doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/blank", map[string]bool{"blanked": false}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("unblank: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if snap := getLiveState(t, h, eventID, token); snap.Blanked {
		t.Fatalf("esperava blanked=false no snapshot público, got %+v", snap)
	}
}

func TestLiveSetAnswersHiddenAppearsInPublicSnapshot(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)
	token, _, _ := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/hide-answers", map[string]bool{"hidden": true}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("hide-answers: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if snap := getLiveState(t, h, eventID, token); !snap.AnswersHidden {
		t.Fatalf("esperava answersHidden=true no snapshot público, got %+v", snap)
	}

	rec = doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/hide-answers", map[string]bool{"hidden": false}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("show-answers: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if snap := getLiveState(t, h, eventID, token); snap.AnswersHidden {
		t.Fatalf("esperava answersHidden=false no snapshot público, got %+v", snap)
	}
}

func TestLiveSetMessageAppearsInSnapshotAndClears(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)
	token, _, _ := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/message", map[string]string{"message": "Voltamos em 5 min"}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("mensagem: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if snap := getLiveState(t, h, eventID, token); snap.Message != "Voltamos em 5 min" {
		t.Fatalf("esperava mensagem no snapshot público, got %+v", snap)
	}

	rec = doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/message", map[string]string{"message": ""}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("limpar mensagem: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if snap := getLiveState(t, h, eventID, token); snap.Message != "" {
		t.Fatalf("esperava mensagem limpa no snapshot público, got %+v", snap)
	}
}

func TestLiveSetMessageRejectsTooLong(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, _, _, _, _, _, _ := setupLiveEvent(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/message", map[string]string{
		"message": strings.Repeat("a", maxLiveMessageLength+1),
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("mensagem muito longa: status esperado 400, got %d", rec.Code)
	}
}

func TestLiveSetInteractionsDefaultsTrueAndToggles(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)
	token, _, _ := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")

	if snap := getLiveState(t, h, eventID, token); !snap.InteractionsEnabled {
		t.Fatalf("esperava interações ligadas por padrão, got %+v", snap)
	}

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/interactions", map[string]bool{"enabled": false}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("desligar interações: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if snap := getLiveState(t, h, eventID, token); snap.InteractionsEnabled {
		t.Fatalf("esperava interações desligadas, got %+v", snap)
	}
}

func TestLiveBlankMessageInteractionsRequireOwnership(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, _, _, _, _, _, _ := setupLiveEvent(t, h)
	other := registerUser2(t, h)

	routes := []struct {
		path string
		body any
	}{
		{"/api/events/" + eventID + "/live/blank", map[string]bool{"blanked": true}},
		{"/api/events/" + eventID + "/live/hide-answers", map[string]bool{"hidden": true}},
		{"/api/events/" + eventID + "/live/message", map[string]string{"message": "oi"}},
		{"/api/events/" + eventID + "/live/interactions", map[string]bool{"enabled": false}},
	}
	for _, rt := range routes {
		rec := doJSON(t, h, http.MethodPost, rt.path, rt.body, []*http.Cookie{other})
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s dono errado: status esperado 404, got %d", rt.path, rec.Code)
		}
		rec = doJSON(t, h, http.MethodPost, rt.path, rt.body, nil)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s sem sessao: status esperado 401, got %d", rt.path, rec.Code)
		}
	}
}

func TestLiveReactRejectsUnlistedEmoji(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)
	token, _, _ := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")

	rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/live/react?token="+token, map[string]string{"emoji": "🍕"}, nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("emoji fora da lista: status esperado 400, got %d", rec.Code)
	}
}

func TestLiveReactAcceptsAllowedEmoji(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)
	token, _, _ := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")

	rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/live/react?token="+token, map[string]string{"emoji": "👍"}, nil)
	if rec.Code != http.StatusOK {
		t.Errorf("emoji permitido: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLiveReactRequiresInteractionsEnabled(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)
	token, _, _ := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/interactions", map[string]bool{"enabled": false}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("desligar interações: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/live/react?token="+token, map[string]string{"emoji": "👍"}, nil)
	if rec.Code != http.StatusForbidden {
		t.Errorf("interações desligadas: status esperado 403, got %d", rec.Code)
	}
}

func TestLiveSubmitQAPersistsAndVisibleToOrganizer(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)
	token, role, _ := liveJoin(t, h, eventID, pin, "ana@exemplo.com")
	if role != auth.LiveRoleParticipant {
		t.Fatalf("esperava ana como participante, got role=%s", role)
	}

	rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/live/qa?token="+token, map[string]string{"text": "Posso ir ao banheiro?"}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("submit qa: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	snap := getLiveAdminState(t, h, eventID, cookie)
	if len(snap.QAInbox) != 1 {
		t.Fatalf("esperava 1 mensagem na caixa, got %+v", snap.QAInbox)
	}
	if snap.QAInbox[0].Email != "ana@exemplo.com" || snap.QAInbox[0].Text != "Posso ir ao banheiro?" {
		t.Errorf("mensagem divergente: %+v", snap.QAInbox[0])
	}
}

func TestLiveSubmitQAGuestHasNoEmail(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)
	token, role, _ := liveJoin(t, h, eventID, pin, "")
	if role != auth.LiveRoleObserver {
		t.Fatalf("esperava convidado como observador, got role=%s", role)
	}

	rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/live/qa?token="+token, map[string]string{"text": "Oi!"}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("submit qa: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	snap := getLiveAdminState(t, h, eventID, cookie)
	if len(snap.QAInbox) != 1 || snap.QAInbox[0].Email != "" {
		t.Fatalf("esperava mensagem de convidado sem e-mail, got %+v", snap.QAInbox)
	}
}

func TestLiveSubmitQARejectsEmptyOrTooLong(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)
	token, _, _ := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")

	cases := []struct {
		name string
		text string
	}{
		{"vazio", ""},
		{"so espacos", "   "},
		{"muito longa", strings.Repeat("a", maxLiveQATextLength+1)},
	}
	for _, tc := range cases {
		rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/live/qa?token="+token, map[string]string{"text": tc.text}, nil)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status esperado 400, got %d", tc.name, rec.Code)
		}
	}
}

func TestLiveSubmitQARequiresInteractionsEnabled(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)
	token, _, _ := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/interactions", map[string]bool{"enabled": false}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("desligar interações: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/live/qa?token="+token, map[string]string{"text": "oi"}, nil)
	if rec.Code != http.StatusForbidden {
		t.Errorf("interações desligadas: status esperado 403, got %d", rec.Code)
	}

	snap := getLiveAdminState(t, h, eventID, cookie)
	if len(snap.QAInbox) != 0 {
		t.Fatalf("nao deveria ter gravado nada com interações desligadas, got %+v", snap.QAInbox)
	}
}

func TestLiveDismissQARemovesFromInbox(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)
	token, _, _ := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")

	doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/live/qa?token="+token, map[string]string{"text": "oi"}, nil)
	snap := getLiveAdminState(t, h, eventID, cookie)
	if len(snap.QAInbox) != 1 {
		t.Fatalf("esperava 1 mensagem antes de dispensar, got %+v", snap.QAInbox)
	}
	messageID := snap.QAInbox[0].ID

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/qa/"+messageID+"/dismiss", nil, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("dismiss: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	snap = getLiveAdminState(t, h, eventID, cookie)
	if len(snap.QAInbox) != 0 {
		t.Fatalf("esperava caixa vazia apos dispensar, got %+v", snap.QAInbox)
	}
}

func TestLiveDismissQARejectsMessageFromAnotherEvent(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookieA, eventA, _, _, _, pinA, _, _ := setupLiveEvent(t, h)
	cookieB, eventB, _, _, _, _, _, _ := setupLiveEvent(t, h)

	tokenA, _, _ := liveJoin(t, h, eventA, pinA, "curioso@exemplo.com")
	doJSON(t, h, http.MethodPost, "/api/public/events/"+eventA+"/live/qa?token="+tokenA, map[string]string{"text": "oi"}, nil)
	snap := getLiveAdminState(t, h, eventA, cookieA)
	if len(snap.QAInbox) != 1 {
		t.Fatalf("esperava 1 mensagem, got %+v", snap.QAInbox)
	}
	messageID := snap.QAInbox[0].ID

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventB+"/live/qa/"+messageID+"/dismiss", nil, []*http.Cookie{cookieB})
	if rec.Code != http.StatusNotFound {
		t.Errorf("mensagem de outro evento: status esperado 404, got %d", rec.Code)
	}
}

// readSSEFrame lê um bloco de evento SSE (linhas "event:"/"data:" seguidas de
// linha em branco) e devolve o nome do evento (vazio pro default) e o
// payload de data. Complementa o readEvent local de TestLiveStreamPushesUpdates,
// que só lida com o caso simples (sem "event:" nomeado).
func readSSEFrame(t *testing.T, reader *bufio.Reader) (eventName, data string) {
	t.Helper()
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("ler evento SSE: %v", err)
		}
		trimmed := strings.TrimRight(line, "\n")
		if trimmed == "" {
			if data != "" {
				return eventName, data
			}
			continue
		}
		if v, ok := strings.CutPrefix(trimmed, "event: "); ok {
			eventName = v
			continue
		}
		if v, ok := strings.CutPrefix(trimmed, "data: "); ok {
			data = v
		}
	}
}

func TestLiveAdminStreamPushesReactionsAndQAUpdates(t *testing.T) {
	h := newTestAPI(t).Handler()
	server := httptest.NewServer(h)
	defer server.Close()
	client := server.Client()

	cookie, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)
	viewerToken, _, code := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")
	if code != http.StatusOK {
		t.Fatalf("join: status %d", code)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/events/"+eventID+"/live/stream", nil)
	if err != nil {
		t.Fatalf("montar request: %v", err)
	}
	req.AddCookie(cookie)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("conectar no stream administrativo: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status esperado 200, got %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)

	name, data := readSSEFrame(t, reader)
	if name != "" {
		t.Fatalf("esperava frame default (snapshot) primeiro, got event=%q", name)
	}
	var initial liveAdminSnapshotDTO
	if err := json.Unmarshal([]byte(data), &initial); err != nil {
		t.Fatalf("decode snapshot inicial: %v", err)
	}
	if len(initial.QAInbox) != 0 {
		t.Fatalf("esperava caixa de q&a vazia inicialmente, got %+v", initial.QAInbox)
	}

	reactRec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/live/react?token="+viewerToken, map[string]string{"emoji": "🎉"}, nil)
	if reactRec.Code != http.StatusOK {
		t.Fatalf("reagir: status esperado 200, got %d: %s", reactRec.Code, reactRec.Body.String())
	}

	name, data = readSSEFrame(t, reader)
	if name != "reaction" {
		t.Fatalf("esperava evento nomeado 'reaction', got %q (data=%s)", name, data)
	}
	var reaction struct {
		Emoji string `json:"emoji"`
	}
	if err := json.Unmarshal([]byte(data), &reaction); err != nil {
		t.Fatalf("decode reação: %v", err)
	}
	if reaction.Emoji != "🎉" {
		t.Errorf("emoji esperado 🎉, got %q", reaction.Emoji)
	}

	qaRec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/live/qa?token="+viewerToken, map[string]string{"text": "Uma pergunta"}, nil)
	if qaRec.Code != http.StatusOK {
		t.Fatalf("submit qa: status esperado 200, got %d: %s", qaRec.Code, qaRec.Body.String())
	}

	name, data = readSSEFrame(t, reader)
	if name != "" {
		t.Fatalf("esperava frame default (snapshot atualizado), got event=%q", name)
	}
	var updated liveAdminSnapshotDTO
	if err := json.Unmarshal([]byte(data), &updated); err != nil {
		t.Fatalf("decode snapshot atualizado: %v", err)
	}
	if len(updated.QAInbox) != 1 || updated.QAInbox[0].Text != "Uma pergunta" {
		t.Fatalf("esperava a nova mensagem no snapshot atualizado, got %+v", updated.QAInbox)
	}
}

func TestLiveJoinWrongPIN(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, _, _, _, _, _, _ := setupLiveEvent(t, h)

	_, _, code := liveJoin(t, h, eventID, "ERRADO", "ana@exemplo.com")
	if code != http.StatusUnauthorized {
		t.Errorf("PIN errado: status esperado 401, got %d", code)
	}
}

// TestLiveStateNeverSerializesNullArrays é regressão de um bug real: slices
// Go zero-value (nil) serializam como JSON null, e o frontend (Svelte,
// PresentationStage) espera sempre array — um "participants": null quebrava
// a tela pública assim que alguém entrava antes de qualquer revelação.
func TestLiveStateNeverSerializesNullArrays(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)
	token, _, _ := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")

	rec := doJSON(t, h, http.MethodGet, "/api/public/events/"+eventID+"/live/state?token="+token, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	raw := rec.Body.String()
	for _, field := range []string{`"pending":null`, `"groups":null`, `"participants":null`} {
		if strings.Contains(raw, field) {
			t.Errorf("payload nao deveria conter %s (frontend espera array vazio, nao null): %s", field, raw)
		}
	}
}

func TestLiveJoinInvalidEmail(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, _, _, _, pin, _, _ := setupLiveEvent(t, h)

	_, _, code := liveJoin(t, h, eventID, pin, "nao-e-email")
	if code != http.StatusBadRequest {
		t.Errorf("e-mail invalido: status esperado 400, got %d", code)
	}
}

func TestLiveJoinEventNotFound(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, _, code := liveJoin(t, h, "999999", "QUALQUER", "ana@exemplo.com")
	if code != http.StatusNotFound {
		t.Errorf("evento inexistente: status esperado 404, got %d", code)
	}
}

func TestLiveStateRequiresValidToken(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, _, _, _, _, _, _ := setupLiveEvent(t, h)

	rec := doJSON(t, h, http.MethodGet, "/api/public/events/"+eventID+"/live/state?token=lixo", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("token invalido: status esperado 401, got %d", rec.Code)
	}
}

func TestLiveStateRejectsTokenFromAnotherEvent(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID1, _, _, _, pin1, _, _ := setupLiveEvent(t, h)
	_, eventID2, _, _, _, _, _, _ := setupLiveEvent(t, h)

	token, _, _ := liveJoin(t, h, eventID1, pin1, "ana@exemplo.com")

	rec := doJSON(t, h, http.MethodGet, "/api/public/events/"+eventID2+"/live/state?token="+token, nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("token de outro evento: status esperado 401, got %d", rec.Code)
	}
}

// TestLiveStateNeverLeaksUnrevealedAnswers é o teste de regressão de
// segurança central da feature: quem não foi revelado pelo admin nunca pode
// aparecer com a própria resposta no payload público, mesmo já tendo
// respondido no banco.
func TestLiveStateNeverLeaksUnrevealedAnswers(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, questionID, optAID, optBID, pin, anaID, biaID := setupLiveEvent(t, h)

	viewerToken, _, _ := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")

	// antes de qualquer revelação: os dois estão pendentes, nenhum grupo tem gente
	snap := getLiveState(t, h, eventID, viewerToken)
	if len(snap.Pending) != 2 {
		t.Fatalf("esperava 2 pendentes, got %d: %+v", len(snap.Pending), snap.Pending)
	}
	for _, g := range snap.Groups {
		if len(g.Participants) != 0 {
			t.Fatalf("nenhum grupo deveria ter participantes antes de revelar (isso vazaria a resposta): %+v", g)
		}
	}

	// admin revela só a ana
	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/reveal", map[string]string{
		"questionId": questionID, "participantId": anaID,
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("revelar: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	snap = getLiveState(t, h, eventID, viewerToken)
	if len(snap.Pending) != 1 || snap.Pending[0].ID != biaID {
		t.Fatalf("bia deveria continuar pendente, got %+v", snap.Pending)
	}
	for _, g := range snap.Groups {
		for _, p := range g.Participants {
			if p.ID == biaID {
				t.Fatalf("resposta da bia vazou no grupo %s antes dela ser revelada", g.Label)
			}
		}
	}
	// ana aparece no grupo certo (Go = optAID)
	found := false
	for _, g := range snap.Groups {
		for _, p := range g.Participants {
			if p.ID == anaID {
				found = true
				if g.Label != "Go" {
					t.Errorf("ana deveria estar no grupo Go, esta em %s", g.Label)
				}
			}
		}
	}
	if !found {
		t.Fatal("ana deveria aparecer revelada em algum grupo apos o reveal")
	}
	_ = optAID
	_ = optBID
}

func TestLiveRevealAllAndReset(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, questionID, _, _, pin, anaID, biaID := setupLiveEvent(t, h)
	viewerToken, _, _ := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/reveal-all", map[string]string{
		"questionId": questionID,
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("revelar todos: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	snap := getLiveState(t, h, eventID, viewerToken)
	if len(snap.Pending) != 0 {
		t.Fatalf("ninguem deveria estar pendente apos revelar todos, got %+v", snap.Pending)
	}
	total := 0
	for _, g := range snap.Groups {
		total += len(g.Participants)
	}
	if total != 2 {
		t.Fatalf("esperava 2 participantes revelados no total, got %d", total)
	}

	rec = doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/reset", map[string]string{
		"questionId": questionID,
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("reiniciar: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	snap = getLiveState(t, h, eventID, viewerToken)
	if len(snap.Pending) != 2 {
		t.Fatalf("todos deveriam voltar a pendente apos reset, got %+v", snap.Pending)
	}
	_ = anaID
	_ = biaID
}

func TestLiveRevealThenUnreveal(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, questionID, _, _, pin, anaID, _ := setupLiveEvent(t, h)
	viewerToken, _, _ := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/reveal", map[string]string{
		"questionId": questionID, "participantId": anaID,
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("revelar: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	snap := getLiveState(t, h, eventID, viewerToken)
	found := false
	for _, g := range snap.Groups {
		found = found || len(g.Participants) > 0
	}
	if len(snap.Pending) != 1 || !found {
		t.Fatalf("esperava ana revelada e 1 pendente, got pending=%+v groups=%+v", snap.Pending, snap.Groups)
	}

	rec = doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/unreveal", map[string]string{
		"questionId": questionID, "participantId": anaID,
	}, []*http.Cookie{cookie})
	if rec.Code != http.StatusOK {
		t.Fatalf("desrevelar: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	snap = getLiveState(t, h, eventID, viewerToken)
	if len(snap.Pending) != 2 {
		t.Fatalf("esperava ana de volta pra pendente, got %+v", snap.Pending)
	}
	for _, g := range snap.Groups {
		if len(g.Participants) != 0 {
			t.Fatalf("esperava nenhum participante revelado apos desrevelar, got %+v", snap.Groups)
		}
	}
}

func TestLiveAdminEndpointsRequireOwnership(t *testing.T) {
	h := newTestAPI(t).Handler()
	_, eventID, questionID, _, _, _, anaID, _ := setupLiveEvent(t, h)

	other := registerUser2(t, h)

	rec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/reveal", map[string]string{
		"questionId": questionID, "participantId": anaID,
	}, []*http.Cookie{other})
	if rec.Code != http.StatusNotFound {
		t.Errorf("dono errado: status esperado 404, got %d", rec.Code)
	}

	rec = doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/reveal", map[string]string{
		"questionId": questionID, "participantId": anaID,
	}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("sem sessao: status esperado 401, got %d", rec.Code)
	}

	rec = doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/unreveal", map[string]string{
		"questionId": questionID, "participantId": anaID,
	}, []*http.Cookie{other})
	if rec.Code != http.StatusNotFound {
		t.Errorf("desrevelar com dono errado: status esperado 404, got %d", rec.Code)
	}
}

func registerUser2(t *testing.T, h http.Handler) *http.Cookie {
	t.Helper()
	rec := doJSON(t, h, http.MethodPost, "/api/auth/register", map[string]string{
		"name": "Bia", "email": "bia-admin@exemplo.com", "password": "segredo",
	}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register: status %d: %s", rec.Code, rec.Body.String())
	}
	return sessionCookie(t, rec)
}

// TestLiveStreamPushesUpdates sobe um servidor HTTP de verdade (httptest não
// dá pra usar http.Flusher com o ResponseRecorder) e confere que revelar um
// participante em outra request chega no SSE já conectado. A leitura do
// stream usa um contexto com timeout (em vez de goroutine + select) pra
// poder chamar t.Fatalf com segurança direto na goroutine do teste.
func TestLiveStreamPushesUpdates(t *testing.T) {
	h := newTestAPI(t).Handler()
	server := httptest.NewServer(h)
	defer server.Close()
	client := server.Client()

	cookie, eventID, questionID, _, _, pin, anaID, _ := setupLiveEvent(t, h)
	token, _, code := liveJoin(t, h, eventID, pin, "curioso@exemplo.com")
	if code != http.StatusOK {
		t.Fatalf("join: status %d", code)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/public/events/"+eventID+"/live/stream?token="+token, nil)
	if err != nil {
		t.Fatalf("montar request: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("conectar no stream: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status esperado 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("content-type esperado text/event-stream, got %s", ct)
	}

	reader := bufio.NewReader(resp.Body)
	readEvent := func() liveSnapshotDTO {
		t.Helper()
		var dataLine string
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				t.Fatalf("ler evento SSE: %v", err)
			}
			trimmed := strings.TrimRight(line, "\n")
			if trimmed == "" {
				continue // linha em branco separando eventos SSE
			}
			dataLine = trimmed
			break
		}
		payload := strings.TrimPrefix(dataLine, "data: ")
		var snap liveSnapshotDTO
		if err := json.Unmarshal([]byte(payload), &snap); err != nil {
			t.Fatalf("decode evento SSE %q: %v", dataLine, err)
		}
		return snap
	}

	initial := readEvent()
	if len(initial.Pending) != 2 {
		t.Fatalf("snapshot inicial deveria ter 2 pendentes, got %d", len(initial.Pending))
	}

	revealRec := doJSON(t, h, http.MethodPost, "/api/events/"+eventID+"/live/reveal", map[string]string{
		"questionId": questionID, "participantId": anaID,
	}, []*http.Cookie{cookie})
	if revealRec.Code != http.StatusOK {
		t.Fatalf("revelar: status esperado 200, got %d: %s", revealRec.Code, revealRec.Body.String())
	}

	updated := readEvent()
	found := false
	for _, g := range updated.Groups {
		for _, p := range g.Participants {
			if p.ID == anaID {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("esperava ana revelada apos o push do SSE, got %+v", updated.Groups)
	}
}
