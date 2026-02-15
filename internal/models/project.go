package models

// Project represents a Todoist project.
type Project struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Color          string `json:"color,omitempty"`
	ParentID       string `json:"parent_id,omitempty"`
	Order          int    `json:"order,omitempty"`
	CommentCount   int    `json:"comment_count,omitempty"`
	IsShared       bool   `json:"is_shared"`
	IsFavorite     bool   `json:"is_favorite"`
	IsInboxProject bool   `json:"is_inbox_project"`
	IsTeamInbox    bool   `json:"is_team_inbox"`
	ViewStyle      string `json:"view_style,omitempty"`
	URL            string `json:"url,omitempty"`
}
