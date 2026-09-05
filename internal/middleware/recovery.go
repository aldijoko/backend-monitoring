package middleware

import (
	"log"

	"github.com/gin-gonic/gin"

	"monitoring-cctv-be/pkg/response"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Printf("panic recovered: %v", recovered)
		response.InternalError(c, "Terjadi kesalahan pada server")
	})
}
