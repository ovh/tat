package main

import (
	"net/http"
	"testing"

	"github.com/ovh/tat/api/tests"
	"github.com/stretchr/testify/assert"
)

var userController = &UsersController{}

func TestMe(t *testing.T) {
	tests.Init(t)

	r := tests.Router(t)
	tests.HandleGET(t, "/me", CheckPassword(), userController.Me)
	req, err := http.NewRequest("GET", r.BasePath()+"/me", nil)
	req.Header.Add("X-Remote-User", "fsamin")

	assert.NoError(t, err)
	response := tests.DoRequest(t, req)
	assert.Equal(t, 200, response.Code)
}
