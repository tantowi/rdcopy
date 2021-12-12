package agent

import (
	"context"
	"log"

	"github.com/mediocregopher/radix/v4"
)

var pattern, sourcePassword, targetPassword string
var scanCount, logInterval int
var sourceUseTLS, targetUseTLS bool

func createClient(ctx context.Context, password string, redisUrl string, enableTLS bool) radix.Client {
	scannerDialer := radix.Dialer{
		AuthPass: password,
	}
	ctx = context.WithValue(ctx, "TLSEnabled", enableTLS)
	client, err := scannerDialer.Dial(ctx, "tcp", redisUrl)
	if err != nil {
		log.Fatal(err)
	}

	return client
}
