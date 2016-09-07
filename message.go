package tat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const (
	// DefaultMessageMaxSize is max size of message, can be overrided by topic
	DefaultMessageMaxSize = 140

	// True in url http way -> string
	True = "true"
	// False in url http way -> string
	False = "false"
	// TreeViewOneTree is onetree value for treeView
	TreeViewOneTree = "onetree"
	// TreeViewFullTree is fulltree value for treeView
	TreeViewFullTree = "fulltree"

	MessageActionUpdate     = "update"
	MessageActionReply      = "reply"
	MessageActionLike       = "like"
	MessageActionUnlike     = "unlike"
	MessageActionLabel      = "label"
	MessageActionUnlabel    = "unlabel"
	MessageActionVoteup     = "voteup"
	MessageActionVotedown   = "votedown"
	MessageActionUnvoteup   = "unvoteup"
	MessageActionUnvotedown = "unvotedown"
	MessageActionRelabel    = "relabel"
	MessageActionConcat     = "concat"
	MessageActionMove       = "move"
	MessageActionTask       = "task"
	MessageActionUntask     = "untask"
)

// Author struct
type Author struct {
	Username string `bson:"username" json:"username"`
	Fullname string `bson:"fullname" json:"fullname"`
}

// Label struct
type Label struct {
	Text  string `bson:"text" json:"text"`
	Color string `bson:"color" json:"color"`
}

// Message struc
type Message struct {
	ID              string    `bson:"_id"             json:"_id"`
	Text            string    `bson:"text"            json:"text"`
	Topic           string    `bson:"topic"           json:"topic"`
	InReplyOfID     string    `bson:"inReplyOfID"     json:"inReplyOfID"`
	InReplyOfIDRoot string    `bson:"inReplyOfIDRoot" json:"inReplyOfIDRoot"`
	NbLikes         int64     `bson:"nbLikes"         json:"nbLikes"`
	Labels          []Label   `bson:"labels"          json:"labels,omitempty"`
	Likers          []string  `bson:"likers"          json:"likers,omitempty"`
	VotersUP        []string  `bson:"votersUP"        json:"votersUP,omitempty"`
	VotersDown      []string  `bson:"votersDown"      json:"votersDown,omitempty"`
	NbVotesUP       int64     `bson:"nbVotesUP"       json:"nbVotesUP"`
	NbVotesDown     int64     `bson:"nbVotesDown"     json:"nbVotesDown"`
	UserMentions    []string  `bson:"userMentions"    json:"userMentions,omitempty"`
	Urls            []string  `bson:"urls"            json:"urls,omitempty"`
	Tags            []string  `bson:"tags"            json:"tags,omitempty"`
	DateCreation    float64   `bson:"dateCreation"    json:"dateCreation"`
	DateUpdate      float64   `bson:"dateUpdate"      json:"dateUpdate"`
	Author          Author    `bson:"author"          json:"author"`
	Replies         []Message `bson:"-"               json:"replies,omitempty"`
	NbReplies       int64     `bson:"nbReplies"       json:"nbReplies"`
}

// MessageCriteria are used to list messages
type MessageCriteria struct {
	Skip                int
	Limit               int
	TreeView            string
	IDMessage           string
	InReplyOfID         string
	InReplyOfIDRoot     string
	AllIDMessage        string // search in IDMessage OR InReplyOfID OR InReplyOfIDRoot
	Text                string
	Topic               string
	Label               string
	NotLabel            string
	AndLabel            string
	Tag                 string
	NotTag              string
	AndTag              string
	Username            string
	DateMinCreation     string
	DateMaxCreation     string
	DateMinUpdate       string
	DateMaxUpdate       string
	LimitMinNbReplies   string
	LimitMaxNbReplies   string
	LimitMinNbVotesUP   string
	LimitMinNbVotesDown string
	LimitMaxNbVotesUP   string
	LimitMaxNbVotesDown string
	OnlyMsgRoot         string
	OnlyCount           string
}

