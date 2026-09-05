package service

import (
	"errors"
	"log"

	"monitoring-cctv-be/internal/dto"
	"monitoring-cctv-be/internal/model"
	"monitoring-cctv-be/internal/repository"
)

type CameraService struct {
	cameras *repository.CameraRepository
	edges   *repository.EdgeRepository
	media   MediaProvider
}

func NewCameraService(cameras *repository.CameraRepository, edges *repository.EdgeRepository, media MediaProvider) *CameraService {
	return &CameraService{cameras: cameras, edges: edges, media: media}
}

var ErrCameraNotFound = errors.New("kamera tidak ditemukan")

func (s *CameraService) toResponse(c model.Camera, edgeByID map[uint]model.Edge, includeSource bool) dto.CameraResponse {
	edge := edgeByID[c.EdgeID]
	resp := dto.CameraResponse{
		ID:             c.ID,
		EdgeID:         c.EdgeID,
		EdgeCode:       edge.Code,
		EdgeName:       edge.Name,
		Name:           c.Name,
		Channel:        c.Channel,
		StreamURL:      c.StreamURL,
		RtspTransport:  c.RtspTransport,
		SourceOnDemand: c.SourceOnDemand,
		StorageDays:    c.StorageDays,
		Status:         c.Status,
		Enabled:        c.Enabled,
		LastSeenAt:     c.LastSeenAt,
		CreatedAt:      c.CreatedAt,
	}
	if includeSource {
		src := c.SourceURL
		resp.SourceURL = &src
	}
	return resp
}

// List backs both /cameras and /live — never includes source_url (RTSP
// credentials), only the computed HLS stream_url.
func (s *CameraService) List(f repository.CameraFilters) (*dto.CameraListResponse, error) {
	cameras, total, err := s.cameras.List(f)
	if err != nil {
		return nil, err
	}

	edgeIDs := make([]uint, 0, len(cameras))
	seen := map[uint]bool{}
	for _, c := range cameras {
		if !seen[c.EdgeID] {
			edgeIDs = append(edgeIDs, c.EdgeID)
			seen[c.EdgeID] = true
		}
	}
	edgeByID, err := s.cameras.EdgeNamesByIDs(edgeIDs)
	if err != nil {
		return nil, err
	}

	items := make([]dto.CameraResponse, 0, len(cameras))
	for _, c := range cameras {
		items = append(items, s.toResponse(c, edgeByID, false))
	}

	return &dto.CameraListResponse{Cameras: items, Total: int(total)}, nil
}

// Get backs the admin edit-form detail endpoint — includes source_url so
// the form can prefill the RTSP source for editing. Callers must restrict
// this to admin/superadmin via RBAC middleware.
func (s *CameraService) Get(id uint) (*dto.CameraResponse, error) {
	c, err := s.cameras.FindByID(id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCameraNotFound
	}
	edgeByID, err := s.cameras.EdgeNamesByIDs([]uint{c.EdgeID})
	if err != nil {
		return nil, err
	}
	resp := s.toResponse(*c, edgeByID, true)
	return &resp, nil
}

func (s *CameraService) Create(req dto.CameraFormPayload) (*dto.CameraResponse, error) {
	if _, err := s.edges.FindByID(req.EdgeID); err != nil {
		return nil, err
	}

	transport := req.RtspTransport
	if transport == "" {
		transport = model.RtspTransportAutomatic
	}
	storageDays := req.StorageDays
	if storageDays <= 0 {
		storageDays = 30
	}
	channel := req.Channel
	if channel <= 0 {
		channel = 1
	}

	c := &model.Camera{
		EdgeID:         req.EdgeID,
		Name:           req.Name,
		Channel:        channel,
		SourceURL:      req.SourceURL,
		RtspTransport:  transport,
		SourceOnDemand: req.SourceOnDemand,
		StorageDays:    storageDays,
		Status:         model.CameraOffline,
		Enabled:        req.Enabled,
	}
	if err := s.cameras.Create(c); err != nil {
		return nil, err
	}

	streamURL, err := s.media.RegisterPath(c.ID, PathConfig{
		Source:         c.SourceURL,
		RtspTransport:  string(c.RtspTransport),
		SourceOnDemand: c.SourceOnDemand,
	})
	if err == nil && streamURL != "" {
		c.StreamURL = streamURL
		_ = s.cameras.Update(c)
	}

	edgeByID, err := s.cameras.EdgeNamesByIDs([]uint{c.EdgeID})
	if err != nil {
		return nil, err
	}
	resp := s.toResponse(*c, edgeByID, false)
	return &resp, nil
}

