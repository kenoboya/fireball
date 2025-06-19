package v1

import (
	"chat-api/internal/model"
	"chat-api/internal/server/grpc/profile/proto"
	"chat-api/pkg/broker"
	"chat-api/pkg/logger"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func (h *Handler) sendMessage(ws *websocket.Conn, request model.CreateMessageRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var (
		wg              sync.WaitGroup
		mu              sync.Mutex
		sendErr         error
		participantsIDs []string
		chatDB          model.ChatDB
	)

	if request.MessageWithData.MessageDB.Content != nil {
		encrypted, err := h.services.MessageEncrypter.Encrypt(*request.MessageWithData.MessageDB.Content)
		if err != nil {
			logger.Error("Failed to encrypted message", zap.Error(err))
			return
		}
		request.MessageWithData.MessageDB.Content = &encrypted
	}

	wg.Add(3)

	go func() {
		defer wg.Done()
		if err := h.services.Messages.SendMessage(ctx, &request); err != nil {
			mu.Lock()
			sendErr = err
			mu.Unlock()
			cancel()
		}
	}()

	go func() {
		defer wg.Done()
		ids, err := h.services.Chats.GetParticipantsOfChat(ctx, request.ChatID)
		mu.Lock()
		if err != nil && sendErr == nil {
			sendErr = err
			cancel()
		}
		participantsIDs = ids
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		chat, err := h.services.Chats.GetChatByChatID(ctx, request.ChatID)
		mu.Lock()
		if err != nil && sendErr == nil {
			sendErr = err
			cancel()
		}
		chatDB = chat
		mu.Unlock()
	}()

	wg.Wait()

	if sendErr != nil {
		logger.Error("Failed during sendMessage flow", zap.Error(sendErr))
		return
	}

	if request.MessageWithData.MessageDB.Content != nil {
		decrypted, err := h.services.MessageEncrypter.Decrypt(*request.MessageWithData.MessageDB.Content)
		if err != nil {
			logger.Error("Failed to decrypted message", zap.Error(err))
			return
		}
		request.MessageWithData.MessageDB.Content = &decrypted
	}

	for _, recipientID := range participantsIDs {
		// Try to get the recipient's WebSocket connection
		webSocketConnection, err := h.services.Auth.GetWebSocket(ctx, recipientID)
		if err != nil {
			if errors.Is(err, model.ErrWebSocketNotFound) {
				// WebSocket not found — fallback to sending a notification instead

				// Get recipient's user brief info for the notification
				responseProfile, err := h.profileClient.GetUserBriefInfo(ctx, &proto.UserRequest{
					SenderID:    request.MessageWithData.MessageDB.SenderID,
					RecipientID: recipientID,
				})
				if err != nil {
					logger.Error("Failed to get profile for notification", zap.Error(err))
					continue
				}

				sender := model.UserBriefInfo{
					UserID:    responseProfile.UserID,
					Name:      responseProfile.Name,
					AvatarURL: responseProfile.AvatarURL,
				}

				// Send notification about message creation

				if err := h.services.Notifications.SendNotification(
					ctx,
					model.NotificationRabbitMQ{
						Exchange:   broker.EXCHANGE_MESSAGE,
						RoutingKey: broker.ROUTING_KEY_MESSAGE_SEND,
					},
					model.NotificationMessage{
						Message: model.MessageBriefInfo{
							MessageID: request.MessageWithData.MessageDB.MessageID,
							SenderID:  request.MessageWithData.MessageDB.SenderID,
							Type:      request.MessageWithData.MessageDB.Type,
							UpdatedAt: request.MessageWithData.MessageDB.CreatedAt,
						},
						Chat: model.ChatBriefInfo{
							ChatID:    chatDB.ChatID,
							CreatorID: chatDB.CreatorID,
							Name:      chatDB.Name,
							Encrypted: chatDB.Encrypted,
							UpdatedAt: chatDB.UpdatedAt,
						},
						Sender:      sender,
						RecipientID: recipientID,
					},
				); err != nil {
					logger.Error("Failed to send notification about new message", zap.Error(err))
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
}

// TODO sendMessageE2EE
