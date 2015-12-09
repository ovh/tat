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

// MessagesController contains all methods about messages manipulation
type MessagesController struct{}

type messagesJSON struct {
	Messages  []models.Message `json:"messages"`
	IsTopicRw bool             `json:"isTopicRw"`
}

type messageJSONOut struct {
	Message models.Message `json:"message"`
	Info    string         `json:"info"`
}

type messageJSON struct {
	ID           string `json:"_id"`
	Text         string `json:"text"`
	Option       string `json:"option"`
	Topic        string
	IDReference  string         `json:"idReference"`
	Action       string         `json:"action"`
	DateCreation int64          `json:"dateCreation"`
	Labels       []models.Label `json:"labels"`
}

func (*MessagesController) buildCriteria(ctx *gin.Context) *models.MessageCriteria {
	c := models.MessageCriteria{}
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
	c.TreeView = ctx.Query("treeView")
	c.IDMessage = ctx.Query("idMessage")
	c.InReplyOfID = ctx.Query("inReplyOfID")
	c.InReplyOfIDRoot = ctx.Query("inReplyOfIDRoot")
	c.AllIDMessage = ctx.Query("allIDMessage")
	c.Text = ctx.Query("text")
	c.Label = ctx.Query("label")
	c.NotLabel = ctx.Query("notLabel")
	c.AndLabel = ctx.Query("andLabel")
	c.Tag = ctx.Query("tag")
	c.NotTag = ctx.Query("notTag")
	c.AndTag = ctx.Query("andTag")
	c.DateMinCreation = ctx.Query("dateMinCreation")
	c.DateMaxCreation = ctx.Query("dateMaxCreation")
	c.DateMinUpdate = ctx.Query("dateMinUpdate")
	c.DateMaxUpdate = ctx.Query("dateMaxUpdate")
	c.Username = ctx.Query("username")
	c.LimitMinNbReplies = ctx.Query("limitMinNbReplies")
	c.LimitMaxNbReplies = ctx.Query("limitMaxNbReplies")
	c.OnlyMsgRoot = ctx.Query("onlyMsgRoot")
	return &c
}

// List messages on one topic, with given criterias
func (m *MessagesController) List(ctx *gin.Context) {
	var criteria = m.buildCriteria(ctx)
	presenceArg := ctx.Query("presence")
	topicIn, err := GetParam(ctx, "topic")
	if err != nil {
		return
	}
	criteria.Topic = topicIn

	// add / if search on topic
	// as topic is in path, it can't start with a /
	if criteria.Topic != "" && string(criteria.Topic[0]) != "/" {
		criteria.Topic = "/" + criteria.Topic
	}

	var topic = models.Topic{}
	err = topic.FindByTopic(criteria.Topic, true)
	if err != nil {
		topicCriteria := ""
		_, topicCriteria, err = m.checkDMTopic(ctx, criteria.Topic)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "topic " + criteria.Topic + " does not exist"})
			return
		}
		// hack to get new created DM Topic
		err := topic.FindByTopic(criteria.Topic, true)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "topic " + criteria.Topic + " does not exist (2)"})
			return
		}
		criteria.Topic = topicCriteria
	}

	out := &messagesJSON{}

	var user models.User
	var e error
	if utils.GetCtxUsername(ctx) != "" {
		user, e = PreCheckUser(ctx)
		if e != nil {
			return
		}
		isReadAccess := topic.IsUserReadAccess(user)
		if !isReadAccess {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "No Read Access to this topic"})
			return
		}
		out.IsTopicRw = topic.IsUserRW(&user)
	} else if !topic.IsROPublic {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "No Public Read Access Public to this topic"})
		return
	} else if topic.IsROPublic && strings.HasPrefix(topic.Topic, "/Private") {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "No Public Read Access to this topic"})
		return
	}

	// send presence
	if presenceArg != "" && !user.IsSystem {
		go func() {
			var presence = models.Presence{}
			err := presence.Upsert(user, topic, presenceArg)
			if err != nil {
				log.Errorf("Error while InsertPresence %s", err)
			}
			go models.WSPresence(&models.WSPresenceJSON{Action: "create", Presence: presence})
		}()

	}

	messages, err := models.ListMessages(criteria)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out.Messages = messages
	ctx.JSON(http.StatusOK, out)
}

