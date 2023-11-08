package agent

import (
	"crypto/tls"
	"log"
	"net"

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
		ReadTimeout:  0,
		WriteTimeout: 0,
		PoolSize:     30,
		MinIdleConns: 30,
		TLSConfig: &tls.Config{
			ServerName: hostname,
		},
	})

	return client
}
