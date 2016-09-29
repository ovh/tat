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
var kafkaOK bool

func initKafka() {
	c := sarama.NewConfig()
	c.ClientID = viper.GetString("kafka_client_id")

	var err error
	producer, err = sarama.NewSyncProducer(strings.Split(viper.GetString("kafka_broker_addresses"), ","), c)
	if err != nil {
		log.Errorf("Error with init sarama:%s (newSyncProducer)", err.Error())
	} else {
		kafkaOK = true
	}

}

// CloseKafka closes producer
func CloseKafka() {
	if producer != nil {
		if err := producer.Close(); err != nil {
			log.Errorf("Error with init sarama:%s (close)", err.Error())
		}
	}
}

// sendOnKafkaTopic send a hook on a topic kafka
func sendOnKafkaTopic(hook tat.Hook, topicKafka string) error {

	if !kafkaOK {
		return fmt.Errorf("sendOnKafkaTopic >> Kafka not initialized")
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
