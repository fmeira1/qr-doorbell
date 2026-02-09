package building

import (
	"encoding/json"
	"net/http"
	"strings"

	"qr-doorbell/internal/storage"
)

type Handler struct {
	service  *Service
	userRepo storage.UserRepository
}

func NewHandler(service *Service, userRepo storage.UserRepository) *Handler {
	return &Handler{service: service, userRepo: userRepo}
}

func (h *Handler) GetBuilding(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/building/")
	building, err := h.service.GetBuilding(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(building)
}

func (h *Handler) GetUnit(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/unit/")
	unit, err := h.service.GetUnit(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unit)
}

func (h *Handler) ListUnits(w http.ResponseWriter, r *http.Request) {
	buildingID := r.URL.Query().Get("building_id")
	if buildingID == "" {
		http.Error(w, "building_id requerido", http.StatusBadRequest)
		return
	}
	units, err := h.service.ListUnitsByBuilding(buildingID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Si occupied=true, filtrar solo unidades con al menos un usuario registrado.
	if r.URL.Query().Get("occupied") == "true" && h.userRepo != nil {
		occupied, err := h.userRepo.ListOccupiedUnitIDs(r.Context())
		if err == nil && len(occupied) > 0 {
			set := make(map[string]bool, len(occupied))
			for _, uid := range occupied {
				set[uid] = true
			}
			filtered := make([]Unit, 0, len(units))
			for _, u := range units {
				if set[u.ID] {
					filtered = append(filtered, u)
				}
			}
			units = filtered
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(units)
}

func (h *Handler) ListBuildings(w http.ResponseWriter, r *http.Request) {
	buildings := h.service.ListBuildings()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(buildings)
}
