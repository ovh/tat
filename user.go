package tat

import (
	"fmt"
)

// Contact User Struct.
type Contact struct {
	Username string `bson:"username" json:"username"`
	Fullname string `bson:"fullname" json:"fullname"`
}

// Auth User Struct
type Auth struct {
	HashedPassword    string `bson:"hashedPassword" json:"-"`
	HashedTokenVerify string `bson:"hashedTokenVerify" json:"-"`
	DateRenewPassword int64  `bson:"dateRenewPassword" json:"dateRenewPassword"`
	DateAskReset      int64  `bson:"dateAskReset" json:"dateAskReset"`
	DateVerify        int64  `bson:"dateVerify" json:"dateVerify"`
	EmailVerified     bool   `bson:"emailVerified" json:"emailVerified"`
}

// User struct
type User struct {
	ID                     string    `bson:"_id"               json:"_id"`
	Username               string    `bson:"username"          json:"username"`
	Fullname               string    `bson:"fullname"          json:"fullname"`
	Email                  string    `bson:"email"             json:"email,omitempty"`
	Groups                 []string  `bson:"-"                 json:"groups,omitempty"`
	IsAdmin                bool      `bson:"isAdmin"           json:"isAdmin,omitempty"`
	IsSystem               bool      `bson:"isSystem"          json:"isSystem,omitempty"`
	IsArchived             bool      `bson:"isArchived"        json:"isArchived,omitempty"`
	CanWriteNotifications  bool      `bson:"canWriteNotifications" json:"canWriteNotifications,omitempty"`
	CanListUsersAsAdmin    bool      `bson:"canListUsersAsAdmin" json:"canListUsersAsAdmin,omitempty"`
	FavoritesTopics        []string  `bson:"favoritesTopics"   json:"favoritesTopics,omitempty"`
	OffNotificationsTopics []string  `bson:"offNotificationsTopics" json:"offNotificationsTopics,omitempty"`
	FavoritesTags          []string  `bson:"favoritesTags"     json:"favoritesTags,omitempty"`
	DateCreation           int64     `bson:"dateCreation"      json:"dateCreation,omitempty"`
	Contacts               []Contact `bson:"contacts"          json:"contacts,omitempty"`
	Auth                   Auth      `bson:"auth" json:"-"`
}

// UsersJSON  represents list of users and count for total
type UsersJSON struct {
	Count int    `json:"count"`
	Users []User `json:"users"`
}

// UserCriteria is used to list users with criterias
type UserCriteria struct {
	Skip            int
	Limit           int
	WithGroups      bool
	IDUser          string
	Username        string
	Fullname        string
	DateMinCreation string
	DateMaxCreation string
}

//ContactsJSON represents a contact for a user, in contacts attribute on a user
type ContactsJSON struct {
	Contacts               []Contact   `json:"contacts"`
	CountContactsPresences int         `json:"countContactsPresences"`
	ContactsPresences      *[]Presence `json:"contactsPresence"`
}

func (c *Client) UserList() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserMe() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserContacts() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserAddContact() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserRemoveContact() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserAddFavoriteTopic() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserRemoveFavoriteTopic() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserEnableNotificationsTopic() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserEnableNotificationsAllTopics() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserDisableNotificationsTopic() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserDisableNotificationsAllTopics() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserAddFavoriteTag() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserRemoveFavoriteTag() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserAdd() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserReset() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserResetSystem() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserConvertToSystem() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserUpdateSystem() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserArchive() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserRename() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserUpdate() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserSetAdmin() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserVerify() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) UserCheck() error {
	return fmt.Errorf("Not Yet Implemented")
}
