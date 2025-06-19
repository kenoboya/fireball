package service

import (
	"chat-api/internal/model"
	repo "chat-api/internal/repository/psql"
	"chat-api/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"
)

type ChatService struct {
	repoMessages  repo.Messages
	repoChats     repo.Chats
	repoMedia     repo.Media
	repoFiles     repo.Files
	repoLocations repo.Locations
	repoPinned    repo.Pinned
}

func NewChatService(ms repo.Messages, ct repo.Chats, md repo.Media, f repo.Files, l repo.Locations, pin repo.Pinned) *ChatService {
	return &ChatService{
		repoMessages:  ms,
		repoChats:     ct,
		repoMedia:     md,
		repoFiles:     f,
		repoLocations: l,
		repoPinned:    pin,
	}
}

func (s *ChatService) CreatePrivateChat(ctx context.Context, request model.CreatePrivateChatRequest) (model.CreatePrivateChatResponse, error) {
	var response model.CreatePrivateChatResponse
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, 5)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if request.Chat.CreatorID == "" || request.Chat.Name == "" || request.Chat.Type == "" {
		return response, model.ErrInvalidParamsOfChat
	}

	if request.InitialMessage.MessageWithData.MessageDB.SenderID == "" || request.InitialMessage.MessageWithData.MessageDB.Status == "" || request.InitialMessage.MessageWithData.MessageDB.Type == "" {
		return response, model.ErrInvalidParamsOfMessage
	}

	messageID, messageCreatedAt, err := s.repoMessages.SetMessage(ctx, request.InitialMessage.MessageWithData.MessageDB)
	if err != nil {
		return response, err
	}

	var chatID int64
	var chatCreatedAt time.Time

	runParallel := func(task func() error) {
		wg.Add(1)
		// Run upload task in a goroutine
		go func() {
			defer wg.Done()
			if err := task(); err != nil {
				if errors.Is(err, model.ErrUploadFile) {
					return
				}
				select {
				case errChan <- err:
				default:
				}
				cancel()
			}
		}()
	}

	runParallel(func() error {
		if request.InitialMessage.MessageWithData.MessageDB.Type == model.MESSAGE_FILE || request.InitialMessage.MessageWithData.MessageDB.Type == model.MESSAGE_MIXED {
			if request.InitialMessage.MessageWithData.Files == nil {
				return model.ErrFilesIsEmpty
			}
			for _, file := range *request.InitialMessage.MessageWithData.Files {
				fileID, err := s.repoFiles.SetFile(ctx, file)
				if err != nil {
					return model.ErrUploadFile
				}
				if err = s.repoMessages.SetBindMessageFile(ctx, messageID, fileID); err != nil {
					return err
				}
				file.FileID = fileID

				mu.Lock()
				if response.Message.MessageWithData.Files == nil {
					response.Message.MessageWithData.Files = new([]model.File)
				}
				*response.Message.MessageWithData.Files = append(*response.Message.MessageWithData.Files, file)
				mu.Unlock()
			}
		}
		return nil
	})

	runParallel(func() error {
		if request.InitialMessage.MessageWithData.MessageDB.Type == model.MESSAGE_MEDIA || request.InitialMessage.MessageWithData.MessageDB.Type == model.MESSAGE_MIXED {
			if request.InitialMessage.MessageWithData.Media == nil {
				return model.ErrMediaIsEmpty
			}
			for _, media := range *request.InitialMessage.MessageWithData.Media {
				mediaID, err := s.repoMedia.SetMedia(ctx, media)
				if err != nil {
					return model.ErrUploadFile
				}
				if err = s.repoMessages.SetBindMessageMedia(ctx, messageID, mediaID); err != nil {
					return err
				}
				media.MediaID = mediaID

				mu.Lock()
				if response.Message.MessageWithData.Media == nil {
					response.Message.MessageWithData.Media = new([]model.Media)
				}
				*response.Message.MessageWithData.Media = append(*response.Message.MessageWithData.Media, media)
				mu.Unlock()
			}
		}
		return nil
	})

	runParallel(func() error {
		if request.InitialMessage.MessageWithData.MessageDB.Type == model.MESSAGE_LOCATION || request.InitialMessage.MessageWithData.MessageDB.Type == model.MESSAGE_MIXED {
			if request.InitialMessage.MessageWithData.Locations == nil {
				return model.ErrLocationIsEmpty
			}
			for _, location := range *request.InitialMessage.MessageWithData.Locations {
				locationID, err := s.repoLocations.SetLocation(ctx, location)
				if err != nil {
					return model.ErrUploadLocation
				}
				if err = s.repoMessages.SetBindMessageLocation(ctx, messageID, locationID); err != nil {
					return err
				}
				location.LocationID = locationID

				mu.Lock()
				if response.Message.MessageWithData.Locations == nil {
					response.Message.MessageWithData.Locations = new([]model.Location)
				}
				*response.Message.MessageWithData.Locations = append(*response.Message.MessageWithData.Locations, location)
				mu.Unlock()
			}
		}
		return nil
	})

	runParallel(func() error {
		tmpChatID, tmpCreatedAt, err := s.repoChats.SetChat(ctx, request.Chat)
		if err == nil {
			mu.Lock()
			chatID = tmpChatID
			chatCreatedAt = tmpCreatedAt
			mu.Unlock()
		}
		return err
	})

	wg.Wait()
	close(errChan)

	var firstErr error
	for err := range errChan {
		if firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		return response, firstErr
	}

	var postWg sync.WaitGroup
	postErrChan := make(chan error, 2)

	runPost := func(task func() error) {
		postWg.Add(1)
		// Run post-creation task in a goroutine
		go func() {
			defer postWg.Done()
			if err := task(); err != nil {
				select {
				case postErrChan <- err:
				default:
				}
				cancel()
			}
		}()
	}

	runPost(func() error {
		return s.repoMessages.SetBindMessageChat(ctx, messageID, chatID)
	})

	runPost(func() error {
		if err := s.repoChats.SetParticipant(ctx, chatID, request.Chat.CreatorID); err != nil {
			return err
		}
		if err := s.repoChats.SetParticipant(ctx, chatID, request.RecipientID); err != nil {
			return err
		}
		return nil
	})

	postWg.Wait()
	close(postErrChan)

	for err := range postErrChan {
		if err != nil {
			return response, err
		}
	}

	request.InitialMessage.MessageWithData.MessageDB.MessageID = messageID
	request.InitialMessage.MessageWithData.MessageDB.CreatedAt = messageCreatedAt
	request.Chat.ChatID = chatID
	request.Chat.CreatedAt = chatCreatedAt

	response.RecipientID = request.RecipientID
	response.Message.MessageWithData.MessageDB = request.InitialMessage.MessageWithData.MessageDB
	response.Chat = request.Chat

	return response, nil
}

