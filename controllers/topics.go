package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/ovh/tat/models"
	"github.com/ovh/tat/utils"
)

// TopicsController contains all methods about topics manipulation
type TopicsController struct{}

type topicsJSON struct {
	Count                int            `json:"count"`
	Topics               []models.Topic `json:"topics"`
	CountTopicsMsgUnread int            `json:"countTopicsMsgUnread"`
	TopicsMsgUnread      map[string]int `json:"topicsMsgUnread"`
}

type topicJSON struct {
	Topic *models.Topic `json:"topic"`
}

type paramTopicUserJSON struct {
	Topic     string `json:"topic"` // topic topic
	Username  string `json:"username"`
	Recursive bool   `json:"recursive"`
}

type topicCreateJSON struct {
	Topic       string `json:"topic" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type topicParameterJSON struct {
	Topic     string `json:"topic"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Recursive bool   `json:"recursive"`
}

func (*TopicsController) buildCriteria(ctx *gin.Context) *models.TopicCriteria {
	c := models.TopicCriteria{}
	skip, e := strconv.Atoi(ctx.DefaultQuery("skip", "0"))
	if e != nil {
		skip = 0
	}
	c.Skip = skip
	limit, e2 := strconv.Atoi(ctx.DefaultQuery("limit", "200"))
	if e2 != nil {
		limit = 200
	}
	c.Limit = limit
	c.IDTopic = ctx.Query("idTopic")
	c.Topic = ctx.Query("topic")
	if c.Topic != "" && !strings.HasPrefix(c.Topic, "/") {
		c.Topic = "/" + c.Topic
	}
	c.Description = ctx.Query("description")
	c.DateMinCreation = ctx.Query("dateMinCreation")
	c.DateMaxCreation = ctx.Query("dateMaxCreation")
	c.GetNbMsgUnread = ctx.Query("getNbMsgUnread")
	c.GetForTatAdmin = ctx.Query("getForTatAdmin")
	return &c
}

// List returns the list of topics that can be viewed by user
func (t *TopicsController) List(ctx *gin.Context) {
	criteria := t.buildCriteria(ctx)
	var user = &models.User{}
	err := user.FindByUsername(utils.GetCtxUsername(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching user."})
		return
	}
	count, topics, err := models.ListTopics(criteria, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching topics."})
		return
	}

	out := &topicsJSON{Topics: topics, Count: count}

	if criteria.GetNbMsgUnread == "true" {
		c := &models.PresenceCriteria{
			Username: user.Username,
		}
		count, presences, err := models.ListPresencesAllFields(c)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		unread := make(map[string]int)
		knownPresence := false
		for _, topic := range topics {
			if utils.ArrayContains(user.OffNotificationsTopics, topic.Topic) {
				continue
			}
			knownPresence = false
			for _, presence := range presences {
				if topic.Topic != presence.Topic {
					continue
				}
				knownPresence = true

				nb, err := models.CountMsgSinceDate(presence.Topic, presence.DatePresence)
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, err)
					return
				}
				unread[presence.Topic] = nb
			}
			if !knownPresence {
				unread[topic.Topic] = -1
			}
		}
		out.TopicsMsgUnread = unread
		out.CountTopicsMsgUnread = count
	}
	ctx.JSON(http.StatusOK, out)
}

