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
	c.LimitMinNbVotesUP = ctx.Query("limitMinNbVotesUP")
	c.LimitMaxNbVotesUP = ctx.Query("limitMaxNbVotesUP")
	c.LimitMinNbVotesDown = ctx.Query("limitMinNbVotesDown")
	c.LimitMaxNbVotesDown = ctx.Query("limitMaxNbVotesDown")
	c.Username = ctx.Query("username")
	c.LimitMinNbReplies = ctx.Query("limitMinNbReplies")
	c.LimitMaxNbReplies = ctx.Query("limitMaxNbReplies")
	c.OnlyMsgRoot = ctx.Query("onlyMsgRoot")
	c.OnlyCount = ctx.Query("onlyCount")
	return &c
}

// List messages on one topic, with given criteria
func (m *MessagesController) List(ctx *gin.Context) {
	var criteria = m.buildCriteria(ctx)

	// we can't use NotLabel or NotTag with fulltree or onetree
	// this avoid potential wrong results associated with a short param limit
	if (criteria.NotLabel != "" || criteria.NotTag != "") &&
		(criteria.TreeView == "fulltree" || criteria.TreeView == "onetree") && criteria.OnlyMsgRoot == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "You can't use fulltree or onetree with NotLabel or NotTag"})
		return
	}

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
	if topic.FindByTopic(criteria.Topic, true, false, false, nil); err != nil {
		topicCriteria := ""
		_, topicCriteria, err = checkDMTopic(ctx, criteria.Topic)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "topic " + criteria.Topic + " does not exist"})
			return
		}
		// hack to get new created DM Topic
		if e := topic.FindByTopic(criteria.Topic, true, false, false, nil); e != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "topic " + criteria.Topic + " does not exist (2)"})
			return
		}
		criteria.Topic = topicCriteria
	}

	if topic.ID == "" {
		log.Errorf("Invalid Topic ID. user: %s topic:%s", user.Username, criteria.Topic)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "topic " + criteria.Topic + " does not exist"})
		return
	}

	out := &models.MessagesJSON{}

	var user models.User
	var e error
	if utils.GetCtxUsername(ctx) != "" {
		user, e = PreCheckUser(ctx)
		if e != nil {
			return
		}
		if isReadAccess := topic.IsUserReadAccess(user); !isReadAccess {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "No Read Access on topic"})
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

	if criteria.OnlyCount == "true" {
		count, e := models.CountMessages(criteria)
		if e != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": e.Error()})
			return
		}
		ctx.JSON(http.StatusOK, &models.MessagesCountJSON{Count: count})
		return
	}

	// send presence
	if presenceArg != "" && !user.IsSystem {
		go func() {
			var presence = models.Presence{}
			if e := presence.Upsert(user, topic, presenceArg); e != nil {
				log.Errorf("Error while InsertPresence %s", e)
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

func (m *MessagesController) preCheckTopic(ctx *gin.Context) (models.MessageJSON, models.Message, models.Topic, error) {
	var topic = models.Topic{}
	var message = models.Message{}
	var messageIn models.MessageJSON
	ctx.Bind(&messageIn)

	topicIn, err := GetParam(ctx, "topic")
	if err != nil {
		return messageIn, message, topic, err
	}
	messageIn.Topic = topicIn

	if messageIn.IDReference == "" || messageIn.Action == "" {
		if efind := topic.FindByTopic(messageIn.Topic, true, true, true, nil); efind != nil {
			topica, _, edm := checkDMTopic(ctx, messageIn.Topic)
			if edm != nil {
				e := errors.New("Topic " + messageIn.Topic + " does not exist")
				ctx.JSON(http.StatusNotFound, gin.H{"error": e.Error()})
				return messageIn, message, topic, e
			}
			topic = *topica
		}
	} else if messageIn.IDReference != "" {
		if efind := message.FindByID(messageIn.IDReference); efind != nil {
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
			messageIn.Action == "voteup" || messageIn.Action == "votedown" ||
			messageIn.Action == "unvoteup" || messageIn.Action == "unvotedown" ||
			messageIn.Action == "relabel" || messageIn.Action == "concat" {
			topicName = m.inverseIfDMTopic(ctx, message.Topics[0])
		} else if messageIn.Action == "move" {
			topicName = topicIn
		} else if messageIn.Action == "task" || messageIn.Action == "untask" {
			topicName, err = m.getTopicNonPrivateTasks(ctx, message.Topics)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return messageIn, message, topic, err
			}
			topicName = m.inverseIfDMTopic(ctx, topicName)
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
		if err = topic.FindByTopic(topicName, true, true, true, nil); err != nil {
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

	if isRw := topic.IsUserRW(&user); !isRw {
		ctx.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("No RW Access to topic " + messageIn.Topic)})
		return
	}

	var message = models.Message{}

	info := ""
	if messageIn.Action == "bookmark" {
		var originalUser = models.User{}
		if err := originalUser.FindByUsername(utils.GetCtxUsername(ctx)); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, errors.New("Error while fetching original user."))
			return
		}
		err := message.Insert(originalUser, topic, messageReference.Text, "", -1, messageReference.Labels, false)
		if err != nil {
			log.Errorf("Error while InsertMessage with action %s : %s", messageIn.Action, err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		info = fmt.Sprintf("New Bookmark created in %s", topic.Topic)
	} else {
		// New root message or reply
		err := message.Insert(user, topic, messageIn.Text, messageIn.IDReference, messageIn.DateCreation, messageIn.Labels, false)
		if err != nil {
			log.Errorf("%s", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		go models.WSMessageNew(&models.WSMessageNewJSON{Topic: topic.Topic})
		info = fmt.Sprintf("Message created in %s", topic.Topic)
	}
	out := &models.MessageJSONOut{Message: message, Info: info}
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

	if !topic.IsUserRW(&user) {
		ctx.AbortWithError(http.StatusForbidden, errors.New("No RW Access to topic : "+messageIn.Topic))
		return
	}

	if messageIn.Action == "label" || messageIn.Action == "unlabel" || messageIn.Action == "relabel" {
		m.addOrRemoveLabel(ctx, &messageIn, messageReference, user, topic)
		return
	}

	if messageIn.Action == "voteup" || messageIn.Action == "votedown" ||
		messageIn.Action == "unvoteup" || messageIn.Action == "unvotedown" {
		m.voteMessage(ctx, &messageIn, messageReference, user, topic)
		return
	}

	if messageIn.Action == "task" || messageIn.Action == "untask" {
		m.addOrRemoveTask(ctx, &messageIn, messageReference, user, topic)
		return
	}

	if messageIn.Action == "update" || messageIn.Action == "concat" {
		m.updateMessage(ctx, &messageIn, messageReference, user, topic)
		return
	}

	if messageIn.Action == "move" {
		m.moveMessage(ctx, &messageIn, messageReference, user, topic)
		return
	}

	ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Action"})
}

// Delete a message
func (m *MessagesController) Delete(ctx *gin.Context) {
	m.messageDelete(ctx, false, false)
}

// DeleteCascade deletes a message and its replies
func (m *MessagesController) DeleteCascade(ctx *gin.Context) {
	m.messageDelete(ctx, true, false)
}

// DeleteCascadeForce deletes a message and its replies, event if a msg is in a
// tasks topic of one user
func (m *MessagesController) DeleteCascadeForce(ctx *gin.Context) {
	m.messageDelete(ctx, true, true)
}

func (m *MessagesController) messageDelete(ctx *gin.Context, cascade, force bool) {
	idMessageIn, err := GetParam(ctx, "idMessage")
	if err != nil {
		return
	}

	message := models.Message{}
	if err = message.FindByID(idMessageIn); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Message %s does not exist", idMessageIn)})
		return
	}

	user, e := PreCheckUser(ctx)
	if e != nil {
		return
	}

	topic, err := m.checkBeforeDelete(ctx, message, user, force)
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
			_, errCheck := m.checkBeforeDelete(ctx, r, user, force)
			if errCheck != nil {
				// ctx writes in checkBeforeDelete
				return
			}
		}
	} else if len(msgs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Could not delete this message, this message have replies"})
		return
	}

	if err = message.Delete(cascade); err != nil {
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
func (m *MessagesController) checkBeforeDelete(ctx *gin.Context, message models.Message, user models.User, force bool) (models.Topic, error) {
	topic := models.Topic{}
	if err := topic.FindByTopic(message.Topics[0], true, false, false, nil); err != nil {
		e := fmt.Sprintf("Topic %s does not exist", message.Topics[0])
		ctx.JSON(http.StatusNotFound, gin.H{"error": e})
		return topic, fmt.Errorf(e)
	}

	if isRw := topic.IsUserRW(&user); !isRw {
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
			// if message is in user's tasks, can delete it
			if topicName == "/Private/"+user.Username+"/Tasks" {
				continue
			}
			// if label done on msg, can delete it
			if !force && !message.ContainsLabel("done") {
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
		if err := message.Like(user); err != nil {
			log.Errorf("Error while like a message %s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		info = "like added"
	} else if action == "unlike" {
		if err := message.Unlike(user); err != nil {
			log.Errorf("Error while unlike a message %s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		info = "like removed"
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid action: " + action)})
		return
	}
	go models.WSMessage(&models.WSMessageJSON{Action: action, Username: user.Username, Message: message})
	ctx.JSON(http.StatusCreated, gin.H{"info": info, "message": message})
}

func (m *MessagesController) addOrRemoveLabel(ctx *gin.Context, messageIn *models.MessageJSON, message models.Message, user models.User, topic models.Topic) {
	if messageIn.Text == "" && messageIn.Action != "relabel" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("Invalid Text for label"))
		return
	}
	info := gin.H{}
	if messageIn.Action == "label" {
		addedLabel, err := message.AddLabel(topic, messageIn.Text, messageIn.Option)
		if err != nil {
			errInfo := fmt.Sprintf("Error while adding a label to a message %s", err.Error())
			log.Errorf(errInfo)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": errInfo})
			return
		}
		info = gin.H{"info": fmt.Sprintf("label %s added to message", addedLabel.Text), "label": addedLabel, "message": message}
	} else if messageIn.Action == "unlabel" {
		if err := message.RemoveLabel(messageIn.Text); err != nil {
			errInfo := fmt.Sprintf("Error while removing a label from a message %s", err.Error())
			log.Errorf(errInfo)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": errInfo})
			return
		}
		info = gin.H{"info": fmt.Sprintf("label %s removed from message", messageIn.Text), "message": message}
	} else if messageIn.Action == "relabel" {
		if err := message.RemoveAllAndAddNewLabel(messageIn.Labels); err != nil {
			errInfo := fmt.Sprintf("Error while removing all labels and add new ones for a message %s", err.Error())
			log.Errorf(errInfo)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": errInfo})
			return
		}
		info = gin.H{"info": fmt.Sprintf("all labels removed and new labels %s added to message", messageIn.Text), "message": message}
	} else {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("Invalid action: "+messageIn.Action))
		return
	}
	go models.WSMessage(&models.WSMessageJSON{Action: messageIn.Action, Username: user.Username, Message: message})
	ctx.JSON(http.StatusCreated, info)
}

func (m *MessagesController) voteMessage(ctx *gin.Context, messageIn *models.MessageJSON, message models.Message, user models.User, topic models.Topic) {
	info := ""
	errInfo := ""
	if messageIn.Action == "voteup" {
		if err := message.VoteUP(user); err != nil {
			errInfo = fmt.Sprintf("Error while vote up a message %s", err.Error())
		}
		info = "Vote UP added to message"
	} else if messageIn.Action == "votedown" {
		if err := message.VoteDown(user); err != nil {
			errInfo = fmt.Sprintf("Error while vote down a message %s", err.Error())
		}
		info = "Vote Down added to message"
	} else if messageIn.Action == "unvoteup" {
		if err := message.UnVoteUP(user); err != nil {
			errInfo = fmt.Sprintf("Error while remove vote up from message %s", err.Error())
		}
		info = "Vote UP removed from message"
	} else if messageIn.Action == "unvotedown" {
		if err := message.UnVoteDown(user); err != nil {
			errInfo = fmt.Sprintf("Error while remove vote down from message %s", err.Error())
		}
		info = "Vote Down removed from message"
	} else {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("Invalid action: "+messageIn.Action))
		return
	}
	if errInfo != "" {
		log.Errorf(errInfo)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errInfo})
		return
	}
	if err := message.FindByID(messageIn.IDReference); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching message after voting"})
		return
	}
	go models.WSMessage(&models.WSMessageJSON{Action: messageIn.Action, Username: user.Username, Message: message})
	ctx.JSON(http.StatusCreated, gin.H{"info": info, "message": message})
}

func (m *MessagesController) addOrRemoveTask(ctx *gin.Context, messageIn *models.MessageJSON, message models.Message, user models.User, topic models.Topic) {
	info := ""
	if messageIn.Action == "task" {
		if message.InReplyOfIDRoot != "" {
			log.Warnf("This message is a reply, you can't task it (%s)", message.ID)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "This message is a reply, you can't task it"})
			return
		}
		if err := message.AddToTasks(user, topic); err != nil {
			log.Errorf("Error while adding a message to tasks %s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error while adding a message to tasks"})
			return
		}
		info = fmt.Sprintf("New Task created in %s", models.GetPrivateTopicTaskName(user))
	} else if messageIn.Action == "untask" {
		if err := message.RemoveFromTasks(user, topic); err != nil {
			log.Errorf("Error while removing a message from tasks %s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		info = fmt.Sprintf("Task removed from %s", models.GetPrivateTopicTaskName(user))
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action: " + messageIn.Action})
		return
	}
	go models.WSMessage(&models.WSMessageJSON{Action: messageIn.Action, Username: user.Username, Message: message})
	ctx.JSON(http.StatusCreated, gin.H{"info": info})
}

func (m *MessagesController) updateMessage(ctx *gin.Context, messageIn *models.MessageJSON, message models.Message, user models.User, topic models.Topic) {
	info := ""

	if !topic.CanUpdateMsg && !topic.CanUpdateAllMsg {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("You can't update a message on topic %s", topic.Topic)})
		return
	}

	if !topic.CanUpdateAllMsg && message.Author.Username != user.Username {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Could not update a message from another user %s than you %s", message.Author.Username, user.Username)})
		return
	}

	err := message.Update(user, topic, messageIn.Text, messageIn.Action)
	if err != nil {
		log.Errorf("Error while update a message %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	info = fmt.Sprintf("Message updated in %s", topic.Topic)

	go models.WSMessage(&models.WSMessageJSON{Action: messageIn.Action, Username: user.Username, Message: message})
	out := &models.MessageJSONOut{Message: message, Info: info}
	ctx.JSON(http.StatusOK, out)
}

func (m *MessagesController) moveMessage(ctx *gin.Context, messageIn *models.MessageJSON, message models.Message, user models.User, topic models.Topic) {
	// Check if user can delete msg on from topic
	_, err := m.checkBeforeDelete(ctx, message, user, true)
	if err != nil {
		// ctx writes in checkBeforeDelete
		return
	}

	// Check if user can write msg from dest topic
	if !topic.IsUserRW(&user) {
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action: " + messageIn.Action})
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

func checkDMTopic(ctx *gin.Context, topicName string) (*models.Topic, string, error) {
	var topic = models.Topic{}

	topicParentName := "/Private/" + utils.GetCtxUsername(ctx) + "/DM"
	if !strings.HasPrefix(topicName, topicParentName+"/") {
		log.Errorf("wrong topic name for DM:" + topicName)
		return &topic, "", errors.New("Wrong tpic name for DM:" + topicName)
	}

	// /Private/usernameFrom/DM/usernameTO
	part := strings.Split(topicName, "/")
	if len(part) != 5 {
		log.Errorf("wrong topic name for DM")
		return &topic, "", errors.New("Wrong topic name for DM:" + topicName)
	}

	var userFrom = models.User{}
	err := userFrom.FindByUsername(utils.GetCtxUsername(ctx))
	if err != nil {
		return &topic, "", errors.New("Error while fetching user.")
	}
	var userTo = models.User{}
	usernameTo := part[4]
	err = userTo.FindByUsername(usernameTo)
	if err != nil {
		return &topic, "", errors.New("Error while fetching user.")
	}

	if err = checkTopicParentDM(userFrom); err != nil {
		return &topic, "", errors.New(err.Error())
	}

	if err = checkTopicParentDM(userTo); err != nil {
		return &topic, "", errors.New(err.Error())
	}

	topic, err = insertTopicDM(userFrom, userTo)
	if err != nil {
		return &topic, "", errors.New(err.Error())
	}

	_, err = insertTopicDM(userTo, userFrom)
	if err != nil {
		return &topic, "", errors.New(err.Error())
	}

	topicCriteria := topicName + "," + "/Private/" + usernameTo + "/DM/" + userFrom.Username
	return &topic, topicCriteria, nil
}

func insertTopicDM(userFrom, userTo models.User) (models.Topic, error) {
	var topic = models.Topic{}
	topicName := "/Private/" + userFrom.Username + "/DM/" + userTo.Username
	topic.Topic = topicName
	topic.Description = userTo.Fullname
	if err := topic.Insert(&userFrom); err != nil {
		log.Errorf("Error while InsertTopic %s", err)
		return topic, err
	}
	return topic, nil
}

func checkTopicParentDM(user models.User) error {
	topicName := "/Private/" + user.Username + "/DM"
	var topicParent = models.Topic{}
	err := topicParent.FindByTopic(topicName, false, false, false, nil)
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
