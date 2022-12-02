// Package advanced provides a groupcache example.
//
// Run:
//
// ./emery --http=:8081 --groupcache-peers=:8082,:8083 --s3bucket myimage-test --key MYK3Yc00l --signduration 30m &
// ./emery --http=:8082 --groupcache-peers=:8081,:8083 --s3bucket myimage-test --key MYK3Yc00l --signduration 30m &
// ./emery --http=:8083 --groupcache-peers=:8081,:8082 --s3bucket myimage-test --key MYK3Yc00l --signduration 30m &
//
// Open http://localhost:8081/_sign/path/to/image.jpg?width=100
// it will redirect to http://localhost:8081/longHMAChashHERE/path/to/image.jpg?expiration=UNIXmilliSTAMP&width=100
//
// Stats are available on http://localhost:8081/stats
package main

import (
	"log"
	"os"
	"time"

	"github.com/spf13/pflag"
)

var (
	flagHTTP            = ":8080"
	flagGroupcache      = int64(128 * (1 << 20))
	flagGroupcachePeers string
	debug               bool

	awsRegion    string
	awsAccessKey string
	awsSecretKey string
	awsS3bucket  string

	hmacKey      string
	hmacDuration time.Duration
)

func init() {
	pflag.BoolVar(&debug, "debug", false, "Enable debugging output")

	pflag.StringVar(&awsRegion, "region", "us-east-1", "AWS Region")
	pflag.StringVar(&awsAccessKey, "accesskey", "", "AWS Access Key. Leave blank to read from env.")
	pflag.StringVar(&awsSecretKey, "secretkey", "", "AWS Secret Key. Leave blank to read from env.")
	pflag.StringVar(&awsS3bucket, "s3bucket", "", "S3 bucket to jail this server to")

	pflag.StringVar(&flagHTTP, "http", flagHTTP, "HTTP listening address:port")
	pflag.Int64Var(&flagGroupcache, "cachesize", flagGroupcache, "Groupcache size in bytes")
	pflag.StringVar(&flagGroupcachePeers, "groupcache-peers", flagGroupcachePeers, "Groupcache peers")

	pflag.StringVar(&hmacKey, "key", "", "Enable and verify request signatures using HMAC")
	pflag.DurationVar(&hmacDuration, "signduration", 0, "With --key, signatures expire after so long")
	pflag.Parse()

	if debug {
		DebugOut = log.New(os.Stderr, "[DEBUG] ", OutFormat)
		TimingOut = log.New(os.Stderr, "[TIMING] ", 0)
	}
}

func main() {
	startHTTPServer()
}
