package models

import "time"

// Task represents a Todoist task.
type Task struct {
	ID            string    `json:"id"`
	Content       string    `json:"content"`
	Description   string    `json:"description"`
	ProjectID     string    `json:"project_id"`
	SectionID     string    `json:"section_id,omitempty"`
	ParentID      string    `json:"parent_id,omitempty"`
	Labels        []string  `json:"labels,omitempty"`
	Priority      int       `json:"priority"`
	Order         int       `json:"child_order,omitempty"`
	Due           *DueDate  `json:"due,omitempty"`
	URL           string    `json:"url,omitempty"`
	CommentCount  int       `json:"note_count,omitempty"`
	IsCompleted   bool      `json:"checked"`
	CreatorID     string    `json:"added_by_uid,omitempty"`
	AssigneeID    string    `json:"responsible_uid,omitempty"`
	CreatedAt     time.Time `json:"added_at"`
	Duration      *Duration `json:"duration,omitempty"`
	UserID        string    `json:"user_id,omitempty"`
	AssignedByUID string    `json:"assigned_by_uid,omitempty"`
	UpdatedAt     string    `json:"updated_at,omitempty"`
	CompletedAt   string    `json:"completed_at,omitempty"`
	IsDeleted     bool      `json:"is_deleted,omitempty"`
}

// DueDate represents a task's due date.
type DueDate struct {
	Date      string `json:"date"`
	String    string `json:"string"`
	Datetime  string `json:"datetime,omitempty"`
	Recurring bool   `json:"recurring"`
	Timezone  string `json:"timezone,omitempty"`
}

// Duration represents a task's duration.
type Duration struct {
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
}
