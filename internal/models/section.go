package models

// Section represents a Todoist section within a project.
type Section struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	Order       int    `json:"section_order,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	AddedAt     string `json:"added_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	IsArchived  bool   `json:"is_archived,omitempty"`
	IsDeleted   bool   `json:"is_deleted,omitempty"`
	IsCollapsed bool   `json:"is_collapsed,omitempty"`
}
