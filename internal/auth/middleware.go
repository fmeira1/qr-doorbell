package auth

import (
	"net/http"
	"strings"
)

// ClaimsFromRequest extrae y valida el JWT del request (cookie o header Authorization).
func ClaimsFromRequest(r *http.Request, secret string) *Claims {
	// 1) Cookie
	if c, err := r.Cookie(cookieName); err == nil && c.Value != "" {
		claims, err := ParseJWT(secret, c.Value)
		if err == nil {
			return claims
		}
	}
	// 2) Bearer
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		claims, err := ParseJWT(secret, token)
		if err == nil {
			return claims
		}
	}
	return nil
}

// RequireAuth es un middleware que exige un usuario autenticado.
// Añade los claims al contexto con key contextKeyClaims.
func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromRequest(r, secret)
			if claims == nil {
				http.Error(w, "no autorizado", http.StatusUnauthorized)
				return
			}
			ctx := WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin es un middleware que exige un usuario autenticado y con is_admin.
func RequireAdmin(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromRequest(r, secret)
			if claims == nil {
				http.Error(w, "no autorizado", http.StatusUnauthorized)
				return
			}
			if !claims.IsAdmin {
				http.Error(w, "requiere rol administrador", http.StatusForbidden)
				return
			}
			ctx := WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
