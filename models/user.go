package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ovh/tat/utils"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Contact User Struct.
type Contact struct {
	Username string `bson:"username" json:"username"`
	Fullname string `bson:"fullname" json:"fullname"`
}

// Auth User Struct
type Auth struct {
	HashedPassword    string `bson:"hashedPassword" json:"-"`
	HashedTokenVerify string `bson:"hashedTokenVerify" json:"-"`
	DateRenewPassword int64  `bson:"dateRenewPassword" json:"dateRenewPassword"`
	DateAskReset      int64  `bson:"dateAskReset" json:"dateAskReset"`
	DateVerify        int64  `bson:"dateVerify" json:"dateVerify"`
	EmailVerified     bool   `bson:"emailVerified" json:"emailVerified"`
}

// User struct
type User struct {
	ID                     string    `bson:"_id"               json:"_id"`
	Username               string    `bson:"username"          json:"username"`
	Fullname               string    `bson:"fullname"          json:"fullname"`
	Email                  string    `bson:"email"             json:"email,omitempty"`
	Groups                 []string  `bson:"-"                 json:"groups,omitempty"`
	IsAdmin                bool      `bson:"isAdmin"           json:"isAdmin,omitempty"`
	IsSystem               bool      `bson:"isSystem"          json:"isSystem,omitempty"`
	IsArchived             bool      `bson:"isArchived"        json:"isArchived,omitempty"`
	CanWriteNotifications  bool      `bson:"canWriteNotifications" json:"canWriteNotifications,omitempty"`
	CanListUsersAsAdmin    bool      `bson:"canListUsersAsAdmin"   json:"canListUsersAsAdmin,omitempty"`
	FavoritesTopics        []string  `bson:"favoritesTopics"   json:"favoritesTopics,omitempty"`
	OffNotificationsTopics []string  `bson:"offNotificationsTopics"   json:"offNotificationsTopics,omitempty"`
	FavoritesTags          []string  `bson:"favoritesTags"     json:"favoritesTags,omitempty"`
	DateCreation           int64     `bson:"dateCreation"      json:"dateCreation,omitempty"`
	Contacts               []Contact `bson:"contacts"          json:"contacts,omitempty"`
	Auth                   Auth      `bson:"auth" json:"-"`
}

// UserCriteria is used to list users with criterias
type UserCriteria struct {
	Skip            int
	Limit           int
	WithGroups      bool
	IDUser          string
	Username        string
	Fullname        string
	DateMinCreation string
	DateMaxCreation string
}

