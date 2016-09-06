package main

import (
	"net/http"
	"testing"

	"github.com/ovh/tat"
	"github.com/ovh/tat/api/tests"
	"github.com/stretchr/testify/assert"
)

var groupsController = GroupsController{}

func TestAddAndDeleteGroup(t *testing.T) {
	tests.Init(t)
	tests.Router(t)
	client := tests.TATClient(t, tests.AdminUser)
	tests.Handle(t, http.MethodPost, "/group", tests.FakeAuthHandler(t, tests.AdminUser, "TAT-TEST", true, false), groupsController.Create)
	tests.Handle(t, http.MethodDelete, "/group/edit/:group", tests.FakeAuthHandler(t, tests.AdminUser, "TAT-TEST", true, false), groupsController.Delete)

	group, err := client.GroupCreate(tat.GroupJSON{
		Description: "Group admin for tests",
		Name:        tests.RandomString(t, 10),
	})
	assert.NoError(t, err)

	err = client.GroupDelete(group.Name)
	assert.NoError(t, err)
}