func (m *MessagesController) preCheckTopic(ctx *gin.Context) (messageJSON, models.Message, models.Topic, error) {
	var topic = models.Topic{}
	var message = models.Message{}
	var messageIn messageJSON
	ctx.Bind(&messageIn)

	topicIn, err := GetParam(ctx, "topic")
	if err != nil {
		return messageIn, message, topic, err
	}
	messageIn.Topic = topicIn

	if messageIn.IDReference == "" || messageIn.Action == "" {
		err := topic.FindByTopic(messageIn.Topic, true)
		if err != nil {
			topic, _, err = m.checkDMTopic(ctx, messageIn.Topic)
			if err != nil {
				e := errors.New("Topic " + messageIn.Topic + " does not exist")
				ctx.JSON(http.StatusNotFound, gin.H{"error": e.Error()})
				return messageIn, message, topic, e
			}
		}
	} else if messageIn.IDReference != "" {
		err := message.FindByID(messageIn.IDReference)
		if err != nil {
			e := errors.New("Message " + messageIn.IDReference + " does not exist")
			ctx.JSON(http.StatusNotFound, gin.H{"error": e.Error()})
			return messageIn, message, topic, e
		}

		topicName := ""
		if messageIn.Action == "update" {
			topicName = messageIn.Topic
		} else if messageIn.Action == "reply" || messageIn.Action == "unbookmark" ||
			messageIn.Action == "like" || messageIn.Action == "unlike" ||
			messageIn.Action == "label" || messageIn.Action == "unlabel" ||
			messageIn.Action == "tag" || messageIn.Action == "untag" {
			topicName = m.inverseIfDMTopic(ctx, message.Topics[0])
		} else if messageIn.Action == "move" {
			topicName = topicIn
		} else if messageIn.Action == "task" || messageIn.Action == "untask" {
			topicName, err = m.getTopicNonPrivateTasks(ctx, message.Topics)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return messageIn, message, topic, err
			}
		} else if messageIn.Action == "bookmark" {
			topicAction := m.getTopicNameFromAction(utils.GetCtxUsername(ctx), messageIn.Action)
			if !strings.HasPrefix(messageIn.Topic, topicAction) {
				e := fmt.Errorf("Invalid Topic name for action %s mTopic %s topicAction:%s ", messageIn.Action, messageIn.Topic, topicAction)
				ctx.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
				return messageIn, message, topic, e
			}
			topicName = messageIn.Topic
		} else {
			e := errors.New("Invalid Call. IDReference not empty with unknown action")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
			return messageIn, message, topic, e
		}
		err = topic.FindByTopic(topicName, true)
		if err != nil {
			e := errors.New("Topic " + topicName + " does not exist")
			ctx.JSON(http.StatusNotFound, gin.H{"error": e.Error()})
			return messageIn, message, topic, e
		}
	} else {
		e := errors.New("Topic and IDReference are null. Wrong request")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return messageIn, message, topic, e
	}
	return messageIn, message, topic, nil
}

