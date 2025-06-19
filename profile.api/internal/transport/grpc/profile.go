package grpc_handler

import (
	"context"
	"profile-api/internal/model"
	"profile-api/internal/server/grpc/proto"
	"profile-api/internal/service"
	"profile-api/pkg/logger"
	"sync"

	"go.uber.org/zap"
)

type ProfileHandler struct {
	services *service.Services
	proto.UnimplementedProfileServiceServer
}

func NewProfileHandler(services *service.Services) *ProfileHandler {
	return &ProfileHandler{services: services}
}

func (h *ProfileHandler) GetUserBriefInfo(ctx context.Context, req *proto.UserRequest) (*proto.UserBriefInfoResponse, error) {
	userBriefInfo, err := h.services.Profiles.GetUserBriefProfileForNotification(ctx, model.UserRequest{
		SenderID:    req.SenderID,
		RecipientID: req.RecipientID,
	})

	if err != nil {
		logger.Error(
			zap.String("handler", "grpc"),
			zap.String("action", "GetUserBriefProfile()"),
			zap.Error(err),
		)
		return &proto.UserBriefInfoResponse{
			Response: &proto.Response{
				Success: false,
				Message: "Failed to get user brief info for notification",
			},
		}, err
	}

	return &proto.UserBriefInfoResponse{
		UserID:    userBriefInfo.UserID,
		Username:  userBriefInfo.Username,
		Name:      userBriefInfo.Name,
		AvatarURL: userBriefInfo.AvatarURL,
		Response: &proto.Response{
			Success: true,
			Message: "Success to get user brief info for notification",
		},
	}, nil
}

func (h *ProfileHandler) GetUsersBriefInfo(ctx context.Context, req *proto.UsersRequest) (*proto.UsersBriefInfoResponse, error) {
	var (
		usersBriefProfile      []model.UserBriefInfo
		usersBriefProfileProto []*proto.UserBriefInfoResponse
		wg                     sync.WaitGroup
		mu                     sync.Mutex
	)

	for _, r := range req.RecipientIDs {
		recipientID := r

		wg.Add(1)
		go func(senderID, recipientID string) {
			defer wg.Done()

			userBriefInfo, err := h.services.Profiles.GetUserBriefProfileForNotification(ctx, model.UserRequest{
				SenderID:    senderID,
				RecipientID: recipientID,
			})
			if err != nil {
				logger.Error(
					zap.String("handler", "grpc"),
					zap.String("action", "GetUserBriefProfile()"),
					zap.Error(err),
				)
				return
			}

			mu.Lock()
			usersBriefProfile = append(usersBriefProfile, userBriefInfo)
			mu.Unlock()
		}(req.SenderID, recipientID)
	}

	wg.Wait()

	for _, userBrief := range usersBriefProfile {
		usersBriefProfileProto = append(usersBriefProfileProto, &proto.UserBriefInfoResponse{
			UserID:    userBrief.UserID,
			Username:  userBrief.Username,
			Name:      userBrief.Name,
			AvatarURL: userBrief.AvatarURL,
		})
	}

	return &proto.UsersBriefInfoResponse{
		UsersBriefInfoResponse: usersBriefProfileProto,
		Response: &proto.Response{
			Success: true,
			Message: "Users fetched successfully",
		},
	}, nil
}

func (h *ProfileHandler) GetUsersProfile(ctx context.Context, req *proto.UsersRequest) (*proto.UsersProfileResponse, error) {
	usersProfile, err := h.services.Profiles.GetUserProfiles(ctx, req.SenderID, req.RecipientIDs)
	if err != nil {
		logger.Error(
			zap.String("handler", "grpc"),
			zap.String("action", "GetUsersProfile()"),
			zap.Error(err),
		)
		return &proto.UsersProfileResponse{
			Response: &proto.Response{
				Success: false,
				Message: "Failed to get users' profiles",
			},
		}, err
	}

	var protoUsersProfile []*proto.User

	for _, userProfile := range usersProfile {
		protoUserProfile := &proto.User{
			UserID:      userProfile.UserID,
			Username:    userProfile.Username,
			DisplayName: userProfile.DisplayName,
			Bio:         userProfile.Bio,
			Email:       userProfile.Email,
			Phone:       userProfile.Phone,
			AvatarURL:   userProfile.AvatarURL,
		}
		protoUsersProfile = append(protoUsersProfile, protoUserProfile)
	}

	return &proto.UsersProfileResponse{
		Users: protoUsersProfile,
		Response: &proto.Response{
			Success: true,
			Message: "Successfully retrieved users' profiles",
		},
	}, nil
}