// CacheKey returns cacke key value
func (m *MessageCriteria) CacheKey() []string {
	s := []string{}
	if m == nil {
		return s
	}
	if m.Topic != "" {
		s = append(s, "Topic="+m.Topic)
	}
	if m.Skip != 0 {
		s = append(s, "Skip="+strconv.Itoa(m.Skip))
	}
	if m.Limit != 0 {
		s = append(s, "Limit="+strconv.Itoa(m.Limit))
	}
	if m.TreeView != "" {
		s = append(s, "TreeView="+m.TreeView)
	}
	if m.IDMessage != "" {
		s = append(s, "IDMessage="+m.IDMessage)
	}
	if m.AllIDMessage != "" {
		s = append(s, "AllIDMessage="+m.AllIDMessage)
	}
	if m.Text != "" {
		s = append(s, "Text="+m.Text)
	}
	if m.NotLabel != "" {
		s = append(s, "NotLabel="+m.NotLabel)
	}
	if m.AndLabel != "" {
		s = append(s, "AndLabel="+m.AndLabel)
	}
	if m.Tag != "" {
		s = append(s, "Tag="+m.Tag)
	}
	if m.NotTag != "" {
		s = append(s, "NotTag="+m.NotTag)
	}
	if m.AndTag != "" {
		s = append(s, "AndTag="+m.AndTag)
	}
	if m.Username != "" {
		s = append(s, "Username="+m.Username)
	}
	if m.DateMinCreation != "" {
		s = append(s, "DateMinCreation="+m.DateMinCreation)
	}
	if m.DateMinCreation != "" {
		s = append(s, "DateMinCreation="+m.DateMinCreation)
	}
	if m.DateMaxCreation != "" {
		s = append(s, "DateMaxCreation="+m.DateMaxCreation)
	}
	if m.DateMinUpdate != "" {
		s = append(s, "DateMinUpdate="+m.DateMinUpdate)
	}
	if m.DateMaxUpdate != "" {
		s = append(s, "DateMaxUpdate="+m.DateMaxUpdate)
	}
	if m.LimitMinNbReplies != "" {
		s = append(s, "LimitMinNbReplies="+m.LimitMinNbReplies)
	}
	if m.LimitMaxNbReplies != "" {
		s = append(s, "LimitMaxNbReplies="+m.LimitMaxNbReplies)
	}
	if m.LimitMinNbVotesUP != "" {
		s = append(s, "LimitMinNbVotesUP="+m.LimitMinNbVotesUP)
	}
	if m.LimitMinNbVotesDown != "" {
		s = append(s, "LimitMinNbVotesDown="+m.LimitMinNbVotesDown)
	}
	if m.LimitMaxNbVotesUP != "" {
		s = append(s, "LimitMaxNbVotesUP="+m.LimitMaxNbVotesUP)
	}
	if m.LimitMaxNbVotesDown != "" {
		s = append(s, "LimitMaxNbVotesDown="+m.LimitMaxNbVotesDown)
	}
	if m.OnlyMsgRoot != "" {
		s = append(s, "OnlyMsgRoot="+m.OnlyMsgRoot)
	}
	if m.OnlyCount != "" {
		s = append(s, "OnlyCount="+m.OnlyCount)
	}

	return s
}

// MessagesJSON represents a message and information if current topic is RW
type MessagesJSON struct {
	Messages     []Message `json:"messages"`
	IsTopicRw    bool      `json:"isTopicRw"`
	IsTopicAdmin bool      `json:"isTopicAdmin"`
}

// MessagesCountJSON represents count of messages
type MessagesCountJSON struct {
	Count int `json:"count"`
}

// MessageJSONOut represents a message and an additional info
type MessageJSONOut struct {
	Message Message `json:"message"`
	Info    string  `json:"info"`
}

// MessageJSON represents a message with action on it
type MessageJSON struct {
	ID           string `json:"_id"`
	Text         string `json:"text"`
	Option       string `json:"option"`
	Topic        string
	IDReference  string   `json:"idReference"`
	Action       string   `json:"action"`
	DateCreation float64  `json:"dateCreation"`
	Labels       []Label  `json:"labels"`
	Options      []string `json:"options"`
	Replies      []string `json:"replies"`
}

// MessageAdd post a tat message
func (c *Client) MessageAdd(message MessageJSON) (*MessageJSONOut, error) {
	if c == nil {
		return nil, ErrClientNotInitiliazed
	}

	if message.Topic == "" {
		return nil, fmt.Errorf("A message must have a Topic")
	}

	path := "/message" + message.Topic

	b, err := json.Marshal(message)
	if err != nil {
		ErrorLogFunc("Error while marshal message: %s", err)
		return nil, err
	}

	body, err := c.reqWant("POST", http.StatusCreated, path, b)
	if err != nil {
		ErrorLogFunc("Error while marshal message for MessageAdd: %s", err)
		return nil, err
	}

	out := &MessageJSONOut{}
	if err := json.Unmarshal(body, out); err != nil {
		return nil, err
	}

	return out, nil
}

func (c *Client) MessageReply() error {
	return fmt.Errorf("Not Yet Implemented")
}

//MessageDelete deletes a message. cascade : delete message and its replies. cascadeForce : delete message and its replies, event if it's in a Tasks Topic of one user
func (c *Client) MessageDelete(id, topic string, cascase bool, cascadeForce bool) error {
	var err error
	if cascase {
		_, err = c.reqWant(http.MethodDelete, 200, "/messages/cascade/"+id+topic, nil)
	} else if cascadeForce {
		_, err = c.reqWant(http.MethodDelete, 200, "/messages/cascadeforce/"+id+topic, nil)
	} else {
		_, err = c.reqWant(http.MethodDelete, 200, "/message/nocascade/"+id+topic, nil)
	}

	if err != nil {
		ErrorLogFunc("Error deleting messages: %s", err)
		return err
	}
	return nil
}

