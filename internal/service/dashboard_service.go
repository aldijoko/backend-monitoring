package service

import (
	"time"

	"gorm.io/gorm"

	"monitoring-cctv-be/internal/dto"
	"monitoring-cctv-be/internal/model"
)

type DashboardService struct {
	db *gorm.DB
}

func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

// Summary returns only aggregates this system actually has data for. There
// is no physical edge hardware reporting bandwidth/CPU/motion events in this
// centralized architecture, so those fields are honestly zeroed rather than
// fabricated — see dto/dashboard_dto.go.
func (s *DashboardService) Summary() (*dto.DashboardData, error) {
	var edges []model.Edge
	if err := s.db.Find(&edges).Error; err != nil {
		return nil, err
	}

	edgesOnline := 0
	for _, e := range edges {
		if e.Status == model.EdgeOnline {
			edgesOnline++
		}
	}

	var camerasTotal, camerasOnline int64
	s.db.Model(&model.Camera{}).Count(&camerasTotal)
	s.db.Model(&model.Camera{}).Where("status IN ?", []string{"online", "recording"}).Count(&camerasOnline)

	var storageUsedBytes int64
	s.db.Model(&model.Recording{}).Select("COALESCE(SUM(size_bytes), 0)").Scan(&storageUsedBytes)

	todayStart := time.Now().Truncate(24 * time.Hour)
	var recordingSecondsToday int64
	s.db.Model(&model.Recording{}).
		Where("started_at >= ?", todayStart).
		Select("COALESCE(SUM(duration_seconds), 0)").Scan(&recordingSecondsToday)

	edgeMetrics := make([]dto.EdgeMetrics, 0, len(edges))
	for _, e := range edges {
		var total, online int64
		s.db.Model(&model.Camera{}).Where("edge_id = ?", e.ID).Count(&total)
		s.db.Model(&model.Camera{}).Where("edge_id = ? AND status IN ?", e.ID, []string{"online", "recording"}).Count(&online)

		uptimePct := 0.0
		if e.Status == model.EdgeOnline {
			uptimePct = 100
		}

		empty := dto.MetricSeries{Data: []dto.TimeseriesPoint{}}
		edgeMetrics = append(edgeMetrics, dto.EdgeMetrics{
			EdgeID:        e.Code,
			EdgeName:      e.Name,
			CPU:           empty,
			Memory:        empty,
			Disk:          empty,
			NetworkIn:     empty,
			NetworkOut:    empty,
			UptimePct:     uptimePct,
			DaysOnline:    0,
			CamerasOnline: int(online),
			CamerasTotal:  int(total),
		})
	}

	storageUsedGB := float64(storageUsedBytes) / 1e9

	return &dto.DashboardData{
		Summary: dto.DashboardSummary{
			EdgesTotal:          len(edges),
			EdgesOnline:         edgesOnline,
			EdgesOffline:        len(edges) - edgesOnline,
			CamerasTotal:        int(camerasTotal),
			CamerasOnline:       int(camerasOnline),
			CamerasOffline:      int(camerasTotal) - int(camerasOnline),
			StorageUsedGB:       storageUsedGB,
			StorageTotalGB:      0,
			StoragePct:          0,
			BandwidthInMbps:     0,
			BandwidthOutMbps:    0,
			RecordingHoursToday: float64(recordingSecondsToday) / 3600,
			MotionEventsToday:   0,
		},
		Edges:        edgeMetrics,
		CameraHealth: []dto.CameraHealthBucket{},
		GeneratedAt:  time.Now().Format(time.RFC3339),
	}, nil
}
