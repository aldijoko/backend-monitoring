package handler

import (
	"github.com/gin-gonic/gin"

	"monitoring-cctv-be/internal/service"
	"monitoring-cctv-be/pkg/response"
)

type DashboardHandler struct {
	dashboard *service.DashboardService
}

func NewDashboardHandler(dashboard *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboard: dashboard}
}

func (h *DashboardHandler) Summary(c *gin.Context) {
	data, err := h.dashboard.Summary()
	if err != nil {
		response.InternalError(c, "Gagal memuat dashboard")
		return
	}
	response.OK(c, 200, data)
}
