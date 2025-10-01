package em

import "encoding/json"

// UserEvents represents user event metadata
type UserEvents struct {
	Referrer     *string `json:"referrer,omitempty"`
	IsMobileUser *bool   `json:"isMobileUser,omitempty"`
}

// CreateUserMetadata represents the metadata for creating a user
type CreateUserMetadata struct {
	AllowAiAttribution  *bool       `json:"allowAiAttribution,omitempty"`
	Bio                 *string     `json:"bio,omitempty"`
	CoverPhotoSizes     *string     `json:"coverPhotoSizes,omitempty"`
	Donation            *string     `json:"donation,omitempty"`
	Handle              *string     `json:"handle,omitempty"`
	Events              *UserEvents `json:"events,omitempty"`
	Location            *string     `json:"location,omitempty"`
	Name                *string     `json:"name,omitempty"`
	ProfilePictureSizes *string     `json:"profilePictureSizes,omitempty"`
	SplUsdcPayoutWallet *string     `json:"splUsdcPayoutWallet,omitempty"`
	Wallet              string      `json:"wallet"`
	Website             *string     `json:"website,omitempty"`
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Metadata CreateUserMetadata `json:"metadata"`
}

// CreateAssociatedWallets represents associated wallet signatures
type CreateAssociatedWallets map[string]struct {
	Signature string `json:"signature"`
}

// CollectiblesMetadata represents collectibles metadata
type CollectiblesMetadata struct {
	Order []string               `json:"order,omitempty"`
	Extra map[string]interface{} `json:",inline"`
}

// PlaylistIdentifier represents a playlist identifier
type PlaylistIdentifier struct {
	Type       string `json:"type"` // "playlist"
	PlaylistID int    `json:"playlist_id"`
}

// ExplorePlaylistIdentifier represents an explore playlist identifier
type ExplorePlaylistIdentifier struct {
	Type       string `json:"type"` // "explore_playlist"
	PlaylistID string `json:"playlist_id"`
}

// PlaylistLibraryIdentifier can be either a regular or explore playlist
type PlaylistLibraryIdentifier interface{}

// PlaylistLibraryFolder represents a folder in the playlist library
type PlaylistLibraryFolder struct {
	ID       string                      `json:"id"`
	Type     string                      `json:"type"` // "folder"
	Name     string                      `json:"name"`
	Contents []PlaylistLibraryIdentifier `json:"contents"`
}

// PlaylistLibrary represents the user's playlist library
type PlaylistLibrary struct {
	Contents []PlaylistLibraryIdentifier `json:"contents"`
}

// UpdateProfileMetadata represents metadata for updating a profile
type UpdateProfileMetadata struct {
	Name                *string          `json:"name,omitempty"`
	Handle              *string          `json:"handle,omitempty"`
	Bio                 *string          `json:"bio,omitempty"`
	Website             *string          `json:"website,omitempty"`
	Donation            *string          `json:"donation,omitempty"`
	Location            *string          `json:"location,omitempty"`
	ProfileType         *string          `json:"profileType,omitempty"` // nullable, can be "label"
	MetadataMultihash   *string          `json:"metadataMultihash,omitempty"`
	Events              *UserEvents      `json:"events,omitempty"`
	IsDeactivated       *bool            `json:"isDeactivated,omitempty"`
	ArtistPickTrackID   *string          `json:"artistPickTrackId,omitempty"`
	AllowAiAttribution  *bool            `json:"allowAiAttribution,omitempty"`
	PlaylistLibrary     *PlaylistLibrary `json:"playlistLibrary,omitempty"`
	TwitterHandle       *string          `json:"twitterHandle,omitempty"`
	InstagramHandle     *string          `json:"instagramHandle,omitempty"`
	TiktokHandle        *string          `json:"tiktokHandle,omitempty"`
	SplUsdcPayoutWallet *string          `json:"splUsdcPayoutWallet,omitempty"` // nullable
}

// UpdateProfileRequest represents a request to update a user profile
type UpdateProfileRequest struct {
	UserID   string                `json:"userId"`
	Events   *UserEvents           `json:"events,omitempty"`
	Metadata UpdateProfileMetadata `json:"metadata"`
}

// FollowUserRequest represents a request to follow a user
type FollowUserRequest struct {
	UserID         string `json:"userId"`
	FolloweeUserID string `json:"followeeUserId"`
}

// UnfollowUserRequest represents a request to unfollow a user
type UnfollowUserRequest struct {
	UserID         string `json:"userId"`
	FolloweeUserID string `json:"followeeUserId"`
}

// SubscribeToUserRequest represents a request to subscribe to a user
type SubscribeToUserRequest struct {
	UserID           string `json:"userId"`
	SubscribeeUserID string `json:"subscribeeUserId"`
}

// UnsubscribeFromUserRequest represents a request to unsubscribe from a user
type UnsubscribeFromUserRequest struct {
	UserID           string `json:"userId"`
	SubscribeeUserID string `json:"subscribeeUserId"`
}

// SendTipRequest represents a request to send a tip
type SendTipRequest struct {
	Amount         int    `json:"amount"` // positive integer
	SenderUserID   string `json:"senderUserId"`
	ReceiverUserID string `json:"receiverUserId"`
}

// SendTipReactionMetadata represents metadata for a tip reaction
type SendTipReactionMetadata struct {
	ReactedTo     string `json:"reactedTo"`
	ReactionValue int    `json:"reactionValue"`
}

// SendTipReactionRequest represents a request to send a tip reaction
type SendTipReactionRequest struct {
	UserID   string                  `json:"userId"`
	Metadata SendTipReactionMetadata `json:"metadata"`
}

// EmailRequest represents an email-related request
type EmailRequest struct {
	EmailOwnerUserID           int      `json:"emailOwnerUserId"`
	ReceivingUserID            int      `json:"receivingUserId"`
	InitialEmailEncryptionUUID int      `json:"initialEmailEncryptionUuid"`
	GranteeUserIDs             []string `json:"granteeUserIds,omitempty"`
	Email                      string   `json:"email"`
}

// Wallet represents a blockchain wallet
type Wallet struct {
	Address string `json:"address"`
	Chain   string `json:"chain"` // "sol" or "eth"
}

// AddAssociatedWalletRequest represents a request to add an associated wallet
type AddAssociatedWalletRequest struct {
	UserID    string `json:"userId"`
	Wallet    Wallet `json:"wallet"`
	Signature string `json:"signature"`
}

// RemoveAssociatedWalletRequest represents a request to remove an associated wallet
type RemoveAssociatedWalletRequest struct {
	UserID string `json:"userId"`
	Wallet Wallet `json:"wallet"`
}

// UpdateCollectiblesRequest represents a request to update collectibles
type UpdateCollectiblesRequest struct {
	UserID       string                `json:"userId"`
	Collectibles *CollectiblesMetadata `json:"collectibles"`
}

// UnmarshalJSON custom unmarshaler for CollectiblesMetadata
func (c *CollectiblesMetadata) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	type Alias CollectiblesMetadata
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	return nil
}
