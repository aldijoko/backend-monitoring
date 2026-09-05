package dto

import (
	"time"

	"monitoring-cctv-be/internal/model"
)

// CameraResponse. SourceURL is only populated for the admin edit-form detail
// endpoint (GetCamera), never for the list endpoint that /cameras and /live
// both consume — see camera_service.go.
type CameraResponse struct {
	ID             uint                `json:"id"`
	EdgeID         uint                `json:"edge_id"`
	EdgeCode       string              `json:"edge_code"`
	EdgeName       string              `json:"edge_name"`
	Name           string              `json:"name"`
	Channel        int                 `json:"channel"`
	SourceURL      *string             `json:"source_url,omitempty"`
	StreamURL      string              `json:"stream_url"`
	RtspTransport  model.RtspTransport `json:"rtsp_transport"`
	SourceOnDemand bool                `json:"source_on_demand"`
	StorageDays    int                 `json:"storage_days"`
	Status         model.CameraStatus  `json:"status"`
	Enabled        bool                `json:"enabled"`
	LastSeenAt     *time.Time          `json:"last_seen_at"`
	CreatedAt      time.Time           `json:"created_at"`
}

type CameraListResponse struct {
	Cameras []CameraResponse `json:"cameras"`
	Total   int              `json:"total"`
}

type CameraFormPayload struct {
	EdgeID         uint                `json:"edge_id" binding:"required"`
	Name           string              `json:"name" binding:"required"`
	Channel        int                 `json:"channel"`
	SourceURL      string              `json:"source_url" binding:"required"`
	RtspTransport  model.RtspTransport `json:"rtsp_transport"`
	SourceOnDemand bool                `json:"source_on_demand"`
	StorageDays    int                 `json:"storage_days"`
	Enabled        bool                `json:"enabled"`
}

// CameraPatchPayload mirrors the frontend's `updateCamera(id, patch:
// Partial<LiveCamera>)` — used both for full-form edits and for the
// lightweight "toggle enabled" action on the /cameras list, so every field
// must be optional and only applied when present.
type CameraPatchPayload struct {
	EdgeID         *uint                `json:"edge_id"`
	Name           *string              `json:"name"`
	Channel        *int                 `json:"channel"`
	SourceURL      *string              `json:"source_url"`
	RtspTransport  *model.RtspTransport `json:"rtsp_transport"`
	SourceOnDemand *bool                `json:"source_on_demand"`
	StorageDays    *int                 `json:"storage_days"`
	Enabled        *bool                `json:"enabled"`
}
