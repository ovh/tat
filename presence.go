package tat

import (
	"time"
)

// UserPresence struct
type UserPresence struct {
	Username string `bson:"username" json:"username"`
	Fullname string `bson:"fullname" json:"fullname"`
}

// Presence struct
type Presence struct {
	ID               string       `bson:"_id,omitempty"    json:"_id"`
	Status           string       `bson:"status"           json:"status"`
	Topic            string       `bson:"topic"            json:"topic"`
	DatePresence     int64        `bson:"datePresence"     json:"datePresence"`
	DateTimePresence time.Time    `bson:"dateTimePresence" json:"dateTimePresence"`
	UserPresence     UserPresence `bson:"userPresence"     json:"userPresence"`
}

// PresenceCriteria used by Presences List
type PresenceCriteria struct {
	Skip            int
	Limit           int
	IDPresence      string
	Status          string
	Topic           string
	Username        string
	DateMinPresence string
	DateMaxPresence string
}

// PresencesJSON represents list of presences with count for total
type PresencesJSON struct {
	Count     int        `json:"count"`
	Presences []Presence `json:"presences"`
}

// PresenceJSONOut represents a presence
type PresenceJSONOut struct {
	Presence Presence `json:"presence"`
}

// PresenceJSON represents a status on a topic
type PresenceJSON struct {
	Status   string `json:"status" binding:"required"`
	Username string `json:"username,omitempty"`
	Topic    string
}
