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

func (*TopicsController) buildCriteria(ctx *gin.Context, user *models.User) *models.TopicCriteria {
	c := models.TopicCriteria{}
	skip, e := strconv.Atoi(ctx.DefaultQuery("skip", "0"))
	if e != nil {
		skip = 0
	}
	c.Skip = skip
	limit, e2 := strconv.Atoi(ctx.DefaultQuery("limit", "500"))
	if e2 != nil {
		limit = 500
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
	c.OnlyFavorites = ctx.Query("onlyFavorites")
	c.GetForTatAdmin = ctx.Query("getForTatAdmin")

	if c.OnlyFavorites == "true" {
		c.Topic = strings.Join(user.FavoritesTopics, ",")
	}
	return &c
}

// List returns the list of topics that can be viewed by user
func (t *TopicsController) List(ctx *gin.Context) {
	var user = &models.User{}
	if err := user.FindByUsername(utils.GetCtxUsername(ctx)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching user."})
		return
	}
	criteria := t.buildCriteria(ctx, user)
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while getting topic in param"})
		return
	}
	var user = models.User{}
	if err = user.FindByUsername(utils.GetCtxUsername(ctx)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching user."})
		return
	}
	topic := &models.Topic{}

	if errfind := topic.FindByTopic(topicRequest, user.IsAdmin, true, true, &user); errfind != nil {
		topic, _, err = checkDMTopic(ctx, topicRequest)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "topic " + topicRequest + " does not exist"})
			return
		}
	}

	if isReadAccess := topic.IsUserReadAccess(user); !isReadAccess {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "No Read Access to this topic: " + user.Username + " " + topic.Topic})
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	if err = topic.Delete(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, "")
}

// Truncate deletes all messages in a topic only if user is Tat admin, or admin on topic
func (t *TopicsController) Truncate(ctx *gin.Context) {
	var paramJSON paramTopicUserJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckUserAdminOnTopic(ctx, paramJSON.Topic)
	if e != nil {
		return
	}

	nbRemoved, err := topic.Truncate()
	if err != nil {
		log.Errorf("Error while truncate topic %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error while truncate topic " + topic.Topic})
		return
	}
	// 201 returns
	ctx.JSON(http.StatusCreated, gin.H{"info": fmt.Sprintf("%d messages removed", nbRemoved)})
}

