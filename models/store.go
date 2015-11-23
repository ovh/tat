package models

import (
	"os"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	databaseName        = "tat"
	collectionGroups    = "groups"
	collectionMessages  = "messages"
	collectionPresences = "presences"
	collectionTopics    = "topics"
	collectionUsers     = "users"
	collectionSockets   = "sockets"
)

// MongoStore stores MongoDB Session and collections
type MongoStore struct {
	session     *mgo.Session
	clGroups    *mgo.Collection
	clMessages  *mgo.Collection
	clPresences *mgo.Collection
	clTopics    *mgo.Collection
	clUsers     *mgo.Collection
	clSockets   *mgo.Collection
}

var _initCtx sync.Once
var _instance *MongoStore

// Store returns mongoDB instance
func Store() *MongoStore {
	return _instance
}

// NewStore initializes a new MongoDB Store
func NewStore() {
	log.Info("Mongodb : create new instance")
	var session *mgo.Session
	var err error

	username := getDbParameter("db_user")
	password := getDbParameter("db_password")
	replicaSetHostnamesTags := getDbParameter("db_rs_tags")

	address := viper.GetString("db_addr")

	if username != "" && password != "" {
		session, err = mgo.Dial("mongodb://" + username + ":" + password + "@" + address)
	} else {
		session, err = mgo.Dial("mongodb://" + address)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("Error with getting hostname: %s", err.Error())
	}

	session.Refresh()
	session.SetMode(mgo.SecondaryPreferred, true)

	if replicaSetHostnamesTags != "" && hostname != "" {
		log.Warnf("SelectServers try selectServer for %s with values %s", hostname, replicaSetHostnamesTags)
		tuples := strings.Split(replicaSetHostnamesTags, ",")
		for _, tuple := range tuples {
			t := strings.Split(tuple, ":")
			tupleHostname := t[0]
			if tupleHostname == hostname {
				tupleTagName := t[1]
				tupleTagValue := t[2]
				log.Warnf("SelectServers attach %s on replicaSet with tagName %s and value %s and %s", hostname, tupleTagName, tupleTagValue)
				session.SelectServers(bson.D{{tupleTagName, tupleTagValue}})
				break
			}
		}
	} else {
		log.Debugf("SelectServers No prefered server to select : %s", replicaSetHostnamesTags)
	}

	if err != nil {
		log.Fatalf("Error with getting Mongodb.Instance on address %s : %s", address, err)
		return
	}

	_instance = &MongoStore{
		session:     session,
		clGroups:    session.DB(databaseName).C(collectionGroups),
		clMessages:  session.DB(databaseName).C(collectionMessages),
		clPresences: session.DB(databaseName).C(collectionPresences),
		clTopics:    session.DB(databaseName).C(collectionTopics),
		clUsers:     session.DB(databaseName).C(collectionUsers),
		clSockets:   session.DB(databaseName).C(collectionSockets),
	}

	initDb()
	ensureIndexes(_instance)
}

// getDbParameter gets value of tat parameter
// return values if not "" AND not "false"
// used by db_user, db_password and db_rs_tags
func getDbParameter(key string) string {
	value := ""
	if viper.GetString(key) != "" && viper.GetString(key) != "false" {
		value = viper.GetString(key)
	}
	return value
}

func initDb() {
	nbTopics, err := CountTopics()
	if err != nil {
		log.Fatalf("Error with getting Mongodb.Instance %s", err)
		return
	}

	if nbTopics == 0 {
		// Create /Private topic
		InitPrivateTopic()
	}
	createDefaultGroup()
}

