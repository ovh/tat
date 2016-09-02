package message

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mvdan/xurls"
	"github.com/ovh/tat"
	"github.com/ovh/tat/api/store"
	topicDB "github.com/ovh/tat/api/topic"
	"github.com/yesnault/hashtag"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const lengthLabel = 100

// InitDB gets all topics, for each topic with "collection" setted, add
// collection to store
func InitDB() {
	topics, err := topicDB.FindAllTopicsWithCollections()
	if err != nil {
		log.Errorf("Error while FindAllTopicsWithCollections:%s", err)
		return
	}
	for _, topic := range topics {
		log.Debugf("Ensure index for topic: %s col: %s", topic.Topic, topic.Collection)
		store.EnsureIndexesMessages(topic.Collection)
	}
}

func buildMessageCriteria(criteria *tat.MessageCriteria, username string) (bson.M, error) {
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
	if criteria.OnlyMsgRoot == tat.True {
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
		bsonDate["$gte"] = tat.TSFromDate(tat.DateFromFloat(i))
	}
	if criteria.DateMaxCreation != "" {
		i, err := strconv.ParseFloat(criteria.DateMaxCreation, 64)
		if err != nil {
			return bson.M{}, fmt.Errorf("Error while parsing dateMaxCreation %s", err)
		}
		bsonDate["$lte"] = tat.TSFromDate(tat.DateFromFloat(i))
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
		bsonDateUpdate["$gte"] = tat.TSFromDate(tat.DateFromFloat(i))
	}
	if criteria.DateMaxUpdate != "" {
		i, err := strconv.ParseFloat(criteria.DateMaxUpdate, 64)
		if err != nil {
			return bson.M{}, fmt.Errorf("Error while parsing dateMaxUpdate %s", err)
		}
		bsonDateUpdate["$lte"] = tat.TSFromDate(tat.DateFromFloat(i))
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

// FindByIDDefaultCollection returns message by given ID
// TODO remove this func after migrate all topic to dedicated
func FindByIDDefaultCollection(message *tat.Message, id string) error {
	err := store.Tat().CDefaultMessages.Find(bson.M{"_id": id}).One(&message)
	if err != nil {
		log.Errorf("Error while fetching message with id %s", id)
	}
	return err
}

// FindByID returns message by given ID
func FindByID(message *tat.Message, id string, topic tat.Topic) error {
	err := store.GetCMessages(topic.Collection).Find(bson.M{"_id": id}).One(&message)
	if err != nil {
		log.Errorf("Error while fetching message with id %s", id)
	}
	return err
}

// ListMessages list messages with given criteria
func ListMessages(criteria *tat.MessageCriteria, username string, topic tat.Topic) ([]tat.Message, error) {
	var messages []tat.Message

	c, errc := buildMessageCriteria(criteria, username)
	if errc != nil {
		return messages, errc
	}
	err := store.GetCMessages(topic.Collection).Find(c).
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

	if criteria.TreeView == tat.TreeViewOneTree || criteria.TreeView == tat.TreeViewFullTree {
		messages, err = initTree(messages, criteria, username, topic)
		if err != nil {
			log.Errorf("Error while Find All Messages (getTree) %s", err)
		}
	}
	if criteria.TreeView == tat.TreeViewOneTree {
		messages, err = OneTreeMessages(messages, 1, criteria, username, topic)
	} else if criteria.TreeView == tat.TreeViewFullTree {
		messages, err = FullTreeMessages(messages, 1, criteria, username, topic)
	}
	if err != nil {
		return messages, err
	}

	if criteria.TreeView == tat.TreeViewOneTree &&
		(criteria.LimitMinNbReplies != "" || criteria.LimitMaxNbReplies != "") {
		return filterNbReplies(messages, criteria)
	}

	return messages, err
}

func filterNbReplies(messages []tat.Message, criteria *tat.MessageCriteria) ([]tat.Message, error) {
	var messagesFiltered []tat.Message
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

func initTree(messages []tat.Message, criteria *tat.MessageCriteria, username string, topic tat.Topic) ([]tat.Message, error) {
	var ids []string
	idMessages := ""
	for i := 0; i <= len(messages)-1; i++ {
		if tat.ArrayContains(ids, messages[i].ID) {
			continue
		}
		ids = append(ids, messages[i].ID)
		idMessages += messages[i].ID + ","
	}

	if idMessages == "" {
		return messages, nil
	}

	c := &tat.MessageCriteria{
		AllIDMessage: idMessages[:len(idMessages)-1],
		NotLabel:     criteria.NotLabel,
		NotTag:       criteria.NotTag,
	}
	var msgs []tat.Message
	cr, errc := buildMessageCriteria(c, username)
	if errc != nil {
		return msgs, errc
	}
	err := store.GetCMessages(topic.Collection).Find(cr).Sort("-dateCreation").All(&msgs)
	if err != nil {
		log.Errorf("Error while Find Messages in getTree %s", err)
		return messages, err
	}
	return msgs, nil
}

// OneTreeMessages returns list msg with only one deep
func OneTreeMessages(messages []tat.Message, nloop int, criteria *tat.MessageCriteria, username string, topic tat.Topic) ([]tat.Message, error) {
	var tree []tat.Message
	if nloop > 25 {
		e := "Infinite loop detected in OneTreeMessages"
		for i := 0; i <= len(messages)-1; i++ {
			e += " id:" + messages[i].ID
		}
		log.Errorf(e)
		return tree, errors.New(e)
	}

	replies := make(map[string][]tat.Message)
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

	ft, err := OneTreeMessages(t, nloop+1, criteria, username, topic)
	return append(ft, tree...), err
}

// FullTreeMessages returns list msg with only full deep
func FullTreeMessages(messages []tat.Message, nloop int, criteria *tat.MessageCriteria, username string, topic tat.Topic) ([]tat.Message, error) {
	var tree []tat.Message
	if nloop > 10 {
		e := "Infinite loop detected in FullTreeMessages"
		for i := 0; i <= len(messages)-1; i++ {
			e += " id:" + messages[i].ID
		}
		log.Errorf(e)
		return tree, errors.New(e)
	}

	replies := make(map[string][]tat.Message)
	var alreadyDone []string

	for i := 0; i <= len(messages)-1; i++ {
		if tat.ArrayContains(alreadyDone, messages[i].ID) {
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
	ft, err := FullTreeMessages(t, nloop+1, criteria, username, topic)
	return append(ft, tree...), err
}

func getTree(messagesIn map[string][]tat.Message, criteria *tat.MessageCriteria, username string, topic tat.Topic) ([]tat.Message, error) {
	var messages []tat.Message

	toDelete := false
	for idMessage := range messagesIn {
		toDelete = false
		c := &tat.MessageCriteria{
			AllIDMessage: idMessage,
			NotLabel:     criteria.NotLabel,
			NotTag:       criteria.NotTag,
		}
		var msgs []tat.Message
		cr, errc := buildMessageCriteria(c, username)
		if errc != nil {
			return msgs, errc
		}
		err := store.GetCMessages(topic.Collection).Find(cr).Sort("-dateCreation").All(&msgs)
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
func Insert(message *tat.Message, user tat.User, topic tat.Topic, text, inReplyOfID string, dateCreation float64, labels []tat.Label, replies []string, isNotificationFromMention bool, messageRoot *tat.Message) error {

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
	if err := CheckAndFixText(message, topic); err != nil {
		return err
	}
	message.ID = bson.NewObjectId().Hex()
	message.InReplyOfID = inReplyOfID

	// 1257894123.456789
	// store ms before comma, 6 after
	now := time.Now()
	dateToStore := tat.TSFromDate(now)

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
		var messageReference = &tat.Message{}
		if messageRoot != nil {
			messageReference = messageRoot
		} else {
			if err := FindByID(messageReference, inReplyOfID, topic); err != nil {
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

		err := store.GetCMessages(topic.Collection).Update(
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
	var author = tat.Author{}
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

	if err := store.GetCMessages(topic.Collection).Insert(message); err != nil {
		log.Errorf("Error while inserting new message %s", err)
		return err
	}

	if !strings.HasPrefix(topic.Topic, topicPrivate) {
		insertNotifications(message, user)
	}
	go topicDB.UpdateTopicTags(&topic, message.Tags)
	go topicDB.UpdateTopicLabels(&topic, message.Labels)
	go topicDB.UpdateTopicLastMessage(&topic, now)

	if len(replies) > 0 {
		for _, textReply := range replies {
			reply := tat.Message{}
			Insert(&reply, user, topic, textReply, message.ID, -1, nil, nil, isNotificationFromMention, message)
		}
	}
	return nil
}

//isUsernameExist retrieve information from user with username
func isUsernameExist(username string) (bool, error) {
	user := tat.User{}
	err := store.Tat().CUsers.
		Find(bson.M{"username": username}).
		Select(bson.M{"username": 1}).
		One(&user)

	if err == mgo.ErrNotFound {
		return false, nil
	} else if err != nil {
		log.Errorf("Error while fetching user with username %s err:%s", username, err)
		return false, err
	}
	return true, nil
}

func extractUsersMentions(text string) []string {
	usernames := hashtag.ExtractMentions(text)
	var usernamesChecked []string

	for _, username := range usernames {
		var user = tat.User{}
		found, err := isUsernameExist(username)
		if found && err == nil {
			usernamesChecked = append(usernamesChecked, user.Username)
		}
	}
	return usernamesChecked
}

func insertNotifications(message *tat.Message, author tat.User) {
	if len(message.UserMentions) == 0 {
		return
	}
	for _, userMention := range message.UserMentions {
		insertNotification(message, author, userMention)
	}
}

func insertNotification(message *tat.Message, author tat.User, usernameMention string) {
	notif := tat.Message{}
	text := fmt.Sprintf("#mention #idMessage:%s #topic:%s %s", message.ID, message.Topic, message.Text)
	topicname := fmt.Sprintf("/Private/%s/Notifications", usernameMention)
	labels := []tat.Label{{Text: "unread", Color: "#d04437"}}
	var topic = tat.Topic{}
	if err := topicDB.FindByTopic(&topic, topicname, false, false, false, nil); err != nil {
		return
	}

	if err := Insert(&notif, author, topic, text, "", -1, labels, nil, true, nil); err != nil {
		// not throw err here, just log
		log.Errorf("Error while inserting notification message for %s, error: %s", usernameMention, err.Error())
	}
}

func checkLabels(labels []tat.Label, labelsToRemove []string) []tat.Label {
	var labelsChecked []tat.Label
	var labelsTextChecked []string
	for _, l := range labels {
		if len(l.Text) < 1 {
			continue
		}
		if len(l.Text) > lengthLabel {
			l.Text = l.Text[0:lengthLabel]
		}
		if !tat.ArrayContains(labelsToRemove, l.Text) && !tat.ArrayContains(labelsTextChecked, l.Text) {
			labelsChecked = append(labelsChecked, l)
			labelsTextChecked = append(labelsTextChecked, l.Text)
		}
	}
	return labelsChecked
}

// CheckAndFixText truncates to maxLength (parameter on topic) characters
// if len < 1, return error
func CheckAndFixText(message *tat.Message, topic tat.Topic) error {
	text := strings.TrimSpace(message.Text)
	if len(text) < 1 {
		return fmt.Errorf("Invalid Text:%s", message.Text)
	}

	maxLength := tat.DefaultMessageMaxSize
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
func Update(message *tat.Message, user tat.User, topic tat.Topic, newText string, action string) error {

	if action == "concat" {
		message.Text += newText
	} else {
		message.Text = newText
	}

	if err := CheckAndFixText(message, topic); err != nil {
		return err
	}

	err := store.GetCMessages(topic.Collection).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{
			"text":         message.Text,
			"dateUpdate":   tat.TSFromNow(),
			"tags":         hashtag.ExtractHashtags(message.Text),
			"userMentions": hashtag.ExtractMentions(message.Text),
			"urls":         xurls.Strict.FindAllString(message.Text, -1),
		}})
	if err != nil {
		log.Errorf("Error while update a message %s", err)
	}

	go topicDB.UpdateTopicTags(&topic, message.Tags)

	return nil
}

// Move moves a message to another topic
func Move(message *tat.Message, user tat.User, fromTopic tat.Topic, toTopic tat.Topic) error {

	// check Delete and RW are done in controller
	c := &tat.MessageCriteria{
		IDMessage: message.ID,
		TreeView:  tat.TreeViewOneTree,
	}

	msgs, err := ListMessages(c, "", fromTopic)
	if err != nil {
		return fmt.Errorf("Error while list Messages in Delete %s", err)
	}
	if len(msgs) != 1 {
		return fmt.Errorf("Error while list Messages in Delete (%s not unique!)", message.ID)
	}

	if fromTopic.Topic == toTopic.Topic {
		_, err = store.GetCMessages(fromTopic.Collection).UpdateAll(
			bson.M{"$or": []bson.M{{"_id": message.ID}, {"inReplyOfIDRoot": message.ID}}},
			bson.M{"$set": bson.M{"topic": toTopic.Topic}})
		if err != nil {
			log.Errorf("Error while update messages (move topic to %s) idMsgRoot:%s err:%s", toTopic.Topic, message.ID, err)
		}
	} else {
		msgsToMove := []tat.Message{msgs[0]}
		msgsToMove = append(msgsToMove, msgs[0].Replies...)
		for _, msgToMove := range msgsToMove {
			msgToMove.Topic = toTopic.Topic
			if errInsert := store.GetCMessages(toTopic.Collection).Insert(msgToMove); errInsert != nil {
				log.Errorf("Move> getClMessages(toTopic).Insert(message), err: %s", errInsert)
				return fmt.Errorf("Error while inserting message to new topic, old message is not deleted")
			}

			if errRemove := store.GetCMessages(fromTopic.Collection).RemoveId(msgToMove.ID); errRemove != nil {
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
func Delete(message *tat.Message, cascade bool, topic tat.Topic) error {
	if message.InReplyOfID != "" {
		var messageParent = &tat.Message{}
		if err := FindByID(messageParent, message.InReplyOfID, topic); err != nil {
			log.Errorf("message > Delete > Error while fetching message parent:%s", err.Error())
			return err
		}
		if err := store.GetCMessages(topic.Collection).Update(
			bson.M{"_id": messageParent.ID},
			bson.M{"$inc": bson.M{"nbReplies": -1}}); err != nil {
			log.Errorf("message > Delete > Error while updating message parent:%s", err.Error())
			return err
		}
	}

	if cascade {
		_, err := store.GetCMessages(topic.Collection).RemoveAll(bson.M{"$or": []bson.M{{"_id": message.ID}, {"inReplyOfIDRoot": message.ID}}})
		return err
	}
	return store.GetCMessages(topic.Collection).Remove(bson.M{"_id": message.ID})
}

func getLabel(message *tat.Message, label string) (int, tat.Label, error) {
	for idx, cur := range message.Labels {
		if cur.Text == label {
			return idx, cur, nil
		}
	}
	l := tat.Label{}
	return -1, l, fmt.Errorf("label %s not found", label)
}

// ContainsLabel returns true if message contains label
func ContainsLabel(message *tat.Message, label string) bool {
	_, _, err := getLabel(message, label)
	return err == nil
}

// IsDoing returns true if message contains label doing or starts with doing:
func IsDoing(message *tat.Message) bool {
	for _, label := range message.Labels {
		if label.Text == "doing" || strings.HasPrefix(label.Text, "doing:") {
			return true
		}
	}
	return false
}

func getTag(message *tat.Message, tag string) (int, string, error) {
	for idx, cur := range message.Tags {
		if cur == tag {
			return idx, cur, nil
		}
	}
	return -1, "", fmt.Errorf("tag %s not found", tag)
}

func containsTag(message *tat.Message, tag string) bool {
	_, _, err := getTag(message, tag)
	return err == nil
}

//AddLabel add a label to a message
//truncated to 100 char in text label
func AddLabel(message *tat.Message, topic tat.Topic, label string, color string) (tat.Label, error) {
	if len(label) > lengthLabel {
		label = label[0:lengthLabel]
	}

	var newLabel = tat.Label{Text: label, Color: color}
	if ContainsLabel(message, label) {
		log.Infof("AddLabel not possible, %s is already a label of message %s", label, message.ID)
		return newLabel, nil
	}

	err := store.GetCMessages(topic.Collection).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": tat.TSFromNow()},
			"$push": bson.M{"labels": newLabel}})

	if err != nil {
		return tat.Label{}, err
	}
	message.Labels = append(message.Labels, newLabel)

	go topicDB.UpdateTopicLabels(&topic, message.Labels)
	return newLabel, nil
}

// RemoveLabel removes label from on message (label text matching)
func RemoveLabel(message *tat.Message, label string, topic tat.Topic) error {
	idxLabel, l, err := getLabel(message, label)
	if err != nil {
		log.Infof("Remove Label is not possible, %s is not a label of this message", label)
		return nil
	}

	err = store.GetCMessages(topic.Collection).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": tat.TSFromNow()},
			"$pull": bson.M{"labels": l}})

	if err != nil {
		return err
	}

	message.Labels = append(message.Labels[:idxLabel], message.Labels[idxLabel+1:]...)
	return nil
}

// RemoveAllAndAddNewLabel removes all labels and add new label on message
func RemoveAllAndAddNewLabel(message *tat.Message, labels []tat.Label, topic tat.Topic) error {
	message.Labels = checkLabels(labels, nil)
	return store.GetCMessages(topic.Collection).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{
			"dateUpdate": tat.TSFromNow(),
			"labels":     message.Labels}})
}

// RemoveSomeAndAddNewLabel removes some labels and add new label on message
func RemoveSomeAndAddNewLabel(message *tat.Message, labels []tat.Label, labelsToRemove []string, topic tat.Topic) error {
	message.Labels = append(message.Labels, labels...)
	return RemoveAllAndAddNewLabel(message, checkLabels(message.Labels, labelsToRemove), topic)
}

// Like add a like to a message
func Like(message *tat.Message, user tat.User, topic tat.Topic) error {
	if tat.ArrayContains(message.Likers, user.Username) {
		return fmt.Errorf("Like not possible, %s is already a liker of this message", user.Username)
	}
	err := store.GetCMessages(topic.Collection).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": tat.TSFromNow()},
			"$inc":  bson.M{"nbLikes": 1},
			"$push": bson.M{"likers": user.Username}})

	if err == nil {
		message.NbLikes++
		message.Likers = append(message.Likers, user.Username)
	}
	return err
}

// Unlike removes a like from one message
func Unlike(message *tat.Message, user tat.User, topic tat.Topic) error {
	if !tat.ArrayContains(message.Likers, user.Username) {
		return fmt.Errorf("Unlike not possible, %s is not a liker of this message", user.Username)
	}
	err := store.GetCMessages(topic.Collection).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": tat.TSFromNow()},
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
func VoteUP(message *tat.Message, user tat.User, topic tat.Topic) error {
	if tat.ArrayContains(message.VotersUP, user.Username) {
		return fmt.Errorf("Vote UP not possible, %s is already a voters UP of this message", user.Username)
	}
	UnVoteDown(message, user, topic)
	return store.GetCMessages(topic.Collection).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": tat.TSFromNow()},
			"$inc":  bson.M{"nbVotesUP": 1},
			"$push": bson.M{"votersUP": user.Username}})
}

// VoteDown add a vote Down to a message
func VoteDown(message *tat.Message, user tat.User, topic tat.Topic) error {
	if tat.ArrayContains(message.VotersDown, user.Username) {
		return fmt.Errorf("Vote Down not possible, %s is already a voters Down of this message", user.Username)
	}
	UnVoteUP(message, user, topic)
	return store.GetCMessages(topic.Collection).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": tat.TSFromNow()},
			"$inc":  bson.M{"nbVotesDown": 1},
			"$push": bson.M{"votersDown": user.Username}})
}

// UnVoteUP removes a vote up from a message
func UnVoteUP(message *tat.Message, user tat.User, topic tat.Topic) error {
	if !tat.ArrayContains(message.VotersUP, user.Username) {
		return fmt.Errorf("Add Vote UP not possible, %s is not a voters UP of this message", user.Username)
	}
	return store.GetCMessages(topic.Collection).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": tat.TSFromNow()},
			"$inc":  bson.M{"nbVotesUP": -1},
			"$pull": bson.M{"votersUP": user.Username}})
}

