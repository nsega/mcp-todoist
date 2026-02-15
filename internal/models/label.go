package models

// Label represents a Todoist label.
type Label struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color,omitempty"`
	Order      int    `json:"order,omitempty"`
	IsFavorite bool   `json:"is_favorite"`
}
