package v1

import (
	"auth-api/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil {
			if err == http.ErrNoCookie {
				h.refresh(c)
				logger.Info("A request to refresh the token has been sent.")
				c.JSON(http.StatusGatewayTimeout, gin.H{"message": "access token was updated"})
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			c.Abort()
			return
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "no token"})
			c.Abort()
			return
		}

		err = h.services.Auth.VerifyToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			c.Abort()
			return
		}
		c.Next()
	}
}
