package geo

import "math"

// DistanceMeters calcula la distancia en metros entre dos puntos (fórmula de Haversine).
// lat1, lon1 y lat2, lon2 están en grados decimales.
func DistanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusM = 6371000 // metros
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}

// WithinRadius devuelve true si el punto (lat2, lon2) está a distancia <= radiusMeters del punto (lat1, lon1).
func WithinRadius(lat1, lon1, lat2, lon2, radiusMeters float64) bool {
	if radiusMeters <= 0 {
		return true
	}
	return DistanceMeters(lat1, lon1, lat2, lon2) <= radiusMeters
}
