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

// sseHeartbeatInterval é o período do "ping" que mantém viva cada conexão SSE
// ao vivo. Muitos proxies/firewalls corporativos derrubam conexão ociosa
// (deixando o socket "meio morto"): o comentário periódico gera tráfego, força
// o buffer do proxy a andar e faz o servidor detectar cliente desconectado.
const sseHeartbeatInterval = 15 * time.Second

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

// handleAoVivoDefinirPergunta godoc
//
// @Summary     Define a pergunta atual da apresentação
// @Tags        ao-vivo (admin)
// @Accept      json
// @Produce     json
// @Param       id   path string                 true "ID do evento"
// @Param       body body liveSetQuestionRequest true "ID da pergunta"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/pergunta [post]
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

// handleAoVivoRevelar godoc
//
// @Summary     Revela a resposta de um participante na pergunta atual
// @Tags        ao-vivo (admin)
// @Accept      json
// @Produce     json
// @Param       id   path string                true "ID do evento"
// @Param       body body liveRevealRequest      true "Pergunta e participante"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/revelar [post]
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

// handleAoVivoOcultar godoc
//
// @Summary     Desfaz a revelação de um participante na pergunta atual
// @Tags        ao-vivo (admin)
// @Accept      json
// @Produce     json
// @Param       id   path string           true "ID do evento"
// @Param       body body liveRevealRequest true "Pergunta e participante"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/ocultar [post]
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

