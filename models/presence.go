package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// UserPresence struct
type UserPresence struct {
	Username string `bson:"username" json:"username"`
	Fullname string `bson:"fullname" json:"fullname"`
}

// Presence struct
type Presence struct {
	ID               string       `bson:"_id,omitempty"    json:"_id"`
	Status           string       `bson:"status"           json:"status"`
	Topic            string       `bson:"topic"            json:"topic"`
	DatePresence     int64        `bson:"datePresence"     json:"datePresence"`
	DateTimePresence time.Time    `bson:"dateTimePresence" json:"dateTimePresence"`
	UserPresence     UserPresence `bson:"userPresence"     json:"userPresence"`
}

// PresenceCriteria used by Presences List
type PresenceCriteria struct {
	Skip            int
	Limit           int
	IDPresence      string
	Status          string
	Topic           string
	Username        string
	DateMinPresence string
	DateMaxPresence string
}

// PresencesJSON represents list of presences with count for total
type PresencesJSON struct {
	Count     int        `json:"count"`
	Presences []Presence `json:"presences"`
}

// PresenceJSONOut represents a presence
type PresenceJSONOut struct {
	Presence Presence `json:"presence"`
}

// PresenceJSON represents a status on a topic
type PresenceJSON struct {
	Status   string `json:"status" binding:"required"`
	Username string `json:"username,omitempty"`
	Topic    string
}

func buildPresenceCriteria(criteria *PresenceCriteria) (bson.M, error) {
	var query = []bson.M{}

	if criteria.Status != "" {
		queryStatus := bson.M{}
		queryStatus["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.Status, ",") {
			queryStatus["$or"] = append(queryStatus["$or"].([]bson.M), bson.M{"status": bson.RegEx{Pattern: "^.*" + regexp.QuoteMeta(val) + ".*$", Options: "i"}})
		}
		query = append(query, queryStatus)
	}
	if criteria.Username != "" {
		queryUsernames := bson.M{}
		queryUsernames["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.Username, ",") {
			queryUsernames["$or"] = append(queryUsernames["$or"].([]bson.M), bson.M{"userPresence.username": val})
		}
		query = append(query, queryUsernames)
	}
	if criteria.Topic != "" {
		queryTopics := bson.M{}
		queryTopics["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.Topic, ",") {
			queryTopics["$or"] = append(queryTopics["$or"].([]bson.M), bson.M{"topic": val})
		}
		query = append(query, queryTopics)
	}

	var bsonDate = bson.M{}

	if criteria.DateMinPresence != "" {
		i, err := strconv.ParseInt(criteria.DateMinPresence, 10, 64)
		if err != nil {
			return bson.M{}, fmt.Errorf("Error while parsing dateMinPresence %s", err)
		}
		tm := time.Unix(i, 0)
		bsonDate["$gte"] = tm.Unix()
	}
	if criteria.DateMaxPresence != "" {
		i, err := strconv.ParseInt(criteria.DateMaxPresence, 10, 64)
		if err != nil {
			return bson.M{}, fmt.Errorf("Error while parsing dateMaxPresence %s", err)
		}
		tm := time.Unix(i, 0)
		bsonDate["$lte"] = tm.Unix()
	}
	if len(bsonDate) > 0 {
		query = append(query, bson.M{"datePresence": bsonDate})
	}

	if len(query) > 0 {
		return bson.M{"$and": query}, nil
	} else if len(query) == 1 {
		return query[0], nil
	}
	return bson.M{}, nil
}

func getFieldsPresence(allFields bool) bson.M {

	if allFields {
		return bson.M{}
	}

	return bson.M{"_id": 0,
		"status":           1,
		"dateTimePresence": 1,
		"datePresence":     1,
		"userPresence":     1,
		"topic":            1,
	}
}

// ListPresencesAllFields returns list of presences, with given criteria
func ListPresencesAllFields(criteria *PresenceCriteria) (int, []Presence, error) {
	return listPresencesInternal(criteria, true)
}

