package cache

import (
	"strings"

	log "github.com/Sirupsen/logrus"
)

// FlushDB flush cache
func FlushDB() (string, error) {
	r := instance.FlushDb()
	if r != nil && r.Err() != nil {
		log.Warnf("Error while flushing Cache: %s", r.Err().Error())
		return "", r.Err()
	}
	re, err := r.Result()
	out := ""
	if err != nil {
		log.Warnf("Error while flushing Cache, result: %s", err)
	} else {
		log.Infof("Cache Redis Cleaned: %s", re)
		needToFlush = false
		out = re
	}
	return out, err
}

// cleanAllByType cleans all keys
func cleanAllByType(key string) {
	keys, _ := Client().SMembers(key).Result()
	if len(keys) > 0 {
		log.Debugf("Clean cache on %d keys %s", len(keys), keys)
		Client().Del(keys...)
		removeSomeMembers(key, keys...)
	}
}

// CleanTopicByName cleans cache for a topic
func CleanTopicByName(topicName string) {
	// TODO To improve to remove only key with topic in arg
	cleanAllByType(Key(TatTopicsKeys()...))
}

// CleanAllTopicsLists cleans all keys
// tat:users:*:topics
// tat:users:*:topics:*
func CleanAllTopicsLists() {
	log.Debugf("Cache CleanAllTopicsLists")
	cleanAllByType(Key(TatTopicsKeys()...))
}

// CleanAllGroups cleans all keys
// tat:users:*:groups
// tat:users:*:groups:*
func CleanAllGroups() {
	log.Debugf("Cache CleanAllGroups")
	cleanAllByType(Key(TatGroupsKeys()...))
}

// CleanUsernames cleans tat:users:<username>
func CleanUsernames(usernames ...string) {
	for _, username := range usernames {
		k := Key("tat", "users", username)
		log.Debugf("Clean username key %s", k)
		Client().Del(k)
	}
}

// CleanMessagesLists cleans tat:messages:<topic>
func CleanMessagesLists(topic string) {
	key := Key(TatMessagesKeys()...)
	keys, _ := Client().SMembers(key).Result()
	keysMembers := []string{}
	members := []string{}
	if len(keys) > 0 {
		for _, k := range keys {
			if strings.HasPrefix(k, "tat:messages:"+topic) {
				log.Debugf("Clean cache on %s", k)
				keysMembers = append(keysMembers, k)
				members = append(members, k)
			}
		}
		Client().Del(keysMembers...)
		removeSomeMembers(key, members...)
	}
}
