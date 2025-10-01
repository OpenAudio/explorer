package em

import "time"

// EthCollectibleGatedConditions represents Ethereum collectible gating
type EthCollectibleGatedConditions struct {
	Chain        string  `json:"chain"` // "eth"
	Address      string  `json:"address"`
	Standard     string  `json:"standard"` // "ERC721" or "ERC1155"
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	ImageURL     *string `json:"imageUrl,omitempty"`
	ExternalLink *string `json:"externalLink,omitempty"`
}

// SolCollectibleGatedConditions represents Solana collectible gating
type SolCollectibleGatedConditions struct {
	Chain        string  `json:"chain"` // "sol"
	Address      string  `json:"address"`
	Name         string  `json:"name"`
	ImageURL     *string `json:"imageUrl,omitempty"`
	ExternalLink *string `json:"externalLink,omitempty"`
}

// CollectibleGatedConditions represents NFT collection gating
type CollectibleGatedConditions struct {
	NFTCollection interface{} `json:"nftCollection,omitempty"` // Can be EthCollectibleGatedConditions or SolCollectibleGatedConditions
}

// FollowGatedConditions represents follow-based gating
type FollowGatedConditions struct {
	FollowUserID int `json:"followUserId"`
}

// TipGatedConditions represents tip-based gating
type TipGatedConditions struct {
	TipUserID int `json:"tipUserId"`
}

// TokenGatedConditions represents token-based gating
type TokenGatedConditions struct {
	TokenGate struct {
		TokenMint   string  `json:"tokenMint"`
		TokenAmount float64 `json:"tokenAmount"`
	} `json:"tokenGate"`
}

// USDCPurchaseConditions represents USDC purchase conditions
type USDCPurchaseConditions struct {
	USDCPurchase struct {
		Price  float64     `json:"price"` // positive
		Splits interface{} `json:"splits"`
	} `json:"usdcPurchase"`
}

// FieldVisibility represents field visibility settings
type FieldVisibility struct {
	Mood      *bool `json:"mood,omitempty"`
	Tags      *bool `json:"tags,omitempty"`
	Genre     *bool `json:"genre,omitempty"`
	Share     *bool `json:"share,omitempty"`
	PlayCount *bool `json:"playCount,omitempty"`
	Remixes   *bool `json:"remixes,omitempty"`
}

// RemixParent represents a parent track in a remix
type RemixParent struct {
	ParentTrackID string `json:"parentTrackId"`
}

// RemixOf represents remix information
type RemixOf struct {
	Tracks []RemixParent `json:"tracks"` // min 1
}

// StemOf represents stem information
type StemOf struct {
	Category      string `json:"category"` // StemCategory enum value
	ParentTrackID string `json:"parentTrackId"`
}

// DDEXResourceContributor represents a DDEX resource contributor
type DDEXResourceContributor struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

// DDEXCopyright represents copyright information
type DDEXCopyright struct {
	Year string `json:"year"`
	Text string `json:"text"`
}

// DDEXRightsController represents rights controller information
type DDEXRightsController struct {
	Name       string   `json:"name"`
	Roles      []string `json:"roles"`
	RightsType string   `json:"rightsType"`
}

// TrackMetadata represents metadata for uploading a track
type TrackMetadata struct {
	TrackID                      *string                   `json:"trackId,omitempty"`
	AiAttributionUserID          *string                   `json:"aiAttributionUserId,omitempty"`
	Description                  *string                   `json:"description,omitempty"`
	FieldVisibility              *FieldVisibility          `json:"fieldVisibility,omitempty"`
	Genre                        string                    `json:"genre"` // required, not null, not "ALL"
	ISRC                         *string                   `json:"isrc,omitempty"`
	IsUnlisted                   *bool                     `json:"isUnlisted,omitempty"`
	ISWC                         *string                   `json:"iswc,omitempty"`
	License                      *string                   `json:"license,omitempty"`
	Mood                         *string                   `json:"mood,omitempty"`
	IsStreamGated                *bool                     `json:"isStreamGated,omitempty"`
	StreamConditions             interface{}               `json:"streamConditions,omitempty"` // Union of gating conditions
	IsDownloadGated              *bool                     `json:"isDownloadGated,omitempty"`
	DownloadConditions           interface{}               `json:"downloadConditions,omitempty"` // Union of gating conditions
	ReleaseDate                  *time.Time                `json:"releaseDate,omitempty"`
	RemixOf                      *RemixOf                  `json:"remixOf,omitempty"`
	StemOf                       *StemOf                   `json:"stemOf,omitempty"`
	Tags                         *string                   `json:"tags,omitempty"`
	Title                        string                    `json:"title"` // required
	Duration                     *float64                  `json:"duration,omitempty"`
	PreviewStartSeconds          *float64                  `json:"previewStartSeconds,omitempty"`
	PlacementHosts               *string                   `json:"placementHosts,omitempty"`
	AudioUploadID                *string                   `json:"audioUploadId,omitempty"`
	TrackCID                     *string                   `json:"trackCid,omitempty"`
	PreviewCID                   *string                   `json:"previewCid,omitempty"`
	OrigFileCID                  *string                   `json:"origFileCid,omitempty"`
	OrigFilename                 *string                   `json:"origFilename,omitempty"`
	IsDownloadable               *bool                     `json:"isDownloadable,omitempty"`
	IsOriginalAvailable          *bool                     `json:"isOriginalAvailable,omitempty"`
	DDEXReleaseIDs               map[string]string         `json:"ddexReleaseIds,omitempty"`
	DDEXApp                      *string                   `json:"ddexApp,omitempty"`
	Artists                      []DDEXResourceContributor `json:"artists,omitempty"`
	ResourceContributors         []DDEXResourceContributor `json:"resourceContributors,omitempty"`
	IndirectResourceContributors []DDEXResourceContributor `json:"indirectResourceContributors,omitempty"`
	RightsController             *DDEXRightsController     `json:"rightsController,omitempty"`
	CopyrightLine                *DDEXCopyright            `json:"copyrightLine,omitempty"`
	ProducerCopyrightLine        *DDEXCopyright            `json:"producerCopyrightLine,omitempty"`
	ParentalWarningType          *string                   `json:"parentalWarningType,omitempty"`
	BPM                          *float64                  `json:"bpm,omitempty"`
	IsCustomBPM                  *bool                     `json:"isCustomBpm,omitempty"`
	MusicalKey                   *string                   `json:"musicalKey,omitempty"`
	IsCustomMusicalKey           *bool                     `json:"isCustomMusicalKey,omitempty"`
	AudioAnalysisErrorCount      *int                      `json:"audioAnalysisErrorCount,omitempty"`
	CommentsDisabled             *bool                     `json:"commentsDisabled,omitempty"`
	IsScheduledRelease           *bool                     `json:"isScheduledRelease,omitempty"`
}

