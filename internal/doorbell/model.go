package doorbell

type Doorbell struct {
	ID         string `json:"id"`
	BuildingID string `json:"building_id"`
	Name       string `json:"name"`
	Active     bool   `json:"active"`
}
