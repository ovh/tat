package group

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ovh/tat"
	"github.com/ovh/tat/api/store"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// InitDB creates default group if necessary
func InitDB() {
	groupname := viper.GetString("default_group")

	// no default group
	if groupname == "" {
		return
	}

	if IsGroupnameExists(groupname) {
		log.Infof("Default Group %s already exist", groupname)
		return
	}

	var group = &tat.Group{
		Name:        groupname,
		Description: "Default Group",
	}

	if err := Insert(group); err != nil {
		log.Errorf("Error while Inserting default group %s", err)
	}
}

// user could be nil
func buildGroupCriteria(criteria *tat.GroupCriteria, user *tat.User) (bson.M, error) {
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

		if user != nil && !user.IsAdmin {
			queryUser := bson.M{}
			queryUser["$or"] = []bson.M{}
			queryUser["$or"] = []bson.M{}
			queryUser["$or"] = append(queryUser["$or"].([]bson.M), bson.M{"adminUsers": bson.M{"$in": [1]string{user.Username}}})
			queryUser["$or"] = append(queryUser["$or"].([]bson.M), bson.M{"users": bson.M{"$in": [1]string{user.Username}}})
			query = append(query, queryUser)
		}
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
			return bson.M{}, fmt.Errorf("Error while parsing dateMinCreation %s", err)
		}
		tm := time.Unix(i, 0)
		bsonDate["$gte"] = tm.Unix()
	}
	if criteria.DateMaxCreation != "" {
		i, err := strconv.ParseInt(criteria.DateMaxCreation, 10, 64)
		if err != nil {
			return bson.M{}, fmt.Errorf("Error while parsing dateMaxCreation %s", err)
		}
		tm := time.Unix(i, 0)
		bsonDate["$lte"] = tm.Unix()
	}
	if len(bsonDate) > 0 {
		query = append(query, bson.M{"dateCreation": bsonDate})
	}

	if len(query) > 0 {
		return bson.M{"$and": query}, nil
	} else if len(query) == 1 {
		return query[0], nil
	}
	return bson.M{}, nil
}

