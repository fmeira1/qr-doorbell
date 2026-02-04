package message

import (
	"context"
	"encoding/json"
	"errors"
	"qr-doorbell/internal/notify"
	"qr-doorbell/internal/storage"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	mu        sync.Mutex
	messages  map[string][]Message
	broadcast func(sessionID string, data []byte)
	telegram  *notify.TelegramNotifier
	baseURL   string

	repo         storage.MessageRepository
	getSessionFn func(sessionID string) (unitID string, unitNumber string, ok bool)
}

func NewService(
	broadcast func(string, []byte),
	telegram *notify.TelegramNotifier,
	baseURL string,
	repo storage.MessageRepository,
	getSessionFn func(sessionID string) (unitID string, unitNumber string, ok bool),
) *Service {
	return &Service{
		messages:     make(map[string][]Message),
		broadcast:    broadcast,
		telegram:     telegram,
		baseURL:      baseURL,
		repo:         repo,
		getSessionFn: getSessionFn,
	}
}

func (s *Service) Create(sessionID string, sender Sender, content string) (Message, error) {
	if sessionID == "" {
		return Message{}, errors.New("session_id vacío")
	}
	if content == "" {
		return Message{}, errors.New("contenido vacío")
	}

	msg := Message{
		ID:        uuid.NewString(),
		SessionID: sessionID,
		Sender:    sender,
		Content:   content,
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	existing := s.messages[sessionID]
	isFirst := len(existing) == 0
	s.messages[sessionID] = append(existing, msg)
	s.mu.Unlock()

	if s.repo != nil {
		_ = s.repo.Store(context.Background(), storage.Message{
			ID:        msg.ID,
			SessionID: msg.SessionID,
			Sender:    string(msg.Sender),
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt,
		})
	}

	if s.broadcast != nil {
		data, _ := json.Marshal(msg)
		s.broadcast(sessionID, data)
	}

	if s.telegram != nil && sender == SenderVisitor && isFirst {
		unitInfo := ""
		if s.getSessionFn != nil {
			_, unitNumber, ok := s.getSessionFn(sessionID)
			if ok && unitNumber != "" {
				unitInfo = "Departamento: " + unitNumber + "\n\n"
			}
		}

		text := "🔔 Alguien está en la puerta\n\n" +
			unitInfo +
			"Mensaje:\n" + content

		chatURL := s.baseURL + "/index.html?session_id=" + sessionID
		s.telegram.NotifyWithButtons(text, sessionID, chatURL)
	}

	return msg, nil
}

func (s *Service) List(sessionID string) ([]Message, error) {
	if sessionID == "" {
		return nil, errors.New("session_id vacío")
	}

	s.mu.Lock()
	inMemory := s.messages[sessionID]
	s.mu.Unlock()

	if len(inMemory) > 0 || s.repo == nil {
		return inMemory, nil
	}

	dbMessages, err := s.repo.ListBySession(context.Background(), sessionID)
	if err != nil {
		return nil, err
	}

	out := make([]Message, 0, len(dbMessages))
	for _, m := range dbMessages {
		out = append(out, Message{
			ID:        m.ID,
			SessionID: m.SessionID,
			Sender:    Sender(m.Sender),
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
		})
	}
	return out, nil
}
