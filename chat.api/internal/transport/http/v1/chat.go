package v1

import (
	"chat-api/internal/model"
	"chat-api/internal/server/grpc/profile/proto"
	"chat-api/pkg/broker"
	"chat-api/pkg/logger"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func (h *Handler) initChatRoutes(router *gin.RouterGroup) {
	chat := router.Group("/chat")
	{
		chat.GET("/initialization", h.initializationOfChats)
		chat.PATCH("/pinned", h.pinnedChat)
		chat.POST("/role", h.addRole)
		chat.POST("/block", h.blockChat)
	}
}

func (h *Handler) blockChat(c *gin.Context) {
	var blockChat model.BlockChat
	userID := h.extractUserIDFromToken(c)
	if userID == "" {
		logger.Error("User ID is empty, unauthorized access")
		newResponse(c, http.StatusUnauthorized, "Unauthorized: token is missing or invalid")
		return
	}

	if err := c.ShouldBindJSON(&blockChat); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if blockChat.UserID != userID {
		logger.Error("Failed to bind request", zap.Error(model.ErrInvalidUserData))
		newResponse(c, http.StatusForbidden, "invalid request body")
		return
	}

	if err := h.services.Chats.SetBlockChat(c.Request.Context(), blockChat); err != nil {
		logger.Error("Failed to block/unlock chat", zap.Error(err))
		newResponse(c, http.StatusInternalServerError, "failed to set block/unblock chat")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully to set block/unblock chat"})
}

func (h *Handler) addRole(c *gin.Context) {
	var chatRole model.ChatRole
	userID := h.extractUserIDFromToken(c)
	if userID == "" {
		return
	}

	if err := c.ShouldBindJSON(&chatRole); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if chatRole.GranterID != userID {
		logger.Error("Failed to bind request", zap.Error(model.ErrInvalidUserData))
		newResponse(c, http.StatusForbidden, "invalid request body")
		return
	}

	if err := h.services.Chats.SetChatRole(c.Request.Context(), chatRole); err != nil {
		logger.Error("Failed to set role", zap.Error(err))
		newResponse(c, http.StatusInternalServerError, "failed to set role")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully set chat role"})
}

func (h *Handler) pinnedChat(c *gin.Context) {
	var pinnedChatWithFlag model.PinnedChatWithFlag
	userID := h.extractUserIDFromToken(c)
	if userID == "" {
		return
	}

	if err := c.ShouldBindJSON(&pinnedChatWithFlag); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		newResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if pinnedChatWithFlag.PinnedChat.UserID != userID {
		logger.Error("Failed to bind request", zap.Error(model.ErrInvalidUserData))
		newResponse(c, http.StatusForbidden, "invalid request body")
		return
	}

	if err := h.services.Chats.UpdatePinnedChat(c.Request.Context(), pinnedChatWithFlag); err != nil {
		logger.Error("Failed to update pinned chat", zap.Error(err))
		newResponse(c, http.StatusInternalServerError, "failed to update pinned chat")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully updated pinned chat"})
}

func (h *Handler) createPrivateChat(ws *websocket.Conn, request model.CreatePrivateChatRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if request.InitialMessage.MessageWithData.MessageDB.Content != nil {
		encrypted, err := h.services.MessageEncrypter.Encrypt(*request.InitialMessage.MessageWithData.MessageDB.Content)
		if err != nil {
			logger.Error("Failed to encrypted message", zap.Error(err))
			return
		}
		request.InitialMessage.MessageWithData.MessageDB.Content = &encrypted
	}

	// Attempt to create a private chat
	response, err := h.services.Chats.CreatePrivateChat(ctx, request)
	if err != nil {
		logger.Error("Failed to create private chat", zap.Error(err))
		return
	}

	if response.Message.MessageWithData.MessageDB.Content != nil {
		decrypted, err := h.services.MessageEncrypter.Decrypt(*response.Message.MessageWithData.MessageDB.Content)
		if err != nil {
			logger.Error("Failed to decrypted message", zap.Error(err))
			return
		}
		response.Message.MessageWithData.MessageDB.Content = &decrypted
	}

	// Try to get the recipient's WebSocket connection
	webSocketConnection, err := h.services.Auth.GetWebSocket(ctx, request.RecipientID)
	if err != nil {
		if errors.Is(err, model.ErrWebSocketNotFound) {
			// WebSocket not found — fallback to sending a notification instead

			// Get recipient's user brief info for the notification
			responseProfile, err := h.profileClient.GetUserBriefInfo(ctx, &proto.UserRequest{
				SenderID:    request.Chat.CreatorID,
				RecipientID: request.RecipientID,
			})
			if err != nil {
				logger.Error("Failed to get profile for notification", zap.Error(err))
				return
			}

			sender := model.UserBriefInfo{
				UserID:    responseProfile.UserID,
				Username:  responseProfile.Username,
				Name:      responseProfile.Name,
				AvatarURL: responseProfile.AvatarURL,
			}

			// Send notification about chat creation
			if err := h.services.Notifications.SendNotification(
				ctx,
				model.NotificationRabbitMQ{
					Exchange:   broker.EXCHANGE_CHAT,
					RoutingKey: broker.ROUTING_KEY_CHAT_CREATED,
				},
				model.NotificationChat{
					Chat: model.ChatBriefInfo{
						ChatID:    response.Chat.ChatID,
						CreatorID: response.Chat.CreatorID,
						Name:      response.Chat.Name,
						Encrypted: response.Chat.Encrypted,
						AvatarURL: response.Chat.AvatarURL,
						UpdatedAt: response.Chat.CreatedAt,
					},
					Sender:      sender,
					RecipientID: request.RecipientID,
				},
			); err != nil {
				logger.Error("Failed to send notification about creating chat", zap.Error(err))
				return
			}

			// Send notification about the first message
			if err := h.services.Notifications.SendNotification(
				ctx,
				model.NotificationRabbitMQ{
					Exchange:   broker.EXCHANGE_MESSAGE,
					RoutingKey: broker.ROUTING_KEY_MESSAGE_SEND,
				},
				model.NotificationMessage{
					Message: model.MessageBriefInfo{
						MessageID: response.Message.MessageWithData.MessageDB.MessageID,
						SenderID:  response.Message.MessageWithData.MessageDB.SenderID,
						Type:      response.Message.MessageWithData.MessageDB.Type,
						UpdatedAt: response.Message.MessageWithData.MessageDB.CreatedAt,
					},
					Chat: model.ChatBriefInfo{
						ChatID:    response.Chat.ChatID,
						CreatorID: response.Chat.CreatorID,
						Name:      response.Chat.Name,
						Encrypted: response.Chat.Encrypted,
						UpdatedAt: response.Chat.CreatedAt,
					},
					Sender:      sender,
					RecipientID: request.RecipientID,
				},
			); err != nil {
				logger.Error("Failed to send notification about new message", zap.Error(err))
				return
			}
			ws.WriteJSON(response)
			return
		}

		// Other unexpected errors while retrieving WebSocket
		logger.Warn("Failed to get WebSocket for user", zap.String("userID", request.RecipientID), zap.Error(err))
		return
	}

	// Validate the WebSocket connection before writing
	if webSocketConnection == nil || webSocketConnection.Conn == nil {
		logger.Warn("WebSocket connection is nil", zap.String("userID", request.RecipientID))
		return
	}

	// Send the created chat response over WebSocket
	if err := webSocketConnection.Conn.WriteJSON(response); err != nil {
		logger.Warn("Failed to send WebSocket message", zap.String("userID", request.RecipientID), zap.Error(err))
	}

	// response for creator
	ws.WriteJSON(response)
}

func (h *Handler) createGroupChat(ws *websocket.Conn, request model.CreateGroupChatRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt to create a group chat
	err := h.services.Chats.CreateGroupChat(ctx, &request)
	if err != nil {
		logger.Error("Failed to create group chat", zap.Error(err))
		return
	}

	for _, recipientID := range request.ParticipantsIDs {
		// Try to get the recipient's WebSocket connection
		webSocketConnection, err := h.services.Auth.GetWebSocket(ctx, recipientID)
		if err != nil {
			if errors.Is(err, model.ErrWebSocketNotFound) {
				// WebSocket not found — fallback to sending a notification instead

				// Get recipient's user brief info for the notification
				responseProfile, err := h.profileClient.GetUserBriefInfo(ctx, &proto.UserRequest{
					SenderID:    request.Chat.CreatorID,
					RecipientID: recipientID,
				})
				if err != nil {
					logger.Error("Failed to get profile for notification", zap.Error(err))
					continue // changed return to continue
				}

				sender := model.UserBriefInfo{
					UserID:    responseProfile.UserID,
					Name:      responseProfile.Name,
					AvatarURL: responseProfile.AvatarURL,
				}

				// Send notification about chat creation
				if err := h.services.Notifications.SendNotification(
					ctx,
					model.NotificationRabbitMQ{
						Exchange:   broker.EXCHANGE_CHAT,
						RoutingKey: broker.ROUTING_KEY_CHAT_CREATED,
					},
					model.NotificationChat{
						Chat: model.ChatBriefInfo{
							ChatID:    request.Chat.ChatID,
							CreatorID: request.Chat.CreatorID,
							Name:      request.Chat.Name,
							Encrypted: request.Chat.Encrypted,
							AvatarURL: request.Chat.AvatarURL,
							UpdatedAt: request.Chat.CreatedAt,
						},
						Sender:      sender,
						RecipientID: recipientID,
					},
				); err != nil {
					logger.Error("Failed to send notification about creating chat", zap.Error(err))
					continue
				}

				continue
			}

			// Other unexpected errors while retrieving WebSocket
			logger.Warn("Failed to get WebSocket for user", zap.String("userID", recipientID), zap.Error(err))
			continue
		}

		// Validate the WebSocket connection before writing
		if webSocketConnection == nil || webSocketConnection.Conn == nil {
			logger.Warn("WebSocket connection is nil", zap.String("userID", recipientID))
			continue
		}

		// Send the created chat response over WebSocket
		if err := webSocketConnection.Conn.WriteJSON(request); err != nil {
			logger.Warn("Failed to send WebSocket message", zap.String("userID", recipientID), zap.Error(err))
		}
	}

	ws.WriteJSON(request)
}

func (h *Handler) initializationOfChats(c *gin.Context) {
	userID := h.extractUserIDFromToken(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	type initResult struct {
		chats       []model.Chat
		pinnedChats []model.PinnedChatInit
		err         error
	}

	resultChan := make(chan initResult, 1)

	go func() {
		var res initResult
		res.chats, res.err = h.services.Chats.InitializeChatsForMessenger(c.Request.Context(), userID)
		if res.err != nil {
			resultChan <- res
			return
		}
		res.pinnedChats, res.err = h.services.Chats.InitializePinnedChatsForMessenger(c.Request.Context(), userID)
		resultChan <- res
	}()

	res := <-resultChan
	if res.err != nil {
		logger.Error("Failed to load chats or pinned chats", zap.Error(res.err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load chats"})
		return
	}

	var userProfilesIDs []string
	for _, chat := range res.chats {
		limit := model.USER_LIMIT_REQUEST
		if len(chat.ParticipantsIDs) < limit {
			limit = len(chat.ParticipantsIDs)
		}
		userProfilesIDs = append(userProfilesIDs, chat.ParticipantsIDs[:limit]...)
	}

	for _, pinnedChat := range res.pinnedChats {
		limit := model.USER_LIMIT_REQUEST
		if len(pinnedChat.Chat.ParticipantsIDs) < limit {
			limit = len(pinnedChat.Chat.ParticipantsIDs)
		}
		userProfilesIDs = append(userProfilesIDs, pinnedChat.Chat.ParticipantsIDs[:limit]...)
	}

	uniqueUserIDs := make(map[string]struct{})
	for _, id := range userProfilesIDs {
		uniqueUserIDs[id] = struct{}{}
	}
	userProfilesIDs = userProfilesIDs[:0]
	for id := range uniqueUserIDs {
		userProfilesIDs = append(userProfilesIDs, id)
	}

	response, err := h.profileClient.GetUsersBriefInfo(context.Background(), &proto.UsersRequest{
		SenderID:     userID,
		RecipientIDs: userProfilesIDs,
	})

	if err != nil {
		logger.Error("Failed to get user profiles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user profiles"})
		return
	}

	initialization := model.MessengerInitializer{
		Chats:        res.chats,
		PinnedChat:   res.pinnedChats,
		UsersProfile: make([]model.UserBriefInfo, 0, len(response.UsersBriefInfoResponse)),
	}

	for _, protoUserProfile := range response.UsersBriefInfoResponse {
		userBrief := model.UserBriefInfo{
			UserID:    protoUserProfile.UserID,
			Username:  protoUserProfile.Username,
			Name:      protoUserProfile.Name,
			AvatarURL: protoUserProfile.AvatarURL,
		}
		initialization.UsersProfile = append(initialization.UsersProfile, userBrief)
	}

	c.JSON(http.StatusOK, initialization)
}
