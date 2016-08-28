package models

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mvdan/xurls"
	"github.com/ovh/tat/utils"
	"github.com/yesnault/hashtag"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// DefaultMessageMaxSize is max size of message, can be overrided by topic
var DefaultMessageMaxSize = 140

const lengthLabel = 100

const (
	True             = "true"
	False            = "false"
	TreeViewOneTree  = "onetree"
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

func buildMessageCriteria(criteria *MessageCriteria, username string) (bson.M, error) {
	var query = []bson.M{}

	if criteria.IDMessage != "" {
		queryIDMessages := bson.M{}
		queryIDMessages["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.IDMessage, ",") {
			queryIDMessages["$or"] = append(queryIDMessages["$or"].([]bson.M), bson.M{"_id": val})
		}
		query = append(query, queryIDMessages)
	}
	if criteria.InReplyOfID != "" {
		queryIDMessages := bson.M{}
		queryIDMessages["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.InReplyOfID, ",") {
			queryIDMessages["$or"] = append(queryIDMessages["$or"].([]bson.M), bson.M{"inReplyOfID": val})
		}
		query = append(query, queryIDMessages)
	}
	if criteria.InReplyOfIDRoot != "" {
		queryIDMessages := bson.M{}
		queryIDMessages["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.InReplyOfIDRoot, ",") {
			queryIDMessages["$or"] = append(queryIDMessages["$or"].([]bson.M), bson.M{"inReplyOfIDRoot": val})
		}
		query = append(query, queryIDMessages)
	}
	if criteria.OnlyMsgRoot == True {
		queryOnlyMsgRoot := bson.M{"inReplyOfIDRoot": bson.M{"$eq": ""}}
		query = append(query, queryOnlyMsgRoot)
	}

	if criteria.AllIDMessage != "" {
		queryIDMessages := bson.M{}
		queryIDMessages["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.AllIDMessage, ",") {
			queryIDMessages["$or"] = append(queryIDMessages["$or"].([]bson.M), bson.M{"_id": val})
			queryIDMessages["$or"] = append(queryIDMessages["$or"].([]bson.M), bson.M{"inReplyOfID": val})
			queryIDMessages["$or"] = append(queryIDMessages["$or"].([]bson.M), bson.M{"inReplyOfIDRoot": val})
		}
		query = append(query, queryIDMessages)
	}

	if criteria.Text != "" {
		query = append(query, bson.M{"text": bson.RegEx{Pattern: "^.*" + regexp.QuoteMeta(criteria.Text) + ".*$", Options: "im"}})
	}

	// Task
	if criteria.Topic == "/Private/"+username+"/Tasks" {
		queryLabels := bson.M{}
		queryLabels["$or"] = []bson.M{{"topic": "/Private/" + username + "/Tasks"}}
		queryLabels["$or"] = append(queryLabels["$or"].([]bson.M), bson.M{"labels": bson.M{"$elemMatch": bson.M{"text": bson.M{"$in": []string{"doing:" + username}}}}})
		query = append(query, queryLabels)
	} else if criteria.Topic != "" {
		queryTopics := bson.M{}
		queryTopics["$or"] = []bson.M{}
		for _, t := range strings.Split(criteria.Topic, ",") {
			queryTopics["$or"] = append(queryTopics["$or"].([]bson.M), bson.M{"topic": t})
		}
		query = append(query, queryTopics)
	}

	if criteria.Username != "" {
		queryUsernames := bson.M{}
		queryUsernames["$or"] = []bson.M{}
		queryUsernames["$or"] = append(queryUsernames["$or"].([]bson.M), bson.M{"author.username": bson.M{"$in": strings.Split(criteria.Username, ",")}})
		query = append(query, queryUsernames)
	}
	if criteria.Label != "" {
		queryLabels := bson.M{"labels": bson.M{"$elemMatch": bson.M{"text": bson.M{"$in": strings.Split(criteria.Label, ",")}}}}
		query = append(query, queryLabels)
	}
	if criteria.AndLabel != "" {
		queryLabels := bson.M{"labels.text": bson.M{"$all": strings.Split(criteria.AndLabel, ",")}}
		query = append(query, queryLabels)
	}
	if criteria.NotLabel != "" {
		for _, val := range strings.Split(criteria.NotLabel, ",") {
			queryLabels := bson.M{"labels.text": bson.M{"$ne": val}}
			query = append(query, queryLabels)
		}
	}
	if criteria.Tag != "" {
		queryTags := bson.M{"tags": bson.M{"$in": strings.Split(criteria.Tag, ",")}}
		query = append(query, queryTags)
	}
	if criteria.AndTag != "" {
		queryTags := bson.M{"tags": bson.M{"$all": strings.Split(criteria.AndTag, ",")}}
		query = append(query, queryTags)
	}
	if criteria.NotTag != "" {
		queryTags := bson.M{"tags": bson.M{"$nin": strings.Split(criteria.NotTag, ",")}}
		query = append(query, queryTags)
	}

	var bsonDate = bson.M{}
	if criteria.DateMinCreation != "" {
		i, err := strconv.ParseFloat(criteria.DateMinCreation, 64)
		if err != nil {
			return bson.M{}, fmt.Errorf("Error while parsing dateMinCreation %s", err)
		}
		bsonDate["$gte"] = utils.TatTSFromDate(utils.DateFromFloat(i))
	}
	if criteria.DateMaxCreation != "" {
		i, err := strconv.ParseFloat(criteria.DateMaxCreation, 64)
		if err != nil {
			return bson.M{}, fmt.Errorf("Error while parsing dateMaxCreation %s", err)
		}
		bsonDate["$lte"] = utils.TatTSFromDate(utils.DateFromFloat(i))
	}
	if len(bsonDate) > 0 {
		query = append(query, bson.M{"dateCreation": bsonDate})
	}

	var bsonDateUpdate = bson.M{}
	if criteria.DateMinUpdate != "" {
		i, err := strconv.ParseFloat(criteria.DateMinUpdate, 64)
		if err != nil {
			return bson.M{}, fmt.Errorf("Error while parsing dateMinUpdate %s", err)
		}
		bsonDateUpdate["$gte"] = utils.TatTSFromDate(utils.DateFromFloat(i))
	}
	if criteria.DateMaxUpdate != "" {
		i, err := strconv.ParseFloat(criteria.DateMaxUpdate, 64)
		if err != nil {
			return bson.M{}, fmt.Errorf("Error while parsing dateMaxUpdate %s", err)
		}
		bsonDateUpdate["$lte"] = utils.TatTSFromDate(utils.DateFromFloat(i))
	}
	if len(bsonDateUpdate) > 0 {
		query = append(query, bson.M{"dateUpdate": bsonDateUpdate})
	}

	var bsonNbVotesUP = bson.M{}
	if criteria.LimitMinNbVotesUP != "" {
		i, err := strconv.ParseFloat(criteria.LimitMinNbVotesUP, 64)
		if err != nil {
			return bson.M{}, fmt.Errorf("Error while parsing limitMinNbVotesUP %s", err)
		}
		bsonNbVotesUP["$gte"] = i
	}
	if criteria.LimitMaxNbVotesUP != "" {
		i, err := strconv.ParseFloat(criteria.LimitMaxNbVotesUP, 64)
		if err != nil {
			return bson.M{}, fmt.Errorf("Error while parsing limitMaxNbVotesUP %s", err)
		}
		bsonNbVotesUP["$lte"] = i
	}
	if len(bsonNbVotesUP) > 0 {
		query = append(query, bson.M{"nbVotesUP": bsonNbVotesUP})
	}

	var bsonNbVotesDown = bson.M{}
	if criteria.LimitMinNbVotesDown != "" {
		i, err := strconv.ParseFloat(criteria.LimitMinNbVotesDown, 64)
		if err != nil {
			return bson.M{}, fmt.Errorf("Error while parsing limitMinNbVotesDown %s", err)
		}
		bsonNbVotesDown["$gte"] = i
	}
	if criteria.LimitMaxNbVotesDown != "" {
		i, err := strconv.ParseFloat(criteria.LimitMaxNbVotesDown, 64)
		if err != nil {
			return bson.M{}, fmt.Errorf("Error while parsing limitMaxNbVotesDown %s", err)
		}
		bsonNbVotesDown["$lte"] = i
	}
	if len(bsonNbVotesDown) > 0 {
		query = append(query, bson.M{"nbVotesDown": bsonNbVotesDown})
	}

	if len(query) > 0 {
		return bson.M{"$and": query}, nil
	} else if len(query) == 1 {
		return query[0], nil
	}
	return bson.M{}, nil
}

// FindByID returns message by given ID
func (message *Message) FindByID(id string, topic Topic) error {
	err := getClMessages(topic).Find(bson.M{"_id": id}).One(&message)
	if err != nil {
		log.Errorf("Error while fetching message with id %s", id)
	}
	return err
}

// ListMessages list messages with given criteria
func ListMessages(criteria *MessageCriteria, username string, topic Topic) ([]Message, error) {
	var messages []Message

	c, errc := buildMessageCriteria(criteria, username)
	if errc != nil {
		return messages, errc
	}
	err := getClMessages(topic).Find(c).
		Sort("-dateCreation").
		Skip(criteria.Skip).
		Limit(criteria.Limit).
		All(&messages)

	if err != nil {
		log.Errorf("Error while Find All Messages %s", err)
	}

	if len(messages) == 0 {
		return messages, nil
	}

	if criteria.TreeView == TreeViewOneTree || criteria.TreeView == TreeViewFullTree {
		messages, err = initTree(messages, criteria, username, topic)
		if err != nil {
			log.Errorf("Error while Find All Messages (getTree) %s", err)
		}
	}
	if criteria.TreeView == TreeViewOneTree {
		messages, err = oneTreeMessages(messages, 1, criteria, username, topic)
	} else if criteria.TreeView == TreeViewFullTree {
		messages, err = fullTreeMessages(messages, 1, criteria, username, topic)
	}
	if err != nil {
		return messages, err
	}

	if criteria.TreeView == TreeViewOneTree &&
		(criteria.LimitMinNbReplies != "" || criteria.LimitMaxNbReplies != "") {
		return filterNbReplies(messages, criteria)
	}

	return messages, err
}

// CountMessages list messages with given criteria
func CountMessages(criteria *MessageCriteria, username string, topic Topic) (int, error) {
	c, errc := buildMessageCriteria(criteria, username)
	if errc != nil {
		return -1, errc
	}
	count, err := getClMessages(topic).Find(c).Count()
	if err != nil {
		log.Errorf("Error while Count Messages %s", err)
	}
	return count, err
}

func filterNbReplies(messages []Message, criteria *MessageCriteria) ([]Message, error) {
	var messagesFiltered []Message
	minReplies := -1

	if criteria.LimitMinNbReplies != "" {
		limitMinNbReplies, err := strconv.Atoi(criteria.LimitMinNbReplies)
		if err != nil {
			log.Errorf("Error while converting LimitMinNbReplies (%s) to int", criteria.LimitMinNbReplies)
		} else {
			minReplies = limitMinNbReplies
		}
	}

	maxReplies := -1
	if criteria.LimitMaxNbReplies != "" {
		limitMaxNbReplies, err := strconv.Atoi(criteria.LimitMaxNbReplies)
		if err != nil {
			log.Errorf("Error while converting LimitMaxNbReplies (%s) to int", criteria.LimitMaxNbReplies)
		} else {
			maxReplies = limitMaxNbReplies
		}
	}

	for _, msg := range messages {
		if (minReplies >= 0 && len(msg.Replies) >= minReplies) ||
			(maxReplies >= 0 && len(msg.Replies) <= maxReplies) ||
			(minReplies >= 0 && maxReplies >= 0 && len(msg.Replies) >= minReplies && len(msg.Replies) <= maxReplies) {
			messagesFiltered = append(messagesFiltered, msg)
		}
	}

	return messagesFiltered, nil
}

func initTree(messages []Message, criteria *MessageCriteria, username string, topic Topic) ([]Message, error) {
	var ids []string
	idMessages := ""
	for i := 0; i <= len(messages)-1; i++ {
		if utils.ArrayContains(ids, messages[i].ID) {
			continue
		}
		ids = append(ids, messages[i].ID)
		idMessages += messages[i].ID + ","
	}

	if idMessages == "" {
		return messages, nil
	}

	c := &MessageCriteria{
		AllIDMessage: idMessages[:len(idMessages)-1],
		NotLabel:     criteria.NotLabel,
		NotTag:       criteria.NotTag,
	}
	var msgs []Message
	cr, errc := buildMessageCriteria(c, username)
	if errc != nil {
		return msgs, errc
	}
	err := getClMessages(topic).Find(cr).Sort("-dateCreation").All(&msgs)
	if err != nil {
		log.Errorf("Error while Find Messages in getTree %s", err)
		return messages, err
	}
	return msgs, nil
}

func oneTreeMessages(messages []Message, nloop int, criteria *MessageCriteria, username string, topic Topic) ([]Message, error) {
	var tree []Message
	if nloop > 25 {
		e := "Infinite loop detected in oneTreeMessages"
		for i := 0; i <= len(messages)-1; i++ {
			e += " id:" + messages[i].ID
		}
		log.Errorf(e)
		return tree, errors.New(e)
	}

	replies := make(map[string][]Message)
	for i := 0; i <= len(messages)-1; i++ {
		if messages[i].InReplyOfIDRoot == "" {
			var replyAdded = false
			for _, msgReply := range replies[messages[i].ID] {
				messages[i].Replies = append(messages[i].Replies, msgReply)
				replyAdded = true
			}
			if replyAdded || nloop > 1 {
				tree = append(tree, messages[i])
				delete(replies, messages[i].ID)
			} else if nloop == 1 && !replyAdded {
				replies[messages[i].ID] = append(replies[messages[i].ID], messages[i])
			}
			continue
		}
		replies[messages[i].InReplyOfIDRoot] = append(replies[messages[i].InReplyOfIDRoot], messages[i])
	}

	if len(replies) == 0 {
		return tree, nil
	}
	t, err := getTree(replies, criteria, username, topic)
	if err != nil {
		return tree, err
	}

	ft, err := oneTreeMessages(t, nloop+1, criteria, username, topic)
	return append(ft, tree...), err
}

func fullTreeMessages(messages []Message, nloop int, criteria *MessageCriteria, username string, topic Topic) ([]Message, error) {
	var tree []Message
	if nloop > 10 {
		e := "Infinite loop detected in fullTreeMessages"
		for i := 0; i <= len(messages)-1; i++ {
			e += " id:" + messages[i].ID
		}
		log.Errorf(e)
		return tree, errors.New(e)
	}

	replies := make(map[string][]Message)
	var alreadyDone []string

	for i := 0; i <= len(messages)-1; i++ {
		if utils.ArrayContains(alreadyDone, messages[i].ID) {
			continue
		}

		var replyAdded = false
		for _, msgReply := range replies[messages[i].ID] {
			messages[i].Replies = append(messages[i].Replies, msgReply)
			delete(replies, messages[i].ID)
			replyAdded = true
		}
		if messages[i].InReplyOfIDRoot == "" {
			if replyAdded || nloop > 1 {
				tree = append(tree, messages[i])
			} else if nloop == 1 && !replyAdded {
				replies[messages[i].ID] = append(replies[messages[i].ID], messages[i])
			}
			continue
		}
		replies[messages[i].InReplyOfID] = append(replies[messages[i].InReplyOfID], messages[i])
		alreadyDone = append(alreadyDone, messages[i].ID)
	}

	if len(replies) == 0 {
		return tree, nil
	}
	t, err := getTree(replies, criteria, username, topic)
	if err != nil {
		return tree, err
	}
	ft, err := fullTreeMessages(t, nloop+1, criteria, username, topic)
	return append(ft, tree...), err
}

func getTree(messagesIn map[string][]Message, criteria *MessageCriteria, username string, topic Topic) ([]Message, error) {
	var messages []Message

	toDelete := false
	for idMessage := range messagesIn {
		toDelete = false
		c := &MessageCriteria{
			AllIDMessage: idMessage,
			NotLabel:     criteria.NotLabel,
			NotTag:       criteria.NotTag,
		}
		var msgs []Message
		cr, errc := buildMessageCriteria(c, username)
		if errc != nil {
			return msgs, errc
		}
		err := getClMessages(topic).Find(cr).Sort("-dateCreation").All(&msgs)
		if err != nil {
			log.Errorf("Error while Find Messages in getTree %s", err)
			return messages, err
		}

		if criteria.NotLabel != "" || criteria.NotTag != "" {
			toDelete = true
		}
		for _, m := range msgs {
			if m.ID == idMessage && m.InReplyOfIDRoot == "" && toDelete {
				toDelete = false
				break
			}
		}

		if toDelete {
			delete(messagesIn, idMessage)
		} else {
			messages = append(messages, msgs...)
		}
	}

	return messages, nil
}

// Insert a new message on one topic
func (message *Message) Insert(user User, topic Topic, text, inReplyOfID string, dateCreation float64, labels []Label, replies []string, isNotificationFromMention bool, messageRoot *Message) error {

	if !isNotificationFromMention {
		notificationsTopic := fmt.Sprintf("/Private/%s/Notifications", user.Username)
		if strings.HasPrefix(topic.Topic, notificationsTopic) {
			if !user.IsSystem {
				return fmt.Errorf("You can't write on your notifications topic")
			} else if user.IsSystem && !user.CanWriteNotifications {
				return fmt.Errorf("This user system %s has no right to write on notifications topic", user.Username)
			}
		}
	}

	message.Text = text
	if err := message.CheckAndFixText(topic); err != nil {
		return err
	}
	message.ID = bson.NewObjectId().Hex()
	message.InReplyOfID = inReplyOfID

	// 1257894123.456789
	// store ms before comma, 6 after
	now := time.Now()
	dateToStore := utils.TatTSFromDate(now)

	if dateCreation > 0 {
		if !topic.CanForceDate {
			return fmt.Errorf("You can't force date on topic %s", topic.Topic)
		}

		if !user.IsSystem {
			return fmt.Errorf("You can't force date on topic %s, you're not a system user", topic.Topic)
		}

		if !topic.CanForceDate {
			return fmt.Errorf("Error while converting dateCreation %f - CanForceDate=false on topic %s", dateCreation, topic.Topic)
		}
		dateToStore = dateCreation
	}

	if inReplyOfID != "" { // reply
		var messageReference = &Message{}
		if messageRoot != nil {
			messageReference = messageRoot
		} else {
			if err := messageReference.FindByID(inReplyOfID, topic); err != nil {
				return err
			}
		}

		if messageReference.InReplyOfID != "" {
			message.InReplyOfIDRoot = messageReference.InReplyOfIDRoot
		} else {
			message.InReplyOfIDRoot = messageReference.ID
		}
		message.Topic = messageReference.Topic

		// if msgRef.dateCreation >= dateToStore -> dateToStore must be after
		if dateToStore <= messageReference.DateCreation {
			dateToStore = messageReference.DateCreation + 1
		}
		messageReference.DateUpdate = dateToStore

		err := getClMessages(topic).Update(
			bson.M{"_id": messageReference.ID},
			bson.M{"$set": bson.M{"dateUpdate": dateToStore},
				"$inc": bson.M{"nbReplies": 1}})

		if err != nil {
			log.Errorf("Error while updating root message for reply %s", err.Error())
			return fmt.Errorf("Error while updating dateUpdate or root message for reply %s", err.Error())
		}

	} else { // root message
		message.Topic = topic.Topic
		topicDM := "/Private/" + user.Username + "/DM/"
		if strings.HasPrefix(topic.Topic, topicDM) {
			part := strings.Split(topic.Topic, "/")
			if len(part) != 5 {
				log.Errorf("wrong topic name for DM")
				return fmt.Errorf("Wrong topic name for DM:%s", topic.Topic)
			}
		}
	}

	message.NbLikes = 0
	var author = Author{}
	author.Username = user.Username
	author.Fullname = user.Fullname
	message.Author = author

	message.DateCreation = dateToStore
	message.DateUpdate = dateToStore
	message.Tags = hashtag.ExtractHashtags(message.Text)
	message.Urls = xurls.Strict.FindAllString(message.Text, -1)

	topicPrivate := "/Private/"
	if !strings.HasPrefix(topic.Topic, topicPrivate) {
		usernamesMentions := extractUsersMentions(message.Text)
		message.UserMentions = usernamesMentions
	}

	if labels != nil {
		message.Labels = checkLabels(labels, nil)
	}

	if err := getClMessages(topic).Insert(message); err != nil {
		log.Errorf("Error while inserting new message %s", err)
		return err
	}

	if !strings.HasPrefix(topic.Topic, topicPrivate) {
		message.insertNotifications(user)
	}
	go topic.UpdateTopicTags(message.Tags)
	go topic.UpdateTopicLabels(message.Labels)
	go topic.UpdateTopicLastMessage(now)

	if len(replies) > 0 {
		for _, textReply := range replies {
			reply := Message{}
			reply.Insert(user, topic, textReply, message.ID, -1, nil, nil, isNotificationFromMention, message)
		}
	}
	return nil
}

func extractUsersMentions(text string) []string {
	usernames := hashtag.ExtractMentions(text)
	var usernamesChecked []string

	for _, username := range usernames {
		var user = User{}
		found, err := user.FindByUsername(username)
		if found && err == nil {
			usernamesChecked = append(usernamesChecked, user.Username)
		}
	}
	return usernamesChecked
}

func (message *Message) insertNotifications(author User) {
	if len(message.UserMentions) == 0 {
		return
	}
	for _, userMention := range message.UserMentions {
		message.insertNotification(author, userMention)
	}
}

func (message *Message) insertNotification(author User, usernameMention string) {
	notif := Message{}
	text := fmt.Sprintf("#mention #idMessage:%s #topic:%s %s", message.ID, message.Topic, message.Text)
	topicname := fmt.Sprintf("/Private/%s/Notifications", usernameMention)
	labels := []Label{{Text: "unread", Color: "#d04437"}}
	var topic = Topic{}
	if err := topic.FindByTopic(topicname, false, false, false, nil); err != nil {
		return
	}

	if err := notif.Insert(author, topic, text, "", -1, labels, nil, true, nil); err != nil {
		// not throw err here, just log
		log.Errorf("Error while inserting notification message for %s, error: %s", usernameMention, err.Error())
	}
}

func checkLabels(labels []Label, labelsToRemove []string) []Label {
	var labelsChecked []Label
	var labelsTextChecked []string
	for _, l := range labels {
		if len(l.Text) < 1 {
			continue
		}
		if len(l.Text) > lengthLabel {
			l.Text = l.Text[0:lengthLabel]
		}
		if !utils.ArrayContains(labelsToRemove, l.Text) && !utils.ArrayContains(labelsTextChecked, l.Text) {
			labelsChecked = append(labelsChecked, l)
			labelsTextChecked = append(labelsTextChecked, l.Text)
		}
	}
	return labelsChecked
}

// CheckAndFixText truncates to maxLength (parameter on topic) characters
// if len < 1, return error
func (message *Message) CheckAndFixText(topic Topic) error {
	text := strings.TrimSpace(message.Text)
	if len(text) < 1 {
		return fmt.Errorf("Invalid Text:%s", message.Text)
	}

	maxLength := DefaultMessageMaxSize
	if topic.MaxLength > 0 {
		maxLength = topic.MaxLength
	}

	if len(text) > maxLength {
		text = text[0:maxLength]
	}
	message.Text = text
	return nil
}

// Update updates a message from database
// action could be concat (for adding additional text to message or update)
func (message *Message) Update(user User, topic Topic, newText string, action string) error {

	if action == "concat" {
		message.Text += newText
	} else {
		message.Text = newText
	}

	err := message.CheckAndFixText(topic)
	if err != nil {
		return err
	}

	err = getClMessages(topic).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{
			"text":         message.Text,
			"dateUpdate":   utils.TatTSFromNow(),
			"tags":         hashtag.ExtractHashtags(message.Text),
			"userMentions": hashtag.ExtractMentions(message.Text),
			"urls":         xurls.Strict.FindAllString(message.Text, -1),
		}})
	if err != nil {
		log.Errorf("Error while update a message %s", err)
	}

	go topic.UpdateTopicTags(message.Tags)

	return nil
}