// ListGroups return all groups matching given criteria
func ListGroups(criteria *tat.GroupCriteria, user *tat.User, isAdmin bool) (int, []tat.Group, error) {
	var groups []tat.Group

	// TODO Cache

	cursor, errl := listGroupsCursor(criteria, user)
	if errl != nil {
		return -1, groups, errl
	}
	count, err := cursor.Count()
	if err != nil {
		log.Errorf("Error while count Groups %s", err)
	}

	selectedFields := bson.M{}
	if criteria.Name == "" {
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

// getGroupsForMemberUser where user is an admin or a member
func getGroupsForMemberUser(user *tat.User) ([]tat.Group, error) {
	var groups []tat.Group
	c := bson.M{}
	c["$or"] = []bson.M{}
	c["$or"] = append(c["$or"].([]bson.M), bson.M{"adminUsers": bson.M{"$in": [1]string{user.Username}}})
	c["$or"] = append(c["$or"].([]bson.M), bson.M{"users": bson.M{"$in": [1]string{user.Username}}})

	err := store.Tat().CGroups.Find(c).All(&groups)
	if err != nil {
		log.Errorf("Error while getting groups for admin user: %s", err.Error())
	}

	return groups, err
}

func listGroupsCursor(criteria *tat.GroupCriteria, user *tat.User) (*mgo.Query, error) {
	c, err := buildGroupCriteria(criteria, user)
	if err != nil {
		return nil, err
	}
	return store.Tat().CGroups.Find(c), nil
}

// Insert insert new group
func Insert(group *tat.Group) error {

	// TODO Clear Cache groups list

	group.ID = bson.NewObjectId().Hex()

	group.DateCreation = time.Now().Unix()
	err := store.Tat().CGroups.Insert(group)
	if err != nil {
		log.Errorf("Error while inserting new group %s", err)
	}
	return err
}

// FindByName returns matching group by groupname
func FindByName(groupname string) (*tat.Group, error) {

	c := &tat.GroupCriteria{
		Name:  groupname,
		Skip:  0,
		Limit: 1,
	}
	n, groups, err := ListGroups(c, nil, true)
	if n != 1 || len(groups) != 1 {
		return nil, fmt.Errorf("Error while fetching group with name %s", groupname)
	}

	return &groups[0], err
}

// IsGroupnameExists return true if groupname exists, false otherwise
func IsGroupnameExists(groupname string) bool {
	_, err := FindByName(groupname)
	if err != nil {
		return false // groupname does not exist
	}
	return true // groupname exists
}

func actionOnSet(group *tat.Group, operand, set, groupname, admin, history string) error {
	// TODO Clear Cache groups list

	err := store.Tat().CGroups.Update(
		bson.M{"_id": group.ID},
		bson.M{operand: bson.M{set: groupname}},
	)
	if err != nil {
		return err
	}
	return addToHistory(group, admin, history+" "+groupname)
}

// AddUser add a user to given group
func AddUser(group *tat.Group, admin string, username string) error {
	return actionOnSet(group, "$addToSet", "users", username, admin, "add")
}

// RemoveUser remove a user from a group
func RemoveUser(group *tat.Group, admin string, username string) error {
	return actionOnSet(group, "$pull", "users", username, admin, "remove")
}

// AddAdminUser add an admin to given group
func AddAdminUser(group *tat.Group, admin string, username string) error {
	return actionOnSet(group, "$addToSet", "adminUsers", username, admin, "add admin")
}

// RemoveAdminUser remove an admin from a group
func RemoveAdminUser(group *tat.Group, admin string, username string) error {
	return actionOnSet(group, "$pull", "adminUsers", username, admin, "remove admin")
}

func addToHistory(group *tat.Group, user string, historyToAdd string) error {
	toAdd := strconv.FormatInt(time.Now().Unix(), 10) + " " + user + " " + historyToAdd
	return store.Tat().CGroups.Update(
		bson.M{"_id": group.ID},
		bson.M{"$addToSet": bson.M{"history": toAdd}},
	)
}

// IsUserAdmin return true if user is admin on this group
func IsUserAdmin(group *tat.Group, username string) bool {
	return tat.ArrayContains(group.AdminUsers, username)
}

// CountGroups returns the total number of groups in db
func CountGroups() (int, error) {
	return store.Tat().CGroups.Count()
}

// Update updates a group : name and description
func Update(group *tat.Group, newGroupname, description string, user *tat.User) error {

	// TODO Clear Cache groups list

	// Check if name already exists -> checked in controller
	err := store.Tat().CGroups.Update(
		bson.M{"_id": group.ID},
		bson.M{"$set": bson.M{"name": newGroupname, "description": description}})

	if err != nil {
		log.Errorf("Error while update group %s to %s:%s", group.Name, newGroupname, err.Error())
		return fmt.Errorf("Error while update group")
	}
	group.Name = newGroupname
	group.Description = description

	return err
}

// Delete deletes a group
func Delete(group *tat.Group, user *tat.User) error {
	// TODO Clear Cache groups list

	if len(group.Users) > 0 {
		return fmt.Errorf("Could not delete this group, this group have Users")
	}
	if len(group.AdminUsers) > 0 {
		return fmt.Errorf("Could not delete this group, this group have Admin Users")
	}

	return store.Tat().CGroups.Remove(bson.M{"_id": group.ID})
}

// ChangeUsernameOnGroups changes a username on groups
func ChangeUsernameOnGroups(oldUsername, newUsername string) {
	// TODO Clear Cache groups list

	// Users
	_, err := store.Tat().CGroups.UpdateAll(
		bson.M{"users": oldUsername},
		bson.M{"$set": bson.M{"users.$": newUsername}})

	if err != nil {
		log.Errorf("Error while changes username from %s to %s on Groups (Users) %s", oldUsername, newUsername, err)
	}

	// AdminUsers
	_, err = store.Tat().CGroups.UpdateAll(
		bson.M{"adminUsers": oldUsername},
		bson.M{"$set": bson.M{"adminUsers.$": newUsername}})

	if err != nil {
		log.Errorf("Error while changes username from %s to %s on Groups (Admins) %s", oldUsername, newUsername, err)
	}
}

// GetUserGroupsOnlyName returns only groupname of user's groups
func GetUserGroupsOnlyName(username string) ([]string, error) {
	groups, err := GetGroups(username)
	if err != nil {
		return []string{}, err
	}

	arr := []string{}
	for _, g := range groups {
		arr = append(arr, g.Name)
	}
	return arr, nil
}

// GetGroups returns all user's groups
func GetGroups(username string) ([]tat.Group, error) {
	var groups []tat.Group

	// TODO Add users to criteria and use Group List

	err := store.Tat().CGroups.Find(bson.M{"users": bson.M{"$in": [1]string{username}}}).
		Sort("-name").
		All(&groups)

	if err != nil {
		log.Errorf("Error while Find groups for user %s error:%s", username, err)
	}
	return groups, err
}