func buildUserCriteria(criteria *UserCriteria) bson.M {
	var query = []bson.M{}
	query = append(query, bson.M{"isArchived": false})

	if criteria.IDUser != "" {
		queryIDUsers := bson.M{}
		queryIDUsers["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.IDUser, ",") {
			queryIDUsers["$or"] = append(queryIDUsers["$or"].([]bson.M), bson.M{"_id": val})
		}
		query = append(query, queryIDUsers)
	}
	if criteria.Username != "" {
		queryUsernames := bson.M{}
		queryUsernames["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.Username, ",") {
			queryUsernames["$or"] = append(queryUsernames["$or"].([]bson.M), bson.M{"username": val})
		}
		query = append(query, queryUsernames)
	}
	if criteria.Fullname != "" {
		queryFullnames := bson.M{}
		queryFullnames["$or"] = []bson.M{}
		for _, val := range strings.Split(criteria.Fullname, ",") {
			queryFullnames["$or"] = append(queryFullnames["$or"].([]bson.M), bson.M{"fullname": val})
		}
		query = append(query, queryFullnames)
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

func getUserListField(isAdmin bool) bson.M {
	if isAdmin {
		return bson.M{"username": 1,
			"fullname":              1,
			"email":                 1,
			"isAdmin":               1,
			"dateCreation":          1,
			"canWriteNotifications": 1,
			"canListUsersAsAdmin":   1,
		}
	}
	return bson.M{"username": 1,
		"fullname": 1,
	}
}

// ListUsers returns users list selected by criteria
func ListUsers(criteria *UserCriteria, isAdmin bool) (int, []User, error) {
	var users []User

	cursor := listUsersCursor(criteria, isAdmin)
	count, err := cursor.Count()
	if err != nil {
		log.Errorf("Error while count Users %s", err)
	}

	err = cursor.Select(getUserListField(isAdmin)).
		Sort("-dateCreation").
		Skip(criteria.Skip).
		Limit(criteria.Limit).
		All(&users)

	if err != nil {
		log.Errorf("Error while Find All Users %s", err)
	}

	// Admin could ask groups for all users. Not perf, but really rare
	if criteria.WithGroups && isAdmin {
		var usersWithGroups []User
		for _, u := range users {
			gs, err := u.GetGroupsOnlyName()
			u.Groups = gs
			log.Infof("User %s, Groups%s", u.Username, u.Groups)
			if err != nil {
				log.Errorf("Error while getting group for user %s, Error:%s", u.Username, err)
			}
			usersWithGroups = append(usersWithGroups, u)
		}
		return count, usersWithGroups, nil
	}
	return count, users, err
}

func listUsersCursor(criteria *UserCriteria, isAdmin bool) *mgo.Query {
	return Store().clUsers.Find(buildUserCriteria(criteria))
}

// Insert a new user, return tokenVerify to user, in order to
// validate account after check email
func (user *User) Insert() (string, error) {
	user.ID = bson.NewObjectId().Hex()

	user.DateCreation = time.Now().Unix()
	user.Auth.DateAskReset = time.Now().Unix()
	user.Auth.EmailVerified = false
	user.IsSystem = false
	user.IsArchived = false
	user.CanWriteNotifications = false
	user.CanListUsersAsAdmin = false
	nbUsers, err := CountUsers()
	if err != nil {
		log.Errorf("Error while count all users%s", err)
		return "", err
	}
	if nbUsers > 0 {
		user.IsAdmin = false
	} else {
		log.Infof("user %s is the first user, he is now admin", user.Username)
		user.IsAdmin = true
	}
	tokenVerify := ""
	tokenVerify, user.Auth.HashedTokenVerify, err = utils.GeneratePassword()
	if err != nil {
		log.Errorf("Error while generate Token Verify for new user %s", err)
		return tokenVerify, err
	}

	err = Store().clUsers.Insert(user)
	if err != nil {
		log.Errorf("Error while inserting new user %s", err)
	}
	return tokenVerify, err
}

// AskReset generate a new saltTokenVerify / hashedTokenVerify
// return tokenVerify (to be sent to user by mail)
func (user *User) AskReset() (string, error) {

	err := user.FindByUsernameAndEmail(user.Username, user.Email)
	if err != nil {
		return "", err
	}

	tokenVerify, hashedTokenVerify, err := utils.GeneratePassword()
	if err != nil {
		log.Errorf("Error while generate Token for reset password %s", err)
		return tokenVerify, err
	}

	err = Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"auth.hashedTokenVerify": hashedTokenVerify,
			"auth.dateAskReset":      time.Now().Unix(),
		}})

	if err != nil {
		log.Errorf("Error while ask reset user %s", err)
	}
	return tokenVerify, err
}

// Verify checks username and tokenVerify, if ok, return true, password if it's a new user
// Password is not stored in Database (only hashedPassword)
// return isNewUser, password, err
func (user *User) Verify(username, tokenVerify string) (bool, string, error) {
	emailVerified, err := user.findByUsernameAndTokenVerify(username, tokenVerify)
	if err != nil {
		return false, "", err
	}

	password, err := user.regenerateAndStoreAuth()

	if !emailVerified {
		log.Debugf("%s is a new user, ask to create his topics", username)
		user.createTopics()
		user.AddDefaultGroup()
	}

	return !emailVerified, password, err
}

func (user *User) regenerateAndStoreAuth() (string, error) {
	password, hashedPassword, err := utils.GeneratePassword()
	if err != nil {
		log.Errorf("Error while genereate password for user %s", err)
		return password, err
	}
	err = Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"auth.hashedTokenVerify": "", // reset tokenVerify
			"auth.hashedPassword":    hashedPassword,
			"auth.dateVerify":        time.Now().Unix(),
			"auth.dateRenewPassword": time.Now().Unix(),
			"auth.emailVerify":       true,
		}})

	if err != nil {
		log.Errorf("Error while updating user %s", err)
	}

	return password, err
}

func (user *User) getFieldsExceptAuth() bson.M {
	return bson.M{"username": 1,
		"fullname":               1,
		"email":                  1,
		"isAdmin":                1,
		"isSystem":               1,
		"isArchived":             1,
		"canWriteNotifications":  1,
		"canListUsersAsAdmin":    1,
		"dateCreation":           1,
		"favoritesTopics":        1,
		"offNotificationsTopics": 1,
		"favoritesTags":          1,
		"contacts":               1,
	}
}

