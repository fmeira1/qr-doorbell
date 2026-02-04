package doorbell

import "errors"

type Service struct {
	data map[string]Doorbell
}

func NewService() *Service {
	return &Service{
		data: map[string]Doorbell{
			"abc123": {
				ID:         "abc123",
				BuildingID: "building-1",
				Name:       "Puerta Principal",
				Active:     true,
			},
		},
	}
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
