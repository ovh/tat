package cache

import (
	log "github.com/Sirupsen/logrus"
)

// cleanAllByType cleans all keys
// tat:users:*:topics
// tat:users:*:topics:*
func cleanAllByType(key string) {
	keys, _ := Client().SMembers(key).Result()
	if len(keys) > 0 {
		log.Debugf("Clean cache on %d keys %s", len(keys), keys)
		Client().Del(keys...)
		removeSomeMembers(key, keys...)
	} else {
		log.Debugf("No cache to clean for key tat:users:*:topics:*")
	}
}

// CleanForUsernames cleans all keys for a username and topics
// tat:users:<username>:topics
// tat:users:<username>:topics:*
func cleanForUsernames(key, ktype string, usernames ...string) {
	for _, username := range usernames {
		log.Debugf("Cache CleanTopics for %s", username)
		k := Key("tat", "users", username, ktype)
		keys, _ := Client().SMembers(k).Result()
		if len(keys) > 0 {
			log.Debugf("Clean cache on %d keys %s", len(keys), keys)
			Client().Del(keys...)
			removeSomeMembers(key, append(keys, k)...)
			removeSomeMembers(k, keys...)
		}
	}
}

// CleanAllTopics cleans all keys
// tat:users:*:topics
// tat:users:*:topics:*
func CleanAllTopics() {
	log.Debugf("Cache CleanAllTopics")
	cleanAllByType(Key(TatTopicsKeys()...))
}

// CleanAllGroups cleans all keys
// tat:users:*:groups
// tat:users:*:groups:*
func CleanAllGroups() {
	log.Debugf("Cache CleanAllTopics")
	cleanAllByType(Key(TatGroupsKeys()...))
}

// CleanTopics cleans all keys for a username and topics
// tat:users:<username>:topics
// tat:users:<username>:topics:*
func CleanTopics(usernames ...string) {
	cleanForUsernames(Key(TatTopicsKeys()...), "topics", usernames...)
}

// CleanGroups cleans all keys for a username and groups
// tat:users:<username>:groups
// tat:users:<username>:groups:*
func CleanGroups(usernames ...string) {
	cleanForUsernames(Key(TatGroupsKeys()...), "groups", usernames...)
}

// CleanUsernames cleans tat:users:<username>
func CleanUsernames(usernames ...string) {
	for _, username := range usernames {
		k := Key("tat", "users", username)
		log.Debugf("Clean username key %s", k)
		Client().Del(k)
	}
}