// Create a new message on one topic
func (m *MessagesController) Create(ctx *gin.Context) {
	messageIn, messageReference, topic, e := m.preCheckTopic(ctx)
	if e != nil {
		return
	}

	user, e := PreCheckUser(ctx)
	if e != nil {
		return
	}

	isRw := topic.IsUserRW(&user)
	if !isRw {
		ctx.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("No RW Access to topic " + messageIn.Topic)})
		return
	}

	var message = models.Message{}

	info := ""
	if messageIn.Action == "bookmark" {
		var originalUser = models.User{}
		err := originalUser.FindByUsername(utils.GetCtxUsername(ctx))
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, errors.New("Error while fetching original user."))
			return
		}
		err = message.Insert(originalUser, topic, messageReference.Text, "", -1, messageReference.Labels, false)
		if err != nil {
			log.Errorf("Error while InsertMessage with action %s : %s", messageIn.Action, err)
			ctx.AbortWithError(http.StatusInternalServerError, errors.New(err.Error()))
			return
		}
		info = fmt.Sprintf("New Bookmark created in %s", topic.Topic)
	} else {
		err := message.Insert(user, topic, messageIn.Text, messageIn.IDReference, messageIn.DateCreation, messageIn.Labels, false)
		if err != nil {
			log.Errorf("%s", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		go models.WSMessageNew(&models.WSMessageNewJSON{Topic: topic.Topic})
		info = fmt.Sprintf("Message created in %s", topic.Topic)
	}
	out := &messageJSONOut{Message: message, Info: info}
	go models.WSMessage(&models.WSMessageJSON{Action: "create", Username: user.Username, Message: message})
	ctx.JSON(http.StatusCreated, out)
}

// Update a message : like, unlike, add label, etc...
func (m *MessagesController) Update(ctx *gin.Context) {
	messageIn, messageReference, topic, e := m.preCheckTopic(ctx)
	if e != nil {
		return
	}

	user, e := PreCheckUser(ctx)
	if e != nil {
		return
	}

	if messageIn.Action == "like" || messageIn.Action == "unlike" {
		m.likeOrUnlike(ctx, messageIn.Action, messageReference, topic, user)
		return
	}

	isRw := topic.IsUserRW(&user)
	if !isRw {
		ctx.AbortWithError(http.StatusForbidden, errors.New("No RW Access to topic : "+messageIn.Topic))
		return
	}

	if messageIn.Action == "label" || messageIn.Action == "unlabel" {
		m.addOrRemoveLabel(ctx, &messageIn, messageReference, user)
		return
	}

	if messageIn.Action == "tag" || messageIn.Action == "untag" {
		m.addOrRemoveTag(ctx, &messageIn, messageReference, user)
		return
	}

	if messageIn.Action == "task" || messageIn.Action == "untask" {
		m.addOrRemoveTask(ctx, &messageIn, messageReference, user, topic)
		return
	}

	if messageIn.Action == "update" {
		m.updateMessage(ctx, &messageIn, messageReference, user, topic)
		return
	}

	if messageIn.Action == "move" {
		m.moveMessage(ctx, &messageIn, messageReference, user, topic)
		return
	}

	ctx.JSON(http.StatusBadRequest, gin.H{"error": "Action invalid."})
}

// Delete a message
func (m *MessagesController) Delete(ctx *gin.Context) {
	m.messageDelete(ctx, false)
}

// DeleteCascade deletes a message and its replies
func (m *MessagesController) DeleteCascade(ctx *gin.Context) {
	m.messageDelete(ctx, true)
}

func (m *MessagesController) messageDelete(ctx *gin.Context, cascade bool) {
	idMessageIn, err := GetParam(ctx, "idMessage")
	if err != nil {
		return
	}

	message := models.Message{}
	err = message.FindByID(idMessageIn)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Message %s does not exist", idMessageIn)})
		return
	}

	user, e := PreCheckUser(ctx)
	if e != nil {
		return
	}

	topic, err := m.checkBeforeDelete(ctx, message, user)
	if err != nil {
		// ctx writes in checkBeforeDelete
		return
	}

	c := &models.MessageCriteria{
		InReplyOfID: message.ID,
		TreeView:    "onetree",
	}

	msgs, err := models.ListMessages(c)
	if err != nil {
		log.Errorf("Error while list Messages in Delete %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while list Messages in Delete"})
		return
	}

	if cascade {
		for _, r := range msgs {
			_, err := m.checkBeforeDelete(ctx, r, user)
			if err != nil {
				// ctx writes in checkBeforeDelete
				return
			}
		}
	} else if len(msgs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Could not delete this message, this message have replies"})
		return
	}

	err = message.Delete(cascade)
	if err != nil {
		log.Errorf("Error while delete a message %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	go models.WSMessage(&models.WSMessageJSON{Action: "delete", Username: user.Username, Message: message})
	ctx.JSON(http.StatusOK, gin.H{"info": fmt.Sprintf("Message deleted from %s", topic.Topic)})
}

// checkBeforeDelete checks
// - if user is RW on topic
// - if topic is Private OR is CanDeleteMsg or CanDeleteAllMsg
func (m *MessagesController) checkBeforeDelete(ctx *gin.Context, message models.Message, user models.User) (models.Topic, error) {
	topic := models.Topic{}
	err := topic.FindByTopic(message.Topics[0], true)
	if err != nil {
		e := fmt.Sprintf("Topic %s does not exist", message.Topics[0])
		ctx.JSON(http.StatusNotFound, gin.H{"error": e})
		return topic, fmt.Errorf(e)
	}
	isRw := topic.IsUserRW(&user)
	if !isRw {
		e := fmt.Sprintf("No RW Access to topic %s", message.Topics[0])
		ctx.JSON(http.StatusForbidden, gin.H{"error": e})
		return topic, fmt.Errorf(e)
	}

	if !strings.HasPrefix(message.Topics[0], "/Private/"+user.Username) && !topic.CanDeleteMsg && !topic.CanDeleteAllMsg {
		if !topic.CanDeleteMsg && !topic.CanDeleteAllMsg {
			e := fmt.Sprintf("You can't delete a message from topic %s", topic.Topic)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": e})
			return topic, fmt.Errorf(e)
		}
		e := fmt.Sprintf("Could not delete a message in a non private topic %s", message.Topics[0])
		ctx.JSON(http.StatusBadRequest, gin.H{"error": e})
		return topic, fmt.Errorf(e)
	}

	if !topic.CanDeleteAllMsg && message.Author.Username != user.Username {
		e := fmt.Sprintf("Could not delete a message from another user %s than you %s", message.Author.Username, user.Username)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": e})
		return topic, fmt.Errorf(e)
	}

	for _, topicName := range message.Topics {
		// if msg is only in tasks topic, ok to delete it
		if strings.HasPrefix(topicName, "/Private/") && strings.HasSuffix(topicName, "/Tasks") && len(message.Topics) > 1 {
			// if label done on msg, can delete it
			if !message.ContainsLabel("done") {
				e := fmt.Sprintf("Could not delete a message in a tasks topic")
				ctx.JSON(http.StatusBadRequest, gin.H{"error": e})
				return topic, fmt.Errorf(e)
			}
		}
	}
	return topic, nil
}

func (m *MessagesController) likeOrUnlike(ctx *gin.Context, action string, message models.Message, topic models.Topic, user models.User) {
	isReadAccess := topic.IsUserReadAccess(user)
	if !isReadAccess {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("No Read Access to topic "+message.Topics[0]))
		return
	}

	info := ""
	if action == "like" {
		err := message.Like(user)
		if err != nil {
			log.Errorf("Error while like a message %s", err)
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		info = "like added"
	} else if action == "unlike" {
		err := message.Unlike(user)
		if err != nil {
			log.Errorf("Error while like a message %s", err)
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		info = "like removed"
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid action : " + action)})
		return
	}
	go models.WSMessage(&models.WSMessageJSON{Action: action, Username: user.Username, Message: message})
	ctx.JSON(http.StatusCreated, gin.H{"info": info})
}

func (m *MessagesController) addOrRemoveLabel(ctx *gin.Context, messageIn *messageJSON, message models.Message, user models.User) {
	if messageIn.Text == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("Invalid Text for label"))
		return
	}
	info := gin.H{}
	if messageIn.Action == "label" {
		addedLabel, err := message.AddLabel(messageIn.Text, messageIn.Option)
		if err != nil {
			log.Errorf("Error while adding a label to a message %s", err)
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		info = gin.H{"info": fmt.Sprintf("label %s added to message", addedLabel.Text), "label": addedLabel, "message": message}
	} else if messageIn.Action == "unlabel" {
		err := message.RemoveLabel(messageIn.Text)
		if err != nil {
			log.Errorf("Error while remove a label from a message %s", err)
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		info = gin.H{"info": fmt.Sprintf("label %s removed from message", messageIn.Text), "message": message}
	} else {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("Invalid action : "+messageIn.Action))
		return
	}
	go models.WSMessage(&models.WSMessageJSON{Action: messageIn.Action, Username: user.Username, Message: message})
	ctx.JSON(http.StatusCreated, info)
}

func (m *MessagesController) addOrRemoveTag(ctx *gin.Context, messageIn *messageJSON, message models.Message, user models.User) {

	if !user.IsSystem {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Invalid Action for non-system user"})
		return
	}

	if messageIn.Text == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("Invalid Text for tag"))
		return
	}

	if messageIn.Action == "tag" {
		err := message.AddTag(messageIn.Text)
		if err != nil {
			log.Errorf("Error while adding a tag to a message %s", err)
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	} else if messageIn.Action == "untag" {
		err := message.RemoveTag(messageIn.Text)
		if err != nil {
			log.Errorf("Error while remove a tag from a message %s", err)
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	} else {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("Invalid action : "+messageIn.Action))
		return
	}
	go models.WSMessage(&models.WSMessageJSON{Action: messageIn.Action, Username: user.Username, Message: message})
	ctx.JSON(http.StatusCreated, "")
}

