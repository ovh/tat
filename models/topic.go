package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ovh/tat/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Topic struct
type Topic struct {
	ID               string           `bson:"_id"          json:"_id,omitempty"`
	Topic            string           `bson:"topic"        json:"topic"`
	Description      string           `bson:"description"  json:"description"`
	ROGroups         []string         `bson:"roGroups"     json:"roGroups,omitempty"`
	RWGroups         []string         `bson:"rwGroups"     json:"rwGroups,omitempty"`
	ROUsers          []string         `bson:"roUsers"      json:"roUsers,omitempty"`
	RWUsers          []string         `bson:"rwUsers"      json:"rwUsers,omitempty"`
	AdminUsers       []string         `bson:"adminUsers"   json:"adminUsers,omitempty"`
	AdminGroups      []string         `bson:"adminGroups"  json:"adminGroups,omitempty"`
	History          []string         `bson:"history"      json:"history"`
	MaxLength        int              `bson:"maxlength"    json:"maxlength"`
	CanForceDate     bool             `bson:"canForceDate" json:"canForceDate"`
	CanUpdateMsg     bool             `bson:"canUpdateMsg" json:"canUpdateMsg"`
	CanDeleteMsg     bool             `bson:"canDeleteMsg" json:"canDeleteMsg"`
	CanUpdateAllMsg  bool             `bson:"canUpdateAllMsg" json:"canUpdateAllMsg"`
	CanDeleteAllMsg  bool             `bson:"canDeleteAllMsg" json:"canDeleteAllMsg"`
	IsROPublic       bool             `bson:"isROPublic"   json:"isROPublic"`
	DateModification int64            `bson:"dateModification" json:"dateModificationn,omitempty"`
	DateCreation     int64            `bson:"dateCreation" json:"dateCreation,omitempty"`
	Parameters       []TopicParameter `bson:"parameters" json:"parameters,omitempty"`
}

// TopicParameter struct, parameter on topics
type TopicParameter struct {
	Key   string `bson:"key"   json:"key"`
	Value string `bson:"value" json:"value"`
}

// TopicCriteria struct, used by List Topic
type TopicCriteria struct {
	Skip            int
	Limit           int
	IDTopic         string
	Topic           string
	Description     string
	DateMinCreation string
	DateMaxCreation string
	GetNbMsgUnread  string
	GetForTatAdmin  string
	Group           string
}

