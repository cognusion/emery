// Package advanced provides a groupcache example.
//
// Run:
//
// ./emery --listen=:8081 --groupcache-peers=:8082,:8083 --s3bucket myimage-test --key MYK3Yc00l --expiration 30m &
// ./emery --listen=:8082 --groupcache-peers=:8081,:8083 --s3bucket myimage-test --key MYK3Yc00l --expiration 30m &
// ./emery --listen=:8083 --groupcache-peers=:8081,:8082 --s3bucket myimage-test --key MYK3Yc00l --expiration 30m &
//
// Open http://localhost:8081/_sign/path/to/image.jpg?width=100
// it will redirect to http://localhost:8081/longHMAChashHERE/path/to/image.jpg?expiration=UNIXmilliSTAMP&width=100
//
// Stats are available on http://localhost:8081/stats
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
)

var (
	configFile string
)

func init() {
	InitConfig() // get the config up and running, so we can use defaults in the CLI

	pflag.Bool(ConfigDebug, config.GetBool(ConfigDebug), "Enable debugging output")
	pflag.StringVar(&configFile, ConfigConfig, config.GetString(ConfigConfig), "Config file to load (comma-separated for multiple)")

	pflag.String(ConfigAwsRegion, config.GetString(ConfigAwsRegion), "AWS Region")
	pflag.String(ConfigAwsAccessKey, config.GetString(ConfigAwsAccessKey), "AWS Access Key. Leave blank to read from env.")
	pflag.String(ConfigAwsSecretKey, config.GetString(ConfigAwsSecretKey), "AWS Secret Key. Leave blank to read from env.")
	pflag.String(ConfigS3Bucket, config.GetString(ConfigS3Bucket), "S3 bucket to jail this server to")

	pflag.String(ConfigListen, config.GetString(ConfigListen), "HTTP listening address:port")
	pflag.Int64(ConfigGroupCacheSize, config.GetInt64(ConfigGroupCacheSize), "Groupcache size in bytes")
	pflag.String(ConfigGroupCachePeers, config.GetString(ConfigGroupCachePeers), "Groupcache peers")

	pflag.String(ConfigHMACKey, config.GetString(ConfigHMACKey), "Enable and verify request signatures using HMAC")
	pflag.String(ConfigHMACSalt, config.GetString(ConfigHMACSalt), "With --key, salts the plaintext before summing")
	pflag.Duration(ConfigHMACExpiration, config.GetDuration(ConfigHMACExpiration), "With --key, signatures expire after so long")
	pflag.Parse()

	err := LoadConfig(configFile, config) // Load the config from configFile into the config
	if err != nil {
		fmt.Printf("Error loading config '%s': %s\n", configFile, err)
		os.Exit(1)
	}

	config.BindPFlags(pflag.CommandLine) // Bind commandline flags to viper config

	// Set up the the loggers

	if config.GetBool(ConfigDebug) {
		DebugOut = log.New(os.Stderr, "[DEBUG] ", OutFormat)
		TimingOut = log.New(os.Stderr, "[TIMING] ", 0)
	}
}

func main() {
	startHTTPServer()
}
