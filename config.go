package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Constants for configuration key strings
const (
	ConfigConfig          = ConfigKey("config")
	ConfigAccessLog       = ConfigKey("accesslog")
	ConfigCommonLog       = ConfigKey("commonlog")
	ConfigDebug           = ConfigKey("debug")
	ConfigDebugLog        = ConfigKey("debuglog")
	ConfigErrorLog        = ConfigKey("errorlog")
	ConfigAwsEC2          = ConfigKey("aws.ec2")
	ConfigAwsRegion       = ConfigKey("aws.region")
	ConfigAwsAccessKey    = ConfigKey("aws.access")
	ConfigAwsSecretKey    = ConfigKey("aws.secret")
	ConfigS3Bucket        = ConfigKey("s3bucket")
	ConfigListen          = ConfigKey("listen")
	ConfigLogAge          = ConfigKey("logage")
	ConfigLogBackups      = ConfigKey("logbackups")
	ConfigLogSize         = ConfigKey("logsize")
	ConfigGroupCacheSize  = ConfigKey("groupcache-size")
	ConfigGroupCachePeers = ConfigKey("groupcache-peers")
	ConfigHMACKey         = ConfigKey("key")
	ConfigHMACSalt        = ConfigKey("salt")
	ConfigHMACExpiration  = ConfigKey("expiration")
)

// ConfigKey is a string type for static config key name consistency
type ConfigKey = string

var config *viper.Viper

// InitConfig creates an config, initialized with defaults and environment-set values, and returns it
func InitConfig() {
	v := viper.New()

	v.AutomaticEnv()
	v.SetEnvPrefix("EMERY")

	loadDefaults(v)

	config = v
}

// LoadConfig read the config file and returns a config object or an error
func LoadConfig(configFilename string, v *viper.Viper) error {

	if configFilename != "" {
		configFilenames := strings.Split(configFilename, ",")
		v.SetConfigFile(configFilenames[0])

		err := v.ReadInConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigParseError); ok {
				return err
			}
			return fmt.Errorf("unable to locate config file '%s': %w", configFilenames[0], err)
		}
		for _, configFile := range configFilenames[1:] {
			file, err := os.Open(configFile) // For read access.
			if err != nil {
				return fmt.Errorf("unable to open config file '%s': %w", configFile, err)
			}
			defer file.Close()
			if err = v.MergeConfig(file); err != nil {
				return fmt.Errorf("unable to parse/merge Config file '%s': %w", configFile, err)
			}
		}
	}

	return nil
}

// loadDefaults sets all the default values for the config. An error may be returned, to support future operations.
func loadDefaults(v *viper.Viper) error {
	v.SetDefault(ConfigDebug, false)                       // Enable vociferous output
	v.SetDefault(ConfigListen, ":8080")                    // ip:port or :port to listen on
	v.SetDefault(ConfigAccessLog, "")                      // Path to file where accesslog, else stdout
	v.SetDefault(ConfigDebugLog, "")                       // Path to file where debug should log to, else stderr
	v.SetDefault(ConfigErrorLog, "")                       // Path to file where errorlog should log to, else stderr
	v.SetDefault(ConfigLogSize, 100)                       // Maximum size, in MB, that the currently log can be before rolling
	v.SetDefault(ConfigLogBackups, 3)                      // Maximum number of rolled logs to keep
	v.SetDefault(ConfigLogAge, 28)                         // Maximum age, in days, to keep rolled logs
	v.SetDefault(ConfigAwsEC2, false)                      // Enable AWS EC2-specific features
	v.SetDefault(ConfigAwsRegion, "us-east-1")             // Set AWS Region
	v.SetDefault(ConfigGroupCacheSize, int64(128*(1<<20))) // Set the local cache size

	return nil
}
