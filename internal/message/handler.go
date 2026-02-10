package message

import (
	"encoding/json"
	"net/http"

	"qr-doorbell/internal/auth"
	"qr-doorbell/internal/session"
)

type Handler struct {
	service        *Service
	sessionService *session.Service
	jwtSecret      string
}

func NewHandler(service *Service, sessionService *session.Service, jwtSecret string) *Handler {
	return &Handler{
		service:        service,
		sessionService: sessionService,
		jwtSecret:      jwtSecret,
	}
}

type createRequest struct {
	SessionID string `json:"session_id"`
	Sender    Sender `json:"sender"`
	Content   string `json:"content"`
}

func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "body inválido", http.StatusBadRequest)
		return
	}

	// Validar que la sesión existe
	sess, found := h.sessionService.Get(req.SessionID)
	if !found {
		http.Error(w, "sesión no encontrada o expirada", http.StatusNotFound)
		return
	}

	// Determinar el sender según la autenticación:
	// - Con JWT válido cuyo unit_id coincide con la sesión → owner
	// - Con JWT válido de admin → owner (puede supervisar cualquier sesión)
	// - Sin JWT → visitor (forzado, no importa qué mande el cliente)
	// - Con JWT pero unit_id no coincide → rechazado
	claims := auth.ClaimsFromRequest(r, h.jwtSecret)
	if claims != nil {
		// Usuario autenticado
		if claims.IsAdmin {
			// Admin puede escribir como owner en cualquier sesión
			req.Sender = SenderOwner
		} else if claims.UnitID == sess.UnitID {
			// Propietario del departamento correcto
			req.Sender = SenderOwner
		} else {
			// Autenticado pero no es dueño de este departamento
			http.Error(w, "no tenés permiso para escribir en esta conversación", http.StatusForbidden)
			return
		}
	} else {
		// Sin autenticación → siempre visitor
		req.Sender = SenderVisitor
	}

	msg, err := h.service.Create(req.SessionID, req.Sender, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")

	// Validar que la sesión existe
	if _, found := h.sessionService.Get(sessionID); !found {
		http.Error(w, "sesión no encontrada o expirada", http.StatusNotFound)
		return
	}

	msgs, err := h.service.List(sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}
