package models

import "time"

// Comment represents a Todoist comment.
type Comment struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id,omitempty"`
	ProjectID string    `json:"project_id,omitempty"`
	Content   string    `json:"content"`
	PostedAt  time.Time `json:"posted_at"`
	PostedUID string    `json:"posted_uid,omitempty"`
	IsDeleted bool      `json:"is_deleted,omitempty"`
}