func (c *Client) MessageDeleteBulk() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) MessageUpdate() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) MessageConcat() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) MessageMove() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) MessageTask() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) MessageUntask() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) MessageLike() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) MessageUnlike() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) MessageVoteUP() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) MessageVoteDown() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) MessageUnVoteUP() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) MessageUnVoteDown() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) MessageLabel() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) MessageUnlabel() error {
	return fmt.Errorf("Not Yet Implemented")
}
func (c *Client) MessageRelabel() error {
	return fmt.Errorf("Not Yet Implemented")
}

//MessageList lists messages on a topic according to criterias
func (c *Client) MessageList(topic string, criteria *MessageCriteria) (*MessagesJSON, error) {
	if criteria == nil {
		criteria = &MessageCriteria{
			Skip:  0,
			Limit: 100,
		}
	}
	criteria.Topic = topic

	u := ""
	if criteria.TreeView != "" {
		u += "&treeView=" + criteria.TreeView
	}
	if criteria.IDMessage != "" {
		u += "&idMessage=" + criteria.IDMessage
	}
	if criteria.InReplyOfID != "" {
		u += "&inReplyOfID=" + criteria.InReplyOfID
	}
	if criteria.InReplyOfIDRoot != "" {
		u += "&inReplyOfIDRoot=" + criteria.InReplyOfIDRoot
	}
	if criteria.AllIDMessage != "" {
		u += "&allIDMessage=" + criteria.AllIDMessage
	}
	if criteria.Text != "" {
		u += "&text=" + criteria.Text
	}
	if criteria.Topic != "" {
		u += "&topic=" + criteria.Topic
	}
	if criteria.Label != "" {
		u += "&label=" + criteria.Label
	}
	if criteria.NotLabel != "" {
		u += "&notLabel=" + criteria.NotLabel
	}
	if criteria.AndLabel != "" {
		u += "&andLabel=" + criteria.AndLabel
	}
	if criteria.Tag != "" {
		u += "&tag=" + criteria.Tag
	}
	if criteria.NotTag != "" {
		u += "&notTag=" + criteria.NotTag
	}
	if criteria.AndTag != "" {
		u += "&andTag=" + criteria.AndTag
	}
	if criteria.DateMinCreation != "" {
		u += "&dateMinCreation=" + criteria.DateMinCreation
	}
	if criteria.DateMaxCreation != "" {
		u += "&dateMaxCreation=" + criteria.DateMaxCreation
	}
	if criteria.DateMinUpdate != "" {
		u += "&dateMinUpdate=" + criteria.DateMinUpdate
	}
	if criteria.DateMaxUpdate != "" {
		u += "&dateMaxUpdate=" + criteria.DateMaxUpdate
	}
	if criteria.Username != "" {
		u += "&username=" + criteria.Username
	}
	if criteria.LimitMinNbReplies != "" {
		u += "&limitMinNbReplies=" + criteria.LimitMinNbReplies
	}
	if criteria.LimitMaxNbReplies != "" {
		u += "&limitMaxNbReplies=" + criteria.LimitMaxNbReplies
	}
	if criteria.LimitMinNbVotesUP != "" {
		u += "&limitMinNbVotesUP=" + criteria.LimitMinNbVotesUP
	}
	if criteria.LimitMaxNbVotesUP != "" {
		u += "&limitMaxNbVotesUP=" + criteria.LimitMaxNbVotesUP
	}
	if criteria.LimitMinNbVotesDown != "" {
		u += "&limitMinNbVotesDown=" + criteria.LimitMinNbVotesDown
	}
	if criteria.LimitMaxNbVotesDown != "" {
		u += "&limitMaxNbVotesDown=" + criteria.LimitMaxNbVotesDown
	}
	if criteria.OnlyMsgRoot == "true" {
		u += "&onlyMsgRoot=true"
	}
	if criteria.OnlyCount == "true" {
		u += "&onlyCount=true"
	}
	path := fmt.Sprintf("/messages%s?skip=%d&limit=%d%s", criteria.Topic, criteria.Skip, criteria.Limit, u)

	body, err := c.reqWant(http.MethodGet, 200, path, nil)
	if err != nil {
		ErrorLogFunc("Error getting messages list: %s", err)
		return nil, err
	}

	DebugLogFunc("Messages List Reponse: %s", string(body))
	var messages = MessagesJSON{}
	if err := json.Unmarshal(body, &messages); err != nil {
		ErrorLogFunc("Error getting messages list: %s", err)
		return nil, err
	}

	return &messages, nil
}
