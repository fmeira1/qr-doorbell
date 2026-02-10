package admin

import (
	"encoding/json"
	"net/http"

	"qr-doorbell/internal/user"
)

type Handler struct {
	userService *user.Service
}

func NewHandler(userService *user.Service) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) Users(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listUsers(w, r)
	case http.MethodPost:
		h.createUser(w, r)
	case http.MethodDelete:
		h.deleteUser(w, r)
	default:
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var in user.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "body inválido", http.StatusBadRequest)
		return
	}
	u, err := h.userService.Create(r.Context(), in)
	if err != nil {
		if err == user.ErrEmailExists {
			http.Error(w, "el email ya está registrado", http.StatusConflict)
			return
		}
		if err == user.ErrInvalidInput {
			http.Error(w, "email, contraseña y unidad son obligatorios", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "falta el parámetro id", http.StatusBadRequest)
		return
	}
	err := h.userService.Delete(r.Context(), id)
	if err != nil {
		if err == user.ErrUserNotFound {
			http.Error(w, "usuario no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	list, err := h.userService.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}
