package socket

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
	"github.com/ovh/tat"
	messageDB "github.com/ovh/tat/api/message"
	presenceDB "github.com/ovh/tat/api/presence"
	topicDB "github.com/ovh/tat/api/topic"
	userDB "github.com/ovh/tat/api/user"
)

// key Socket.instance, value websocket.Conn
var activeUsers = struct {
	sync.RWMutex
	m map[string]*tat.Socket
}{m: make(map[string]*tat.Socket)}

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
func Open(socket *tat.Socket) {
	socket.MutexWrite = &sync.Mutex{}
	m, err := getMsgConnect(socket)
	if err != nil {
		log.Errorf("Error while getting first Msg %s", err.Error())
		return
	}

	err = actionConnect(socket, m)
	if err != nil {
		log.Errorf("Error while first connection try %s", err.Error())
		return
	}

	go writePing(socket)
	read(socket)
}

func read(socket *tat.Socket) {
	socket.Connection.SetReadLimit(maxMessageSize)
	socket.Connection.SetPongHandler(func(string) error {
		log.Debugf("Pong with %s", socket.Instance)
		socket.Connection.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		m, err := getMsg(socket)
		if err != nil {
			break
		}
		work(socket, m)
	}
	log.Debugf("Out Loop for %s", socket.Instance)
	close(socket)
}

func getMsgConnect(socket *tat.Socket) (tat.WSConnectJSON, error) {
	m := tat.WSConnectJSON{}
	err := socket.Connection.ReadJSON(&m)
	if err != nil {
		log.Debugf("Invalid ReadJSON %s", err)
		return tat.WSConnectJSON{}, err
	}
	log.Debugf("Action Connect from new user")

	return m, nil
}

func getMsg(socket *tat.Socket) (tat.WSJSON, error) {
	m := tat.WSJSON{}
	err := socket.Connection.ReadJSON(&m)
	if err != nil {
		log.Debugf("Invalid ReadJSON %s", err)
		return tat.WSJSON{}, err
	}
	log.Infof("Action %s from %s", m.Action, socket.Instance)

	return m, nil
}

func writePing(socket *tat.Socket) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		close(socket)
	}()

	for {
		select {
		case <-ticker.C:
			log.Debugf("Write ping for %s", socket.Instance)
			if err := ping(socket); err != nil {
				log.Debugf("Error ping for %s", socket.Instance)
				return
			}
		}
	}
}

func close(socket *tat.Socket) {
	log.Debugf("Call close for %s", socket.Instance)
	deleteUserFromAll(socket)
	socket.Connection.Close()
}

// CloseSocketOfUsername closes socket for a username
func CloseSocketOfUsername(username string) {
	log.Debugf("Call Socket for username %s", username)

	for _, s := range activeUsers.m {
		if s.Username == username {
			close(s)
		}
	}
}

func actionConnect(socket *tat.Socket, msg tat.WSConnectJSON) error {
	user := tat.User{}
	username := strings.Trim(msg.Username, "")
	found, err := userDB.FindByUsernameAndPassword(&user, username, strings.Trim(msg.Password, ""))
	if !found {
		return fmt.Errorf("Invalid credentials for username %s, err:%s", username, err.Error())
	} else if err != nil {
		return fmt.Errorf("Error while connecting username %s, err:%s", username, err.Error())
	}

	log.Infof("WS Connection Ok for %s", username)

	salt, err := userDB.GenerateSalt()
	if err != nil {
		return err
	}

	instance := user.Username + "$" + salt
	socket.IsAdmin = user.IsAdmin
	socket.Username = user.Username
	socket.Instance = instance
	activeUsers.Lock()
	activeUsers.m[instance] = socket
	activeUsers.Unlock()
	write(socket, gin.H{"action": "connect", "result": fmt.Sprintf("connect OK to Tat Engine"), "status": http.StatusOK})
	return nil
}