// Move moves a message to another topic
func (message *Message) Move(user User, fromTopic Topic, toTopic Topic) error {

	// check Delete and RW are done in controller
	c := &MessageCriteria{
		IDMessage: message.ID,
		TreeView:  TreeViewOneTree,
	}

	msgs, err := ListMessages(c, "", fromTopic)
	if err != nil {
		return fmt.Errorf("Error while list Messages in Delete %s", err)
	}
	if len(msgs) != 1 {
		return fmt.Errorf("Error while list Messages in Delete (%s not unique!)", message.ID)
	}

	// here, ok, we can move
	topicUpdate := []string{toTopic.Topic}

	if fromTopic.Topic == toTopic.Topic {
		_, err = getClMessages(fromTopic).UpdateAll(
			bson.M{"$or": []bson.M{{"_id": message.ID}, {"inReplyOfIDRoot": message.ID}}},
			bson.M{"$set": bson.M{"topics": topicUpdate, "topic": toTopic.Topic}})
		if err != nil {
			log.Errorf("Error while update messages (move topic to %s) idMsgRoot:%s err:%s", toTopic.Topic, message.ID, err)
		}
	} else {
		for _, msgToMove := range msgs {
			msgToMove.Topic = toTopic.Topic
			if errInsert := getClMessages(toTopic).Insert(msgToMove); errInsert != nil {
				log.Errorf("Move> getClMessages(toTopic).Insert(message), err: %s", errInsert)
				return fmt.Errorf("Error while inserting message to new topic, old message is not deleted")
			}

			if errRemove := getClMessages(fromTopic).RemoveId(msgToMove.ID); errRemove != nil {
				log.Errorf("Move> getClMessages(toTopic).RemoveId(message), err: %s", errRemove)
				return fmt.Errorf("Error while removing message from old topic")
			}
		}
	}

	if err != nil {
		log.Errorf("Error while update messages (move topic to %s) idMsgRoot:%s err:%s", toTopic.Topic, message.ID, err)
	}

	return nil
}

