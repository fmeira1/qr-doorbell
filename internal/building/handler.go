package building

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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(units)
}

func (h *Handler) ListBuildings(w http.ResponseWriter, r *http.Request) {
	buildings := h.service.ListBuildings()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(buildings)
}
