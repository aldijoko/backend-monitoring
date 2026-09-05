package service

import (
	"errors"
	"strings"

	"monitoring-cctv-be/internal/dto"
	"monitoring-cctv-be/internal/model"
	"monitoring-cctv-be/internal/repository"
)

type EdgeService struct {
	edges *repository.EdgeRepository
}

func NewEdgeService(edges *repository.EdgeRepository) *EdgeService {
	return &EdgeService{edges: edges}
}

var ErrEdgeNotFound = errors.New("edge tidak ditemukan")
var ErrEdgeCodeRequired = errors.New("kode edge wajib diisi")
var ErrEdgeCodeTaken = errors.New("kode edge sudah digunakan")

func (s *EdgeService) List(p repository.EdgeListParams) ([]repository.EdgeWithStats, int64, error) {
	return s.edges.List(p)
}

func (s *EdgeService) Get(id uint) (*model.Edge, error) {
	e, err := s.edges.FindByID(id)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, ErrEdgeNotFound
	}
	return e, nil
}

func (s *EdgeService) Create(req dto.EdgeFormPayload) (*model.Edge, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		return nil, ErrEdgeCodeRequired
	}

	existing, err := s.edges.FindByCode(code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEdgeCodeTaken
	}

	e := &model.Edge{
		Code:   code,
		Name:   req.Name,
		Status: model.EdgePending,
	}
	if req.Hostname != "" {
		e.Hostname = &req.Hostname
	}
	if req.IPAddress != "" {
		e.IPAddress = &req.IPAddress
	}
	if err := s.edges.Create(e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *EdgeService) Update(id uint, req dto.EdgeFormPayload) (*model.Edge, error) {
	e, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	e.Name = req.Name
	if req.Hostname != "" {
		e.Hostname = &req.Hostname
	} else {
		e.Hostname = nil
	}
	if req.IPAddress != "" {
		e.IPAddress = &req.IPAddress
	} else {
		e.IPAddress = nil
	}
	if err := s.edges.Update(e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *EdgeService) Delete(id uint) error {
	if _, err := s.Get(id); err != nil {
		return err
	}
	return s.edges.Delete(id)
}