func (m *MessagesController) addOrRemoveTask(ctx *gin.Context, messageIn *messageJSON, message models.Message, user models.User, topic models.Topic) {
	info := ""
	if messageIn.Action == "task" {
		err := message.AddToTasks(user, topic)
		if err != nil {
			log.Errorf("Error while adding a message to tasks %s", err)
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		info = fmt.Sprintf("New Task created in %s", models.GetPrivateTopicTaskName(user))
	} else if messageIn.Action == "untask" {
		err := message.RemoveFromTasks(user, topic)
		if err != nil {
			log.Errorf("Error while remove a message from tasks %s", err)
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		info = fmt.Sprintf("Task removed from %s", models.GetPrivateTopicTaskName(user))
	} else {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("Invalid action : "+messageIn.Action))
		return
	}
	go models.WSMessage(&models.WSMessageJSON{Action: messageIn.Action, Username: user.Username, Message: message})
	ctx.JSON(http.StatusCreated, gin.H{"info": info})
}

func (m *MessagesController) updateMessage(ctx *gin.Context, messageIn *messageJSON, message models.Message, user models.User, topic models.Topic) {
	info := ""
	if messageIn.Action == "update" {

		if !topic.CanUpdateMsg && !topic.CanUpdateAllMsg {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("You can't update a message on topic %s", topic.Topic)})
			return
		}

		if !topic.CanUpdateAllMsg && message.Author.Username != user.Username {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Could not update a message from another user %s than you %s", message.Author.Username, user.Username)})
			return
		}

		message.Text = messageIn.Text
		err := message.Update(user, topic)
		if err != nil {
			log.Errorf("Error while update a message %s", err)
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		info = fmt.Sprintf("Message updated in %s", topic.Topic)
	} else {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("Invalid action : "+messageIn.Action))
		return
	}
	go models.WSMessage(&models.WSMessageJSON{Action: messageIn.Action, Username: user.Username, Message: message})
	out := &messageJSONOut{Message: message, Info: info}
	ctx.JSON(http.StatusOK, out)
}

