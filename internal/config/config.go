package config

import (
	"fmt"
	"os"
	"strconv"
)

// AppConfig contiene toda la configuración de la aplicación.
// Se carga principalmente desde variables de entorno para facilitar
// la ejecución en diferentes entornos (local, Docker, producción, etc.).
type AppConfig struct {
	// Port es el puerto HTTP donde se expone el servidor.
	// Ejemplo: "8081"
	Port string

	// BaseURL es la URL pública desde la cual se accederá al frontend.
	// Se utiliza principalmente para construir links en las notificaciones
	// de Telegram (por ejemplo: https://mi-dominio.com).
	BaseURL string

	// TelegramToken es el token del bot de Telegram.
	TelegramToken string

	// TelegramChatID es el chat (o canal) donde se enviarán las notificaciones.
	TelegramChatID string

	// DatabaseURL es el DSN/URL de conexión a Postgres.
	// Si está vacío, la aplicación funcionará en modo "in-memory"
	// para sesiones y mensajes.
	DatabaseURL string

	// JWTSecret es la clave para firmar los JWT de sesión. Obligatorio para auth.
	JWTSecret string

	// AdminEmail y AdminPassword definen el usuario admin inicial (bootstrap).
	// Si están ambos definidos, al arrancar se crea/actualiza ese usuario como admin.
	AdminEmail    string
	AdminPassword string

	// Geovalla: si GeofenceRadiusMeters > 0 y las coordenadas están definidas,
	// al tocar el timbre se exige que el visitante esté dentro del radio.
	DoorbellLat          float64 // DOORBELL_LAT (ej. -34.6037)
	DoorbellLon          float64 // DOORBELL_LON (ej. -58.3816)
	GeofenceRadiusMeters float64 // GEOFENCE_RADIUS_METERS (ej. 50)
}

// Load lee la configuración desde variables de entorno y aplica
// valores por defecto razonables para desarrollo local.
//
// Variables soportadas:
//   - PORT               (opcional, por defecto "8081")
//   - BASE_URL           (opcional, por defecto "http://localhost:8081")
//   - TELEGRAM_TOKEN     (opcional, si está vacío no se enviarán notificaciones)
//   - TELEGRAM_CHAT_ID   (opcional, si está vacío no se enviarán notificaciones)
//   - DOORBELL_LAT       (opcional, latitud del timbre para geovalla)
//   - DOORBELL_LON       (opcional, longitud del timbre para geovalla)
//   - GEOFENCE_RADIUS_METERS (opcional, radio en metros; si > 0 se exige estar dentro para tocar timbre)
func Load() (*AppConfig, error) {
	port := getenvDefault("PORT", "8081")
	baseURL := getenvDefault("BASE_URL", fmt.Sprintf("http://localhost:%s", port))

	cfg := &AppConfig{
		Port:           port,
		BaseURL:        baseURL,
		TelegramToken:  os.Getenv("TELEGRAM_TOKEN"),
		TelegramChatID: os.Getenv("TELEGRAM_CHAT_ID"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		AdminEmail:     os.Getenv("ADMIN_EMAIL"),
		AdminPassword:  os.Getenv("ADMIN_PASSWORD"),
	}
	if v, err := strconv.ParseFloat(os.Getenv("DOORBELL_LAT"), 64); err == nil {
		cfg.DoorbellLat = v
	}
	if v, err := strconv.ParseFloat(os.Getenv("DOORBELL_LON"), 64); err == nil {
		cfg.DoorbellLon = v
	}
	if v, err := strconv.ParseFloat(os.Getenv("GEOFENCE_RADIUS_METERS"), 64); err == nil && v > 0 {
		cfg.GeofenceRadiusMeters = v
	}

	return cfg, nil
}

// Addr devuelve la dirección de escucha para http.ListenAndServe, por ejemplo ":8081".
func (c *AppConfig) Addr() string {
	return ":" + c.Port
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
