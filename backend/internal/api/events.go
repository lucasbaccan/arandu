package api

import (
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
	eventStatusPreparation = "PREPARATION"
	maxEventTitleLength    = 120
	autoPINRetries         = 5
)

var pinRe = regexp.MustCompile(`^[a-zA-Z0-9_\-]{1,25}$`)

type eventDTO struct {
	ID          string `json:"id"`
	OwnerID     string `json:"ownerId"`
	Title       string `json:"title"`
	PINCode     string `json:"pinCode"`
	Status      string `json:"status"`
	ShowRanking bool   `json:"configShowRanking"`
	CreatedAt   string `json:"createdAt"`
}

func toEventDTO(e store.Event) eventDTO {
	return eventDTO{
		ID:          strconv.FormatInt(e.ID, 10),
		OwnerID:     strconv.FormatInt(e.OwnerID, 10),
		Title:       e.Title,
		PINCode:     e.PINCode,
		Status:      e.Status,
		ShowRanking: e.ShowRanking,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
	}
}

type createEventRequest struct {
	Title   string `json:"title"`
	PINCode string `json:"pinCode"`
}

func (a *API) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	var req createEventRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.PINCode = strings.TrimSpace(req.PINCode)

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
		ev, err := a.store.CreateEvent(r.Context(), store.Event{
			ID:      a.ids.NextID(),
			OwnerID: ownerID,
			Title:   req.Title,
			PINCode: pin,
			Status:  eventStatusPreparation,
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

func (a *API) handleListEvents(w http.ResponseWriter, r *http.Request) {
	ownerID := userIDFromContext(r.Context())
	events, err := a.store.ListEventsByOwner(r.Context(), ownerID)
	if err != nil {
		log.Printf("api: listar eventos: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao listar eventos.")
		return
	}
	dtos := make([]eventDTO, 0, len(events))
	for _, e := range events {
		dtos = append(dtos, toEventDTO(e))
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": dtos})
}

func (a *API) handleGetEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseEventID(w, r)
	if !ok {
		return
	}
	ownerID := userIDFromContext(r.Context())
	ev, err := a.store.FindEventByIDAndOwner(r.Context(), id, ownerID)
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
	ShowRanking bool   `json:"configShowRanking"`
}

func (a *API) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseEventID(w, r)
	if !ok {
		return
	}
	ownerID := userIDFromContext(r.Context())

	var req updateEventRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.PINCode = strings.TrimSpace(req.PINCode)

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

	current, err := a.store.FindEventByIDAndOwner(r.Context(), id, ownerID)
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

	ev, err := a.store.UpdateEvent(r.Context(), store.Event{
		ID:          id,
		OwnerID:     ownerID,
		Title:       req.Title,
		PINCode:     pin,
		Status:      current.Status,
		ShowRanking: req.ShowRanking,
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
