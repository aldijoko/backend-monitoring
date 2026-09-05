package dto

import "monitoring-cctv-be/internal/model"

type RecordingListResponse struct {
	Items  []model.Recording `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

type ArchiveBulkRequest struct {
	IDs    []uint `json:"ids" binding:"required"`
	Format string `json:"format"`
}

type ArchiveBulkResponse struct {
	ArchiveURL string `json:"archive_url"`
	SizeBytes  int64  `json:"size_bytes"`
	ExpiresAt  string `json:"expires_at"`
	ItemCount  int    `json:"item_count"`
}
