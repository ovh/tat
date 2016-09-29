package hook

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/ovh/tat"
)

// InitHooks initializes hooks
func InitHooks() {
	initKafka()
}

func sendHook(hook tat.Hook, topic tat.Topic) {
	for _, p := range topic.Parameters {
		if strings.HasPrefix(p.Key, "webhook") {
			if err := sendWebHook(hook, p.Value); err != nil {
				log.Errorf("sendHook webhook err:%s", err)
			}
		} else if strings.HasPrefix(p.Key, "kafka-topic") {
			if err := sendOnKafkaTopic(hook, p.Value); err != nil {
				log.Errorf("sendHook kafka err:%s", err)
			}
		}
	}
}
