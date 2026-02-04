package building

type Building struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Unit struct {
	ID         string `json:"id"`
	BuildingID string `json:"building_id"`
	Number     string `json:"number"` // ej: "3A", "PB B", "5C"
	Name       string `json:"name"`   // opcional: "Departamento 3A"
	Active     bool   `json:"active"`
}
