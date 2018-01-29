package hook

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/ovh/tat"
	"github.com/spf13/viper"
)

var producer sarama.SyncProducer
var hookKafkaEnabled bool

func initKafka() {
	if viper.GetString("kafka_user") == "" || viper.GetString("kafka_password") == "" || viper.GetString("kafka_broker_addresses") == "" {
		log.Infof("No Kafka configured")
		return
	}

	c := sarama.NewConfig()
	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = true
	c.Net.TLS.Enable = true
	c.Net.SASL.Enable = true
	c.Net.SASL.User = viper.GetString("kafka_user")
	c.Net.SASL.Password = viper.GetString("kafka_password")

	var err error
	producer, err = sarama.NewSyncProducer(strings.Split(viper.GetString("kafka_broker_addresses"), ","), c)
	if err != nil {
		log.Errorf("Error with init sarama:%s (newSyncProducer)", err.Error())
	} else {
		hookKafkaEnabled = true
	}
	log.Infof("Kafka used at %s", viper.GetString("kafka_broker_addresses"))
}

// closeKafka closes producer
func closeKafka() {
	if producer != nil {
		if err := producer.Close(); err != nil {
			log.Errorf("Error with init sarama:%s (close)", err.Error())
		}
	}
}

// sendOnKafkaTopic send a hook on a topic kafka
func sendOnKafkaTopic(hook *tat.HookJSON, topicKafka string, topic tat.Topic) error {
	log.Debugf("sendOnKafkaTopic enter for post on kafka topic %s setted on tat topic %s", topicKafka, topic.Topic)

	if !hookKafkaEnabled {
		return fmt.Errorf("sendOnKafkaTopic: Kafka not initialized")
	}

	data, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{Topic: topicKafka, Value: sarama.ByteEncoder(data)}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}
	log.Debugf("Event sent to topic %s partition %d offset %d", topicKafka, partition, offset)
	return nil
}