// Delete deletes a message from database
func (message *Message) Delete(cascade bool, topic Topic) error {
	if message.InReplyOfID != "" {
		var messageParent = &Message{}
		if err := messageParent.FindByID(message.InReplyOfID, topic); err != nil {
			log.Errorf("message > Delete > Error while fetching message parent:%s", err.Error())
			return err
		}
		if err := getClMessages(topic).Update(
			bson.M{"_id": messageParent.ID},
			bson.M{"$inc": bson.M{"nbReplies": -1}}); err != nil {
			log.Errorf("message > Delete > Error while updating message parent:%s", err.Error())
			return err
		}
	}

	if cascade {
		_, err := getClMessages(topic).RemoveAll(bson.M{"$or": []bson.M{{"_id": message.ID}, {"inReplyOfIDRoot": message.ID}}})
		return err
	}
	return getClMessages(topic).Remove(bson.M{"_id": message.ID})
}

func (message *Message) getLabel(label string) (int, Label, error) {
	for idx, cur := range message.Labels {
		if cur.Text == label {
			return idx, cur, nil
		}
	}
	l := Label{}
	return -1, l, fmt.Errorf("label %s not found", label)
}

// ContainsLabel returns true if message contains label
func (message *Message) ContainsLabel(label string) bool {
	_, _, err := message.getLabel(label)
	return err == nil
}

