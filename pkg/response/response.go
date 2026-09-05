package response

import "github.com/gin-gonic/gin"

// ErrorBody mirrors the frontend's ApiError type (src/lib/types/api.ts) —
// client.ts reads `code`/`message`/`details` directly, with no envelope.
type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// OK writes the data directly at the top level — client.ts does
// `return (await res.json()) as TResponse` with no unwrapping.
func OK(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}

func Fail(c *gin.Context, status int, code string, message string) {
	c.JSON(status, ErrorBody{Code: code, Message: message})
}

func BadRequest(c *gin.Context, message string) { Fail(c, 400, "bad_request", message) }

func Unauthorized(c *gin.Context, message string) { Fail(c, 401, "unauthorized", message) }

func Forbidden(c *gin.Context, message string) { Fail(c, 403, "forbidden", message) }

func NotFound(c *gin.Context, message string) { Fail(c, 404, "not_found", message) }

func Conflict(c *gin.Context, message string) { Fail(c, 409, "conflict", message) }

func InternalError(c *gin.Context, message string) { Fail(c, 500, "internal_error", message) }
