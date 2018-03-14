package main

import (
	"testing"

	"github.com/ovh/tat"
	"github.com/ovh/tat/api/tests"
	"github.com/stretchr/testify/assert"
)

var groupsController = GroupsController{}

func TestAddAndDeleteGroup(t *testing.T) {
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

	group, err := client.GroupCreate(tat.GroupJSON{
		Description: "Group admin for tests",
		Name:        tests.RandomString(t, 10),
	})
	assert.NoError(t, err)

	err = client.GroupDelete(group.Name)
	assert.NoError(t, err)
}

func TestGroupsList(t *testing.T) {
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

	// Create a test group
	group, errGroupCreate := client.GroupCreate(tat.GroupJSON{
		Description: "Group admin for tests",
		Name:        tests.RandomString(t, 10),
	})
	assert.NotNil(t, group)
	assert.NoError(t, errGroupCreate)

	// Ensure that the created group will be deleted at the end of the test
	defer client.GroupDelete(group.Name)

	// Search the created group by name
	groupsName, errGroupListName := client.GroupList(&tat.GroupCriteria{
		Name: group.Name,
	})
	assert.NotNil(t, groupsName)
	assert.NoError(t, errGroupListName)
	assert.Equal(t, 1, len(groupsName.Groups))
	t.Log(groupsName)

	// Search the created group with a regular expression: use only the first 5 characters of the group and match the rest with ".*"
	groupsNameRegex, errGroupListNameRegex := client.GroupList(&tat.GroupCriteria{
		NameRegex: group.Name[:4] + ".*",
	})
	assert.NotNil(t, groupsNameRegex)
	assert.NoError(t, errGroupListNameRegex)
	assert.Equal(t, 1, len(groupsNameRegex.Groups))
	t.Log(groupsNameRegex)
}