// FindByUsernameAndPassword search username, use user's salt to generates hashedPassword
// and check username + hashedPassword in DB
func (user *User) FindByUsernameAndPassword(username, password string) error {
	var tmpUser = User{}
	err := Store().clUsers.
		Find(bson.M{"username": username}).
		Select(bson.M{"auth.hashedPassword": 1, "auth.saltPassword": 1}).
		One(&tmpUser)
	if err != nil {
		return fmt.Errorf("Error while fetching hash with username %s", username)
	}

	if !utils.IsCheckValid(password, tmpUser.Auth.HashedPassword) {
		return fmt.Errorf("Error while checking user %s with given password", username)
	}

	// ok, user is checked, get all fields now
	return user.FindByUsername(username)
}

// TrustUsername create user is not already registered
func (user *User) TrustUsername(username string) error {

	if !IsUsernameExists(username) {

		user.Username = username
		user.setEmailAndFullnameFromTrustedUsername()

		tokenVerify, err := user.Insert()
		if err != nil {
			return fmt.Errorf("TrustUsername, Error while Insert user %s : %s", username, err.Error())
		}

		_, _, err = user.Verify(username, tokenVerify)
		if err != nil {
			return fmt.Errorf("TrustUsername, Error while verify : %s", err.Error())
		}

		log.Infof("User %s created by TrustUsername", username)
	}

	// ok, user is checked, get all fields now
	return user.FindByUsername(username)
}

func (user *User) setEmailAndFullnameFromTrustedUsername() {
	conf := viper.GetString("trusted_usernames_emails_fullnames")
	if len(conf) < 2 {
		user.setEmailFromDefaultDomain()
		user.Fullname = user.Username
		return
	}

	tuples := strings.Split(conf, ",")

	for _, tuple := range tuples {
		t := strings.Split(tuple, ":")
		if len(t) != 3 {
			log.Errorf("Misconfiguration of trusted_usernames_emails tuple:%s", tuple)
			continue
		}
		usernameTuple := t[0]
		emailTuple := t[1]
		fullnameTuple := t[2]
		if usernameTuple == user.Username && emailTuple != "" && fullnameTuple != "" {
			user.Email = emailTuple
			user.Fullname = strings.Replace(fullnameTuple, "_", " ", -1)
			return
		}
	}
	// default behaviour
	user.setEmailFromDefaultDomain()
	user.Fullname = user.Username
}

func (user *User) setEmailFromDefaultDomain() {
	user.Email = user.Username + "@" + viper.GetString("default_domain")
}

// FindByUsernameAndPassword search username, use user's salt to generates tokenVerify
// and check username + hashedTokenVerify in DB
func (user *User) findByUsernameAndTokenVerify(username, tokenVerify string) (bool, error) {
	var tmpUser = User{}
	err := Store().clUsers.
		Find(bson.M{"username": username}).
		Select(bson.M{"auth.emailVerify": 1, "auth.hashedTokenVerify": 1, "auth.saltTokenVerify": 1, "auth.dateAskReset": 1}).
		One(&tmpUser)
	if err != nil {
		return false, fmt.Errorf("Error while fetching hashed Token Verify with username %s", username)
	}

	// dateAskReset more than 30 min, expire token
	if time.Since(time.Unix(tmpUser.Auth.DateAskReset, 0)).Minutes() > 30 {
		return false, fmt.Errorf("Token Validation expired. Please ask a reset of your password with username %s", username)
	}
	if !utils.IsCheckValid(tokenVerify, tmpUser.Auth.HashedTokenVerify) {
		return false, fmt.Errorf("Error while checking user %s with given token", username)
	}

	// ok, user is checked, get all fields now
	err = user.FindByUsername(username)
	if err != nil {
		return false, err
	}

	return tmpUser.Auth.EmailVerified, nil
}

//FindByUsernameAndEmail retrieve information from user with username
func (user *User) FindByUsernameAndEmail(username, email string) error {
	err := Store().clUsers.
		Find(bson.M{"username": username, "email": email}).
		Select(user.getFieldsExceptAuth()).
		One(&user)
	if err != nil {
		log.Errorf("Error while fetching user with username %s", username)
	}
	return err
}

