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

type CameraHandler struct {
	cameras *service.CameraService
}

func NewCameraHandler(cameras *service.CameraService) *CameraHandler {
	return &CameraHandler{cameras: cameras}
}

func (h *CameraHandler) List(c *gin.Context) {
	f := repository.CameraFilters{
		Status: c.Query("status"),
		Q:      c.Query("q"),
		// Unset (0) means unlimited — callers like /live and the edge detail
		// page depend on getting every matching camera back.
		Limit:  atoiOr(c.Query("limit"), 0),
		Offset: atoiOr(c.Query("offset"), 0),
	}
	if v := c.Query("edge_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			eid := uint(id)
			f.EdgeID = &eid
		}
	}

	res, err := h.cameras.List(f)
	if err != nil {
		response.InternalError(c, "Gagal memuat data kamera")
		return
	}
	response.OK(c, 200, res)
}

func (h *CameraHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	cam, err := h.cameras.Get(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, 200, cam)
}

func (h *CameraHandler) Create(c *gin.Context) {
	var req dto.CameraFormPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data kamera tidak valid")
		return
	}
	cam, err := h.cameras.Create(req)
	if err != nil {
		response.InternalError(c, "Gagal membuat kamera")
		return
	}
	response.OK(c, 201, cam)
}

func (h *CameraHandler) Patch(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	var req dto.CameraPatchPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data kamera tidak valid")
		return
	}
	cam, err := h.cameras.Patch(uint(id), req)
	if err != nil {
		if errors.Is(err, service.ErrCameraNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "Gagal memperbarui kamera")
		return
	}
	response.OK(c, 200, cam)
}

func (h *CameraHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	if err := h.cameras.Delete(uint(id)); err != nil {
		response.NotFound(c, err.Error())
		return
	}
	c.Status(204)
}
