package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"auth-api/internal/model"
	"auth-api/internal/service"
	mock_service "auth-api/internal/service/mocks"
	"auth-api/pkg/broker"
	"auth-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/magiconair/properties/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/mock/gomock"
)

func TestHandler_SignUp(t *testing.T) {
	logger.InitLogger()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mock_service.NewMockAuth(ctrl)

	email := "test@example.com"
	phone := "+1234567890"
	now := time.Date(2025, 6, 16, 15, 4, 5, 0, time.UTC)
	mockTokens := model.Tokens{
		AccessToken:  "mocked-access-token",
		RefreshToken: "mocked-refresh-token",
	}

	mockUsers := []model.User{
		{
			UserID:       bson.NewObjectID(),
			Username:     "testuser1",
			Password:     "hashed-password",
			Email:        &email,
			Blocked:      model.UNBLOCKED,
			RegisteredAt: now,
		},
		{
			UserID:       bson.NewObjectID(),
			Username:     "testuser2",
			Password:     "hashed-password",
			Phone:        &phone,
			Blocked:      model.UNBLOCKED,
			RegisteredAt: now,
		},
		{
			UserID:       bson.NewObjectID(),
			Username:     "testuser3",
			Password:     "hashed-password",
			Email:        &email,
			Phone:        &phone,
			Blocked:      model.UNBLOCKED,
			RegisteredAt: now,
		},
	}

	type mockBehavior func(s *mock_service.MockAuth, userRequest model.UserSignUp)

	testTable := []struct {
		name                string
		inputBody           string
		inputRequest        model.UserSignUp
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK-with-email",
			inputBody: `{"user":{"username":"testuser1","password":"P@ssw0rd!","email":"test@example.com"},"verify-code":{"recipient":"test@example.com","code":"123456"}}`,
			inputRequest: model.UserSignUp{
				User: model.User{
					Username: "testuser1",
					Password: "P@ssw0rd!",
					Email:    &email,
					Blocked:  model.UNBLOCKED,
				},
				VerifyCode: model.VerifyCodeInput{
					Recipient: email,
					Code:      "123456",
				},
			},
			mockBehavior: func(s *mock_service.MockAuth, userRequest model.UserSignUp) {
				s.EXPECT().SignUp(gomock.Any(), userSignUpEquals(userRequest)).Return(mockTokens, mockUsers[0], nil)
				s.EXPECT().GetAccessTokenTTL().AnyTimes().Return(time.Minute * 15)
				s.EXPECT().GetRefreshTokenTTL().AnyTimes().Return(time.Hour * 24 * 7)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: makeExpectedResponseSignUp(mockUsers[0]),
		},
		{
			name:      "OK-with-phone",
			inputBody: `{"user":{"username":"testuser2","password":"P@ssw0rd!","phone":"+1234567890"},"verify-code":{"recipient":"+1234567890","code":"123456"}}`,
			inputRequest: model.UserSignUp{
				User: model.User{
					Username: "testuser2",
					Password: "P@ssw0rd!",
					Phone:    &phone,
					Blocked:  model.UNBLOCKED,
				},
				VerifyCode: model.VerifyCodeInput{
					Recipient: phone,
					Code:      "123456",
				},
			},
			mockBehavior: func(s *mock_service.MockAuth, userRequest model.UserSignUp) {
				s.EXPECT().SignUp(gomock.Any(), userSignUpEquals(userRequest)).Return(mockTokens, mockUsers[1], nil)
				s.EXPECT().GetAccessTokenTTL().AnyTimes().Return(time.Minute * 15)
				s.EXPECT().GetRefreshTokenTTL().AnyTimes().Return(time.Hour * 24 * 7)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: makeExpectedResponseSignUp(mockUsers[1]),
		},
		{
			name:      "OK-with-phoneAndEmail",
			inputBody: `{"user":{"username":"testuser3","password":"P@ssw0rd!","email":"test@example.com","phone":"+1234567890"},"verify-code":{"recipient":"+1234567890","code":"123456"}}`,
			inputRequest: model.UserSignUp{
				User: model.User{
					Username: "testuser3",
					Password: "P@ssw0rd!",
					Email:    &email,
					Phone:    &phone,
					Blocked:  model.UNBLOCKED,
				},
				VerifyCode: model.VerifyCodeInput{
					Recipient: phone,
					Code:      "123456",
				},
			},
			mockBehavior: func(s *mock_service.MockAuth, userRequest model.UserSignUp) {
				s.EXPECT().SignUp(gomock.Any(), userSignUpEquals(userRequest)).Return(mockTokens, mockUsers[2], nil)
				s.EXPECT().GetAccessTokenTTL().AnyTimes().Return(time.Minute * 15)
				s.EXPECT().GetRefreshTokenTTL().AnyTimes().Return(time.Hour * 24 * 7)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: makeExpectedResponseSignUp(mockUsers[2]),
		},
		{
			name:                "EmailInvalid",
			inputBody:           `{"user":{"username":"baduser","password":"P@ssw0rd!","email":"bad-email"},"verify-code":{"recipient":"bad-email","code":"123456"}}`,
			inputRequest:        model.UserSignUp{},
			mockBehavior:        func(s *mock_service.MockAuth, userRequest model.UserSignUp) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"user validation failed: user data is invalid: email has invalid structure"}`,
		},
		{
			name:                "PhoneInvalid",
			inputBody:           `{"user":{"username":"baduser","password":"P@ssw0rd!","phone":"invalid-phone"},"verify-code":{"recipient":"invalid-phone","code":"123456"}}`,
			inputRequest:        model.UserSignUp{},
			mockBehavior:        func(s *mock_service.MockAuth, userRequest model.UserSignUp) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"user validation failed: user data is invalid: phone has invalid structure"}`,
		},
		{
			name:                "EmailAndPhoneIsEmpty",
			inputBody:           `{"user":{"username":"baduser","password":"P@ssw0rd!"},"verify-code":{"recipient":"","code":"123456"}}`,
			inputRequest:        model.UserSignUp{},
			mockBehavior:        func(s *mock_service.MockAuth, userRequest model.UserSignUp) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"user validation failed: user data is invalid: email and phone can't both be empty"}`,
		},
		{
			name:                "PasswordEmpty",
			inputBody:           `{"user":{"username":"baduser","password":"","email":"test@example.com"},"verify-code":{"recipient":"test@example.com","code":"123456"}}`,
			inputRequest:        model.UserSignUp{},
			mockBehavior:        func(s *mock_service.MockAuth, userRequest model.UserSignUp) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"user validation failed: user data is invalid: password has invalid structure"}`,
		},
		{
			name:                "PasswordInvalidStructure",
			inputBody:           `{"user":{"username":"baduser","password":"simplepass","email":"test@example.com"},"verify-code":{"recipient":"test@example.com","code":"123456"}}`,
			inputRequest:        model.UserSignUp{},
			mockBehavior:        func(s *mock_service.MockAuth, userRequest model.UserSignUp) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"user validation failed: user data is invalid: password has invalid structure"}`,
		},
		{
			name:                "UsernameEmpty",
			inputBody:           `{"user":{"username":"","password":"P@ssw0rd!","email":"test@example.com"},"verify-code":{"recipient":"test@example.com","code":"123456"}}`,
			inputRequest:        model.UserSignUp{},
			mockBehavior:        func(s *mock_service.MockAuth, userRequest model.UserSignUp) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"user validation failed: user data is invalid: username has invalid structure"}`,
		},
		{
			name:                "UsernameTooShort",
			inputBody:           `{"user":{"username":"a","password":"P@ssw0rd!","email":"test@example.com"},"verify-code":{"recipient":"test@example.com","code":"123456"}}`,
			inputRequest:        model.UserSignUp{},
			mockBehavior:        func(s *mock_service.MockAuth, userRequest model.UserSignUp) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"user validation failed: user data is invalid: username has invalid structure"}`,
		},
		{
			name:                "CodeInvalid",
			inputBody:           `{"user":{"username":"baduser","password":"P@ssw0rd!","email":"test@example.com"},"verify-code":{"recipient":"bad-email","code":"12346"}}`,
			inputRequest:        model.UserSignUp{},
			mockBehavior:        func(s *mock_service.MockAuth, userRequest model.UserSignUp) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"verify-code validation failed: user data is invalid: code must be exactly 6 characters"}`,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(mockAuth, tc.inputRequest)

			services := &service.Services{
				Auth: mockAuth,
			}
			handler := NewHandler(services)
			r := gin.New()
			r.POST("/sign-up", handler.signUp)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/sign-up", bytes.NewBufferString(tc.inputBody))
			r.ServeHTTP(w, req)

			assert.Equal(t, w.Code, tc.expectedStatusCode)
			assert.Equal(t, w.Body.String(), tc.expectedRequestBody)
		})
	}
}

