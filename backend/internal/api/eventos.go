package api

import (
	"context"
	"crypto/rand"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"devopsconecta/backend/internal/store"
)

const (
	eventStatusPreparation      = "PREPARATION"
	eventStatusOpenForAnswers   = "OPEN_FOR_ANSWERS"
	eventStatusClosedForAnswers = "CLOSED_FOR_ANSWERS"
	maxEventTitleLength         = 120
	autoPINRetries              = 5
)

var pinRe = regexp.MustCompile(`^[a-zA-Z0-9_\-]{1,25}$`)

type eventDTO struct {
	ID                  string `json:"id"`
	OwnerID             string `json:"ownerId"`
	Title               string `json:"title"`
	PINCode             string `json:"pinCode"`
	Status              string `json:"status"`
	ShowRanking         bool   `json:"configShowRanking"`
	AllowEdit           bool   `json:"allowEdit"`
	InteractionsEnabled bool   `json:"interactionsEnabled"`
	CreatedAt           string `json:"createdAt"`
	QuestionCount       int    `json:"questionCount"`
	ParticipantCount    int    `json:"participantCount"`
}

func toEventDTO(e store.Event) eventDTO {
	return eventDTO{
		ID:                  strconv.FormatInt(e.ID, 10),
		OwnerID:             strconv.FormatInt(e.OwnerID, 10),
		Title:               e.Title,
		PINCode:             e.PINCode,
		Status:              e.Status,
		ShowRanking:         e.ShowRanking,
		AllowEdit:           e.AllowEdit,
		InteractionsEnabled: e.InteractionsEnabled,
		CreatedAt:           e.CreatedAt.Format(time.RFC3339),
	}
}

func toEventSummaryDTO(es store.EventSummary) eventDTO {
	dto := toEventDTO(es.Event)
	dto.QuestionCount = es.QuestionCount
	dto.ParticipantCount = es.ParticipantCount
	return dto
}

type createEventRequest struct {
	Title   string `json:"title"`
	PINCode string `json:"pinCode"`
}

// handleCriarEvento godoc
//
// @Summary     Cria um evento
// @Description Sem pinCode, gera um PIN numérico de 6 dígitos automaticamente.
// @Tags        eventos
// @Accept      json
// @Produce     json
// @Param       body body createEventRequest true "Título e (opcional) PIN"
// @Success     201 {object} eventEnvelope
// @Failure     400 {object} errorResponse
// @Failure     409 {object} errorResponse "PIN em uso"
// @Security    cookieAuth
// @Router      /api/eventos [post]
func (a *API) handleCriarEvento(w http.ResponseWriter, r *http.Request) {
	var req createEventRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.PINCode = strings.ToUpper(strings.TrimSpace(req.PINCode))

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "Informe o título do evento.")
		return
	}
	if len(req.Title) > maxEventTitleLength {
		writeError(w, http.StatusBadRequest, "Título muito longo (máximo "+strconv.Itoa(maxEventTitleLength)+" caracteres).")
		return
	}

	customPIN := req.PINCode != ""
	if customPIN && !pinRe.MatchString(req.PINCode) {
		writeError(w, http.StatusBadRequest, "O PIN deve ter 1 a 25 caracteres: letras, números, _ ou -.")
		return
	}

	ownerID := userIDFromContext(r.Context())

	attempts := 1
	if !customPIN {
		attempts = autoPINRetries
	}
	for i := 0; i < attempts; i++ {
		pin := req.PINCode
		if !customPIN {
			pin = generatePIN()
		}
		ev, err := a.store.CriarEvento(r.Context(), store.Event{
			ID:                  a.ids.NextID(),
			OwnerID:             ownerID,
			Title:               req.Title,
			PINCode:             pin,
			Status:              eventStatusPreparation,
			AllowEdit:           true,
			InteractionsEnabled: true,
		})
		if errors.Is(err, store.ErrPinTaken) && !customPIN {
			continue
		}
		if errors.Is(err, store.ErrPinTaken) {
			writeError(w, http.StatusConflict, "Este PIN já está em uso. Escolha outro.")
			return
		}
		if err != nil {
			log.Printf("api: criar evento: %v", err)
			writeError(w, http.StatusInternalServerError, "Erro interno ao criar o evento.")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"event": toEventDTO(ev)})
		return
	}
	writeError(w, http.StatusConflict, "Não foi possível gerar um PIN livre. Tente novamente.")
}

