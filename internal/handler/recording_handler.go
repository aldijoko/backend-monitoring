package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"monitoring-cctv-be/internal/dto"
	"monitoring-cctv-be/internal/repository"
	"monitoring-cctv-be/internal/service"
	"monitoring-cctv-be/pkg/response"
)

type RecordingHandler struct {
	recordings *service.RecordingService
}

func NewRecordingHandler(recordings *service.RecordingService) *RecordingHandler {
	return &RecordingHandler{recordings: recordings}
}

func (h *RecordingHandler) List(c *gin.Context) {
	p := repository.RecordingListParams{
		Search: c.Query("search"),
		Limit:  atoiOr(c.Query("limit"), 10),
		Offset: atoiOr(c.Query("offset"), 0),
		Sort:   c.Query("sort"),
		Order:  c.Query("order"),
	}
	if v := c.Query("edge_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			eid := uint(id)
			p.EdgeID = &eid
		}
	}
	if v := c.Query("camera_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			cid := uint(id)
			p.CameraID = &cid
		}
	}
	if v := c.Query("date_from"); v != "" {
		p.DateFrom = &v
	}
	if v := c.Query("date_to"); v != "" {
		p.DateTo = &v
	}

	res, err := h.recordings.List(p)
	if err != nil {
		response.InternalError(c, "Gagal memuat data recording")
		return
	}
	response.OK(c, 200, res)
}

func (h *RecordingHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	rec, err := h.recordings.Get(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, 200, rec)
}

func (h *RecordingHandler) Cameras(c *gin.Context) {
	opts, err := h.recordings.Cameras()
	if err != nil {
		response.InternalError(c, "Gagal memuat daftar kamera")
		return
	}
	response.OK(c, 200, opts)
}

func (h *RecordingHandler) ArchiveBulk(c *gin.Context) {
	var req dto.ArchiveBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload tidak valid")
		return
	}
	res, err := h.recordings.ArchiveBulk(req.IDs, req.Format)
	if err != nil {
		if errors.Is(err, service.ErrNoRecordingsSelected) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, "Gagal membuat arsip")
		return
	}
	response.OK(c, 200, res)
}
