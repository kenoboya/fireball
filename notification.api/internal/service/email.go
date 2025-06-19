package service

import (
	"context"
	"fmt"
	"notification-api/internal/config"
	"notification-api/internal/model"
	repo "notification-api/internal/repository"
	"notification-api/pkg/logger"

	"gopkg.in/gomail.v2"
)

type EmailService struct {
	emailConfig            config.EmailConfig
	verificationRepository repo.Verification
}

func NewEmailService(emailConfig config.EmailConfig, verificationRepository repo.Verification) *EmailService {
	return &EmailService{
		emailConfig:            emailConfig,
		verificationRepository: verificationRepository,
	}
}

func (s *EmailService) SendVerifyCodeToEmail(ctx context.Context, vc model.VerifyCodeInput) error {
	message := gomail.NewMessage()

	message.SetHeader("From", s.emailConfig.Email)
	message.SetHeader("To", vc.Recipient)
	message.SetHeader("Subject", "Fireball Messenger — Verification Code")

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
  <body style="font-family: Arial, sans-serif; line-height: 1.6;">
    <p>Hello,</p>

    <p>Your verification code for <strong>Fireball Messenger</strong> is:</p>

    <pre style="
      background-color: #1e1e1e;
      color: #ffffff;
      padding: 15px;
      border-radius: 8px;
      font-size: 24px;
      letter-spacing: 3px;
      text-align: center;
    "><strong>%s</strong></pre>

    <p>If you did not request this code, please ignore this message.</p>

    <p>Thank you,<br>
    Fireball Messenger Team</p>
  </body>
</html>
`, vc.Code)

	message.SetBody("text/html", body)

	dialer := gomail.NewDialer(
		s.emailConfig.Smtp.Host,
		s.emailConfig.Smtp.Port,
		s.emailConfig.Smtp.Username,
		s.emailConfig.Smtp.Password,
	)

	if s.emailConfig.Smtp.Port == 465 {
		dialer.SSL = true
	}

	if err := dialer.DialAndSend(message); err != nil {
		logger.Errorf("Failed to send verification email to %s: %v", vc.Recipient, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	if err := s.verificationRepository.SetRecordVerificationLog(ctx, vc, model.EMAIL); err != nil {
		logger.Errorf("Failed to record verification email: %v", err)
		return fmt.Errorf("failed to record verification log: %w", err)
	}

	logger.Infof("Verification code sent to %s", vc.Recipient)
	return nil
}
