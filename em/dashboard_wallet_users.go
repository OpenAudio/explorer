package em

// SignatureData represents signature data for wallet or user
type SignatureData struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

// CreateDashboardWalletUserRequest represents a request to create a dashboard wallet user
type CreateDashboardWalletUserRequest struct {
	Wallet          string         `json:"wallet"` // Ethereum address
	UserID          string         `json:"userId"`
	WalletSignature *SignatureData `json:"walletSignature,omitempty"` // Message format: "Connecting Audius user id a93jl at 39823489" OR "Connecting Audius user @jill1990 at 39823489"
	UserSignature   *SignatureData `json:"userSignature,omitempty"`   // Message format: "Connecting Audius protocol dashboard wallet 0x6c9CA7D9580d4e8286B0628c0300A2A1235a8e2E at 39823489"
}

// DeleteDashboardWalletUserRequest represents a request to delete a dashboard wallet user
type DeleteDashboardWalletUserRequest struct {
	UserID string `json:"userId"`
	Wallet string `json:"wallet"` // Ethereum address
}