func work(socket *tat.Socket, msg tat.WSJSON) {
	log.Debugf("Work with %s, %+v", socket.Instance, msg)
	switch msg.Action {
	case "subscribeMessages":
		actionSubscribeMessages(socket, msg)
	case "unsubscribeMessages":
		actionUnsubscribeMessages(socket, msg)
	case "subscribeMessagesNew":
		actionSubscribeMessagesNew(socket, msg)
	case "unsubscribeMessagesNew":
		actionUnsubscribeMessagesNew(socket, msg)
	case "subscribePresences":
		actionSubscribePresences(socket, msg)
	case "unsubscribePresences":
		actionUnsubscribePresences(socket, msg)
	case "subscribeUsers":
		actionSubscribeUsers(socket, msg)
	case "unsubscribeUsers":
		actionUnsubscribeUsers(socket, msg)
	case "writePresence":
		actionWritePresence(socket, msg)
	default:
		log.Errorf("Invalid Action %s", msg.Action)
	}
}

func preCheckWSTopics(socket *tat.Socket, msg tat.WSJSON) ([]tat.Topic, tat.User, error) {
	var user = tat.User{}
	found, err := userDB.FindByUsername(&user, socket.Username)
	if !found || err != nil {
		m := fmt.Sprintf("Internal Error getting User for action %s", msg.Action)
		log.Errorf("%s :%s", m, err)
		write(socket, gin.H{"action": msg.Action, "result": m, "status": http.StatusInternalServerError})
		return []tat.Topic{}, tat.User{}, errors.New(m)
	}

	c := tat.TopicCriteria{}
	c.Skip = 0
	c.Limit = 1000
	if len(msg.Topics) == 1 && strings.Trim(msg.Topics[0], " ") == "all" {
		// nothing, we select all topics
	} else {
		c.Topic = strings.Join(msg.Topics, ",")
	}

	_, topics, err := topicDB.ListTopics(&c, &user, false, false, false)
	if err != nil {
		m := fmt.Sprintf("Error while getting topics for action %s", msg.Action)
		write(socket, gin.H{"action": msg.Action, "result": m, "status": http.StatusBadRequest})
		return nil, user, errors.New(m)
	}
	return topics, user, nil
}

func actionSubscribeMessages(socket *tat.Socket, msg tat.WSJSON) {
	topics, _, err := preCheckWSTopics(socket, msg)
	if err != nil {
		return
	}
	sVal := subscriptionVal{
		instance: socket.Instance,
		treeView: msg.TreeView,
	}
	subscriptionMessages.Lock()
	for _, topic := range topics {
		if subscriptionArrContains(subscriptionMessages.m[topic.Topic], socket.Instance) {
			write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("%s topic %s KO : already Subscribe", msg.Action, topic.Topic), "status": http.StatusConflict})
		} else {
			subscriptionMessages.m[topic.Topic] = append(subscriptionMessages.m[topic.Topic], sVal)
			write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("%s topic %s OK", msg.Action, topic.Topic), "status": http.StatusOK})
		}
	}
	subscriptionMessages.Unlock()
}

func actionUnsubscribeMessages(socket *tat.Socket, msg tat.WSJSON) {
	topics, _, err := preCheckWSTopics(socket, msg)
	if err != nil {
		return
	}

	for _, topic := range topics {
		deleteUserFromMessage(socket, topic.Topic)
		write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("%s topic %s OK", msg.Action, topic.Topic), "status": http.StatusOK})
	}
}

func actionSubscribeMessagesNew(socket *tat.Socket, msg tat.WSJSON) {
	topics, _, err := preCheckWSTopics(socket, msg)
	if err != nil {
		return
	}
	sVal := subscriptionVal{
		instance: socket.Instance,
	}
	subscriptionMessagesNew.Lock()
	for _, topic := range topics {
		if subscriptionArrContains(subscriptionMessagesNew.m[topic.Topic], socket.Instance) {
			write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("%s topic %s KO : already Subscribe", msg.Action, topic.Topic), "status": http.StatusConflict})
		} else {
			subscriptionMessagesNew.m[topic.Topic] = append(subscriptionMessagesNew.m[topic.Topic], sVal)
			write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("%s topic %s OK", msg.Action, topic.Topic), "status": http.StatusOK})
		}
	}
	subscriptionMessagesNew.Unlock()
}

func actionUnsubscribeMessagesNew(socket *tat.Socket, msg tat.WSJSON) {
	topics, _, err := preCheckWSTopics(socket, msg)
	if err != nil {
		return
	}

	for _, topic := range topics {
		deleteUserFromMessageCount(socket, topic.Topic)
		write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("%s topic %s OK", msg.Action, topic.Topic), "status": http.StatusOK})
	}
}

