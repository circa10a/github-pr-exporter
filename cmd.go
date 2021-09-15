package main

import flag "github.com/spf13/pflag"

const (
	defaultConfigFile          = "./config.yaml"
	defaultPort                = 8080
	defaultDaysAgo             = 90
	defaultIgnoreUserNamespace = false
	defaultInterval            = 21600
	defaultRateLimitInteral    = 6
)

var configFile *string = flag.String("config", defaultConfigFile, "Path to config file")
var port *int = flag.Int("port", defaultPort, "What port to listen on")
var daysAgo *int = flag.Int("days-ago", defaultDaysAgo, "How many days back to search for pull requests")
var ignoreUserNamespace *bool = flag.Bool("ignore-user-repos", defaultIgnoreUserNamespace, "Ignore the user's own repos")
var interval *int = flag.Int("interval", defaultInterval, "How many seconds to wait before refreshing pull request data. Defaults to 6 hours")
