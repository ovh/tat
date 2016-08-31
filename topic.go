package tat

import (
	"fmt"
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
	IsROPublic           bool             `bson:"isROPublic"   json:"isROPublic"`
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

type ParamTopicUserJSON struct {
	Topic     string `json:"topic"` // topic topic
	Username  string `json:"username"`
	Recursive bool   `json:"recursive"`
}

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