// IsDoing returns true if message contains label doing or starts with doing:
func (message *Message) IsDoing() bool {
	for _, label := range message.Labels {
		if label.Text == "doing" || strings.HasPrefix(label.Text, "doing:") {
			return true
		}
	}
	return false
}

func (message *Message) getTag(tag string) (int, string, error) {
	for idx, cur := range message.Tags {
		if cur == tag {
			return idx, cur, nil
		}
	}
	return -1, "", fmt.Errorf("tag %s not found", tag)
}

func (message *Message) containsTag(tag string) bool {
	_, _, err := message.getTag(tag)
	return err == nil
}

//AddLabel add a label to a message
//truncated to 100 char in text label
func (message *Message) AddLabel(topic Topic, label string, color string) (Label, error) {
	if len(label) > lengthLabel {
		label = label[0:lengthLabel]
	}

	var newLabel = Label{Text: label, Color: color}
	if message.ContainsLabel(label) {
		log.Infof("AddLabel not possible, %s is already a label of message %s", label, message.ID)
		return newLabel, nil
	}

	err := getClMessages(topic).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": utils.TatTSFromNow()},
			"$push": bson.M{"labels": newLabel}})

	if err != nil {
		return Label{}, err
	}
	message.Labels = append(message.Labels, newLabel)

	go topic.UpdateTopicLabels(message.Labels)
	return newLabel, nil
}