func buildTopicCriteria(criteria *TopicCriteria, user *User) bson.M {
	var query = []bson.M{}

	if criteria.IDTopic != "" {
		queryIDTopics := bson.M{}
		queryIDTopics["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.IDTopic, ",") {
			queryIDTopics["$or"] = append(queryIDTopics["$or"].([]bson.M), bson.M{"_id": val})
		}
		query = append(query, queryIDTopics)
	}
	if criteria.Topic != "" {
		queryTopics := bson.M{}
		queryTopics["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.Topic, ",") {
			queryTopics["$or"] = append(queryTopics["$or"].([]bson.M), bson.M{"topic": val})
		}
		query = append(query, queryTopics)
	}
	if criteria.Description != "" {
		queryDescriptions := bson.M{}
		queryDescriptions["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.Description, ",") {
			queryDescriptions["$or"] = append(queryDescriptions["$or"].([]bson.M), bson.M{"description": val})
		}
		query = append(query, queryDescriptions)
	}
	if criteria.Group != "" {
		queryGroups := bson.M{}
		queryGroups["$or"] = []bson.M{}
		queryGroups["$or"] = append(queryGroups["$or"].([]bson.M), bson.M{"adminGroups": bson.M{"$in": strings.Split(criteria.Group, ",")}})
		queryGroups["$or"] = append(queryGroups["$or"].([]bson.M), bson.M{"roGroups": bson.M{"$in": strings.Split(criteria.Group, ",")}})
		queryGroups["$or"] = append(queryGroups["$or"].([]bson.M), bson.M{"rwGroups": bson.M{"$in": strings.Split(criteria.Group, ",")}})
		query = append(query, queryGroups)
	}

	var bsonDate = bson.M{}

	if criteria.DateMinCreation != "" {
		i, err := strconv.ParseInt(criteria.DateMinCreation, 10, 64)
		if err != nil {
			log.Errorf("Error while parsing dateMinCreation %s", err)
		}
		tm := time.Unix(i, 0)

		if err == nil {
			bsonDate["$gte"] = tm.Unix()
		} else {
			log.Errorf("Error while parsing dateMinCreation %s", err)
		}
	}
	if criteria.DateMaxCreation != "" {
		i, err := strconv.ParseInt(criteria.DateMaxCreation, 10, 64)
		if err != nil {
			log.Errorf("Error while parsing dateMaxCreation %s", err)
		}
		tm := time.Unix(i, 0)

		if err == nil {
			bsonDate["$lte"] = tm.Unix()
		} else {
			log.Errorf("Error while parsing dateMaxCreation %s", err)
		}
	}
	if len(bsonDate) > 0 {
		query = append(query, bson.M{"dateCreation": bsonDate})
	}

	if criteria.GetForTatAdmin == "true" && user.IsAdmin {
		// requester is tat Admin and wants all topics, except /Private/* topics
		query = append(query, bson.M{
			"topic": bson.M{"$not": bson.RegEx{Pattern: "^\\/Private\\/.*", Options: "i"}},
		})
	} else if criteria.GetForTatAdmin == "true" && !user.IsAdmin {
		log.Warnf("User %s (not a TatAdmin) try to list all topics as an admin", user.Username)
	} else {
		bsonUser := []bson.M{}
		bsonUser = append(bsonUser, bson.M{"roUsers": bson.M{"$in": [1]string{user.Username}}})
		bsonUser = append(bsonUser, bson.M{"rwUsers": bson.M{"$in": [1]string{user.Username}}})
		bsonUser = append(bsonUser, bson.M{"adminUsers": bson.M{"$in": [1]string{user.Username}}})
		userGroups, err := user.GetGroupsOnlyName()
		if err != nil {
			log.Errorf("Error with getting groups for user %s", err)
		} else {
			bsonUser = append(bsonUser, bson.M{"roGroups": bson.M{"$in": userGroups}})
			bsonUser = append(bsonUser, bson.M{"rwGroups": bson.M{"$in": userGroups}})
			bsonUser = append(bsonUser, bson.M{"adminGroups": bson.M{"$in": userGroups}})
		}
		query = append(query, bson.M{"$or": bsonUser})
	}

	if len(query) > 0 {
		return bson.M{"$and": query}
	} else if len(query) == 1 {
		return query[0]
	}
	return bson.M{}
}

func getTopicSelectedFields(isAdmin bool) bson.M {
	if !isAdmin {
		return bson.M{
			"topic":           1,
			"description":     1,
			"isROPublic":      1,
			"canForceDate":    1,
			"canUpdateMsg":    1,
			"canDeleteMsg":    1,
			"canUpdateAllMsg": 1,
			"canDeleteAllMsg": 1,
			"maxlength":       1,
			"parameters":      1,
		}
	}
	return bson.M{}
}

// CountTopics return the total number of topics in db
func CountTopics() (int, error) {
	return Store().clTopics.Count()
}

// ListTopics returns list of topics, matching criterias
func ListTopics(criteria *TopicCriteria, user *User) (int, []Topic, error) {
	var topics []Topic

	cursor := listTopicsCursor(criteria, user)
	count, err := cursor.Count()
	if err != nil {
		log.Errorf("Error while count Topics %s", err)
	}

	err = cursor.Select(getTopicSelectedFields(user.IsAdmin)).
		Sort("topic").
		Skip(criteria.Skip).
		Limit(criteria.Limit).
		All(&topics)

	if err != nil {
		log.Errorf("Error while Find Topics %s", err)
	}

	if user.IsAdmin {
		return count, topics, err
	}

	var topicsUser []Topic
	// Get all topics where user is admin
	topicsMember, err := getTopicsForMemberUser(user, nil)
	if err != nil {
		return count, topics, err
	}

	for _, topic := range topics {
		added := false
		for _, topicMember := range topicsMember {
			if topic.ID == topicMember.ID {
				topic.AdminGroups = topicMember.AdminGroups
				topic.AdminUsers = topicMember.AdminUsers
				topic.ROUsers = topicMember.ROUsers
				topic.RWUsers = topicMember.RWUsers
				topic.RWGroups = topicMember.RWGroups
				topic.ROGroups = topicMember.ROGroups
				topicsUser = append(topicsUser, topic)
				added = true
				break
			}
		}
		if !added {
			topicsUser = append(topicsUser, topic)
		}
	}

	return count, topicsUser, err
}

