package auth

import "context"

type contextKey string

const contextKeyClaims contextKey = "claims"

// WithClaims guarda los claims en el contexto.
func WithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, contextKeyClaims, c)
}

// ClaimsFromContext devuelve los claims del contexto (o nil).
func ClaimsFromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(contextKeyClaims).(*Claims)
	return c
}
