package agent

import (
	"crypto/tls"
	"log"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
)

var pattern, sourcePassword, targetPassword string
var scanCount, logInterval int

func createClient(redisAddr string, password string) *redis.Client {
	hostname, _, err := net.SplitHostPort(redisAddr)
	if err != nil {
		log.Fatal(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     password,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 12 * time.Second,
		MinIdleConns: 40,
		PoolSize:     40,
		TLSConfig: &tls.Config{
			ServerName: hostname,
		},
	})

	return client
}
