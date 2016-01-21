package models

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ovh/tat/utils"
)

// WSPresenceJSON is used by Tat websocket
type WSPresenceJSON struct {
	Action   string   `json:"action"`
	Presence Presence `json:"presence"`
}

// WSMessageJSON is used by Tat websocket
// From Tat to client
type WSMessageJSON struct {
	Action   string  `json:"action"`
	Username string  `json:"username"`
	Message  Message `json:"message"`
}

// WSMessageNewJSON is used by Tat weNewet
// From Tat to client
type WSMessageNewJSON struct {
	Topic string `json:"topic"`
}

// WSUserJSON is used by Tat websocket
// From Tat to client
type WSUserJSON struct {
	Action   string `json:"action"`
	Username string `json:"username"`
}

// WSJSON represents a json from client to tat, except connect action
type WSJSON struct {
	Action   string   `json:"action"`
	Status   string   `json:"status"`
	TreeView string   `json:"treeView"`
	Topics   []string `json:"topics"`
}

// WSConnectJSON represents a json from client to tat, connect action
type WSConnectJSON struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Socket struct
type Socket struct {
	Connection *websocket.Conn
	IsAdmin    bool
	username   string
	instance   string
	mutexWrite *sync.Mutex
}

// key Socket.instance, value websocket.Conn
var activeUsers = struct {
	sync.RWMutex
	m map[string]*Socket
}{m: make(map[string]*Socket)}

type subscriptionVal struct {
	instance string
	treeView string
}

// key topic, list of subscriptionMsgVal
var subscriptionMessages = struct {
	sync.RWMutex
	m map[string][]subscriptionVal
}{m: make(map[string][]subscriptionVal)}

// key topic, list of subscriptionVal
var subscriptionMessagesNew = struct {
	sync.RWMutex
	m map[string][]subscriptionVal
}{m: make(map[string][]subscriptionVal)}

// key topic, list of subscriptionVal
var subscriptionPresences = struct {
	sync.RWMutex
	m map[string][]subscriptionVal
}{m: make(map[string][]subscriptionVal)}

// key Socket.instance
var subscriptionUsers = struct {
	sync.RWMutex
	m map[string]int
}{m: make(map[string]int)}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// SocketsDump returns all struct data
func SocketsDump() gin.H {
	return gin.H{
		"activeUsers":             activeUsers.m,
		"subscriptionMessages":    fmt.Sprintf("%+v", subscriptionMessages.m),
		"subscriptionMessagesNew": fmt.Sprintf("%+v", subscriptionMessagesNew.m),
		"subscriptionPresences":   fmt.Sprintf("%+v", subscriptionPresences.m),
		"subscriptionUsers":       fmt.Sprintf("%+v", subscriptionUsers.m),
	}
}

// Open initializes a new Tat Socket and read / write on websocket
func (socket *Socket) Open() {
	socket.mutexWrite = &sync.Mutex{}
	m, err := socket.getMsgConnect()
	if err != nil {
		log.Errorf("Error while getting first Msg %s", err.Error())
		return
	}

	err = socket.actionConnect(m)
	if err != nil {
		log.Errorf("Error while first connection try %s", err.Error())
		return
	}

	go socket.writePing()
	socket.read()
}

func (socket *Socket) read() {
	socket.Connection.SetReadLimit(maxMessageSize)
	socket.Connection.SetPongHandler(func(string) error {
		log.Debugf("Pong with %s", socket.instance)
		socket.Connection.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		m, err := socket.getMsg()
		if err != nil {
			break
		}
		socket.work(m)
	}
	log.Debugf("Out Loop for %s", socket.instance)
	socket.close()
}

func (socket *Socket) getMsgConnect() (WSConnectJSON, error) {
	m := WSConnectJSON{}
	err := socket.Connection.ReadJSON(&m)
	if err != nil {
		log.Debugf("Invalid ReadJSON %s", err)
		return WSConnectJSON{}, err
	}
	log.Debugf("Action Connect from new user")

	return m, nil
}

func (socket *Socket) getMsg() (WSJSON, error) {
	m := WSJSON{}
	err := socket.Connection.ReadJSON(&m)
	if err != nil {
		log.Debugf("Invalid ReadJSON %s", err)
		return WSJSON{}, err
	}
	log.Infof("Action %s from %s", m.Action, socket.instance)

	return m, nil
}

func (socket *Socket) writePing() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		socket.close()
	}()

	for {
		select {
		case <-ticker.C:
			log.Debugf("Write ping for %s", socket.instance)
			if err := socket.ping(); err != nil {
				log.Debugf("Error ping for %s", socket.instance)
				return
			}
		}
	}
}

