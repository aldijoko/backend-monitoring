package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"monitoring-cctv-be/pkg/jwt"
	"monitoring-cctv-be/pkg/response"
)

const ContextUserIDKey = "user_id"
const ContextRoleKey = "role"

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "Unauthorized")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")

		claims, err := jwt.ParseToken(secret, token)
		if err != nil {
			response.Unauthorized(c, "Unauthorized")
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextRoleKey, claims.Role)
		c.Next()
	}
}

// RequireRole gates a route to specific roles — a plain allow-list check
// against the JWT's role claim. No DB-backed permission matrix; see
// CLAUDE.md for why this project deliberately doesn't need one.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role, _ := c.Get(ContextRoleKey)
		roleStr, _ := role.(string)
		if !allowed[roleStr] {
			response.Forbidden(c, "Anda tidak memiliki akses untuk aksi ini")
			c.Abort()
			return
		}
		c.Next()
	}
}