// UnVoteDown removes a vote down from a message
func UnVoteDown(message *tat.Message, user tat.User, topic tat.Topic) error {
	if !tat.ArrayContains(message.VotersDown, user.Username) {
		return fmt.Errorf("Remove Vote Down not possible, %s is not a voters Down of this message", user.Username)
	}
	return store.GetCMessages(topic.Collection).Update(
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"dateUpdate": tat.TSFromNow()},
			"$inc":  bson.M{"nbVotesDown": -1},
			"$pull": bson.M{"votersDown": user.Username}})
}

// GetPrivateTopicTaskName return Tasks tat.Topic name of user
func GetPrivateTopicTaskName(user tat.User) string {
	return "/Private/" + user.Username + "/Tasks"
}

func addOrRemoveFromTasks(message *tat.Message, action string, user tat.User, topic tat.Topic) error {
	if action != "pull" && action != "push" {
		return fmt.Errorf("Wrong action to add or remove tasks:%s", action)
	}

	idRoot := message.ID
	if message.InReplyOfIDRoot != "" {
		idRoot = message.InReplyOfIDRoot
	}

	msgReply := &tat.Message{}
	text := "Take this thread into my tasks"
	if action == "pull" {
		text = "Remove this thread from my tasks"

		nDoing := 0
		for _, cur := range message.Labels {
			if strings.HasPrefix(cur.Text, "doing:") {
				nDoing++
			}
			if cur.Text == "doing:"+user.Username {
				RemoveLabel(message, "doing:"+user.Username, topic)
			} else if cur.Text == "done:"+user.Username {
				RemoveLabel(message, "done:"+user.Username, topic)
			} else if cur.Text == "done" {
				RemoveLabel(message, "done", topic)
			}
		}
		if nDoing >= 1 {
			RemoveLabel(message, "doing", topic)
		}
	} else { // push
		if !ContainsLabel(message, "doing") {
			AddLabel(message, topic, "doing", "#5484ed")
		}
		if !ContainsLabel(message, "doing:"+user.Username) {
			AddLabel(message, topic, "doing:"+user.Username, "#5484ed")
		}
		if ContainsLabel(message, "open") {
			RemoveLabel(message, "open", topic)
		}
		if ContainsLabel(message, "done") {
			RemoveLabel(message, "done", topic)
		}
	}

	return Insert(msgReply, user, topic, text, idRoot, -1, nil, nil, false, nil)
}

