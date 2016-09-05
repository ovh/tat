package tat

import "strconv"

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
