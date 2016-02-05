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

// Group struct
type Group struct {
	ID           string   `bson:"_id"          json:"_id"`
	Name         string   `bson:"name"         json:"name"`
	Description  string   `bson:"description"  json:"description"`
	Users        []string `bson:"users"        json:"users,omitempty"`
	AdminUsers   []string `bson:"adminUsers"   json:"adminUsers,omitempty"`
	DateCreation int64    `bson:"dateCreation" json:"dateCreation,omitempty"`
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
func ListGroups(criteria *GroupCriteria, user *User, isAdmin bool) (int, []Group, error) {
	var groups []Group

	cursor := listGroupsCursor(criteria)
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
	if isAdmin {
		return count, groups, err
	}

	var groupsUser []Group
	// Get all groups where user is admin
	groupsMember, err := getGroupsForMemberUser(user)
	if err != nil {
		return count, groups, err
	}

	for _, group := range groups {
		added := false
		for _, groupMember := range groupsMember {
			if group.ID == groupMember.ID {
				groupsUser = append(groupsUser, groupMember)
				added = true
				break
			}
		}
		if !added {
			groupsUser = append(groupsUser, group)
		}
	}

	return count, groupsUser, err
}

// getGroupsForMemberUser where user is an admin or a member
func getGroupsForMemberUser(user *User) ([]Group, error) {
	var groups []Group
	c := bson.M{}
	c["$or"] = []bson.M{}
	c["$or"] = append(c["$or"].([]bson.M), bson.M{"adminUsers": bson.M{"$in": [1]string{user.Username}}})
	c["$or"] = append(c["$or"].([]bson.M), bson.M{"users": bson.M{"$in": [1]string{user.Username}}})

	err := Store().clGroups.Find(c).All(&groups)
	if err != nil {
		log.Errorf("Error while getting groups for admin user: %s", err.Error())
	}

	return groups, err
}

func listGroupsCursor(criteria *GroupCriteria) *mgo.Query {
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
	return Store().clGroups.Update(
		bson.M{"_id": group.ID},
		bson.M{"$addToSet": bson.M{"history": toAdd}},
	)
}

// IsUserAdmin return true if user is admin on this group
func (group *Group) IsUserAdmin(user *User) bool {
	return utils.ArrayContains(group.AdminUsers, user.Username)
}

// CountGroups returns the total number of groups in db
func CountGroups() (int, error) {
	return Store().clGroups.Count()
}

// Update updates a group : name and description
func (group *Group) Update(newGroupname, description string, user *User) error {

	// Check if name already exists -> checked in controller
	err := Store().clGroups.Update(
		bson.M{"_id": group.ID},
		bson.M{"$set": bson.M{"name": newGroupname, "description": description}})

	if err != nil {
		log.Errorf("Error while update group %s to %s:%s", group.Name, newGroupname, err.Error())
		return fmt.Errorf("Error while update group")
	}

	if newGroupname != group.Name {
		changeGroupnameOnTopics(group.Name, newGroupname)
	}

	return err
}

// Delete deletes a group
func (group *Group) Delete(user *User) error {
	if len(group.Users) > 0 {
		return fmt.Errorf("Could not delete this group, this group have Users")
	}
	if len(group.AdminUsers) > 0 {
		return fmt.Errorf("Could not delete this group, this group have Admin Users")
	}

	c := TopicCriteria{}
	c.Skip = 0
	c.Limit = 10
	c.Group = group.Name

	count, topics, err := ListTopics(&c, user)
	if err != nil {
		log.Errorf("Error while getting topics associated to group %s:%s", group.Name, err.Error())
		return fmt.Errorf("Error while getting topics associated to group")
	}

	if len(topics) > 0 {
		e := fmt.Sprintf("Group %s associated to %d topic, you can't delete it", group.Name, count)
		log.Errorf(e)
		return fmt.Errorf(e)
	}

	return Store().clGroups.Remove(bson.M{"_id": group.ID})
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
