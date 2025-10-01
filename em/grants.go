package em

// CreateGrantRequest represents a request to create a grant
type CreateGrantRequest struct {
	UserID    string `json:"userId"`
	AppAPIKey string `json:"appApiKey"` // must be valid API key
}

// AddManagerRequest represents a request to add a manager
type AddManagerRequest struct {
	UserID        string `json:"userId"`
	ManagerUserID string `json:"managerUserId"`
}

// RemoveManagerRequest represents a request to remove a manager
type RemoveManagerRequest struct {
	UserID        string `json:"userId"`
	ManagerUserID string `json:"managerUserId"`
}

// RevokeGrantRequest represents a request to revoke a grant
type RevokeGrantRequest struct {
	UserID    string `json:"userId"`
	AppAPIKey string `json:"appApiKey"` // must be valid API key
}

// ApproveGrantRequest represents a request to approve a grant
type ApproveGrantRequest struct {
	UserID        string `json:"userId"`
	GrantorUserID string `json:"grantorUserId"`
}
