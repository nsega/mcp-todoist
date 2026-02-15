package models

import "time"

// Task represents a Todoist task.
type Task struct {
	ID           string    `json:"id"`
	Content      string    `json:"content"`
	Description  string    `json:"description"`
	ProjectID    string    `json:"project_id"`
	SectionID    string    `json:"section_id,omitempty"`
	ParentID     string    `json:"parent_id,omitempty"`
	Labels       []string  `json:"labels,omitempty"`
	Priority     int       `json:"priority"`
	Order        int       `json:"order,omitempty"`
	Due          *DueDate  `json:"due,omitempty"`
	URL          string    `json:"url,omitempty"`
	CommentCount int       `json:"comment_count,omitempty"`
	IsCompleted  bool      `json:"is_completed"`
	CreatorID    string    `json:"creator_id,omitempty"`
	AssigneeID   string    `json:"assignee_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	Duration     *Duration `json:"duration,omitempty"`
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
