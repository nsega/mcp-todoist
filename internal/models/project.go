package models

// Project represents a Todoist project.
type Project struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Color          string `json:"color,omitempty"`
	ParentID       string `json:"parent_id,omitempty"`
	Order          int    `json:"child_order,omitempty"`
	IsShared       bool   `json:"is_shared"`
	IsFavorite     bool   `json:"is_favorite"`
	IsInboxProject bool   `json:"inbox_project"`
	IsTeamInbox    bool   `json:"is_team_inbox,omitempty"`
	ViewStyle      string `json:"view_style,omitempty"`
	URL            string `json:"url,omitempty"`
	CreatorUID     string `json:"creator_uid,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
	IsArchived     bool   `json:"is_archived,omitempty"`
	IsDeleted      bool   `json:"is_deleted,omitempty"`
	Description    string `json:"description,omitempty"`
	CanAssignTasks bool   `json:"can_assign_tasks,omitempty"`
}
