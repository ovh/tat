package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ovh/tat"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

var instance *tat.Client

// Client return a new Tat Client
func Client() *tat.Client {
	ReadConfig()
	if instance != nil {
		return instance
	}

	// Init basic auth credentials
	var basicAuthUsername, basicAuthPassword string
	if viper.GetBool("basic-auth") {
		// Only try to get basic auth credentials when --basic-auth flag is set

		// First, try to get basic auth credentials from the configuration file
		basicAuthUsername = viper.GetString("basicAuthUsername")
		basicAuthPassword = viper.GetString("basicAuthPassword")

		// If the username is not defined, ask the user to provide it
		if basicAuthUsername == "" {
			fmt.Println("No basic auth username found")
			fmt.Print("Please enter your username: ")
			fmt.Scan(&basicAuthUsername)
		}

		// If the password is not defined, ask the user to provide it
		if basicAuthPassword == "" {
			fmt.Printf("No basic auth password found for username %v\n", basicAuthUsername)
			fmt.Printf("Please enter the password for username %v: ", basicAuthUsername)
			pwd, errReadPassword := terminal.ReadPassword(0)
			Check(errReadPassword)
			basicAuthPassword = string(pwd)
		}
	}

	tc, err := tat.NewClient(tat.Options{
		URL:                   viper.GetString("url"),
		Username:              viper.GetString("username"),
		Password:              viper.GetString("password"),
		BasicAuthUsername:     basicAuthUsername, // Always send the basic auth credentials (if --basic-auth is not set, the value will be empty)
		BasicAuthPassword:     basicAuthPassword, // Always send the basic auth credentials (if --basic-auth is not set, the value will be empty)
		Referer:               "tatcli.v." + tat.Version,
		SSLInsecureSkipVerify: viper.GetBool("sslInsecureSkipVerify"),
	})

	if err != nil {
		log.Fatalf("Error while create new Tat Client: %s", err)
	}

	tat.DebugLogFunc = log.Debugf

	if Debug {
		tat.IsDebug = true
	}

	// Set the instance for future use and return it
	instance = tc
	return tc
}

// GetSkipLimit gets skip and limit in args array
// default skip to 0 and limit to 10
func GetSkipLimit(args []string) (int, int) {
	skip := "0"
	limit := "10"
	if len(args) == 3 {
		skip = args[1]
		limit = args[2]
	} else if len(args) == 2 {
		skip = args[0]
		limit = args[1]
	}
	s, e1 := strconv.Atoi(skip)
	Check(e1)
	l, e2 := strconv.Atoi(limit)
	Check(e2)
	return s, l
}

func getJSON(s []byte) string {
	if Pretty {
		var out bytes.Buffer
		json.Indent(&out, s, "", "  ")
		return out.String()
	}
	return string(s)
}

// Print prints json return
func Print(v interface{}) {
	switch v.(type) {
	case []byte:
		fmt.Printf("%s", v)
	default:
		out, err := tat.Sprint(v)
		Check(err)
		fmt.Print(getJSON(out))
	}
}

// Check checks error, if != nil, throw panic
func Check(e error) {
	if e != nil {
		if ShowStackTrace {
			panic(e)
		} else {
			log.Fatalf("%s", e)
		}
	}
}