func actionSubscribePresences(socket *tat.Socket, msg tat.WSJSON) {
	topics, _, err := preCheckWSTopics(socket, msg)
	if err != nil {
		return
	}
	sVal := subscriptionVal{
		instance: socket.Instance,
	}
	subscriptionPresences.Lock()
	for _, topic := range topics {
		if subscriptionArrContains(subscriptionPresences.m[topic.Topic], socket.Instance) {
			write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("%s topic %s KO : already Subscribe", msg.Action, topic.Topic), "status": http.StatusConflict})
		} else {
			subscriptionPresences.m[topic.Topic] = append(subscriptionPresences.m[topic.Topic], sVal)
			write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("%s presence %s OK", msg.Action, topic.Topic), "status": http.StatusOK})
		}
	}
	subscriptionPresences.Unlock()
}

func actionUnsubscribePresences(socket *tat.Socket, msg tat.WSJSON) {
	topics, _, err := preCheckWSTopics(socket, msg)
	if err != nil {
		return
	}
	for _, topic := range topics {
		deleteUserFromPresence(socket, topic.Topic)
		write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("%s presence %s OK", msg.Action, topic.Topic), "status": http.StatusOK})
	}
}

func actionWritePresence(socket *tat.Socket, msg tat.WSJSON) {

	if len(msg.Topics) < 1 {
		write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("Invalid topic for action %s", msg.Action), "status": http.StatusBadRequest})
		return
	}
	topics, user, err := preCheckWSTopics(socket, msg)
	if err != nil {
		return
	}
	for _, topic := range topics {
		var presence = tat.Presence{}
		err := presenceDB.Upsert(&presence, user, topic, msg.Status)
		if err != nil {
			write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("Error while %s on topic %s OK", msg.Action, topic.Topic), "status": http.StatusInternalServerError})
		} else {
			write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("%s on topic %s OK", msg.Action, topic.Topic), "status": http.StatusOK})
			go WSPresence(&tat.WSPresenceJSON{Action: "create", Presence: presence})
		}
	}
}

func actionSubscribeUsers(socket *tat.Socket, msg tat.WSJSON) error {
	if socket.IsAdmin {
		subscriptionUsers.Lock()
		subscriptionUsers.m[socket.Instance] = 0
		subscriptionUsers.Unlock()
		write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("%s OK", msg.Action), "status": http.StatusOK})
	} else {
		write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("%s KO", msg.Action), "status": http.StatusForbidden})
	}
	return nil
}

func actionUnsubscribeUsers(socket *tat.Socket, msg tat.WSJSON) error {
	subscriptionUsers.Lock()
	delete(subscriptionUsers.m, socket.Instance)
	subscriptionUsers.Unlock()
	write(socket, gin.H{"action": msg.Action, "result": fmt.Sprintf("%s OK", msg.Action), "status": http.StatusOK})
	return nil
}

func write(socket *tat.Socket, w gin.H) {
	go func(socket *tat.Socket) {
		socket.MutexWrite.Lock()
		defer socket.MutexWrite.Unlock()
		socket.Connection.SetWriteDeadline(time.Now().Add(writeWait))
		if err := socket.Connection.WriteJSON(w); err != nil {
			log.Errorf("Error while WriteJSON:%s", err.Error())
		}
	}(socket)
}

// ping writes a ping message
func ping(socket *tat.Socket) error {
	return socket.Connection.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait))
}

func deleteUserFromAll(socket *tat.Socket) {
	deleteUserFromMessages(socket)
	deleteUserFromPresences(socket)
	subscriptionUsers.Lock()
	delete(subscriptionUsers.m, socket.Instance)
	subscriptionUsers.Unlock()
	activeUsers.Lock()
	delete(activeUsers.m, socket.Instance)
	activeUsers.Unlock()
}

func deleteUserFromPresences(socket *tat.Socket) {
	subscriptionPresences.Lock()
	deleteUserFromAllList(socket, subscriptionPresences.m)
	subscriptionPresences.Unlock()
}

func deleteUserFromPresence(socket *tat.Socket, topicName string) {
	subscriptionPresences.Lock()
	deleteUserFromList(socket, topicName, subscriptionPresences.m)
	subscriptionPresences.Unlock()
}

func deleteUserFromMessages(socket *tat.Socket) {
	subscriptionMessages.Lock()
	deleteUserFromAllList(socket, subscriptionMessages.m)
	subscriptionMessages.Unlock()
}

