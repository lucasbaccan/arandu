package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"devopsconecta/backend/internal/auth"
	"devopsconecta/backend/internal/store"
)

// passwordResetTokenTTL é a validade do link de redefinição de senha gerado
// pelo super admin — curto de propósito: é um link pra repassar na hora
// (chat, e-mail), não uma credencial de longa duração.
const passwordResetTokenTTL = time.Hour

type adminUserDTO struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	CreatedAt  string `json:"createdAt"`
	EventCount int    `json:"eventCount"`
}

func toAdminUserDTO(us store.UserSummary) adminUserDTO {
	return adminUserDTO{
		ID:         strconv.FormatInt(us.ID, 10),
		Email:      us.Email,
		Name:       us.Name,
		Role:       us.Role,
		CreatedAt:  us.CreatedAt.Format(time.RFC3339),
		EventCount: us.EventCount,
	}
}

// handleAdminListarUsuarios godoc
//
// @Summary     Lista todos os usuários
// @Description Somente super admin — cada usuário vem com sua contagem de eventos.
// @Tags        admin
// @Produce     json
// @Success     200 {object} adminUsersResponse
// @Failure     403 {object} errorResponse "acesso restrito ao super admin"
// @Security    cookieAuth
// @Router      /api/admin/usuarios [get]
func (a *API) handleAdminListarUsuarios(w http.ResponseWriter, r *http.Request) {
	users, err := a.store.ListarUsuariosComContagemDeEventos(r.Context())
	if err != nil {
		log.Printf("api: listar usuários: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao listar usuários.")
		return
	}
	dtos := make([]adminUserDTO, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, toAdminUserDTO(u))
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": dtos})
}

func parseUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID de usuário inválido.")
		return 0, false
	}
	return id, true
}

