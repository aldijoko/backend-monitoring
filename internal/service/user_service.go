package service

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"strings"

	"monitoring-cctv-be/internal/dto"
	"monitoring-cctv-be/internal/model"
	"monitoring-cctv-be/internal/repository"
)

type UserService struct {
	users *repository.UserRepository
}

func NewUserService(users *repository.UserRepository) *UserService {
	return &UserService{users: users}
}

var ErrUsernameTaken = errors.New("username sudah digunakan")
var ErrUserNotFound = errors.New("user tidak ditemukan")
var ErrCannotDeleteSelf = errors.New("tidak dapat menghapus akun sendiri")

func (s *UserService) List(p repository.UserListParams) (*dto.UserListResponse, error) {
	users, total, err := s.users.List(p)
	if err != nil {
		return nil, err
	}
	limit := p.Limit
	if limit <= 0 {
		limit = 10
	}
	return &dto.UserListResponse{Items: users, Total: total, Limit: limit, Offset: p.Offset}, nil
}

func (s *UserService) Get(id uint) (*model.User, error) {
	u, err := s.users.FindByID(id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *UserService) Create(req dto.CreateUserRequest) (*model.User, error) {
	existing, err := s.users.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUsernameTaken
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	u := &model.User{
		Username:     req.Username,
		Email:        &req.Email,
		PasswordHash: hash,
		Role:         req.Role,
		IsActive:     isActive,
	}
	if err := s.users.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Update(id uint, req dto.UpdateUserRequest) (*model.User, error) {
	u, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	u.Email = &req.Email
	u.Role = req.Role
	u.IsActive = req.IsActive
	if err := s.users.Update(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) SetActive(id uint, isActive bool) (*model.User, error) {
	u, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	u.IsActive = isActive
	if err := s.users.Update(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Delete(requesterID, id uint) error {
	if requesterID == id {
		return ErrCannotDeleteSelf
	}
	u, err := s.Get(id)
	if err != nil {
		return err
	}
	return s.users.Delete(u.ID)
}

func (s *UserService) ResetPassword(id uint) (string, error) {
	u, err := s.Get(id)
	if err != nil {
		return "", err
	}
	temp := generateTempPassword()
	hash, err := HashPassword(temp)
	if err != nil {
		return "", err
	}
	u.PasswordHash = hash
	if err := s.users.Update(u); err != nil {
		return "", err
	}
	return temp, nil
}

func generateTempPassword() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "CCTV-" + strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b))[:8]
}
