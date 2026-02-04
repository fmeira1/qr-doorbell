package user

import (
	"context"
	"sync"
	"time"

	"qr-doorbell/internal/storage"

	"github.com/google/uuid"
)

// MemoryRepository implementa UserRepository en memoria (sin Postgres).
type MemoryRepository struct {
	mu      sync.RWMutex
	byID    map[string]storage.User
	byEmail map[string]string // email -> id
}

// NewMemoryRepository crea un repositorio de usuarios en memoria.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		byID:    make(map[string]storage.User),
		byEmail: make(map[string]string),
	}
}

func (r *MemoryRepository) Store(ctx context.Context, u storage.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[u.ID] = u
	r.byEmail[u.Email] = u.ID
	return nil
}

func (r *MemoryRepository) GetByID(ctx context.Context, id string) (storage.User, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	return u, ok, nil
}

func (r *MemoryRepository) GetByEmail(ctx context.Context, email string) (storage.User, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[email]
	if !ok {
		return storage.User{}, false, nil
	}
	u := r.byID[id]
	return u, true, nil
}

func (r *MemoryRepository) List(ctx context.Context) ([]storage.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]storage.User, 0, len(r.byID))
	for _, u := range r.byID {
		out = append(out, u)
	}
	return out, nil
}

func (r *MemoryRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.byID[userID]
	if !ok {
		return nil
	}
	u.PasswordHash = passwordHash
	r.byID[userID] = u
	return nil
}

// EnsureAdmin crea o actualiza el usuario admin desde email/password (para bootstrap).
func (r *MemoryRepository) EnsureAdmin(ctx context.Context, email, passwordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if id, ok := r.byEmail[email]; ok {
		u := r.byID[id]
		u.PasswordHash = passwordHash
		u.IsAdmin = true
		r.byID[id] = u
		return nil
	}
	u := storage.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: passwordHash,
		UnitID:       "",
		IsAdmin:      true,
		CreatedAt:    time.Now(),
	}
	r.byID[u.ID] = u
	r.byEmail[u.Email] = u.ID
	return nil
}