func (m *MessagesController) moveMessage(ctx *gin.Context, messageIn *messageJSON, message models.Message, user models.User, topic models.Topic) {
	// Check if user can delete msg on from topic
	_, err := m.checkBeforeDelete(ctx, message, user)
	if err != nil {
		// ctx writes in checkBeforeDelete
		return
	}

	// Check if user can write msg from dest topic
	isRw := topic.IsUserRW(&user)
	if !isRw {
		ctx.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("No RW Access to topic %s", topic.Topic)})
		return
	}

	// check if message is a reply -> not possible
	if message.InReplyOfIDRoot != "" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("You can't move a reply message")})
		return
	}

	info := ""
	if messageIn.Action == "move" {
		err := message.Move(user, topic)
		if err != nil {
			log.Errorf("Error while move a message to topic: %s err: %s", topic.Topic, err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while move a message to topic %s", topic.Topic)})
			return
		}
		info = fmt.Sprintf("Message move to %s", topic.Topic)
	} else {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("Invalid action : "+messageIn.Action))
		return
	}
	go models.WSMessage(&models.WSMessageJSON{Action: messageIn.Action, Username: user.Username, Message: message})
	ctx.JSON(http.StatusCreated, gin.H{"info": info})
}

func (m *MessagesController) getTopicNameFromAction(username, action string) string {
	return "/Private/" + username + "/" + strings.Title(action) + "s"
}

