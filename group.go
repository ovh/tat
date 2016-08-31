package tat

// Group struct
type Group struct {
	ID           string   `bson:"_id"          json:"_id"`
	Name         string   `bson:"name"         json:"name"`
	Description  string   `bson:"description"  json:"description"`
	Users        []string `bson:"users"        json:"users,omitempty"`
	AdminUsers   []string `bson:"adminUsers"   json:"adminUsers,omitempty"`
	DateCreation int64    `bson:"dateCreation" json:"dateCreation,omitempty"`
}

// GroupCriteria is used by List all Groups
type GroupCriteria struct {
	Skip            int
	Limit           int
	IDGroup         string
	Name            string
	Description     string
	DateMinCreation string
	DateMaxCreation string
}

type GroupsJSON struct {
	Count  int     `json:"count"`
	Groups []Group `json:"groups"`
}

type ParamUserJSON struct {
	Groupname string `json:"groupname"`
	Username  string `json:"username"`
}

type GroupJSON struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type ParamGroupJSON struct {
	Topic     string `json:"topic"`
	Groupname string `json:"groupname"`
	Recursive bool   `json:"recursive"`
}