// UploadTrackRequest represents a request to upload a track
type UploadTrackRequest struct {
	UserID   string        `json:"userId"`
	Metadata TrackMetadata `json:"metadata"`
}

// TrackFilesMetadata represents metadata for uploading track files (genre is optional)
type TrackFilesMetadata struct {
	TrackMetadata
	Genre *string `json:"genre,omitempty"` // Override to make optional
}

// UploadTrackFilesRequest represents a request to upload track files
type UploadTrackFilesRequest struct {
	UserID   string             `json:"userId"`
	Metadata TrackFilesMetadata `json:"metadata"`
}

// UpdateTrackRequest represents a request to update a track
type UpdateTrackRequest struct {
	UserID          string         `json:"userId"`
	TrackID         string         `json:"trackId"`
	Metadata        *TrackMetadata `json:"metadata,omitempty"` // Partial metadata
	GeneratePreview *bool          `json:"generatePreview,omitempty"`
}

// DeleteTrackRequest represents a request to delete a track
type DeleteTrackRequest struct {
	UserID  string `json:"userId"`
	TrackID string `json:"trackId"`
}

// FavoriteTrackMetadata represents metadata for favoriting a track
type FavoriteTrackMetadata struct {
	IsSaveOfRepost bool `json:"isSaveOfRepost"`
}

// FavoriteTrackRequest represents a request to favorite a track
type FavoriteTrackRequest struct {
	UserID   string                 `json:"userId"`
	TrackID  string                 `json:"trackId"`
	Metadata *FavoriteTrackMetadata `json:"metadata,omitempty"`
}

// UnfavoriteTrackRequest represents a request to unfavorite a track
type UnfavoriteTrackRequest struct {
	UserID  string `json:"userId"`
	TrackID string `json:"trackId"`
}

// RepostTrackMetadata represents metadata for reposting a track
type RepostTrackMetadata struct {
	IsRepostOfRepost bool `json:"isRepostOfRepost"`
}

// RepostTrackRequest represents a request to repost a track
type RepostTrackRequest struct {
	UserID   string               `json:"userId"`
	TrackID  string               `json:"trackId"`
	Metadata *RepostTrackMetadata `json:"metadata,omitempty"`
}

// UnrepostTrackRequest represents a request to unrepost a track
type UnrepostTrackRequest struct {
	UserID  string `json:"userId"`
	TrackID string `json:"trackId"`
}

// RecordTrackDownloadRequest represents a request to record a track download
type RecordTrackDownloadRequest struct {
	UserID  *string `json:"userId,omitempty"`
	TrackID string  `json:"trackId"`
}

// ShareTrackRequest represents a request to share a track
type ShareTrackRequest struct {
	UserID  string `json:"userId"`
	TrackID string `json:"trackId"`
}

// GetPurchaseTrackInstructionsRequest represents a request to get purchase instructions
type GetPurchaseTrackInstructionsRequest struct {
	UserID            string      `json:"userId"`
	TrackID           string      `json:"trackId"`
	Price             interface{} `json:"price"`                 // number or bigint
	ExtraAmount       interface{} `json:"extraAmount,omitempty"` // number or bigint
	IncludeNetworkCut *bool       `json:"includeNetworkCut,omitempty"`
}

// PurchaseTrackRequest represents a request to purchase a track
type PurchaseTrackRequest struct {
	UserID            string      `json:"userId"`
	TrackID           string      `json:"trackId"`
	Price             interface{} `json:"price"`                 // number or bigint
	ExtraAmount       interface{} `json:"extraAmount,omitempty"` // number or bigint
	IncludeNetworkCut *bool       `json:"includeNetworkCut,omitempty"`
	Wallet            *string     `json:"wallet,omitempty"` // Solana public key
}