func (m *MessagesController) inverseIfDMTopic(ctx *gin.Context, topicName string) string {
	if !strings.HasPrefix(topicName, "/Private/") {
		return topicName
	}
	if !strings.HasSuffix(topicName, "/DM/"+utils.GetCtxUsername(ctx)) {
		return topicName
	}

	// /Private/usernameFrom/DM/usernameTO
	part := strings.Split(topicName, "/")
	if len(part) != 5 {
		return topicName
	}
	return "/Private/" + utils.GetCtxUsername(ctx) + "/DM/" + part[2]
}

func (m *MessagesController) getTopicNonPrivateTasks(ctx *gin.Context, topics []string) (string, error) {
	// if msg is only in topic Tasks
	topicTasks := "/Private/" + utils.GetCtxUsername(ctx) + "/Tasks"
	for _, name := range topics {
		if !strings.HasPrefix(name, "/Private") {
			return name, nil
		}
		if !strings.HasPrefix(topics[0], topicTasks) {
			return name, nil
		}
	}
	return "", errors.New("Could not get non private task topic")
}

func (m *MessagesController) checkDMTopic(ctx *gin.Context, topicName string) (models.Topic, string, error) {
	var topic = models.Topic{}

	topicParentName := "/Private/" + utils.GetCtxUsername(ctx) + "/DM"
	if !strings.HasPrefix(topicName, topicParentName+"/") {
		log.Errorf("wrong topic name for DM:" + topicName)
		return topic, "", errors.New("Wrong tpic name for DM:" + topicName)
	}

	// /Private/usernameFrom/DM/usernameTO
	part := strings.Split(topicName, "/")
	if len(part) != 5 {
		log.Errorf("wrong topic name for DM")
		return topic, "", errors.New("Wrong topic name for DM:" + topicName)
	}

	var userFrom = models.User{}
	err := userFrom.FindByUsername(utils.GetCtxUsername(ctx))
	if err != nil {
		return topic, "", errors.New("Error while fetching user.")
	}
	var userTo = models.User{}
	usernameTo := part[4]
	err = userTo.FindByUsername(usernameTo)
	if err != nil {
		return topic, "", errors.New("Error while fetching user.")
	}

	err = m.checkTopicParentDM(userFrom)
	if err != nil {
		return topic, "", errors.New(err.Error())
	}

	err = m.checkTopicParentDM(userTo)
	if err != nil {
		return topic, "", errors.New(err.Error())
	}

	topic, err = m.insertTopicDM(userFrom, userTo)
	if err != nil {
		return topic, "", errors.New(err.Error())
	}

	_, err = m.insertTopicDM(userTo, userFrom)
	if err != nil {
		return topic, "", errors.New(err.Error())
	}

	topicCriteria := topicName + "," + "/Private/" + usernameTo + "/DM/" + userFrom.Username
	return topic, topicCriteria, nil
}

func (*MessagesController) insertTopicDM(userFrom, userTo models.User) (models.Topic, error) {
	var topic = models.Topic{}
	topicName := "/Private/" + userFrom.Username + "/DM/" + userTo.Username
	topic.Topic = topicName
	topic.Description = userTo.Fullname
	err := topic.Insert(&userFrom)
	if err != nil {
		log.Errorf("Error while InsertTopic %s", err)
		return topic, err
	}
	return topic, nil
}

func (*MessagesController) checkTopicParentDM(user models.User) error {
	topicName := "/Private/" + user.Username + "/DM"
	var topicParent = models.Topic{}
	err := topicParent.FindByTopic(topicName, false)
	if err != nil {
		topicParent.Topic = topicName
		topicParent.Description = "DM Topics"
		err = topicParent.Insert(&user)
		if err != nil {
			log.Errorf("Error while InsertTopic Parent %s", err)
			return err
		}
	}
	return nil
}
