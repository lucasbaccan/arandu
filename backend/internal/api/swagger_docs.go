package api

// Este arquivo só existe para a documentação Swagger (swag). Os handlers
// respondem com map[string]any montado na hora (ver writeJSON) em vez de um
// tipo nomeado — o que é ótimo pro código, mas o swag precisa de um tipo real
// pra descrever o formato de cada resposta numa anotação "@Success". Os tipos
// abaixo espelham exatamente esses mapas; nenhum é construído em tempo de
// execução.

type okResponse struct {
	OK bool `json:"ok"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type userEnvelope struct {
	User userDTO `json:"user"`
}

type contaConfiguracaoResponse struct {
	MinPasswordLength   int  `json:"minPasswordLength"`
	RegistrationEnabled bool `json:"registrationEnabled"`
	EmailEnabled        bool `json:"emailEnabled"`
}

type validTokenResponse struct {
	Valid bool `json:"valid"`
}

type adminUsersResponse struct {
	Users []adminUserDTO `json:"users"`
}

type adminResetLinkResponse struct {
	Token string `json:"token"`
	// ExpiresAt no formato RFC3339.
	ExpiresAt string `json:"expiresAt"`
}

type adminEventsResponse struct {
	Events []adminEventDTO `json:"events"`
}

type adminConfigResponse struct {
	RegistrationEnabled bool `json:"registrationEnabled"`
	EmailEnabled        bool `json:"emailEnabled"`
}

type eventEnvelope struct {
	Event eventDTO `json:"event"`
}

type eventsEnvelope struct {
	Events []eventDTO `json:"events"`
}

type questionEnvelope struct {
	Question questionDTO `json:"question"`
}

type questionsEnvelope struct {
	Questions []questionDTO `json:"questions"`
}

type respostasEnvelope struct {
	ParticipantCount int                      `json:"participantCount"`
	Participants     []participantResponseDTO `json:"participants"`
}

type enviarRespostasResponse struct {
	OK bool `json:"ok"`
	// EditToken permite voltar depois e editar as mesmas respostas (link
	// pessoal, sem login) — só presente quando o evento permite edição.
	EditToken string `json:"editToken"`
}

type pinResolvidoResponse struct {
	ID string `json:"id"`
}

type nomeExisteResponse struct {
	Exists bool `json:"exists"`
}

type publicoEventoResponse struct {
	Event     publicEventDTO      `json:"event"`
	Questions []publicQuestionDTO `json:"questions"`
}

type publicoParticipanteEnvelope struct {
	Participant publicParticipantDTO `json:"participant"`
}

type liveEntrarResponse struct {
	// Token de visitante (JWT) — usado como "?token=" nas rotas
	// /api/publico/eventos/{id}/ao-vivo/*.
	Token string `json:"token"`
	// Role: "participant" (já respondeu, com e-mail reconhecido) ou
	// "observer" (só assiste).
	Role string `json:"role"`
}

// removerPerguntaAoVivoRequest espelha o struct anônimo usado em
// handleAoVivoRemoverPergunta (o handler não dá nome a ele, mas o swag
// precisa de um tipo nomeado pra documentar o corpo da requisição).
type removerPerguntaAoVivoRequest struct {
	ClientID string `json:"clientId"`
}

type enviarPerguntaAoVivoResponse struct {
	OK        bool   `json:"ok"`
	MessageID string `json:"messageId"`
	Text      string `json:"text"`
}