func ensureIndexes(store *MongoStore) {

	listIndex(store.clMessages, false)
	listIndex(store.clTopics, false)
	listIndex(store.clGroups, false)
	listIndex(store.clUsers, false)
	listIndex(store.clPresences, false)

	ensureIndex(store.clMessages, mgo.Index{Key: []string{"topics", "-dateUpdate", "-dateCreation"}})
	ensureIndex(store.clMessages, mgo.Index{Key: []string{"topics", "-dateCreation"}})
	ensureIndex(store.clMessages, mgo.Index{Key: []string{"tags"}})
	ensureIndex(store.clMessages, mgo.Index{Key: []string{"labels.text"}})
	ensureIndex(store.clMessages, mgo.Index{Key: []string{"inReplyOfID"}})
	ensureIndex(store.clMessages, mgo.Index{Key: []string{"inReplyOfIDRoot"}})
	ensureIndex(store.clTopics, mgo.Index{Key: []string{"topic"}, Unique: true})
	ensureIndex(store.clGroups, mgo.Index{Key: []string{"name"}, Unique: true})
	ensureIndex(store.clUsers, mgo.Index{Key: []string{"username"}, Unique: true})
	ensureIndex(store.clUsers, mgo.Index{Key: []string{"email"}, Unique: true})
	ensureIndex(store.clPresences, mgo.Index{Key: []string{"topic", "-dateTimePresence"}})
}

func listIndex(col *mgo.Collection, drop bool) {
	indexes, err := col.Indexes()
	if err != nil {
		log.Warnf("Error while getting index : %s", err)
	}
	for _, index := range indexes {
		log.Warnf("Info Index : Col %s : %+v", col.Name, index)
		if drop {
			err := col.DropIndex(index.Key...)
			if err != nil {
				log.Warnf("Error while dropping index : %s", err)
			}
		}
	}
}

func ensureIndex(col *mgo.Collection, index mgo.Index) {
	err := col.EnsureIndex(index)
	if err != nil {
		log.Fatalf("Error while creating index on %s:%s", col.Name, err)
		return
	}
}

func createDefaultGroup() {
	groupname := viper.GetString("default_group")

	// no default group
	if groupname == "" {
		return
	}

	if IsGroupnameExists(groupname) {
		log.Infof("Default Group %s already exist", groupname)
		return
	}

	var group = Group{
		Name:        groupname,
		Description: "Default Group",
	}

	err := group.Insert()
	if err != nil {
		log.Errorf("Error while Inserting default group %s", err)
	}

}

// RefreshStore calls Refresh on mongoDB Store, in order to avoid lost connection
func RefreshStore() {
	_instance.session.Refresh()
}

// DBServerStatus returns serverStatus cmd
func DBServerStatus() (bson.M, error) {
	result := bson.M{}
	err := _instance.session.Run(bson.D{{"serverStatus", 1}}, &result)
	return result, err
}

// DBStats returns dbstats cmd
func DBStats() (bson.M, error) {
	result := bson.M{}
	err := _instance.session.DB("tat").Run(bson.D{{"dbStats", 1}, {"scale", 1024}}, &result)
	return result, err
}

// GetCollectionNames returns collection names
func GetCollectionNames() ([]string, error) {
	return _instance.session.DB("tat").CollectionNames()
}

// DBStatsCollection returns stats for given collection
func DBStatsCollection(colName string) (bson.M, error) {
	result := bson.M{}
	err := _instance.session.DB("tat").Run(bson.D{{"collStats", colName}, {"scale", 1024}, {"indexDetails", true}}, &result)
	return result, err
}

// DBReplSetGetStatus returns replSetGetStatus cmd
func DBReplSetGetStatus() (bson.M, error) {
	result := bson.M{}
	err := _instance.session.Run(bson.D{{"replSetGetStatus", 1}}, &result)
	return result, err
}

// DBReplSetGetConfig returns replSetGetConfig cmd
func DBReplSetGetConfig() (bson.M, error) {
	result := bson.M{}
	err := _instance.session.Run(bson.D{{"replSetGetConfig", 1}}, &result)
	return result, err
}

// GetSlowestQueries returns the slowest queries
func GetSlowestQueries() ([]map[string]interface{}, error) {
	col := _instance.session.DB("tat").C("system.profile")
	var results []map[string]interface{}
	err := col.Find(bson.M{}).
		Sort("-millis").
		Skip(0).
		Limit(10).
		All(&results)
	return results, err
}
