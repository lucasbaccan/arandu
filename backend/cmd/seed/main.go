// Comando seed popula o banco com dados fake pra explorar o app sem precisar
// criar tudo manualmente: um usuário demo, 10 eventos, cada um com 10
// perguntas (5 abertas, 5 de múltipla escolha) e um bom punhado de
// participantes (com e sem foto) já respondendo. Roda direto contra o
// mesmo store/config usados pelo servidor (respeitando DATABASE_PATH), então
// os dados aparecem no mesmo banco que `make dev`/`make run` usam.
//
// Uso: go run ./cmd/seed   (ou `make seed`)
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"devopsconecta/backend/internal/auth"
	"devopsconecta/backend/internal/config"
	"devopsconecta/backend/internal/ids"
	"devopsconecta/backend/internal/store"
)

const (
	demoEmail    = "demo@demo.com"
	demoPassword = "demo"
	demoName     = "Demo"

	numEvents               = 10
	minParticipantsPerEvent = 15
	maxParticipantsPerEvent = 22

	// maxTitleLen espelha maxQuestionTitleLength de internal/api/questions.go.
	maxTitleLen = 300

	roleOpen15    = "open15"
	roleOpenGroup = "opengroup"
)

func main() {
	dbPath := flag.String("db", "", "caminho do banco SQLite (padrão: DATABASE_PATH ou ./data/app.db)")
	flag.Parse()

	cfg := config.Load()
	if *dbPath != "" {
		cfg.DatabasePath = *dbPath
	}

	db, err := store.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("seed: %v", err)
	}
	defer db.Close()
	if err := store.Migrate(db); err != nil {
		log.Fatalf("seed: %v", err)
	}

	st := store.New(db)
	gen := ids.NewGenerator(int64(cfg.SnowflakeNode))
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	ctx := context.Background()

	log.Printf("Banco: %s", cfg.DatabasePath)

	user, err := ensureDemoUser(ctx, db, st, gen)
	if err != nil {
		log.Fatalf("seed: usuário demo: %v", err)
	}
	log.Printf("Usuário demo pronto (id %d): %s / senha %q", user.ID, user.Email, demoPassword)

	var totalQuestions, totalParticipants, totalAnswers int
	for i := 0; i < numEvents; i++ {
		nq, np, na, err := seedEvent(ctx, st, gen, rng, user.ID, i)
		if err != nil {
			log.Fatalf("seed: evento %d: %v", i+1, err)
		}
		totalQuestions += nq
		totalParticipants += np
		totalAnswers += na
	}

	log.Printf("Pronto! %d eventos, %d perguntas, %d participantes, %d respostas.", numEvents, totalQuestions, totalParticipants, totalAnswers)
	log.Printf("Login: %s / %s", demoEmail, demoPassword)
}

// ensureDemoUser cria o usuário demo na primeira vez; em reexecuções,
// reaproveita o usuário existente e apaga os eventos antigos dele (cascade
// cuida de perguntas/participantes/respostas), pra `make seed` ser
// idempotente e não empilhar dados a cada chamada.
func ensureDemoUser(ctx context.Context, db *sql.DB, st *store.Store, gen *ids.Generator) (store.User, error) {
	existing, err := st.FindUserByEmail(ctx, demoEmail)
	if err == nil {
		if _, err := db.ExecContext(ctx, `DELETE FROM events WHERE owner_id = ?`, existing.ID); err != nil {
			return store.User{}, fmt.Errorf("limpar eventos antigos do usuário demo: %w", err)
		}
		return existing, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return store.User{}, err
	}

	hash, err := auth.HashPassword(demoPassword)
	if err != nil {
		return store.User{}, err
	}
	return st.CreateUser(ctx, store.User{
		ID:           gen.NextID(),
		Email:        demoEmail,
		Name:         demoName,
		PasswordHash: hash,
		AuthProvider: "email",
	})
}

func seedEvent(ctx context.Context, st *store.Store, gen *ids.Generator, rng *rand.Rand, ownerID int64, idx int) (numQuestions, numParticipants, numAnswers int, err error) {
	event, err := st.CreateEvent(ctx, store.Event{
		ID:                  gen.NextID(),
		OwnerID:             ownerID,
		Title:               eventTitles[idx%len(eventTitles)],
		PINCode:             fmt.Sprintf("DEMO%02d", idx+1),
		Status:              eventStatuses[idx%len(eventStatuses)],
		ShowRanking:         rng.Intn(2) == 0,
		InteractionsEnabled: true,
	})
	if err != nil {
		return 0, 0, 0, fmt.Errorf("criar evento: %w", err)
	}

	questions, err := seedQuestions(ctx, st, gen, rng, event.ID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("criar perguntas: %w", err)
	}

	participants, err := seedParticipants(ctx, st, gen, rng, event.ID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("criar participantes: %w", err)
	}

	numAnswers, err = seedAnswers(ctx, st, gen, rng, questions, participants)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("criar respostas: %w", err)
	}

	log.Printf("  Evento %q (PIN %s, %s): %d perguntas, %d participantes, %d respostas",
		event.Title, event.PINCode, event.Status, len(questions), len(participants), numAnswers)
	return len(questions), len(participants), numAnswers, nil
}

type seededQuestion struct {
	store.QuestionWithOptions
	Role string
}

