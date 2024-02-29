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
		ReadTimeout:  10,
		WriteTimeout: 12,
		MinIdleConns: 40,
		PoolSize:     40,
		TLSConfig: &tls.Config{
			ServerName: hostname,
		},
	})

	return client
}
