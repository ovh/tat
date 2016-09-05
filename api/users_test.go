package main

import (
	"net/http"
	"testing"

	"github.com/ovh/tat/api/tests"
	"github.com/stretchr/testify/assert"
)

var userController = &UsersController{}

// TestUserMe tests non-admin user, authenticated on tat
// GET on /user/me
func TestUserMe(t *testing.T) {
	tests.Init(t)

	r := tests.Router(t)
	tests.Handle(t, http.MethodGet, "/me", CheckPassword(), userController.Me)
	req, err := http.NewRequest(http.MethodGet, r.BasePath()+"/me", nil)
	req.Header.Add("X-Remote-User", "fsamin")

	assert.NoError(t, err)
	response := tests.DoRequest(t, req)
	assert.Equal(t, 200, response.Code)
}
