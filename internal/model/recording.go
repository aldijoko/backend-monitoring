package model

import "time"

type Recording struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	EdgeID          uint      `gorm:"not null;index" json:"edge_id"`
	EdgeCode        string    `gorm:"size:32" json:"edge_code"`
	CameraID        uint      `gorm:"not null;index" json:"camera_id"`
	CameraName      string    `gorm:"size:255" json:"camera_name"`
	Filename        string    `gorm:"size:255;not null" json:"filename"`
	StartedAt       time.Time `json:"started_at"`
	EndedAt         time.Time `json:"ended_at"`
	SizeBytes       int64     `json:"size_bytes"`
	StoragePath     *string   `gorm:"size:512" json:"storage_path,omitempty"`
	DurationSeconds *int      `json:"duration_seconds,omitempty"`
	ThumbnailURL    *string   `gorm:"size:512" json:"thumbnail_url,omitempty"`
	PlaybackURL     *string   `gorm:"size:512" json:"playback_url,omitempty"`
	CreatedAt       time.Time `json:"-"`
}

func (Recording) TableName() string { return "recordings" }