// ComputeTags computes tags on one topic
func (t *TopicsController) ComputeTags(ctx *gin.Context) {
	var paramJSON paramTopicUserJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckUserAdminOnTopic(ctx, paramJSON.Topic)
	if e != nil {
		return
	}

	nbComputed, err := topic.ComputeTags()
	if err != nil {
		log.Errorf("Error while compute tags on topic %s: %s", topic.Topic, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error while compute tags on topic " + topic.Topic})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"info": fmt.Sprintf("%d tags computed", nbComputed)})
}

// ComputeLabels computes labels on one topic
func (t *TopicsController) ComputeLabels(ctx *gin.Context) {
	var paramJSON paramTopicUserJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckUserAdminOnTopic(ctx, paramJSON.Topic)
	if e != nil {
		return
	}

	nbComputed, err := topic.ComputeLabels()
	if err != nil {
		log.Errorf("Error while compute labels on topic %s: %s", topic.Topic, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error while compute labels on topic " + topic.Topic})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"info": fmt.Sprintf("%d labels computed", nbComputed)})
}

// TruncateTags clear tags on one topic
func (t *TopicsController) TruncateTags(ctx *gin.Context) {
	var paramJSON paramTopicUserJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckUserAdminOnTopic(ctx, paramJSON.Topic)
	if e != nil {
		return
	}

	if err := topic.TruncateTags(); err != nil {
		log.Errorf("Error while clear tags on topic %s: %s", topic.Topic, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error while clear tags on topic " + topic.Topic})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"info": fmt.Sprintf("%d tags cleared", len(topic.Tags))})
}

// TruncateLabels clear labels on one topic
func (t *TopicsController) TruncateLabels(ctx *gin.Context) {
	var paramJSON paramTopicUserJSON
	ctx.Bind(&paramJSON)
	topic, e := t.preCheckUserAdminOnTopic(ctx, paramJSON.Topic)
	if e != nil {
		return
	}

	if err := topic.TruncateLabels(); err != nil {
		log.Errorf("Error while clear labels on topic %s: %s", topic.Topic, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error while clear labels on topic " + topic.Topic})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"info": fmt.Sprintf("%d labels cleared", len(topic.Labels))})
}

// preCheckUser checks if user in paramJSON exists and if current user is admin on topic
func (t *TopicsController) preCheckUser(ctx *gin.Context, paramJSON *paramTopicUserJSON) (models.Topic, error) {
	if userExists := models.IsUsernameExists(paramJSON.Username); !userExists {
		e := errors.New("username " + paramJSON.Username + " does not exist")
		ctx.AbortWithError(http.StatusInternalServerError, e)
		return models.Topic{}, e
	}
	return t.preCheckUserAdminOnTopic(ctx, paramJSON.Topic)
}

// preCheckGroup checks if group exists and is admin on topic
func (t *TopicsController) preCheckGroup(ctx *gin.Context, paramJSON *paramGroupJSON) (models.Topic, error) {
	if groupExists := models.IsGroupnameExists(paramJSON.Groupname); !groupExists {
		e := errors.New("groupname" + paramJSON.Groupname + " does not exist")
		ctx.AbortWithError(http.StatusInternalServerError, e)
		return models.Topic{}, e
	}
	return t.preCheckUserAdminOnTopic(ctx, paramJSON.Topic)
}

func (t *TopicsController) preCheckUserAdminOnTopic(ctx *gin.Context, topicName string) (models.Topic, error) {
	topic := models.Topic{}
	if errfind := topic.FindByTopic(topicName, true, false, false, nil); errfind != nil {
		e := errors.New(errfind.Error())
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
		ctx.JSON(http.StatusForbidden, gin.H{"error": e})
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
		log.Errorf("Error while adding read only user: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		log.Errorf("Error while adding read write user: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	if err := topic.AddAdminUser(utils.GetCtxUsername(ctx), paramJSON.Username, paramJSON.Recursive); err != nil {
		log.Errorf("Error while adding admin user: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		log.Errorf("Error while removing read only user: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	if err := topic.RemoveRwUser(utils.GetCtxUsername(ctx), paramJSON.Username, paramJSON.Recursive); err != nil {
		log.Errorf("Error while removing read write user: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	if err := topic.RemoveAdminUser(utils.GetCtxUsername(ctx), paramJSON.Username, paramJSON.Recursive); err != nil {
		log.Errorf("Error while removing admin user: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	if err := topic.AddRoGroup(utils.GetCtxUsername(ctx), paramJSON.Groupname, paramJSON.Recursive); err != nil {
		log.Errorf("Error while adding admin read only group: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	if err := topic.AddRwGroup(utils.GetCtxUsername(ctx), paramJSON.Groupname, paramJSON.Recursive); err != nil {
		log.Errorf("Error while adding admin read write group: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	if err := topic.AddAdminGroup(utils.GetCtxUsername(ctx), paramJSON.Groupname, paramJSON.Recursive); err != nil {
		log.Errorf("Error while adding admin admin group: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		log.Errorf("Error while adding parameter: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		log.Errorf("Error while removing parameter: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		log.Errorf("Error while removing read only group: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		log.Errorf("Error while removing read write group: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		log.Errorf("Error while removing admin group: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, "")
}

type paramsJSON struct {
	Topic               string                  `json:"topic"`
	MaxLength           int                     `json:"maxlength"`
	CanForceDate        bool                    `json:"canForceDate"`
	CanUpdateMsg        bool                    `json:"canUpdateMsg"`
	CanDeleteMsg        bool                    `json:"canDeleteMsg"`
	CanUpdateAllMsg     bool                    `json:"canUpdateAllMsg"`
	CanDeleteAllMsg     bool                    `json:"canDeleteAllMsg"`
	IsROPublic          bool                    `json:"isROPublic"`
	IsAutoComputeTags   bool                    `json:"isAutoComputeTags"`
	IsAutoComputeLabels bool                    `json:"isAutoComputeLabels"`
	Recursive           bool                    `json:"recursive"`
	Parameters          []models.TopicParameter `json:"parameters"`
}

// SetParam update Topic Parameters : MaxLength, CanForeceDate, CanUpdateMsg, CanDeleteMsg, CanUpdateAllMsg, CanDeleteAllMsg, IsROPublic
// admin only, except on Private topic
func (t *TopicsController) SetParam(ctx *gin.Context) {
	var paramsJSON paramsJSON
	ctx.Bind(&paramsJSON)

	topic := models.Topic{}
	var err error
	if strings.HasPrefix(paramsJSON.Topic, "/Private/"+utils.GetCtxUsername(ctx)) {
		err := topic.FindByTopic(paramsJSON.Topic, false, false, false, nil)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching topic /Private/" + utils.GetCtxUsername(ctx)})
			return
		}
	} else {
		topic, err = t.preCheckUserAdminOnTopic(ctx, paramsJSON.Topic)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}
	}

	err = topic.SetParam(utils.GetCtxUsername(ctx),
		paramsJSON.Recursive,
		paramsJSON.MaxLength,
		paramsJSON.CanForceDate,
		paramsJSON.CanUpdateMsg,
		paramsJSON.CanDeleteMsg,
		paramsJSON.CanUpdateAllMsg,
		paramsJSON.CanDeleteAllMsg,
		paramsJSON.IsROPublic,
		paramsJSON.IsAutoComputeTags,
		paramsJSON.IsAutoComputeLabels,
		paramsJSON.Parameters)

	if err != nil {
		log.Errorf("Error while setting parameters: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"info": fmt.Sprintf("Topic %s updated", topic.Topic)})
}

// AllComputeTags Compute tags on all topics
func (t *TopicsController) AllComputeTags(ctx *gin.Context) {
	// It's only for admin, admin already checked in route
	info, err := models.AllTopicsComputeTags()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"info": info})
}

// AllComputeLabels Compute tags on all topics
func (t *TopicsController) AllComputeLabels(ctx *gin.Context) {
	// It's only for admin, admin already checked in route
	info, err := models.AllTopicsComputeLabels()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"info": info})
}

type attributeJSON struct {
	ParamName  string `json:"paramName"`
	ParamValue string `json:"paramValue"`
}

// AllSetParam set a param on all topics
func (t *TopicsController) AllSetParam(ctx *gin.Context) {
	// It's only for admin, admin already checked in route
	var param attributeJSON
	ctx.Bind(&param)

	info, err := models.AllTopicsSetParam(param.ParamName, param.ParamValue)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"info": info})
}

// AllComputeReplies computes replies on all topics
func (t *TopicsController) AllComputeReplies(ctx *gin.Context) {
	// It's only for admin, admin already checked in route
	var param attributeJSON
	ctx.Bind(&param)

	info, err := models.AllTopicsComputeReplies()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"info": info})
}