func (s *ChatService) CreateGroupChat(ctx context.Context, request *model.CreateGroupChatRequest) error {
	var (
		wg            sync.WaitGroup
		chatID        int64
		chatCreatedAt time.Time
	)

	// Buffered error channel for up to 2 errors
	errChan := make(chan error, 2)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Validate input
	if request.Chat.CreatorID == "" || request.Chat.Name == "" || request.Chat.Type == "" {
		return model.ErrInvalidParamsOfChat
	}

	if request.ChatAction.UserID == "" {
		return model.ErrInvalidParamsOfMessage
	}

	request.Chat.Type = model.CHAT_ACTION_CREATE

	// Save chat and capture generated ID and timestamp
	chatID, chatCreatedAt, err := s.repoChats.SetChat(ctx, request.Chat)
	if err != nil {
		return err
	}
	request.Chat.CreatedAt = chatCreatedAt
	request.Chat.UpdatedAt = chatCreatedAt
	request.ChatAction.ChatID = chatID

	runParallel := func(task func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := task(); err != nil {
				select {
				case errChan <- err:
					cancel()
				default:
					cancel()
				}
			}
		}()
	}

	// Participants
	for _, participantID := range request.ParticipantsIDs {
		pid := participantID
		runParallel(func() error {
			return s.repoChats.SetParticipant(ctx, chatID, pid)
		})
	}

	// Save chat action
	runParallel(func() error {
		return s.repoChats.SetAction(ctx, request.ChatAction)
	})

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	request.Chat.ChatID = chatID
	request.Chat.CreatedAt = chatCreatedAt
	request.ChatAction.ChatID = chatID
	return nil
}

