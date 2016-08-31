package tat

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
