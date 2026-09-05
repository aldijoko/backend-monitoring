package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"monitoring-cctv-be/internal/dto"
	"monitoring-cctv-be/internal/service"
	"monitoring-cctv-be/pkg/response"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Username dan password wajib diisi")
		return
	}

	res, err := h.auth.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrUserInactive) {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, "Gagal login")
		return
	}

	response.OK(c, 200, res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.Status(204)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "refresh_token wajib diisi")
		return
	}

	res, err := h.auth.Refresh(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "Refresh token tidak valid")
		return
	}

	response.OK(c, 200, res)
}
