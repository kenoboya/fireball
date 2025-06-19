package service

import (
	"context"
	"encoding/json"
	"fmt"
	"notification-api/internal/config"
	"notification-api/internal/model"
	repo "notification-api/internal/repository"
	"notification-api/pkg/logger"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type PhoneService struct {
	twilioConfig           config.TwilioConfig
	verificationRepository repo.Verification
}

func NewPhoneService(twilioConfig config.TwilioConfig, verificationRepository repo.Verification) *PhoneService {
	return &PhoneService{
		twilioConfig:           twilioConfig,
		verificationRepository: verificationRepository,
	}
}

func (s *PhoneService) SendVerifyCodeToPhone(ctx context.Context, vc model.VerifyCodeInput) error {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: s.twilioConfig.AccountSID,
		Password: s.twilioConfig.AuthToken,
	})

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(vc.Recipient)
	params.SetFrom(s.twilioConfig.PhoneNumber)
	params.SetBody(fmt.Sprintf("Your verification code is: %s", vc.Code))

	logger.Info("Sending SMS",
		"to", vc.Recipient,
		"from", s.twilioConfig.PhoneNumber,
		"code", vc.Code)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		logger.Error("Failed to send SMS", "error", err)
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	responseJSON, _ := json.Marshal(resp)

	if err := s.verificationRepository.SetRecordVerificationLog(ctx, vc, model.SMS); err != nil {
		logger.Errorf("Failed to record verification SMS: %v", err)
		return fmt.Errorf("failed to record verification log: %w", err)
	}

	logger.Info("SMS sent successfully", "response", string(responseJSON))
	return nil
}