// getTopicsForMemberUser where user is an admin or a member
func getTopicsForMemberUser(user *User, topic *Topic) ([]Topic, error) {
	var topics []Topic

	userGroups, err := user.GetGroupsOnlyName()
	c := bson.M{}
	c["$or"] = []bson.M{}
	c["$or"] = append(c["$or"].([]bson.M), bson.M{"adminUsers": bson.M{"$in": [1]string{user.Username}}})
	c["$or"] = append(c["$or"].([]bson.M), bson.M{"roUsers": bson.M{"$in": [1]string{user.Username}}})
	c["$or"] = append(c["$or"].([]bson.M), bson.M{"rwUsers": bson.M{"$in": [1]string{user.Username}}})
	if len(userGroups) > 0 {
		c["$or"] = append(c["$or"].([]bson.M), bson.M{"adminGroups": bson.M{"$in": userGroups}})
		c["$or"] = append(c["$or"].([]bson.M), bson.M{"roGroups": bson.M{"$in": userGroups}})
		c["$or"] = append(c["$or"].([]bson.M), bson.M{"rwGroups": bson.M{"$in": userGroups}})
	}

	if topic != nil {
		c["$and"] = []bson.M{}
		c["$and"] = append(c["$and"].([]bson.M), bson.M{"topic": topic.Topic})
	}

	err = Store().clTopics.Find(c).All(&topics)
	if err != nil {
		log.Errorf("Error while getting topics for member user: %s", err.Error())
	}

	return topics, err
}

func listTopicsCursor(criteria *TopicCriteria, user *User) *mgo.Query {
	return Store().clTopics.Find(buildTopicCriteria(criteria, user))
}

// InitPrivateTopic insert topic "/Private"
func InitPrivateTopic() {
	topic := &Topic{
		ID:              bson.NewObjectId().Hex(),
		Topic:           "/Private",
		Description:     "Private Topics",
		DateCreation:    time.Now().Unix(),
		MaxLength:       DefaultMessageMaxSize,
		CanForceDate:    false,
		CanUpdateMsg:    false,
		CanDeleteMsg:    false,
		CanUpdateAllMsg: false,
		CanDeleteAllMsg: false,
		IsROPublic:      false,
	}
	err := Store().clTopics.Insert(topic)
	log.Infof("Initialize /Private Topic")
	if err != nil {
		log.Fatalf("Error while initialize /Private Topic %s", err)
	}
}

