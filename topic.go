package tat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Topic struct
type Topic struct {
	ID                   string           `bson:"_id"          json:"_id,omitempty"`
	Collection           string           `bson:"collection"   json:"collection"`
	Topic                string           `bson:"topic"        json:"topic"`
	Description          string           `bson:"description"  json:"description"`
	ROGroups             []string         `bson:"roGroups"     json:"roGroups,omitempty"`
	RWGroups             []string         `bson:"rwGroups"     json:"rwGroups,omitempty"`
	ROUsers              []string         `bson:"roUsers"      json:"roUsers,omitempty"`
	RWUsers              []string         `bson:"rwUsers"      json:"rwUsers,omitempty"`
	AdminUsers           []string         `bson:"adminUsers"   json:"adminUsers,omitempty"`
	AdminGroups          []string         `bson:"adminGroups"  json:"adminGroups,omitempty"`
	History              []string         `bson:"history"      json:"history"`
	MaxLength            int              `bson:"maxlength"    json:"maxlength"`
	CanForceDate         bool             `bson:"canForceDate" json:"canForceDate"`
	CanUpdateMsg         bool             `bson:"canUpdateMsg" json:"canUpdateMsg"`
	CanDeleteMsg         bool             `bson:"canDeleteMsg" json:"canDeleteMsg"`
	CanUpdateAllMsg      bool             `bson:"canUpdateAllMsg" json:"canUpdateAllMsg"`
	CanDeleteAllMsg      bool             `bson:"canDeleteAllMsg" json:"canDeleteAllMsg"`
	AdminCanUpdateAllMsg bool             `bson:"adminCanUpdateAllMsg" json:"adminCanUpdateAllMsg"`
	AdminCanDeleteAllMsg bool             `bson:"adminCanDeleteAllMsg" json:"adminCanDeleteAllMsg"`
	IsAutoComputeTags    bool             `bson:"isAutoComputeTags" json:"isAutoComputeTags"`
	IsAutoComputeLabels  bool             `bson:"isAutoComputeLabels" json:"isAutoComputeLabels"`
	DateModification     int64            `bson:"dateModification" json:"dateModificationn,omitempty"`
	DateCreation         int64            `bson:"dateCreation" json:"dateCreation,omitempty"`
	DateLastMessage      int64            `bson:"dateLastMessage" json:"dateLastMessage,omitempty"`
	Parameters           []TopicParameter `bson:"parameters" json:"parameters,omitempty"`
	Tags                 []string         `bson:"tags" json:"tags,omitempty"`
	Labels               []Label          `bson:"labels" json:"labels,omitempty"`
}

// TopicParameter struct, parameter on topics
type TopicParameter struct {
	Key   string `bson:"key"   json:"key"`
	Value string `bson:"value" json:"value"`
}

// TopicCriteria struct, used by List Topic
type TopicCriteria struct {
	Skip                 int
	Limit                int
	IDTopic              string
	Topic                string
	TopicPath            string
	Description          string
	DateMinCreation      string
	DateMaxCreation      string
	GetNbMsgUnread       string
	OnlyFavorites        string
	GetForTatAdmin       string
	GetForAllTasksTopics bool
	Group                string
}

// CacheKey returns cacke key value
func (t *TopicCriteria) CacheKey() []string {
	var s = []string{}
	if t == nil {
		return s
	}
	if t.Skip != 0 {
		s = append(s, "skip="+strconv.Itoa(t.Skip))
	}
	if t.Limit != 0 {
		s = append(s, "limit="+strconv.Itoa(t.Limit))
	}
	if t.IDTopic != "" {
		s = append(s, "id_topic="+t.IDTopic)
	}
	if t.Topic != "" {
		s = append(s, "topic="+t.Topic)
	}
	if t.TopicPath != "" {
		s = append(s, "topic_path="+t.TopicPath)
	}
	if t.Description != "" {
		s = append(s, "description="+t.Description)
	}
	if t.DateMinCreation != "" {
		s = append(s, "date_min_creation="+t.DateMinCreation)
	}
	if t.DateMaxCreation != "" {
		s = append(s, "date_max_creation="+t.DateMaxCreation)
	}
	if t.GetNbMsgUnread != "" {
		s = append(s, "get_nb_msg_unread="+t.GetNbMsgUnread)
	}
	if t.OnlyFavorites != "" {
		s = append(s, "only_favorites="+t.OnlyFavorites)
	}
	if t.GetForTatAdmin != "" {
		s = append(s, "get_for_tat_admin="+t.GetForTatAdmin)
	}
	if t.GetForAllTasksTopics {
		s = append(s, "get_for_all_tasks_topics="+strconv.FormatBool(t.GetForAllTasksTopics))
	}
	if t.Group != "" {
		s = append(s, "group="+t.Group)
	}
	return s
}

