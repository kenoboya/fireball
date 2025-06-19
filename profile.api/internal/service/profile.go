package service

import (
	"context"
	"profile-api/internal/model"
	repo "profile-api/internal/repository"
	"profile-api/pkg/logger"
	"regexp"
	"sync"
)

type ProfileService struct {
	repo         repo.Profile
	repoContacts repo.Contacts
}

func NewProfileService(repo repo.Profile, repoContacts repo.Contacts) *ProfileService {
	return &ProfileService{
		repo:         repo,
		repoContacts: repoContacts,
	}
}

func (s *ProfileService) SetProfile(ctx context.Context, user model.User) error {
	return s.repo.SetProfile(ctx, user)
}

func (s *ProfileService) getUserBriefProfileByID(ctx context.Context, userID string, request model.UserRequest) (model.UserBriefInfo, error) {
	userBI, err := s.repo.GetUserBriefProfile(ctx, userID)
	if err != nil {
		return model.UserBriefInfo{}, err
	}

	alias, err := s.repoContacts.GetAlias(ctx, request)
	if err != nil {
		logger.Infof("failed to get alias %s", err)
		return userBI, nil
	}

	if alias != "" {
		userBI.Name = alias
	}

	return userBI, nil
}

func (s *ProfileService) GetUserBriefProfileForNotification(ctx context.Context, request model.UserRequest) (model.UserBriefInfo, error) {
	return s.getUserBriefProfileByID(ctx, request.SenderID, request)
}

func (s *ProfileService) GetUserBriefProfile(ctx context.Context, request model.UserRequest) (model.UserBriefInfo, error) {
	return s.getUserBriefProfileByID(ctx, request.RecipientID, request)
}

func (s *ProfileService) GetUserProfiles(ctx context.Context, senderID string, recipientIDs []string) ([]model.User, error) {
	var users []model.User
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, recipientID := range recipientIDs {
		wg.Add(1)

		go func(senderID, recipientID string) {
			defer wg.Done()
			userProfile, err := s.repo.GetByUserID(ctx, recipientID)
			if err != nil {
				select {
				case errChan <- err:
					cancel()
				default:
				}
				return
			}

			alias, err := s.repoContacts.GetAlias(ctx, model.UserRequest{
				SenderID:    senderID,
				RecipientID: recipientID,
			})

			if err == nil {
				if alias != "" {
					userProfile.DisplayName = alias
				}
			}

			mu.Lock()
			users = append(users, userProfile)
			mu.Unlock()
		}(senderID, recipientID)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errChan:
		return nil, err
	case <-done:
		return users, nil
	}
}

func (s *ProfileService) GetByUserID(ctx context.Context, userID string) (model.User, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *ProfileService) UpdateProfile(ctx context.Context, user model.User) error {
	return s.repo.UpdateProfile(ctx, user)
}

func (s *ProfileService) DeleteProfile(ctx context.Context, userID string) error {
	return s.repo.DeleteProfile(ctx, userID)
}

func (s *ProfileService) GetContacts(ctx context.Context, senderID string) ([]model.User, error) {
	var (
		users []model.User
		wg    sync.WaitGroup
		mu    sync.Mutex
	)

	contacts, err := s.repoContacts.GetContacts(ctx, senderID)
	if err != nil {
		return nil, err
	}

	for _, c := range contacts {
		contact := c
		wg.Add(1)

		go func(contact model.Contact) {
			defer wg.Done()

			user, err := s.repo.GetByUserID(ctx, contact.UserRequest.RecipientID)
			if err != nil {
				logger.Errorf("Failed to get user %s: %v", contact.UserRequest.RecipientID, err)
				return
			}

			if contact.Alias != "" {
				user.DisplayName = contact.Alias
			}

			mu.Lock()
			users = append(users, user)
			mu.Unlock()
		}(contact)
	}

	wg.Wait()

	return users, nil
}

func (s *ProfileService) GetContact(ctx context.Context, request model.UserRequest) (model.User, error) {
	var (
		err  error
		user model.User
	)

	contact, err := s.repoContacts.GetContact(ctx, request)
	if err != nil {
		return model.User{}, err
	}

	user, err = s.repo.GetByUserID(ctx, contact.UserRequest.RecipientID)
	if err != nil {
		logger.Errorf("Failed to get user %s: %v", contact.UserRequest.RecipientID, err)
		return model.User{}, err
	}

	if contact.Alias != "" {
		user.DisplayName = contact.Alias
	}

	return user, nil
}

func (s *ProfileService) SearchProfile(ctx context.Context, userSearchRequest model.UserSearchRequest) ([]model.UserBriefInfo, error) {
	userSearchRequest.Nickname = cleanNickname(userSearchRequest.Nickname)
	return s.repo.SearchProfile(ctx, userSearchRequest)
}

func cleanNickname(nick string) string {
	re := regexp.MustCompile(`[^a-zA-Zа-яА-ЯёЁ]+`)

	cleaned := re.ReplaceAllString(nick, "")

	return cleaned
}
