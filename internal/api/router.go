package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"qr-doorbell/internal/admin"
	"qr-doorbell/internal/auth"
	"qr-doorbell/internal/building"
	"qr-doorbell/internal/config"
	"qr-doorbell/internal/doorbell"
	"qr-doorbell/internal/message"
	"qr-doorbell/internal/notify"
	"qr-doorbell/internal/rate"
	"qr-doorbell/internal/realtime"
	"qr-doorbell/internal/session"
	"qr-doorbell/internal/storage"
	"qr-doorbell/internal/user"

	"strings"
	"time"
)

func NewRouter(cfg *config.AppConfig, db *storage.DB) http.Handler {
	mux := http.NewServeMux()
	hub := realtime.NewHub()

	var msgRepo storage.MessageRepository
	var sessionRepo storage.SessionRepository
	var userRepo storage.UserRepository
	if db != nil {
		msgRepo = storage.NewPostgresMessageRepository(db.SQL)
		sessionRepo = storage.NewPostgresSessionRepository(db.SQL)
		userRepo = storage.NewPostgresUserRepository(db.SQL)
	} else {
		userRepo = user.NewMemoryRepository()
	}

	var telegram *notify.TelegramNotifier
	if cfg.TelegramToken != "" && cfg.TelegramChatID != "" {
		telegram = notify.NewTelegramNotifier(
			cfg.TelegramToken,
			cfg.TelegramChatID,
		)
	}

	// Servicios
	doorbellService := doorbell.NewService()
	buildingService := building.NewService()
	sessionService := session.NewService(5*time.Minute, sessionRepo)
	userService := user.NewService(userRepo)

	// Bootstrap admin si están definidos ADMIN_EMAIL y ADMIN_PASSWORD
	if cfg.AdminEmail != "" && cfg.AdminPassword != "" {
		if err := userService.EnsureAdmin(context.Background(), cfg.AdminEmail, cfg.AdminPassword); err != nil {
			log.Printf("aviso: no se pudo crear/actualizar admin: %v", err)
		} else {
			log.Println("Admin bootstrap OK:", cfg.AdminEmail)
		}
	}

	// Auth y admin (solo si JWT_SECRET está definido)
	if cfg.JWTSecret != "" {
		authHandler := auth.NewHandler(userService, cfg.JWTSecret)
		adminHandler := admin.NewHandler(userService)
		mux.HandleFunc("/auth/login", authHandler.Login)
		mux.HandleFunc("/auth/logout", authHandler.Logout)
		mux.Handle("/auth/me", auth.RequireAuth(cfg.JWTSecret)(http.HandlerFunc(authHandler.Me)))
		mux.Handle("/auth/change-password", auth.RequireAuth(cfg.JWTSecret)(http.HandlerFunc(authHandler.ChangePassword)))
		mux.Handle("/admin/users", auth.RequireAdmin(cfg.JWTSecret)(http.HandlerFunc(adminHandler.Users)))
	}

	// Función helper para obtener info de sesión (unit) desde message service
	getSessionInfo := func(sessionID string) (unitID string, unitNumber string, ok bool) {
		sess, found := sessionService.Get(sessionID)
		if !found {
			return "", "", false
		}
		// Intentar obtener el número del unit desde building service
		unit, err := buildingService.GetUnit(sess.UnitID)
		if err != nil {
			return sess.UnitID, "", true
		}
		return sess.UnitID, unit.Number, true
	}

	messageService := message.NewService(
		hub.Broadcast,
		telegram,
		cfg.BaseURL,
		msgRepo,
		getSessionInfo,
	)

	if telegram != nil {
		telegram.StartPolling(func(sessionID, text string) {
			if _, err := messageService.Create(sessionID, message.SenderOwner, text); err != nil {
				fmt.Println("telegram onReply: error creando mensaje:", err, "sessionID:", sessionID)
			}
		})
	}

	//handlers y llamadas
	doorbellHandler := doorbell.NewHandler(doorbellService)
	buildingHandler := building.NewHandler(buildingService, userRepo)
	sessionHandler := session.NewHandler(sessionService, cfg)
	messageHandler := message.NewHandler(messageService)

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/doorbell/", doorbellHandler.GetDoorbell)
	mux.HandleFunc("/building/", buildingHandler.GetBuilding)
	mux.HandleFunc("/unit/", buildingHandler.GetUnit)
	mux.HandleFunc("/units", buildingHandler.ListUnits)
	mux.HandleFunc("/buildings", buildingHandler.ListBuildings)
	// rate limiting sencillo por IP+ruta
	limiter := rate.NewLimiter(rate.LimiterConfig{
		Requests: 10,
		Window:   5 * time.Second,
	})

	mux.Handle("/session/start", rate.RateLimitMiddleware(limiter, http.HandlerFunc(sessionHandler.StartSession)))
	mux.Handle("/message", rate.RateLimitMiddleware(limiter, http.HandlerFunc(messageHandler.CreateMessage)))
	mux.Handle("/messages", rate.RateLimitMiddleware(limiter, http.HandlerFunc(messageHandler.ListMessages)))
	mux.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		sessionID := strings.TrimPrefix(r.URL.Path, "/ws/")
		hub.HandleWS(w, r, sessionID)
	})

	// Servir archivos estáticos (frontend)
	staticDir := filepath.Join(".", "web", "public")
	fs := http.FileServer(http.Dir(staticDir))
	mux.Handle("/", fs)

	// Middlewares globales: recovery (atrapa panics) + logging (loguea requests).
	// Se aplican en orden inverso: logging corre primero, recovery envuelve todo.
	var handler http.Handler = mux
	handler = LoggingMiddleware(handler)
	handler = RecoveryMiddleware(handler)

	return handler
}