// ParamTopicUserJSON is used to update a parameter on topic
type ParamTopicUserJSON struct {
	Topic     string `json:"topic"` // topic topic
	Username  string `json:"username"`
	Recursive bool   `json:"recursive"`
}

// TopicCreateJSON is used to create a parameter on topic
type TopicCreateJSON struct {
	Topic       string `json:"topic" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type TopicParameterJSON struct {
	Topic     string `json:"topic"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Recursive bool   `json:"recursive"`
}

// TopicsJSON represents struct used by Engine while returns list of topics
type TopicsJSON struct {
	Count                int            `json:"count"`
	Topics               []Topic        `json:"topics"`
	CountTopicsMsgUnread int            `json:"countTopicsMsgUnread"`
	TopicsMsgUnread      map[string]int `json:"topicsMsgUnread"`
}

// TopicJSON represents struct used by Engine while returns one topic
type TopicJSON struct {
	Topic *Topic `json:"topic"`
}

// CheckAndFixNameTopic Add a / to topic name is it is not present
// return an error if length of name is < 4 or > 100
func CheckAndFixNameTopic(topicName string) (string, error) {
	name := strings.TrimSpace(topicName)

	if len(name) > 1 && string(name[0]) != "/" {
		name = "/" + name
	}

	if len(name) < 4 {
		return topicName, fmt.Errorf("Invalid topic length (3 or more characters, beginning with slash. Ex: /ABC): %s", topicName)
	}

	if len(name)-1 == strings.LastIndex(name, "/") {
		name = name[0 : len(name)-1]
	}

	if len(name) > 100 {
		return topicName, fmt.Errorf("Invalid topic length (max 100 characters):%s", topicName)
	}

	return name, nil
}

// TopicCreate creates a topic
func (c *Client) TopicCreate(t TopicCreateJSON) (*Topic, error) {
	if c == nil {
		return nil, ErrClientNotInitiliazed
	}

	b, err := json.Marshal(t)
	if err != nil {
		ErrorLogFunc("Error while marshal topic: %s", err)
		return nil, err
	}

	res, err := c.reqWant("POST", http.StatusCreated, "/topic", b)
	if err != nil {
		ErrorLogFunc("Error while marshal message for CreateTopic: %s", err)
		return nil, err
	}

	DebugLogFunc("createTopicResponse : %s", string(res))

	newTopic := &Topic{}
	if err := json.Unmarshal(res, newTopic); err != nil {
		return nil, err
	}

	return newTopic, nil
}

//TopicList list all topics according to criterias. Default behavior (criteria is Nil) will limit 10 topics.
func (c *Client) TopicList(criteria *TopicCriteria) (*TopicsJSON, error) {
	if c == nil {
		return nil, ErrClientNotInitiliazed
	}

	if criteria == nil {
		criteria = &TopicCriteria{
			Skip:  0,
			Limit: 10,
		}
	}

	var n string
	if criteria.Topic != "" {
		n += "&topic=" + criteria.Topic
	}
	if criteria.TopicPath != "" {
		n += "&topicPath=" + criteria.TopicPath
	}
	if criteria.IDTopic != "" {
		n += "&idTopic=" + criteria.IDTopic
	}
	if criteria.Description != "" {
		n += "&Description=" + criteria.Description
	}
	if criteria.DateMinCreation != "" {
		n += "&DateMinCreation=" + criteria.DateMinCreation
	}
	if criteria.DateMaxCreation != "" {
		n += "&DateMaxCreation=" + criteria.DateMaxCreation
	}
	if criteria.GetNbMsgUnread != "" {
		n += "&getNbMsgUnread=" + criteria.GetNbMsgUnread
	}
	if criteria.OnlyFavorites != "" {
		n += "&onlyFavorites=" + criteria.OnlyFavorites
	}
	if criteria.GetForTatAdmin == "true" {
		n += "&getForTatAdmin=" + criteria.GetForTatAdmin
	}

	path := fmt.Sprintf("/topics?skip=%d&limit=%d%s", criteria.Skip, criteria.Limit, n)

	body, err := c.reqWant(http.MethodGet, 200, path, nil)
	if err != nil {
		ErrorLogFunc("Error getting topic list: %s", err)
		return nil, err
	}

	DebugLogFunc("Topic List Reponse: %s", string(body))
	var topics = TopicsJSON{}
	if err := json.Unmarshal(body, &topics); err != nil {
		ErrorLogFunc("Error getting topic list: %s", err)
		return nil, err
	}

	return &topics, nil
}

