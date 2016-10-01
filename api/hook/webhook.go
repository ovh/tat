package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/facebookgo/httpcontrol"
	"github.com/ovh/tat"
	"github.com/spf13/viper"
)

var hookWebhookEnabled bool

func initWebhook() {
	hookWebhookEnabled = viper.GetBool("webhooks_enabled")
}

func sendWebHook(hook *tat.HookJSON, path string, topic tat.Topic, headerName, headerValue string) error {
	log.Debugf("sendWebHook >> enter for post webhook setted on topic %s", topic.Topic)

	data, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("POST", path, bytes.NewReader(data))

	if headerName != "" && headerValue != "" {
		req.Header.Add(headerName, headerValue)
	}

	c := &http.Client{
		Transport: &httpcontrol.Transport{
			RequestTimeout: 5 * time.Second,
			MaxTries:       3,
		},
	}

	resp, err := c.Do(req)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		return fmt.Errorf("sendWebHook >> err:%s", err)
	}

	body, errb := ioutil.ReadAll(resp.Body)
	if errb != nil {
		return fmt.Errorf("sendWebHook >> Error with ioutil.ReadAll %s", errb.Error())
	}

	if resp != nil && resp.StatusCode > 300 {
		log.Errorf("sendWebHook, err received: %d, body:%s", resp.StatusCode, body)
	} else {
		log.Debugf("Response from webhook %s", body)
	}

	return nil
}
