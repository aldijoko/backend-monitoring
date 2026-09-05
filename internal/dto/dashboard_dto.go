package dto

// Mirrors frontend src/lib/types/dashboard.ts. Note: this app has no
// physical edge hardware reporting CPU/memory/disk/network telemetry (the
// architecture is single-site/centralized — see CLAUDE.md). Those
// MetricSeries are returned with an empty Data slice rather than fabricated
// numbers; only the fields backed by real camera/recording data are
// populated. See camera_service.go / edge_service.go for what's real.

type TimeseriesPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

type MetricSeries struct {
	Name  string            `json:"name"`
	Unit  string            `json:"unit"`
	Color string            `json:"color"`
	Data  []TimeseriesPoint `json:"data"`
}

type EdgeMetrics struct {
	EdgeID        string       `json:"edge_id"`
	EdgeName      string       `json:"edge_name"`
	CPU           MetricSeries `json:"cpu"`
	Memory        MetricSeries `json:"memory"`
	Disk          MetricSeries `json:"disk"`
	NetworkIn     MetricSeries `json:"network_in"`
	NetworkOut    MetricSeries `json:"network_out"`
	UptimePct     float64      `json:"uptime_pct"`
	DaysOnline    int          `json:"days_online"`
	CamerasOnline int          `json:"cameras_online"`
	CamerasTotal  int          `json:"cameras_total"`
}

type CameraHealthBucket struct {
	Range string `json:"range"`
	Count int    `json:"count"`
	Color string `json:"color"`
}

type DashboardSummary struct {
	EdgesTotal           int     `json:"edges_total"`
	EdgesOnline          int     `json:"edges_online"`
	EdgesOffline         int     `json:"edges_offline"`
	CamerasTotal         int     `json:"cameras_total"`
	CamerasOnline        int     `json:"cameras_online"`
	CamerasOffline       int     `json:"cameras_offline"`
	StorageUsedGB        float64 `json:"storage_used_gb"`
	StorageTotalGB       float64 `json:"storage_total_gb"`
	StoragePct           float64 `json:"storage_pct"`
	BandwidthInMbps      float64 `json:"bandwidth_in_mbps"`
	BandwidthOutMbps     float64 `json:"bandwidth_out_mbps"`
	RecordingHoursToday  float64 `json:"recording_hours_today"`
	MotionEventsToday    int     `json:"motion_events_today"`
}

type DashboardData struct {
	Summary      DashboardSummary     `json:"summary"`
	Edges        []EdgeMetrics        `json:"edges"`
	CameraHealth []CameraHealthBucket `json:"camera_health"`
	GeneratedAt  string               `json:"generated_at"`
}
