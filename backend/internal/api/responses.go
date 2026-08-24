package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"devopsconecta/backend/internal/store"
)

type answerDTO struct {
	QuestionID    string `json:"questionId"`
	QuestionTitle string `json:"questionTitle"`
	QuestionType  string `json:"questionType"`
	OptionID      string `json:"optionId"`
	OptionText    string `json:"optionText"`
	Text          string `json:"text"`
}

type participantResponseDTO struct {
	ID        string      `json:"id"`
	Email     string      `json:"email"`
	Name      string      `json:"name"`
	Photo     string      `json:"photo"`
	EditToken string      `json:"editToken"`
	CreatedAt string      `json:"createdAt"`
	Answers   []answerDTO `json:"answers"`
}

func (a *API) handleListResponses(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}

	questions, err := a.store.ListQuestionsByEvent(r.Context(), eventID)
	if err != nil {
		log.Printf("api: listar perguntas para respostas: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao buscar as respostas.")
		return
	}
	qByID := make(map[int64]store.QuestionWithOptions, len(questions))
	for _, q := range questions {
		qByID[q.ID] = q
	}

	participants, err := a.store.ListParticipantsByEvent(r.Context(), eventID)
	if err != nil {
		log.Printf("api: listar participantes: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao buscar as respostas.")
		return
	}

	dtos := make([]participantResponseDTO, 0, len(participants))
	for _, p := range participants {
		answers, err := a.store.ListAnswersByParticipant(r.Context(), p.ID)
		if err != nil {
			log.Printf("api: listar respostas do participante: %v", err)
			writeError(w, http.StatusInternalServerError, "Erro interno ao buscar as respostas.")
			return
		}
		ansByQuestion := make(map[int64]store.Answer, len(answers))
		for _, ans := range answers {
			ansByQuestion[ans.QuestionID] = ans
		}

		// Percorre as perguntas na ordem atual (order_index), não na ordem em
		// que as respostas foram salvas, para refletir reordenações feitas
		// pelo organizador depois do envio.
		adtos := make([]answerDTO, 0, len(questions))
		for _, q := range questions {
			ans, found := ansByQuestion[q.ID]
			if !found {
				continue
			}
			dto := answerDTO{
				QuestionID:    strconv.FormatInt(q.ID, 10),
				QuestionTitle: q.Title,
				QuestionType:  q.Type,
				Text:          ans.FreeText,
			}
			if ans.OptionID != 0 {
				dto.OptionID = strconv.FormatInt(ans.OptionID, 10)
				for _, o := range q.Options {
					if o.ID == ans.OptionID {
						dto.OptionText = o.TextLabel
						break
					}
				}
			}
			adtos = append(adtos, dto)
		}

		dtos = append(dtos, participantResponseDTO{
			ID:        strconv.FormatInt(p.ID, 10),
			Email:     p.Email,
			Name:      p.Name,
			Photo:     a.photoURL(p),
			EditToken: p.EditToken,
			CreatedAt: p.CreatedAt.Format(time.RFC3339),
			Answers:   adtos,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"participantCount": len(dtos),
		"participants":     dtos,
	})
}

type updateAnswerRequest struct {
	OptionID string `json:"optionId"`
	Text     string `json:"text"`
}

func (a *API) handleUpdateAnswer(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}

	participantID, err := strconv.ParseInt(r.PathValue("participantId"), 10, 64)
	if err != nil || participantID <= 0 {
		writeError(w, http.StatusBadRequest, "ID de participante inválido.")
		return
	}
	questionID, err := strconv.ParseInt(r.PathValue("questionId"), 10, 64)
	if err != nil || questionID <= 0 {
		writeError(w, http.StatusBadRequest, "ID de pergunta inválido.")
		return
	}

	participant, err := a.store.FindParticipantByID(r.Context(), participantID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Participante não encontrado.")
		return
	}
	if err != nil {
		log.Printf("api: buscar participante: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao atualizar a resposta.")
		return
	}
	if participant.EventID != eventID {
		writeError(w, http.StatusNotFound, "Participante não encontrado.")
		return
	}

	questions, err := a.store.ListQuestionsByEvent(r.Context(), eventID)
	if err != nil {
		log.Printf("api: listar perguntas: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao atualizar a resposta.")
		return
	}
	var question *store.QuestionWithOptions
	for i := range questions {
		if questions[i].ID == questionID {
			question = &questions[i]
			break
		}
	}
	if question == nil {
		writeError(w, http.StatusNotFound, "Pergunta não encontrada.")
		return
	}

	var req updateAnswerRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var optionID int64
	var text string
	if question.Type == questionTypeOpenText {
		text = strings.TrimSpace(req.Text)
		if len(text) > maxFreeTextLength {
			writeError(w, http.StatusBadRequest, "Resposta muito longa.")
			return
		}
	} else {
		if req.OptionID == "" {
			writeError(w, http.StatusBadRequest, "Selecione uma opção.")
			return
		}
		optionID, err = strconv.ParseInt(req.OptionID, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Opção inválida.")
			return
		}
		valid := false
		for _, o := range question.Options {
			if o.ID == optionID {
				valid = true
				break
			}
		}
		if !valid {
			writeError(w, http.StatusBadRequest, "Opção inválida.")
			return
		}
	}

	if err := a.store.UpdateAnswer(r.Context(), participantID, questionID, optionID, text); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Resposta não encontrada.")
		return
	} else if err != nil {
		log.Printf("api: atualizar resposta: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao atualizar a resposta.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type updateParticipantPhotoRequest struct {
	Photo string `json:"photo"`
}

// handleUpdateParticipantPhoto permite ao organizador definir ou corrigir a
// foto de um participante (ex: participante pediu atualização por fora do
// link de edição, ou não enviou foto ao responder).
func (a *API) handleUpdateParticipantPhoto(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}

	participantID, err := strconv.ParseInt(r.PathValue("participantId"), 10, 64)
	if err != nil || participantID <= 0 {
		writeError(w, http.StatusBadRequest, "ID de participante inválido.")
		return
	}

	participant, err := a.store.FindParticipantByID(r.Context(), participantID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Participante não encontrado.")
		return
	}
	if err != nil {
		log.Printf("api: buscar participante: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao atualizar a foto.")
		return
	}
	if participant.EventID != eventID {
		writeError(w, http.StatusNotFound, "Participante não encontrado.")
		return
	}

	var req updateParticipantPhotoRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Photo) > maxPhotoDataURLLength {
		writeError(w, http.StatusBadRequest, "Foto muito grande.")
		return
	}
	if req.Photo != "" && !strings.HasPrefix(req.Photo, "data:image/") {
		writeError(w, http.StatusBadRequest, "Formato de foto inválido.")
		return
	}

	if err := a.store.UpdateParticipantPhoto(r.Context(), participantID, req.Photo); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Participante não encontrado.")
		return
	} else if err != nil {
		log.Printf("api: atualizar foto do participante: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao atualizar a foto.")
		return
	}
	// Foto alterada ou removida: invalida o arquivo em disco pra próxima
	// requisição rematerializar com o novo valor (ou devolver 404).
	a.photos.Remove(participantID)

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
