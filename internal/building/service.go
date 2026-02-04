package building

import "errors"

type Service struct {
	buildings map[string]Building
	units     map[string]Unit
}

func NewService() *Service {
	// Por ahora, un solo edificio con unidades de ejemplo
	buildingID := "building-1"
	buildings := map[string]Building{
		buildingID: {
			ID:   buildingID,
			Name: "Edificio Principal",
		},
	}

	// Nomenclatura: PB A y B; 1C, 1D, 1E; 2F, 2G, 2H (2 plantas)
	units := map[string]Unit{
		"unit-1": {
			ID:         "unit-1",
			BuildingID: buildingID,
			Number:     "PB A",
			Name:       "Planta Baja A",
			Active:     true,
		},
		"unit-2": {
			ID:         "unit-2",
			BuildingID: buildingID,
			Number:     "PB B",
			Name:       "Planta Baja B",
			Active:     true,
		},
		"unit-3": {
			ID:         "unit-3",
			BuildingID: buildingID,
			Number:     "1C",
			Name:       "Primer Piso C",
			Active:     true,
		},
		"unit-4": {
			ID:         "unit-4",
			BuildingID: buildingID,
			Number:     "1D",
			Name:       "Primer Piso D",
			Active:     true,
		},
		"unit-5": {
			ID:         "unit-5",
			BuildingID: buildingID,
			Number:     "1E",
			Name:       "Primer Piso E",
			Active:     true,
		},
		"unit-6": {
			ID:         "unit-6",
			BuildingID: buildingID,
			Number:     "2F",
			Name:       "Segundo Piso F",
			Active:     true,
		},
		"unit-7": {
			ID:         "unit-7",
			BuildingID: buildingID,
			Number:     "2G",
			Name:       "Segundo Piso G",
			Active:     true,
		},
		"unit-8": {
			ID:         "unit-8",
			BuildingID: buildingID,
			Number:     "2H",
			Name:       "Segundo Piso H",
			Active:     true,
		},
	}

	return &Service{
		buildings: buildings,
		units:     units,
	}
}

func (s *Service) GetBuilding(id string) (Building, error) {
	if id == "" {
		return Building{}, errors.New("building id vacío")
	}
	building, ok := s.buildings[id]
	if !ok {
		return Building{}, errors.New("edificio no encontrado")
	}
	return building, nil
}

func (s *Service) GetUnit(id string) (Unit, error) {
	if id == "" {
		return Unit{}, errors.New("unit id vacío")
	}
	unit, ok := s.units[id]
	if !ok {
		return Unit{}, errors.New("departamento no encontrado")
	}
	return unit, nil
}

// unitDisplayOrder define el orden de aparición: PB A, PB B, 1C, 1D, 1E, 2F, 2G, 2H
var unitDisplayOrder = []string{"unit-1", "unit-2", "unit-3", "unit-4", "unit-5", "unit-6", "unit-7", "unit-8"}

func (s *Service) ListUnitsByBuilding(buildingID string) ([]Unit, error) {
	if buildingID == "" {
		return nil, errors.New("building id vacío")
	}
	var result []Unit
	for _, id := range unitDisplayOrder {
		unit, ok := s.units[id]
		if ok && unit.BuildingID == buildingID && unit.Active {
			result = append(result, unit)
		}
	}
	return result, nil
}

func (s *Service) ListBuildings() []Building {
	var result []Building
	for _, b := range s.buildings {
		result = append(result, b)
	}
	return result
}