//FindByUsername retrieve information from user with username
func (user *User) FindByUsername(username string) error {
	err := Store().clUsers.
		Find(bson.M{"username": username}).
		Select(user.getFieldsExceptAuth()).
		One(&user)
	if err != nil {
		log.Errorf("Error while fetching user with username %s", username)
	}
	return err
}

//FindByFullname retrieve information from user with fullname
func (user *User) FindByFullname(fullname string) error {
	err := Store().clUsers.
		Find(bson.M{"fullname": fullname}).
		Select(user.getFieldsExceptAuth()).
		One(&user)
	if err != nil {
		log.Errorf("Error while fetching user with fullname %s", fullname)
	}
	return err
}

//FindByEmail retrieve information from user with email
func (user *User) FindByEmail(email string) error {
	err := Store().clUsers.
		Find(bson.M{"email": email}).
		Select(user.getFieldsExceptAuth()).
		One(&user)
	if err != nil {
		log.Errorf("Error while fetching user with email %s", email)
	}
	return err
}

// IsEmailExists return true if email is already used, false otherwise
func IsEmailExists(email string) bool {
	var user = User{}

	err := user.FindByEmail(email)
	if err != nil {
		return false // user does not exist
	}
	return true // user exists
}

// IsUsernameExists return true if usernamer is already used, false otherwise
func IsUsernameExists(username string) bool {
	var user = User{}

	err := user.FindByUsername(username)
	if err != nil {
		return false // user does not exist
	}
	return true // user exists
}

// IsFullnameExists return true if fullname is already used, false otherwise
func IsFullnameExists(fullname string) bool {
	var user = User{}

	err := user.FindByFullname(fullname)
	if err != nil {
		return false // user does not exist
	}
	return true // user exists
}