// handleAdminExcluirUsuario apaga a conta (e, junto, todos os eventos dela).
// Quem está agindo não pode se auto-excluir por aqui: essa tela é para
// gerenciar outras contas, e um super admin sem outra forma de virar admin de
// novo (o papel só é atribuído implicitamente ao primeiro cadastro do banco)
// ficaria travado fora da própria administração.
// handleAdminExcluirUsuario godoc
//
// @Summary     Exclui um usuário (e seus eventos)
// @Description Somente super admin. Quem está agindo não pode se auto-excluir por aqui.
// @Tags        admin
// @Produce     json
// @Param       id path string true "ID do usuário"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse "auto-exclusão bloqueada"
// @Failure     403 {object} errorResponse "acesso restrito ao super admin"
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/admin/usuarios/{id} [delete]
func (a *API) handleAdminExcluirUsuario(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUserID(w, r)
	if !ok {
		return
	}
	if id == userIDFromContext(r.Context()) {
		writeError(w, http.StatusBadRequest, "Você não pode excluir a própria conta por aqui.")
		return
	}
	if err := a.store.ExcluirUsuario(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Usuário não encontrado.")
			return
		}
		log.Printf("api: excluir usuário: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao excluir o usuário.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// handleAdminGerarLinkRedefinicao gera um token de uso único (o super admin
// nunca escolhe nem vê a senha da pessoa) para ela mesma definir uma nova
// senha em /redefinir-senha. O admin copia o link retornado e repassa por
// fora (chat, e-mail) — o app não envia nada.
// handleAdminGerarLinkRedefinicao godoc
//
// @Summary     Gera um link de redefinição de senha para um usuário
// @Description Somente super admin. Token de uso único, válido por 1h — o admin copia e envia por fora (o app não envia nada); gerar de novo invalida o anterior.
// @Tags        admin
// @Produce     json
// @Param       id path string true "ID do usuário"
// @Success     200 {object} adminResetLinkResponse
// @Failure     403 {object} errorResponse "acesso restrito ao super admin"
// @Failure     404 {object} errorResponse
// @Security    cookieAuth
// @Router      /api/admin/usuarios/{id}/link-redefinicao [post]
func (a *API) handleAdminGerarLinkRedefinicao(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUserID(w, r)
	if !ok {
		return
	}
	if _, err := a.store.BuscarUsuarioPorID(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Usuário não encontrado.")
			return
		}
		log.Printf("api: buscar usuário: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}

	token, err := generateResetToken()
	if err != nil {
		log.Printf("api: gerar token de redefinição: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao gerar o link.")
		return
	}
	expiresAt := time.Now().Add(passwordResetTokenTTL)
	if err := a.store.CriarTokenRedefinicaoSenha(r.Context(), token, id, expiresAt); err != nil {
		log.Printf("api: salvar token de redefinição: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao gerar o link.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token":     token,
		"expiresAt": expiresAt.Format(time.RFC3339),
	})
}

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type adminEventDTO struct {
	eventDTO
	OwnerName  string `json:"ownerName"`
	OwnerEmail string `json:"ownerEmail"`
}

// handleAdminListarEventos godoc
//
// @Summary     Lista todos os eventos, de todos os organizadores
// @Description Somente super admin.
// @Tags        admin
// @Produce     json
// @Success     200 {object} adminEventsResponse
// @Failure     403 {object} errorResponse "acesso restrito ao super admin"
// @Security    cookieAuth
// @Router      /api/admin/eventos [get]
func (a *API) handleAdminListarEventos(w http.ResponseWriter, r *http.Request) {
	events, err := a.store.ListarTodosResumosDeEventos(r.Context())
	if err != nil {
		log.Printf("api: listar todos os eventos: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao listar eventos.")
		return
	}
	dtos := make([]adminEventDTO, 0, len(events))
	for _, e := range events {
		dtos = append(dtos, adminEventDTO{
			eventDTO:   toEventSummaryDTO(e.EventSummary),
			OwnerName:  e.OwnerName,
			OwnerEmail: e.OwnerEmail,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": dtos})
}

// handleAdminBuscarConfiguracoes godoc
//
// @Summary     Configurações administrativas
// @Description Somente super admin. Hoje só tem o toggle de cadastro aberto.
// @Tags        admin
// @Produce     json
// @Success     200 {object} adminConfigResponse
// @Failure     403 {object} errorResponse "acesso restrito ao super admin"
// @Security    cookieAuth
// @Router      /api/admin/configuracoes [get]
func (a *API) handleAdminBuscarConfiguracoes(w http.ResponseWriter, r *http.Request) {
	enabled, err := a.store.RegistroHabilitado(r.Context())
	if err != nil {
		log.Printf("api: buscar configurações: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao buscar configurações.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"registrationEnabled": enabled})
}

type updateAdminConfigRequest struct {
	RegistrationEnabled bool `json:"registrationEnabled"`
}

// handleAdminAtualizarConfiguracoes godoc
//
// @Summary     Atualiza configurações administrativas
// @Description Somente super admin. Liga/desliga o cadastro de novas contas organizadoras.
// @Tags        admin
// @Accept      json
// @Produce     json
// @Param       body body updateAdminConfigRequest true "Nova configuração"
// @Success     200 {object} adminConfigResponse
// @Failure     400 {object} errorResponse
// @Failure     403 {object} errorResponse "acesso restrito ao super admin"
// @Security    cookieAuth
// @Router      /api/admin/configuracoes [patch]
func (a *API) handleAdminAtualizarConfiguracoes(w http.ResponseWriter, r *http.Request) {
	var req updateAdminConfigRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.DefinirRegistroHabilitado(r.Context(), req.RegistrationEnabled); err != nil {
		log.Printf("api: atualizar configurações: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao salvar configurações.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"registrationEnabled": req.RegistrationEnabled})
}

// --- Redefinição de senha via link (público, sem sessão) ---

// handlePublicoValidarTokenRedefinicao godoc
//
// @Summary     Valida um link de redefinição de senha
// @Description Público, sem sessão. Usado por /redefinir-senha antes de mostrar o formulário de nova senha.
// @Tags        conta
// @Produce     json
// @Param       token query string true "Token do link gerado pelo super admin"
// @Success     200 {object} validTokenResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse "link inválido ou expirado"
// @Router      /api/publico/redefinir-senha [get]
func (a *API) handlePublicoValidarTokenRedefinicao(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "Link inválido.")
		return
	}
	rt, err := a.store.BuscarTokenRedefinicaoSenha(r.Context(), token)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Link inválido ou expirado.")
		return
	}
	if err != nil {
		log.Printf("api: validar token de redefinição: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	if time.Now().After(rt.ExpiresAt) {
		writeError(w, http.StatusNotFound, "Link inválido ou expirado.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"valid": true})
}

type redefinirSenhaRequest struct {
	Token     string `json:"token"`
	NovaSenha string `json:"novaSenha"`
}

// handlePublicoRedefinirSenha godoc
//
// @Summary     Define uma nova senha a partir do link de redefinição
// @Description Público, sem sessão. O token é de uso único e expira em 1h.
// @Tags        conta
// @Accept      json
// @Produce     json
// @Param       body body redefinirSenhaRequest true "Token e nova senha"
// @Success     200 {object} okResponse
// @Failure     400 {object} errorResponse "link inválido/expirado ou senha fora das regras"
// @Router      /api/publico/redefinir-senha [post]
func (a *API) handlePublicoRedefinirSenha(w http.ResponseWriter, r *http.Request) {
	var req redefinirSenhaRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	rt, err := a.store.BuscarTokenRedefinicaoSenha(r.Context(), req.Token)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusBadRequest, "Link inválido ou expirado.")
		return
	}
	if err != nil {
		log.Printf("api: buscar token de redefinição: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	if time.Now().After(rt.ExpiresAt) {
		writeError(w, http.StatusBadRequest, "Link inválido ou expirado.")
		return
	}

	if len(req.NovaSenha) < a.cfg.MinPasswordLength {
		writeError(w, http.StatusBadRequest, "A nova senha deve ter pelo menos "+strconv.Itoa(a.cfg.MinPasswordLength)+" caracteres.")
		return
	}
	if len(req.NovaSenha) > maxPasswordLength {
		writeError(w, http.StatusBadRequest, "A nova senha é muito longa (máximo "+strconv.Itoa(maxPasswordLength)+" caracteres).")
		return
	}

	hash, err := auth.HashPassword(req.NovaSenha)
	if err != nil {
		log.Printf("api: redefinir senha: hash: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno.")
		return
	}
	if err := a.store.AtualizarSenha(r.Context(), rt.UserID, hash); err != nil {
		log.Printf("api: redefinir senha: salvar: %v", err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao salvar a nova senha.")
		return
	}
	_ = a.store.ExcluirTokenRedefinicaoSenha(r.Context(), req.Token)

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
