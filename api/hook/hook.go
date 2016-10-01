package hook

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/ovh/tat"
)

// InitHooks initializes hooks
func InitHooks() {
	initKafka()
	initWebhook()
	initXMPPHook()
}

// CloseHooks closes hooks
func CloseHooks() {
	closeKafka()
}

// SendHook sends a hook if topic contains hook parameter
func SendHook(hook *tat.HookJSON, topic tat.Topic) {
	go innerSendHook(hook, topic)
}

// GetCapabilities returns tat capabilities about hooks
func GetCapabilities() []tat.CapabilitieHook {
	hooks := []tat.CapabilitieHook{
		{HookType: tat.HookTypeKafka, HookEnabled: hookKafkaEnabled},
		{HookType: tat.HookTypeWebHook, HookEnabled: hookWebhookEnabled},
		{HookType: tat.HookTypeXMPP, HookEnabled: hookXMPPEnabled},
	}
	return hooks
}

func innerSendHook(hook *tat.HookJSON, topic tat.Topic) {
	for _, p := range topic.Parameters {
		h := &tat.HookJSON{
			HookMessage: hook.HookMessage,
			Hook: tat.Hook{
				Type:        p.Key,
				Destination: p.Value,
			},
		}
		if strings.HasPrefix(p.Key, tat.HookTypeWebHook) {
			if err := sendWebHook(h, p.Value, topic, "", ""); err != nil {
				log.Errorf("sendHook webhook err:%s", err)
			}
		} else if strings.HasPrefix(p.Key, tat.HookTypeKafka) {
			if err := sendOnKafkaTopic(h, p.Value, topic); err != nil {
				log.Errorf("sendHook kafka err:%s", err)
			}
		} else if p.Key == tat.HookTypeXMPP || p.Key == tat.HookTypeXMPPOut {
			if err := sendXMPP(h, p.Value, topic); err != nil {
				log.Errorf("sendHook XMPP err:%s", err)
			}
		}
	}
}