// RemoveLabel removes label from on message (label text matching)
func (message *Message) RemoveLabel(label string, topic Topic) error {
	idxLabel, l, err := message.getLabel(label)
	if err != nil {
		log.Infof("Remove Label is not possible, %s is not a label of this message", label)
		return nil
	}

	err = getClMessages(topic).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": utils.TatTSFromNow()},
			"$pull": bson.M{"labels": l}})

	if err != nil {
		return err
	}

	message.Labels = append(message.Labels[:idxLabel], message.Labels[idxLabel+1:]...)
	return nil
}

// RemoveAllAndAddNewLabel removes all labels and add new label on message
func (message *Message) RemoveAllAndAddNewLabel(labels []Label, topic Topic) error {
	message.Labels = checkLabels(labels, nil)
	return getClMessages(topic).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{
			"dateUpdate": utils.TatTSFromNow(),
			"labels":     message.Labels}})
}

// RemoveSomeAndAddNewLabel removes some labels and add new label on message
func (message *Message) RemoveSomeAndAddNewLabel(labels []Label, labelsToRemove []string, topic Topic) error {
	message.Labels = append(message.Labels, labels...)
	return message.RemoveAllAndAddNewLabel(checkLabels(message.Labels, labelsToRemove), topic)
}

// Like add a like to a message
func (message *Message) Like(user User, topic Topic) error {
	if utils.ArrayContains(message.Likers, user.Username) {
		return fmt.Errorf("Like not possible, %s is already a liker of this message", user.Username)
	}
	err := getClMessages(topic).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": utils.TatTSFromNow()},
			"$inc":  bson.M{"nbLikes": 1},
			"$push": bson.M{"likers": user.Username}})

	if err == nil {
		message.NbLikes++
		message.Likers = append(message.Likers, user.Username)
	}
	return err
}

