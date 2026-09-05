package model

import "time"

type EdgeStatus string

const (
	EdgeOnline  EdgeStatus = "online"
	EdgeOffline EdgeStatus = "offline"
	EdgePending EdgeStatus = "pending"
)

type Edge struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Code          string     `gorm:"uniqueIndex;size:32;not null" json:"code"`
	Name          string     `gorm:"size:255;not null" json:"name"`
	Hostname      *string    `gorm:"size:255" json:"hostname,omitempty"`
	IPAddress     *string    `gorm:"size:64" json:"ip_address,omitempty"`
	Location      *string    `gorm:"size:255" json:"location,omitempty"`
	Status        EdgeStatus `gorm:"size:16;not null;default:pending" json:"status"`
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"-"`
}

func (Edge) TableName() string { return "edges" }