// Patch applies only the fields present in the request — it backs the
// single PATCH /cameras/:id endpoint the frontend uses both for full-form
// edits and for the lightweight "toggle enabled" action.
func (s *CameraService) Patch(id uint, req dto.CameraPatchPayload) (*dto.CameraResponse, error) {
	c, err := s.cameras.FindByID(id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCameraNotFound
	}

	pathConfigChanged := false

	if req.EdgeID != nil {
		c.EdgeID = *req.EdgeID
	}
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Channel != nil {
		c.Channel = *req.Channel
	}
	if req.SourceURL != nil && *req.SourceURL != c.SourceURL {
		c.SourceURL = *req.SourceURL
		pathConfigChanged = true
	}
	if req.RtspTransport != nil && *req.RtspTransport != c.RtspTransport {
		c.RtspTransport = *req.RtspTransport
		pathConfigChanged = true
	}
	if req.SourceOnDemand != nil && *req.SourceOnDemand != c.SourceOnDemand {
		c.SourceOnDemand = *req.SourceOnDemand
		pathConfigChanged = true
	}
	if req.StorageDays != nil {
		c.StorageDays = *req.StorageDays
	}
	if req.Enabled != nil {
		c.Enabled = *req.Enabled
	}

	if err := s.cameras.Update(c); err != nil {
		return nil, err
	}

	if pathConfigChanged {
		streamURL, err := s.media.UpdatePath(c.ID, PathConfig{
			Source:         c.SourceURL,
			RtspTransport:  string(c.RtspTransport),
			SourceOnDemand: c.SourceOnDemand,
		})
		if err == nil && streamURL != "" {
			c.StreamURL = streamURL
			_ = s.cameras.Update(c)
		}
	}

	edgeByID, err := s.cameras.EdgeNamesByIDs([]uint{c.EdgeID})
	if err != nil {
		return nil, err
	}
	resp := s.toResponse(*c, edgeByID, false)
	return &resp, nil
}

func (s *CameraService) Delete(id uint) error {
	c, err := s.cameras.FindByID(id)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCameraNotFound
	}
	_ = s.media.RemovePath(id)
	return s.cameras.Delete(id)
}

// ReconcileMediaPaths re-registers every enabled camera's path with
// mediamtx. mediamtx does not persist paths registered via its config API —
// they live in-memory only — so a mediamtx restart forgets every camera
// path even though the DB still has the right source_url, leaving cameras
// stuck offline until something happens to PATCH them. Call this once at
// API startup, before the status poller's first tick, so a mediamtx
// restart during dev doesn't require manually re-saving every camera.
// Per-camera failures (e.g. mediamtx unreachable) are logged and skipped —
// not fatal, since the status poller will keep reporting them offline.
func (s *CameraService) ReconcileMediaPaths() {
	cameras, _, err := s.cameras.List(repository.CameraFilters{})
	if err != nil {
		log.Printf("reconcile media paths: list cameras: %v", err)
		return
	}
	for i := range cameras {
		c := &cameras[i]
		if !c.Enabled {
			continue
		}
		streamURL, err := s.media.UpdatePath(c.ID, PathConfig{
			Source:         c.SourceURL,
			RtspTransport:  string(c.RtspTransport),
			SourceOnDemand: c.SourceOnDemand,
		})
		if err != nil {
			log.Printf("reconcile media paths: camera %d: %v", c.ID, err)
			continue
		}
		if streamURL != "" && streamURL != c.StreamURL {
			c.StreamURL = streamURL
			if err := s.cameras.Update(c); err != nil {
				log.Printf("reconcile media paths: camera %d: update stream_url: %v", c.ID, err)
			}
		}
	}
}