// ListPresences returns list of presences, but only field status, dateTimePresence,datePresence,userPresence
func ListPresences(criteria *PresenceCriteria) (int, []Presence, error) {
	return listPresencesInternal(criteria, false)
}

func listPresencesInternal(criteria *PresenceCriteria, allFields bool) (int, []Presence, error) {
	var presences []Presence

	cursor, errl := listPresencesCursor(criteria, allFields)
	if errl != nil {
		return -1, presences, errl
	}
	count, err := cursor.Count()
	if err != nil {
		return -1, presences, fmt.Errorf("Error while count Presences %s", err)
	}
	err = cursor.Select(getFieldsPresence(allFields)).
		Sort("-datePresence").
		Skip(criteria.Skip).
		Limit(criteria.Limit).
		All(&presences)

	if err != nil {
		log.Errorf("Error while Find All Presences %s", err)
	}
	return count, presences, err
}

func listPresencesCursor(criteria *PresenceCriteria, allFields bool) (*mgo.Query, error) {
	c, err := buildPresenceCriteria(criteria)
	if err != nil {
		return nil, err
	}
	return Store().clPresences.Find(c), nil
}

// Upsert insert of update a presence (user / topic)
func (presence *Presence) Upsert(user User, topic Topic, status string) error {
	presence.Status = status
	if err := presence.checkAndFixStatus(); err != nil {
		return err
	}
	//	presence.ID = bson.NewObjectId().Hex()
	presence.Topic = topic.Topic
	var userPresence = UserPresence{}
	userPresence.Username = user.Username
	userPresence.Fullname = user.Fullname
	presence.UserPresence = userPresence
	now := time.Now()
	presence.DatePresence = now.Unix()
	presence.DateTimePresence = time.Now()

	//selector := ]bson.M{}
	//selector = append(selector, bson.M{"userpresence.username": userPresence.Username})
	//selector = append(selector, bson.M{"topic": topic.Topic})
	selector := bson.M{"topic": topic.Topic, "userPresence.username": userPresence.Username}
	if _, err := Store().clPresences.Upsert(selector, presence); err != nil {
		log.Errorf("Error while inserting new presence %s", err)
	}
	return nil
}

// truncate to 140 characters
// if len < 1, return error
func (presence *Presence) checkAndFixStatus() error {
	status := strings.TrimSpace(presence.Status)
	if len(status) < 1 {
		return fmt.Errorf("Invalid Status:%s", presence.Status)
	}

	validStatus := [...]string{"online", "offline", "busy"}
	find := false
	for _, s := range validStatus {
		if s == presence.Status {
			find = true
			break
		}
	}

	if !find {
		return fmt.Errorf("Invalid Status, should be online or offline or busy :%s", presence.Status)
	}
	presence.Status = status
	return nil
}

func changeAuthorUsernameOnPresences(oldUsername, newUsername string) error {
	_, err := Store().clPresences.UpdateAll(
		bson.M{"userPresence.username": oldUsername},
		bson.M{"$set": bson.M{"userPresence.username": newUsername}})

	if err != nil {
		log.Errorf("Error while update username from %s to %s on Presences %s", oldUsername, newUsername, err)
	}

	return err
}

// CountPresences returns the total number of presences in db
func CountPresences() (int, error) {
	return Store().clPresences.Count()
}

// Delete all presences of one user on one topic
func (presence *Presence) Delete(user User, topic Topic) error {
	_, err := Store().clPresences.RemoveAll(bson.M{"userPresence.username": user.Username, "topic": topic.Topic})
	return err
}

// CheckAllPresences detects duplicate
func CheckAllPresences() ([]bson.M, error) {

	pipeline := []bson.M{
		{"$group": bson.M{"_id": bson.M{"username": "$userPresence.username", "topic": "$topic"}, "count": bson.M{"$sum": 1}}},
		{"$match": bson.M{"count": bson.M{"$gt": 1}}},
	}

	pipe := Store().clPresences.Pipe(pipeline)
	results := []bson.M{}

	err := pipe.All(&results)
	return results, err
}
