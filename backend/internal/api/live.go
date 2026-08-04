package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"devopsconecta/backend/internal/auth"
	"devopsconecta/backend/internal/store"
)

const liveViewerTokenHours = 12

// --- Admin (autenticado, dono do evento) ---

type liveSetQuestionRequest struct {
	QuestionID string `json:"questionId"`
}

func (a *API) handleLiveSetQuestion(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	var req liveSetQuestionRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	questionID, err := strconv.ParseInt(req.QuestionID, 10, 64)
	if err != nil || questionID <= 0 {
		writeError(w, http.StatusBadRequest, "ID de pergunta inválido.")
		return
	}
	a.live.SetCurrentQuestion(eventID, questionID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type liveRevealRequest struct {
	QuestionID    string `json:"questionId"`
	ParticipantID string `json:"participantId"`
}

func (a *API) handleLiveReveal(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	var req liveRevealRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	questionID, err1 := strconv.ParseInt(req.QuestionID, 10, 64)
	participantID, err2 := strconv.ParseInt(req.ParticipantID, 10, 64)
	if err1 != nil || err2 != nil || questionID <= 0 || participantID <= 0 {
		writeError(w, http.StatusBadRequest, "Dados inválidos.")
		return
	}
	a.live.Reveal(eventID, questionID, participantID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type liveRevealAllRequest struct {
	QuestionID string `json:"questionId"`
}

func (a *API) handleLiveRevealAll(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	var req liveRevealAllRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	questionID, err := strconv.ParseInt(req.QuestionID, 10, 64)
	if err != nil || questionID <= 0 {
		writeError(w, http.StatusBadRequest, "ID de pergunta inválido.")
		return
	}
	participants, err := a.store.ListParticipantsByEvent(r.Context(), eventID)
	if err != nil {
		log.Printf("api: listar participantes para revelar todos: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	ids := make([]int64, len(participants))
	for i, p := range participants {
		ids[i] = p.ID
	}
	a.live.RevealAll(eventID, questionID, ids)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type liveResetRequest struct {
	QuestionID string `json:"questionId"`
}

func (a *API) handleLiveReset(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	var req liveResetRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	questionID, err := strconv.ParseInt(req.QuestionID, 10, 64)
	if err != nil || questionID <= 0 {
		writeError(w, http.StatusBadRequest, "ID de pergunta inválido.")
		return
	}
	a.live.Reset(eventID, questionID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// --- Público (sem login, PIN + e-mail) ---

type liveJoinRequest struct {
	PINCode string `json:"pinCode"`
	Email   string `json:"email"`
}

// handleLiveJoin autoriza o acesso à apresentação pública com PIN + e-mail.
// PIN errado nunca autoriza. Com o PIN certo: e-mail de quem já respondeu
// vira "participante", qualquer outro e-mail vira "observador" (só assiste).
func (a *API) handleLiveJoin(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}
	var req liveJoinRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	pin := strings.ToUpper(strings.TrimSpace(req.PINCode))
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if pin == "" {
		writeError(w, http.StatusBadRequest, "Informe o PIN do evento.")
		return
	}
	if !isValidEmail(email) {
		writeError(w, http.StatusBadRequest, "Informe um e-mail válido.")
		return
	}

	event, err := a.store.FindEventByID(r.Context(), eventID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Evento não encontrado.")
		return
	}
	if err != nil {
		log.Printf("api: buscar evento para entrada na apresentação: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	if event.PINCode != pin {
		writeError(w, http.StatusUnauthorized, "PIN inválido.")
		return
	}

	role := auth.LiveRoleObserver
	var participantID int64
	participant, err := a.store.FindParticipantByEventAndEmail(r.Context(), eventID, email)
	if err == nil {
		role = auth.LiveRoleParticipant
		participantID = participant.ID
	} else if !errors.Is(err, store.ErrNotFound) {
		log.Printf("api: buscar participante para entrada na apresentação: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}

	token, err := auth.NewLiveViewerToken(a.cfg.JWTSecret, eventID, participantID, role, liveViewerTokenHours)
	if err != nil {
		log.Printf("api: gerar token de visitante: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "role": role})
}

// resolveLiveViewer valida o token de visitante (query param) e confere que
// ele foi emitido pro evento da URL.
func (a *API) resolveLiveViewer(w http.ResponseWriter, r *http.Request) (int64, auth.LiveClaims, bool) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return 0, auth.LiveClaims{}, false
	}
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	claims, err := auth.VerifyLiveViewerToken(a.cfg.JWTSecret, token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Sessão inválida.")
		return 0, auth.LiveClaims{}, false
	}
	claimEventID, err := strconv.ParseInt(claims.EventID, 10, 64)
	if err != nil || claimEventID != eventID {
		writeError(w, http.StatusUnauthorized, "Sessão inválida.")
		return 0, auth.LiveClaims{}, false
	}
	return eventID, claims, true
}

func (a *API) handleLiveState(w http.ResponseWriter, r *http.Request) {
	eventID, _, ok := a.resolveLiveViewer(w, r)
	if !ok {
		return
	}
	snapshot, err := a.buildLiveSnapshot(r.Context(), eventID)
	if err != nil {
		log.Printf("api: montar snapshot da apresentação: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

// handleLiveStream mantém uma conexão SSE aberta, mandando o snapshot atual
// assim que conecta e de novo a cada mudança no estado da apresentação.
func (a *API) handleLiveStream(w http.ResponseWriter, r *http.Request) {
	eventID, _, ok := a.resolveLiveViewer(w, r)
	if !ok {
		return
	}
	flusher, isFlusher := w.(http.Flusher)
	if !isFlusher {
		writeError(w, http.StatusInternalServerError, "Streaming não suportado.")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	writeSnapshot := func() bool {
		snapshot, err := a.buildLiveSnapshot(r.Context(), eventID)
		if err != nil {
			log.Printf("api: montar snapshot da apresentação: %v", err)
			return false
		}
		data, err := json.Marshal(snapshot)
		if err != nil {
			log.Printf("api: serializar snapshot da apresentação: %v", err)
			return false
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	if !writeSnapshot() {
		return
	}

	sub, unsubscribe := a.live.Subscribe(eventID)
	defer unsubscribe()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-sub:
			if !writeSnapshot() {
				return
			}
		}
	}
}

// --- Snapshot: ponto crítico de segurança ---
//
// Quem não foi revelado na pergunta atual nunca aparece com dado de
// resposta no payload — só em Pending, com identidade e nada mais. Isso é
// filtrado aqui no servidor, não escondido só no client.

type liveOptionDTO struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type liveQuestionDTO struct {
	ID      string          `json:"id"`
	Title   string          `json:"title"`
	Type    string          `json:"type"`
	Options []liveOptionDTO `json:"options"`
}

type liveParticipantDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Photo string `json:"photo"`
}

type liveGroupDTO struct {
	Label        string               `json:"label"`
	Participants []liveParticipantDTO `json:"participants"`
}

type liveSnapshotDTO struct {
	EventTitle        string               `json:"eventTitle"`
	Questions         []liveQuestionDTO    `json:"questions"`
	CurrentQuestionID string               `json:"currentQuestionId"`
	Pending           []liveParticipantDTO `json:"pending"`
	Groups            []liveGroupDTO       `json:"groups"`
}

func toLiveParticipantDTO(p store.Participant) liveParticipantDTO {
	return liveParticipantDTO{ID: strconv.FormatInt(p.ID, 10), Email: p.Email, Photo: p.Photo}
}

type revealedLiveAnswer struct {
	participant store.Participant
	optionID    int64
	text        string
}

func (a *API) buildLiveSnapshot(ctx context.Context, eventID int64) (liveSnapshotDTO, error) {
	event, err := a.store.FindEventByID(ctx, eventID)
	if err != nil {
		return liveSnapshotDTO{}, fmt.Errorf("buscar evento: %w", err)
	}
	questions, err := a.store.ListQuestionsByEvent(ctx, eventID)
	if err != nil {
		return liveSnapshotDTO{}, fmt.Errorf("listar perguntas: %w", err)
	}
	participants, err := a.store.ListParticipantsByEvent(ctx, eventID)
	if err != nil {
		return liveSnapshotDTO{}, fmt.Errorf("listar participantes: %w", err)
	}

	qDTOs := make([]liveQuestionDTO, 0, len(questions))
	var current *store.QuestionWithOptions
	state := a.live.Get(eventID)
	currentQuestionID := state.CurrentQuestionID
	if currentQuestionID == 0 && len(questions) > 0 {
		currentQuestionID = questions[0].ID
	}
	for i := range questions {
		q := &questions[i]
		opts := make([]liveOptionDTO, 0, len(q.Options))
		for _, o := range q.Options {
			opts = append(opts, liveOptionDTO{ID: strconv.FormatInt(o.ID, 10), Text: o.TextLabel})
		}
		qDTOs = append(qDTOs, liveQuestionDTO{
			ID: strconv.FormatInt(q.ID, 10), Title: q.Title, Type: q.Type, Options: opts,
		})
		if q.ID == currentQuestionID {
			current = q
		}
	}

	// Pending/Groups sempre inicializados como slice vazio (nunca nil) — um
	// nil aqui serializa como JSON null, e o frontend espera sempre um array.
	snapshot := liveSnapshotDTO{
		EventTitle: event.Title,
		Questions:  qDTOs,
		Pending:    []liveParticipantDTO{},
		Groups:     []liveGroupDTO{},
	}
	if current == nil {
		return snapshot, nil
	}
	snapshot.CurrentQuestionID = strconv.FormatInt(current.ID, 10)

	revealed := state.Revealed[current.ID]
	pending := make([]liveParticipantDTO, 0, len(participants))
	revealedAnswers := make([]revealedLiveAnswer, 0, len(participants))
	for _, p := range participants {
		if !revealed[p.ID] {
			pending = append(pending, toLiveParticipantDTO(p))
			continue
		}
		answers, err := a.store.ListAnswersByParticipant(ctx, p.ID)
		if err != nil {
			return liveSnapshotDTO{}, fmt.Errorf("listar respostas do participante: %w", err)
		}
		var found *store.Answer
		for i := range answers {
			if answers[i].QuestionID == current.ID {
				found = &answers[i]
				break
			}
		}
		if found == nil {
			// revelado mas sem resposta pra essa pergunta ainda — trata
			// como pendente pra não travar o snapshot.
			pending = append(pending, toLiveParticipantDTO(p))
			continue
		}
		revealedAnswers = append(revealedAnswers, revealedLiveAnswer{participant: p, optionID: found.OptionID, text: found.FreeText})
	}

	snapshot.Pending = pending
	snapshot.Groups = buildLiveGroups(current, revealedAnswers)
	return snapshot, nil
}

func buildLiveGroups(q *store.QuestionWithOptions, revealed []revealedLiveAnswer) []liveGroupDTO {
	if q.Type == questionTypeOpenText {
		order := make([]string, 0, len(revealed))
		byLabel := make(map[string]*liveGroupDTO, len(revealed))
		for _, r := range revealed {
			label := normalizeOpenText(r.text)
			if label == "" {
				label = "—"
			}
			g, ok := byLabel[label]
			if !ok {
				g = &liveGroupDTO{Label: label, Participants: []liveParticipantDTO{}}
				byLabel[label] = g
				order = append(order, label)
			}
			g.Participants = append(g.Participants, toLiveParticipantDTO(r.participant))
		}
		groups := make([]liveGroupDTO, 0, len(order))
		for _, label := range order {
			groups = append(groups, *byLabel[label])
		}
		return groups
	}

	order := make([]int64, 0, len(q.Options))
	byOption := make(map[int64]*liveGroupDTO, len(q.Options))
	for _, opt := range q.Options {
		byOption[opt.ID] = &liveGroupDTO{Label: opt.TextLabel, Participants: []liveParticipantDTO{}}
		order = append(order, opt.ID)
	}
	for _, r := range revealed {
		g, ok := byOption[r.optionID]
		if !ok {
			continue
		}
		g.Participants = append(g.Participants, toLiveParticipantDTO(r.participant))
	}
	groups := make([]liveGroupDTO, 0, len(order))
	for _, id := range order {
		groups = append(groups, *byOption[id])
	}
	return groups
}

// normalizeOpenText espelha o normalizeOpenText do frontend
// (Presentation.svelte): trim + primeira letra maiúscula, pra agrupar quem
// respondeu a mesma coisa no mesmo balão.
func normalizeOpenText(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}
	r := []rune(trimmed)
	return strings.ToUpper(string(r[0])) + string(r[1:])
}
