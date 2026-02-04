package session

import (
	"context"
	"errors"
	"sync"
	"time"

	"qr-doorbell/internal/storage"

	"github.com/google/uuid"
)

type Service struct {
	mu       sync.Mutex
	sessions map[string]Session
	ttl      time.Duration

	repo storage.SessionRepository
}

func NewService(ttl time.Duration, repo storage.SessionRepository) *Service {
	return &Service{
		sessions: make(map[string]Session),
		ttl:      ttl,
		repo:     repo,
	}
}

func (s *Service) Start(doorbellID, unitID string) (Session, error) {
	if doorbellID == "" {
		return Session{}, errors.New("doorbell_id vacío")
	}
	if unitID == "" {
		return Session{}, errors.New("unit_id vacío")
	}

	now := time.Now()
	session := Session{
		ID:         uuid.NewString(),
		DoorbellID: doorbellID,
		UnitID:     unitID,
		Status:     StatusOpen,
		CreatedAt:  now,
		ExpiresAt:  now.Add(s.ttl),
	}

	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()

	if s.repo != nil {
		_ = s.repo.Store(context.Background(), storage.Session{
			ID:         session.ID,
			DoorbellID: session.DoorbellID,
			UnitID:     session.UnitID,
			Status:     string(session.Status),
			CreatedAt:  session.CreatedAt,
			ExpiresAt:  session.ExpiresAt,
		})
	}

	return session, nil
}

func (s *Service) Get(id string) (Session, bool) {
	s.mu.Lock()
	session, ok := s.sessions[id]
	s.mu.Unlock()

	if ok {
		if time.Now().After(session.ExpiresAt) {
			s.mu.Lock()
			delete(s.sessions, id)
			s.mu.Unlock()
			return Session{}, false
		}
		return session, true
	}

	if s.repo == nil {
		return Session{}, false
	}

	dbSession, found, err := s.repo.Get(context.Background(), id)
	if err != nil || !found {
		return Session{}, false
	}

	if time.Now().After(dbSession.ExpiresAt) {
		return Session{}, false
	}

	return Session{
		ID:         dbSession.ID,
		DoorbellID: dbSession.DoorbellID,
		UnitID:     dbSession.UnitID,
		Status:     Status(dbSession.Status),
		CreatedAt:  dbSession.CreatedAt,
		ExpiresAt:  dbSession.ExpiresAt,
	}, true
}
