package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"devopsconecta/backend/internal/store"
)

const (
	maxPhotoDataURLLength = 400_000
	maxFreeTextLength     = 2000
)

type publicEventDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	AnswersOpen bool   `json:"answersOpen"`
}

type publicOptionDTO struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type publicQuestionDTO struct {
	ID      string            `json:"id"`
	Title   string            `json:"title"`
	Type    string            `json:"type"`
	Options []publicOptionDTO `json:"options"`
}

type publicAnswerDTO struct {
	QuestionID string `json:"questionId"`
	OptionID   string `json:"optionId"`
	Text       string `json:"text"`
}

type publicParticipantDTO struct {
	Email   string            `json:"email"`
	Photo   string            `json:"photo"`
	Answers []publicAnswerDTO `json:"answers"`
}

// handlePublicGetParticipant resolve o participante pelo token de edição
// (link enviado a ele após o primeiro envio), para pré-preencher o
// formulário de resposta. O token é o único segredo que autoriza o acesso;
// token inválido ou de outro evento retorna 404 (não revela se o e-mail
// existe).
func (a *API) handlePublicGetParticipant(w http.ResponseWriter, r *http.Request) {
	id, ok := parseEventID(w, r)
	if !ok {
		return
	}
	token := strings.TrimSpace(r.URL.Query().Get("token"))

	participant, err := a.store.FindParticipantByEventAndToken(r.Context(), id, token)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Link de edição inválido ou expirado.")
		return
	}
	if err != nil {
		log.Printf("api: buscar participante por token: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao buscar suas respostas.")
		return
	}

	answers, err := a.store.ListAnswersByParticipant(r.Context(), participant.ID)
	if err != nil {
		log.Printf("api: listar respostas do participante: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao buscar suas respostas.")
		return
	}

	adtos := make([]publicAnswerDTO, 0, len(answers))
	for _, ans := range answers {
		dto := publicAnswerDTO{QuestionID: strconv.FormatInt(ans.QuestionID, 10), Text: ans.FreeText}
		if ans.OptionID != 0 {
			dto.OptionID = strconv.FormatInt(ans.OptionID, 10)
		}
		adtos = append(adtos, dto)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"participant": publicParticipantDTO{
			Email:   participant.Email,
			Photo:   participant.Photo,
			Answers: adtos,
		},
	})
}

