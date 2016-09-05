package tat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/facebookgo/httpcontrol"
)

// Client representes a Client configuration to connect to api
type Client struct {
	Username       string
	Password       string
	URL            string
	Referer        string
	RequestTimeout time.Duration
	MaxTries       uint
}

func (client *Client) initHeaders(req *http.Request) error {
	if client.Username == "" || client.Password == "" || client.Referer == "" {
		return fmt.Errorf("usernname, password and referer have to be setted")
	}
	req.Header.Set(TatHeaderUsername, client.Username)
	req.Header.Set(TatHeaderPassword, client.Password)
	req.Header.Set(TatHeaderXTatRefererLower, client.Referer)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")
	return nil
}

// MessageAdd post a tat message
func (client *Client) MessageAdd(message MessageJSON) error {

	if message.Topic == "" {
		return fmt.Errorf("A message must have a Topic")
	}

	path := "/message" + message.Topic

	jsonStr, err := json.Marshal(message)
	if err != nil {
		log.Errorf("Error while marshal message for postMessage")
	}

	if _, err := client.reqWant("POST", http.StatusCreated, path, jsonStr); err != nil {
		return err
	}

	log.Debugf("Post %s done", message.Text)
	return nil
}

func (client *Client) reqWant(method string, wantCode int, path string, jsonStr []byte) ([]byte, error) {
	if client.URL == "" {
		return []byte{}, fmt.Errorf("Invalid Configuration : invalid URL")
	}

	requestPath := client.URL + path
	var req *http.Request
	if jsonStr != nil {
		req, _ = http.NewRequest(method, requestPath, bytes.NewReader(jsonStr))
	} else {
		req, _ = http.NewRequest(method, requestPath, nil)
	}

	client.initHeaders(req)

	httpClient := &http.Client{
		Transport: &httpcontrol.Transport{
			RequestTimeout: client.RequestTimeout,
			MaxTries:       client.MaxTries,
		},
	}
	resp, err := httpClient.Do(req)

	defer resp.Body.Close()

	if resp.StatusCode != wantCode {
		log.Error(fmt.Sprintf("Response Status:%s", resp.Status))
		log.Error(fmt.Sprintf("Request path:%s", requestPath))
		log.Error(fmt.Sprintf("Request:%s", string(jsonStr)))
		log.Error(fmt.Sprintf("Response Headers:%s", resp.Header))
		body, _ := ioutil.ReadAll(resp.Body)
		log.Error(fmt.Sprintf("Response Body:%s", string(body)))
		return []byte{}, fmt.Errorf("Response code %d with Body:%s", resp.StatusCode, string(body))
	}
	log.Debugf("%s %s", method, requestPath)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error with ioutil.ReadAll %s", err.Error())
		return []byte{}, fmt.Errorf("Error with ioutil.ReadAll %s", err.Error())
	}
	return body, nil
}
