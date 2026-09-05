package service

import (
	"log"
	"time"

	"monitoring-cctv-be/internal/model"
	"monitoring-cctv-be/internal/repository"
)

// CameraStatusPoller periodically asks mediamtx whether each enabled
// camera's path is actually receiving a stream, and writes that back to
// cameras.status/last_seen_at — then derives edges.status/last_heartbeat
// from the aggregate (see EdgeRepository.RecalculateStatus). This is the
// only source of "online/offline" truth; nothing else updates camera status.
type CameraStatusPoller struct {
	cameras *repository.CameraRepository
	edges   *repository.EdgeRepository
	media   *MediaMTXService
	every   time.Duration
}

func NewCameraStatusPoller(cameras *repository.CameraRepository, edges *repository.EdgeRepository, media *MediaMTXService, every time.Duration) *CameraStatusPoller {
	return &CameraStatusPoller{cameras: cameras, edges: edges, media: media, every: every}
}

func (p *CameraStatusPoller) Run(stop <-chan struct{}) {
	ticker := time.NewTicker(p.every)
	defer ticker.Stop()

	p.tick()
	for {
		select {
		case <-ticker.C:
			p.tick()
		case <-stop:
			return
		}
	}
}

func (p *CameraStatusPoller) tick() {
	cameras, _, err := p.cameras.List(repository.CameraFilters{})
	if err != nil {
		log.Printf("status poller: list cameras failed: %v", err)
		return
	}

	touchedEdges := map[uint]bool{}

	for _, cam := range cameras {
		if !cam.Enabled {
			continue
		}

		status, err := p.media.GetPathStatus(cam.ID)
		if err != nil {
			log.Printf("status poller: camera %d: %v", cam.ID, err)
			continue
		}

		newStatus := model.CameraOffline
		var lastSeen *string
		if status.Exists && status.Ready {
			newStatus = model.CameraOnline
			now := time.Now().Format(time.RFC3339)
			lastSeen = &now
		}

		if newStatus != cam.Status {
			if err := p.cameras.UpdateStatus(cam.ID, newStatus, "", lastSeen); err != nil {
				log.Printf("status poller: update camera %d failed: %v", cam.ID, err)
				continue
			}
		} else if lastSeen != nil {
			_ = p.cameras.UpdateStatus(cam.ID, newStatus, "", lastSeen)
		}

		touchedEdges[cam.EdgeID] = true
	}

	for edgeID := range touchedEdges {
		if err := p.edges.RecalculateStatus(edgeID); err != nil {
			log.Printf("status poller: recalc edge %d failed: %v", edgeID, err)
		}
	}
}