// Unlike removes a like from one message
func (message *Message) Unlike(user User, topic Topic) error {
	if !utils.ArrayContains(message.Likers, user.Username) {
		return fmt.Errorf("Unlike not possible, %s is not a liker of this message", user.Username)
	}
	err := getClMessages(topic).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": utils.TatTSFromNow()},
			"$inc":  bson.M{"nbLikes": -1},
			"$pull": bson.M{"likers": user.Username}})

	if err == nil {
		message.NbLikes--
		likers := []string{}
		for _, l := range message.Likers {
			if l != user.Username {
				likers = append(likers, l)
			}
		}
		message.Likers = likers
	}

	return err
}

// VoteUP add a vote UP to a message
func (message *Message) VoteUP(user User, topic Topic) error {
	if utils.ArrayContains(message.VotersUP, user.Username) {
		return fmt.Errorf("Vote UP not possible, %s is already a voters UP of this message", user.Username)
	}
	message.UnVoteDown(user, topic)
	return getClMessages(topic).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": utils.TatTSFromNow()},
			"$inc":  bson.M{"nbVotesUP": 1},
			"$push": bson.M{"votersUP": user.Username}})
}

// VoteDown add a vote Down to a message
func (message *Message) VoteDown(user User, topic Topic) error {
	if utils.ArrayContains(message.VotersDown, user.Username) {
		return fmt.Errorf("Vote Down not possible, %s is already a voters Down of this message", user.Username)
	}
	message.UnVoteUP(user, topic)
	return getClMessages(topic).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": utils.TatTSFromNow()},
			"$inc":  bson.M{"nbVotesDown": 1},
			"$push": bson.M{"votersDown": user.Username}})
}

