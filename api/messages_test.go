package main

import (
	"testing"

	"github.com/ovh/tat"
	"github.com/ovh/tat/api/tests"
	"github.com/stretchr/testify/assert"
)

var messagesCtrl = &MessagesController{}

func TestMessagesList(t *testing.T) {
	tests.Init(t)
	router := tests.Router(t)
	client := tests.TATClient(t, tests.AdminUser)

	initRoutesGroups(router, tests.FakeAuthHandler(t, tests.AdminUser, "X-TAT-TEST", true, false))
	initRoutesMessages(router, tests.FakeAuthHandler(t, tests.AdminUser, "X-TAT-TEST", true, false))
	initRoutesPresences(router, tests.FakeAuthHandler(t, tests.AdminUser, "X-TAT-TEST", true, false))
	initRoutesTopics(router, tests.FakeAuthHandler(t, tests.AdminUser, "X-TAT-TEST", true, false))
	initRoutesUsers(router, tests.FakeAuthHandler(t, tests.AdminUser, "X-TAT-TEST", true, false))
	initRoutesStats(router, tests.FakeAuthHandler(t, tests.AdminUser, "X-TAT-TEST", true, false))
	initRoutesSystem(router, tests.FakeAuthHandler(t, tests.AdminUser, "X-TAT-TEST", true, false))
	initRoutesSockets(router, tests.FakeAuthHandler(t, tests.AdminUser, "X-TAT-TEST", true, false))

	topic, err := client.TopicCreate(tat.TopicCreateJSON{
		Topic:       "/" + tests.RandomString(t, 10),
		Description: "this is a test",
	})

	assert.NotNil(t, topic)
	assert.NoError(t, err)
	if topic != nil {
		t.Logf("Topic %s created", topic.Topic)
	}

	defer client.TopicDelete(topic.Topic)
	defer client.TopicTruncate(topic.Topic)

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
	if len(messages.Messages) != 2 {
		t.Fail()
		return
	}

	if messages.Messages[0].DateCreation < messages.Messages[1].DateCreation {
		t.Log("Wrong order")
		t.Fail()
	}

	err = client.MessageDelete(message.Message.ID, topic.Topic, false, false)
	assert.NoError(t, err)

	messages, err = client.MessageList(topic.Topic, nil)
	assert.NotNil(t, topic)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(messages.Messages))

}
