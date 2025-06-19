package v1

import (
	"auth-api/internal/model"
	"auth-api/pkg/broker"
	"auth-api/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *Handler) initAuthRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
		auth.GET("/refresh", h.refresh)
		auth.GET("/verify", h.AuthMiddleware(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Authenticated"})
		})
		auth.POST("/verify-code", h.verifyCode)
		auth.POST("/social/callback", h.socialAuth)
	}
}

func (h *Handler) signUp(c *gin.Context) {
	var request model.UserSignUp

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := request.Validate(); err != nil {
		logger.Error("Failed to validate request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	tokens, user, err := h.services.Auth.SignUp(c.Request.Context(), request)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.SetCookie("access_token", tokens.AccessToken, int(h.services.Auth.GetAccessTokenTTL().Seconds()), "/", "localhost", false, true)
	c.SetCookie("refresh_token", tokens.RefreshToken, int(h.services.Auth.GetRefreshTokenTTL().Seconds()), "/", "localhost", false, true)

	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "User signed up successfully",
	})
}

func (h *Handler) signIn(c *gin.Context) {
	var userSignIn model.UserSignIn
	if err := c.BindJSON(&userSignIn); err != nil {
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := userSignIn.Validate(); err != nil {
		logger.Error("Failed to validate request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	tokens, user, err := h.services.Auth.SignIn(c.Request.Context(), userSignIn)
	if err != nil {
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.SetCookie("access_token", tokens.AccessToken, int(h.services.Auth.GetAccessTokenTTL().Seconds()), "/", "localhost", false, true)
	c.SetCookie("refresh_token", tokens.RefreshToken, int(h.services.Auth.GetRefreshTokenTTL().Seconds()), "/", "localhost", false, true)

	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "User signed in successfully",
	})

}

func (h *Handler) refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		newResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	if refreshToken == "" {
		newResponse(c, http.StatusUnauthorized, "refresh token is empty")
		return
	}

	tokens, err := h.services.Auth.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.SetCookie("access_token", tokens.AccessToken, int(h.services.Auth.GetAccessTokenTTL().Seconds()), "/", "localhost", false, true)
	c.SetCookie("refresh_token", tokens.RefreshToken, int(h.services.Auth.GetRefreshTokenTTL().Seconds()), "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Token refreshed successfully"})
}

func (h *Handler) verifyCode(c *gin.Context) {
	var request model.VerifyInput

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := request.Validate(); err != nil {
		logger.Error("Failed to validate request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	vc, err := h.services.Auth.VerifyCode(c.Request.Context(), request.Recipient)
	if err != nil {
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	notificationRabbitMQ := model.NotificationRabbitMQ{
		Exchange: broker.EXCHANGE_VERIFY_CODE,
	}

	switch request.Type {
	case model.EMAIL:
		notificationRabbitMQ.RoutingKey = broker.ROUTING_KEY_VERIFY_CODE_EMAIL
	case model.PHONE:
		notificationRabbitMQ.RoutingKey = broker.ROUTING_KEY_VERIFY_CODE_PHONE
	default:
		newResponse(c, http.StatusBadRequest, "unsupported notification type")
		return
	}

	if err := h.services.Notifications.SendNotification(c.Request.Context(), notificationRabbitMQ, vc); err != nil {
		logger.Error("Failed to send notification", zap.Error(err))
		newResponse(c, http.StatusInternalServerError, "failed to send notification")
		return
	}

	newResponse(c, http.StatusOK, "verification code sent successfully")
}

func (h *Handler) socialAuth(c *gin.Context) {
	var request model.SocialMediaRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := request.Validate(); err != nil {
		logger.Error("Failed to validate request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	tokens, user, err := h.services.Auth.EntranceViaSocialMedia(c.Request.Context(), request)
	if err != nil {
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.SetCookie("access_token", tokens.AccessToken, int(h.services.Auth.GetAccessTokenTTL().Seconds()), "/", "localhost", false, true)
	c.SetCookie("refresh_token", tokens.RefreshToken, int(h.services.Auth.GetRefreshTokenTTL().Seconds()), "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "User signed in successfully",
	})

}