func (socket *Socket) close() {
	log.Debugf("Call close for %s", socket.instance)
	socket.deleteUserFromAll()
	socket.Connection.Close()
}

func closeSocketOfUsername(username string) {
	log.Debugf("Call Socket for username %s", username)

	for _, s := range activeUsers.m {
		if s.username == username {
			s.close()
		}
	}
}

func (socket *Socket) actionConnect(msg WSConnectJSON) error {
	user := User{}
	username := strings.Trim(msg.Username, "")
	err := user.FindByUsernameAndPassword(username, strings.Trim(msg.Password, ""))
	if err != nil {
		return fmt.Errorf("Invalid credentials for username %s, err:%s", username, err.Error())
	}

	log.Infof("WS Connection Ok for %s", username)

	if err != nil {
		log.Errorf("Error with connect: %s", err.Error())
	}
	salt, err := utils.GenerateSalt()
	if err != nil {
		return err
	}

	instance := user.Username + "$" + salt
	socket.IsAdmin = user.IsAdmin
	socket.username = user.Username
	socket.instance = instance
	activeUsers.Lock()
	activeUsers.m[instance] = socket
	activeUsers.Unlock()
	socket.write(gin.H{"action": "connect", "result": fmt.Sprintf("connect OK to Tat Engine"), "status": http.StatusOK})
	return nil
}

func (socket *Socket) work(msg WSJSON) {
	log.Debugf("Work with %s, %+v", socket.instance, msg)
	switch msg.Action {
	case "subscribeMessages":
		socket.actionSubscribeMessages(msg)
	case "unsubscribeMessages":
		socket.actionUnsubscribeMessages(msg)
	case "subscribeMessagesNew":
		socket.actionSubscribeMessagesNew(msg)
	case "unsubscribeMessagesNew":
		socket.actionUnsubscribeMessagesNew(msg)
	case "subscribePresences":
		socket.actionSubscribePresences(msg)
	case "unsubscribePresences":
		socket.actionUnsubscribePresences(msg)
	case "subscribeUsers":
		socket.actionSubscribeUsers(msg)
	case "unsubscribeUsers":
		socket.actionUnsubscribeUsers(msg)
	case "writePresence":
		socket.actionWritePresence(msg)
	default:
		log.Errorf("Invalid Action %s", msg.Action)
	}
}

func (socket *Socket) preCheckWSTopics(msg WSJSON) ([]Topic, User, error) {
	count, topics, user, err := socket.getTopicsOfUser(msg)
	if err != nil {
		return topics, user, err
	}

	// at least one topic found with "all"
	if count > 1 {
		return topics, user, nil
	}

	if len(msg.Topics) < 1 {
		m := fmt.Sprintf("Invalid number of args (%d) for action %s", len(msg.Topics), msg.Action)
		socket.write(gin.H{"action": msg.Action, "result": m, "status": http.StatusBadRequest})
		return []Topic{}, User{}, errors.New(m)
	}
	topics = make([]Topic, len(msg.Topics))

	for i, topicName := range msg.Topics {
		var topic = Topic{}
		err := topic.FindByTopic(strings.Trim(topicName, " "), true, nil)
		if err != nil {
			m := fmt.Sprintf("Invalid topic (%s) for action %s", topicName, msg.Action)
			log.Errorf("%s error:%s", m, err.Error())
			socket.write(gin.H{"action": msg.Action, "result": m, "status": http.StatusBadRequest})
			return []Topic{}, User{}, errors.New(m)
		}

		isReadAccess := topic.IsUserReadAccess(user)
		if !isReadAccess {
			m := fmt.Sprintf("No Read Access on topic %s for action %s", topicName, msg.Action)
			socket.write(gin.H{"action": msg.Action, "result": m, "status": http.StatusForbidden})
			return []Topic{}, User{}, errors.New(m)
		}
		topics[i] = topic
	}
	return topics, user, nil
}