// OneTopic returns only requested topic, and only if user has read access
func (t *TopicsController) OneTopic(ctx *gin.Context) {
	topicRequest, err := GetParam(ctx, "topic")
	if err != nil {
		return
	}
	var user = models.User{}
	err = user.FindByUsername(utils.GetCtxUsername(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching user."})
		return
	}
	topic := &models.Topic{}
	errfinding := topic.FindByTopic(topicRequest, user.IsAdmin)
	if errfinding != nil {
		ctx.JSON(http.StatusInternalServerError, errfinding)
		return
	}

	isReadAccess := topic.IsUserReadAccess(user)
	if !isReadAccess {
		ctx.JSON(http.StatusInternalServerError, errors.New("No Read Access to this topic: "+user.Username+" "+topic.Topic))
		return
	}
	out := &topicJSON{Topic: topic}
	ctx.JSON(http.StatusOK, out)
}

// Create creates a new topic
func (*TopicsController) Create(ctx *gin.Context) {
	var topicIn topicCreateJSON
	ctx.Bind(&topicIn)

	var user = models.User{}
	err := user.FindByUsername(utils.GetCtxUsername(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching user."})
		return
	}

	var topic models.Topic
	topic.Topic = topicIn.Topic
	topic.Description = topicIn.Description

	err = topic.Insert(&user)
	if err != nil {
		log.Errorf("Error while InsertTopic %s", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusCreated, topic)
}

// Delete deletes requested topic only if user is Tat admin, or admin on topic
func (t *TopicsController) Delete(ctx *gin.Context) {

	topicRequest, err := GetParam(ctx, "topic")
	if err != nil {
		return
	}

	var user = models.User{}
	err = user.FindByUsername(utils.GetCtxUsername(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching user."})
		return
	}

	paramJSON := paramTopicUserJSON{
		Topic:     topicRequest,
		Username:  user.Username,
		Recursive: false,
	}

	topic, e := t.preCheckUser(ctx, &paramJSON)
	if e != nil {
		return
	}
	err = topic.Delete(&user)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, "")
}

func (t *TopicsController) preCheckUser(ctx *gin.Context, paramJSON *paramTopicUserJSON) (models.Topic, error) {
	usernameExists := models.IsUsernameExists(paramJSON.Username)

	if !usernameExists {
		e := errors.New("username " + paramJSON.Username + " does not exist")
		ctx.AbortWithError(http.StatusInternalServerError, e)
		return models.Topic{}, e
	}

	return t.preCheckUserAdminOnTopic(ctx, paramJSON.Topic)
}

func (t *TopicsController) preCheckGroup(ctx *gin.Context, paramJSON *paramGroupJSON) (models.Topic, error) {
	groupnameExists := models.IsGroupnameExists(paramJSON.Groupname)

	if !groupnameExists {
		e := errors.New("groupname" + paramJSON.Groupname + " does not exist")
		ctx.AbortWithError(http.StatusInternalServerError, e)
		return models.Topic{}, e
	}

	return t.preCheckUserAdminOnTopic(ctx, paramJSON.Topic)
}

func (t *TopicsController) preCheckUserAdminOnTopic(ctx *gin.Context, topicName string) (models.Topic, error) {
	topic := models.Topic{}
	errfinding := topic.FindByTopic(topicName, true)
	if errfinding != nil {
		e := errors.New(errfinding.Error())
		ctx.AbortWithError(http.StatusInternalServerError, e)
		return topic, e
	}

	if utils.IsTatAdmin(ctx) { // if Tat admin, ok
		return topic, nil
	}

	user, err := PreCheckUser(ctx)
	if err != nil {
		return models.Topic{}, err
	}

	if !topic.IsUserAdmin(&user) {
		e := fmt.Errorf("user %s is not admin on topic %s", user.Username, topic.Topic)
		ctx.AbortWithError(http.StatusForbidden, e)
		return models.Topic{}, e
	}

	return topic, nil
}

// AddRoUser add a readonly user on selected topic
func (t *TopicsController) AddRoUser(ctx *gin.Context) {
	var paramJSON paramTopicUserJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckUser(ctx, &paramJSON)
	if e != nil {
		return
	}
	err := topic.AddRoUser(utils.GetCtxUsername(ctx), paramJSON.Username, paramJSON.Recursive)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}
	ctx.JSON(http.StatusCreated, "")
}

// AddRwUser add a read / write user on selected topic
func (t *TopicsController) AddRwUser(ctx *gin.Context) {
	var paramJSON paramTopicUserJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckUser(ctx, &paramJSON)
	if e != nil {
		return
	}

	err := topic.AddRwUser(utils.GetCtxUsername(ctx), paramJSON.Username, paramJSON.Recursive)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, "")
}

// AddAdminUser add an admin user on selected topic
func (t *TopicsController) AddAdminUser(ctx *gin.Context) {
	var paramJSON paramTopicUserJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckUser(ctx, &paramJSON)
	if e != nil {
		return
	}

	err := topic.AddAdminUser(utils.GetCtxUsername(ctx), paramJSON.Username, paramJSON.Recursive)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, "")
}

// RemoveRoUser removes a readonly user on selected topic
func (t *TopicsController) RemoveRoUser(ctx *gin.Context) {
	var paramJSON paramTopicUserJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckUser(ctx, &paramJSON)
	if e != nil {
		return
	}

	err := topic.RemoveRoUser(utils.GetCtxUsername(ctx), paramJSON.Username, paramJSON.Recursive)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, "")
}

// RemoveRwUser removes a read / write user on selected topic
func (t *TopicsController) RemoveRwUser(ctx *gin.Context) {
	var paramJSON paramTopicUserJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckUser(ctx, &paramJSON)
	if e != nil {
		return
	}

	err := topic.RemoveRwUser(utils.GetCtxUsername(ctx), paramJSON.Username, paramJSON.Recursive)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}
	ctx.JSON(http.StatusCreated, "")
}

// RemoveAdminUser removes an admin user on selected topic
func (t *TopicsController) RemoveAdminUser(ctx *gin.Context) {
	var paramJSON paramTopicUserJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckUser(ctx, &paramJSON)
	if e != nil {
		return
	}

	err := topic.RemoveAdminUser(utils.GetCtxUsername(ctx), paramJSON.Username, paramJSON.Recursive)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}
	ctx.JSON(http.StatusCreated, "")
}

type paramGroupJSON struct {
	Topic     string `json:"topic"`
	Groupname string `json:"groupname"`
	Recursive bool   `json:"recursive"`
}

// AddRoGroup add a readonly group on selected topic
func (t *TopicsController) AddRoGroup(ctx *gin.Context) {
	var paramJSON paramGroupJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckGroup(ctx, &paramJSON)
	if e != nil {
		return
	}
	err := topic.AddRoGroup(utils.GetCtxUsername(ctx), paramJSON.Groupname, paramJSON.Recursive)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, "")

}

