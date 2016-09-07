package cache

import (
	"strings"

	log "github.com/Sirupsen/logrus"
)

// cleanAllByType cleans all keys
func cleanAllByType(key string) {
	keys, _ := Client().SMembers(key).Result()
	if len(keys) > 0 {
		log.Debugf("Clean cache on %d keys %s", len(keys), keys)
		Client().Del(keys...)
		removeSomeMembers(key, keys...)
	}
}

// CleanForUsernames cleans all keys for a username and topics
// tat:users:<username>:topics
// tat:users:<username>:topics:*
func cleanTopicsListsForUsernames(key, ktype string, usernames ...string) {
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

// CleanAllTopicsLists cleans all keys
// tat:users:*:topics
// tat:users:*:topics:*
func CleanAllTopicsLists() {
	log.Debugf("Cache CleanAllTopicsLists")
	cleanAllByType(Key(TatTopicsKeys()...))
	cleanAllByType(Key(TatMessagesKeys()...))
}

// CleanAllGroups cleans all keys
// tat:users:*:groups
// tat:users:*:groups:*
func CleanAllGroups() {
	log.Debugf("Cache CleanAllTopics")
	cleanAllByType(Key(TatGroupsKeys()...))
}

// CleanTopicsList cleans all keys for a username and topics
// tat:users:<username>:topics
// tat:users:<username>:topics:*
func CleanTopicsList(usernames ...string) {
	cleanTopicsListsForUsernames(Key(TatTopicsKeys()...), "topics", usernames...)
}

// CleanGroups cleans all keys for a username and groups
// tat:users:<username>:groups
// tat:users:<username>:groups:*
func CleanGroups(usernames ...string) {
	cleanTopicsListsForUsernames(Key(TatGroupsKeys()...), "groups", usernames...)
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