// AddToTasks add a message to user's tasks tat.Topic
func AddToTasks(message *tat.Message, user tat.User, topic tat.Topic) error {
	return addOrRemoveFromTasks(message, "push", user, topic)
}

// RemoveFromTasks removes a task from user's Tasks tat.Topic
func RemoveFromTasks(message *tat.Message, user tat.User, topic tat.Topic) error {
	return addOrRemoveFromTasks(message, "pull", user, topic)
}

// CountMsgSinceDate return number of messages created on one topic from a given date
func CountMsgSinceDate(topic tat.Topic, date int64) (int, error) {
	nb, err := store.GetCMessages(topic.Collection).Find(bson.M{"topic": topic.Topic, "dateCreation": bson.M{"$gte": date}}).Count()
	if err != nil {
		log.Errorf("Error while count messages with topic %s and dateCreation lte:%d err:%s", topic.Topic, date, err.Error())
	}
	return nb, err
}

// ChangeUsernameOnMessages changes username of a user on all msg
func ChangeUsernameOnMessages(oldUsername, newUsername string) error {
	if err := changeAuthorUsernameOnMessages(oldUsername, newUsername); err != nil {
		return err
	}
	if err := ChangeUsernameOnMessagesTopics(oldUsername, newUsername); err != nil {
		return err
	}
	return nil
}