func (s *ChatService) loadChatDetails(ctx context.Context, chatDB model.ChatDB) (model.Chat, error) {
	var (
		chat     model.Chat
		messages []model.Message
		muMsg    sync.Mutex
		wgMsg    sync.WaitGroup

		wgMeta  sync.WaitGroup
		metaErr error
		metaMu  sync.Mutex

		participantsIDs []string
		actions         *[]model.ChatAction
		roles           *[]model.ChatRole
	)

	chat.ChatDB = chatDB

	wgMeta.Add(3)

	// 1. Get participants
	go func() {
		defer wgMeta.Done()
		result, err := s.repoChats.GetAllParticipantsByChatID(ctx, chatDB.ChatID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			metaMu.Lock()
			metaErr = err
			metaMu.Unlock()
			return
		}
		if err == nil {
			participantsIDs = result
		}
	}()

	// 2. Get chat actions
	go func() {
		defer wgMeta.Done()
		result, err := s.repoChats.GetAllActionsWithLimit(ctx, chatDB.ChatID, model.ACTION_LIMIT_REQUEST)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			metaMu.Lock()
			metaErr = err
			metaMu.Unlock()
			return
		}
		if err == nil {
			actions = &result
		}
	}()

	// 3. Get chat roles
	go func() {
		defer wgMeta.Done()
		result, err := s.repoChats.GetAllChatRoles(ctx, chatDB.ChatID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			metaMu.Lock()
			metaErr = err
			metaMu.Unlock()
			return
		}
		if err == nil {
			roles = &result
		}
	}()

	messagesDB, err := s.repoMessages.GetMessagesByChatIDWithLimit(ctx, chatDB.ChatID, model.MESSAGE_LIMIT_REQUEST)
	if err != nil {
		return chat, err
	}

	for _, messageDB := range messagesDB {
		wgMsg.Add(1)
		go func(messageDB model.MessageDB) {
			defer wgMsg.Done()
			var message model.Message
			message.MessageWithData.MessageDB = messageDB

			var innerWg sync.WaitGroup
			innerWg.Add(5)

			// 1. Media
			go func() {
				defer innerWg.Done()
				media, err := s.repoMedia.GetMediaFileByMessageID(ctx, messageDB.MessageID)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					logger.Warnf("Failed to load media, err: %s", err)
					return
				}
				if err == nil {
					message.MessageWithData.Media = &media
				}
			}()

			// 2. Location
			go func() {
				defer innerWg.Done()
				loc, err := s.repoLocations.GetLocationsByMessageID(ctx, messageDB.MessageID)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					logger.Warnf("Failed to load location, err: %s", err)
					return
				}
				if err == nil {
					message.MessageWithData.Locations = &loc
				}
			}()

			// 3. Files
			go func() {
				defer innerWg.Done()
				files, err := s.repoFiles.GetFilesByMessageID(ctx, messageDB.MessageID)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					logger.Warnf("Failed to load files, err: %s", err)
					return
				}
				if err == nil {
					message.MessageWithData.Files = &files
				}
			}()

			// 4. PinnedMessage
			go func() {
				defer innerWg.Done()
				pinned, err := s.repoPinned.GetPinnedMessageByMessageID(ctx, messageDB.MessageID)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					logger.Warnf("Failed to load pinned messages, err: %s", err)
					return
				}
				if err == nil {
					message.PinnedMessage = &pinned
				}
			}()

			// 5. Actions
			go func() {
				defer innerWg.Done()
				acts, err := s.repoMessages.GetAllActions(ctx, messageDB.MessageID)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					logger.Warnf("Failed to load message action, err: %s", err)
					return
				}
				if err == nil {
					message.Action = &acts
				}
			}()

			innerWg.Wait()

			muMsg.Lock()
			messages = append(messages, message)
			muMsg.Unlock()
		}(messageDB)
	}

	wgMsg.Wait()

	wgMeta.Wait()
	if metaErr != nil {
		return chat, metaErr
	}

	chat.Messages = messages
	chat.ParticipantsIDs = participantsIDs
	chat.ChatAction = actions
	chat.ChatRoles = roles

	return chat, nil
}