// GetGroupsOnlyName returns only groupname  of user's groups
func (user *User) GetGroupsOnlyName() ([]string, error) {
	groups, err := user.GetGroups()
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
func (user *User) GetGroups() ([]Group, error) {
	var groups []Group

	err := Store().clGroups.Find(bson.M{"users": bson.M{"$in": [1]string{user.Username}}}).
		Sort("-name").
		All(&groups)

	if err != nil {
		log.Errorf("Error while Find groups for user %s error:%s", user.Username, err)
	}
	return groups, err
}

func (user *User) getFavoriteTopic(topic string) (string, error) {
	for _, cur := range user.FavoritesTopics {
		if cur == topic {
			return cur, nil
		}
	}
	l := ""
	return l, fmt.Errorf("topic %s not found in favorites topics of user", topic)
}

func (user *User) containsFavoriteTopic(topic string) bool {
	_, err := user.getFavoriteTopic(topic)
	if err == nil {
		return true
	}
	return false
}

// AddFavoriteTopic add a favorite topic to user
func (user *User) AddFavoriteTopic(topic string) error {
	if user.containsFavoriteTopic(topic) {
		return fmt.Errorf("AddFavoriteTopic not possible, %s is already a favorite topic", topic)
	}

	err := Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$push": bson.M{"favoritesTopics": topic}})
	if err != nil {
		return err
	}
	return nil
}

// RemoveFavoriteTopic removes a favorite topic from user
func (user *User) RemoveFavoriteTopic(topic string) error {
	topicName, err := CheckAndFixNameTopic(topic)
	if err != nil {
		return err
	}

	t, err := user.getFavoriteTopic(topicName)
	if err != nil {
		return fmt.Errorf("Remove favorite topic is not possible, %s is not a favorite of this user", topicName)
	}

	err = Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$pull": bson.M{"favoritesTopics": t}})

	if err != nil {
		return err
	}
	return nil
}

func (user *User) containsOffNotificationsTopic(topic string) bool {
	_, err := user.getOffNotificationsTopic(topic)
	if err == nil {
		return true
	}
	return false
}

func (user *User) getOffNotificationsTopic(topic string) (string, error) {
	for _, cur := range user.OffNotificationsTopics {
		if cur == topic {
			return cur, nil
		}
	}
	l := ""
	return l, fmt.Errorf("topic %s not found in off notifications topics of user", topic)
}

// EnableNotificationsTopic remove topic from user list offNotificationsTopics
func (user *User) EnableNotificationsTopic(topic string) error {

	topicName, err := CheckAndFixNameTopic(topic)
	if err != nil {
		return err
	}

	t, err := user.getOffNotificationsTopic(topicName)
	if err != nil {
		return fmt.Errorf("Enable notifications on topic %s is not possible, notifications are already enabled", topicName)
	}

	err = Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$pull": bson.M{"offNotificationsTopics": t}})

	if err != nil {
		return err
	}
	return nil
}

// DisableNotificationsTopic add topic to user list offNotificationsTopics
func (user *User) DisableNotificationsTopic(topic string) error {
	if user.containsOffNotificationsTopic(topic) {
		return fmt.Errorf("DisableNotificationsTopic not possible, notifications are already off on topic %s", topic)
	}

	err := Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$push": bson.M{"offNotificationsTopics": topic}})
	if err != nil {
		return err
	}
	return nil

}

func (user *User) getFavoriteTag(tag string) (string, error) {
	for _, cur := range user.FavoritesTags {
		if cur == tag {
			return cur, nil
		}
	}
	l := ""
	return l, fmt.Errorf("topic %s not found in favorites tags of user", tag)
}

func (user *User) containsFavoriteTag(tag string) bool {
	_, err := user.getFavoriteTag(tag)
	if err == nil {
		return true
	}
	return false
}

// AddFavoriteTag Add a favorite tag to user
func (user *User) AddFavoriteTag(tag string) error {
	if user.containsFavoriteTag(tag) {
		return fmt.Errorf("AddFavoriteTag not possible, %s is already a favorite tag", tag)
	}
	err := Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$push": bson.M{"favoritesTags": tag}})
	if err != nil {
		return err
	}
	return nil
}

// RemoveFavoriteTag remove a favorite tag from user
func (user *User) RemoveFavoriteTag(tag string) error {
	t, err := user.getFavoriteTag(tag)
	if err != nil {
		return fmt.Errorf("Remove favorite tag is not possible, %s is not a favorite of this user", tag)
	}

	err = Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$pull": bson.M{"favoritesTags": t}})

	if err != nil {
		return err
	}
	return nil
}

func (user *User) getContact(contactUsername string) (Contact, error) {
	for _, cur := range user.Contacts {
		if cur.Username == contactUsername {
			return cur, nil
		}
	}
	l := Contact{}
	return l, fmt.Errorf("contact %s not found", contactUsername)
}

func (user *User) containsContact(contactUsername string) bool {
	_, err := user.getContact(contactUsername)
	if err == nil {
		return true
	}
	return false
}

// AddContact add a contact to user
func (user *User) AddContact(contactUsername string, contactFullname string) error {
	if user.containsContact(contactUsername) {
		return fmt.Errorf("AddContact not possible, %s is already a contact of this user", contactUsername)
	}
	var newContact = &Contact{Username: contactUsername, Fullname: contactFullname}

	err := Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$push": bson.M{"contacts": newContact}})

	if err != nil {
		return err
	}
	return nil
}

// RemoveContact removes a contact from user
func (user *User) RemoveContact(contactUsername string) error {
	l, err := user.getContact(contactUsername)
	if err != nil {
		return fmt.Errorf("Remove Contact is not possible, %s is not a contact of this user", contactUsername)
	}

	err = Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$pull": bson.M{"contacts": l}})

	if err != nil {
		return err
	}
	return nil
}

// ConvertToSystem set attribute IsSysetm to true and suffix mail with a random string. If
// canWriteNotifications is true, this system user can write into /Private/username/Notifications topics
// canListUsersAsAdmin is true, this system user can view all user's fields as an admin (email, etc...)
// returns password, err
func (user *User) ConvertToSystem(userAdmin string, canWriteNotifications, canListUsersAsAdmin bool) (string, error) {
	email := fmt.Sprintf("%s$system$by$%s$%d", user.Email, userAdmin, time.Now().Unix())
	err := Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"email":                 email,
			"isSystem":              true,
			"canWriteNotifications": canWriteNotifications,
			"canListUsersAsAdmin":   canListUsersAsAdmin,
			"auth.emailVerified":    true,
		}})

	if err != nil {
		return "", err
	}

	return user.regenerateAndStoreAuth()
}

// UpdateSystemUser updates flags CanWriteNotifications and CanListUsersAsAdmin
func (user *User) UpdateSystemUser(canWriteNotifications, canListUsersAsAdmin bool) error {
	return Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"canWriteNotifications": canWriteNotifications,
			"canListUsersAsAdmin":   canListUsersAsAdmin,
		}})
}