func changeAuthorUsernameOnMessages(oldUsername, newUsername string) error {

	// default messages collection
	_, err := store.Tat().Session.DB(store.DatabaseName).C(store.CollectionDefaultMessages).UpdateAll(
		bson.M{"author.username": oldUsername},
		bson.M{"$set": bson.M{"author.username": newUsername}})

	if err != nil {
		log.Errorf("Error while update username from %s to %s on Messages err:%s", oldUsername, newUsername, err.Error())
	}

	// and all dedicated messages collections
	topics, errFindAll := topicDB.FindAllTopicsWithCollections()
	if errFindAll != nil {
		return errFindAll
	}

	for _, topic := range topics {
		_, err := store.GetCMessages(topic.Collection).UpdateAll(
			bson.M{"author.username": oldUsername},
			bson.M{"$set": bson.M{"author.username": newUsername}})

		if err != nil {
			log.Errorf("Error while update username from %s to %s on Messages err:%s", oldUsername, newUsername, err.Error())
			return err
		}
	}
	return nil
}

// ChangeUsernameOnMessagesTopics change username on topics
func ChangeUsernameOnMessagesTopics(oldUsername, newUsername string) error {
	var topics []tat.Topic
	errFindTopics := store.Tat().CTopics.Find(bson.M{"topic": bson.RegEx{Pattern: "^/Private/" + oldUsername + "/", Options: "i"}}).
		Select(bson.M{"_id": 1, "collection": 1, "topic": 1}).
		All(&topics)
	if errFindTopics != nil {
		return errFindTopics
	}

	collections := []string{store.CollectionDefaultMessages}
	for _, topic := range topics {
		if topic.Collection != "" {
			collections = append(collections, topic.Collection)
		}
	}
	for _, collection := range collections {
		var messages []tat.Message
		err := store.Tat().Session.DB(store.DatabaseName).C(collection).Find(
			bson.M{"topic": bson.RegEx{Pattern: "^/Private/" + oldUsername + "/", Options: "i"}}).All(&messages)

		if err != nil {
			log.Errorf("Error while getting messages to update username from %s to %s on tat.Topics %s", oldUsername, newUsername, err)
		}

		// Not perf, check to update all msgs in a collection
		for _, msg := range messages {
			if errUpdate := store.Tat().Session.DB(store.DatabaseName).C(collection).
				Update(bson.M{"_id": msg.ID}, bson.M{"$set": bson.M{"topic": msg.Topic}}); errUpdate != nil {
				log.Errorf("Error while update topic on message %s name from username %s to username %s on collection %s, err:%s", msg.ID, oldUsername, newUsername, collection, errUpdate)
			}
		}
	}

	return nil
}

