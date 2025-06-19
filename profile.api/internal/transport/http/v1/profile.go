package v1

import (
	"net/http"
	"profile-api/internal/model"
	"profile-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *Handler) initAuthRoutes(router *gin.RouterGroup) {
	profile := router.Group("/profiles")
	{
		profile.POST("", h.SetUserProfile)
		profile.GET("", h.GetUserProfile)
		profile.POST("/search", h.SearchProfile)

		contact := profile.Group("/contacts")
		{
			contact.POST("", h.SetProfilesOfUserContacts)
			contact.GET("", h.GetProfileOfUserContact)
			contact.GET("/all", h.GetProfilesOfUserContacts)
		}
	}
}

func (h *Handler) SetUserProfile(c *gin.Context) {
	var request model.User

	userID := h.extractUserIDFromToken(c)
	if userID == "" {
		logger.Error("User ID is empty, unauthorized access")
		newResponse(c, http.StatusUnauthorized, "Unauthorized: token is missing or invalid")
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if request.UserID != userID {
		logger.Error("Request user_id isn't user_id from", zap.Error(model.ErrInvalidUserData))
		newResponse(c, http.StatusForbidden, model.ErrInvalidUserData.Error())
		return
	}

	if err := h.services.Profiles.SetProfile(c.Request.Context(), request); err != nil {
		logger.Error("Failed to set profile", zap.Error(err))
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User profile created successfully",
	})
}

func (h *Handler) SetProfilesOfUserContacts(c *gin.Context) {
	var request model.Contact

	userID := h.extractUserIDFromToken(c)
	if userID == "" {
		logger.Error("User ID is empty, unauthorized access")
		newResponse(c, http.StatusUnauthorized, "Unauthorized: token is missing or invalid")
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if request.UserRequest.SenderID != userID {
		logger.Error("Request user_id isn't user_id from", zap.Error(model.ErrInvalidUserData))
		newResponse(c, http.StatusForbidden, model.ErrInvalidUserData.Error())
		return
	}

	if err := h.services.Contacts.SetContact(c.Request.Context(), request); err != nil {
		logger.Error("Failed to set contact", zap.Error(err))
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User contact was set successfully",
	})
}

func (h *Handler) GetUserProfile(c *gin.Context) {
	var request model.UserRequest

	userID := h.extractUserIDFromToken(c)
	if userID == "" {
		logger.Error("User ID is empty, unauthorized access")
		newResponse(c, http.StatusUnauthorized, "Unauthorized: token is missing or invalid")
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if request.SenderID != userID {
		logger.Error("Request user_id isn't user_id from", zap.Error(model.ErrInvalidUserData))
		newResponse(c, http.StatusForbidden, model.ErrInvalidUserData.Error())
		return
	}

	userBriefInfo, err := h.services.Profiles.GetUserBriefProfile(c.Request.Context(), request)
	if err != nil {
		logger.Error("Failed to get user brief profile", zap.Error(err))
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_brief_info": userBriefInfo,
		"message":         "User profile retrieved successfully",
	})

}

func (h *Handler) GetProfilesOfUserContacts(c *gin.Context) {
	userID := h.extractUserIDFromToken(c)
	if userID == "" {
		logger.Error("User ID is empty, unauthorized access")
		newResponse(c, http.StatusUnauthorized, "Unauthorized: token is missing or invalid")
		return
	}

	users, err := h.services.Profiles.GetContacts(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get contacts", zap.Error(err))
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":   users,
		"message": "Contacts retrieved successfully",
	})
}

func (h *Handler) GetProfileOfUserContact(c *gin.Context) {
	var request model.UserRequest
	userID := h.extractUserIDFromToken(c)
	if userID == "" {
		logger.Error("User ID is empty, unauthorized access")
		newResponse(c, http.StatusUnauthorized, "Unauthorized: token is missing or invalid")
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if request.SenderID != userID {
		logger.Error("Request user_id isn't user_id from", zap.Error(model.ErrInvalidUserData))
		newResponse(c, http.StatusForbidden, model.ErrInvalidUserData.Error())
		return
	}

	user, err := h.services.Profiles.GetContact(c.Request.Context(), request)
	if err != nil {
		logger.Error("Failed to get contact", zap.Error(err))
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":   user,
		"message": "Contact retrieved successfully",
	})

}

func (h *Handler) SearchProfile(c *gin.Context) {
	var searchRequest model.UserSearchRequest
	userID := h.extractUserIDFromToken(c)
	if userID == "" {
		logger.Error("User ID is empty, unauthorized access")
		newResponse(c, http.StatusUnauthorized, "Unauthorized: token is missing or invalid")
		return
	}

	if err := c.ShouldBindJSON(&searchRequest); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	foundProfile, err := h.services.Profiles.SearchProfile(c.Request.Context(), searchRequest)
	if err != nil {
		logger.Error("Failed to find profile", zap.Error(err))
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"found_profile": foundProfile,
		"message":       "Profile was found successfully",
	})
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
