package model

import "time"

const (
	MEDIA_TYPE_JPEG = "image/jpeg"
	MEDIA_TYPE_PNG  = "image/png"
	MEDIA_TYPE_GIF  = "image/gif"
	MEDIA_TYPE_MP4  = "video/mp4"
	MEDIA_TYPE_WEBM = "video/webm"
	MEDIA_TYPE_WEBP = "image/webp"
)

type Media struct {
	MediaID    int64
	URL        string
	Type       string
	Size       int64
	UploadedAt time.Time
}

type MessageMedia struct {
	MessageID int64
	MediaID   int64
}
