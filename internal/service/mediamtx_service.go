package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// MediaMTXService talks to mediamtx's runtime config API to register one
// path per camera — mediamtx itself handles the RTSP pull, reconnect, and
// HLS transcode, so this is the only piece of glue code needed. See
// docs/mediamtx.md (or CLAUDE.md) for why this replaces manual ffmpeg
// process management.
type MediaMTXService struct {
	apiURL     string
	hlsBaseURL string
	client     *http.Client
}

func NewMediaMTXService(apiURL, hlsBaseURL string) *MediaMTXService {
	return &MediaMTXService{apiURL: apiURL, hlsBaseURL: hlsBaseURL, client: &http.Client{}}
}

func pathName(cameraID uint) string {
	return fmt.Sprintf("camera-%d", cameraID)
}

func (m *MediaMTXService) streamURL(cameraID uint) string {
	return fmt.Sprintf("%s/%s/index.m3u8", m.hlsBaseURL, pathName(cameraID))
}

func (m *MediaMTXService) do(method, path string, body any) error {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return err
		}
	}
	req, err := http.NewRequest(method, m.apiURL+path, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		return fmt.Errorf("mediamtx API %s %s failed: %s", method, path, res.Status)
	}
	return nil
}

func pathBody(cfg PathConfig) map[string]any {
	return map[string]any{
		"source":         cfg.Source,
		"rtspTransport":  cfg.RtspTransport,
		"sourceOnDemand": cfg.SourceOnDemand,
	}
}

func (m *MediaMTXService) RegisterPath(cameraID uint, cfg PathConfig) (string, error) {
	name := pathName(cameraID)
	if err := m.do(http.MethodPost, "/v3/config/paths/add/"+name, pathBody(cfg)); err != nil {
		return "", err
	}
	return m.streamURL(cameraID), nil
}

func (m *MediaMTXService) UpdatePath(cameraID uint, cfg PathConfig) (string, error) {
	name := pathName(cameraID)
	if err := m.do(http.MethodPatch, "/v3/config/paths/patch/"+name, pathBody(cfg)); err != nil {
		// path may not exist yet (e.g. created before mediamtx was wired up) — fall back to add
		if err2 := m.do(http.MethodPost, "/v3/config/paths/add/"+name, pathBody(cfg)); err2 != nil {
			return "", err
		}
	}
	return m.streamURL(cameraID), nil
}

func (m *MediaMTXService) RemovePath(cameraID uint) error {
	name := pathName(cameraID)
	return m.do(http.MethodDelete, "/v3/config/paths/delete/"+name, nil)
}

// PathStatus reports whether mediamtx currently has an active source for
// the given camera path — used by the status poller to derive camera/edge
// online state (see camera_status_poller.go).
type PathStatus struct {
	Ready  bool
	Exists bool
}

func (m *MediaMTXService) GetPathStatus(cameraID uint) (PathStatus, error) {
	name := pathName(cameraID)
	req, err := http.NewRequest(http.MethodGet, m.apiURL+"/v3/paths/get/"+name, nil)
	if err != nil {
		return PathStatus{}, err
	}
	res, err := m.client.Do(req)
	if err != nil {
		return PathStatus{}, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return PathStatus{Exists: false}, nil
	}
	if res.StatusCode >= 300 {
		return PathStatus{}, fmt.Errorf("mediamtx API get path failed: %s", res.Status)
	}

	var payload struct {
		Ready bool `json:"ready"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return PathStatus{}, err
	}
	return PathStatus{Exists: true, Ready: payload.Ready}, nil
}
