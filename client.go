package tat

import (
	"fmt"

	"github.com/ovh/tat"
)

//Client is the main struct to call TAT API
type Client struct {
	url      string
	user     string
	password string
}

//Options is a struct to initialize a TAT client
type Options struct {
	URL      string
	User     string
	Password string
}

//NewClient initialize a TAT client
func NewClient(opts Options) (*Client, error) {
	if opts.URL == "" || opts.User == "" || opts.Password == "" {
		return nil, fmt.Errorf("Invalid configuration")
	}
	c := &Client{
		url:      opts.URL,
		user:     opts.User,
		password: opts.Password,
	}

	return c, nil
}

func (c *Client) CreateTopic(t *TopicCreateJSON) (*tat.Topic, error) {
	if c == nil {
		return
	}
}
