package tat

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
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
	Username   string
	Instance   string
	MutexWrite *sync.Mutex
}

// SocketDump returns current sockets
func (c *Client) SocketDump() ([]byte, error) {
	return c.simpleGetAndGetBytes("/sockets/dump")
}

// SocketMessages TODO
func (c *Client) SocketMessages() error {
	return fmt.Errorf("NOT YET IMPLEMENTED IN SDK")
}

// SocketMessagesNew TODO
func (c *Client) SocketMessagesNew() error {
	return fmt.Errorf("NOT YET IMPLEMENTED IN SDK")
}

// SocketInteractive TODO
func (c *Client) SocketInteractive() error {
	return fmt.Errorf("NOT YET IMPLEMENTED IN SDK")
}

// SocketUsers TODO
func (c *Client) SocketUsers() error {
	return fmt.Errorf("NOT YET IMPLEMENTED IN SDK")
}