// handleListarEventos godoc
//
// @Summary     Lista os eventos do organizador autenticado
// @Tags        eventos
// @Produce     json
// @Success     200 {object} eventsEnvelope
// @Security    cookieAuth
// @Router      /api/eventos [get]
func (a *API) handleListarEventos(w http.ResponseWriter, r *http.Request) {
	ownerID := userIDFromContext(r.Context())
	events, err := a.store.ListarResumosDeEventosPorDono(r.Context(), ownerID)
	if err != nil {
		log.Printf("api: listar eventos: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao listar eventos.")
		return
	}
	dtos := make([]eventDTO, 0, len(events))
	for _, e := range events {
		dtos = append(dtos, toEventSummaryDTO(e))
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": dtos})
}

// autorizarAcessoEvento verifica se quem está autenticado pode agir sobre o
// evento: o dono, ou o super admin (que controla todos os eventos, de
// qualquer organizador). Acesso negado responde como "não encontrado" — não
// vaza pra quem não é dono nem admin que o evento existe e é de outra
// pessoa. Devolve o evento com o OwnerID real, que os handlers de escrita
// (atualizar/excluir) precisam usar em vez do id de quem está agindo, quando
// os dois forem diferentes (edição feita pelo super admin).
func (a *API) autorizarAcessoEvento(ctx context.Context, eventID int64) (store.Event, error) {
	ev, err := a.store.BuscarEventoPorID(ctx, eventID)
	if err != nil {
		return store.Event{}, err
	}
	userID := userIDFromContext(ctx)
	if ev.OwnerID == userID {
		return ev, nil
	}
	u, err := a.store.BuscarUsuarioPorID(ctx, userID)
	if err != nil || u.Role != store.RoleSuperAdmin {
		return store.Event{}, store.ErrNotFound
	}
	return ev, nil
}

// handleBuscarEvento godoc
//
// @Summary     Busca um evento por ID
// @Description Dono do evento ou super admin — qualquer outro recebe 404 (não revela que o evento existe).
// @Tags        eventos
// @Produce     json
// @Param       id path string true "ID do evento"
// @Success     200 {object} eventEnvelope
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id} [get]
func (a *API) handleBuscarEvento(w http.ResponseWriter, r *http.Request) {
	id, ok := parseEventID(w, r)
	if !ok {
		return
	}
	ev, err := a.autorizarAcessoEvento(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Evento não encontrado.")
		return
	}
	if err != nil {
		log.Printf("api: buscar evento: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao buscar o evento.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"event": toEventDTO(ev)})
}

type updateEventRequest struct {
	Title       string `json:"title"`
	PINCode     string `json:"pinCode"`
	Status      string `json:"status"`
	ShowRanking bool   `json:"configShowRanking"`
	AllowEdit   bool   `json:"allowEdit"`
}

// handleAtualizarEvento godoc
//
// @Summary     Atualiza um evento
// @Description Dono do evento ou super admin. status aceita "" (mantém), OPEN_FOR_ANSWERS ou CLOSED_FOR_ANSWERS.
// @Tags        eventos
// @Accept      json
// @Produce     json
// @Param       id   path string             true "ID do evento"
// @Param       body body updateEventRequest true "Campos a atualizar"
// @Success     200 {object} eventEnvelope
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Failure     409 {object} errorResponse "PIN em uso"
// @Security    cookieAuth
// @Router      /api/eventos/{id} [patch]
func (a *API) handleAtualizarEvento(w http.ResponseWriter, r *http.Request) {
	id, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req updateEventRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.PINCode = strings.ToUpper(strings.TrimSpace(req.PINCode))

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "Informe o título do evento.")
		return
	}
	if len(req.Title) > maxEventTitleLength {
		writeError(w, http.StatusBadRequest, "Título muito longo (máximo "+strconv.Itoa(maxEventTitleLength)+" caracteres).")
		return
	}
	if req.PINCode != "" && !pinRe.MatchString(req.PINCode) {
		writeError(w, http.StatusBadRequest, "O PIN deve ter 1 a 25 caracteres: letras, números, _ ou -.")
		return
	}
	if req.Status != "" && req.Status != eventStatusOpenForAnswers && req.Status != eventStatusClosedForAnswers {
		writeError(w, http.StatusBadRequest, "Status inválido.")
		return
	}

	current, err := a.autorizarAcessoEvento(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Evento não encontrado.")
		return
	}
	if err != nil {
		log.Printf("api: buscar evento: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao atualizar o evento.")
		return
	}

	pin := current.PINCode
	if req.PINCode != "" {
		pin = req.PINCode
	}
	status := current.Status
	if req.Status != "" {
		status = req.Status
	}

	ev, err := a.store.AtualizarEvento(r.Context(), store.Event{
		ID:          id,
		OwnerID:     current.OwnerID,
		Title:       req.Title,
		PINCode:     pin,
		Status:      status,
		ShowRanking: req.ShowRanking,
		AllowEdit:   req.AllowEdit,
	})
	if errors.Is(err, store.ErrPinTaken) {
		writeError(w, http.StatusConflict, "Este PIN já está em uso. Escolha outro.")
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Evento não encontrado.")
		return
	}
	if err != nil {
		log.Printf("api: atualizar evento: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao atualizar o evento.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"event": toEventDTO(ev)})
}

// handleDeletarEvento godoc
//
// @Summary     Exclui um evento
// @Description Dono do evento ou super admin. Apaga em cascata perguntas, participantes e respostas.
// @Tags        eventos
// @Produce     json
// @Param       id path string true "ID do evento"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id} [delete]
func (a *API) handleDeletarEvento(w http.ResponseWriter, r *http.Request) {
	id, ok := parseEventID(w, r)
	if !ok {
		return
	}
	ev, err := a.autorizarAcessoEvento(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Evento não encontrado.")
		return
	}
	if err != nil {
		log.Printf("api: buscar evento: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao excluir o evento.")
		return
	}
	if err := a.store.DeletarEvento(r.Context(), id, ev.OwnerID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Evento não encontrado.")
			return
		}
		log.Printf("api: deletar evento: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao excluir o evento.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func parseEventID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID de evento inválido.")
		return 0, false
	}
	return id, true
}

func generatePIN() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "000000"
	}
	digits := make([]byte, 6)
	for i, x := range b {
		digits[i] = '0' + x%10
	}
	return string(digits)
}
