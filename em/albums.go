package em

import "time"

// GetAlbumRequest represents a request to get an album
type GetAlbumRequest struct {
	UserID  *string `json:"userId,omitempty"`
	AlbumID string  `json:"albumId"`
}

// GetAlbumsRequest represents a request to get multiple albums
type GetAlbumsRequest struct {
	UserID *string  `json:"userId,omitempty"`
	ID     []string `json:"id"`
}

// GetAlbumTracksRequest represents a request to get tracks from an album
type GetAlbumTracksRequest struct {
	AlbumID string `json:"albumId"`
}

// CreateAlbumMetadata represents metadata for creating an album
type CreateAlbumMetadata struct {
	AlbumName             string                    `json:"albumName"`
	IsPrivate             *bool                     `json:"isPrivate,omitempty"`
	Description           *string                   `json:"description,omitempty"` // max 1000 chars
	License               *string                   `json:"license,omitempty"`
	ReleaseDate           *time.Time                `json:"releaseDate,omitempty"`
	DDEXReleaseIDs        map[string]string         `json:"ddexReleaseIds,omitempty"`
	DDEXApp               *string                   `json:"ddexApp,omitempty"`
	UPC                   *string                   `json:"upc,omitempty"`
	Artists               []DDEXResourceContributor `json:"artists,omitempty"`
	CopyrightLine         *DDEXCopyright            `json:"copyrightLine,omitempty"`
	ProducerCopyrightLine *DDEXCopyright            `json:"producerCopyrightLine,omitempty"`
	ParentalWarningType   *string                   `json:"parentalWarningType,omitempty"`
	IsStreamGated         *bool                     `json:"isStreamGated,omitempty"`
	StreamConditions      *USDCPurchaseConditions   `json:"streamConditions,omitempty"`
	IsDownloadGated       *bool                     `json:"isDownloadGated,omitempty"`
	DownloadConditions    *USDCPurchaseConditions   `json:"downloadConditions,omitempty"`
	IsScheduledRelease    *bool                     `json:"isScheduledRelease,omitempty"`
}

// CreateAlbumRequest represents a request to create an album
type CreateAlbumRequest struct {
	AlbumID  *string             `json:"albumId,omitempty"`
	UserID   string              `json:"userId"`
	Metadata CreateAlbumMetadata `json:"metadata"`
	TrackIDs []string            `json:"trackIds,omitempty"`
}

// UploadAlbumMetadata extends CreateAlbumMetadata with required fields
type UploadAlbumMetadata struct {
	CreateAlbumMetadata
	Genre string  `json:"genre"` // required
	Mood  *string `json:"mood,omitempty"`
	Tags  *string `json:"tags,omitempty"`
}

// AlbumTrackMetadata represents track metadata within an album
type AlbumTrackMetadata struct {
	TrackMetadata
	Genre              *string     `json:"genre,omitempty"`              // Override to make optional
	Mood               *string     `json:"mood,omitempty"`               // Override to make optional
	Tags               *string     `json:"tags,omitempty"`               // Override to make optional
	IsStreamGated      *bool       `json:"isStreamGated,omitempty"`      // Override to make optional
	StreamConditions   interface{} `json:"streamConditions,omitempty"`   // Override to make optional
	IsDownloadable     *bool       `json:"isDownloadable,omitempty"`     // Override to make optional
	DownloadConditions interface{} `json:"downloadConditions,omitempty"` // Override to make optional
}

