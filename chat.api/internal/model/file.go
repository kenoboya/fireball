package model

import "time"

type File struct {
	FileID     int64     `json:"file_id,omitempty" db:"file_id"`
	URL        string    `json:"url" db:"url"`
	Type       string    `json:"type" db:"type"`
	Size       int64     `json:"size,omitempty" db:"size"`
	UploadedAt time.Time `json:"uploaded_at,omitempty" db:"uploaded_at"`
}

type MessageFile struct {
	MessageID int64
	FileID    int64
}
