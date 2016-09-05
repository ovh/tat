package main

import (
	"testing"

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

	topic, err := createTopic(t)
	assert.NotNil(t, topic)
	assert.NoError(t, err)
	t.Logf("Topic %s created", topic.Topic)
}
