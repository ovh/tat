package tat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

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
	UserUsername    string
}

// CacheKey returns cacke key value
func (g *GroupCriteria) CacheKey() []string {
	var s = []string{}
	if g == nil {
		return s
	}
	if g.Skip != 0 {
		s = append(s, "skip="+strconv.Itoa(g.Skip))
	}
	if g.Limit != 0 {
		s = append(s, "limit="+strconv.Itoa(g.Limit))
	}
	if g.IDGroup != "" {
		s = append(s, "id_group="+g.IDGroup)
	}
	if g.Name != "" {
		s = append(s, "name="+g.Name)
	}
	if g.Description != "" {
		s = append(s, "description="+g.Description)
	}
	if g.DateMinCreation != "" {
		s = append(s, "date_min_creation="+g.DateMinCreation)
	}
	if g.DateMaxCreation != "" {
		s = append(s, "date_max_creation="+g.DateMaxCreation)
	}
	if g.UserUsername != "" {
		s = append(s, "user_username="+g.UserUsername)
	}
	return s
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

func (c *Client) GroupList() error {
	return fmt.Errorf("Not Yet Implemented")
}

//GroupCreate creates a group
func (c *Client) GroupCreate(g GroupJSON) (*Group, error) {
	if c == nil {
		return nil, ErrClientNotInitiliazed
	}

	b, err := json.Marshal(g)
	if err != nil {
		ErrorLogFunc("Error while marshal group: %s", err)
		return nil, err
	}

	res, err := c.reqWant(http.MethodPost, http.StatusCreated, "/group", b)
	if err != nil {
		ErrorLogFunc("Error while marshal group for GroupCreate: %s", err)
		return nil, err
	}

	DebugLogFunc("GroupCreate : %s", string(res))

	newGroup := &Group{}
	if err := json.Unmarshal(res, newGroup); err != nil {
		return nil, err
	}

	return newGroup, nil
}

func (c *Client) GroupUpdate(groupname string) error {
	return fmt.Errorf("Not Yet Implemented")
}

//GroupDelete delete a group
func (c *Client) GroupDelete(groupname string) error {
	_, err := c.reqWant(http.MethodDelete, http.StatusOK, "/group/edit/"+groupname, nil)
	if err != nil {
		ErrorLogFunc("Error while deleting group: %s", err)
		return err
	}
	return nil
}

func (c *Client) GroupAddUser() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) GroupDeleteUser() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) GroupAddAdminUser() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) GroupDeleteAdminUser() error {
	return fmt.Errorf("Not Yet Implemented")
}