// Insert creates a new topic. User is read write on topic
func (topic *Topic) Insert(user *User) error {
	err := topic.CheckAndFixName()
	if err != nil {
		return err
	}

	isParentRootTopic, parentTopic, err := topic.getParentTopic()
	if !isParentRootTopic {
		if err != nil {
			return fmt.Errorf("Parent Topic not found %s", topic.Topic)
		}

		// If user create a Topic in /Private/username, no check or RW to create
		if !strings.HasPrefix(topic.Topic, "/Private/"+user.Username) {
			// check if user can create topic in /topic
			hasRW := parentTopic.IsUserAdmin(user)
			if !hasRW {
				return fmt.Errorf("No RW access to parent topic %s", parentTopic.Topic)
			}
		}
	} else if !user.IsAdmin { // no parent topic, check admin
		return fmt.Errorf("No write access to create parent topic %s", topic.Topic)
	}

	var existing = &Topic{}
	err = existing.FindByTopic(topic.Topic, true, nil)
	if err == nil {
		return fmt.Errorf("Topic Already Exists : %s", topic.Topic)
	}

	topic.ID = bson.NewObjectId().Hex()
	topic.DateCreation = time.Now().Unix()
	topic.MaxLength = DefaultMessageMaxSize // topic MaxLenth messages
	topic.CanForceDate = false
	topic.IsROPublic = false

	if !isParentRootTopic {
		topic.ROGroups = parentTopic.ROGroups
		topic.RWGroups = parentTopic.RWGroups
		topic.ROUsers = parentTopic.ROUsers
		topic.RWUsers = parentTopic.RWUsers
		topic.AdminUsers = parentTopic.AdminUsers
		topic.AdminGroups = parentTopic.AdminGroups
		topic.MaxLength = parentTopic.MaxLength
		topic.CanForceDate = parentTopic.CanForceDate
		// topic.CanUpdateMsg can be set by user.createTopics for new users
		// with CanUpdateMsg=true
		if !topic.CanUpdateMsg {
			topic.CanUpdateMsg = parentTopic.CanUpdateMsg
		}
		// topic.CanDeleteMsg can be set by user.createTopics for new users
		// with CanDeleteMsg=true
		if !topic.CanDeleteMsg {
			topic.CanDeleteMsg = parentTopic.CanDeleteMsg
		}
		topic.CanUpdateAllMsg = parentTopic.CanUpdateAllMsg
		topic.CanDeleteAllMsg = parentTopic.CanDeleteAllMsg
		topic.IsROPublic = parentTopic.IsROPublic
		topic.Parameters = parentTopic.Parameters
	}

	err = Store().clTopics.Insert(topic)
	if err != nil {
		log.Errorf("Error while inserting new topic %s", err)
	}

	h := fmt.Sprintf("create a new topic :%s", topic.Topic)
	err = topic.addToHistory(bson.M{"_id": topic.ID}, user.Username, h)
	if err != nil {
		log.Errorf("Error while inserting history for new topic %s", err)
	}
	err = topic.AddRwUser(user.Username, user.Username, false)

	return err
}

// Delete deletes a topic from database
func (topic *Topic) Delete(user *User) error {

	// If user delete a Topic under /Private/username, no check or RW to delete
	if !strings.HasPrefix(topic.Topic, "/Private/"+user.Username) {
		// check if user is Tat admin or admin on this topic
		hasRW := topic.IsUserAdmin(user)
		if !hasRW {
			return fmt.Errorf("No RW access to topic %s (to delete it)", topic.Topic)
		}
	}

	c := &MessageCriteria{Topic: topic.Topic}
	msgs, err := ListMessages(c)
	if err != nil {
		return fmt.Errorf("Error while list Messages in Delete %s", err)
	}

	if len(msgs) > 0 {
		return fmt.Errorf("Could not delete this topic, this topic have messages")
	}

	return Store().clTopics.Remove(bson.M{"_id": topic.ID})
}

// Truncate removes all messages in a topic
func (topic *Topic) Truncate() (int, error) {
	changeInfo, err := Store().clMessages.RemoveAll(bson.M{"topics": bson.M{"$in": [1]string{topic.Topic}}})
	if err != nil {
		return 0, err
	}
	return changeInfo.Removed, err
}

// Get parent topic
// If it is a "root topic", like /myTopic, return true, nil, nil
func (topic *Topic) getParentTopic() (bool, *Topic, error) {
	index := strings.LastIndex(topic.Topic, "/")
	if index == 0 {
		return true, nil, nil
	}
	var nameParent = topic.Topic[0:index]
	var parentTopic = &Topic{}
	err := parentTopic.FindByTopic(nameParent, true, nil)
	if err != nil {
		log.Errorf("Error while fetching parent topic %s", err)
	}
	return false, parentTopic, err
}

// FindByTopic returns topic by topicName.
func (topic *Topic) FindByTopic(topicIn string, isAdmin bool, user *User) error {
	topic.Topic = topicIn
	err := topic.CheckAndFixName()
	if err != nil {
		return err
	}
	err = Store().clTopics.Find(bson.M{"topic": topic.Topic}).
		Select(getTopicSelectedFields(isAdmin)).
		One(&topic)

	if err != nil {
		log.Debugf("Error while fetching topic %s", topic.Topic)
		// TODO DM
		return err
	}

	if user != nil {
		// Get all topics where user is admin
		topicsMember, err := getTopicsForMemberUser(user, topic)
		if err != nil {
			return err
		}

		if len(topicsMember) == 1 {
			topic.AdminGroups = topicsMember[0].AdminGroups
			topic.AdminUsers = topicsMember[0].AdminUsers
			topic.ROUsers = topicsMember[0].ROUsers
			topic.RWUsers = topicsMember[0].RWUsers
			topic.RWGroups = topicsMember[0].RWGroups
			topic.ROGroups = topicsMember[0].ROGroups
		}
	}
	return err
}