// handleAoVivoRevelarTodos godoc
//
// @Summary     Revela todos os participantes pendentes na pergunta atual
// @Tags        ao-vivo (admin)
// @Accept      json
// @Produce     json
// @Param       id   path string                  true "ID do evento"
// @Param       body body liveRevealAllRequest    true "ID da pergunta"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/revelar-todos [post]
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
	answers, err := a.store.ListarRespostasPorPergunta(r.Context(), questionID)
	if err != nil {
		log.Printf("api: listar respondentes para revelar todos: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	ids := make([]int64, 0, len(answers))
	for participantID := range answers {
		ids = append(ids, participantID)
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

// handleAoVivoReiniciar godoc
//
// @Summary     Reinicia a revelação de uma pergunta (volta todos a pendente)
// @Tags        ao-vivo (admin)
// @Accept      json
// @Produce     json
// @Param       id   path string             true "ID do evento"
// @Param       body body liveResetRequest   true "ID da pergunta"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/reiniciar [post]
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
// handleAoVivoReiniciarTudo godoc
//
// @Summary     Reinicia a revelação de TODAS as perguntas do evento
// @Tags        ao-vivo (admin)
// @Produce     json
// @Param       id path string true "ID do evento"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/reiniciar-tudo [post]
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

// handleAoVivoDefinirEmBranco godoc
//
// @Summary     Liga/desliga a tela em branco na apresentação
// @Tags        ao-vivo (admin)
// @Accept      json
// @Produce     json
// @Param       id   path string                  true "ID do evento"
// @Param       body body liveSetBlankedRequest   true "blanked"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/em-branco [post]
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

// handleAoVivoOcultarRespostas godoc
//
// @Summary     Liga/desliga a ocultação das respostas na apresentação
// @Tags        ao-vivo (admin)
// @Accept      json
// @Produce     json
// @Param       id   path string                       true "ID do evento"
// @Param       body body liveSetAnswersHiddenRequest  true "hidden"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/ocultar-respostas [post]
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
// handleAoVivoOcultarNomes godoc
//
// @Summary     Liga/desliga a legenda de nome sob cada rosto na apresentação
// @Tags        ao-vivo (admin)
// @Accept      json
// @Produce     json
// @Param       id   path string                     true "ID do evento"
// @Param       body body liveSetNamesHiddenRequest  true "hidden"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/ocultar-nomes [post]
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

// validPresentDensityModes: "modo_amplo" = Amplo, o padrão (testa todas as
// colunas e fica com a maior escala), "modo_compacto" = Compacto, o
// automático antigo (menor nº de colunas que cabe), "1".."4" = força esse nº
// de colunas — ver PresentationStage.svelte (forceCols/amplo) e fitDensity
// lá. "" (legado) segue aceito e os clientes interpretam como o padrão
// (modo_amplo).
var validPresentDensityModes = map[string]bool{
	"": true, "modo_compacto": true, "modo_amplo": true, "1": true, "2": true, "3": true, "4": true,
}

// handleAoVivoDefinirModoDensidade escolhe o modo de densidade do placar de
// respostas (colunas × escala das pílulas) — reflete em tempo real tanto na
// janela de apresentação (/stage/{id}/present) quanto na tela pública
// (/audience/{id}), já que as duas leem o mesmo snapshot (buildLiveSnapshot).
// Escolhido pelos botões no rodapé de /stage.
// handleAoVivoDefinirModoDensidade godoc
//
// @Summary     Define o modo de densidade do placar de respostas
// @Description mode: "" ou "modo_amplo" (padrão, maior escala possível), "modo_compacto" (menor nº de colunas que cabe), ou "1".."4" (força o nº de colunas).
// @Tags        ao-vivo (admin)
// @Accept      json
// @Produce     json
// @Param       id   path string                            true "ID do evento"
// @Param       body body liveSetPresentDensityModeRequest  true "mode"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/modo-densidade [post]
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

// handleAoVivoDefinirMensagem godoc
//
// @Summary     Define (ou limpa, com message vazio) o aviso mostrado na apresentação
// @Tags        ao-vivo (admin)
// @Accept      json
// @Produce     json
// @Param       id   path string                true "ID do evento"
// @Param       body body liveSetMessageRequest  true "message"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/mensagem [post]
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

// handleAoVivoDefinirInteracoes godoc
//
// @Summary     Liga/desliga as interações da plateia (reações e Q&A)
// @Tags        ao-vivo (admin)
// @Accept      json
// @Produce     json
// @Param       id   path string                       true "ID do evento"
// @Param       body body liveSetInteractionsRequest  true "enabled"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/interacoes [post]
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

// handleAoVivoDispensarPergunta godoc
//
// @Summary     Dispensa (arquiva) uma mensagem de Q&A da caixa de entrada do organizador
// @Tags        ao-vivo (admin)
// @Produce     json
// @Param       id        path string true "ID do evento"
// @Param       messageId path string true "ID da mensagem de Q&A"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/perguntas/{messageId}/dispensar [post]
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
// handleAoVivoRemoverPergunta godoc
//
// @Summary     Remove a própria pergunta/recado de Q&A
// @Description Só quem enviou pode remover: o clientId do corpo precisa bater com o clientId salvo no envio.
// @Tags        ao-vivo (público)
// @Accept      json
// @Produce     json
// @Param       id        path  string                                    true "ID do evento"
// @Param       messageId path  string                                    true "ID da mensagem de Q&A"
// @Param       token     query string                                    true "Token de visitante"
// @Param       body      body  removerPerguntaAoVivoRequest              true "clientId de quem enviou"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    liveViewerToken
// @Router      /api/publico/eventos/{id}/ao-vivo/perguntas/{messageId} [delete]
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
// handleAoVivoEstadoAdmin godoc
//
// @Summary     Snapshot administrativo do estado ao vivo (sem SSE)
// @Description Equivalente sem streaming de GET /ao-vivo/fluxo — útil pra buscar o estado uma vez (ex: testes, fallback de polling).
// @Tags        ao-vivo (admin)
// @Produce     json
// @Param       id path string true "ID do evento"
// @Success     200 {object} liveAdminSnapshotDTO
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/estado [get]
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
// handleAoVivoFluxoAdmin godoc
//
// @Summary     Stream (SSE) do estado ao vivo administrativo
// @Description Server-Sent Events: um frame liveAdminSnapshotDTO ao conectar, de novo a cada mudança, e um heartbeat a cada 15s. Não é uma chamada request/response comum — não dá pra "experimentar" pelo Swagger UI (a conexão fica aberta); consuma com EventSource no navegador ou um cliente HTTP que aceite streaming.
// @Tags        ao-vivo (admin)
// @Produce     text/event-stream
// @Param       id path string true "ID do evento"
// @Success     200 {object} liveAdminSnapshotDTO
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/fluxo [get]
func (a *API) handleAoVivoFluxoAdmin(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	a.streamLiveSSE(w, r, eventID, func(ctx context.Context) (any, error) {
		return a.buildAdminLiveSnapshot(ctx, eventID)
	})
}

// handleAoVivoEstadoApresentacao é o par autenticado de handleAoVivoEstado: mesmo
// snapshot somente-leitura (pergunta atual, pendentes, grupos revelados) que
// a plateia vê, mas restrito ao dono do evento — sem PIN nem token de
// visitante. Alimenta a janela "Modo apresentação", que só o organizador
// pode abrir.
// handleAoVivoEstadoApresentacao godoc
//
// @Summary     Snapshot do estado de apresentação, para o organizador (sem SSE)
// @Description Mesmo formato que a plateia vê (pergunta atual, pendentes, grupos revelados), mas restrito ao dono do evento — alimenta a janela "Modo apresentação".
// @Tags        ao-vivo (admin)
// @Produce     json
// @Param       id path string true "ID do evento"
// @Success     200 {object} liveSnapshotDTO
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/apresentacao/estado [get]
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
// handleAoVivoFluxoApresentacao godoc
//
// @Summary     Stream (SSE) do estado de apresentação, para o organizador
// @Description Server-Sent Events, mesmo formato de GET /ao-vivo/fluxo (público) mas autenticado — não requer PIN nem token de visitante. Não dá pra "experimentar" pelo Swagger UI (conexão fica aberta); consuma com EventSource.
// @Tags        ao-vivo (admin)
// @Produce     text/event-stream
// @Param       id path string true "ID do evento"
// @Success     200 {object} liveSnapshotDTO
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/eventos/{id}/ao-vivo/apresentacao/fluxo [get]
func (a *API) handleAoVivoFluxoApresentacao(w http.ResponseWriter, r *http.Request) {
	eventID, ok := a.resolveEventOwner(w, r)
	if !ok {
		return
	}
	a.streamLiveSSE(w, r, eventID, func(ctx context.Context) (any, error) {
		return a.buildLiveSnapshot(ctx, eventID)
	})
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
// handleAoVivoEntrar godoc
//
// @Summary     Entra na apresentação ao vivo com PIN (+ e-mail opcional)
// @Description Público, sem sessão. PIN errado nunca autoriza. E-mail de quem já respondeu vira "participant"; e-mail em branco ou de quem não respondeu vira "observer" (só assiste). O token devolvido é usado como "?token=" nas demais rotas /ao-vivo/* públicas.
// @Tags        ao-vivo (público)
// @Accept      json
// @Produce     json
// @Param       id   path string             true "ID do evento"
// @Param       body body liveJoinRequest    true "PIN e e-mail (opcional)"
// @Success     200 {object} liveEntrarResponse
// @Failure     400 {object} errorResponse
// @Failure     401 {object} errorResponse "PIN inválido"
// @Failure     404 {object} errorResponse
// @Router      /api/publico/eventos/{id}/ao-vivo/entrar [post]
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
// handleAoVivoReagir godoc
//
// @Summary     Manda uma reação de emoji (efêmera, não grava no banco)
// @Description emoji precisa ser um dos permitidos: 👍 ❤️ 😂 🎉 👏. Exige interações habilitadas no evento.
// @Tags        ao-vivo (público)
// @Accept      json
// @Produce     json
// @Param       id    path  string           true "ID do evento"
// @Param       token query string           true "Token de visitante"
// @Param       body  body  liveReactRequest true "emoji"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse
// @Failure     401 {object} errorResponse "sessão inválida"
// @Failure     403 {object} errorResponse "interações desativadas"
// @Security    liveViewerToken
// @Router      /api/publico/eventos/{id}/ao-vivo/reagir [post]
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
// handleAoVivoEnviarPergunta godoc
//
// @Summary     Envia uma pergunta/recado para a caixa de Q&A do organizador
// @Description Fica só na caixa privada do organizador — nunca aparece pra outros participantes. clientId (opcional) permite remover a própria mensagem depois. Exige interações habilitadas no evento.
// @Tags        ao-vivo (público)
// @Accept      json
// @Produce     json
// @Param       id    path  string             true "ID do evento"
// @Param       token query string             true "Token de visitante"
// @Param       body  body  liveSubmitQARequest true "Mensagem"
// @Success     200 {object} enviarPerguntaAoVivoResponse
// @Failure     400 {object} errorResponse
// @Failure     401 {object} errorResponse "sessão inválida"
// @Failure     403 {object} errorResponse "interações desativadas"
// @Security    liveViewerToken
// @Router      /api/publico/eventos/{id}/ao-vivo/perguntas [post]
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

// handleAoVivoEstado godoc
//
// @Summary     Snapshot público do estado ao vivo (sem SSE)
// @Description Equivalente sem streaming de GET /ao-vivo/fluxo (público).
// @Tags        ao-vivo (público)
// @Produce     json
// @Param       id    path  string true "ID do evento"
// @Param       token query string true "Token de visitante"
// @Success     200 {object} liveSnapshotDTO
// @Failure     401 {object} errorResponse "sessão inválida"
// @Security    liveViewerToken
// @Router      /api/publico/eventos/{id}/ao-vivo/estado [get]
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
// handleAoVivoFluxo godoc
//
// @Summary     Stream (SSE) do estado ao vivo, para participantes/observadores
// @Description Server-Sent Events: um frame liveSnapshotDTO ao conectar, de novo a cada mudança, e um heartbeat a cada 15s. Também emite eventos nomeados "reaction" (\{"emoji":"👍"\}) para reações ao vivo. Não dá pra "experimentar" pelo Swagger UI (conexão fica aberta); consuma com EventSource no navegador.
// @Tags        ao-vivo (público)
// @Produce     text/event-stream
// @Param       id    path  string true "ID do evento"
// @Param       token query string true "Token de visitante"
// @Success     200 {object} liveSnapshotDTO
// @Failure     401 {object} errorResponse "sessão inválida"
// @Security    liveViewerToken
// @Router      /api/publico/eventos/{id}/ao-vivo/fluxo [get]
func (a *API) handleAoVivoFluxo(w http.ResponseWriter, r *http.Request) {
	eventID, _, ok := a.resolveLiveViewer(w, r)
	if !ok {
		return
	}
	a.streamLiveSSE(w, r, eventID, func(ctx context.Context) (any, error) {
		return a.buildLiveSnapshot(ctx, eventID)
	})
}

// setSSEHeaders configura os cabeçalhos de um stream SSE. O X-Accel-Buffering
// desliga o buffering do proxy reverso (Traefik/nginx/Cloudflare): sem isso,
// frames pequenos podem ficar presos no buffer do proxy e chegar atrasados ou
// nunca — exatamente o que aparece em redes corporativas atrás de proxy.
func setSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
}

// streamLiveSSE mantém uma conexão SSE aberta pra um evento: envia o snapshot
// ao conectar, de novo a cada mudança de estado, e um heartbeat periódico que
// mantém a conexão viva (proxy/firewall corporativo derruba ou silencia
// conexão ociosa) e detecta cliente desconectado.
func (a *API) streamLiveSSE(w http.ResponseWriter, r *http.Request, eventID int64, build func(context.Context) (any, error)) {
	flusher, isFlusher := w.(http.Flusher)
	if !isFlusher {
		writeError(w, http.StatusInternalServerError, "Streaming não suportado.")
		return
	}
	setSSEHeaders(w)
	w.WriteHeader(http.StatusOK)

	writeSnapshot := func() bool {
		snapshot, err := build(r.Context())
		if err != nil {
			log.Printf("api: montar snapshot: %v", err)
			return false
		}
		data, err := json.Marshal(snapshot)
		if err != nil {
			log.Printf("api: serializar snapshot: %v", err)
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

	heartbeat := time.NewTicker(sseHeartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			// Reenvia o snapshot (não um comentário ": ping"): o EventSource
			// descarta comentários, então só um data frame deixa a tela saber
			// que o stream segue vivo durante uma pausa da apresentação — é o
			// que permite o cliente detectar "sem conexão" de verdade.
			if !writeSnapshot() {
				return
			}
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

	var currentOtherOptionID int64
	for _, opt := range current.Options {
		if opt.IsOther {
			currentOtherOptionID = opt.ID
			break
		}
	}

	revealed := state.Revealed[current.ID]
	pending := make([]liveParticipantDTO, 0, len(participants))
	revealedAnswers := make([]revealedLiveAnswer, 0, len(participants))
	// allOpenTexts/allOtherTexts só são preenchidos pra OPEN_TEXT e pra opção
	// "Outro" de uma GROUP, respectivamente: ao contrário das opções fixas
	// (que já existem de antemão e sempre aparecem, mesmo com 0 pessoas),
	// esses baldes só existem pelo que foi digitado — então pra eles
	// aparecerem na tela antes de qualquer revelação (igual às opções
	// fixas), é preciso saber o texto de todo mundo que respondeu, revelado
	// ou não. A identidade de quem ainda não foi revelado nunca é anexada a
	// esse texto no payload — só usamos aqui pra descobrir quais baldes
	// existem.
	var allOpenTexts []string
	var allOtherTexts []string
	// Quem não está no mapa não respondeu a esta pergunta (ex.: pergunta criada
	// depois do prazo) e fica fora do snapshot — não aparece como pendente nem
	// em nenhum grupo, porque não tem opção/resposta pra encaixar.
	answersByParticipant, err := a.store.ListarRespostasPorPergunta(ctx, current.ID)
	if err != nil {
		return liveSnapshotDTO{}, fmt.Errorf("listar respostas da pergunta: %w", err)
	}
	for _, p := range participants {
		found, ok := answersByParticipant[p.ID]
		if !ok {
			continue
		}
		if !revealed[p.ID] {
			pending = append(pending, a.toLiveParticipantDTO(p))
		} else {
			revealedAnswers = append(revealedAnswers, revealedLiveAnswer{participant: p, optionID: found.OptionID, text: found.FreeText})
		}
		if current.Type == questionTypeOpenText {
			allOpenTexts = append(allOpenTexts, found.FreeText)
		} else if currentOtherOptionID != 0 && found.OptionID == currentOtherOptionID {
			allOtherTexts = append(allOtherTexts, found.FreeText)
		}
	}

	snapshot.Pending = pending
	snapshot.Groups = a.buildLiveGroups(current, revealedAnswers, allOpenTexts, allOtherTexts)
	return snapshot, nil
}

// buildTextGroups agrupa por texto normalizado (trim + primeira letra
// maiúscula), um balde por resposta distinta — usado tanto por perguntas
// OPEN_TEXT quanto pela opção "Outro" dentro de uma pergunta GROUP (nos dois
// casos o balde não existe de antemão, só depois de alguém escrever algo).
// allTexts cobre quem ainda não foi revelado, pra o balde já aparecer na
// tela antes da revelação (mesma lógica de allOpenTexts em buildLiveSnapshot).
func (a *API) buildTextGroups(allTexts []string, revealed []revealedLiveAnswer) []liveGroupDTO {
	order := make([]string, 0, len(allTexts))
	byLabel := make(map[string]*liveGroupDTO, len(allTexts))
	for _, text := range allTexts {
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
			// não deveria acontecer (allTexts inclui os revelados também),
			// mas cria o balde se faltar por algum motivo.
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

// buildLiveGroups agrupa as respostas reveladas em "baldes" pro telão. Pra
// GROUP, cada opção fixa já existe de antemão (aparece com 0 pessoas até
// alguém ser revelado nela) — exceto a opção "Outro", que não é uma escolha
// compartilhável: cada pessoa escreveu a própria resposta, então ela usa o
// mesmo agrupamento por texto de OPEN_TEXT (allOtherTexts), em vez de um
// único balde genérico "Outro" misturando respostas diferentes.
func (a *API) buildLiveGroups(q *store.QuestionWithOptions, revealed []revealedLiveAnswer, allOpenTexts, allOtherTexts []string) []liveGroupDTO {
	if q.Type == questionTypeOpenText {
		return a.buildTextGroups(allOpenTexts, revealed)
	}

	var otherOptionID int64
	for _, opt := range q.Options {
		if opt.IsOther {
			otherOptionID = opt.ID
			break
		}
	}

	order := make([]int64, 0, len(q.Options))
	byOption := make(map[int64]*liveGroupDTO, len(q.Options))
	for _, opt := range q.Options {
		if opt.ID == otherOptionID {
			continue
		}
		byOption[opt.ID] = &liveGroupDTO{Label: opt.TextLabel, Participants: []liveParticipantDTO{}}
		order = append(order, opt.ID)
	}

	otherRevealed := make([]revealedLiveAnswer, 0)
	for _, r := range revealed {
		if otherOptionID != 0 && r.optionID == otherOptionID {
			otherRevealed = append(otherRevealed, r)
			continue
		}
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
	if otherOptionID != 0 {
		groups = append(groups, a.buildTextGroups(allOtherTexts, otherRevealed)...)
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
