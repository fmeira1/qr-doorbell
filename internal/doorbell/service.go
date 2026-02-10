package doorbell

import "errors"

type Service struct {
	data map[string]Doorbell
}

func NewService(doorbellID string) *Service {
	if doorbellID == "" {
		doorbellID = "abc123"
	}
	return &Service{
		data: map[string]Doorbell{
			doorbellID: {
				ID:         doorbellID,
				BuildingID: "building-1",
				Name:       "Puerta Principal",
				Active:     true,
			},
		},
	}
}

// GetDoorbellID devuelve el ID del timbre activo (solo hay uno).
func (s *Service) GetDoorbellID() string {
	for id := range s.data {
		return id
	}
	return ""
}

func (s *Service) GetByID(id string) (Doorbell, error) {
	if id == "" {
		return Doorbell{}, errors.New("id vacío")
	}

	doorbell, ok := s.data[id]
	if !ok {
		return Doorbell{}, errors.New("doorbell no encontrado")
	}

	return doorbell, nil
}