// IsTopicExists return true if topic exists, false otherwise
func IsTopicExists(topic string) bool {
	var t = Topic{}
	return t.FindByTopic(topic, false, nil) == nil // no error, return true
}

// FindByID return topic, matching given id
func (topic *Topic) FindByID(id string, isAdmin bool) error {
	err := Store().clTopics.Find(bson.M{"_id": id}).
		Select(getTopicSelectedFields(isAdmin)).
		One(&topic)
	if err != nil {
		log.Errorf("Error while fecthing topic with id:%s", id)
	}
	return err
}

// SetParam update param maxLength, canForceDate, canUpdateMsg, canDeleteMsg, canUpdateAllMsg, canDeleteAllMsg, isROPublic, parameters on topic
func (topic *Topic) SetParam(username string, recursive bool, maxLength int,
	canForceDate, canUpdateMsg, canDeleteMsg, canUpdateAllMsg, canDeleteAllMsg,
	isROPublic bool, parameters []TopicParameter) error {

	var selector bson.M

	if recursive {
		selector = bson.M{"topic": bson.RegEx{Pattern: "^" + topic.Topic + ".*$"}}
	} else {
		selector = bson.M{"_id": topic.ID}
	}

	if maxLength <= 0 {
		maxLength = DefaultMessageMaxSize
	}

	update := bson.M{
		"maxlength":       maxLength,
		"canForceDate":    canForceDate,
		"canUpdateMsg":    canUpdateMsg,
		"canDeleteMsg":    canDeleteMsg,
		"canUpdateAllMsg": canUpdateAllMsg,
		"canDeleteAllMsg": canDeleteAllMsg,
		"isROPublic":      isROPublic,
	}

	if parameters != nil {
		update["parameters"] = parameters
	}
	_, err := Store().clTopics.UpdateAll(selector, bson.M{"$set": update})

	if err != nil {
		log.Errorf("Error while updateAll parameters : %s", err.Error())
		return err
	}
	h := fmt.Sprintf("update param to maxlength:%d, canForceDate:%t, canUpdateMsg:%t, canDeleteMsg:%t, canUpdateAllMsg:%t, canDeleteAllMsg:%t, isROPublic:%t", maxLength, canForceDate, canUpdateMsg, canDeleteMsg, canUpdateAllMsg, canDeleteAllMsg, isROPublic)
	return topic.addToHistory(selector, username, h)
}

func (topic *Topic) actionOnSetParameter(operand, set, admin string, newParam TopicParameter, recursive bool, history string) error {

	var selector bson.M

	if recursive {
		selector = bson.M{"topic": bson.RegEx{Pattern: "^" + topic.Topic + ".*$"}}
	} else {
		selector = bson.M{"_id": topic.ID}
	}

	var err error
	if operand == "$pull" {
		_, err = Store().clTopics.UpdateAll(
			selector,
			bson.M{operand: bson.M{set: bson.M{"key": newParam.Key}}},
		)
	} else {
		_, err = Store().clTopics.UpdateAll(
			selector,
			bson.M{operand: bson.M{set: bson.M{"key": newParam.Key, "value": newParam.Value}}},
		)
	}

	if err != nil {
		return err
	}
	return topic.addToHistory(selector, admin, history+" "+newParam.Key+":"+newParam.Value)
}

func (topic *Topic) actionOnSet(operand, set, username, admin string, recursive bool, history string) error {

	var selector bson.M

	if recursive {
		selector = bson.M{"topic": bson.RegEx{Pattern: "^" + topic.Topic + ".*$"}}
	} else {
		selector = bson.M{"_id": topic.ID}
	}

	_, err := Store().clTopics.UpdateAll(
		selector,
		bson.M{operand: bson.M{set: username}},
	)

	if err != nil {
		return err
	}
	return topic.addToHistory(selector, admin, history+" "+username)
}

// AddRoUser add a read only user to topic
func (topic *Topic) AddRoUser(admin string, username string, recursive bool) error {
	return topic.actionOnSet("$addToSet", "roUsers", username, admin, recursive, "add to ro")
}