func (socket *Socket) getTopicsOfUser(msg WSJSON) (int, []Topic, User, error) {
	var user = User{}
	err := user.FindByUsername(socket.username)
	if err != nil {
		m := fmt.Sprintf("Internal Error getting User for action %s", msg.Action)
		log.Errorf("%s :%s", m, err)
		socket.write(gin.H{"action": msg.Action, "result": m, "status": http.StatusInternalServerError})
		return 0, []Topic{}, User{}, errors.New(m)
	}

	if len(msg.Topics) == 1 && strings.Trim(msg.Topics[0], " ") == "all" {
		c := TopicCriteria{}
		c.Skip = 0
		c.Limit = 1000

		count, topics, err := ListTopics(&c, &user)
		if err != nil {
			m := fmt.Sprintf("Error while getting topics for action %s", msg.Action)
			socket.write(gin.H{"action": msg.Action, "result": m, "status": http.StatusBadRequest})
			return count, []Topic{}, user, errors.New(m)
		}
		return count, topics, user, nil
	}
	return 0, []Topic{}, user, nil
}

func (socket *Socket) actionSubscribeMessages(msg WSJSON) {
	topics, _, err := socket.preCheckWSTopics(msg)
	if err != nil {
		return
	}
	sVal := subscriptionVal{
		instance: socket.instance,
		treeView: msg.TreeView,
	}
	subscriptionMessages.Lock()
	for _, topic := range topics {
		if subscriptionArrContains(subscriptionMessages.m[topic.Topic], socket.instance) {
			socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("%s topic %s KO : already Subscribe", msg.Action, topic.Topic), "status": http.StatusConflict})
		} else {
			subscriptionMessages.m[topic.Topic] = append(subscriptionMessages.m[topic.Topic], sVal)
			socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("%s topic %s OK", msg.Action, topic.Topic), "status": http.StatusOK})
		}
	}
	subscriptionMessages.Unlock()
}

func (socket *Socket) actionUnsubscribeMessages(msg WSJSON) {
	topics, _, err := socket.preCheckWSTopics(msg)
	if err != nil {
		return
	}

	for _, topic := range topics {
		socket.deleteUserFromMessage(topic.Topic)
		socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("%s topic %s OK", msg.Action, topic.Topic), "status": http.StatusOK})
	}
}

func (socket *Socket) actionSubscribeMessagesNew(msg WSJSON) {
	topics, _, err := socket.preCheckWSTopics(msg)
	if err != nil {
		return
	}
	sVal := subscriptionVal{
		instance: socket.instance,
	}
	subscriptionMessagesNew.Lock()
	for _, topic := range topics {
		if subscriptionArrContains(subscriptionMessagesNew.m[topic.Topic], socket.instance) {
			socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("%s topic %s KO : already Subscribe", msg.Action, topic.Topic), "status": http.StatusConflict})
		} else {
			subscriptionMessagesNew.m[topic.Topic] = append(subscriptionMessagesNew.m[topic.Topic], sVal)
			socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("%s topic %s OK", msg.Action, topic.Topic), "status": http.StatusOK})
		}
	}
	subscriptionMessagesNew.Unlock()
}

func (socket *Socket) actionUnsubscribeMessagesNew(msg WSJSON) {
	topics, _, err := socket.preCheckWSTopics(msg)
	if err != nil {
		return
	}

	for _, topic := range topics {
		socket.deleteUserFromMessageCount(topic.Topic)
		socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("%s topic %s OK", msg.Action, topic.Topic), "status": http.StatusOK})
	}
}

func (socket *Socket) actionSubscribePresences(msg WSJSON) {
	topics, _, err := socket.preCheckWSTopics(msg)
	if err != nil {
		return
	}
	sVal := subscriptionVal{
		instance: socket.instance,
	}
	subscriptionPresences.Lock()
	for _, topic := range topics {
		if subscriptionArrContains(subscriptionPresences.m[topic.Topic], socket.instance) {
			socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("%s topic %s KO : already Subscribe", msg.Action, topic.Topic), "status": http.StatusConflict})
		} else {
			subscriptionPresences.m[topic.Topic] = append(subscriptionPresences.m[topic.Topic], sVal)
			socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("%s presence %s OK", msg.Action, topic.Topic), "status": http.StatusOK})
		}
	}
	subscriptionPresences.Unlock()
}

