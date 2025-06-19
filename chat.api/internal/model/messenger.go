package model

type MessengerInitializer struct {
	Chats        []Chat
	PinnedChat   []PinnedChatInit
	UsersProfile []UserBriefInfo
}
