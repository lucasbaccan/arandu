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
	"time"

	"devopsconecta/backend/internal/auth"
	"devopsconecta/backend/internal/live"
	"devopsconecta/backend/internal/store"
)

const liveViewerTokenHours = 12

// allowedReactionEmojis é o mesmo conjunto usado no frontend
// (frontend/src/components/ReactionBar.svelte) — mudar um lado exige mudar o
// outro.
var allowedReactionEmojis = map[string]bool{
	"👍": true, "❤️": true, "😂": true, "🎉": true, "👏": true,
}

const (
	maxLiveMessageLength = 300
	maxLiveQATextLength  = 500
)

// --- Admin (autenticado, dono do evento) ---

type liveSetQuestionRequest struct {
	QuestionID string `json:"questionId"`
}

func (a *API) handleAoVivoDefinirPergunta(w http.ResponseWriter, r *http.Request) {
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
	if err := a.store.DefinirPerguntaAtualDoEvento(r.Context(), eventID, questionID); err != nil {
		log.Printf("api: salvar pergunta atual: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.SetCurrentQuestion(eventID, questionID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type liveRevealRequest struct {
	QuestionID    string `json:"questionId"`
	ParticipantID string `json:"participantId"`
}

func (a *API) handleAoVivoRevelar(w http.ResponseWriter, r *http.Request) {
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
	if err := a.store.RevelarResposta(r.Context(), eventID, questionID, participantID); err != nil {
		log.Printf("api: salvar revelação: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.Reveal(eventID, questionID, participantID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *API) handleAoVivoOcultar(w http.ResponseWriter, r *http.Request) {
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
	if err := a.store.OcultarResposta(r.Context(), eventID, questionID, participantID); err != nil {
		log.Printf("api: salvar desrevelação: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.Unreveal(eventID, questionID, participantID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type liveRevealAllRequest struct {
	QuestionID string `json:"questionId"`
}

func (a *API) handleAoVivoRevelarTodos(w http.ResponseWriter, r *http.Request) {
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
	participants, err := a.store.ListarParticipantesPorEvento(r.Context(), eventID)
	if err != nil {
		log.Printf("api: listar participantes para revelar todos: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	ids := make([]int64, len(participants))
	for i, p := range participants {
		ids[i] = p.ID
	}
	if err := a.store.RevelarTodasRespostas(r.Context(), eventID, questionID, ids); err != nil {
		log.Printf("api: salvar revelar todos: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.RevealAll(eventID, questionID, ids)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type liveResetRequest struct {
	QuestionID string `json:"questionId"`
}

func (a *API) handleAoVivoReiniciar(w http.ResponseWriter, r *http.Request) {
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
	if err := a.store.ReiniciarReveladosDaPergunta(r.Context(), eventID, questionID); err != nil {
		log.Printf("api: salvar reinício da revelação: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.Reset(eventID, questionID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// handleAoVivoReiniciarTudo limpa a revelação de todas as perguntas do evento de
// uma vez — botão "Reiniciar tudo" em /stage.
func (a *API) handleAoVivoReiniciarTudo(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	if err := a.store.ReiniciarReveladosDoEvento(r.Context(), eventID); err != nil {
		log.Printf("api: salvar reinício de toda a revelação: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.ResetAll(eventID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type liveSetBlankedRequest struct {
	Blanked bool `json:"blanked"`
}

func (a *API) handleAoVivoDefinirEmBranco(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	var req liveSetBlankedRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.DefinirEmBrancoDoEvento(r.Context(), eventID, req.Blanked); err != nil {
		log.Printf("api: salvar tela em branco: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.SetBlanked(eventID, req.Blanked)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type liveSetAnswersHiddenRequest struct {
	Hidden bool `json:"hidden"`
}

func (a *API) handleAoVivoOcultarRespostas(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	var req liveSetAnswersHiddenRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.DefinirRespostasOcultasDoEvento(r.Context(), eventID, req.Hidden); err != nil {
		log.Printf("api: salvar ocultar respostas: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.SetAnswersHidden(eventID, req.Hidden)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type liveSetNamesHiddenRequest struct {
	Hidden bool `json:"hidden"`
}

// handleAoVivoOcultarNomes liga/desliga a legenda de nome sob cada rosto na
// janela de apresentação (/stage/{id}/present) e na tela da plateia
// (/audience/{id}) — o painel do organizador (/stage) sempre mostra os
// nomes, esse switch não afeta ele.
func (a *API) handleAoVivoOcultarNomes(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	var req liveSetNamesHiddenRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.DefinirNomesOcultosDoEvento(r.Context(), eventID, req.Hidden); err != nil {
		log.Printf("api: salvar ocultar nomes: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.SetNamesHidden(eventID, req.Hidden)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type liveSetPresentDensityModeRequest struct {
	Mode string `json:"mode"`
}

// validPresentDensityModes: "" = automático, "smart" = testa todas as
// colunas e fica com a maior escala, "1".."4" = força esse nº de colunas —
// ver PresentationStage.svelte (forceCols/smart) e fitDensity lá.
var validPresentDensityModes = map[string]bool{
	"": true, "smart": true, "1": true, "2": true, "3": true, "4": true,
}

// handleAoVivoDefinirModoDensidade escolhe o modo de densidade do placar de
// respostas (colunas × escala das pílulas) — reflete em tempo real tanto na
// janela de apresentação (/stage/{id}/present) quanto na tela pública
// (/audience/{id}), já que as duas leem o mesmo snapshot (buildLiveSnapshot).
// Escolhido pelos botões no rodapé de /stage.
func (a *API) handleAoVivoDefinirModoDensidade(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	var req liveSetPresentDensityModeRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !validPresentDensityModes[req.Mode] {
		writeError(w, http.StatusBadRequest, "Modo de apresentação inválido.")
		return
	}
	if err := a.store.DefinirModoDensidadeDoEvento(r.Context(), eventID, req.Mode); err != nil {
		log.Printf("api: salvar modo de densidade da apresentação: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.SetPresentDensityMode(eventID, req.Mode)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type liveSetMessageRequest struct {
	Message string `json:"message"`
}

func (a *API) handleAoVivoDefinirMensagem(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	var req liveSetMessageRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	msg := strings.TrimSpace(req.Message)
	if len(msg) > maxLiveMessageLength {
		writeError(w, http.StatusBadRequest, "Mensagem muito longa (máximo "+strconv.Itoa(maxLiveMessageLength)+" caracteres).")
		return
	}
	if err := a.store.DefinirMensagemDoEvento(r.Context(), eventID, msg); err != nil {
		log.Printf("api: salvar aviso: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.SetMessage(eventID, msg)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type liveSetInteractionsRequest struct {
	Enabled bool `json:"enabled"`
}

func (a *API) handleAoVivoDefinirInteracoes(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	var req liveSetInteractionsRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.DefinirInteracoesHabilitadas(r.Context(), eventID, req.Enabled); err != nil {
		log.Printf("api: alternar interações: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.Touch(eventID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *API) handleAoVivoDispensarPergunta(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	messageID, err := strconv.ParseInt(r.PathValue("messageId"), 10, 64)
	if err != nil || messageID <= 0 {
		writeError(w, http.StatusBadRequest, "ID de mensagem inválido.")
		return
	}
	msg, err := a.store.BuscarPerguntaAoVivoPorID(r.Context(), messageID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Mensagem não encontrada.")
		return
	}
	if err != nil {
		log.Printf("api: buscar mensagem de q&a: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	if msg.EventID != eventID {
		writeError(w, http.StatusNotFound, "Mensagem não encontrada.")
		return
	}
	if err := a.store.DispensarPerguntaAoVivo(r.Context(), messageID); err != nil {
		log.Printf("api: dispensar mensagem de q&a: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.Touch(eventID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// handleAoVivoRemoverPergunta deixa QUEM PERGUNTOU remover a própria pergunta: o
// navegador envia o mesmo clientId usado no envio (identificador persistido
// pelo frontend) e a mensagem só é removida se o clientId bater e pertencer
// ao evento da URL. Qualquer divergência responde "não encontrada" pra não
// revelar a existência da mensagem a terceiros.
func (a *API) handleAoVivoRemoverPergunta(w http.ResponseWriter, r *http.Request) {
	eventID, _, ok := a.resolveLiveViewer(w, r)
	if !ok {
		return
	}
	messageID, err := strconv.ParseInt(r.PathValue("messageId"), 10, 64)
	if err != nil || messageID <= 0 {
		writeError(w, http.StatusBadRequest, "ID de mensagem inválido.")
		return
	}
	var req struct {
		ClientID string `json:"clientId"`
	}
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	clientID := strings.TrimSpace(req.ClientID)
	if clientID == "" {
		writeError(w, http.StatusBadRequest, "Identificador do navegador ausente.")
		return
	}
	msg, err := a.store.BuscarPerguntaAoVivoPorID(r.Context(), messageID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Mensagem não encontrada.")
		return
	}
	if err != nil {
		log.Printf("api: buscar mensagem de q&a: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	if msg.EventID != eventID || msg.ClientID != clientID {
		writeError(w, http.StatusNotFound, "Mensagem não encontrada.")
		return
	}
	if err := a.store.RemoverPerguntaAoVivoPorCliente(r.Context(), messageID, clientID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Mensagem não encontrada.")
			return
		}
		log.Printf("api: remover mensagem de q&a: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.Touch(eventID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// handleAoVivoEstadoAdmin é o equivalente autenticado de handleAoVivoEstado — deixa
// buscar o snapshot do organizador sem abrir SSE (útil pra testes e como
// fallback), espelhando o par state/stream que já existe pro público.
func (a *API) handleAoVivoEstadoAdmin(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	snapshot, err := a.buildAdminLiveSnapshot(r.Context(), eventID)
	if err != nil {
		log.Printf("api: montar snapshot administrativo: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

// handleAoVivoFluxoAdmin espelha handleAoVivoFluxo, mas pro organizador
// autenticado: snapshot administrativo (inclui a caixa de Q&A privada) + a
// mesma fila de reações ao vivo que os participantes veem.
func (a *API) handleAoVivoFluxoAdmin(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
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
		snapshot, err := a.buildAdminLiveSnapshot(r.Context(), eventID)
		if err != nil {
			log.Printf("api: montar snapshot administrativo: %v", err)
			return false
		}
		data, err := json.Marshal(snapshot)
		if err != nil {
			log.Printf("api: serializar snapshot administrativo: %v", err)
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
	reactionSub, unsubscribeReactions := a.live.SubscribeReactions(eventID)
	defer unsubscribeReactions()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-sub:
			if !writeSnapshot() {
				return
			}
		case ev := <-reactionSub:
			if !writeReactionSSE(w, flusher, ev) {
				return
			}
		}
	}
}

// handleAoVivoEstadoApresentacao é o par autenticado de handleAoVivoEstado: mesmo
// snapshot somente-leitura (pergunta atual, pendentes, grupos revelados) que
// a plateia vê, mas restrito ao dono do evento — sem PIN nem token de
// visitante. Alimenta a janela "Modo apresentação", que só o organizador
// pode abrir.
func (a *API) handleAoVivoEstadoApresentacao(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	snapshot, err := a.buildLiveSnapshot(r.Context(), eventID)
	if err != nil {
		log.Printf("api: montar snapshot de apresentação: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

// handleAoVivoFluxoApresentacao espelha handleAoVivoFluxo (mesmo snapshot,
// atualizado a cada mudança), mas autenticado pro dono do evento — a janela
// de apresentação não tem token de visitante nem PIN.
func (a *API) handleAoVivoFluxoApresentacao(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
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
			log.Printf("api: montar snapshot de apresentação: %v", err)
			return false
		}
		data, err := json.Marshal(snapshot)
		if err != nil {
			log.Printf("api: serializar snapshot de apresentação: %v", err)
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
	reactionSub, unsubscribeReactions := a.live.SubscribeReactions(eventID)
	defer unsubscribeReactions()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-sub:
			if !writeSnapshot() {
				return
			}
		case ev := <-reactionSub:
			if !writeReactionSSE(w, flusher, ev) {
				return
			}
		}
	}
}

// --- Público (sem login, PIN + e-mail) ---

type liveJoinRequest struct {
	PINCode string `json:"pinCode"`
	Email   string `json:"email"`
}

// handleAoVivoEntrar autoriza o acesso à apresentação pública com PIN e,
// opcionalmente, e-mail. PIN errado nunca autoriza. Com o PIN certo: e-mail
// de quem já respondeu vira "participante"; e-mail em branco (entrada como
// convidado) ou de quem não respondeu vira "observador" (só assiste).
func (a *API) handleAoVivoEntrar(w http.ResponseWriter, r *http.Request) {
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
	if email != "" && !isValidEmail(email) {
		writeError(w, http.StatusBadRequest, "Informe um e-mail válido.")
		return
	}

	event, err := a.store.BuscarEventoPorID(r.Context(), eventID)
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
	if email != "" {
		participant, err := a.store.BuscarParticipantePorEventoEEmail(r.Context(), eventID, email)
		if err == nil {
			role = auth.LiveRoleParticipant
			participantID = participant.ID
		} else if !errors.Is(err, store.ErrNotFound) {
			log.Printf("api: buscar participante para entrada na apresentação: %v", err)
			writeError(w, http.StatusInternalServerError, "Erro interno.")
			return
		}
	}

	token, err := auth.NewLiveViewerToken(a.cfg.JWTSecret, eventID, participantID, role, liveViewerTokenHours)
	if err != nil {
		log.Printf("api: gerar token de visitante: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "role": role})
}

type liveReactRequest struct {
	Emoji string `json:"emoji"`
}

// handleAoVivoReagir transmite uma reação de emoji pra quem estiver assistindo
// (participantes + organizador), sem gravar nada no banco — é efêmera por
// decisão de produto.
func (a *API) handleAoVivoReagir(w http.ResponseWriter, r *http.Request) {
	eventID, _, ok := a.resolveLiveViewer(w, r)
	if !ok {
		return
	}
	var req liveReactRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !allowedReactionEmojis[req.Emoji] {
		writeError(w, http.StatusBadRequest, "Emoji não permitido.")
		return
	}
	event, err := a.store.BuscarEventoPorID(r.Context(), eventID)
	if err != nil {
		log.Printf("api: buscar evento para reação: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	if !event.InteractionsEnabled {
		writeError(w, http.StatusForbidden, "As interações estão desativadas no momento.")
		return
	}
	a.live.BroadcastReaction(eventID, req.Emoji)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type liveSubmitQARequest struct {
	Text string `json:"text"`
	// ClientID identifica o navegador que mandou (UUID persistido pelo
	// frontend) — é o que permite a própria pessoa remover a pergunta depois.
	ClientID string `json:"clientId"`
}

// handleAoVivoEnviarPergunta grava uma pergunta/recado de um participante pro
// organizador. Fica só na caixa de entrada privada do organizador — nunca é
// exposta a outros participantes nem no snapshot público.
func (a *API) handleAoVivoEnviarPergunta(w http.ResponseWriter, r *http.Request) {
	eventID, claims, ok := a.resolveLiveViewer(w, r)
	if !ok {
		return
	}
	var req liveSubmitQARequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	text := strings.TrimSpace(req.Text)
	if text == "" {
		writeError(w, http.StatusBadRequest, "Escreva uma mensagem.")
		return
	}
	if len(text) > maxLiveQATextLength {
		writeError(w, http.StatusBadRequest, "Mensagem muito longa (máximo "+strconv.Itoa(maxLiveQATextLength)+" caracteres).")
		return
	}
	// clientID é opcional (mensagens antigas/outros clientes não têm); se vier,
	// só precisa ser curto — o dono do navegador usa pra remover a própria
	// pergunta (ver handleAoVivoRemoverPergunta).
	clientID := strings.TrimSpace(req.ClientID)
	if len(clientID) > 64 {
		writeError(w, http.StatusBadRequest, "Identificador do navegador inválido.")
		return
	}
	event, err := a.store.BuscarEventoPorID(r.Context(), eventID)
	if err != nil {
		log.Printf("api: buscar evento para q&a: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	if !event.InteractionsEnabled {
		writeError(w, http.StatusForbidden, "As interações estão desativadas no momento.")
		return
	}

	var participantID int64
	if claims.Role == auth.LiveRoleParticipant && claims.ParticipantID != "" {
		participantID, _ = strconv.ParseInt(claims.ParticipantID, 10, 64)
	}
	messageID := a.ids.NextID()
	if _, err := a.store.CriarPerguntaAoVivo(r.Context(), store.LiveQAMessage{
		ID:            messageID,
		EventID:       eventID,
		ParticipantID: participantID,
		ClientID:      clientID,
		Text:          text,
	}); err != nil {
		log.Printf("api: gravar mensagem de q&a: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	a.live.Touch(eventID)
	// Devolve o id e o texto pra o próprio navegador poder listar/remover a
	// pergunta dele (o snapshot público não expõe a caixa de entrada).
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "messageId": strconv.FormatInt(messageID, 10), "text": text})
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

func (a *API) handleAoVivoEstado(w http.ResponseWriter, r *http.Request) {
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

// handleAoVivoFluxo mantém uma conexão SSE aberta, mandando o snapshot atual
// assim que conecta e de novo a cada mudança no estado da apresentação.
func (a *API) handleAoVivoFluxo(w http.ResponseWriter, r *http.Request) {
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
	reactionSub, unsubscribeReactions := a.live.SubscribeReactions(eventID)
	defer unsubscribeReactions()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-sub:
			if !writeSnapshot() {
				return
			}
		case ev := <-reactionSub:
			if !writeReactionSSE(w, flusher, ev) {
				return
			}
		}
	}
}

// writeReactionSSE escreve um evento SSE nomeado ("reaction"), separado dos
// frames default de snapshot — o frontend escuta com
// EventSource.addEventListener('reaction', ...) ao lado do onmessage normal.
func writeReactionSSE(w http.ResponseWriter, flusher http.Flusher, ev live.ReactionEvent) bool {
	data, err := json.Marshal(map[string]string{"emoji": ev.Emoji})
	if err != nil {
		return false
	}
	if _, err := fmt.Fprintf(w, "event: reaction\ndata: %s\n\n", data); err != nil {
		return false
	}
	flusher.Flush()
	return true
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
	Name  string `json:"name"`
	Photo string `json:"photo"`
}

type liveGroupDTO struct {
	Label        string               `json:"label"`
	Participants []liveParticipantDTO `json:"participants"`
}

type liveSnapshotDTO struct {
	EventTitle          string               `json:"eventTitle"`
	Questions           []liveQuestionDTO    `json:"questions"`
	CurrentQuestionID   string               `json:"currentQuestionId"`
	Pending             []liveParticipantDTO `json:"pending"`
	Groups              []liveGroupDTO       `json:"groups"`
	Blanked             bool                 `json:"blanked"`
	Message             string               `json:"message"`
	InteractionsEnabled bool                 `json:"interactionsEnabled"`
	AnswersHidden       bool                 `json:"answersHidden"`
	NamesHidden         bool                 `json:"namesHidden"`
	// PresentDensityMode escolhe colunas × escala do placar de respostas em
	// PresentationStage.svelte — ver live.EventState.PresentDensityMode.
	PresentDensityMode string `json:"presentDensityMode"`
}

type liveQAMessageDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
}

type liveAdminSnapshotDTO struct {
	Blanked             bool                `json:"blanked"`
	Message             string              `json:"message"`
	InteractionsEnabled bool                `json:"interactionsEnabled"`
	AnswersHidden       bool                `json:"answersHidden"`
	NamesHidden         bool                `json:"namesHidden"`
	PresentDensityMode  string              `json:"presentDensityMode"`
	CurrentQuestionID   string              `json:"currentQuestionId"`
	// Revealed mapeia questionID (string) -> participantIDs (string) já
	// revelados nessa pergunta — deixa o painel do organizador retomar de
	// onde parou depois de um F5 ou ao reabrir /stage, em vez de sempre
	// recomeçar do zero.
	Revealed map[string][]string `json:"revealed"`
	QAInbox  []liveQAMessageDTO  `json:"qaInbox"`
}

// buildAdminLiveSnapshot cobre o que só existe do lado do servidor:
// blank/aviso/interações/revelação (pra restaurar depois de um F5 ou ao
// reabrir /stage) e a caixa de Q&A privada.
func (a *API) buildAdminLiveSnapshot(ctx context.Context, eventID int64) (liveAdminSnapshotDTO, error) {
	event, err := a.store.BuscarEventoPorID(ctx, eventID)
	if err != nil {
		return liveAdminSnapshotDTO{}, fmt.Errorf("buscar evento: %w", err)
	}
	state := a.live.Get(eventID)
	messages, err := a.store.ListarPerguntasAoVivoPorEvento(ctx, eventID)
	if err != nil {
		return liveAdminSnapshotDTO{}, fmt.Errorf("listar mensagens de q&a: %w", err)
	}

	qaDTOs := make([]liveQAMessageDTO, 0, len(messages))
	for _, m := range messages {
		email := ""
		name := ""
		if m.ParticipantID != 0 {
			if p, err := a.store.BuscarParticipantePorID(ctx, m.ParticipantID); err == nil {
				email = p.Email
				name = p.Name
			}
		}
		qaDTOs = append(qaDTOs, liveQAMessageDTO{
			ID:        strconv.FormatInt(m.ID, 10),
			Email:     email,
			Name:      name,
			Text:      m.Text,
			CreatedAt: m.CreatedAt.Format(time.RFC3339),
		})
	}

	revealed := make(map[string][]string, len(state.Revealed))
	for questionID, set := range state.Revealed {
		ids := make([]string, 0, len(set))
		for participantID := range set {
			ids = append(ids, strconv.FormatInt(participantID, 10))
		}
		revealed[strconv.FormatInt(questionID, 10)] = ids
	}

	currentQuestionID := ""
	if state.CurrentQuestionID != 0 {
		currentQuestionID = strconv.FormatInt(state.CurrentQuestionID, 10)
	}

	return liveAdminSnapshotDTO{
		Blanked:             state.Blanked,
		Message:             state.Message,
		InteractionsEnabled: event.InteractionsEnabled,
		AnswersHidden:       state.AnswersHidden,
		NamesHidden:         state.NamesHidden,
		PresentDensityMode:  state.PresentDensityMode,
		CurrentQuestionID:   currentQuestionID,
		Revealed:            revealed,
		QAInbox:             qaDTOs,
	}, nil
}

func (a *API) toLiveParticipantDTO(p store.Participant) liveParticipantDTO {
	return liveParticipantDTO{ID: strconv.FormatInt(p.ID, 10), Email: p.Email, Name: p.Name, Photo: a.photoURL(p)}
}

type revealedLiveAnswer struct {
	participant store.Participant
	optionID    int64
	text        string
}

func (a *API) buildLiveSnapshot(ctx context.Context, eventID int64) (liveSnapshotDTO, error) {
	event, err := a.store.BuscarEventoPorID(ctx, eventID)
	if err != nil {
		return liveSnapshotDTO{}, fmt.Errorf("buscar evento: %w", err)
	}
	questions, err := a.store.ListarPerguntasPorEvento(ctx, eventID)
	if err != nil {
		return liveSnapshotDTO{}, fmt.Errorf("listar perguntas: %w", err)
	}
	participants, err := a.store.ListarParticipantesPorEvento(ctx, eventID)
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
		EventTitle:          event.Title,
		Questions:           qDTOs,
		Pending:             []liveParticipantDTO{},
		Groups:              []liveGroupDTO{},
		Blanked:             state.Blanked,
		Message:             state.Message,
		InteractionsEnabled: event.InteractionsEnabled,
		AnswersHidden:       state.AnswersHidden,
		NamesHidden:         state.NamesHidden,
		PresentDensityMode:  state.PresentDensityMode,
	}
	if current == nil {
		return snapshot, nil
	}
	snapshot.CurrentQuestionID = strconv.FormatInt(current.ID, 10)

	revealed := state.Revealed[current.ID]
	pending := make([]liveParticipantDTO, 0, len(participants))
	revealedAnswers := make([]revealedLiveAnswer, 0, len(participants))
	// allOpenTexts só é preenchido pra perguntas de resposta aberta: ao
	// contrário de múltipla escolha (cujas opções já existem de antemão e
	// sempre aparecem, mesmo com 0 pessoas), os "baldes" de resposta aberta
	// só existem pelo que foi digitado — então pra eles aparecerem na tela
	// antes de qualquer revelação (igual às opções fixas), é preciso saber
	// o texto de todo mundo que respondeu, revelado ou não. A identidade de
	// quem ainda não foi revelado nunca é anexada a esse texto no payload —
	// só usamos aqui pra descobrir quais baldes existem.
	var allOpenTexts []string
	for _, p := range participants {
		if !revealed[p.ID] {
			pending = append(pending, a.toLiveParticipantDTO(p))
			if current.Type == questionTypeOpenText {
				if text, ok := freeTextAnswerFor(ctx, a.store, p.ID, current.ID); ok {
					allOpenTexts = append(allOpenTexts, text)
				}
			}
			continue
		}
		answers, err := a.store.ListarRespostasPorParticipante(ctx, p.ID)
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
			pending = append(pending, a.toLiveParticipantDTO(p))
			continue
		}
		revealedAnswers = append(revealedAnswers, revealedLiveAnswer{participant: p, optionID: found.OptionID, text: found.FreeText})
		if current.Type == questionTypeOpenText {
			allOpenTexts = append(allOpenTexts, found.FreeText)
		}
	}

	snapshot.Pending = pending
	snapshot.Groups = a.buildLiveGroups(current, revealedAnswers, allOpenTexts)
	return snapshot, nil
}

// freeTextAnswerFor busca a resposta de texto livre de um participante pra
// uma pergunta específica, sem expor o restante das respostas dele.
func freeTextAnswerFor(ctx context.Context, st *store.Store, participantID, questionID int64) (string, bool) {
	answers, err := st.ListarRespostasPorParticipante(ctx, participantID)
	if err != nil {
		return "", false
	}
	for i := range answers {
		if answers[i].QuestionID == questionID {
			return answers[i].FreeText, true
		}
	}
	return "", false
}

func (a *API) buildLiveGroups(q *store.QuestionWithOptions, revealed []revealedLiveAnswer, allOpenTexts []string) []liveGroupDTO {
	if q.Type == questionTypeOpenText {
		order := make([]string, 0, len(allOpenTexts))
		byLabel := make(map[string]*liveGroupDTO, len(allOpenTexts))
		for _, text := range allOpenTexts {
			label := normalizeOpenText(text)
			if label == "" {
				label = "—"
			}
			if _, ok := byLabel[label]; !ok {
				byLabel[label] = &liveGroupDTO{Label: label, Participants: []liveParticipantDTO{}}
				order = append(order, label)
			}
		}
		for _, r := range revealed {
			label := normalizeOpenText(r.text)
			if label == "" {
				label = "—"
			}
			g, ok := byLabel[label]
			if !ok {
				// não deveria acontecer (allOpenTexts inclui os revelados
				// também), mas cria o balde se faltar por algum motivo.
				g = &liveGroupDTO{Label: label, Participants: []liveParticipantDTO{}}
				byLabel[label] = g
				order = append(order, label)
			}
			g.Participants = append(g.Participants, a.toLiveParticipantDTO(r.participant))
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
		g.Participants = append(g.Participants, a.toLiveParticipantDTO(r.participant))
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
