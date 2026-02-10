package doorbell

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetActiveID devuelve el doorbell_id activo (endpoint público).
func (h *Handler) GetActiveID(w http.ResponseWriter, r *http.Request) {
	id := h.service.GetDoorbellID()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"doorbell_id": id})
}

func (h *Handler) GetDoorbell(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/doorbell/")

	doorbell, err := h.service.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doorbell)
}
