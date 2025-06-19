package v1

import (
	"net/http"
	"notification-api/internal/model"
	"notification-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *Handler) initNotificationRoutes(router *gin.RouterGroup) {
	notifications := router.Group("/notifications")
	{
		notifications.GET("", h.GetNotifications)
		notifications.PATCH("/chats/mute", h.SetUserChatNotification)
	}
}

func (h *Handler) SetUserChatNotification(c *gin.Context) {
	var request model.UserNotification

	userID := h.extractUserIDFromToken(c)
	if userID == "" {
		logger.Error("User ID is empty, unauthorized access")
		newResponse(c, http.StatusUnauthorized, "Unauthorized: token is missing or invalid")
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.UserID != userID {
		logger.Error("Request user_id isn't user_id from", zap.Error(model.ErrInvalidUserData))
		newResponse(c, http.StatusForbidden, model.ErrInvalidUserData.Error())
		return
	}

	if err := h.services.Notifications.SetUserChatNotificationStatus(c.Request.Context(), request); err != nil {
		logger.Error("Failed to set chat notification status", zap.Error(err))
		newResponse(c, http.StatusInternalServerError, "Failed to update notification status")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification status successfully updated"})
}

func (h *Handler) GetNotifications(c *gin.Context) {
	userID := h.extractUserIDFromToken(c)
	if userID == "" {
		logger.Error("User ID is empty, unauthorized access")
		newResponse(c, http.StatusUnauthorized, "Unauthorized: token is missing or invalid")
		return
	}

	notificationResponse, err := h.services.Notifications.GetNotifications(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get notification by user_id", zap.Error(err))
		newResponse(c, http.StatusInternalServerError, "Failed to get notification by user_id")
		return
	}

	c.JSON(http.StatusOK, notificationResponse)
}

func (h *Handler) extractUserIDFromToken(c *gin.Context) string {
	token, err := c.Cookie("access_token")
	if err != nil || token == "" {
		newResponse(c, http.StatusUnauthorized, "Authorization token is missing")
		return ""
	}

	userID, err := h.services.Auth.ValidateToken(token)
	if err != nil {
		newResponse(c, http.StatusUnauthorized, "Invalid token")
		return ""
	}

	return userID
}