// UnVoteUP removes a vote up from a message
func (message *Message) UnVoteUP(user User, topic Topic) error {
	if !utils.ArrayContains(message.VotersUP, user.Username) {
		return fmt.Errorf("Add Vote UP not possible, %s is not a voters UP of this message", user.Username)
	}
	return getClMessages(topic).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": utils.TatTSFromNow()},
			"$inc":  bson.M{"nbVotesUP": -1},
			"$pull": bson.M{"votersUP": user.Username}})
}

// UnVoteDown removes a vote down from a message
func (message *Message) UnVoteDown(user User, topic Topic) error {
	if !utils.ArrayContains(message.VotersDown, user.Username) {
		return fmt.Errorf("Remove Vote Down not possible, %s is not a voters Down of this message", user.Username)
	}
	return getClMessages(topic).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": utils.TatTSFromNow()},
			"$inc":  bson.M{"nbVotesDown": -1},
			"$pull": bson.M{"votersDown": user.Username}})
}

// GetPrivateTopicTaskName return Tasks Topic name of user
func GetPrivateTopicTaskName(user User) string {
	return "/Private/" + user.Username + "/Tasks"
}

func (message *Message) addOrRemoveFromTasks(action string, user User, topic Topic) error {
	if action != "pull" && action != "push" {
		return fmt.Errorf("Wrong action to add or remove tasks:%s", action)
	}

	idRoot := message.ID
	if message.InReplyOfIDRoot != "" {
		idRoot = message.InReplyOfIDRoot
	}

	msgReply := &Message{}
	text := "Take this thread into my tasks"
	if action == "pull" {
		text = "Remove this thread from my tasks"

		nDoing := 0
		for _, cur := range message.Labels {
			if strings.HasPrefix(cur.Text, "doing:") {
				nDoing++
			}
			if cur.Text == "doing:"+user.Username {
				message.RemoveLabel("doing:"+user.Username, topic)
			} else if cur.Text == "done:"+user.Username {
				message.RemoveLabel("done:"+user.Username, topic)
			} else if cur.Text == "done" {
				message.RemoveLabel("done", topic)
			}
		}
		if nDoing >= 1 {
			message.RemoveLabel("doing", topic)
		}
	} else { // push
		if !message.ContainsLabel("doing") {
			message.AddLabel(topic, "doing", "#5484ed")
		}
		if !message.ContainsLabel("doing:" + user.Username) {
			message.AddLabel(topic, "doing:"+user.Username, "#5484ed")
		}
		if message.ContainsLabel("open") {
			message.RemoveLabel("open", topic)
		}
		if message.ContainsLabel("done") {
			message.RemoveLabel("done", topic)
		}
	}

	return msgReply.Insert(user, topic, text, idRoot, -1, nil, nil, false, nil)
}

// AddToTasks add a message to user's tasks Topic
func (message *Message) AddToTasks(user User, topic Topic) error {
	return message.addOrRemoveFromTasks("push", user, topic)
}

// RemoveFromTasks removes a task from user's Tasks Topic
func (message *Message) RemoveFromTasks(user User, topic Topic) error {
	return message.addOrRemoveFromTasks("pull", user, topic)
}

