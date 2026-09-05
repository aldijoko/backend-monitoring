package service

// PathConfig carries the subset of a camera's connection settings that
// actually map to mediamtx per-path config (see pathDefaults in
// mediamtx.yml) — as opposed to fields like resolution/fps/codec that
// mediamtx has no way to apply since it only remuxes, never transcodes.
type PathConfig struct {
	Source         string
	RtspTransport  string
	SourceOnDemand bool
}

// MediaProvider abstracts the media-plane (mediamtx) that actually pulls
// RTSP and serves HLS — camera_service.go depends on this interface only,
// so it works standalone before mediamtx is wired up (see NoopMediaProvider)
// and against the real thing once MediaMTXService (Phase 4) is injected.
type MediaProvider interface {
	RegisterPath(cameraID uint, cfg PathConfig) (streamURL string, err error)
	UpdatePath(cameraID uint, cfg PathConfig) (streamURL string, err error)
	RemovePath(cameraID uint) error
}

// NoopMediaProvider lets camera CRUD work before mediamtx is deployed — it
// returns an empty stream_url so the frontend list still renders, just
// without a playable stream yet.
type NoopMediaProvider struct{}

func (NoopMediaProvider) RegisterPath(_ uint, _ PathConfig) (string, error) { return "", nil }
func (NoopMediaProvider) UpdatePath(_ uint, _ PathConfig) (string, error)   { return "", nil }
func (NoopMediaProvider) RemovePath(_ uint) error                           { return nil }
