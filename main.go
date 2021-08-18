package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/google/go-github/v37/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	flag "github.com/spf13/pflag"
)

// User is a github user
type User string

// Users are a slice of github users
type Users []User

// Config represents config file read in from --config
type Config struct {
	Config struct {
		Users `yaml:"users"`
	} `yaml:"config"`
}

func (c *Config) read() *Config {
	yamlFile, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

var (
	pullRequestsGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pull_request",
		Help: "GitHub pull request",
	}, []string{"user", "created_at", "link", "status"})
	pullRequestsTotalGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pull_request_total",
		Help: "Total number of pull requests opened by all users",
	}, []string{})
)

func init() {
	flag.Parse()
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}

func main() {
	config := &Config{}
	config.read()
	log.Infof("read %d users from config file", len(config.Config.Users))

	ctx := context.Background()
	client := github.NewClient(nil)

	go func() {
		for {
			now := time.Now()
			past := now.AddDate(0, 0, -*daysAgo)
			beginningSearchDate := fmt.Sprintf("%d-%02d-%02d", past.Year(), past.Month(), past.Day())
			// Track total pull requests for the total counter
			pullRequestCount := 0
			log.Info("searching for pull requests")
			// Loop all users passed in via config file
			for _, user := range config.Config.Users {
				// Get all pull requests opened by a user
				pullRequests := user.getPullRequests(ctx, client, beginningSearchDate, *ignoreUserNamespace)
				for _, pullRequest := range pullRequests {
					// Create new metric for each pull request
					pullRequestsGauge.With(prometheus.Labels{
						"user":       user.String(),
						"created_at": pullRequest.CreatedAt,
						"link":       pullRequest.PullRequestURL,
						"status":     pullRequest.Status,
					}).Set(1)
					// Increase the total counter
					pullRequestCount++
				}
				// Unauthenticated clients rate limit is 10 requests per minute
				// This sleep ensures no rate limits occur. The result is 1000 user searches every 90 minutes
				time.Sleep(time.Second * defaultRateLimitInteral)
			}
			// Set total counter
			pullRequestsTotalGauge.With(prometheus.Labels{}).Set(float64(pullRequestCount))
			// Wait for configured refresh interval
			log.Infof("finished searching for pull requests. sleeping for %d seconds", *interval)
			time.Sleep(time.Duration(*interval) * time.Second)
		}
	}()
	http.Handle("/metrics", promhttp.Handler())
	log.Infof("starting github-pr-exporter on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}