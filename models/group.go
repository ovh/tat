package models

import (
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ovh/tat/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Group struct
type Group struct {
	ID           string   `bson:"_id"          json:"_id"`
	Name         string   `bson:"name"         json:"name"`
	Description  string   `bson:"description"  json:"description"`
	Users        []string `bson:"users"        json:"users,omitempty"`
	AdminUsers   []string `bson:"adminUsers"   json:"adminUsers,omitempty"`
	DateCreation int64    `bson:"dateCreation" json:"dateCreation"`
}

// GroupCriteria is used by List all Groups
type GroupCriteria struct {
	Skip            int
	Limit           int
	IDGroup         string
	Name            string
	Description     string
	DateMinCreation string
	DateMaxCreation string
}

func buildGroupCriteria(criteria *GroupCriteria) bson.M {
	var query = []bson.M{}

	if criteria.IDGroup != "" {
		queryIDGroups := bson.M{}
		queryIDGroups["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.IDGroup, ",") {
			queryIDGroups["$or"] = append(queryIDGroups["$or"].([]bson.M), bson.M{"_id": val})
		}
		query = append(query, queryIDGroups)
	}
	if criteria.Name != "" {
		queryNames := bson.M{}
		queryNames["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.Name, ",") {
			queryNames["$or"] = append(queryNames["$or"].([]bson.M), bson.M{"name": val})
		}
		query = append(query, queryNames)
	}
	if criteria.Description != "" {
		queryDescriptions := bson.M{}
		queryDescriptions["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.Description, ",") {
			queryDescriptions["$or"] = append(queryDescriptions["$or"].([]bson.M), bson.M{"description": val})
		}
		query = append(query, queryDescriptions)
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

	if len(query) > 0 {
		return bson.M{"$and": query}
	} else if len(query) == 1 {
		return query[0]
	}
	return bson.M{}
}

// ListGroups return all groups matching given criterias
func ListGroups(criteria *GroupCriteria, isAdmin bool) (int, []Group, error) {
	var groups []Group

	cursor := listGroupsCursor(criteria, isAdmin)
	count, err := cursor.Count()
	if err != nil {
		log.Errorf("Error while count Groups %s", err)
	}

	selectedFields := bson.M{}
	if !isAdmin {
		selectedFields = bson.M{"name": 1, "description": 1}
	}
	err = cursor.Select(selectedFields).
		Sort("name").
		Skip(criteria.Skip).
		Limit(criteria.Limit).
		All(&groups)

	if err != nil {
		log.Errorf("Error while Find All Groups %s", err)
	}
	return count, groups, err
}

func listGroupsCursor(criteria *GroupCriteria, isAdmin bool) *mgo.Query {
	return Store().clGroups.Find(buildGroupCriteria(criteria))
}

// Insert insert new group
func (group *Group) Insert() error {
	group.ID = bson.NewObjectId().Hex()

	group.DateCreation = time.Now().Unix()
	err := Store().clGroups.Insert(group)
	if err != nil {
		log.Errorf("Error while inserting new group %s", err)
	}
	return err
}

// FindByName returns matching group by groupname
func (group *Group) FindByName(groupname string) error {
	err := Store().clGroups.Find(bson.M{"name": groupname}).One(&group)
	if err != nil {
		log.Errorf("Error while fetching group with name %s", groupname)
	}
	return err
}

// IsGroupnameExists return true if groupname exists, false otherwise
func IsGroupnameExists(groupname string) bool {
	var group = Group{}
	err := group.FindByName(groupname)
	if err != nil {
		return false // groupname does not exist
	}
	return true // groupname exists
}

func (group *Group) actionOnSet(operand, set, groupname, admin, history string) error {
	err := Store().clGroups.Update(
		bson.M{"_id": group.ID},
		bson.M{operand: bson.M{set: groupname}},
	)
	if err != nil {
		return err
	}
	return group.addToHistory(admin, history+" "+groupname)
}

// AddUser add a user to given group
func (group *Group) AddUser(admin string, username string) error {
	return group.actionOnSet("$addToSet", "users", username, admin, "add")
}

// RemoveUser remove a user from a group
func (group *Group) RemoveUser(admin string, username string) error {
	return group.actionOnSet("$pull", "users", username, admin, "remove")
}

// AddAdminUser add an admin to given group
func (group *Group) AddAdminUser(admin string, username string) error {
	return group.actionOnSet("$addToSet", "adminUsers", username, admin, "add admin")
}

// RemoveAdminUser remove an admin from a group
func (group *Group) RemoveAdminUser(admin string, username string) error {
	return group.actionOnSet("$pull", "adminUsers", username, admin, "remove admin")
}

func (group *Group) addToHistory(user string, historyToAdd string) error {
	toAdd := strconv.FormatInt(time.Now().Unix(), 10) + " " + user + " " + historyToAdd
	err := Store().clGroups.Update(
		bson.M{"_id": group.ID},
		bson.M{"$addToSet": bson.M{"history": toAdd}},
	)
	return err
}

// IsUserAdmin return true if user is admin on this group
func (group *Group) IsUserAdmin(user *User) bool {
	if utils.ArrayContains(group.AdminUsers, user.Username) {
		return true
	}
	return false
}

// CountGroups returns the total number of groups in db
func CountGroups() (int, error) {
	return Store().clGroups.Count()
}

func changeUsernameOnGroups(oldUsername, newUsername string) {
	// Users
	_, err := Store().clGroups.UpdateAll(
		bson.M{"users": oldUsername},
		bson.M{"$set": bson.M{"users.$": newUsername}})

	if err != nil {
		log.Errorf("Error while changes username from %s to %s on Groups (Users) %s", oldUsername, newUsername, err)
	}

	// AdminUsers
	_, err = Store().clGroups.UpdateAll(
		bson.M{"adminUsers": oldUsername},
		bson.M{"$set": bson.M{"adminUsers.$": newUsername}})

	if err != nil {
		log.Errorf("Error while changes username from %s to %s on Groups (Admins) %s", oldUsername, newUsername, err)
	}

}
