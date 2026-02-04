package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const cookieName = "qrdoorbell_token"
const defaultExpire = 7 * 24 * time.Hour

var ErrNoSecret = errors.New("JWT_SECRET no configurado")

// Claims es el payload del JWT (subject = user ID).
type Claims struct {
	jwt.RegisteredClaims
	UserID  string `json:"uid"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"admin"`
}

// NewJWT genera un token para el usuario.
func NewJWT(secret string, userID, email string, isAdmin bool) (string, error) {
	if secret == "" {
		return "", ErrNoSecret
	}
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(defaultExpire)),
		},
		UserID:  userID,
		Email:   email,
		IsAdmin: isAdmin,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// ParseJWT valida el token y devuelve los claims.
func ParseJWT(secret, tokenString string) (*Claims, error) {
	if secret == "" {
		return nil, ErrNoSecret
	}
	t, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, errors.New("token inválido")
	}
	return claims, nil
}