// ResetSystemUserPassword reset a password for a system user
// returns newPassword
func (user *User) ResetSystemUserPassword() (string, error) {
	if !user.IsSystem {
		return "", fmt.Errorf("Reset password not possible, %s is not a system user", user.Username)
	}
	return user.regenerateAndStoreAuth()
}

// ConvertToAdmin set attribute IsAdmin to true
func (user *User) ConvertToAdmin(userAdmin string) error {
	log.Warnf("%s grant %s to admin", userAdmin, user.Username)
	err := Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"isAdmin": true}})

	if err != nil {
		return err
	}
	return nil
}

// Archive changes username of one user and set attribute email, username to random string
func (user *User) Archive(userAdmin string) error {
	newFullname := fmt.Sprintf("%s$archive$by$%s$%d", user.Fullname, userAdmin, time.Now().Unix())
	newUsername := fmt.Sprintf("%s$archive$by$%s$%d", user.Username, userAdmin, time.Now().Unix())
	email := fmt.Sprintf("%s$archive$by$%s$%d", user.Email, userAdmin, time.Now().Unix())
	err := Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"email": email, "fullname": newFullname, "isArchived": true}})

	if err != nil {
		return err
	}
	return user.Rename(newUsername)
}

// Rename changes username of one user
func (user *User) Rename(newUsername string) error {
	if IsUsernameExists(newUsername) {
		return fmt.Errorf("Username %s already exists", newUsername)
	}

	err := Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"username": newUsername}})

	if err != nil {
		return err
	}

	closeSocketOfUsername(user.Username)
	changeUsernameOnMessages(user.Username, newUsername)
	changeUsernameOnTopics(user.Username, newUsername)
	changeUsernameOnGroups(user.Username, newUsername)
	changeAuthorUsernameOnPresences(user.Username, newUsername)
	return nil
}

// Update changes fullname and email of user
func (user *User) Update(newFullname, newEmail string) error {

	if user.Email != newEmail && IsEmailExists(newEmail) {
		return fmt.Errorf("Email %s already exists", newEmail)
	}

	if user.Fullname != newFullname && IsFullnameExists(newFullname) {
		return fmt.Errorf("Fullname %s already exists", newFullname)
	}

	err := Store().clUsers.Update(
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"fullname": newFullname, "email": newEmail}})

	if err != nil {
		return err
	}
	return nil
}

// CountUsers returns the total number of users in db
func CountUsers() (int, error) {
	return Store().clUsers.Count()
}

func (user *User) createTopics() error {
	err := user.CreatePrivateTopic("")
	if err != nil {
		return err
	}
	err = user.CreatePrivateTopic("Bookmarks")
	if err != nil {
		return err
	}
	err = user.CreatePrivateTopic("Tasks")
	if err != nil {
		return err
	}
	err = user.CreatePrivateTopic("Notifications")
	if err != nil {
		return err
	}
	return nil
}

// CreatePrivateTopic creates a Private Topic. Name of topic will be :
// /Private/username and if subTopic != "", it will be :
// /Private/username/subTopic
// CanUpdateMsg, CanDeleteMsg set to true
func (user *User) CreatePrivateTopic(subTopic string) error {
	topic := "/Private/" + user.Username
	description := "Private Topic"

	if subTopic != "" {
		topic = fmt.Sprintf("%s/%s", topic, subTopic)
		description = fmt.Sprintf("%s - %s of %s", description, subTopic, user.Username)
	} else {
		description = fmt.Sprintf("%s - %s", description, user.Username)
	}
	t := &Topic{
		Topic:        topic,
		Description:  description,
		CanUpdateMsg: true,
		CanDeleteMsg: true,
	}
	e := t.Insert(user)
	if e != nil {
		log.Errorf("Error while creating Private topic %s: %s", topic, e.Error())
	}
	return e
}

// AddDefaultGroup add default group to user
func (user *User) AddDefaultGroup() error {
	groupname := viper.GetString("default_group")

	// no default group
	if groupname == "" {
		return nil
	}

	group := Group{}
	errfinding := group.FindByName(groupname)
	if errfinding != nil {
		e := fmt.Errorf("Error while fetching default group : %s", errfinding.Error())
		return e
	}
	err := group.AddUser("Tat", user.Username)
	if err != nil {
		e := fmt.Errorf("Error while adding user to default group : %s", err.Error())
		return e
	}
	return nil
}