// seedQuestions cria as 10 perguntas do evento: 5 OPEN_TEXT (incluindo a
// pergunta com 15 respostas distintas e a de agrupamento por capitalização)
// e 5 GROUP (múltipla escolha), com exatamente uma pergunta bem longa de
// cada tipo.
func seedQuestions(ctx context.Context, st *store.Store, gen *ids.Generator, rng *rand.Rand, eventID int64) ([]seededQuestion, error) {
	generic := pickStrings(rng, genericOpenQuestions, 3)
	generic[rng.Intn(len(generic))] = truncateTitle(longOpenTitle)

	openDefs := []struct{ title, role string }{
		{open15Question, roleOpen15},
		{openGroupQuestion, roleOpenGroup},
		{generic[0], ""},
		{generic[1], ""},
		{generic[2], ""},
	}
	rng.Shuffle(len(openDefs), func(i, j int) { openDefs[i], openDefs[j] = openDefs[j], openDefs[i] })

	result := make([]seededQuestion, 0, 10)
	for _, def := range openDefs {
		q, err := st.CreateQuestion(ctx, store.Question{
			ID:         gen.NextID(),
			EventID:    eventID,
			Title:      def.title,
			Type:       "OPEN_TEXT",
			LayoutView: "TIMELINE",
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("criar pergunta aberta: %w", err)
		}
		result = append(result, seededQuestion{QuestionWithOptions: q, Role: def.role})
	}

	groupDefs := pickGroupQuestions(rng, 5)
	groupDefs[rng.Intn(len(groupDefs))].Title = truncateTitle(longGroupTitle)
	for _, gq := range groupDefs {
		opts := make([]store.QuestionOption, 0, len(gq.Options))
		for _, label := range gq.Options {
			opts = append(opts, store.QuestionOption{ID: gen.NextID(), TextLabel: label})
		}
		q, err := st.CreateQuestion(ctx, store.Question{
			ID:         gen.NextID(),
			EventID:    eventID,
			Title:      gq.Title,
			Type:       "GROUP",
			LayoutView: "TIMELINE",
		}, opts)
		if err != nil {
			return nil, fmt.Errorf("criar pergunta de múltipla escolha: %w", err)
		}
		result = append(result, seededQuestion{QuestionWithOptions: q})
	}

	return result, nil
}

// seedParticipants cria entre 15 e 22 participantes (nomes/e-mails únicos
// tirados do pool), metade com foto (avatar gerado) e metade sem.
func seedParticipants(ctx context.Context, st *store.Store, gen *ids.Generator, rng *rand.Rand, eventID int64) ([]store.Participant, error) {
	n := minParticipantsPerEvent + rng.Intn(maxParticipantsPerEvent-minParticipantsPerEvent+1)
	if n > len(namePool) {
		n = len(namePool)
	}
	order := rng.Perm(len(namePool))

	participants := make([]store.Participant, 0, n)
	for i := 0; i < n; i++ {
		ps := namePool[order[i]]
		email := ps.Local + "@exemplo.com"
		photo := ""
		if i%2 == 0 {
			photo = avatarDataURL(email)
		}
		created, err := st.UpsertParticipant(ctx, store.Participant{
			ID:      gen.NextID(),
			EventID: eventID,
			Email:   email,
			Name:    ps.Name,
			Photo:   photo,
		})
		if err != nil {
			return nil, fmt.Errorf("criar participante %s: %w", email, err)
		}
		participants = append(participants, created)
	}
	return participants, nil
}

func seedAnswers(ctx context.Context, st *store.Store, gen *ids.Generator, rng *rand.Rand, questions []seededQuestion, participants []store.Participant) (int, error) {
	total := 0
	for pi, p := range participants {
		answers := make([]store.Answer, 0, len(questions))
		for _, q := range questions {
			a := store.Answer{ID: gen.NextID(), QuestionID: q.ID}
			if q.Type == "GROUP" {
				a.OptionID = q.Options[rng.Intn(len(q.Options))].ID
			} else {
				a.FreeText = openAnswerFor(rng, q.Role, pi)
			}
			answers = append(answers, a)
		}
		if err := st.ReplaceAnswers(ctx, p.ID, answers); err != nil {
			return total, fmt.Errorf("gravar respostas do participante %d: %w", p.ID, err)
		}
		total += len(answers)
	}
	return total, nil
}

// openAnswerFor decide o texto livre de acordo com o papel da pergunta:
//   - roleOpen15: um valor do pool de 15 textos distintos, um por
//     participante (repete só se houver mais participantes que o pool).
//   - roleOpenGroup: uma palavra do pool de linguagens, com a primeira letra
//     ora maiúscula ora minúscula — os dois primeiros participantes são
//     forçados pra "python"/"Python", garantindo pelo menos um grupo com
//     duplicata de capitalização diferente independente do sorteio.
//   - senão: uma resposta genérica qualquer do pool.
func openAnswerFor(rng *rand.Rand, role string, participantIdx int) string {
	switch role {
	case roleOpen15:
		return open15Pool[participantIdx%len(open15Pool)]
	case roleOpenGroup:
		if participantIdx == 0 {
			return "python"
		}
		if participantIdx == 1 {
			return "Python"
		}
		word := groupWordPool[rng.Intn(len(groupWordPool))]
		if rng.Intn(2) == 0 {
			return strings.ToUpper(word[:1]) + word[1:]
		}
		return word
	default:
		return genericAnswerPool[rng.Intn(len(genericAnswerPool))]
	}
}

func pickStrings(rng *rand.Rand, pool []string, n int) []string {
	order := rng.Perm(len(pool))
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = pool[order[i]]
	}
	return out
}

func pickGroupQuestions(rng *rand.Rand, n int) []groupQuestion {
	order := rng.Perm(len(groupQuestionPool))
	out := make([]groupQuestion, n)
	for i := 0; i < n; i++ {
		out[i] = groupQuestionPool[order[i]]
	}
	return out
}

func truncateTitle(s string) string {
	r := []rune(s)
	if len(r) <= maxTitleLen {
		return s
	}
	return string(r[:maxTitleLen])
}
