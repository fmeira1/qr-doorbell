package api

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime/debug"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// RecoveryMiddleware atrapa cualquier panic en un handler HTTP y devuelve 500
// en vez de dejar que el proceso se caiga. Loguea el stack trace completo.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				log.Printf("[PANIC RECOVERY] %s %s — error: %v\n%s",
					r.Method, r.URL.Path, err, stack)

				// Si todavía no se escribió nada al response, devolver 500.
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "Error interno del servidor")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware loguea cada request con método, ruta y status code.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("%s %s %d", r.Method, r.URL.Path, rw.status)
	})
}

// responseWriter es un wrapper que captura el status code para logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.wroteHeader = true
	}
	return rw.ResponseWriter.Write(b)
}

// Hijack implementa http.Hijacker para que WebSocket upgrades funcionen
// a través del LoggingMiddleware.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("el ResponseWriter subyacente no soporta Hijack")
	}
	return h.Hijack()
}
