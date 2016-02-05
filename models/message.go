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
	"gopkg.in/mgo.v2/bson"
)

// DefaultMessageMaxSize is max size of message, can be overrided by topic
var DefaultMessageMaxSize = 140

const lengthLabel = 100

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
	Topics          []string  `bson:"topics"          json:"topics"`
	InReplyOfID     string    `bson:"inReplyOfID"     json:"inReplyOfID"`
	InReplyOfIDRoot string    `bson:"inReplyOfIDRoot" json:"inReplyOfIDRoot"`
	NbLikes         int64     `bson:"nbLikes"         json:"nbLikes"`
	Labels          []Label   `bson:"labels"          json:"labels,omitempty"`
	Likers          []string  `bson:"likers"          json:"likers,omitempty"`
	UserMentions    []string  `bson:"userMentions"    json:"userMentions,omitempty"`
	Urls            []string  `bson:"urls"            json:"urls,omitempty"`
	Tags            []string  `bson:"tags"            json:"tags,omitempty"`
	DateCreation    float64   `bson:"dateCreation"    json:"dateCreation"`
	DateUpdate      float64   `bson:"dateUpdate"      json:"dateUpdate"`
	Author          Author    `bson:"author"          json:"author"`
	Replies         []Message `bson:"-"               json:"replies,omitempty"`
}

// MessageCriteria are used to list messages
type MessageCriteria struct {
	Skip              int
	Limit             int
	TreeView          string
	IDMessage         string
	InReplyOfID       string
	InReplyOfIDRoot   string
	AllIDMessage      string // search in IDMessage OR InReplyOfID OR InReplyOfIDRoot
	Text              string
	Topic             string
	Label             string
	NotLabel          string
	AndLabel          string
	Tag               string
	NotTag            string
	AndTag            string
	Username          string
	DateMinCreation   string
	DateMaxCreation   string
	DateMinUpdate     string
	DateMaxUpdate     string
	LimitMinNbReplies string
	LimitMaxNbReplies string
	OnlyMsgRoot       string
}

func buildMessageCriteria(criteria *MessageCriteria) bson.M {
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
	if criteria.OnlyMsgRoot == "true" {
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
		queryTexts := bson.M{}
		queryTexts["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.Text, ",") {
			queryTexts["$or"] = append(queryTexts["$or"].([]bson.M), bson.M{"text": bson.RegEx{Pattern: "^.*" + regexp.QuoteMeta(val) + ".*$", Options: "im"}})
		}
		query = append(query, queryTexts)
	}
	if criteria.Topic != "" {
		queryTopics := bson.M{}
		queryTopics["$or"] = []bson.M{}
		queryTopics["$or"] = append(queryTopics["$or"].([]bson.M), bson.M{"topics": bson.M{"$in": strings.Split(criteria.Topic, ",")}})
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
			log.Errorf("Error while parsing dateMinCreation %s", err)
		} else {
			tm := utils.DateFromFloat(i)
			bsonDate["$gte"] = tm.Unix()
		}
	}
	if criteria.DateMaxCreation != "" {
		i, err := strconv.ParseFloat(criteria.DateMaxCreation, 64)
		if err != nil {
			log.Errorf("Error while parsing dateMaxCreation %s", err)
		}
		tm := utils.DateFromFloat(i)

		if err == nil {
			// TODO use floating point
			bsonDate["$lte"] = tm.Unix()
		} else {
			log.Errorf("Error while parsing dateMaxCreation %s", err)
		}
	}
	if len(bsonDate) > 0 {
		query = append(query, bson.M{"dateCreation": bsonDate})
	}

	var bsonDateUpdate = bson.M{}
	if criteria.DateMinUpdate != "" {
		i, err := strconv.ParseFloat(criteria.DateMinUpdate, 64)
		if err != nil {
			log.Errorf("Error while parsing dateMinUpdate %s", err)
		}
		tm := utils.DateFromFloat(i)

		if err == nil {
			// TODO use floating point
			bsonDateUpdate["$gte"] = tm.Unix()
		} else {
			log.Errorf("Error while parsing dateMinUpdate %s", err)
		}
	}
	if criteria.DateMaxUpdate != "" {
		i, err := strconv.ParseFloat(criteria.DateMaxUpdate, 64)
		if err != nil {
			log.Errorf("Error while parsing dateMaxUpdate %s", err)
		}
		tm := utils.DateFromFloat(i)

		if err == nil {
			// TODO use floating point
			bsonDateUpdate["$lte"] = tm.Unix()
		} else {
			log.Errorf("Error while parsing dateMaxUpdate %s", err)
		}
	}
	if len(bsonDateUpdate) > 0 {
		query = append(query, bson.M{"dateUpdate": bsonDateUpdate})
	}

	if len(query) > 0 {
		return bson.M{"$and": query}
	} else if len(query) == 1 {
		return query[0]
	}
	return bson.M{}
}

