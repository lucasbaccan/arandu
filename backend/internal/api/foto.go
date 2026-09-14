package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"devopsconecta/backend/internal/photos"
	"devopsconecta/backend/internal/store"
)

// photoURL devolve a URL pública da foto de um participante ("" quando ele
// não tem foto). As respostas JSON nunca mais carregam o base64 — a foto é
// servida como arquivo estático por GET /api/fotos/{id}.
func (a *API) photoURL(p store.Participant) string {
	if p.Photo == "" {
		return ""
	}
	return a.photos.URL(p.ID)
}

// handleFotoParticipante serve a foto de um participante como arquivo. Na
// primeira vez o arquivo é materializado em disco a partir do data URL salvo
// no banco (mais lento); nas seguintes, é servido direto do disco (rápido).
// A foto não exige autenticação: ela já é exibida publicamente na tela da
// apresentação.
// handleFotoParticipante godoc
//
// @Summary     Serve a foto de um participante
// @Description Público (a foto já aparece na apresentação). Primeira chamada materializa o arquivo a partir do banco; as seguintes servem direto do disco com cache.
// @Tags        fotos
// @Produce     png
// @Produce     jpeg
// @Param       participantId path string true "ID do participante"
// @Success     200 {file} file
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse "sem foto"
// @Router      /api/fotos/{participantId} [get]
func (a *API) handleFotoParticipante(w http.ResponseWriter, r *http.Request) {
	participantID, err := strconv.ParseInt(r.PathValue("participantId"), 10, 64)
	if err != nil || participantID <= 0 {
		writeError(w, http.StatusBadRequest, "ID de participante inválido.")
		return
	}

	err = a.photos.Serve(w, r, participantID, func() (string, error) {
		p, err := a.store.BuscarParticipantePorID(r.Context(), participantID)
		if err != nil {
			return "", err
		}
		return p.Photo, nil
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) || errors.Is(err, photos.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Foto não encontrada.")
			return
		}
		log.Printf("api: servir foto do participante %d: %v", participantID, err)
		writeError(w, http.StatusInternalServerError, "Erro interno ao servir a foto.")
	}
}
