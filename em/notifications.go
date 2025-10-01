package em

// MarkAllNotificationsAsViewedRequest represents a request to mark all notifications as viewed
type MarkAllNotificationsAsViewedRequest struct {
	UserID string `json:"userId"`
}

// UpdatePlaylistLastViewedAtRequest represents a request to update playlist last viewed timestamp
type UpdatePlaylistLastViewedAtRequest struct {
	PlaylistID string `json:"playlistId"`
	UserID     string `json:"userId"`
}

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	Data interface{} `json:"data"`
}
