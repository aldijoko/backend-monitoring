package service

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"monitoring-cctv-be/internal/dto"
	"monitoring-cctv-be/internal/model"
	"monitoring-cctv-be/internal/repository"
	"monitoring-cctv-be/pkg/jwt"
)

type AuthService struct {
	users            *repository.UserRepository
	jwtSecret        string
	accessTTL        time.Duration
	refreshTTL       time.Duration
}

func NewAuthService(users *repository.UserRepository, jwtSecret string, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

var ErrInvalidCredentials = errors.New("username atau password salah")
var ErrUserInactive = errors.New("akun tidak aktif, hubungi superadmin")

func (s *AuthService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.users.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrUserInactive
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	access, err := jwt.GenerateToken(s.jwtSecret, user.ID, string(user.Role), s.accessTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := jwt.GenerateToken(s.jwtSecret, user.ID, string(user.Role), s.refreshTTL)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         *user,
	}, nil
}

func (s *AuthService) Refresh(refreshToken string) (*dto.RefreshResponse, error) {
	claims, err := jwt.ParseToken(s.jwtSecret, refreshToken)
	if err != nil {
		return nil, errors.New("refresh token tidak valid")
	}

	user, err := s.users.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, ErrUserInactive
	}

	access, err := jwt.GenerateToken(s.jwtSecret, user.ID, string(user.Role), s.accessTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := jwt.GenerateToken(s.jwtSecret, user.ID, string(user.Role), s.refreshTTL)
	if err != nil {
		return nil, err
	}

	return &dto.RefreshResponse{AccessToken: access, RefreshToken: refresh}, nil
}

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func EnsureBootstrapAdmin(users *repository.UserRepository, username, email, password string) error {
	existing, err := users.FindByUsername(username)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	e := email
	return users.Create(&model.User{
		Username:     username,
		Email:        &e,
		PasswordHash: hash,
		Role:         model.RoleSuperadmin,
		IsActive:     true,
	})
}
