package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// TestParticipantPhotoServedAsFile cobre o fluxo completo das fotos como
// arquivo: o payload da lista de respostas traz só a URL, o endpoint serve os
// bytes decodificados do base64 do banco, e a troca/remoção da foto invalida
// o arquivo em disco (rematerialização).
func TestParticipantPhotoServedAsFile(t *testing.T) {
	h := newTestAPI(t).Handler()
	cookie, eventID, groupQID, openQID, optAID, _ := setupPublicEvent(t, h)

	submit := func(email, name, photo string) string {
		t.Helper()
		rec := doJSON(t, h, http.MethodPost, "/api/public/events/"+eventID+"/submit", map[string]any{
			"email": email,
			"name":  name,
			"photo": photo,
			"answers": []map[string]string{
				{"questionId": groupQID, "optionId": optAID},
				{"questionId": openQID, "text": "Pizza"},
			},
		}, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("submit %s: status esperado 200, got %d: %s", email, rec.Code, rec.Body.String())
		}
		return participantIDFromResponses(t, h, cookie, eventID, email)
	}

	withPhoto := submit("ana@exemplo.com", "Ana", "data:image/jpeg;base64,Zm9vCg==")
	withoutPhoto := submit("bia@exemplo.com", "Bia", "")

	// 1. A listagem devolve URL, nunca o base64.
	rec := doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	var resp struct {
		Participants []participantResponseDTO `json:"participants"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode respostas: %v", err)
	}
	if got := resp.Participants[0].Photo; got != "/api/photos/"+withPhoto {
		t.Errorf("foto deveria ser a URL do arquivo, got %q (id %s)", got, withPhoto)
	}
	if got := resp.Participants[1].Photo; got != "" {
		t.Errorf("participante sem foto deveria vir com photo vazio, got %q", got)
	}

	// 2. Primeiro GET materializa a partir do base64 do banco.
	rec = doJSON(t, h, http.MethodGet, "/api/photos/"+withPhoto, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("primeiro GET da foto: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "foo\n" {
		t.Errorf("conteúdo esperado 'foo\\n', got %q", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "image/jpeg") {
		t.Errorf("content-type esperado image/jpeg, got %q", ct)
	}

	// 3. Segundo GET (arquivo já em disco) continua funcionando.
	rec = doJSON(t, h, http.MethodGet, "/api/photos/"+withPhoto, nil, nil)
	if rec.Code != http.StatusOK || rec.Body.String() != "foo\n" {
		t.Errorf("segundo GET: status %d, body %q", rec.Code, rec.Body.String())
	}

	// 4. Participante sem foto: 404.
	rec = doJSON(t, h, http.MethodGet, "/api/photos/"+withoutPhoto, nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("foto de participante sem foto: esperado 404, got %d", rec.Code)
	}

	// 5. Trocar a foto invalida o arquivo; o próximo GET serve o novo valor.
	rec = doJSON(t, h, http.MethodPatch,
		"/api/events/"+eventID+"/responses/"+withPhoto+"/photo",
		map[string]any{"photo": "data:image/png;base64,YmFyCg=="}, []*http.Cookie{cookie},
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("atualizar foto: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, h, http.MethodGet, "/api/photos/"+withPhoto, nil, nil)
	if rec.Code != http.StatusOK || rec.Body.String() != "bar\n" {
		t.Errorf("após troca: status %d, body %q (esperado 'bar\\n')", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "image/png") {
		t.Errorf("content-type esperado image/png, got %q", ct)
	}

	// 6. Remover a foto invalida o arquivo; o próximo GET devolve 404.
	rec = doJSON(t, h, http.MethodPatch,
		"/api/events/"+eventID+"/responses/"+withPhoto+"/photo",
		map[string]any{"photo": ""}, []*http.Cookie{cookie},
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("remover foto: status esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, h, http.MethodGet, "/api/photos/"+withPhoto, nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("após remoção: esperado 404, got %d", rec.Code)
	}
}

func TestParticipantPhotoInvalidID(t *testing.T) {
	h := newTestAPI(t).Handler()
	for _, id := range []string{"abc", "0", "-1"} {
		rec := doJSON(t, h, http.MethodGet, "/api/photos/"+id, nil, nil)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("id %q: esperado 400, got %d", id, rec.Code)
		}
	}
	rec := doJSON(t, h, http.MethodGet, "/api/photos/999999", nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("participante inexistente: esperado 404, got %d", rec.Code)
	}
}

// participantIDFromResponses busca o ID do participante pelo e-mail na lista
// de respostas (a rota pública não expõe o ID numérico).
func participantIDFromResponses(t *testing.T, h http.Handler, cookie *http.Cookie, eventID, email string) string {
	t.Helper()
	rec := doJSON(t, h, http.MethodGet, "/api/events/"+eventID+"/responses", nil, []*http.Cookie{cookie})
	var resp struct {
		Participants []participantResponseDTO `json:"participants"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode respostas: %v", err)
	}
	for _, p := range resp.Participants {
		if p.Email == email {
			return p.ID
		}
	}
	t.Fatalf("participante %s não encontrado", email)
	return ""
}
