package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/ovh/tat/models"
	"github.com/ovh/tat/utils"
)

// PresencesController contains all methods about presences manipulation
type PresencesController struct{}

func (*PresencesController) buildCriteria(ctx *gin.Context) *models.PresenceCriteria {
	c := models.PresenceCriteria{}
	skip, e := strconv.Atoi(ctx.DefaultQuery("skip", "0"))
	if e != nil {
		skip = 0
	}
	c.Skip = skip
	limit, e2 := strconv.Atoi(ctx.DefaultQuery("limit", "100"))
	if e2 != nil {
		limit = 10
	}
	c.Limit = limit
	c.IDPresence = ctx.Query("idPresence")
	c.Status = ctx.Query("status")
	c.Username = ctx.Query("username")
	c.DateMinPresence = ctx.Query("dateMinPresence")
	c.DateMaxPresence = ctx.Query("dateMaxPresence")
	return &c
}

// List list presences with given criteria
func (m *PresencesController) List(ctx *gin.Context) {
	topicIn, err := GetParam(ctx, "topic")
	if err != nil {
		return
	}
	criteria := m.buildCriteria(ctx)
	criteria.Topic = topicIn
	m.listWithCriteria(ctx, criteria)
}

func (m *PresencesController) listWithCriteria(ctx *gin.Context, criteria *models.PresenceCriteria) {
	user, e := m.preCheckUser(ctx)
	if e != nil {
		return
	}
	var topic = models.Topic{}
	if err := topic.FindByTopic(criteria.Topic, true, false, false, nil); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("topic "+criteria.Topic+" does not exist"))
		return
	}

	if isReadAccess := topic.IsUserReadAccess(user); !isReadAccess {
		ctx.AbortWithError(http.StatusForbidden, errors.New("No Read Access to this topic."))
		return
	}
	// add / if search on topic
	// as topic is in path, it can't start with a /
	if criteria.Topic != "" && string(criteria.Topic[0]) != "/" {
		criteria.Topic = "/" + criteria.Topic
	}

	topicDM := "/Private/" + utils.GetCtxUsername(ctx) + "/DM/"
	if strings.HasPrefix(criteria.Topic, topicDM) {
		part := strings.Split(criteria.Topic, "/")
		if len(part) != 5 {
			log.Errorf("wrong topic name for DM")
			ctx.AbortWithError(http.StatusInternalServerError, errors.New("Wrong topic name for DM:"+criteria.Topic))
			return
		}
		topicInverse := "/Private/" + part[4] + "/DM/" + utils.GetCtxUsername(ctx)
		criteria.Topic = criteria.Topic + "," + topicInverse
	}

	count, presences, err := models.ListPresences(criteria)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	out := &models.PresencesJSON{
		Count:     count,
		Presences: presences,
	}
	ctx.JSON(http.StatusOK, out)
}

func (m *PresencesController) preCheckTopic(ctx *gin.Context) (models.PresenceJSON, models.Topic, error) {
	var topic = models.Topic{}
	var presenceIn models.PresenceJSON
	ctx.Bind(&presenceIn)

	topicIn, err := GetParam(ctx, "topic")
	if err != nil {
		return presenceIn, topic, err
	}
	presenceIn.Topic = topicIn

	if err = topic.FindByTopic(presenceIn.Topic, true, false, false, nil); err != nil {
		e := errors.New("Topic " + presenceIn.Topic + " does not exist")
		ctx.AbortWithError(http.StatusInternalServerError, e)
		return presenceIn, topic, e
	}
	return presenceIn, topic, nil
}

func (*PresencesController) preCheckUser(ctx *gin.Context) (models.User, error) {
	var user = models.User{}
	if err := user.FindByUsername(utils.GetCtxUsername(ctx)); err != nil {
		e := errors.New("Error while fetching user.")
		ctx.AbortWithError(http.StatusInternalServerError, e)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": e.Error()})
		return user, e
	}
	return user, nil
}

func (m *PresencesController) create(ctx *gin.Context) {
	presenceIn, topic, e := m.preCheckTopic(ctx)
	if e != nil {
		return
	}

	user, e := m.preCheckUser(ctx)
	if e != nil {
		return
	}

	if isReadAccess := topic.IsUserReadAccess(user); !isReadAccess {
		e := errors.New("No Read Access to topic " + presenceIn.Topic + " for user " + user.Username)
		ctx.AbortWithError(http.StatusForbidden, e)
		ctx.JSON(http.StatusForbidden, gin.H{"error": e.Error()})
		return
	}

	var presence = models.Presence{}
	if err := presence.Upsert(user, topic, presenceIn.Status); err != nil {
		log.Errorf("Error while InsertPresence %s", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go models.WSPresence(&models.WSPresenceJSON{Action: "create", Presence: presence})

	//out := &presenceJSONOut{Presence: presence}
	//ctx.JSON(http.StatusCreated, nil)
}

// CreateAndGet creates a presence and get presences on current topic
func (m *PresencesController) CreateAndGet(ctx *gin.Context) {
	m.create(ctx)
	if ctx.IsAborted() {
		return
	}

	fiften := strconv.FormatInt(time.Now().Unix()-15, 10)

	topicIn, _ := GetParam(ctx, "topic") // no error possible here
	criteria := &models.PresenceCriteria{
		Skip:            0,
		Limit:           1000,
		Topic:           topicIn,
		DateMinPresence: fiften,
	}

	m.listWithCriteria(ctx, criteria)
}

// Delete deletes all presences of one user, on one topic
func (m *PresencesController) Delete(ctx *gin.Context) {
	presenceIn, topic, e := m.preCheckTopic(ctx)
	if e != nil {
		return
	}

	var user = models.User{}

	user, e = m.preCheckUser(ctx)
	if e != nil {
		return
	}

	if user.IsAdmin {
		if err := user.FindByUsername(presenceIn.Username); err != nil {
			e := errors.New("Error while fetching user " + presenceIn.Username + " for delete presence.")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": e.Error()})
			return
		}
	} else if isReadAccess := topic.IsUserReadAccess(user); !isReadAccess {
		e := errors.New("No Read Access to topic " + presenceIn.Topic + " for user " + user.Username)
		ctx.AbortWithError(http.StatusForbidden, e)
		ctx.JSON(http.StatusForbidden, gin.H{"error": e.Error()})
		return
	}

	var presence = models.Presence{}
	if err := presence.Delete(user, topic); err != nil {
		log.Errorf("Error while DeletePresence %s", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go models.WSPresence(&models.WSPresenceJSON{Action: "delete", Presence: presence})
	ctx.JSON(http.StatusOK, nil)
}

// CheckAllPresences checks presences, delete double
func (m *PresencesController) CheckAllPresences(ctx *gin.Context) {
	// admin check in route
	statsPresences, err := models.CheckAllPresences()
	if err != nil {
		log.Errorf("Error while get models.CheckAllPresences %s", err)
	}

	now := time.Now()
	ctx.JSON(http.StatusOK, gin.H{
		"date":           now.Unix(),
		"dateHuman":      now,
		"statsPresences": statsPresences,
	})
}
