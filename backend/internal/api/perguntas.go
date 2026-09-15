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
	questionTypeGroup    = "GROUP"
	questionTypeOpenText = "OPEN_TEXT"

	maxQuestionTitleLength = 300
	maxOptionTextLength    = 120
	minGroupOptions        = 2
	maxOptions             = 10

	// otherOptionLabel é o texto fixo da opção sintética adicionada quando
	// allowOther é true — não é digitada pelo organizador (ver
	// createQuestionRequest.AllowOther).
	otherOptionLabel = "Outro"
)

type questionOptionDTO struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	IsOther bool   `json:"isOther,omitempty"`
}

type questionDTO struct {
	ID         string              `json:"id"`
	EventID    string              `json:"eventId"`
	Title      string              `json:"title"`
	Type       string              `json:"type"`
	OrderIndex int64               `json:"orderIndex"`
	Options    []questionOptionDTO `json:"options"`
}

func toQuestionDTO(q store.QuestionWithOptions) questionDTO {
	opts := make([]questionOptionDTO, 0, len(q.Options))
	for _, o := range q.Options {
		opts = append(opts, questionOptionDTO{
			ID:      strconv.FormatInt(o.ID, 10),
			Text:    o.TextLabel,
			IsOther: o.IsOther,
		})
	}
	return questionDTO{
		ID:         strconv.FormatInt(q.ID, 10),
		EventID:    strconv.FormatInt(q.EventID, 10),
		Title:      q.Title,
		Type:       q.Type,
		OrderIndex: q.OrderIndex,
		Options:    opts,
	}
}

// otherOption cria a opção sintética "Outro", incluída além das opções do
// organizador quando allowOther é true. Só se aplica a perguntas com opções
// (GROUP) — OPEN_TEXT já é resposta livre por definição.
func (a *API) otherOption() store.QuestionOption {
	return store.QuestionOption{ID: a.ids.NextID(), TextLabel: otherOptionLabel, IsOther: true}
}

// handleListarPerguntas godoc
//
// @Summary     Lista as perguntas de um evento
// @Tags        perguntas
// @Produce     json
// @Param       id path string true "ID do evento"
// @Success     200 {object} questionsEnvelope
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/perguntas [get]
func (a *API) handleListarPerguntas(w http.ResponseWriter, r *http.Request) {
	id, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	questions, err := a.store.ListarPerguntasPorEvento(r.Context(), id)
	if err != nil {
		log.Printf("api: listar perguntas: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao listar perguntas.")
		return
	}
	dtos := make([]questionDTO, 0, len(questions))
	for _, q := range questions {
		dtos = append(dtos, toQuestionDTO(q))
	}
	writeJSON(w, http.StatusOK, map[string]any{"questions": dtos})
}

type createQuestionRequest struct {
	Title      string   `json:"title"`
	Type       string   `json:"type"`
	Options    []string `json:"options"`
	AllowOther bool     `json:"allowOther"`
}

// handleCriarPergunta godoc
//
// @Summary     Cria uma pergunta no evento
// @Description type: GROUP (opções, min. 2) ou OPEN_TEXT (sem opções). allowOther adiciona uma opção "Outro" com resposta livre opcional (só GROUP).
// @Tags        perguntas
// @Accept      json
// @Produce     json
// @Param       id   path string                true "ID do evento"
// @Param       body body createQuestionRequest true "Pergunta"
// @Success     201 {object} questionEnvelope
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/perguntas [post]
func (a *API) handleCriarPergunta(w http.ResponseWriter, r *http.Request) {
	id, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}

	var req createQuestionRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Type = strings.TrimSpace(req.Type)

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "Informe o texto da pergunta.")
		return
	}
	if len(req.Title) > maxQuestionTitleLength {
		writeError(w, http.StatusBadRequest, "Pergunta muito longa (máximo "+strconv.Itoa(maxQuestionTitleLength)+" caracteres).")
		return
	}
	if req.Type != questionTypeGroup && req.Type != questionTypeOpenText {
		writeError(w, http.StatusBadRequest, "Tipo de pergunta inválido. Use GROUP ou OPEN_TEXT.")
		return
	}

	options, msg := normalizeOptions(req.Options, req.Type)
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	optionEnts := make([]store.QuestionOption, 0, len(options)+1)
	for _, text := range options {
		optionEnts = append(optionEnts, store.QuestionOption{ID: a.ids.NextID(), TextLabel: text})
	}
	if req.AllowOther && req.Type == questionTypeGroup {
		optionEnts = append(optionEnts, a.otherOption())
	}

	question, err := a.store.CriarPergunta(r.Context(), store.Question{
		ID:      a.ids.NextID(),
		EventID: id,
		Title:   req.Title,
		Type:    req.Type,
	}, optionEnts)
	if err != nil {
		log.Printf("api: criar pergunta: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao criar a pergunta.")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"question": toQuestionDTO(question)})
}