// AddRwUser add a read write user to topic
func (topic *Topic) AddRwUser(admin string, username string, recursive bool) error {
	return topic.actionOnSet("$addToSet", "rwUsers", username, admin, recursive, "add to ro")
}

// AddAdminUser add a read write user to topic
func (topic *Topic) AddAdminUser(admin string, username string, recursive bool) error {
	return topic.actionOnSet("$addToSet", "adminUsers", username, admin, recursive, "add to admin")
}

// RemoveRoUser removes a read only user from topic
func (topic *Topic) RemoveRoUser(admin string, username string, recursive bool) error {
	return topic.actionOnSet("$pull", "roUsers", username, admin, recursive, "remove from ro")
}

// RemoveAdminUser removes a read only user from topic
func (topic *Topic) RemoveAdminUser(admin string, username string, recursive bool) error {
	return topic.actionOnSet("$pull", "roUsers", username, admin, recursive, "remove from admin")
}

// RemoveRwUser removes a read write user from topic
func (topic *Topic) RemoveRwUser(admin string, username string, recursive bool) error {
	return topic.actionOnSet("$pull", "rwUsers", username, admin, recursive, "remove from rw")
}

// AddRoGroup add a read only group to topic
func (topic *Topic) AddRoGroup(admin string, username string, recursive bool) error {
	return topic.actionOnSet("$addToSet", "roGroups", username, admin, recursive, "add to ro")
}

// AddRwGroup add a read write group to topic
func (topic *Topic) AddRwGroup(admin string, username string, recursive bool) error {
	return topic.actionOnSet("$addToSet", "rwGroups", username, admin, recursive, "add to ro")
}

// AddAdminGroup add a admin group to topic
func (topic *Topic) AddAdminGroup(admin string, username string, recursive bool) error {
	return topic.actionOnSet("$addToSet", "adminGroups", username, admin, recursive, "add to admin")
}

// RemoveAdminGroup removes a read write group from topic
func (topic *Topic) RemoveAdminGroup(admin string, username string, recursive bool) error {
	return topic.actionOnSet("$pull", "adminGroups", username, admin, recursive, "remove from admin")
}

// RemoveRoGroup removes a read only group from topic
func (topic *Topic) RemoveRoGroup(admin string, username string, recursive bool) error {
	return topic.actionOnSet("$pull", "roGroups", username, admin, recursive, "remove from ro")
}

// RemoveRwGroup removes a read write group from topic
func (topic *Topic) RemoveRwGroup(admin string, username string, recursive bool) error {
	return topic.actionOnSet("$pull", "rwGroups", username, admin, recursive, "remove from rw")
}

// AddParameter add a parameter to the topic
func (topic *Topic) AddParameter(admin string, parameterKey string, parameterValue string, recursive bool) error {
	return topic.actionOnSetParameter("$addToSet", "parameters", admin, TopicParameter{Key: parameterKey, Value: parameterValue}, recursive, "add to parameter")
}

// RemoveParameter removes a read only user from topic
func (topic *Topic) RemoveParameter(admin string, parameterKey string, parameterValue string, recursive bool) error {
	return topic.actionOnSetParameter("$pull", "parameters", admin, TopicParameter{Key: parameterKey, Value: ""}, recursive, "remove from parameters")
}

func (topic *Topic) addToHistory(selector bson.M, user string, historyToAdd string) error {
	toAdd := strconv.FormatInt(time.Now().Unix(), 10) + " " + user + " " + historyToAdd
	_, err := Store().clTopics.UpdateAll(
		selector,
		bson.M{"$addToSet": bson.M{"history": toAdd}},
	)
	return err
}

// IsUserRW return true if user can write on a this topic
// Check personal access to topic, and group access
func (topic *Topic) IsUserRW(user *User) bool {
	if utils.ArrayContains(topic.RWUsers, user.Username) ||
		utils.ArrayContains(topic.AdminUsers, user.Username) {
		return true
	}
	userGroups, err := user.GetGroups()
	if err != nil {
		log.Errorf("Error while fetching user groups")
		return false
	}

	var groups []string
	for _, g := range userGroups {
		groups = append(groups, g.Name)
	}

	return utils.ItemInBothArrays(topic.RWGroups, groups)
}