//TopicDelete delete a topics
func (c *Client) TopicDelete(topicPath string) error {
	_, err := c.reqWant(http.MethodDelete, 200, "/topic"+topicPath, nil)
	if err != nil {
		ErrorLogFunc("Error deleting topic list: %s", err)
		return err
	}
	return nil
}

func (c *Client) TopicTruncate() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicAddRoUser() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicComputeLabels() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicTruncateLabels() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicComputeTags() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicTruncateTags() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicAllComputeLabels() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicAllComputeTags() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicAllComputeReplies() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicAllSetParam() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicAddRwUser() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicAddAdminUser() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicDeleteRoUser() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicDeleteRwUser() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicDeleteAdminUser() error {
	return fmt.Errorf("Not Yet Implemented")
}

//TopicAddRoGroup adds a read-only group on a topic
func (c *Client) TopicAddRoGroup(topic, groupname string, recursive bool) error {
	t := map[string]interface{}{
		"topic":     topic,
		"groupname": groupname,
		"recursive": recursive,
	}
	b, err := json.Marshal(t)
	if err != nil {
		ErrorLogFunc("Error marshalling %v: %s", t, err)
		return err
	}

	if _, err = c.reqWant(http.MethodPut, 200, "/topic/add/rogroup", b); err != nil {
		ErrorLogFunc("Error adding RO Group: %s", err)
		return err
	}
	return nil
}

//TopicAddRwGroup adds a read-write group on a topic
func (c *Client) TopicAddRwGroup(topic, groupname string, recursive bool) error {
	t := map[string]interface{}{
		"topic":     topic,
		"groupname": groupname,
		"recursive": recursive,
	}
	b, err := json.Marshal(t)
	if err != nil {
		ErrorLogFunc("Error marshalling %v: %s", t, err)
		return err
	}

	if _, err = c.reqWant(http.MethodPut, 200, "/topic/add/rwgroup", b); err != nil {
		ErrorLogFunc("Error adding RW Group: %s", err)
		return err
	}
	return nil
}

//TopicAddAdminGroup adds an admin group on a topic
func (c *Client) TopicAddAdminGroup(topic, groupname string, recursive bool) error {
	t := map[string]interface{}{
		"topic":     topic,
		"groupname": groupname,
		"recursive": recursive,
	}
	b, err := json.Marshal(t)
	if err != nil {
		ErrorLogFunc("Error marshalling %v: %s", t, err)
		return err
	}

	if _, err = c.reqWant(http.MethodPut, 200, "/topic/add/admingroup", b); err != nil {
		ErrorLogFunc("Error adding Admin Group: %s", err)
		return err
	}
	return nil
}

//TopicDeleteRoGroup deletes a read-only group on a topic
func (c *Client) TopicDeleteRoGroup(topic, groupname string, recursive bool) error {
	t := map[string]interface{}{
		"topic":     topic,
		"groupname": groupname,
		"recursive": recursive,
	}
	b, err := json.Marshal(t)
	if err != nil {
		ErrorLogFunc("Error marshalling %v: %s", t, err)
		return err
	}

	if _, err = c.reqWant(http.MethodPut, 200, "/topic/remove/rogroup", b); err != nil {
		ErrorLogFunc("Error deleting RO Group: %s", err)
		return err
	}
	return nil
}

//TopicDeleteRwGroup deletes a read-write group on a topic
func (c *Client) TopicDeleteRwGroup(topic, groupname string, recursive bool) error {
	t := map[string]interface{}{
		"topic":     topic,
		"groupname": groupname,
		"recursive": recursive,
	}
	b, err := json.Marshal(t)
	if err != nil {
		ErrorLogFunc("Error marshalling %v: %s", t, err)
		return err
	}

	if _, err = c.reqWant(http.MethodPut, 200, "/topic/remove/rwgroup", b); err != nil {
		ErrorLogFunc("Error deleting RW Group: %s", err)
		return err
	}
	return nil
}

//TopicDeleteAdminGroup deletes an admin group on a topic
func (c *Client) TopicDeleteAdminGroup(topic, groupname string, recursive bool) error {
	t := map[string]interface{}{
		"topic":     topic,
		"groupname": groupname,
		"recursive": recursive,
	}
	b, err := json.Marshal(t)
	if err != nil {
		ErrorLogFunc("Error marshalling %v: %s", t, err)
		return err
	}

	if _, err = c.reqWant(http.MethodPut, 200, "/topic/remove/admingroup", b); err != nil {
		ErrorLogFunc("Error deleting Admin Group: %s", err)
		return err
	}
	return nil
}

func (c *Client) TopicAddParameter() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicDeleteParameter() error {
	return fmt.Errorf("Not Yet Implemented")
}

func (c *Client) TopicParameter() error {
	return fmt.Errorf("Not Yet Implemented")
}
