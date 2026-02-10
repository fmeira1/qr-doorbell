package user

import (
	"context"
	"errors"
	"time"

	"qr-doorbell/internal/storage"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailExists  = errors.New("el email ya está registrado")
	ErrInvalidLogin = errors.New("email o contraseña incorrectos")
	ErrInvalidInput = errors.New("datos inválidos")
	ErrUserNotFound = errors.New("usuario no encontrado")
)

const bcryptCost = 10

type Service struct {
	repo storage.UserRepository
}

func NewService(repo storage.UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *Service) checkPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// EnsureAdmin crea o actualiza el usuario admin (bootstrap desde env).
func (s *Service) EnsureAdmin(ctx context.Context, email, password string) error {
	if email == "" || password == "" {
		return nil
	}
	hash, err := s.hashPassword(password)
	if err != nil {
		return err
	}
	return s.repo.EnsureAdmin(ctx, email, string(hash))
}

// Create crea un usuario (solo admin). Password en claro, se hashea aquí.
func (s *Service) Create(ctx context.Context, in CreateInput) (User, error) {
	if in.Email == "" || in.Password == "" || in.UnitID == "" {
		return User{}, ErrInvalidInput
	}
	_, ok, err := s.repo.GetByEmail(ctx, in.Email)
	if err != nil {
		return User{}, err
	}
	if ok {
		return User{}, ErrEmailExists
	}
	hash, err := s.hashPassword(in.Password)
	if err != nil {
		return User{}, err
	}
	u := storage.User{
		ID:           uuid.New().String(),
		Email:        in.Email,
		PasswordHash: string(hash),
		UnitID:       in.UnitID,
		IsAdmin:      false,
		CreatedAt:    time.Now(),
	}
	if err := s.repo.Store(ctx, u); err != nil {
		return User{}, err
	}
	return toUser(u), nil
}

// Login verifica email/password y devuelve el usuario si es correcto.
func (s *Service) Login(ctx context.Context, email, password string) (User, error) {
	if email == "" || password == "" {
		return User{}, ErrInvalidLogin
	}
	su, ok, err := s.repo.GetByEmail(ctx, email)
	if err != nil || !ok {
		return User{}, ErrInvalidLogin
	}
	if !s.checkPassword(su.PasswordHash, password) {
		return User{}, ErrInvalidLogin
	}
	return toUser(su), nil
}

// GetByID devuelve un usuario por ID (sin password).
func (s *Service) GetByID(ctx context.Context, id string) (User, error) {
	su, ok, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	if !ok {
		return User{}, ErrUserNotFound
	}
	return toUser(su), nil
}

// List devuelve todos los usuarios (sin password).
func (s *Service) List(ctx context.Context) ([]User, error) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]User, len(list))
	for i := range list {
		out[i] = toUser(list[i])
	}
	return out, nil
}

// Delete elimina un usuario por ID. No permite eliminar admins.
func (s *Service) Delete(ctx context.Context, id string) error {
	u, ok, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUserNotFound
	}
	if u.IsAdmin {
		return errors.New("no se puede eliminar un administrador")
	}
	return s.repo.Delete(ctx, id)
}

// ChangePassword cambia la contraseña del usuario (verificando la actual).
func (s *Service) ChangePassword(ctx context.Context, userID string, in ChangePasswordInput) error {
	if in.NewPassword == "" || len(in.NewPassword) < 6 {
		return ErrInvalidInput
	}
	su, ok, err := s.repo.GetByID(ctx, userID)
	if err != nil || !ok {
		return ErrUserNotFound
	}
	if !s.checkPassword(su.PasswordHash, in.CurrentPassword) {
		return ErrInvalidLogin
	}
	hash, err := s.hashPassword(in.NewPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, userID, string(hash))
}

func toUser(s storage.User) User {
	return User{
		ID:        s.ID,
		Email:     s.Email,
		UnitID:    s.UnitID,
		IsAdmin:   s.IsAdmin,
		CreatedAt: s.CreatedAt,
	}
}
