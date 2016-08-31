package cache

import (
	"strings"

	"gopkg.in/redis.v4"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

var instance Cache

//Client returns Cache interface
func Client() Cache {
	redisHosts := viper.GetString("redis_hosts")
	redisPassword := viper.GetString("redis_password")
	redisHostsArray := strings.Split(redisHosts, ",")

	if instance != nil {
		goto testInstance
	}

	if redisHosts == "" {
		//Mode in memory
		instance = &LocalCache{}
		goto testInstance
	}

	if len(redisHostsArray) > 1 {
		//Mode in cluster
	} else {
		//Mode master
		opts := &redis.Options{
			Addr:     redisHosts,
			Password: redisPassword,
		}
		instance = redis.NewClient(opts)
	}

testInstance:
	if err := instance.Ping().Err(); err != nil {
		logrus.Errorf("Unable to ping redis at %s : %s", redisHosts, err)
	}

	return instance
}

//Key convert string array in redis key
func Key(s ...string) string {
	var escape = func(s string) string {
		return strings.Replace(s, ":", "_", -1)
	}

	for i := range s {
		s[i] = escape(s[i])
	}

	return strings.Join(s, ":")
}