// handleRemoverPergunta godoc
//
// @Summary  Remove uma pergunta
// @Tags     perguntas
// @Param    id         path string true "ID do evento"
// @Param    questionId path string true "ID da pergunta"
// @Success  204 "removida"
// @Failure  400 {object} errorResponse
// @Failure  404 {object} errorResponse
// @Security cookieAuth
// @Router   /api/eventos/{id}/perguntas/{questionId} [delete]
func (a *API) handleRemoverPergunta(w http.ResponseWriter, r *http.Request) {
	id, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}

	questionID, ok := parseQuestionID(w, r)
	if !ok {
		return
	}

	if err := a.store.RemoverPergunta(r.Context(), questionID, id); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Pergunta não encontrada.")
		return
	} else if err != nil {
		log.Printf("api: remover pergunta: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao remover a pergunta.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type updateQuestionRequest struct {
	Title      string   `json:"title"`
	Options    []string `json:"options"`
	AllowOther bool     `json:"allowOther"`
}

// handleAtualizarPergunta godoc
//
// @Summary     Atualiza o título/opções de uma pergunta
// @Description O tipo (GROUP/OPEN_TEXT) não muda depois de criada — só título, opções e a opção "Outro" (allowOther).
// @Tags        perguntas
// @Accept      json
// @Produce     json
// @Param       id         path string                true "ID do evento"
// @Param       questionId path string                true "ID da pergunta"
// @Param       body       body updateQuestionRequest true "Novo título/opções"
// @Success     200 {object} questionEnvelope
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/perguntas/{questionId} [patch]
func (a *API) handleAtualizarPergunta(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}

	questionID, ok := parseQuestionID(w, r)
	if !ok {
		return
	}

	var req updateQuestionRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "Informe o texto da pergunta.")
		return
	}
	if len(req.Title) > maxQuestionTitleLength {
		writeError(w, http.StatusBadRequest, "Pergunta muito longa (máximo "+strconv.Itoa(maxQuestionTitleLength)+" caracteres).")
		return
	}

	existing, err := a.store.ListarPerguntasPorEvento(r.Context(), eventID)
	if err != nil {
		log.Printf("api: buscar pergunta para editar: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao atualizar a pergunta.")
		return
	}
	var current *store.QuestionWithOptions
	for i := range existing {
		if existing[i].ID == questionID {
			current = &existing[i]
			break
		}
	}
	if current == nil {
		writeError(w, http.StatusNotFound, "Pergunta não encontrada.")
		return
	}

	options, msg := normalizeOptions(req.Options, current.Type)
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	optionEnts := make([]store.QuestionOption, 0, len(options)+1)
	for _, text := range options {
		optionEnts = append(optionEnts, store.QuestionOption{ID: a.ids.NextID(), TextLabel: text})
	}
	if req.AllowOther && current.Type == questionTypeGroup {
		optionEnts = append(optionEnts, a.otherOption())
	}

	question, err := a.store.AtualizarPergunta(r.Context(), store.Question{
		ID:         questionID,
		EventID:    eventID,
		Title:      req.Title,
		Type:       current.Type,
		OrderIndex: current.OrderIndex,
	}, optionEnts)
	if err != nil {
		log.Printf("api: atualizar pergunta: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao atualizar a pergunta.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"question": toQuestionDTO(question)})
}

func parseQuestionID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("questionId"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID de pergunta inválido.")
		return 0, false
	}
	return id, true
}

