package service

import (
	"chat-api/internal/model"
	repo "chat-api/internal/repository/psql"
	"context"
	"errors"
	"sync"
)

type MessageService struct {
	repoMessages  repo.Messages
	repoMedia     repo.Media
	repoFiles     repo.Files
	repoLocations repo.Locations
}

func NewMessageService(
	repoMessages repo.Messages,
	repoFiles repo.Files,
	repoMedia repo.Media,
	repoLocations repo.Locations,
) *MessageService {
	return &MessageService{
		repoMessages:  repoMessages,
		repoFiles:     repoFiles,
		repoMedia:     repoMedia,
		repoLocations: repoLocations,
	}
}

func (s *MessageService) SendMessage(ctx context.Context, createMessageRequest *model.CreateMessageRequest) error {
	var err error
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, 4)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	messageID, createMessageTime, err := s.repoMessages.SetMessage(ctx, createMessageRequest.MessageWithData.MessageDB)
	if err != nil {
		return err
	}

	createMessageRequest.MessageWithData.MessageDB.CreatedAt = createMessageTime
	createMessageRequest.MessageWithData.MessageDB.UpdatedAt = createMessageTime
	createMessageRequest.MessageWithData.MessageDB.MessageID = messageID

	runParallel := func(task func() error) {
		wg.Add(1)
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
		if createMessageRequest.MessageWithData.MessageDB.Type == model.MESSAGE_FILE ||
			createMessageRequest.MessageWithData.MessageDB.Type == model.MESSAGE_MIXED {
			if createMessageRequest.MessageWithData.Files == nil {
				return model.ErrFilesIsEmpty
			}
			for i := range *createMessageRequest.MessageWithData.Files {
				file := &(*createMessageRequest.MessageWithData.Files)[i]
				fileID, err := s.repoFiles.SetFile(ctx, *file)
				if err != nil {
					return model.ErrUploadFile
				}
				if err = s.repoMessages.SetBindMessageFile(ctx, messageID, fileID); err != nil {
					return err
				}
				mu.Lock()
				file.FileID = fileID
				mu.Unlock()
			}
		}
		return nil
	})

	runParallel(func() error {
		if createMessageRequest.MessageWithData.MessageDB.Type == model.MESSAGE_MEDIA ||
			createMessageRequest.MessageWithData.MessageDB.Type == model.MESSAGE_MIXED {
			if createMessageRequest.MessageWithData.Media == nil {
				return model.ErrMediaIsEmpty
			}
			for i := range *createMessageRequest.MessageWithData.Media {
				media := &(*createMessageRequest.MessageWithData.Media)[i]
				mediaID, err := s.repoMedia.SetMedia(ctx, *media)
				if err != nil {
					return model.ErrUploadFile
				}
				if err = s.repoMessages.SetBindMessageMedia(ctx, messageID, mediaID); err != nil {
					return err
				}
				mu.Lock()
				media.MediaID = mediaID
				mu.Unlock()
			}
		}
		return nil
	})

	runParallel(func() error {
		if createMessageRequest.MessageWithData.MessageDB.Type == model.MESSAGE_LOCATION ||
			createMessageRequest.MessageWithData.MessageDB.Type == model.MESSAGE_MIXED {
			if createMessageRequest.MessageWithData.Locations == nil {
				return model.ErrLocationIsEmpty
			}
			for i := range *createMessageRequest.MessageWithData.Locations {
				location := &(*createMessageRequest.MessageWithData.Locations)[i]
				locationID, err := s.repoLocations.SetLocation(ctx, *location)
				if err != nil {
					return model.ErrUploadLocation
				}
				if err = s.repoMessages.SetBindMessageLocation(ctx, messageID, locationID); err != nil {
					return err
				}
				mu.Lock()
				location.LocationID = locationID
				mu.Unlock()
			}
		}
		return nil
	})

	runParallel(func() error {
		return s.repoMessages.SetBindMessageChat(ctx, messageID, createMessageRequest.ChatID)
	})

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}