func (socket *Socket) actionUnsubscribePresences(msg WSJSON) {
	topics, _, err := socket.preCheckWSTopics(msg)
	if err != nil {
		return
	}
	for _, topic := range topics {
		socket.deleteUserFromPresence(topic.Topic)
		socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("%s presence %s OK", msg.Action, topic.Topic), "status": http.StatusOK})
	}
}

func (socket *Socket) actionWritePresence(msg WSJSON) {

	if len(msg.Topics) < 1 {
		socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("Invalid topic for action %s", msg.Action), "status": http.StatusBadRequest})
		return
	}
	topics, user, err := socket.preCheckWSTopics(msg)
	if err != nil {
		return
	}
	for _, topic := range topics {
		var presence = Presence{}
		err := presence.Upsert(user, topic, msg.Status)
		if err != nil {
			socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("Error while %s on topic %s OK", msg.Action, topic.Topic), "status": http.StatusInternalServerError})
		} else {
			socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("%s on topic %s OK", msg.Action, topic.Topic), "status": http.StatusOK})
			go WSPresence(&WSPresenceJSON{Action: "create", Presence: presence})
		}
	}
}

func (socket *Socket) actionSubscribeUsers(msg WSJSON) error {
	if socket.IsAdmin {
		subscriptionUsers.Lock()
		subscriptionUsers.m[socket.instance] = 0
		subscriptionUsers.Unlock()
		socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("%s OK", msg.Action), "status": http.StatusOK})
	} else {
		socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("%s KO", msg.Action), "status": http.StatusForbidden})
	}
	return nil
}

func (socket *Socket) actionUnsubscribeUsers(msg WSJSON) error {
	subscriptionUsers.Lock()
	delete(subscriptionUsers.m, socket.instance)
	subscriptionUsers.Unlock()
	socket.write(gin.H{"action": msg.Action, "result": fmt.Sprintf("%s OK", msg.Action), "status": http.StatusOK})
	return nil
}

func (socket *Socket) write(w gin.H) {
	go func(socket *Socket) {
		socket.mutexWrite.Lock()
		defer socket.mutexWrite.Unlock()
		socket.Connection.SetWriteDeadline(time.Now().Add(writeWait))
		if err := socket.Connection.WriteJSON(w); err != nil {
			log.Errorf("Error while WriteJSON:%s", err.Error())
		}
	}(socket)
}

// ping writes a ping message
func (socket *Socket) ping() error {
	return socket.Connection.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait))
}

func (socket *Socket) deleteUserFromAll() {
	socket.deleteUserFromMessages()
	socket.deleteUserFromPresences()
	subscriptionUsers.Lock()
	delete(subscriptionUsers.m, socket.instance)
	subscriptionUsers.Unlock()
	activeUsers.Lock()
	delete(activeUsers.m, socket.instance)
	activeUsers.Unlock()
}

func (socket *Socket) deleteUserFromPresences() {
	subscriptionPresences.Lock()
	socket.deleteUserFromAllList(subscriptionPresences.m)
	subscriptionPresences.Unlock()
}

func (socket *Socket) deleteUserFromPresence(topicName string) {
	subscriptionPresences.Lock()
	socket.deleteUserFromList(topicName, subscriptionPresences.m)
	subscriptionPresences.Unlock()
}

func (socket *Socket) deleteUserFromMessages() {
	subscriptionMessages.Lock()
	socket.deleteUserFromAllList(subscriptionMessages.m)
	subscriptionMessages.Unlock()
}

func (socket *Socket) deleteUserFromMessage(topicName string) {
	subscriptionMessages.Lock()
	socket.deleteUserFromList(topicName, subscriptionMessages.m)
	subscriptionMessages.Unlock()
}

func (socket *Socket) deleteUserFromMessagesCount() {
	subscriptionMessagesNew.Lock()
	socket.deleteUserFromAllList(subscriptionMessagesNew.m)
	subscriptionMessagesNew.Unlock()
}

func (socket *Socket) deleteUserFromMessageCount(topicName string) {
	subscriptionMessagesNew.Lock()
	socket.deleteUserFromList(topicName, subscriptionMessagesNew.m)
	subscriptionMessagesNew.Unlock()
}

