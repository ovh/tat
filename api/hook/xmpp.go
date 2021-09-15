package hook

import (
	"fmt"
	"strings"

	"github.com/ovh/tat"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var hookXMPPEnabled bool

func initXMPPHook() {
	hookXMPPEnabled = viper.GetString("tat2xmpp_url") != ""
}

func sendXMPP(hook *tat.HookJSON, path string, topic tat.Topic) error {
	if hook.HookMessage.MessageJSONOut.Message.Author.Username == viper.GetString("tat2xmpp_username") {
		log.Debugf("sendXMPP: Skip msg from %s on topic %s", viper.GetString("tat2xmpp_username"), topic.Topic)
		return nil
	}
	log.Debugf("sendXMPP: enter for post XMPP via tat2XMPP setted on topic %s", topic.Topic)

	// We split the XMPP servers and keys (comma separated lists)
	tat2xmppServers := strings.Split(viper.GetString("tat2xmpp_url"), ",")
	tat2xmppKeys := strings.Split(viper.GetString("tat2xmpp_key"), ",")

	// We must have the same number of servers and keys: one key for one server
	if len(tat2xmppServers) != len(tat2xmppKeys) {
		return fmt.Errorf("the number of XMPP servers differs from the number of provided keys (%v servers and %v keys)", len(tat2xmppServers), len(tat2xmppKeys))
	}

	// Go through the servers and send the hook with the right key
	// The right key for the right server is determined by the declaration order:
	// the first server goes with the first key, the second server goes with the second key...
	for index, tat2xmppServer := range tat2xmppServers {
		errSendWebHook := sendWebHook(hook, tat2xmppServer+"/hook", topic, tat.HookTat2XMPPHeaderKey, tat2xmppKeys[index])
		if errSendWebHook != nil {
			// If an error is encountered, abort everything and return the error because we should not encounter any error
			// even if we are sending the wrong destination to the wrong server (tat2xmpp will handle that and return no error)
			// So an error is not normal and we should return it immediately
			return errSendWebHook
		}
	}

	return nil
}
