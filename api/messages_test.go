package main

import (
	"net/http"
	"testing"

	"github.com/ovh/tat"
	"github.com/ovh/tat/api/tests"
	"github.com/stretchr/testify/assert"
)

var messagesCtrl = &MessagesController{}

func TestMessagesList(t *testing.T) {
	tests.Init(t)
	tests.Router(t)

	tests.Handle(t, http.MethodPost, "/topic", tests.FakeAuthHandler(t, tests.AdminUser, "TAT-TEST", true, false), topicsController.Create)
	tests.Handle(t, http.MethodPut, "/topic/truncate", tests.FakeAuthHandler(t, tests.AdminUser, "TAT-TEST", true, false), topicsController.Truncate)
	tests.Handle(t, http.MethodPut, "/topic/param", tests.FakeAuthHandler(t, tests.AdminUser, "TAT-TEST", true, false), topicsController.SetParam)
	tests.Handle(t, http.MethodDelete, "/topic/*topic", tests.FakeAuthHandler(t, tests.AdminUser, "TAT-TEST", true, false), topicsController.Delete)
	tests.Handle(t, http.MethodPost, "/message/*topic", tests.FakeAuthHandler(t, tests.AdminUser, "TAT-TEST", true, false), messagesCtrl.Create)
	tests.Handle(t, http.MethodDelete, "/message/nocascade/:idMessage/*topic", tests.FakeAuthHandler(t, tests.AdminUser, "TAT-TEST", true, false), messagesCtrl.Delete)
	tests.Handle(t, http.MethodGet, "/messages/*topic", tests.FakeAuthHandler(t, tests.AdminUser, "TAT-TEST", true, false), messagesCtrl.List)

	client := tests.TATClient(t, tests.AdminUser)
	topic, err := client.TopicCreate(tat.TopicCreateJSON{
		Topic:       "/" + tests.RandomString(t, 10),
		Description: "this is a test",
	})

	assert.NotNil(t, topic)
	assert.NoError(t, err)
	if topic != nil {
		t.Logf("Topic %s created", topic.Topic)
	}

	err = client.TopicParameter(topic.Topic, false, tat.TopicParameters{
		CanDeleteMsg:         true,
		AdminCanDeleteAllMsg: true,
	})
	assert.NoError(t, err)

	message, err := client.MessageAdd(tat.MessageJSON{
		Text:  "test test",
		Topic: topic.Topic,
	})
	assert.NotNil(t, message)
	assert.NoError(t, err)
	if topic != nil {
		t.Log(message.Info)
	}

	messages, err := client.MessageList(topic.Topic, nil)
	assert.NotNil(t, topic)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(messages.Messages))

	messages, err = client.MessageList(topic.Topic, nil)
	assert.NotNil(t, topic)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(messages.Messages))

	message, err = client.MessageAdd(tat.MessageJSON{
		Text:  "test2 test2",
		Topic: topic.Topic,
	})
	assert.NotNil(t, message)
	assert.NoError(t, err)
	if topic != nil {
		t.Log(message.Message)
	}

	messages, err = client.MessageList(topic.Topic, nil)
	assert.NotNil(t, topic)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(messages.Messages))

	err = client.MessageDelete(message.Message.ID, topic.Topic, false, false)
	assert.NoError(t, err)

	messages, err = client.MessageList(topic.Topic, nil)
	assert.NotNil(t, topic)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(messages.Messages))

	client.TopicTruncate(topic.Topic)
	client.TopicDelete(topic.Topic)

}
