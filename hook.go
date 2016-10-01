package tat

import (
	"fmt"
)

// HookJSON represents a json sent to an external system, for a event about a message
type HookMessageJSON struct {
	Action         string          `json:"action"`
	Username       string          `json:"username"`
	MessageJSONOut *MessageJSONOut `json:"message"`
}

// HookJSON represents a json sent to an external system
type HookJSON struct {
	Hook        Hook             `json:"hook"`
	HookMessage *HookMessageJSON `json:"hookMessage"`
}

var HooksType = []string{HookTypeWebHook, HookTypeKafka, HookTypeXMPP}

var (
	HookTypeWebHook = "tathook-webhook"
	HookTypeKafka   = "tathook-kafka"
	HookTypeXMPP    = "tathook-xmpp"
	HookTypeXMPPOut = "tathook-xmpp-out"
	HookTypeXMPPIn  = "tathook-xmpp-in"
)

type Hook struct {
	Type        string `json:"type"` // in HooksType
	Destination string `json:"destination"`
}

func checkHook(h Hook) error {
	if h.Type != HookTypeKafka && h.Type != HookTypeWebHook {
		return fmt.Errorf("Invalid Hook type: %s", h.Type)
	}
	if h.Destination == "" {
		return fmt.Errorf("Invalid hook, destination is empty")
	}
	return nil
}

var HookTat2XMPPHeaderKey = "Tat2xmppkey"
