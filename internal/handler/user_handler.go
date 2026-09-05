package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"monitoring-cctv-be/internal/dto"
	"monitoring-cctv-be/internal/middleware"
	"monitoring-cctv-be/internal/repository"
	"monitoring-cctv-be/internal/service"
	"monitoring-cctv-be/pkg/response"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) List(c *gin.Context) {
	p := repository.UserListParams{
		Search: c.Query("search"),
		Role:   c.Query("role"),
		Limit:  atoiOr(c.Query("limit"), 10),
		Offset: atoiOr(c.Query("offset"), 0),
	}
	if v := c.Query("is_active"); v != "" {
		b := v == "true"
		p.IsActive = &b
	}

	res, err := h.users.List(p)
	if err != nil {
		response.InternalError(c, "Gagal memuat data user")
		return
	}
	response.OK(c, 200, res)
}

func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	u, err := h.users.Get(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, 200, u)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data user tidak valid")
		return
	}
	u, err := h.users.Create(req)
	if err != nil {
		if errors.Is(err, service.ErrUsernameTaken) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, "Gagal membuat user")
		return
	}
	response.OK(c, 201, u)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data user tidak valid")
		return
	}
	u, err := h.users.Update(uint(id), req)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, 200, u)
}

func (h *UserHandler) SetActive(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	var req dto.SetActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload tidak valid")
		return
	}
	u, err := h.users.SetActive(uint(id), req.IsActive)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, 200, u)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	requesterID, _ := c.Get(middleware.ContextUserIDKey)
	reqID, _ := requesterID.(uint)

	if err := h.users.Delete(reqID, uint(id)); err != nil {
		if errors.Is(err, service.ErrCannotDeleteSelf) {
			response.BadRequest(c, err.Error())
			return
		}
		response.NotFound(c, err.Error())
		return
	}
	c.Status(204)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	temp, err := h.users.ResetPassword(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, 200, dto.ResetPasswordResponse{TemporaryPassword: temp})
}

func atoiOr(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}