// IsUserReadAccess  return true if user has read access to topic
func (topic *Topic) IsUserReadAccess(user User) bool {
	currentTopic := topic

	if topic.IsROPublic {
		return true
	}

	// if user not admin, reload topic with admin rights
	if !user.IsAdmin {
		currentTopic = &Topic{}
		e := currentTopic.FindByID(topic.ID, true)
		if e != nil {
			return false
		}
	}

	if utils.ArrayContains(currentTopic.ROUsers, user.Username) ||
		utils.ArrayContains(currentTopic.RWUsers, user.Username) ||
		utils.ArrayContains(currentTopic.AdminUsers, user.Username) {
		return true
	}
	userGroups, err := user.GetGroups()
	if err != nil {
		log.Errorf("Error while fetching user groups for user %s", user.Username)
		return false
	}

	var groups []string
	for _, g := range userGroups {
		groups = append(groups, g.Name)
	}

	return utils.ItemInBothArrays(currentTopic.RWGroups, groups) ||
		utils.ItemInBothArrays(currentTopic.ROGroups, groups) ||
		utils.ItemInBothArrays(currentTopic.AdminGroups, groups)
}

// IsUserAdmin return true if user is Tat admin or is admin on this topic
// Check personal access to topic, and group access
func (topic *Topic) IsUserAdmin(user *User) bool {

	if user.IsAdmin {
		return true
	}

	if utils.ArrayContains(topic.AdminUsers, user.Username) {
		return true
	}

	userGroups, err := user.GetGroups()
	if err != nil {
		log.Errorf("Error while fetching user groups")
		return false
	}

	var groups []string
	for _, g := range userGroups {
		groups = append(groups, g.Name)
	}

	if utils.ItemInBothArrays(topic.AdminGroups, groups) {
		return true
	}

	// user is "Admin" on his /Private/usrname topics
	return strings.HasPrefix(topic.Topic, "/Private/"+user.Username)
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

// CheckAndFixName Add a / to topic name is it is not present
// return an error if length of name is < 4 or > 100
func (topic *Topic) CheckAndFixName() error {
	name, err := CheckAndFixNameTopic(topic.Topic)
	if err != nil {
		return err
	}
	topic.Topic = name
	return nil
}

func changeUsernameOnTopics(oldUsername, newUsername string) {
	changeNameOnSet("username", "roUsers", oldUsername, newUsername)
	changeNameOnSet("username", "rwUsers", oldUsername, newUsername)
	changeNameOnSet("username", "adminUsers", oldUsername, newUsername)
	changeUsernameOnPrivateTopics(oldUsername, newUsername)
}

func changeGroupnameOnTopics(oldGroupname, newGroupname string) {
	changeNameOnSet("groupname", "roGroups", oldGroupname, newGroupname)
	changeNameOnSet("groupname", "rwGroups", oldGroupname, newGroupname)
	changeNameOnSet("groupname", "adminGroups", oldGroupname, newGroupname)
}

func changeNameOnSet(typeChange, set, oldname, newname string) {
	_, err := Store().clTopics.UpdateAll(
		bson.M{set: oldname},
		bson.M{"$set": bson.M{set + ".$": newname}})

	if err != nil {
		log.Errorf("Error while changes %s from %s to %s on Topics (%s) %s", typeChange, oldname, newname, set, err)
	}
}

func changeUsernameOnPrivateTopics(oldUsername, newUsername string) error {
	var topics []Topic

	err := Store().clTopics.Find(
		bson.M{
			"topic": bson.RegEx{
				Pattern: "^/Private/" + oldUsername + ".*$", Options: "i",
			}}).All(&topics)

	if err != nil {
		log.Errorf("Error while getting topic with username %s for rename to %s on Topics %s", oldUsername, newUsername, err)
	}

	for _, topic := range topics {
		newTopicName := strings.Replace(topic.Topic, oldUsername, newUsername, 1)
		err := Store().clTopics.Update(
			bson.M{"_id": topic.ID},
			bson.M{"$set": bson.M{"topic": newTopicName}},
		)
		if err != nil {
			log.Errorf("Error while update Topic name from %s to %s :%s", topic.Topic, newTopicName, err)
		}
	}

	return err
}