func (s *ChatService) InitializeChatsForMessenger(ctx context.Context, userID string) ([]model.Chat, error) {
	var (
		chats   []model.Chat
		muChat  sync.Mutex
		errChan = make(chan error, 1)
		wgChat  sync.WaitGroup
	)

	chatsDB, err := s.repoChats.GetAllChatsByUserIDWithLimit(ctx, userID, model.CHAT_LIMIT_REQUEST)
	if err != nil {
		return nil, err
	}

	for _, chatDB := range chatsDB {
		wgChat.Add(1)
		go func(chatDB model.ChatDB) {
			defer wgChat.Done()

			chat, err := s.loadChatDetails(ctx, chatDB)
			if err != nil {
				select {
				case errChan <- err:
				default:
				}
				return
			}

			muChat.Lock()
			chats = append(chats, chat)
			muChat.Unlock()
		}(chatDB)
	}

	done := make(chan struct{})
	go func() {
		wgChat.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errChan:
		return nil, err
	case <-done:
		return chats, nil
	}
}

func (s *ChatService) InitializePinnedChatsForMessenger(ctx context.Context, userID string) ([]model.PinnedChatInit, error) {
	var (
		chats   []model.PinnedChatInit
		muChat  sync.Mutex
		errChan = make(chan error, 1)
		wgChat  sync.WaitGroup
	)

	pinnedChats, err := s.repoPinned.GetPinnedChatsByUserIDWithLimit(ctx, userID, model.PINNED_CHAT_LIMIT_REQUEST)
	if err != nil {
		return nil, err
	}

	for _, pinnedChat := range pinnedChats {
		wgChat.Add(1)
		go func(pinned model.PinnedChat) {
			defer wgChat.Done()

			chatDB, err := s.repoChats.GetChatByChatID(ctx, pinned.ChatID)
			if err != nil {
				select {
				case errChan <- err:
				default:
				}
				return
			}

			chat, err := s.loadChatDetails(ctx, chatDB)
			if err != nil {
				select {
				case errChan <- err:
				default:
				}
				return
			}

			pinnedChatInit := model.PinnedChatInit{
				Chat:       chat,
				PinnedChat: pinned,
			}

			muChat.Lock()
			chats = append(chats, pinnedChatInit)
			muChat.Unlock()
		}(pinnedChat)
	}

	done := make(chan struct{})
	go func() {
		wgChat.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errChan:
		return nil, err
	case <-done:
		return chats, nil
	}
}

func (s *ChatService) GetParticipantsOfChat(ctx context.Context, chatID int64) ([]string, error) {
	return s.repoChats.GetAllParticipantsByChatID(ctx, chatID)
}

func (s *ChatService) GetChatByChatID(ctx context.Context, chatID int64) (model.ChatDB, error) {
	return s.repoChats.GetChatByChatID(ctx, chatID)
}

func (s *ChatService) UpdatePinnedChat(ctx context.Context, pinnedChatWithFlag model.PinnedChatWithFlag) error {
	exists, err := s.repoPinned.IsPinnedChatExists(ctx, pinnedChatWithFlag.PinnedChat)
	if err != nil {
		return err
	}

	if exists {
		if !pinnedChatWithFlag.Fix {
			return s.repoPinned.DeletePinnedChat(ctx, pinnedChatWithFlag.PinnedChat)
		}
		// if flag is true
		return s.repoPinned.UpdatePinnedChat(ctx, pinnedChatWithFlag.PinnedChat)
	}

	if pinnedChatWithFlag.Fix {
		return s.repoPinned.SetPinnedChat(ctx, pinnedChatWithFlag.PinnedChat)
	}
	return nil
}

func (s *ChatService) SetChatRole(ctx context.Context, chatRole model.ChatRole) error {
	return s.repoChats.SetChatRole(ctx, chatRole)
}

func (s *ChatService) SetBlockChat(ctx context.Context, blockChat model.BlockChat) error {
	exists, err := s.repoChats.IsBlockedChatExists(ctx, blockChat.ChatID, blockChat.UserID)
	if err != nil {
		return err
	}

	if blockChat.Blocked {
		if !exists {
			return s.repoChats.SetBlockChat(ctx, blockChat.ChatID, blockChat.UserID)
		}
	} else {
		if exists {
			return s.repoChats.DeleteBlockUser(ctx, blockChat.ChatID, blockChat.UserID)
		}
	}
	return nil
}
