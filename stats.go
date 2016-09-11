package tat

import (
	"encoding/json"
	"time"
)

// StatsCountJSON contains all globals counters
type StatsCountJSON struct {
	Date      int64     `json:"date"`
	DateHuman time.Time `json:"dateHuman"`
	Version   string    `json:"version"`
	Groups    int       `json:"groups"`
	Messages  int       `json:"messages"`
	Presences int       `json:"presences"`
	Topics    int       `json:"topics"`
	Users     int       `json:"users"`
}

// StatsCount calls GET /stats/count
func (c *Client) StatsCount() (*StatsCountJSON, error) {
	body, err := c.reqWant("GET", 200, "/stats/count", nil)
	if err != nil {
		return nil, err
	}

	out := &StatsCountJSON{}
	if err := json.Unmarshal(body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// StatsDistribution returns Stats Distribution per topics and per users
func (c *Client) StatsDistribution() ([]byte, error) {
	return c.simpleGetAndGetBytes("/stats/distribution")
}

// StatsDBStats returns DB Stats
func (c *Client) StatsDBStats() ([]byte, error) {
	return c.simpleGetAndGetBytes("/stats/db/stats")
}

// StatsDBServerStatus returns DB Server Status
func (c *Client) StatsDBServerStatus() ([]byte, error) {
	return c.simpleGetAndGetBytes("/stats/db/serverStatus")
}

// StatsDBReplSetGetConfig returns DB Relica Set Config
func (c *Client) StatsDBReplSetGetConfig() ([]byte, error) {
	return c.simpleGetAndGetBytes("/stats/db/replSetGetConfig")
}

// StatsDBReplSetGetStatus returns Replica Set Status
func (c *Client) StatsDBReplSetGetStatus() ([]byte, error) {
	return c.simpleGetAndGetBytes("/stats/db/replSetGetStatus")
}

// StatsDBCollections returns nb msg for each collections
func (c *Client) StatsDBCollections() ([]byte, error) {
	return c.simpleGetAndGetBytes("/stats/db/collections")
}

// StatsDBSlowestQueries returns DB slowest Queries
func (c *Client) StatsDBSlowestQueries() ([]byte, error) {
	return c.simpleGetAndGetBytes("/stats/db/slowestQueries")
}

// StatsInstance returns DB Instance
func (c *Client) StatsInstance() ([]byte, error) {
	return c.simpleGetAndGetBytes("/stats/instance")
}
