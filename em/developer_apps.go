package em

// CreateDeveloperAppRequest represents a request to create a developer app
type CreateDeveloperAppRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"` // max 128 chars
	ImageURL    *string `json:"imageUrl,omitempty"`    // max 2000 chars, must be http/https URL
	UserID      string  `json:"userId"`
}

// UpdateDeveloperAppRequest represents a request to update a developer app
type UpdateDeveloperAppRequest struct {
	AppAPIKey   string  `json:"appApiKey"` // must be valid API key
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"` // max 128 chars
	ImageURL    *string `json:"imageUrl,omitempty"`    // max 2000 chars, must be http/https URL
	UserID      string  `json:"userId"`
}

// DeleteDeveloperAppRequest represents a request to delete a developer app
type DeleteDeveloperAppRequest struct {
	UserID    string `json:"userId"`
	AppAPIKey string `json:"appApiKey"` // must be valid API key
}
