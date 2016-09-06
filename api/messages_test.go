package main

import (
	"net/http"
	"testing"

	"github.com/ovh/tat"
	"github.com/ovh/tat/api/tests"
	"github.com/stretchr/testify/assert"
)

var messagesController = &MessagesController{}
var topicsController = &TopicsController{}
var usersController = &UsersController{}

// TestUserMe tests non-admin user, authenticated on tat
// GET on /user/me
func TestMessagesList(t *testing.T) {
	tests.Init(t)
	tests.Router(t)

	tests.Handle(t, http.MethodPost, "/topic", tests.FakeAuthHandler(t, tests.AdminUser, "TAT-TEST", true, false), topicsController.Create)
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
}
