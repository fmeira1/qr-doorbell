package auth

import (
	"encoding/json"
	"net/http"

	"qr-doorbell/internal/user"
)

type Handler struct {
	userService *user.Service
	jwtSecret   string
}

func NewHandler(userService *user.Service, jwtSecret string) *Handler {
	return &Handler{userService: userService, jwtSecret: jwtSecret}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "credenciales inválidas", http.StatusBadRequest)
		return
	}
	u, err := h.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == user.ErrInvalidLogin {
			http.Error(w, "email o contraseña incorrectos", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token, err := NewJWT(h.jwtSecret, u.ID, u.Email, u.UnitID, u.IsAdmin)
	if err != nil {
		http.Error(w, "error generando sesión", http.StatusInternalServerError)
		return
	}
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(defaultExpire.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromRequest(r, h.jwtSecret)
	if claims == nil {
		http.Error(w, "no autorizado", http.StatusUnauthorized)
		return
	}
	u, err := h.userService.GetByID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "usuario no encontrado", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	claims := ClaimsFromRequest(r, h.jwtSecret)
	if claims == nil {
		http.Error(w, "no autorizado", http.StatusUnauthorized)
		return
	}
	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "datos inválidos", http.StatusBadRequest)
		return
	}
	err := h.userService.ChangePassword(r.Context(), claims.UserID, user.ChangePasswordInput{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
	if err != nil {
		if err == user.ErrInvalidLogin {
			http.Error(w, "contraseña actual incorrecta", http.StatusUnauthorized)
			return
		}
		if err == user.ErrInvalidInput {
			http.Error(w, "la nueva contraseña debe tener al menos 6 caracteres", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