// handlePublicResolvePIN acha o evento a partir só do PIN, pra tela inicial
// (sem ID na URL) poder checar se o código existe antes de avançar pro
// próximo passo. Não revela mais que o ID — status e detalhes do evento só
// aparecem nas rotas que já exigem PIN validado (join da live).
func (a *API) handlePublicResolvePIN(w http.ResponseWriter, r *http.Request) {
	pin := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("pin")))
	if pin == "" || !pinRe.MatchString(pin) {
		writeError(w, http.StatusNotFound, "Código não encontrado.")
		return
	}

	event, err := a.store.FindEventByPIN(r.Context(), pin)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Código não encontrado.")
		return
	}
	if err != nil {
		log.Printf("api: buscar evento por pin: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"id": strconv.FormatInt(event.ID, 10)})
}

func (a *API) handlePublicGetEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseEventID(w, r)
	if !ok {
		return
	}

	ev, err := a.store.FindEventByID(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Evento não encontrado.")
		return
	}
	if err != nil {
		log.Printf("api: buscar evento público: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao buscar o evento.")
		return
	}

	questions, err := a.store.ListQuestionsByEvent(r.Context(), id)
	if err != nil {
		log.Printf("api: listar perguntas públicas: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao buscar as perguntas.")
		return
	}

	dtos := make([]publicQuestionDTO, 0, len(questions))
	for _, q := range questions {
		opts := make([]publicOptionDTO, 0, len(q.Options))
		for _, o := range q.Options {
			opts = append(opts, publicOptionDTO{ID: strconv.FormatInt(o.ID, 10), Text: o.TextLabel})
		}
		dtos = append(dtos, publicQuestionDTO{
			ID:      strconv.FormatInt(q.ID, 10),
			Title:   q.Title,
			Type:    q.Type,
			Options: opts,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"event": publicEventDTO{
			ID:          strconv.FormatInt(ev.ID, 10),
			Title:       ev.Title,
			AnswersOpen: ev.Status == eventStatusOpenForAnswers,
		},
		"questions": dtos,
	})
}

type submitAnswerRequest struct {
	QuestionID string `json:"questionId"`
	OptionID   string `json:"optionId"`
	Text       string `json:"text"`
}

type submitRequest struct {
	Email     string                `json:"email"`
	Photo     string                `json:"photo"`
	EditToken string                `json:"editToken"`
	Answers   []submitAnswerRequest `json:"answers"`
}

func (a *API) handleSubmitAnswers(w http.ResponseWriter, r *http.Request) {
	id, ok := parseEventID(w, r)
	if !ok {
		return
	}

	ev, err := a.store.FindEventByID(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Evento não encontrado.")
		return
	} else if err != nil {
		log.Printf("api: buscar evento para envio: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao enviar as respostas.")
		return
	}
	if ev.Status != eventStatusOpenForAnswers {
		writeError(w, http.StatusForbidden, "As respostas não estão abertas para este evento.")
		return
	}

	var req submitRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.EditToken = strings.TrimSpace(req.EditToken)
	if req.EditToken != "" {
		p, err := a.store.FindParticipantByEventAndToken(r.Context(), id, req.EditToken)
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Link de edição inválido ou expirado.")
			return
		}
		if err != nil {
			log.Printf("api: buscar participante por token: %v", err)
			writeError(w, http.StatusInternalServerError, "Erro interno ao enviar as respostas.")
			return
		}
		req.Email = p.Email
	} else {
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		if !isValidEmail(req.Email) {
			writeError(w, http.StatusBadRequest, "Informe um e-mail válido.")
			return
		}
	}
	if len(req.Photo) > maxPhotoDataURLLength {
		writeError(w, http.StatusBadRequest, "Foto muito grande.")
		return
	}
	if req.Photo != "" && !strings.HasPrefix(req.Photo, "data:image/") {
		writeError(w, http.StatusBadRequest, "Formato de foto inválido.")
		return
	}

	questions, err := a.store.ListQuestionsByEvent(r.Context(), id)
	if err != nil {
		log.Printf("api: listar perguntas para envio: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao enviar as respostas.")
		return
	}
	if len(questions) == 0 {
		writeError(w, http.StatusBadRequest, "Este evento não tem perguntas.")
		return
	}

	byID := make(map[int64]store.QuestionWithOptions, len(questions))
	for _, q := range questions {
		byID[q.ID] = q
	}

	if len(req.Answers) != len(questions) {
		writeError(w, http.StatusBadRequest, "Responda todas as perguntas antes de enviar.")
		return
	}

	seen := make(map[int64]bool, len(req.Answers))
	answers := make([]store.Answer, 0, len(req.Answers))
	for _, ans := range req.Answers {
		qID, err := strconv.ParseInt(ans.QuestionID, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Pergunta inválida.")
			return
		}
		q, ok := byID[qID]
		if !ok {
			writeError(w, http.StatusBadRequest, "Pergunta não pertence a este evento.")
			return
		}
		if seen[qID] {
			writeError(w, http.StatusBadRequest, "Resposta duplicada para a mesma pergunta.")
			return
		}
		seen[qID] = true

		answer := store.Answer{ID: a.ids.NextID(), QuestionID: qID}
		if q.Type == questionTypeOpenText {
			text := strings.TrimSpace(ans.Text)
			if text == "" {
				writeError(w, http.StatusBadRequest, `Responda a pergunta "`+q.Title+`".`)
				return
			}
			if len(text) > maxFreeTextLength {
				writeError(w, http.StatusBadRequest, "Resposta muito longa.")
				return
			}
			answer.FreeText = text
		} else {
			optID, err := strconv.ParseInt(ans.OptionID, 10, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, `Selecione uma opção para "`+q.Title+`".`)
				return
			}
			valid := false
			for _, o := range q.Options {
				if o.ID == optID {
					valid = true
					break
				}
			}
			if !valid {
				writeError(w, http.StatusBadRequest, "Opção inválida.")
				return
			}
			answer.OptionID = optID
		}
		answers = append(answers, answer)
	}

	participant, err := a.store.UpsertParticipant(r.Context(), store.Participant{
		ID:      a.ids.NextID(),
		EventID: id,
		Email:   req.Email,
		Photo:   req.Photo,
	})
	if err != nil {
		log.Printf("api: salvar participante: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao enviar as respostas.")
		return
	}

	if err := a.store.ReplaceAnswers(r.Context(), participant.ID, answers); err != nil {
		log.Printf("api: salvar respostas: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao enviar as respostas.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "editToken": participant.EditToken})
}