// CountAllMessages returns the total number of messages in db
func CountAllMessages() (int, error) {

	count, errDefault := store.Tat().Session.DB(store.DatabaseName).C(store.CollectionDefaultMessages).Count()
	if errDefault != nil {
		return -1, errDefault
	}

	topics, errFindAll := topicDB.FindAllTopicsWithCollections()
	if errFindAll != nil {
		return -1, errFindAll
	}

	for _, topic := range topics {
		c, errCount := store.GetCMessages(topic.Collection).Count()
		if errCount != nil {
			return -1, errCount
		}
		count += c
	}
	return count, nil
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
	pipe := store.Tat().CDefaultMessages.Pipe(pipeline)
	results := []bson.M{}

	err := pipe.All(&results)
	return results, err
}

// CountMessages list messages with given criteria
func CountMessages(criteria *tat.MessageCriteria, username string, topic tat.Topic) (int, error) {
	c, errc := buildMessageCriteria(criteria, username)
	if errc != nil {
		return -1, errc
	}
	count, err := store.GetCMessages(topic.Collection).Find(c).Count()
	if err != nil {
		log.Errorf("Error while Count Messages %s", err)
	}
	return count, err
}

// ComputeReplies re-compute replies for all messages in one topic
func ComputeReplies(topic tat.Topic) (int, error) {

	log.Debugf("ComputeReplies on topic %s", topic.Topic)

	nbCompute := 0
	var messages []tat.Message

	var query = []bson.M{}
	query = append(query, bson.M{"topic": topic.Topic})
	query = append(query, bson.M{"inReplyOfID": bson.M{"$exists": true, "$ne": ""}})
	if err := store.GetCMessages(topic.Collection).Find(bson.M{"$and": query}).All(&messages); err != nil {
		log.Errorf("Error while find messages for compute replies on topic %s: %s", topic.Topic, err)
	}
	log.Debugf("ComputeReplies query %s", query)
	log.Debugf("ComputeReplies on topic %s, %d msg", topic.Topic, len(messages))

	for _, msg := range messages {
		if msg.InReplyOfID == "" {
			continue
		}
		c := &tat.MessageCriteria{InReplyOfID: msg.InReplyOfID}
		if nb, err := CountMessages(c, "", topic); err == nil {
			err := store.GetCMessages(topic.Collection).Update(
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

// AllTopicsComputeReplies computes Replies on all topics
func AllTopicsComputeReplies() (string, error) {
	var topics []tat.Topic
	err := store.Tat().CTopics.Find(bson.M{}).
		Select(topicDB.GetTopicSelectedFields(true, false, false, false)).
		Sort("topic").
		All(&topics)

	if err != nil {
		log.Errorf("Error while getting all topics for compute replies")
		return "", err
	}

	nOk := 1
	for _, topic := range topics {
		nbCompute, err := ComputeReplies(topic)
		if err != nil {
			log.Errorf("Error while compute replies on topic %s: %s", topic.Topic, err.Error())
		} else {
			log.Infof(" %d replies compute on topic %s", nbCompute, topic.Topic)
			nOk++
		}
	}

	return fmt.Sprintf("Replies computed on %d topics", nOk), nil
}

// MigrateMessagesToDedicatedTopic migrates a topic, from default to dedicated
func MigrateMessagesToDedicatedTopic(topic *tat.Topic, limit int) (int, error) {
	criteria := &tat.MessageCriteria{Topic: topic.Topic}

	c, errc := buildMessageCriteria(criteria, "")
	if errc != nil {
		return -1, errc
	}
	var msgsToMigrate []tat.Message
	err := store.Tat().Session.DB(store.DatabaseName).C(store.CollectionDefaultMessages).Find(c).
		Sort("-dateCreation").
		Skip(0).
		Limit(limit).
		All(&msgsToMigrate)

	if err != nil {
		log.Errorf("MigrateMessagesToDedicatedTopic> Error while Find msg %s", err)
		return -1, err
	}

	if len(msgsToMigrate) == 0 {
		log.Warnf("MigrateMessagesToDedicatedTopic> No message to migrate for topic %s", topic.Topic)
		return 0, nil
	}

	nMigrated := 0
	for _, msgToMigrate := range msgsToMigrate {
		if errInsert := store.Tat().Session.DB(store.DatabaseName).C(topic.Collection).Insert(msgToMigrate); errInsert != nil {
			log.Errorf("MigrateMessagesToDedicatedTopic> getClMessages(toTopic).Insert(message), err: %s", errInsert)
			return nMigrated, errInsert
		}

		if errRemove := store.Tat().Session.DB(store.DatabaseName).C(store.CollectionDefaultMessages).RemoveId(msgToMigrate.ID); errRemove != nil {
			log.Errorf("MigrateMessagesToDedicatedTopic> getClMessages(toTopic).RemoveId(message), err: %s", errRemove)
			return nMigrated, errRemove
		}
		nMigrated++
	}

	return nMigrated, nil
}