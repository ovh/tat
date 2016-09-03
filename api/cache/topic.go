package cache

import (
	log "github.com/Sirupsen/logrus"
)

// CleanAllTopics cleans all keys
// tat:users:*:topics
// tat:users:*:topics:*
func CleanAllTopics() {
	keys, _ := Client().Keys(Key("tat", "users", "*", "topics*")).Result()
	if len(keys) > 0 {
		log.Debugf("Clean cache on %d keys %s", len(keys), keys)
		Client().Del(keys...)
	} else {
		log.Debugf("No cache to clean for key tat:users:*:topics:*")
	}
}
