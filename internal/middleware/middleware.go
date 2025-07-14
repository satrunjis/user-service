package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/satrunjis/user-service/internal/service"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() 
		if len(c.Errors) > 0 && !c.Writer.Written() {
			err := c.Errors.Last().Err

			switch e := err.(type) {
			case *service.ServiceError:
				handleServiceError(c, e)
			default:
				handleUnexpectedError(c, err)
			}
		}
	}
}

func handleServiceError(c *gin.Context, err *service.ServiceError) {
	status := http.StatusInternalServerError

	switch err.Code {
	case service.ErrCodeNotFound:
		status = http.StatusNotFound
	case service.ErrCodeInvalidInput:
		status = http.StatusBadRequest
	case service.ErrCodeAlreadyExists:
		status = http.StatusConflict
	case service.ErrCodeInternal:
		status = http.StatusInternalServerError
	}

	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    err.Code,
			"message": err.Message,
		},
	})
}

func handleUnexpectedError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		},
	})
}
func Cors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Разрешить все домены (для разработки)
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