func deleteUserFromMessage(socket *tat.Socket, topicName string) {
	subscriptionMessages.Lock()
	deleteUserFromList(socket, topicName, subscriptionMessages.m)
	subscriptionMessages.Unlock()
}

func deleteUserFromMessageCount(socket *tat.Socket, topicName string) {
	subscriptionMessagesNew.Lock()
	deleteUserFromList(socket, topicName, subscriptionMessagesNew.m)
	subscriptionMessagesNew.Unlock()
}

func deleteUserFromAllList(socket *tat.Socket, lst map[string][]subscriptionVal) {
	for key := range lst {
		deleteUserFromList(socket, key, lst)
	}
}

func deleteUserFromList(socket *tat.Socket, key string, lst map[string][]subscriptionVal) {
	for i, u := range lst[key] {
		if u.instance == socket.Instance {
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
func WSMessageNew(msg *tat.WSMessageNewJSON) {
	w := gin.H{"eventMsgNew": msg}

	subscriptionMessagesNew.RLock()
	for _, sVal := range subscriptionMessagesNew.m[msg.Topic] {
		activeUsers.RLock()
		write(activeUsers.m[sVal.instance], w)
		activeUsers.RUnlock()
	}
	subscriptionMessagesNew.RUnlock()
}

// WSMessage writes event messages
func WSMessage(msg *tat.WSMessageJSON, topic tat.Topic) {
	w := gin.H{"eventMsg": msg}

	var oneTree, fullTree, msgs []tat.Message
	var wOneTree, wFullTree = gin.H{}, gin.H{}
	c := &tat.MessageCriteria{AllIDMessage: msg.Message.InReplyOfIDRoot}
	isGetMsgs := false

	subscriptionMessages.RLock()
	for _, sVal := range subscriptionMessages.m[msg.Message.Topic] {
		activeUsers.RLock()

		if msg.Message.InReplyOfIDRoot == "" || sVal.treeView == "" {
			write(activeUsers.m[sVal.instance], w)
			activeUsers.RUnlock()
			continue
		}

		// below is for treeView

		// getting messages related to ID, only once
		if !isGetMsgs {
			msgs, _ = messageDB.ListMessages(c, "", topic)
			isGetMsgs = true
		}

		switch sVal.treeView {
		case "onetree":
			// only once to call oneTreeMessages
			if len(oneTree) == 0 {
				oneTree, _ = messageDB.OneTreeMessages(msgs, 1, c, "", topic)
				if len(oneTree) == 0 {
					// fake init oneTree
					oneTree = []tat.Message{msg.Message}
					wOneTree = gin.H{"eventMsgNew": msg}
				} else {
					wOneTree = gin.H{"eventMsgNew": msg, "oneTree": oneTree[0]}
				}
			}
			write(activeUsers.m[sVal.instance], wOneTree)
		case "fulltree":
			// only once to call oneTreeMessages
			if len(fullTree) == 0 {
				fullTree, _ = messageDB.FullTreeMessages(msgs, 1, c, "", topic)
				if len(fullTree) == 0 {
					// fake init fullTree
					fullTree = []tat.Message{msg.Message}
					wFullTree = gin.H{"eventMsgNew": msg}
				} else {
					wFullTree = gin.H{"eventMsgNew": msg, "fullTree": fullTree[0]}
				}
			}
			write(activeUsers.m[sVal.instance], wFullTree)
		default:
			log.Warnf("Invalid souscription tree, send no tree")
			write(activeUsers.m[sVal.instance], w)
		}
		activeUsers.RUnlock()
	}
	subscriptionMessages.RUnlock()
}

// WSPresence writes event presences
func WSPresence(p *tat.WSPresenceJSON) {
	w := gin.H{"eventPresence": p}
	subscriptionPresences.RLock()
	for _, sVal := range subscriptionPresences.m[p.Presence.Topic] {
		activeUsers.RLock()
		write(activeUsers.m[sVal.instance], w)
		activeUsers.RUnlock()
	}
	subscriptionPresences.RUnlock()
}

// WSUser writes event users
func WSUser(u *tat.WSUserJSON) {
	w := gin.H{"eventUser": u}
	subscriptionUsers.RLock()
	for instance := range subscriptionUsers.m {
		activeUsers.Lock()
		write(activeUsers.m[instance], w)
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