func (socket *Socket) deleteUserFromAllList(lst map[string][]subscriptionVal) {
	for key := range lst {
		socket.deleteUserFromList(key, lst)
	}
}

func (socket *Socket) deleteUserFromList(key string, lst map[string][]subscriptionVal) {
	for i, u := range lst[key] {
		if u.instance == socket.instance {
			if len(lst[key]) <= 1 {
				delete(lst, key)
				break
			} else {
				lst[key] = append(lst[key][:i], lst[key][i+1:]...)
			}
		}
	}
}

// WSMessageNew writes event messagesCount
func WSMessageNew(msg *WSMessageNewJSON) {
	w := gin.H{"eventMsgNew": msg}

	subscriptionMessagesNew.RLock()
	for _, sVal := range subscriptionMessagesNew.m[msg.Topic] {
		activeUsers.RLock()
		activeUsers.m[sVal.instance].write(w)
		activeUsers.RUnlock()
	}
	subscriptionMessagesNew.RUnlock()
}

// WSMessage writes event messages
func WSMessage(msg *WSMessageJSON) {
	w := gin.H{"eventMsg": msg}

	var oneTree, fullTree, msgs []Message
	var wOneTree, wFullTree = gin.H{}, gin.H{}
	c := &MessageCriteria{}
	c.AllIDMessage = msg.Message.InReplyOfIDRoot
	isGetMsgs := false

	subscriptionMessages.RLock()
	for _, sVal := range subscriptionMessages.m[msg.Message.Topics[0]] {
		activeUsers.RLock()

		if msg.Message.InReplyOfIDRoot == "" || sVal.treeView == "" {
			activeUsers.m[sVal.instance].write(w)
			activeUsers.RUnlock()
			continue
		}

		// below is for treeView

		// getting messages related to ID, only once
		if !isGetMsgs {
			msgs, _ = ListMessages(c)
			isGetMsgs = true
		}

		switch sVal.treeView {
		case "onetree":
			// only once to call oneTreeMessages
			if len(oneTree) == 0 {
				oneTree, _ = oneTreeMessages(msgs, 1, c)
				if len(oneTree) == 0 {
					// fake init oneTree
					oneTree = []Message{msg.Message}
					wOneTree = gin.H{"eventMsgNew": msg}
				} else {
					wOneTree = gin.H{"eventMsgNew": msg, "oneTree": oneTree[0]}
				}
			}
			activeUsers.m[sVal.instance].write(wOneTree)
		case "fulltree":
			// only once to call oneTreeMessages
			if len(fullTree) == 0 {
				fullTree, _ = fullTreeMessages(msgs, 1, c)
				if len(fullTree) == 0 {
					// fake init fullTree
					fullTree = []Message{msg.Message}
					wFullTree = gin.H{"eventMsgNew": msg}
				} else {
					wFullTree = gin.H{"eventMsgNew": msg, "fullTree": fullTree[0]}
				}
			}
			activeUsers.m[sVal.instance].write(wFullTree)
		default:
			log.Warnf("Invalid souscription tree, send no tree")
			activeUsers.m[sVal.instance].write(w)
		}
		activeUsers.RUnlock()
	}
	subscriptionMessages.RUnlock()
}

// WSPresence writes event presences
func WSPresence(p *WSPresenceJSON) {
	w := gin.H{"eventPresence": p}
	subscriptionPresences.RLock()
	for _, sVal := range subscriptionPresences.m[p.Presence.Topic] {
		activeUsers.RLock()
		activeUsers.m[sVal.instance].write(w)
		activeUsers.RUnlock()
	}
	subscriptionPresences.RUnlock()
}

// WSUser writes event users
func WSUser(u *WSUserJSON) {
	w := gin.H{"eventUser": u}
	subscriptionUsers.RLock()
	for instance := range subscriptionUsers.m {
		activeUsers.Lock()
		activeUsers.m[instance].write(w)
		activeUsers.RLock()
	}
	subscriptionUsers.RUnlock()
}

func subscriptionArrContains(array []subscriptionVal, instance string) bool {
	for _, cur := range array {
		if cur.instance == instance {
			return true
		}
	}
	return false
}
