package main

import (
	"net/http"
	"testing"

	"github.com/ovh/tat"
	"github.com/ovh/tat/api/tests"
	"github.com/stretchr/testify/assert"
)

func createTopic(t *testing.T, path string) (*tat.Topic, error) {
	tests.Handle(t, http.MethodPost, "/topic", tests.FakeAuthHandler(t, tests.AdminUser, "TAT-TEST", true, false), topicsController.Create)
	client := tests.TATClient(t, tests.AdminUser)
	return client.TopicCreate(tat.TopicCreateJSON{
		Topic:       path + "/" + tests.RandomString(t, 10),
		Description: "this is a test",
	})
}

func TestTopicCreate(t *testing.T) {
	tests.Init(t)
	tests.Router(t)
	topic, err := createTopic(t, "")
	assert.NotNil(t, topic)
	assert.NoError(t, err)
	if topic == nil {
		t.Fail()
		return
	}
	t.Logf("Topic %s created", topic.Topic)
	assert.NotZero(t, topic.ID)
}
