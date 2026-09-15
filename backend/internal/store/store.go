package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

var (
	ErrNotFound   = errors.New("store: registro não encontrado")
	ErrEmailTaken = errors.New("store: e-mail já cadastrado")
)

// Open abre o SQLite com WAL e uma única conexão (single writer).
func Open(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("store: criar diretório: %w", err)
		}
	}

	dsn := path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: abrir banco: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("store: ping: %w", err)
	}
	return db, nil
}

func Migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id            INTEGER PRIMARY KEY,
			email         TEXT    NOT NULL UNIQUE,
			name          TEXT    NOT NULL,
			password_hash TEXT    NOT NULL,
			auth_provider TEXT    NOT NULL DEFAULT 'email',
			created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS events (
			id                  INTEGER PRIMARY KEY,
			owner_id            INTEGER NOT NULL REFERENCES users(id),
			title               TEXT    NOT NULL,
			pin_code            TEXT    NOT NULL,
			status              TEXT    NOT NULL DEFAULT 'PREPARATION',
			config_show_ranking BOOLEAN NOT NULL DEFAULT 0,
			allow_edit          BOOLEAN NOT NULL DEFAULT 1,
			created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE UNIQUE INDEX IF NOT EXISTS idx_events_pin_active
			ON events(pin_code) WHERE status != 'FINISHED';

		CREATE TABLE IF NOT EXISTS questions (
			id          INTEGER PRIMARY KEY,
			event_id    INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
			title       TEXT    NOT NULL,
			type        TEXT    NOT NULL,
			order_index INTEGER NOT NULL DEFAULT 0,
			created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS question_options (
			id          INTEGER PRIMARY KEY,
			question_id INTEGER NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
			text_label  TEXT    NOT NULL,
			is_other    BOOLEAN NOT NULL DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_questions_event ON questions(event_id);
		CREATE INDEX IF NOT EXISTS idx_question_options_question ON question_options(question_id);

		CREATE TABLE IF NOT EXISTS participants (
			id         INTEGER PRIMARY KEY,
			event_id   INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
			email      TEXT    NOT NULL,
			name       TEXT    NOT NULL DEFAULT '',
			photo      TEXT    NOT NULL DEFAULT '',
			edit_token TEXT    NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(event_id, email)
		);

		CREATE TABLE IF NOT EXISTS answers (
			id             INTEGER PRIMARY KEY,
			question_id    INTEGER NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
			participant_id INTEGER NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
			option_id      INTEGER REFERENCES question_options(id) ON DELETE CASCADE,
			free_text      TEXT    NOT NULL DEFAULT '',
			created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(question_id, participant_id)
		);

		CREATE INDEX IF NOT EXISTS idx_participants_event ON participants(event_id);
		CREATE INDEX IF NOT EXISTS idx_answers_question ON answers(question_id);
		CREATE INDEX IF NOT EXISTS idx_answers_participant ON answers(participant_id);

		CREATE TABLE IF NOT EXISTS live_qa_messages (
			id             INTEGER PRIMARY KEY,
			event_id       INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
			participant_id INTEGER REFERENCES participants(id) ON DELETE CASCADE,
			text           TEXT    NOT NULL,
			dismissed      BOOLEAN NOT NULL DEFAULT 0,
			created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_live_qa_messages_event ON live_qa_messages(event_id);

		CREATE TABLE IF NOT EXISTS event_live_state (
			event_id              INTEGER PRIMARY KEY REFERENCES events(id) ON DELETE CASCADE,
			current_question_id   INTEGER NOT NULL DEFAULT 0,
			blanked               BOOLEAN NOT NULL DEFAULT 0,
			message               TEXT    NOT NULL DEFAULT '',
			answers_hidden        BOOLEAN NOT NULL DEFAULT 0,
			names_hidden          BOOLEAN NOT NULL DEFAULT 0,
			present_density_mode  TEXT    NOT NULL DEFAULT ''
		);

		CREATE TABLE IF NOT EXISTS revealed_answers (
			event_id       INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
			question_id    INTEGER NOT NULL,
			participant_id INTEGER NOT NULL,
			PRIMARY KEY (event_id, question_id, participant_id)
		);

		CREATE INDEX IF NOT EXISTS idx_revealed_answers_event ON revealed_answers(event_id);
	`)
	if err != nil {
		return fmt.Errorf("store: migração: %w", err)
	}

	// Bancos criados antes da coluna edit_token existir: adiciona sem quebrar
	// dados existentes (CREATE TABLE IF NOT EXISTS acima não altera tabelas já
	// criadas).
	if _, err := db.Exec(`ALTER TABLE participants ADD COLUMN edit_token TEXT NOT NULL DEFAULT ''`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("store: migração (edit_token): %w", err)
		}
	}

	if _, err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_participants_edit_token
			ON participants(edit_token) WHERE edit_token != '';
	`); err != nil {
		return fmt.Errorf("store: migração (índice edit_token): %w", err)
	}

	// Bancos criados antes de interactions_enabled existir: adiciona sem
	// quebrar dados existentes, default true (reações/Q&A ligados por padrão).
	if _, err := db.Exec(`ALTER TABLE events ADD COLUMN interactions_enabled BOOLEAN NOT NULL DEFAULT 1`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("store: migração (interactions_enabled): %w", err)
		}
	}

	// Bancos criados antes da coluna name existir: adiciona sem quebrar dados
	// existentes (CREATE TABLE IF NOT EXISTS acima não altera tabelas já
	// criadas). Participantes antigos ficam com name vazio até reenviarem.
	if _, err := db.Exec(`ALTER TABLE participants ADD COLUMN name TEXT NOT NULL DEFAULT ''`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("store: migração (participants.name): %w", err)
		}
	}

	// Bancos criados antes de allow_edit existir: adiciona sem quebrar dados
	// existentes, default true (link de edição ativo por padrão).
	if _, err := db.Exec(`ALTER TABLE events ADD COLUMN allow_edit BOOLEAN NOT NULL DEFAULT 1`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("store: migração (allow_edit): %w", err)
		}
	}

	// Bancos criados antes de present_density_mode existir: adiciona sem
	// quebrar dados existentes, default '' (automático).
	if _, err := db.Exec(`ALTER TABLE event_live_state ADD COLUMN present_density_mode TEXT NOT NULL DEFAULT ''`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("store: migração (present_density_mode): %w", err)
		}
	}

	// Os modos de densidade mudaram de nome: 'smart' virou 'modo_amplo' e
	// 'auto' virou 'modo_compacto'. Atualiza valores antigos persistidos pra
	// não cair no comportamento errado (idempotente: roda de novo sem efeito).
	if _, err := db.Exec(`UPDATE event_live_state SET present_density_mode = 'modo_amplo' WHERE present_density_mode = 'smart'`); err != nil {
		return fmt.Errorf("store: migração (modo_amplo): %w", err)
	}
	if _, err := db.Exec(`UPDATE event_live_state SET present_density_mode = 'modo_compacto' WHERE present_density_mode = 'auto'`); err != nil {
		return fmt.Errorf("store: migração (modo_compacto): %w", err)
	}

	// client_id identifica o navegador que mandou a pergunta de Q&A — quem
	// perguntou pode remover a própria mensagem (ver RemoverPerguntaAoVivoPorCliente).
	// Bancos antigos ficam com '' (mensagem sem dono de navegador, não removível
	// pelo participante).
	if _, err := db.Exec(`ALTER TABLE live_qa_messages ADD COLUMN client_id TEXT NOT NULL DEFAULT ''`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("store: migração (live_qa_messages.client_id): %w", err)
		}
	}

	// Bancos criados antes do papel super admin existir: adiciona a coluna com
	// default 'organizer' (não quebra dados existentes).
	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'organizer'`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("store: migração (users.role): %w", err)
		}
	}

	// Bootstrap do super admin em bancos que já tinham usuários antes desta
	// coluna existir: promove o mais antigo. Idempotente — não faz nada se já
	// existe alguém com o papel (inclusive o caminho normal, onde CriarUsuario
	// já atribui o papel ao primeiro cadastro).
	if _, err := db.Exec(`
		UPDATE users SET role = 'super_admin'
		WHERE id = (SELECT id FROM users ORDER BY created_at ASC, id ASC LIMIT 1)
		  AND NOT EXISTS (SELECT 1 FROM users WHERE role = 'super_admin')
	`); err != nil {
		return fmt.Errorf("store: migração (bootstrap super admin): %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS app_settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS password_reset_tokens (
			token      TEXT PRIMARY KEY,
			user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user ON password_reset_tokens(user_id);
	`); err != nil {
		return fmt.Errorf("store: migração (app_settings/password_reset_tokens): %w", err)
	}

	// layout_view nunca foi usado pelo frontend (a apresentação ao vivo sempre
	// renderizou do mesmo jeito, independente do valor) — removido do schema.
	// Bancos criados antes desta remoção ainda têm a coluna (CREATE TABLE IF
	// NOT EXISTS acima não altera tabelas já criadas).
	if _, err := db.Exec(`ALTER TABLE questions DROP COLUMN layout_view`); err != nil {
		if !strings.Contains(err.Error(), "no such column") {
			return fmt.Errorf("store: migração (remover questions.layout_view): %w", err)
		}
	}

	// Bancos criados antes da opção "Outro" existir: adiciona sem quebrar
	// dados existentes, default false (nenhuma opção era "Outro" até aqui).
	if _, err := db.Exec(`ALTER TABLE question_options ADD COLUMN is_other BOOLEAN NOT NULL DEFAULT 0`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("store: migração (question_options.is_other): %w", err)
		}
	}

	return nil
}
