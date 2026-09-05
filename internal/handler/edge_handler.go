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

type EdgeHandler struct {
	edges *service.EdgeService
}

func NewEdgeHandler(edges *service.EdgeService) *EdgeHandler {
	return &EdgeHandler{edges: edges}
}

func (h *EdgeHandler) List(c *gin.Context) {
	limit := atoiOr(c.Query("limit"), 10)
	offset := atoiOr(c.Query("offset"), 0)

	items, total, err := h.edges.List(repository.EdgeListParams{
		Search: c.Query("search"),
		Status: c.Query("status"),
		Limit:  limit,
		Offset: offset,
		Sort:   c.Query("sort"),
		Order:  c.Query("order"),
	})
	if err != nil {
		response.InternalError(c, "Gagal memuat data edge")
		return
	}

	response.OK(c, 200, gin.H{"items": items, "total": total, "limit": limit, "offset": offset})
}

func (h *EdgeHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	e, err := h.edges.Get(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, 200, e)
}

func (h *EdgeHandler) Create(c *gin.Context) {
	var req dto.EdgeFormPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data edge tidak valid")
		return
	}
	e, err := h.edges.Create(req)
	if err != nil {
		if errors.Is(err, service.ErrEdgeCodeRequired) {
			response.BadRequest(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrEdgeCodeTaken) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, "Gagal membuat edge")
		return
	}
	response.OK(c, 201, e)
}

func (h *EdgeHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	var req dto.EdgeFormPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data edge tidak valid")
		return
	}
	e, err := h.edges.Update(uint(id), req)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, 200, e)
}

func (h *EdgeHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	if err := h.edges.Delete(uint(id)); err != nil {
		response.NotFound(c, err.Error())
		return
	}
	c.Status(204)
}
