package cache

import (
	"strings"

	"gopkg.in/redis.v4"

	log "github.com/Sirupsen/logrus"
	"github.com/ovh/tat"
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
		opts := &redis.ClusterOptions{
			Addrs:    redisHostsArray,
			Password: redisPassword,
		}
		instance = redis.NewClusterClient(opts)
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
		log.Errorf("Unable to ping redis at %s: %s", redisHosts, err)
	}

	return instance
}

// TestInstanceAtStartup pings redis and display error log if no redis, and Info
// log is redis is here
func TestInstanceAtStartup() {
	if viper.GetString("redis_hosts") == "" {
		log.Infof("TAT is NOT linked to a redis")
		return
	}
	if err := Client().Ping().Err(); err != nil {
		log.Errorf("Unable to ping redis at %s: %s", viper.GetString("redis_hosts"), err)
	} else {
		log.Infof("TAT is linked to redis %s", viper.GetString("redis_hosts"))
	}
}

//CriteriaKey returns the Redis Key
func CriteriaKey(i tat.CacheableCriteria, s ...string) string {
	k := i.CacheKey()
	return Key(s...) + ":" + Key(k...)
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

func removeSomeMembers(key string, members ...string) {
	m := make([]interface{}, len(members))
	for i, member := range members {
		m[i] = member
	}
	Client().SRem(key, m...)
}
