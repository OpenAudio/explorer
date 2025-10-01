package em

// CommentMetadata represents metadata for a comment
type CommentMetadata struct {
	Body            *string `json:"body,omitempty"`
	CommentID       *int    `json:"commentId,omitempty"`
	UserID          int     `json:"userId"`
	EntityID        int     `json:"entityId"`
	EntityType      *string `json:"entityType,omitempty"` // For now just tracks are supported
	ParentCommentID *int    `json:"parentCommentId,omitempty"`
	TrackTimestampS *int    `json:"trackTimestampS,omitempty"`
	Mentions        []int   `json:"mentions,omitempty"`
}
