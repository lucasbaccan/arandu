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
	questionTypeGroup      = "GROUP"
	questionTypeIndividual = "INDIVIDUAL"
	defaultLayoutView      = "TIMELINE"

	maxQuestionTitleLength = 300
	maxOptionTextLength    = 120
	minOptions             = 1
	minGroupOptions        = 2
	maxOptions             = 10
)

var allowedLayoutViews = map[string]bool{
	"TIMELINE": true,
	"DUAL":     true,
	"CLOUD":    true,
	"CENTER":   true,
}

type questionOptionDTO struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type questionDTO struct {
	ID         string              `json:"id"`
	EventID    string              `json:"eventId"`
	Title      string              `json:"title"`
	Type       string              `json:"type"`
	LayoutView string              `json:"layoutView"`
	OrderIndex int64               `json:"orderIndex"`
	Options    []questionOptionDTO `json:"options"`
}

func toQuestionDTO(q store.QuestionWithOptions) questionDTO {
	opts := make([]questionOptionDTO, 0, len(q.Options))
	for _, o := range q.Options {
		opts = append(opts, questionOptionDTO{
			ID:   strconv.FormatInt(o.ID, 10),
			Text: o.TextLabel,
		})
	}
	return questionDTO{
		ID:         strconv.FormatInt(q.ID, 10),
		EventID:    strconv.FormatInt(q.EventID, 10),
		Title:      q.Title,
		Type:       q.Type,
		LayoutView: q.LayoutView,
		OrderIndex: q.OrderIndex,
		Options:    opts,
	}
}

func (a *API) handleListQuestions(w http.ResponseWriter, r *http.Request) {
	id, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	questions, err := a.store.ListQuestionsByEvent(r.Context(), id)
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
	LayoutView string   `json:"layoutView"`
	Options    []string `json:"options"`
}

func (a *API) handleCreateQuestion(w http.ResponseWriter, r *http.Request) {
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
	req.LayoutView = strings.TrimSpace(req.LayoutView)

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "Informe o texto da pergunta.")
		return
	}
	if len(req.Title) > maxQuestionTitleLength {
		writeError(w, http.StatusBadRequest, "Pergunta muito longa (máximo "+strconv.Itoa(maxQuestionTitleLength)+" caracteres).")
		return
	}
	if req.Type != questionTypeGroup && req.Type != questionTypeIndividual {
		writeError(w, http.StatusBadRequest, "Tipo de pergunta inválido. Use GROUP ou INDIVIDUAL.")
		return
	}
	if req.LayoutView == "" {
		req.LayoutView = defaultLayoutView
	}
	if !allowedLayoutViews[req.LayoutView] {
		writeError(w, http.StatusBadRequest, "Layout de pergunta inválido.")
		return
	}

	options, msg := normalizeOptions(req.Options, req.Type)
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	optionEnts := make([]store.QuestionOption, 0, len(options))
	for _, text := range options {
		optionEnts = append(optionEnts, store.QuestionOption{ID: a.ids.NextID(), TextLabel: text})
	}

	question, err := a.store.CreateQuestion(r.Context(), store.Question{
		ID:         a.ids.NextID(),
		EventID:    id,
		Title:      req.Title,
		Type:       req.Type,
		LayoutView: req.LayoutView,
	}, optionEnts)
	if err != nil {
		log.Printf("api: criar pergunta: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao criar a pergunta.")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"question": toQuestionDTO(question)})
}

func (a *API) handleDeleteQuestion(w http.ResponseWriter, r *http.Request) {
	id, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}

	questionID, ok := parseQuestionID(w, r)
	if !ok {
		return
	}

	if err := a.store.DeleteQuestion(r.Context(), questionID, id); errors.Is(err, store.ErrNotFound) {
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
	Title   string   `json:"title"`
	Options []string `json:"options"`
}

func (a *API) handleUpdateQuestion(w http.ResponseWriter, r *http.Request) {
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

	options, msg := normalizeOptions(req.Options, questionTypeGroup)
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	existing, err := a.store.ListQuestionsByEvent(r.Context(), eventID)
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

	optionEnts := make([]store.QuestionOption, 0, len(options))
	for _, text := range options {
		optionEnts = append(optionEnts, store.QuestionOption{ID: a.ids.NextID(), TextLabel: text})
	}

	question, err := a.store.UpdateQuestion(r.Context(), store.Question{
		ID:         questionID,
		EventID:    eventID,
		Title:      req.Title,
		Type:       current.Type,
		LayoutView: current.LayoutView,
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

func (a *API) handleReorderQuestions(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}

	var req reorderQuestionsRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	existing, err := a.store.ListQuestionsByEvent(r.Context(), eventID)
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

	if err := a.store.ReorderQuestions(r.Context(), eventID, ids); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Pergunta não encontrada.")
		return
	} else if err != nil {
		log.Printf("api: reordenar perguntas: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao reordenar perguntas.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// resolveEventOwner valida o ID do evento e a posse pelo usuário autenticado.
func (a *API) resolveEventOwner(w http.ResponseWriter, r *http.Request) (int64, bool) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return 0, false
	}
	ownerID := userIDFromContext(r.Context())
	if _, err := a.store.FindEventByIDAndOwner(r.Context(), eventID, ownerID); err != nil {
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

// normalizeOptions valida e limpa as opções, exigindo mínimo por tipo de pergunta.
func normalizeOptions(options []string, qType string) ([]string, string) {
	if len(options) < minOptions || len(options) > maxOptions {
		return nil, "Informe entre 1 e " + strconv.Itoa(maxOptions) + " opções."
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
	if qType == questionTypeGroup && len(clean) < minGroupOptions {
		return nil, "Perguntas em grupo precisam de pelo menos " + strconv.Itoa(minGroupOptions) + " opções."
	}
	return clean, ""
}
