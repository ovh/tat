package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/ovh/tat/models"
	"github.com/ovh/tat/utils"
)

// GroupsController contains all methods about groups manipulation
type GroupsController struct{}

type groupsJSON struct {
	Count  int            `json:"count"`
	Groups []models.Group `json:"groups"`
}

type paramUserJSON struct {
	Groupname string `json:"groupname"`
	Username  string `json:"username"`
}

type groupJSON struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

func (*GroupsController) buildCriteria(ctx *gin.Context) *models.GroupCriteria {
	c := models.GroupCriteria{}
	skip, e := strconv.Atoi(ctx.DefaultQuery("skip", "0"))
	if e != nil {
		skip = 0
	}
	c.Skip = skip

	limit, e2 := strconv.Atoi(ctx.DefaultQuery("limit", "100"))
	if e2 != nil {
		limit = 100
	}
	c.Limit = limit
	c.IDGroup = ctx.Query("idGroup")
	c.Name = ctx.Query("name")
	c.Description = ctx.Query("description")
	c.DateMinCreation = ctx.Query("dateMinCreation")
	c.DateMaxCreation = ctx.Query("dateMaxCreation")
	return &c
}

// List list groups with given criterias
func (g *GroupsController) List(ctx *gin.Context) {
	var criteria models.GroupCriteria
	ctx.Bind(&criteria)

	user, err := PreCheckUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching current user"})
		return
	}
	count, groups, err := models.ListGroups(g.buildCriteria(ctx), &user, utils.IsTatAdmin(ctx))
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	out := &groupsJSON{
		Count:  count,
		Groups: groups,
	}
	ctx.JSON(http.StatusOK, out)
}

// Create creates a new group
func (*GroupsController) Create(ctx *gin.Context) {
	var groupJSON groupJSON
	ctx.Bind(&groupJSON)

	var groupIn models.Group
	groupIn.Name = groupJSON.Name
	groupIn.Description = groupJSON.Description

	err := groupIn.Insert()
	if err != nil {
		log.Errorf("Error while InsertGroup %s", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusCreated, groupIn)
}

func (*GroupsController) preCheckUser(ctx *gin.Context, paramJSON *paramUserJSON) (models.Group, error) {
	usernameExists := models.IsUsernameExists(paramJSON.Username)
	group := models.Group{}
	if !usernameExists {
		e := errors.New("username " + paramJSON.Username + " does not exist")
		ctx.AbortWithError(http.StatusInternalServerError, e)
		return group, e
	}
	errfinding := group.FindByName(paramJSON.Groupname)
	if errfinding != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errfinding)
		return group, errfinding
	}

	if utils.IsTatAdmin(ctx) { // if Tat admin, ok
		return group, nil
	}

	user, err := PreCheckUser(ctx)
	if err != nil {
		return models.Group{}, err
	}

	if !group.IsUserAdmin(&user) {
		e := fmt.Errorf("user %s is not admin on group %s", user.Username, group.Name)
		ctx.AbortWithError(http.StatusInternalServerError, e)
		return models.Group{}, e
	}

	return group, nil
}

type groupUpdateJSON struct {
	Name        string `json:"newName" binding:"required"`
	Description string `json:"newDescription" binding:"required"`
}

// Update a group
// only for Tat admin
func (g *GroupsController) Update(ctx *gin.Context) {
	var paramJSON groupUpdateJSON
	ctx.Bind(&paramJSON)

	group, err := GetParam(ctx, "group")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Group in query"})
		return
	}

	groupToUpdate := models.Group{}
	err = groupToUpdate.FindByName(group)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Group"})
		return
	}

	if paramJSON.Name != groupToUpdate.Name {
		groupnameExists := models.IsGroupnameExists(paramJSON.Name)

		if groupnameExists {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Group Name already exists"})
			return
		}
	}

	user, err := PreCheckUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error while fetching user"})
		return
	}

	err = groupToUpdate.Update(paramJSON.Name, paramJSON.Description, &user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error while update group"})
		return
	}
	ctx.JSON(http.StatusCreated, "")
}

// Delete deletes requested group
// only for Tat admin
func (g *GroupsController) Delete(ctx *gin.Context) {
	var paramJSON groupUpdateJSON
	ctx.Bind(&paramJSON)

	group, err := GetParam(ctx, "group")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Group in query"})
		return
	}

	groupToDelete := models.Group{}
	err = groupToDelete.FindByName(group)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Group"})
		return
	}

	user, err := PreCheckUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error while fetching user"})
		return
	}

	err = groupToDelete.Delete(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while deleting Group: %s", err.Error())})
		return
	}
	ctx.JSON(http.StatusOK, "")
}

// AddUser add a user to a group
func (g *GroupsController) AddUser(ctx *gin.Context) {
	var paramJSON paramUserJSON
	ctx.Bind(&paramJSON)
	group, e := g.preCheckUser(ctx, &paramJSON)
	if e != nil {
		return
	}
	err := group.AddUser(utils.GetCtxUsername(ctx), paramJSON.Username)
	if err != nil {
		return
	}
	ctx.JSON(http.StatusCreated, "")
}

// RemoveUser removes user from a group
func (g *GroupsController) RemoveUser(ctx *gin.Context) {
	var paramJSON paramUserJSON
	ctx.Bind(&paramJSON)
	group, e := g.preCheckUser(ctx, &paramJSON)
	if e != nil {
		return
	}

	err := group.RemoveUser(utils.GetCtxUsername(ctx), paramJSON.Username)
	if err != nil {
		return
	}
	ctx.JSON(http.StatusOK, "")
}

// AddAdminUser add a user to a group
func (g *GroupsController) AddAdminUser(ctx *gin.Context) {
	var paramJSON paramUserJSON
	ctx.Bind(&paramJSON)
	group, e := g.preCheckUser(ctx, &paramJSON)
	if e != nil {
		return
	}
	err := group.AddAdminUser(utils.GetCtxUsername(ctx), paramJSON.Username)
	if err != nil {
		return
	}
	ctx.JSON(http.StatusCreated, "")
}

// RemoveAdminUser removes user from a group
func (g *GroupsController) RemoveAdminUser(ctx *gin.Context) {
	var paramJSON paramUserJSON
	ctx.Bind(&paramJSON)
	group, e := g.preCheckUser(ctx, &paramJSON)
	if e != nil {
		return
	}

	err := group.RemoveAdminUser(utils.GetCtxUsername(ctx), paramJSON.Username)
	if err != nil {
		return
	}
	ctx.JSON(http.StatusOK, "")
}
