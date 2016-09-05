package tests

import (
	"flag"
	"net/http"
	"strconv"
	"sync"
	"testing"

	"net/http/httptest"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/ovh/tat"
	"github.com/ovh/tat/api/store"
	"github.com/spf13/viper"
)

var initiliazed = false
var dbAddr, dbUser, dbPassword string
var mutex = sync.Mutex{}
var testsRouterGroups = map[*testing.T]*gin.RouterGroup{}
var testsEngine = map[*testing.T]*gin.Engine{}
var testsIndex = 0

// Init the test context with the database
func Init(t *testing.T) {
	if initiliazed {
		return
	}
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
	})

	flag.StringVar(&dbAddr, "db-addr", "127.0.0.1:27017", "Address of the mongodb server")
	flag.StringVar(&dbUser, "db-user", "", "User to authenticate with the mongodb server")
	flag.StringVar(&dbPassword, "db-password", "", "Password to authenticate with the mongodb server")

	flag.Parse()

	viper.Set("db_addr", dbAddr)
	viper.Set("db_user", dbUser)
	viper.Set("db_password", dbPassword)
	viper.Set("header_trust_username", "X-Remote-User")

	if err := store.NewStore(); err != nil {
		t.Errorf("Error initializing test context : %s", err)
		t.Fail()
		return
	}

	log.Infof(">>> Connected to database %s", dbAddr)
	initiliazed = true
}

// Router prepare a gin router for test purpose
func Router(t *testing.T) *gin.RouterGroup {
	mutex.Lock()
	defer mutex.Unlock()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		c.Status(200)
	})
	testsIndex++
	g := r.Group("test" + strconv.Itoa(testsIndex))
	testsRouterGroups[t] = g
	testsEngine[t] = r

	return g
}

// DoRequest executes request for tests
func DoRequest(t *testing.T, r *http.Request) *httptest.ResponseRecorder {
	router := testsEngine[t]
	if router == nil {
		t.Fail()
		return nil
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w
}

// Handle associates a method & path on an handler (h)
func Handle(t *testing.T, method, path string, handler ...gin.HandlerFunc) {
	g := testsRouterGroups[t]
	if g == nil {
		t.Fail()
		return
	}
	handle(g, method, path, handler...)
	return
}

func handle(g *gin.RouterGroup, m string, s string, h ...gin.HandlerFunc) {
	g.Handle(m, s, h...)
}

func FakeAuthHandler(t *testing.T, username string, referer string, isAdmin bool, isSystem bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(tat.TatHeaderUsername, username)
		ctx.Set("Tat_isAdmin", isAdmin)
		ctx.Set("Tat_isSystem", isSystem)
		ctx.Set(tat.TatHeaderXTatRefererLower, referer)
	}
}

// TATClient is a client
func TATClient(t *testing.T, username string) *tat.Client {
	g := testsRouterGroups[t].BasePath()
	client, _ := tat.NewClient(tat.Options{
		URL:      g,
		Username: username,
		Password: "no_password_for_tests",
	})
	tat.HTTPClient = getTestHTTPClient(t)
	tat.ErrorLogFunc = t.Errorf
	tat.DebugLogFunc = t.Logf
	return client
}
