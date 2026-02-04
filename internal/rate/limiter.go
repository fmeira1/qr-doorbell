package rate

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// LimiterConfig define los parámetros básicos del rate limiting.
type LimiterConfig struct {
	// Requests es la cantidad máxima de requests permitidos
	// por ventana de tiempo.
	Requests int
	// Window es la duración de la ventana de tiempo.
	Window time.Duration
}

type visitor struct {
	remaining int
	resetAt   time.Time
}

// Limiter es un rate limiter muy simple basado en ventanas fijas
// por clave (normalmente IP+ruta).
type Limiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	cfg      LimiterConfig
}

// NewLimiter crea un limiter en memoria.
func NewLimiter(cfg LimiterConfig) *Limiter {
	if cfg.Requests <= 0 {
		cfg.Requests = 30
	}
	if cfg.Window <= 0 {
		cfg.Window = 10 * time.Second
	}
	return &Limiter{
		visitors: make(map[string]*visitor),
		cfg:      cfg,
	}
}

// Allow devuelve true si la request está permitida para la clave dada.
func (l *Limiter) Allow(key string) bool {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	v, ok := l.visitors[key]
	if !ok || now.After(v.resetAt) {
		l.visitors[key] = &visitor{
			remaining: l.cfg.Requests - 1,
			resetAt:   now.Add(l.cfg.Window),
		}
		return true
	}

	if v.remaining <= 0 {
		return false
	}

	v.remaining--
	return true
}

// RateLimitMiddleware crea un middleware que limita las requests
// por IP y ruta.
//
// Ejemplo de uso:
//
//	lim := rate.NewLimiter(rate.LimiterConfig{Requests: 10, Window: time.Second})
//	mux.Handle("/path", rate.RateLimitMiddleware(lim, handler))
func RateLimitMiddleware(l *Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		// Clave = IP + ruta.
		key := ip + "|" + r.URL.Path

		if !l.Allow(key) {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("too many requests"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
