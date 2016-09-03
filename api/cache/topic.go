package cache

import (
	log "github.com/Sirupsen/logrus"
)

const keyTatTopicsKeys = "tat:topics:keys"

// CleanAllTopics cleans all keys
// tat:users:*:topics
// tat:users:*:topics:*
func CleanAllTopics() {
	log.Debugf("Cache CleanAllTopics > enter")
	keys, _ := Client().SMembers(keyTatTopicsKeys).Result()
	if len(keys) > 0 {
		log.Debugf("Clean cache on %d keys %s", len(keys), keys)
		Client().Del(keys...)
		removeSomeMembers(keyTatTopicsKeys, keys)
	} else {
		log.Debugf("No cache to clean for key tat:users:*:topics:*")
	}
}

// CleanTopics cleans all keys for a username and topics
// tat:users:<username>:topics
// tat:users:<username>:topics:*
func CleanTopics(username string) {
	log.Debugf("Cache CleanTopics for %s", username)
	keys, _ := Client().SMembers(Key("tat", "users", username, "topics")).Result()
	if len(keys) > 0 {
		log.Debugf("Clean cache on %d keys %s", len(keys), keys)
		Client().Del(keys...)
		removeSomeMembers(keyTatTopicsKeys, keys)
	} else {
		log.Debugf("No cache to clean for vakey tat:users:%s:topics", keys)
	}
}
