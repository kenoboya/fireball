package model

import (
	"auth-api/pkg/logger"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

const (
	BLOCKED   = "blocked"
	UNBLOCKED = "unblocked"

	ADMIN_ROLE   = "admin"
	PREMIUM_ROLE = "premium"
	USER_ROLE    = "user"

	LENGTH_CODE = 6

	EMAIL = "email"
	PHONE = "phone"

	MIN_LENGTH_OF_USERNAME = 2
	MAX_LENGTH_OF_USERNAME = 35
)

var (
	passwordRegex = regexp.MustCompile(`^[A-Za-z\d!@#$%^&*]{8,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex    = regexp.MustCompile(`^\+?[0-9]{7,15}$`) // Accepts international format, digits only
)

type User struct {
	UserID       bson.ObjectID `bson:"_id,omitempty"`
	Username     string        `bson:"username"`
	Password     string        `json:"password,omitempty" bson:"password"`
	Email        *string       `bson:"email" json:"email,omitempty"`
	Phone        *string       `bson:"phone" json:"phone,omitempty"`
	Blocked      string        `bson:"blocked,omitempty" json:"blocked,omitempty"`
	RegisteredAt time.Time     `bson:"registered_at,omitempty" json:"registered_at,omitempty"`
}

func (u *User) Validate() error {
	if u.Username == "" || !isValidUsername(u.Username) {
		logger.Error("Username has invalid structure", zap.Error(ErrInvalidUserData))
		return fmt.Errorf("%w: username has invalid structure", ErrInvalidUserData)
	}

	if u.Password == "" || !isValidPassword(u.Password) {
		logger.Error("Password has invalid structure", zap.Error(ErrInvalidUserData))
		return fmt.Errorf("%w: password has invalid structure", ErrInvalidUserData)
	}

	if u.Email == nil && u.Phone == nil {
		logger.Error("Email and phone can't both be empty", zap.Error(ErrInvalidUserData))
		return fmt.Errorf("%w: email and phone can't both be empty", ErrInvalidUserData)
	}

	var emailValid, phoneValid bool
	if u.Email != nil {
		emailValid = isValidEmail(*u.Email)
	} else {
		emailValid = false
	}

	if u.Phone != nil {
		phoneValid = isValidPhone(*u.Phone)
	} else {
		phoneValid = false
	}

	if emailValid || phoneValid {
		return nil
	}

	if u.Email != nil && !emailValid {
		logger.Error("Email has invalid structure", zap.Error(ErrInvalidUserData))
		return fmt.Errorf("%w: email has invalid structure", ErrInvalidUserData)
	}

	if u.Phone != nil && !phoneValid {
		logger.Error("Phone has invalid structure", zap.Error(ErrInvalidUserData))
		return fmt.Errorf("%w: phone has invalid structure", ErrInvalidUserData)
	}
	return nil
}

func (u *User) IsBlocked() bool {
	return u.Blocked == BLOCKED
}

type UserSignUp struct {
	User       User            `json:"user"`
	VerifyCode VerifyCodeInput `json:"verify-code"`
}

func (r *UserSignUp) Validate() error {
	var errValidation error
	if errValidation = r.User.Validate(); errValidation != nil {
		return fmt.Errorf("user validation failed: %w", errValidation)
	}
	if errValidation = r.VerifyCode.Validate(); errValidation != nil {
		return fmt.Errorf("verify-code validation failed: %w", errValidation)
	}
	return nil
}

type UserSignIn struct {
	VerifyCode VerifyCodeInput `json:"verify-code"`
	Login      string          `json:"login"`
	Password   string          `json:"password"`
}

func (r *UserSignIn) Validate() error {
	var errValidation error
	if errValidation = r.VerifyCode.Validate(); errValidation != nil {
		return fmt.Errorf("verify-code validation failed: %w", errValidation)
	}

	if !isValidEmail(r.Login) && !isValidPhone(r.Login) && !isValidUsername(r.Login) {
		return fmt.Errorf("login validation failed")
	}

	if !isValidPassword(r.Password) {
		return fmt.Errorf("password validation failed")
	}

	return nil
}

type VerifyCodeInput struct {
	Recipient string `json:"recipient"`
	Code      string `json:"code"`
}

func (r *VerifyCodeInput) Validate() error {
	if len(r.Code) != LENGTH_CODE {
		logger.Error("Code must be exactly 6 characters", zap.Error(ErrInvalidUserData))
		return fmt.Errorf("%w: code must be exactly %d characters", ErrInvalidUserData, LENGTH_CODE)
	}

	if r.Recipient == "" || (!isValidEmail(r.Recipient) && !isValidPhone(r.Recipient)) {
		logger.Error("Recipient must be a valid email or phone", zap.Error(ErrInvalidUserData))
		return fmt.Errorf("%w: recipient must be a valid email or phone", ErrInvalidUserData)
	}

	return nil
}

type VerifyInput struct {
	Recipient string `json:"recipient"`
	Type      string `json:"type"` // email or phone
}

func (r *VerifyInput) Validate() error {
	if r.Recipient == "" {
		logger.Error("Recipient is empty", zap.Error(ErrInvalidUserData))
		return fmt.Errorf("%w: recipient is empty", ErrInvalidUserData)
	}

	switch r.Type {
	case EMAIL:
		if !isValidEmail(r.Recipient) {
			logger.Error("Recipient email has invalid structure", zap.Error(ErrInvalidUserData))
			return fmt.Errorf("%w: recipient email has invalid structure", ErrInvalidUserData)
		}
	case PHONE:
		if !isValidPhone(r.Recipient) {
			logger.Error("Phone has invalid structure", zap.Error(ErrInvalidUserData))
			return fmt.Errorf("%w: phone has invalid structure", ErrInvalidUserData)
		}
	default:
		logger.Error("invalid type", zap.Error(ErrInvalidUserData))
		return fmt.Errorf("%w: invalid type", ErrInvalidUserData)
	}

	return nil
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func isValidPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

func isValidPassword(password string) bool {
	if !passwordRegex.MatchString(password) {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	specials := "!@#$%^&*"

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case strings.ContainsRune(specials, ch):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func isValidUsername(username string) bool {
	length := len(username)
	if length > MAX_LENGTH_OF_USERNAME || length < MIN_LENGTH_OF_USERNAME {
		return false
	}
	return usernameRegex.MatchString(username)
}
