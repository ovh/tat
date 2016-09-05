package tat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"log"

	"github.com/facebookgo/httpcontrol"
)

// Client representes a Client configuration to connect to api
type Client struct {
	username       string
	password       string
	url            string
	referer        string
	requestTimeout time.Duration
	maxTries       uint
}

//Options is a struct to initialize a TAT client
type Options struct {
	Username       string
	Password       string
	URL            string
	Referer        string
	RequestTimeout time.Duration
	MaxTries       uint
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var HTTPClient httpClient

// DebugLogFunc is a function that logs the provided message with optional fmt.Sprintf-style arguments. By default, logs to the default log.Logger.
var DebugLogFunc func(string, ...interface{})

// ErrorLogFunc is a function that logs the provided message with optional fmt.Sprintf-style arguments. By default, logs to the default log.Logger.
var ErrorLogFunc = log.Printf

//ErrClientNotInitiliazed is a predifined Error
var ErrClientNotInitiliazed = fmt.Errorf("Client is not initialized")

//NewClient initialize a TAT client
func NewClient(opts Options) (*Client, error) {
	if opts.URL == "" || opts.Username == "" || opts.Password == "" {
		return nil, fmt.Errorf("Invalid configuratio:. usersane, password and referer have to be setted")
	}
	c := &Client{
		url:            opts.URL,
		username:       opts.Username,
		password:       opts.Password,
		referer:        "TAT-SDK-" + Version,
		requestTimeout: time.Minute,
		maxTries:       5,
	}
	if opts.Referer != "" {
		c.referer = opts.Referer
	}
	if opts.RequestTimeout != time.Duration(0) {
		c.requestTimeout = opts.RequestTimeout
	}
	if opts.MaxTries != 0 {
		c.maxTries = opts.MaxTries
	}

	return c, nil
}

func (c *Client) initHeaders(req *http.Request) error {
	if c == nil {
		return ErrClientNotInitiliazed
	}

	req.Header.Set(TatHeaderUsername, c.username)
	req.Header.Set(TatHeaderPassword, c.password)
	req.Header.Set(TatHeaderXTatRefererLower, c.referer)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")
	return nil
}

func (c *Client) reqWant(method string, wantCode int, path string, jsonStr []byte) ([]byte, error) {
	if c == nil {
		return nil, ErrClientNotInitiliazed
	}

	requestPath := c.url + path
	var req *http.Request
	if jsonStr != nil {
		req, _ = http.NewRequest(method, requestPath, bytes.NewReader(jsonStr))
	} else {
		req, _ = http.NewRequest(method, requestPath, nil)
	}

	c.initHeaders(req)

	if HTTPClient == nil {
		HTTPClient = &http.Client{
			Transport: &httpcontrol.Transport{
				RequestTimeout: c.requestTimeout,
				MaxTries:       c.maxTries,
			},
		}
	}
	resp, err := HTTPClient.Do(req)

	defer resp.Body.Close()

	if resp.StatusCode != wantCode {
		ErrorLogFunc("Response Status:%s", resp.Status)
		ErrorLogFunc("Request path:%s", requestPath)
		ErrorLogFunc("Request:%s", string(jsonStr))
		ErrorLogFunc("Response Headers:%s", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		ErrorLogFunc("Response Body:%s", string(body))
		return []byte{}, fmt.Errorf("Response code %d with Body:%s", resp.StatusCode, string(body))
	}
	DebugLogFunc("%s %s", method, requestPath)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrorLogFunc("Error with ioutil.ReadAll %s", err)
		return nil, fmt.Errorf("Error with ioutil.ReadAll %s", err.Error())
	}
	return body, nil
}

// CreateTopic creates a topic
func (c *Client) CreateTopic(t TopicCreateJSON) (*Topic, error) {
	if c == nil {
		return nil, ErrClientNotInitiliazed
	}

	b, err := json.Marshal(t)
	if err != nil {
		ErrorLogFunc("Error while marshal topic: %s", err)
		return nil, err
	}

	res, err := c.reqWant("POST", http.StatusCreated, "/topic", b)
	if err != nil {
		ErrorLogFunc("Error while marshal message for CreateTopic: %s", err)
		return nil, err
	}

	DebugLogFunc("createTopicResponse : %s", string(res))

	newTopic := &Topic{}
	if err := json.Unmarshal(res, newTopic); err != nil {
		return nil, err
	}

	return newTopic, nil
}

// AddMessage post a tat message
func (c *Client) AddMessage(message MessageJSON) error {
	if c == nil {
		return ErrClientNotInitiliazed
	}

	if message.Topic == "" {
		return fmt.Errorf("A message must have a Topic")
	}

	path := "/message" + message.Topic

	b, err := json.Marshal(message)
	if err != nil {
		ErrorLogFunc("Error while marshal message: %s", err)
		return err
	}

	if _, err := c.reqWant("POST", http.StatusCreated, path, b); err != nil {
		ErrorLogFunc("Error while marshal message for AddMessage: %s", err)
		return err
	}

	DebugLogFunc("Post %s done", message.Text)
	return nil
}
