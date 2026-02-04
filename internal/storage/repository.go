package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Message representa un mensaje persistido. Refleja internal/message.Message
// pero separado para no crear dependencias circulares.
type Message struct {
	ID        string
	SessionID string
	Sender    string
	Content   string
	CreatedAt time.Time
}

// Session representa una sesión persistida. Similar a internal/session.Session.
type Session struct {
	ID         string
	DoorbellID string
	UnitID     string
	Status     string
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

// User representa un usuario (propietario o admin) persistido.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	UnitID       string
	IsAdmin      bool
	CreatedAt    time.Time
}

// UserRepository define las operaciones de persistencia de usuarios.
type UserRepository interface {
	Store(ctx context.Context, u User) error
	GetByID(ctx context.Context, id string) (User, bool, error)
	GetByEmail(ctx context.Context, email string) (User, bool, error)
	List(ctx context.Context) ([]User, error)
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	// EnsureAdmin crea o actualiza un usuario admin por email (bootstrap desde env).
	EnsureAdmin(ctx context.Context, email, passwordHash string) error
}

// MessageRepository define las operaciones de persistencia de mensajes.
type MessageRepository interface {
	Store(ctx context.Context, m Message) error
	ListBySession(ctx context.Context, sessionID string) ([]Message, error)
}

// SessionRepository define las operaciones de persistencia de sesiones.
type SessionRepository interface {
	Store(ctx context.Context, s Session) error
	Get(ctx context.Context, id string) (Session, bool, error)
}

// PostgresMessageRepository es una implementación de MessageRepository
// usando Postgres.
type PostgresMessageRepository struct {
	db *sql.DB
}

func NewPostgresMessageRepository(db *sql.DB) *PostgresMessageRepository {
	return &PostgresMessageRepository{db: db}
}

func (r *PostgresMessageRepository) Store(ctx context.Context, m Message) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO messages (id, session_id, sender, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, m.ID, m.SessionID, m.Sender, m.Content, m.CreatedAt)
	return err
}

func (r *PostgresMessageRepository) ListBySession(ctx context.Context, sessionID string) ([]Message, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, session_id, sender, content, created_at
		FROM messages
		WHERE session_id = $1
		ORDER BY created_at ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Sender, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}

	return out, rows.Err()
}

// PostgresSessionRepository es una implementación de SessionRepository
// usando Postgres.
type PostgresSessionRepository struct {
	db *sql.DB
}

func NewPostgresSessionRepository(db *sql.DB) *PostgresSessionRepository {
	return &PostgresSessionRepository{db: db}
}

func (r *PostgresSessionRepository) Store(ctx context.Context, s Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (id, doorbell_id, unit_id, status, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, s.ID, s.DoorbellID, s.UnitID, s.Status, s.CreatedAt, s.ExpiresAt)
	return err
}

func (r *PostgresSessionRepository) Get(ctx context.Context, id string) (Session, bool, error) {
	var s Session
	err := r.db.QueryRowContext(ctx, `
		SELECT id, doorbell_id, unit_id, status, created_at, expires_at
		FROM sessions
		WHERE id = $1
	`, id).Scan(&s.ID, &s.DoorbellID, &s.UnitID, &s.Status, &s.CreatedAt, &s.ExpiresAt)
	if err == sql.ErrNoRows {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, err
	}
	return s, true, nil
}

// PostgresUserRepository implementa UserRepository con Postgres.
type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Store(ctx context.Context, u User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, unit_id, is_admin, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, u.ID, u.Email, u.PasswordHash, u.UnitID, u.IsAdmin, u.CreatedAt)
	return err
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (User, bool, error) {
	var u User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, unit_id, is_admin, created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.UnitID, &u.IsAdmin, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, err
	}
	return u, true, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (User, bool, error) {
	var u User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, unit_id, is_admin, created_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.UnitID, &u.IsAdmin, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, err
	}
	return u, true, nil
}

func (r *PostgresUserRepository) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, email, password_hash, unit_id, is_admin, created_at
		FROM users ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.UnitID, &u.IsAdmin, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *PostgresUserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, passwordHash, userID)
	return err
}

func (r *PostgresUserRepository) EnsureAdmin(ctx context.Context, email, passwordHash string) error {
	u, ok, err := r.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if ok {
		_, err = r.db.ExecContext(ctx, `UPDATE users SET password_hash = $1, is_admin = true WHERE id = $2`, passwordHash, u.ID)
		return err
	}
	id := uuid.New().String()
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, unit_id, is_admin, created_at)
		VALUES ($1, $2, $3, '', true, $4)
	`, id, email, passwordHash, time.Now())
	return err
}
