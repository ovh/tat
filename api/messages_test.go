package main

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ovh/tat/api/tests"
	"github.com/stretchr/testify/assert"
)

var messagesController = &MessagesController{}
var topicsController = &TopicsController{}

func createTopic(t *testing.T, r *gin.RouterGroup) (string, error) {
	tests.Handle(t, http.MethodPost, "/topic", CheckPassword(), topicsController.Create)

}

// TestUserMe tests non-admin user, authenticated on tat
// GET on /user/me
func TestMessagesList(t *testing.T) {
	tests.Init(t)

	r := tests.Router(t)
	tests.Handle(t, http.MethodGet, "/*topic", CheckPassword(), messagesController.List)

	req, err := http.NewRequest(http.MethodGet, r.BasePath()+"/me", nil)
	req.Header.Add("X-Remote-User", "fsamin")

	assert.NoError(t, err)
	response := tests.DoRequest(t, req)
	assert.Equal(t, 200, response.Code)
}