type reorderQuestionsRequest struct {
	QuestionIDs []string `json:"questionIds"`
}

// handleReordenarPerguntas godoc
//
// @Summary     Reordena as perguntas do evento
// @Description questionIds deve conter TODAS as perguntas do evento, na nova ordem desejada.
// @Tags        perguntas
// @Accept      json
// @Produce     json
// @Param       id   path string                  true "ID do evento"
// @Param       body body reorderQuestionsRequest true "IDs das perguntas na nova ordem"
// @Success     204 "reordenado"
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/perguntas/ordem [put]
func (a *API) handleReordenarPerguntas(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}

	var req reorderQuestionsRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	existing, err := a.store.ListarPerguntasPorEvento(r.Context(), eventID)
	if err != nil {
		log.Printf("api: listar perguntas para reordenar: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao reordenar perguntas.")
		return
	}
	if len(req.QuestionIDs) != len(existing) {
		writeError(w, http.StatusBadRequest, "A ordem deve conter todas as perguntas do evento.")
		return
	}

	ids := make([]int64, 0, len(req.QuestionIDs))
	seen := make(map[string]bool, len(req.QuestionIDs))
	for _, raw := range req.QuestionIDs {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			writeError(w, http.StatusBadRequest, "ID de pergunta inválido.")
			return
		}
		if seen[raw] {
			writeError(w, http.StatusBadRequest, "A ordem contém perguntas repetidas.")
			return
		}
		seen[raw] = true
		ids = append(ids, id)
	}
	for _, q := range existing {
		if !seen[strconv.FormatInt(q.ID, 10)] {
			writeError(w, http.StatusBadRequest, "A ordem deve conter apenas perguntas deste evento.")
			return
		}
	}

	if err := a.store.ReordenarPerguntas(r.Context(), eventID, ids); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Pergunta não encontrada.")
		return
	} else if err != nil {
		log.Printf("api: reordenar perguntas: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao reordenar perguntas.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// resolveEventOwner valida o ID do evento e o acesso de quem está autenticado
// (dono do evento, ou super admin — ver autorizarAcessoEvento em eventos.go).
func (a *API) resolveEventOwner(w http.ResponseWriter, r *http.Request) (int64, bool) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return 0, false
	}
	if _, err := a.autorizarAcessoEvento(r.Context(), eventID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Evento não encontrado.")
			return 0, false
		}
		log.Printf("api: buscar evento: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao buscar o evento.")
		return 0, false
	}
	return eventID, true
}

// normalizeOptions valida e limpa as opções de uma pergunta GROUP (mínimo
// minGroupOptions, máximo maxOptions). Perguntas de resposta aberta
// (OPEN_TEXT) não têm opções.
func normalizeOptions(options []string, qType string) ([]string, string) {
	if qType == questionTypeOpenText {
		return []string{}, ""
	}
	if len(options) > maxOptions {
		return nil, "Informe até " + strconv.Itoa(maxOptions) + " opções."
	}
	seen := make(map[string]bool, len(options))
	clean := make([]string, 0, len(options))
	for _, o := range options {
		o = strings.TrimSpace(o)
		if o == "" {
			return nil, "Opções não podem ser vazias."
		}
		if len(o) > maxOptionTextLength {
			return nil, "Opção muito longa (máximo " + strconv.Itoa(maxOptionTextLength) + " caracteres)."
		}
		if seen[o] {
			return nil, "As opções devem ser diferentes entre si."
		}
		seen[o] = true
		clean = append(clean, o)
	}
	if len(clean) < minGroupOptions {
		return nil, "Perguntas em grupo precisam de pelo menos " + strconv.Itoa(minGroupOptions) + " opções."
	}
	return clean, ""
}