// AddRwGroup add a read write group on selected topic
func (t *TopicsController) AddRwGroup(ctx *gin.Context) {
	var paramJSON paramGroupJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckGroup(ctx, &paramJSON)
	if e != nil {
		return
	}

	err := topic.AddRwGroup(utils.GetCtxUsername(ctx), paramJSON.Groupname, paramJSON.Recursive)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, "")
}

// AddAdminGroup add an admin group on selected topic
func (t *TopicsController) AddAdminGroup(ctx *gin.Context) {
	var paramJSON paramGroupJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckGroup(ctx, &paramJSON)
	if e != nil {
		return
	}

	err := topic.AddAdminGroup(utils.GetCtxUsername(ctx), paramJSON.Groupname, paramJSON.Recursive)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, "")
}

// AddParameter add a parameter on selected topic
func (t *TopicsController) AddParameter(ctx *gin.Context) {
	var topicParameterJSON topicParameterJSON
	ctx.Bind(&topicParameterJSON)
	topic, e := t.preCheckUserAdminOnTopic(ctx, topicParameterJSON.Topic)
	if e != nil {
		return
	}

	err := topic.AddParameter(utils.GetCtxUsername(ctx), topicParameterJSON.Key, topicParameterJSON.Value, topicParameterJSON.Recursive)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, "")
}

// RemoveParameter add a parameter on selected topic
func (t *TopicsController) RemoveParameter(ctx *gin.Context) {
	var topicParameterJSON topicParameterJSON
	ctx.Bind(&topicParameterJSON)

	topic, e := t.preCheckUserAdminOnTopic(ctx, topicParameterJSON.Topic)
	if e != nil {
		return
	}

	err := topic.RemoveParameter(utils.GetCtxUsername(ctx), topicParameterJSON.Key, topicParameterJSON.Value, topicParameterJSON.Recursive)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, "")
}

// RemoveRoGroup removes a read only group on selected topic
func (t *TopicsController) RemoveRoGroup(ctx *gin.Context) {
	var paramJSON paramGroupJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckGroup(ctx, &paramJSON)
	if e != nil {
		return
	}

	err := topic.RemoveRoGroup(utils.GetCtxUsername(ctx), paramJSON.Groupname, paramJSON.Recursive)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, "")
}

// RemoveRwGroup removes a read write group on selected topic
func (t *TopicsController) RemoveRwGroup(ctx *gin.Context) {
	var paramJSON paramGroupJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckGroup(ctx, &paramJSON)
	if e != nil {
		return
	}

	err := topic.RemoveRwGroup(utils.GetCtxUsername(ctx), paramJSON.Groupname, paramJSON.Recursive)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, "")
}

// RemoveAdminGroup removes an admin group on selected topic
func (t *TopicsController) RemoveAdminGroup(ctx *gin.Context) {
	var paramJSON paramGroupJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckGroup(ctx, &paramJSON)
	if e != nil {
		return
	}

	err := topic.RemoveAdminGroup(utils.GetCtxUsername(ctx), paramJSON.Groupname, paramJSON.Recursive)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, "")
}

type paramJSON struct {
	Topic           string `json:"topic"`
	MaxLength       int    `json:"maxlength"`
	CanForceDate    bool   `json:"canForceDate"`
	CanUpdateMsg    bool   `json:"canUpdateMsg"`
	CanDeleteMsg    bool   `json:"canDeleteMsg"`
	CanUpdateAllMsg bool   `json:"canUpdateAllMsg"`
	CanDeleteAllMsg bool   `json:"canDeleteAllMsg"`
	IsROPublic      bool   `json:"isROPublic"`
	Recursive       bool   `json:"recursive"`
}

// SetParam update Topic Parameters : MaxLength, CanForeceDate, CanUpdateMsg, CanDeleteMsg, CanUpdateAllMsg, CanDeleteAllMsg, IsROPublic
// admin only, except on Private topic
func (t *TopicsController) SetParam(ctx *gin.Context) {
	var paramJSON paramJSON
	ctx.Bind(&paramJSON)

	topic := models.Topic{}
	var err error
	if strings.HasPrefix(paramJSON.Topic, "/Private/"+utils.GetCtxUsername(ctx)) {
		err := topic.FindByTopic(paramJSON.Topic, false)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching topic /Private/" + utils.GetCtxUsername(ctx)})
			return
		}
	} else {
		topic, err = t.preCheckUserAdminOnTopic(ctx, paramJSON.Topic)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}
	}

	err = topic.SetParam(utils.GetCtxUsername(ctx),
		paramJSON.Recursive,
		paramJSON.MaxLength,
		paramJSON.CanForceDate,
		paramJSON.CanUpdateMsg,
		paramJSON.CanDeleteMsg,
		paramJSON.CanUpdateAllMsg,
		paramJSON.CanDeleteAllMsg,
		paramJSON.IsROPublic)

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}
	ctx.JSON(http.StatusCreated, "")
}
