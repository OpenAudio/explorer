package em

// EventEventType represents the event type enum
type EventEventType string

// EventEntityType represents the entity type enum
type EventEntityType string

// EventMetadata represents metadata for an event
type EventMetadata struct {
	UserID     int                    `json:"userId"`
	EventID    int                    `json:"eventId"`
	EventType  *EventEventType        `json:"eventType,omitempty"`
	EntityType *EventEntityType       `json:"entityType,omitempty"`
	EntityID   *int                   `json:"entityId,omitempty"`
	EndDate    *string                `json:"endDate,omitempty"` // ISO format date string
	EventData  map[string]interface{} `json:"eventData,omitempty"`
}

// CreateEventRequest represents a request to create an event
type CreateEventRequest struct {
	UserID     int                    `json:"userId"`
	EventID    *int                   `json:"eventId,omitempty"` // optional for creation
	EventType  *EventEventType        `json:"eventType,omitempty"`
	EntityType *EventEntityType       `json:"entityType,omitempty"`
	EntityID   *int                   `json:"entityId,omitempty"`
	EndDate    *string                `json:"endDate,omitempty"` // ISO format date string
	EventData  map[string]interface{} `json:"eventData,omitempty"`
}

// UpdateEventRequest represents a request to update an event
type UpdateEventRequest struct {
	UserID    int                    `json:"userId"`
	EventID   int                    `json:"eventId"`
	EndDate   *string                `json:"endDate,omitempty"`
	EventData map[string]interface{} `json:"eventData,omitempty"`
}

// DeleteEventRequest represents a request to delete an event
type DeleteEventRequest struct {
	UserID  int `json:"userId"`
	EventID int `json:"eventId"`
}
