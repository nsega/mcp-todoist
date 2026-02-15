package models

// Section represents a Todoist section within a project.
type Section struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Order     int    `json:"order,omitempty"`
}