// FindByID returns message by given ID
func (message *Message) FindByID(id string) error {
	err := Store().clMessages.Find(bson.M{"_id": id}).One(&message)
	if err != nil {
		log.Errorf("Error while fecthing message with id %s", id)
	}
	return err
}

// ListMessages list messages with given criteria
func ListMessages(criteria *MessageCriteria) ([]Message, error) {
	var messages []Message

	err := Store().clMessages.Find(buildMessageCriteria(criteria)).
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

	if criteria.TreeView == "onetree" || criteria.TreeView == "fulltree" {
		messages, err = initTree(messages, criteria)
		if err != nil {
			log.Errorf("Error while Find All Messages (getTree) %s", err)
		}
	}
	if criteria.TreeView == "onetree" {
		messages, err = oneTreeMessages(messages, 1, criteria)
	} else if criteria.TreeView == "fulltree" {
		messages, err = fullTreeMessages(messages, 1, criteria)
	}
	if err != nil {
		return messages, err
	}

	if criteria.TreeView == "onetree" &&
		(criteria.LimitMinNbReplies != "" || criteria.LimitMaxNbReplies != "") {
		return filterNbReplies(messages, criteria)
	}

	return messages, err
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

func initTree(messages []Message, criteria *MessageCriteria) ([]Message, error) {
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
	err := Store().clMessages.Find(buildMessageCriteria(c)).Sort("-dateCreation").All(&msgs)
	if err != nil {
		log.Errorf("Error while Find Messages in getTree %s", err)
		return messages, err
	}
	return msgs, nil
}

func oneTreeMessages(messages []Message, nloop int, criteria *MessageCriteria) ([]Message, error) {
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
	t, err := getTree(replies, criteria)
	if err != nil {
		return tree, err
	}

	ft, err := oneTreeMessages(t, nloop+1, criteria)
	return append(ft, tree...), err
}

func fullTreeMessages(messages []Message, nloop int, criteria *MessageCriteria) ([]Message, error) {
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
	t, err := getTree(replies, criteria)
	if err != nil {
		return tree, err
	}
	ft, err := fullTreeMessages(t, nloop+1, criteria)
	return append(ft, tree...), err
}

func getTree(messagesIn map[string][]Message, criteria *MessageCriteria) ([]Message, error) {
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
		err := Store().clMessages.Find(buildMessageCriteria(c)).Sort("-dateCreation").All(&msgs)
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
func (message *Message) Insert(user User, topic Topic, text, inReplyOfID string, dateCreation float64, labels []Label, isNotificationFromMention bool) error {

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
	err := message.CheckAndFixText(topic)
	if err != nil {
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
			return fmt.Errorf("Error while converting dateCreation %f - error:%s", dateCreation, err.Error())
		}
		dateToStore = dateCreation
	}

	if inReplyOfID != "" { // reply
		var messageReference = &Message{}
		if err := messageReference.FindByID(inReplyOfID); err != nil {
			return err
		}
		if messageReference.InReplyOfID != "" {
			message.InReplyOfIDRoot = messageReference.InReplyOfIDRoot
		} else {
			message.InReplyOfIDRoot = messageReference.ID
		}
		message.Topics = messageReference.Topics

		// if msgRef.dateCreation >= dateToStore -> dateToStore must be after
		if dateToStore <= messageReference.DateCreation {
			dateToStore = messageReference.DateCreation + 1
		}
		messageReference.DateUpdate = dateToStore

		err := Store().clMessages.Update(
			bson.M{"_id": messageReference.ID},
			bson.M{"$set": bson.M{"dateUpdate": dateToStore}})

		if err != nil {
			log.Errorf("Error while updating root message for reply %s", err.Error())
			return fmt.Errorf("Error while updating dateUpdate or root message for reply %s", err.Error())
		}

	} else { // root message
		message.Topics = append(message.Topics, topic.Topic)
		topicDM := "/Private/" + user.Username + "/DM/"
		if strings.HasPrefix(topic.Topic, topicDM) {
			part := strings.Split(topic.Topic, "/")
			if len(part) != 5 {
				log.Errorf("wrong topic name for DM")
				return fmt.Errorf("Wrong topic name for DM:%s", topic.Topic)
			}
			topicInverse := "/Private/" + part[4] + "/DM/" + user.Username
			message.Topics = append(message.Topics, topicInverse)
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
		message.Labels = checkLabels(labels)
	}

	if err := Store().clMessages.Insert(message); err != nil {
		log.Errorf("Error while inserting new message %s", err)
		return err
	}

	if !strings.HasPrefix(topic.Topic, topicPrivate) {
		message.insertNotifications(user)
	}
	return nil
}

func extractUsersMentions(text string) []string {
	usernames := hashtag.ExtractMentions(text)
	var usernamesChecked []string

	for _, username := range usernames {
		var user = User{}
		if err := user.FindByUsername(username); err == nil {
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
	text := fmt.Sprintf("#mention #idMessage:%s #topic:%s %s", message.ID, message.Topics[0], message.Text)
	topicname := fmt.Sprintf("/Private/%s/Notifications", usernameMention)
	labels := []Label{Label{Text: "unread", Color: "#d04437"}}
	var topic = Topic{}
	if err := topic.FindByTopic(topicname, false, nil); err != nil {
		return
	}

	if err := notif.Insert(author, topic, text, "", -1, labels, true); err != nil {
		// not throw err here, just log
		log.Errorf("Error while inserting notification message for %s, error: %s", usernameMention, err.Error())
	}
}

func checkLabels(labels []Label) []Label {
	var labelsChecked []Label
	for _, l := range labels {
		if len(l.Text) > lengthLabel {
			l.Text = l.Text[0:lengthLabel]
		}
		labelsChecked = append(labelsChecked, l)
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

	err = Store().clMessages.Update(
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

	return nil
}

// Move moves a message to another topic
func (message *Message) Move(user User, newTopic Topic) error {

	// check Delete and RW are done in controller
	c := &MessageCriteria{
		IDMessage: message.ID,
		TreeView:  "onetree",
	}

	msgs, err := ListMessages(c)
	if err != nil {
		return fmt.Errorf("Error while list Messages in Delete %s", err)
	}
	if len(msgs) != 1 {
		return fmt.Errorf("Error while list Messages in Delete (%s not unique!)", message.ID)
	}

	for _, e := range msgs[0].Replies {
		if len(e.Topics) != 1 {
			return fmt.Errorf("A reply belongs more than one topic, you can't move this thread.")
		}
	}

	// here, ok, we can move
	topicUpdate := []string{newTopic.Topic}
	_, err = Store().clMessages.UpdateAll(
		bson.M{"$or": []bson.M{bson.M{"_id": message.ID}, bson.M{"inReplyOfIDRoot": message.ID}}},
		bson.M{"$set": bson.M{"topics": topicUpdate}})

	if err != nil {
		log.Errorf("Error while update messages (move topic to %s) idMsgRoot:%s err:%s", newTopic.Topic, message.ID, err)
	}

	return nil
}

// Delete deletes a message from database
func (message *Message) Delete(cascade bool) error {
	if cascade {
		_, err := Store().clMessages.RemoveAll(bson.M{"$or": []bson.M{bson.M{"_id": message.ID}, bson.M{"inReplyOfIDRoot": message.ID}}})
		return err
	}
	return Store().clMessages.Remove(bson.M{"_id": message.ID})
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
func (message *Message) AddLabel(label string, color string) (Label, error) {
	if len(label) > lengthLabel {
		label = label[0:lengthLabel]
	}

	if message.ContainsLabel(label) {
		return Label{}, fmt.Errorf("AddLabel not possible, %s is already a label of this message", label)
	}
	var newLabel = Label{Text: label, Color: color}

	err := Store().clMessages.Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": utils.TatTSFromNow()},
			"$push": bson.M{"labels": newLabel}})

	if err != nil {
		return Label{}, err
	}
	message.Labels = append(message.Labels, newLabel)
	return newLabel, nil
}

// RemoveLabel removes label from on message (label text matching)
func (message *Message) RemoveLabel(label string) error {
	idxLabel, l, err := message.getLabel(label)
	if err != nil {
		return fmt.Errorf("Remove Label is not possible, %s is not a label of this message", label)
	}

	err = Store().clMessages.Update(
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
func (message *Message) RemoveAllAndAddNewLabel(labels []Label) error {
	message.Labels = checkLabels(labels)
	return Store().clMessages.Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{
			"dateUpdate": utils.TatTSFromNow(),
			"labels":     message.Labels}})
}

// Like add a like to a message
func (message *Message) Like(user User) error {
	if utils.ArrayContains(message.Likers, user.Username) {
		return fmt.Errorf("Like not possible, %s is already a liker of this message", user.Username)
	}
	err := Store().clMessages.Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": utils.TatTSFromNow()},
			"$inc":  bson.M{"nbLikes": 1},
			"$push": bson.M{"likers": user.Username}})

	if err != nil {
		return err
	}
	return nil
}

// Unlike removes a like from one message
func (message *Message) Unlike(user User) error {
	if !utils.ArrayContains(message.Likers, user.Username) {
		return fmt.Errorf("Unlike not possible, %s is not a liker of this message", user.Username)
	}
	err := Store().clMessages.Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": utils.TatTSFromNow()},
			"$inc":  bson.M{"nbLikes": -1},
			"$pull": bson.M{"likers": user.Username}})

	if err != nil {
		return err
	}
	return nil
}

// GetPrivateTopicTaskName return Tasks Topic name of user
func GetPrivateTopicTaskName(user User) string {
	return "/Private/" + user.Username + "/Tasks"
}

func (message *Message) addOrRemoveFromTasks(action string, user User, topic Topic) error {
	if action != "pull" && action != "push" {
		return fmt.Errorf("Wrong action to add or remove tasks:%s", action)
	}
	topicTasksName := GetPrivateTopicTaskName(user)
	idRoot := message.ID
	if message.InReplyOfIDRoot != "" {
		idRoot = message.InReplyOfIDRoot
	}

	_, err := Store().clMessages.UpdateAll(
		bson.M{"$or": []bson.M{bson.M{"_id": idRoot}, bson.M{"inReplyOfIDRoot": idRoot}}},
		bson.M{"$" + action: bson.M{"topics": topicTasksName}})

	if err != nil {
		return err
	}

	msgReply := &Message{}
	text := "Take this thread into my tasks"
	if action == "pull" {
		text = "Remove this thread from my tasks"
	}
	return msgReply.Insert(user, topic, text, idRoot, -1, nil, false)
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
func CountMsgSinceDate(topic string, date int64) (int, error) {
	nb, err := Store().clMessages.Find(bson.M{"topics": bson.M{"$in": [1]string{topic}}, "dateCreation": bson.M{"$gte": date}}).Count()
	if err != nil {
		log.Errorf("Error while count message with topic %s and dateCreation lte:%d", topic, date)
	}
	return nb, err
}

func changeUsernameOnMessages(oldUsername, newUsername string) {
	changeAuthorUsernameOnMessages(oldUsername, newUsername)
	changeUsernameOnMessagesTopics(oldUsername, newUsername)
}

func changeAuthorUsernameOnMessages(oldUsername, newUsername string) error {
	_, err := Store().clMessages.UpdateAll(
		bson.M{"author.username": oldUsername},
		bson.M{"$set": bson.M{"author.username": newUsername}})

	if err != nil {
		log.Errorf("Error while update username from %s to %s on Messages %s", oldUsername, newUsername, err)
	}

	return err
}

func changeUsernameOnMessagesTopics(oldUsername, newUsername string) error {
	var messages []Message

	err := Store().clMessages.Find(
		bson.M{
			"topics": bson.RegEx{Pattern: "^/Private/" + oldUsername + "/", Options: "i"},
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

		err := Store().clMessages.Update(
			bson.M{"_id": msg.ID},
			bson.M{"$set": bson.M{"topics": msg.Topics}},
		)

		if err != nil {
			log.Errorf("Error while update topic on message %s name from username %s to username %s :%s", msg.ID, oldUsername, newUsername, err)
		}
	}

	return err
}

// CountMessages returns the total number of messages in db
func CountMessages() (int, error) {
	return Store().clMessages.Count()
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
	pipe := Store().clMessages.Pipe(pipeline)
	results := []bson.M{}

	err := pipe.All(&results)
	return results, err
}
