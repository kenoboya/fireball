package repo

import (
	"chat-api/internal/model"
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	Pinned    Pinned
	Media     Media
	Locations Locations
	Files     Files
	Messages  Messages
	Chats     Chats
}

func NewRepositories(db *sqlx.DB) *Repositories {
	return &Repositories{
		Pinned:    NewPinnedRepo(db),
		Media:     NewMediaRepo(db),
		Locations: NewLocationsRepo(db),
		Files:     NewFilesRepo(db),
		Messages:  NewMessagesRepo(db),
		Chats:     NewChatsRepo(db),
	}
}

type Pinned interface {
	SetPinnedMessage(ctx context.Context, pinMessage model.PinnedMessage) error
	GetPinnedMessagesByChatID(ctx context.Context, chatID int64) ([]model.PinnedMessage, error)
	GetPinnedMessageByMessageID(ctx context.Context, messageID int64) (model.PinnedMessage, error)
	DeletePinnedMessage(ctx context.Context, pinMessage model.PinnedMessage) error

	SetPinnedChat(ctx context.Context, pinChat model.PinnedChat) error
	GetPinnedChatsByUserID(ctx context.Context, userID string) ([]model.PinnedChat, error)
	GetPinnedChatsByUserIDWithLimit(ctx context.Context, userID string, limit int) ([]model.PinnedChat, error)
	IsPinnedChatExists(ctx context.Context, pinChat model.PinnedChat) (bool, error)
	UpdatePinnedChat(ctx context.Context, pinChat model.PinnedChat) error
	DeletePinnedChat(ctx context.Context, pinChat model.PinnedChat) error
}

type Media interface {
	SetMedia(ctx context.Context, media model.Media) (mediaID int64, err error)
	GetMediaByMediaID(ctx context.Context, mediaID int64) (model.Media, error)
	GetMediaFileByMessageID(ctx context.Context, messageID int64) ([]model.Media, error)
	GetMediaByChatID(ctx context.Context, chatID int64) ([]model.Media, error)
	DeleteMedia(ctx context.Context, mediaID int64) error
}

type Locations interface {
	SetLocation(ctx context.Context, location model.Location) (locationID int64, err error)
	GetLocationByLocationID(ctx context.Context, locationID int64) (model.Location, error)
	GetLocationsByMessageID(ctx context.Context, messageID int64) ([]model.Location, error)
	GetLocationsByChatID(ctx context.Context, chatID int64) ([]model.Location, error)
	DeleteLocation(ctx context.Context, locationID int64) error
}

type Files interface {
	SetFile(ctx context.Context, file model.File) (fileID int64, err error)
	GetFileByFileID(ctx context.Context, fileID int64) (model.File, error)
	GetFilesByMessageID(ctx context.Context, messageID int64) ([]model.File, error)
	GetFilesByChatID(ctx context.Context, chatID int64) ([]model.File, error)
	DeleteFile(ctx context.Context, fileID int64) error
}

type Messages interface {
	SetMessage(ctx context.Context, message model.MessageDB) (messageID int64, createdAt time.Time, err error)
	SetAction(ctx context.Context, messageAction model.MessageAction) error
	GetAllActions(ctx context.Context, messageID int64) ([]model.MessageAction, error)
	GetMessageByMessageID(ctx context.Context, messageID int64) (model.MessageDB, error)
	GetMessagesByChatID(ctx context.Context, chatID int64) ([]model.MessageDB, error)
	GetMessagesByChatIDWithLimit(ctx context.Context, chatID int64, limit int) ([]model.MessageDB, error)
	// UpdateChat(ctx context.Context, chat model.Chat, chatID int64) error // НУЖНО ПОДУМАТЬ НУЖНО ЛИ ВОЗРАЩАТЬ ЕЩЁ ЧТО-ТО + ДОЛЖНА БЫТЬ ЛОГИКА ДЕЙСТВИЙ
	DeleteMessage(ctx context.Context, messageID int64) error

	SetBindMessageMedia(ctx context.Context, messageID, mediaID int64) error
	SetBindMessageLocation(ctx context.Context, messageID, locationID int64) error
	SetBindMessageFile(ctx context.Context, messageID, fileID int64) error
	SetBindMessageChat(ctx context.Context, messageID, chatID int64) error
}

type Chats interface {
	SetChat(ctx context.Context, chat model.ChatDB) (chatID int64, createdAt time.Time, err error)
	SetParticipant(ctx context.Context, chatID int64, userID string) error
	SetChatRole(ctx context.Context, chatRole model.ChatRole) error
	SetBlockChat(ctx context.Context, chatID int64, userID string) error
	SetAction(ctx context.Context, chatAction model.ChatAction) error
	GetAllActions(ctx context.Context, chatID int64) ([]model.ChatAction, error)
	GetAllActionsWithLimit(ctx context.Context, chatID int64, limit int) ([]model.ChatAction, error)
	GetChatByChatID(ctx context.Context, chatID int64) (model.ChatDB, error)
	GetAllChatRoles(ctx context.Context, chatID int64) ([]model.ChatRole, error)
	GetAllParticipantsByChatID(ctx context.Context, chatID int64) ([]string, error)
	GetAllChatsByUserID(ctx context.Context, userID string) ([]model.ChatDB, error)
	GetAllChatsByUserIDWithLimit(ctx context.Context, userID string, limit int) ([]model.ChatDB, error)
	IsBlockedChatExists(ctx context.Context, chatID int64, userID string) (bool, error)
	// UpdateChat(ctx context.Context, chat model.Chat, chatID int64) error // НУЖНО ПОДУМАТЬ НУЖНО ЛИ ВОЗРАЩАТЬ ЕЩЁ ЧТО-ТО + ДОЛЖНА БЫТЬ ЛОГИКА ДЕЙСТВИЙ
	DeleteBlockUser(ctx context.Context, chatID int64, userID string) error
	DeleteChat(ctx context.Context, chatID int64) error
}