func TestHandler_SignIn(t *testing.T) {
	logger.InitLogger()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mock_service.NewMockAuth(ctrl)

	email := "test@example.com"
	phone := "+1234567890"
	now := time.Date(2025, 6, 16, 15, 4, 5, 0, time.UTC)
	mockTokens := model.Tokens{
		AccessToken:  "mocked-access-token",
		RefreshToken: "mocked-refresh-token",
	}

	mockUsers := []model.User{
		{
			UserID:       bson.NewObjectID(),
			Username:     "testuser1",
			Password:     "hashed-password",
			Email:        &email,
			Blocked:      model.UNBLOCKED,
			RegisteredAt: now,
		},
		{
			UserID:       bson.NewObjectID(),
			Username:     "testuser2",
			Password:     "hashed-password",
			Phone:        &phone,
			Blocked:      model.UNBLOCKED,
			RegisteredAt: now,
		},
	}

	type mockBehavior func(s *mock_service.MockAuth, userRequest model.UserSignIn)

	testTable := []struct {
		name                string
		inputBody           string
		inputRequest        model.UserSignIn
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK-1",
			inputBody: `{"login":"kazyalka212","password":"P@ssw0rd!","verify-code":{"recipient":"test@example.com","code":"123456"}}`,
			inputRequest: model.UserSignIn{
				Login:    "kazyalka212",
				Password: "P@ssw0rd!",
				VerifyCode: model.VerifyCodeInput{
					Recipient: email,
					Code:      "123456",
				},
			},
			mockBehavior: func(s *mock_service.MockAuth, userRequest model.UserSignIn) {
				s.EXPECT().SignIn(gomock.Any(), userRequest).Return(mockTokens, mockUsers[0], nil)
				s.EXPECT().GetAccessTokenTTL().AnyTimes().Return(time.Minute * 15)
				s.EXPECT().GetRefreshTokenTTL().AnyTimes().Return(time.Hour * 24 * 7)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: makeExpectedResponseSignIn(mockUsers[0]),
		},
		{
			name:      "OK-2",
			inputBody: `{"login":"+1234567890","password":"P@ssw0rd!","verify-code":{"recipient":"+1234567890","code":"123456"}}`,
			inputRequest: model.UserSignIn{
				Login:    phone,
				Password: "P@ssw0rd!",
				VerifyCode: model.VerifyCodeInput{
					Recipient: phone,
					Code:      "123456",
				},
			},
			mockBehavior: func(s *mock_service.MockAuth, userRequest model.UserSignIn) {
				s.EXPECT().SignIn(gomock.Any(), userRequest).Return(mockTokens, mockUsers[1], nil)
				s.EXPECT().GetAccessTokenTTL().AnyTimes().Return(time.Minute * 15)
				s.EXPECT().GetRefreshTokenTTL().AnyTimes().Return(time.Hour * 24 * 7)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: makeExpectedResponseSignIn(mockUsers[1]),
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(mockAuth, tc.inputRequest)

			services := &service.Services{
				Auth: mockAuth,
			}
			handler := NewHandler(services)
			r := gin.New()
			r.POST("/sign-in", handler.signIn)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/sign-in", bytes.NewBufferString(tc.inputBody))
			r.ServeHTTP(w, req)

			assert.Equal(t, w.Code, tc.expectedStatusCode)
			assert.Equal(t, w.Body.String(), tc.expectedRequestBody)
		})
	}
}

