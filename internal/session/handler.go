package session

import (
	"encoding/json"
	"net/http"

	"qr-doorbell/internal/config"
	"qr-doorbell/internal/geo"
)

type Handler struct {
	service *Service
	cfg     *config.AppConfig
}

func NewHandler(service *Service, cfg *config.AppConfig) *Handler {
	return &Handler{service: service, cfg: cfg}
}

type startRequest struct {
	DoorbellID string   `json:"doorbell_id"`
	UnitID     string   `json:"unit_id"`
	Lat        *float64 `json:"lat"`
	Lon        *float64 `json:"lon"`
}

func (h *Handler) StartSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req startRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "body inválido", http.StatusBadRequest)
		return
	}

	// Geovalla: si está configurada, exigir ubicación y validar que esté dentro del radio
	if h.cfg != nil && h.cfg.GeofenceRadiusMeters > 0 {
		if req.Lat == nil || req.Lon == nil {
			http.Error(w, "Para tocar el timbre hay que permitir la ubicación.", http.StatusBadRequest)
			return
		}
		if !geo.WithinRadius(h.cfg.DoorbellLat, h.cfg.DoorbellLon, *req.Lat, *req.Lon, h.cfg.GeofenceRadiusMeters) {
			http.Error(w, "Tenés que estar cerca del edificio para tocar el timbre.", http.StatusForbidden)
			return
		}
	}

	session, err := h.service.Start(req.DoorbellID, req.UnitID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}
