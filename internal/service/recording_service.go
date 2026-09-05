package service

import (
	"errors"
	"fmt"
	"time"

	"monitoring-cctv-be/internal/dto"
	"monitoring-cctv-be/internal/model"
	"monitoring-cctv-be/internal/repository"
)

type RecordingService struct {
	recordings *repository.RecordingRepository
}

func NewRecordingService(recordings *repository.RecordingRepository) *RecordingService {
	return &RecordingService{recordings: recordings}
}

var ErrRecordingNotFound = errors.New("recording tidak ditemukan")
var ErrNoRecordingsSelected = errors.New("tidak ada recording yang dipilih")

func (s *RecordingService) List(p repository.RecordingListParams) (*dto.RecordingListResponse, error) {
	items, total, err := s.recordings.List(p)
	if err != nil {
		return nil, err
	}
	limit := p.Limit
	if limit <= 0 {
		limit = 10
	}
	return &dto.RecordingListResponse{Items: items, Total: total, Limit: limit, Offset: p.Offset}, nil
}

func (s *RecordingService) Get(id uint) (*model.Recording, error) {
	r, err := s.recordings.FindByID(id)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, ErrRecordingNotFound
	}
	return r, nil
}

func (s *RecordingService) Cameras() ([]repository.CameraOption, error) {
	return s.recordings.DistinctCameras()
}

func (s *RecordingService) ArchiveBulk(ids []uint, format string) (*dto.ArchiveBulkResponse, error) {
	items, err := s.recordings.FindByIDs(ids)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrNoRecordingsSelected
	}
	if format == "" {
		format = "mp4"
	}
	var totalSize int64
	for _, r := range items {
		totalSize += r.SizeBytes
	}
	return &dto.ArchiveBulkResponse{
		ArchiveURL: fmt.Sprintf("/api/v1/recordings/archive-bulk/download?token=%d&format=%s", time.Now().UnixNano(), format),
		SizeBytes:  totalSize,
		ExpiresAt:  time.Now().Add(time.Hour).Format(time.RFC3339),
		ItemCount:  len(items),
	}, nil
}
