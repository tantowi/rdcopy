package agent

import (
	"context"
	"crypto/tls"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var pattern string

// var sourcePassword, targetPassword string
var scanCount, logInterval int

// createClient initializes a new Redis client with the provided address and password.
// It sets up the connection with appropriate timeouts and TLS configuration.
// The RedisUrl should be in the format "redis://user:password@example.com:6379/db" or rediss://user:password@example.com:6379/db for secure connections.
// If the address is invalid, return error
func createClient(redisUrl string) (*redis.Client, error) {

	prl, err := url.Parse(redisUrl)
	if err != nil {
		return nil, errors.New(`invalid address format. Expected format is "redis://user:password@example.com:6379/db" or rediss://user:password@example.com:6379/db for secure connections`)
	}

	if prl.Scheme != "redis" && prl.Scheme != "rediss" {
		return nil, errors.New("invalid scheme. Expected 'redis' or 'rediss'")
	}

	host := prl.Hostname()
	if host == "" {
		return nil, errors.New("host cannot be empty")
	}

	var cport = prl.Port()
	if cport == "" || cport == "0" {
		cport = "6379" // Default Redis port
	}

	_, err = strconv.Atoi(cport)
	if err != nil {
		return nil, errors.New("invalid port number: " + cport)
	}

	var addr = prl.Host // Combine the host and port into a single address

	var user string
	var pass string
	if prl.User != nil {
		user = prl.User.Username()    // Get the username from the URL
		pass, _ = prl.User.Password() // Get the password from the URL, if available
	}

	cdb := prl.Path
	if cdb == "" {
		cdb = "0" // Default Redis database
	}

	db, err := strconv.Atoi(cdb)
	if err != nil {
		return nil, errors.New("Invalid database: " + cdb)
	}

	var options = redis.Options{
		Addr:         addr,             // Use the full address including host and port
		Username:     user,             // Use the username from the URL
		DB:           db,               // Use the database number from the URL
		Password:     pass,             // Use the password from the URL, if available
		ReadTimeout:  60 * time.Second, // Set read timeout to 60 seconds
		WriteTimeout: 60 * time.Second, // Set write timeout to 60 seconds
		MinIdleConns: 10,               // Set minimum idle connections to 10
		PoolSize:     10,               // Set the maximum number of connections in the pool to 10
	}

	// If the scheme is "rediss", enable TLS
	if prl.Scheme == "rediss" {
		options.TLSConfig = &tls.Config{
			ServerName: host,             // Use the host for server name verification
			MinVersion: tls.VersionTLS12, // Ensure a minimum TLS version
		}
	}

	// Create a new Redis client with the specified options
	// The client will use the provided address, password, and TLS configuration
	var client = redis.NewClient(&options)

	if client == nil {
		return nil, errors.New("failed to create Redis client")
	}

	// Set a timeout for the connection to ensure it doesn't hang indefinitely
	// Use a context with a timeout to avoid long waits on connection issues
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.New("failed to connect to Redis: " + err.Error())
	}

	return client, nil
}