// UpdateAlbumMetadata represents metadata for updating an album
type UpdateAlbumMetadata struct {
	AlbumName             *string                   `json:"albumName,omitempty"`
	IsPrivate             *bool                     `json:"isPrivate,omitempty"`
	Description           *string                   `json:"description,omitempty"`
	License               *string                   `json:"license,omitempty"`
	ReleaseDate           *time.Time                `json:"releaseDate,omitempty"`
	DDEXReleaseIDs        map[string]string         `json:"ddexReleaseIds,omitempty"`
	DDEXApp               *string                   `json:"ddexApp,omitempty"`
	UPC                   *string                   `json:"upc,omitempty"`
	Artists               []DDEXResourceContributor `json:"artists,omitempty"`
	CopyrightLine         *DDEXCopyright            `json:"copyrightLine,omitempty"`
	ProducerCopyrightLine *DDEXCopyright            `json:"producerCopyrightLine,omitempty"`
	ParentalWarningType   *string                   `json:"parentalWarningType,omitempty"`
	IsStreamGated         *bool                     `json:"isStreamGated,omitempty"`
	StreamConditions      *USDCPurchaseConditions   `json:"streamConditions,omitempty"`
	IsDownloadGated       *bool                     `json:"isDownloadGated,omitempty"`
	DownloadConditions    *USDCPurchaseConditions   `json:"downloadConditions,omitempty"`
	IsScheduledRelease    *bool                     `json:"isScheduledRelease,omitempty"`
	Genre                 *string                   `json:"genre,omitempty"`
	Mood                  *string                   `json:"mood,omitempty"`
	Tags                  *string                   `json:"tags,omitempty"`
	PlaylistContents      []PlaylistContent         `json:"playlistContents,omitempty"`
}

// UploadAlbumRequest represents a request to upload an album with tracks
type UploadAlbumRequest struct {
	UserID         string               `json:"userId"`
	Metadata       UploadAlbumMetadata  `json:"metadata"`
	TrackMetadatas []AlbumTrackMetadata `json:"trackMetadatas"`
}

// UpdateAlbumRequest represents a request to update an album
type UpdateAlbumRequest struct {
	UserID   string              `json:"userId"`
	AlbumID  string              `json:"albumId"`
	Metadata UpdateAlbumMetadata `json:"metadata"`
}

// DeleteAlbumRequest represents a request to delete an album
type DeleteAlbumRequest struct {
	UserID  string `json:"userId"`
	AlbumID string `json:"albumId"`
}

// FavoriteAlbumMetadata represents metadata for favoriting an album
type FavoriteAlbumMetadata struct {
	IsSaveOfRepost bool `json:"isSaveOfRepost"`
}

// FavoriteAlbumRequest represents a request to favorite an album
type FavoriteAlbumRequest struct {
	UserID   string                 `json:"userId"`
	AlbumID  string                 `json:"albumId"`
	Metadata *FavoriteAlbumMetadata `json:"metadata,omitempty"`
}

// UnfavoriteAlbumRequest represents a request to unfavorite an album
type UnfavoriteAlbumRequest struct {
	UserID  string `json:"userId"`
	AlbumID string `json:"albumId"`
}

// RepostAlbumMetadata represents metadata for reposting an album
type RepostAlbumMetadata struct {
	IsRepostOfRepost bool `json:"isRepostOfRepost"`
}

// RepostAlbumRequest represents a request to repost an album
type RepostAlbumRequest struct {
	UserID   string               `json:"userId"`
	AlbumID  string               `json:"albumId"`
	Metadata *RepostAlbumMetadata `json:"metadata,omitempty"`
}

// UnrepostAlbumRequest represents a request to unrepost an album
type UnrepostAlbumRequest struct {
	UserID  string `json:"userId"`
	AlbumID string `json:"albumId"`
}

// GetPurchaseAlbumInstructionsRequest represents a request to get album purchase instructions
type GetPurchaseAlbumInstructionsRequest struct {
	UserID            string      `json:"userId"`
	AlbumID           string      `json:"albumId"`
	Price             interface{} `json:"price"`                 // number or bigint
	ExtraAmount       interface{} `json:"extraAmount,omitempty"` // number or bigint
	IncludeNetworkCut *bool       `json:"includeNetworkCut,omitempty"`
}

// PurchaseAlbumRequest represents a request to purchase an album
type PurchaseAlbumRequest struct {
	UserID            string      `json:"userId"`
	AlbumID           string      `json:"albumId"`
	Price             interface{} `json:"price"`                 // number or bigint
	ExtraAmount       interface{} `json:"extraAmount,omitempty"` // number or bigint
	IncludeNetworkCut *bool       `json:"includeNetworkCut,omitempty"`
	Wallet            *string     `json:"wallet,omitempty"` // Solana public key
}
