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
			created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE UNIQUE INDEX IF NOT EXISTS idx_events_pin_active
			ON events(pin_code) WHERE status != 'FINISHED';

		CREATE TABLE IF NOT EXISTS questions (
			id          INTEGER PRIMARY KEY,
			event_id    INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
			title       TEXT    NOT NULL,
			type        TEXT    NOT NULL,
			layout_view TEXT    NOT NULL DEFAULT '',
			order_index INTEGER NOT NULL DEFAULT 0,
			created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS question_options (
			id          INTEGER PRIMARY KEY,
			question_id INTEGER NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
			text_label  TEXT    NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_questions_event ON questions(event_id);
		CREATE INDEX IF NOT EXISTS idx_question_options_question ON question_options(question_id);

		CREATE TABLE IF NOT EXISTS participants (
			id         INTEGER PRIMARY KEY,
			event_id   INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
			email      TEXT    NOT NULL,
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

	return nil
}