func TestHandler_Refresh(t *testing.T) {
	logger.InitLogger()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mock_service.NewMockAuth(ctrl)

	mockTokens := model.Tokens{
		AccessToken:  "mocked-access-token",
		RefreshToken: "mocked-refresh-token",
	}

	type mockBehavior func(s *mock_service.MockAuth, refreshToken string)

	testTable := []struct {
		name                string
		refreshToken        string
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:         "OK-1",
			refreshToken: "old-refresh-token",
			mockBehavior: func(s *mock_service.MockAuth, refreshToken string) {
				s.EXPECT().Refresh(gomock.Any(), refreshToken).Return(mockTokens, nil)
				s.EXPECT().GetAccessTokenTTL().AnyTimes().Return(time.Minute * 15)
				s.EXPECT().GetRefreshTokenTTL().AnyTimes().Return(time.Hour * 24 * 7)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: `{"message":"Token refreshed successfully"}`,
		},
		{
			name:                "MissingCookie",
			refreshToken:        "",
			mockBehavior:        func(s *mock_service.MockAuth, refreshToken string) {},
			expectedStatusCode:  http.StatusUnauthorized,
			expectedRequestBody: `{"message":"http: named cookie not present"}`,
		},
		{
			name:                "EmptyRefreshToken",
			refreshToken:        "",
			mockBehavior:        func(s *mock_service.MockAuth, refreshToken string) {},
			expectedStatusCode:  http.StatusUnauthorized,
			expectedRequestBody: `{"message":"http: named cookie not present"}`,
		},

		{
			name:         "InvalidRefreshToken",
			refreshToken: "invalid-token",
			mockBehavior: func(s *mock_service.MockAuth, refreshToken string) {
				s.EXPECT().Refresh(gomock.Any(), refreshToken).Return(model.Tokens{}, fmt.Errorf("invalid refresh token"))
			},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"invalid refresh token"}`,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(mockAuth, tc.refreshToken)

			services := &service.Services{
				Auth: mockAuth,
			}
			handler := NewHandler(services)
			r := gin.New()
			r.POST("/refresh", handler.refresh)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/refresh", nil)

			if tc.refreshToken != "" {
				req.AddCookie(&http.Cookie{Name: "refresh_token", Value: tc.refreshToken})
			}

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatusCode, w.Code)
			assert.Equal(t, tc.expectedRequestBody, w.Body.String())
		})
	}
}

func TestHandler_VerifyCode(t *testing.T) {
	logger.InitLogger()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	email := "test@example.com"
	phone := "+1234567890"

	type mockBehavior func(auth *mock_service.MockAuth, notifications *mock_service.MockNotifications, recipient string, notification model.NotificationRabbitMQ, verifyCode model.VerifyCodeInput)

	testTable := []struct {
		name                string
		inputBody           string
		inputRequest        model.VerifyInput
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK with email",
			inputBody: `{"recipient":"` + email + `","type":"email"}`,
			inputRequest: model.VerifyInput{
				Recipient: email,
				Type:      "email",
			},
			mockBehavior: func(auth *mock_service.MockAuth, notifications *mock_service.MockNotifications, recipient string, notification model.NotificationRabbitMQ, verifyCode model.VerifyCodeInput) {
				auth.EXPECT().VerifyCode(gomock.Any(), recipient).Return(verifyCode, nil)
				notifications.EXPECT().SendNotification(gomock.Any(), notification, verifyCode).Return(nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: `{"message":"verification code sent successfully"}`,
		},
		{
			name:      "OK with phone",
			inputBody: `{"recipient":"` + phone + `","type":"phone"}`,
			inputRequest: model.VerifyInput{
				Recipient: phone,
				Type:      "phone",
			},
			mockBehavior: func(auth *mock_service.MockAuth, notifications *mock_service.MockNotifications, recipient string, notification model.NotificationRabbitMQ, verifyCode model.VerifyCodeInput) {
				auth.EXPECT().VerifyCode(gomock.Any(), recipient).Return(verifyCode, nil)
				notifications.EXPECT().SendNotification(gomock.Any(), notification, verifyCode).Return(nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: `{"message":"verification code sent successfully"}`,
		},
		{
			name:      "Unsupported type",
			inputBody: `{"recipient":"` + email + `","type":"fax"}`,
			inputRequest: model.VerifyInput{
				Recipient: email,
				Type:      "fax",
			},
			mockBehavior: func(auth *mock_service.MockAuth, notifications *mock_service.MockNotifications, recipient string, notification model.NotificationRabbitMQ, verifyCode model.VerifyCodeInput) {
				// Ничего не ожидается, так как ошибка будет до вызова VerifyCode
			},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"user data is invalid: invalid type"}`,
		},
		{
			name:      "VerifyCode returns error",
			inputBody: `{"recipient":"` + email + `","type":"email"}`,
			inputRequest: model.VerifyInput{
				Recipient: email,
				Type:      "email",
			},
			mockBehavior: func(auth *mock_service.MockAuth, notifications *mock_service.MockNotifications, recipient string, notification model.NotificationRabbitMQ, verifyCode model.VerifyCodeInput) {
				auth.EXPECT().VerifyCode(gomock.Any(), recipient).Return(model.VerifyCodeInput{}, errors.New("verify failed"))
			},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"verify failed"}`,
		},
		{
			name:      "SendNotification returns error",
			inputBody: `{"recipient":"` + phone + `","type":"phone"}`,
			inputRequest: model.VerifyInput{
				Recipient: phone,
				Type:      "phone",
			},
			mockBehavior: func(auth *mock_service.MockAuth, notifications *mock_service.MockNotifications, recipient string, notification model.NotificationRabbitMQ, verifyCode model.VerifyCodeInput) {
				auth.EXPECT().VerifyCode(gomock.Any(), recipient).Return(verifyCode, nil)
				notifications.EXPECT().SendNotification(gomock.Any(), notification, verifyCode).Return(errors.New("send failed"))
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedRequestBody: `{"message":"failed to send notification"}`,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			mockAuthCtrl := mock_service.NewMockAuth(ctrl)
			mockNotifCtrl := mock_service.NewMockNotifications(ctrl)

			notification := model.NotificationRabbitMQ{
				Exchange: broker.EXCHANGE_VERIFY_CODE,
			}
			if tc.inputRequest.Type == "email" {
				notification.RoutingKey = broker.ROUTING_KEY_VERIFY_CODE_EMAIL
			} else if tc.inputRequest.Type == "phone" {
				notification.RoutingKey = broker.ROUTING_KEY_VERIFY_CODE_PHONE
			}

			verifyCode := model.VerifyCodeInput{
				Recipient: tc.inputRequest.Recipient,
				Code:      "123456",
			}

			tc.mockBehavior(mockAuthCtrl, mockNotifCtrl, tc.inputRequest.Recipient, notification, verifyCode)

			services := &service.Services{
				Auth:          mockAuthCtrl,
				Notifications: mockNotifCtrl,
			}
			handler := NewHandler(services)
			r := gin.New()
			r.POST("/verify-code", handler.verifyCode)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/verify-code", bytes.NewBufferString(tc.inputBody))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatusCode, w.Code)
			assert.Equal(t, tc.expectedRequestBody, w.Body.String())
		})
	}
}

type userSignUpMatcher struct {
	expected model.UserSignUp
}

func (m userSignUpMatcher) Matches(x interface{}) bool {
	actual, ok := x.(model.UserSignUp)
	if !ok {
		return false
	}

	if actual.User.Username != m.expected.User.Username {
		println("username mismatch:", actual.User.Username, m.expected.User.Username)
		return false
	}

	if (actual.User.Email == nil) != (m.expected.User.Email == nil) {
		println("email nil mismatch")
		return false
	}
	if actual.User.Email != nil && m.expected.User.Email != nil && *actual.User.Email != *m.expected.User.Email {
		println("email value mismatch:", *actual.User.Email, *m.expected.User.Email)
		return false
	}

	if (actual.User.Phone == nil) != (m.expected.User.Phone == nil) {
		println("phone nil mismatch")
		return false
	}
	if actual.User.Phone != nil && m.expected.User.Phone != nil && *actual.User.Phone != *m.expected.User.Phone {
		println("phone value mismatch:", *actual.User.Phone, *m.expected.User.Phone)
		return false
	}

	if actual.VerifyCode.Recipient != m.expected.VerifyCode.Recipient {
		println("verify recipient mismatch:", actual.VerifyCode.Recipient, m.expected.VerifyCode.Recipient)
		return false
	}
	if actual.VerifyCode.Code != m.expected.VerifyCode.Code {
		println("verify code mismatch:", actual.VerifyCode.Code, m.expected.VerifyCode.Code)
		return false
	}

	return true
}

func (m userSignUpMatcher) String() string {
	return "matches UserSignUp with expected fields"
}

func userSignUpEquals(expected model.UserSignUp) gomock.Matcher {
	return userSignUpMatcher{expected: expected}
}

func makeExpectedResponseSignUp(user model.User) string {
	resp := map[string]interface{}{
		"user":    user,
		"message": "User signed up successfully",
	}
	bytes, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func makeExpectedResponseSignIn(user model.User) string {
	resp := map[string]interface{}{
		"user":    user,
		"message": "User signed in successfully",
	}
	bytes, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
