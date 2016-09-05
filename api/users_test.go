package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ovh/tat"
	"github.com/ovh/tat/api/tests"
	"github.com/stretchr/testify/assert"
)

var userController = &UsersController{}

// TestUserMe tests non-admin user, authenticated on tat
// GET on /user/me, check HTTP 200
func TestUserMe(t *testing.T) {
	tests.Init(t)

	r := tests.Router(t)
	tests.Handle(t, http.MethodGet, "/me", CheckPassword(), userController.Me)
	req, err := http.NewRequest(http.MethodGet, r.BasePath()+"/me", nil)
	req.Header.Add("X-Remote-User", tests.AdminUser)

	assert.NoError(t, err)
	response := tests.DoRequest(t, req)
	assert.Equal(t, 200, response.Code)
}

func createUser(t *testing.T) ([]byte, error) {
	tests.Handle(t, http.MethodPost, "/user", usersController.Create)
	username := "tat.integration.test.users." + tests.RandomString(t, 10)
	client := tests.TATClient(t, "")
	return client.UserAdd(tat.UserCreateJSON{
		Username: username,
		Fullname: fmt.Sprintf("User %s created for Tat Integration Test", username),
		Email:    fmt.Sprintf("%s@tat.foo", username),
	})
}

func TestCreateUser(t *testing.T) {
	tests.Init(t)
	tests.Router(t)
	r, err := createUser(t)
	assert.NotNil(t, r)
	assert.NoError(t, err)
	t.Logf("User created, return from tat:%s", r)
}
