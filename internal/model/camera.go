package model

import "time"

type CameraStatus string

const (
	CameraOnline    CameraStatus = "online"
	CameraOffline   CameraStatus = "offline"
	CameraRecording CameraStatus = "recording"
	CameraError     CameraStatus = "error"
)

type RtspTransport string

const (
	RtspTransportAutomatic RtspTransport = "automatic"
	RtspTransportTCP       RtspTransport = "tcp"
	RtspTransportUDP       RtspTransport = "udp"
)

// Camera. SourceURL holds the RTSP source (with embedded credentials) that
// mediamtx pulls from — it must never reach the public live-view response,
// only the admin edit-form detail endpoint. StreamURL is the HLS output
// mediamtx serves, and is what the frontend actually plays.
type Camera struct {
	ID             uint          `gorm:"primaryKey" json:"id"`
	EdgeID         uint          `gorm:"not null;index" json:"edge_id"`
	Name           string        `gorm:"size:255;not null" json:"name"`
	Channel        int           `gorm:"not null;default:1" json:"channel"`
	SourceURL      string        `gorm:"size:512;not null" json:"-"`
	StreamURL      string        `gorm:"size:512" json:"stream_url"`
	RtspTransport  RtspTransport `gorm:"size:16;not null;default:automatic" json:"rtsp_transport"`
	SourceOnDemand bool          `gorm:"not null;default:false" json:"source_on_demand"`
	StorageDays    int           `gorm:"not null;default:30" json:"storage_days"`
	Status         CameraStatus  `gorm:"size:16;not null;default:offline" json:"status"`
	Enabled        bool          `gorm:"not null;default:true" json:"enabled"`
	LastSeenAt     *time.Time    `json:"last_seen_at"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"-"`
}

func (Camera) TableName() string { return "cameras" }