// CountMsgSinceDate return number of messages created on one topic from a given date
func CountMsgSinceDate(topic Topic, date int64) (int, error) {
	nb, err := getClMessages(topic).Find(bson.M{"topic": topic.Topic, "dateCreation": bson.M{"$gte": date}}).Count()
	if err != nil {
		log.Errorf("Error while count messages with topic %s and dateCreation lte:%d err:%s", topic.Topic, date, err.Error())
	}
	return nb, err
}

// ListTags returns all tags on one topic
func ListTags(topic Topic) ([]string, error) {
	var tags []string
	err := getClMessages(topic).
		Find(bson.M{"topic": topic.Topic}).
		Distinct("tags", &tags)
	if err != nil {
		log.Errorf("Error while getting tags on topic %s, err:%s", topic.Topic, err.Error())
	}
	return tags, err
}

// ListLabels returns all labels on one topic
func ListLabels(topic Topic) ([]Label, error) {
	var labels []Label
	err := getClMessages(topic).
		Find(bson.M{"topic": topic.Topic}).
		Distinct("labels", &labels)
	if err != nil {
		log.Errorf("Error while getting labels on topic %s, err:%s", topic.Topic, err.Error())
	}
	return labels, err
}

func changeUsernameOnMessages(oldUsername, newUsername string) error {
	if err := changeAuthorUsernameOnMessages(oldUsername, newUsername); err != nil {
		return err
	}
	if err := changeUsernameOnMessagesTopics(oldUsername, newUsername); err != nil {
		return err
	}
	return nil
}

func changeAuthorUsernameOnMessages(oldUsername, newUsername string) error {
	// TODO on all topics (with dedicated topics)
	return fmt.Errorf("Not Yet Implemented")
	/*_, err := getClMessages().UpdateAll(
		bson.M{"author.username": oldUsername},
		bson.M{"$set": bson.M{"author.username": newUsername}})

	if err != nil {
		log.Errorf("Error while update username from %s to %s on Messages err:%s", oldUsername, newUsername, err.Error())
	}

	return err*/
}

func changeUsernameOnMessagesTopics(oldUsername, newUsername string) error {

	// TODO on all topics (with dedicated topics)
	return fmt.Errorf("Not Yet Implemented")

	/*
		var messages []Message
		err := getClMessages().Find(
			bson.M{
				"topic": bson.RegEx{Pattern: "^/Private/" + oldUsername + "/", Options: "i"},
			}).All(&messages)

		if err != nil {
			log.Errorf("Error while getting messages to update username from %s to %s on Topics %s", oldUsername, newUsername, err)
		}

		for _, msg := range messages {
			msg.Topics = []string{}
			for _, topic := range msg.Topics {
				newTopicName := strings.Replace(topic, oldUsername, newUsername, 1)
				msg.Topics = append(msg.Topics, newTopicName)
			}

			if errUpdate := getClMessages().Update(bson.M{"_id": msg.ID}, bson.M{"$set": bson.M{"topic": msg.Topic}}); errUpdate != nil {
				log.Errorf("Error while update topic on message %s name from username %s to username %s :%s", msg.ID, oldUsername, newUsername, errUpdate)
			}
		}

		return err
	*/
}

// CountAllMessages returns the total number of messages in db
func CountAllMessages() (int, error) {
	// TODO on all topics (with dedicated topics)
	return -1, fmt.Errorf("Not Yet Implemented")
	// return getClMessages().Count()
}

// ComputeReplies re-compute replies for all messages in one topic
func ComputeReplies(topic Topic) (int, error) {

	log.Debugf("ComputeReplies on topic %s", topic.Topic)

	nbCompute := 0
	var messages []Message

	var query = []bson.M{}
	query = append(query, bson.M{"topic": topic.Topic})
	query = append(query, bson.M{"inReplyOfID": bson.M{"$exists": true, "$ne": ""}})
	if err := getClMessages(topic).Find(bson.M{"$and": query}).All(&messages); err != nil {
		log.Errorf("Error while find messages for compute replies on topic %s: %s", topic.Topic, err)
	}
	log.Debugf("ComputeReplies query %s", query)
	log.Debugf("ComputeReplies on topic %s, %d msg", topic.Topic, len(messages))

	for _, msg := range messages {
		if msg.InReplyOfID == "" {
			continue
		}
		c := &MessageCriteria{InReplyOfID: msg.InReplyOfID}
		if nb, err := CountMessages(c, "", topic); err == nil {
			err := getClMessages(topic).Update(
				bson.M{"_id": msg.InReplyOfID},
				bson.M{"$set": bson.M{"nbReplies": nb}})
			if err != nil {
				log.Errorf("Error while updating message for compute replies:%s", err.Error())
			} else {
				nbCompute++
			}
		}
	}

	return nbCompute, nil
}

// DistributionMessages returns distribution of messages per topic
func DistributionMessages(col string) ([]bson.M, error) {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": bson.M{col: "$" + col},
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$sort": bson.M{
				"count": -1,
			},
		},
	}
	pipe := Store().clDefaultMessages.Pipe(pipeline)
	results := []bson.M{}

	err := pipe.All(&results)
	return results, err
}

func getClMessages(topic Topic) *mgo.Collection {
	if topic.Collection != "" {
		return Store().session.DB(databaseName).C(topic.Collection)
	}
	return Store().clDefaultMessages
}
